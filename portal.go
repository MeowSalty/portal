package portal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MeowSalty/portal/adapter"
	"github.com/MeowSalty/portal/health"
	"github.com/MeowSalty/portal/selector"
	"github.com/MeowSalty/portal/stats"
	"github.com/MeowSalty/portal/types"
)

// SelectorStrategy 定义选择器策略类型
type SelectorStrategy string

const (
	RandomSelectorStrategy SelectorStrategy = "random"
	LRUSelectorStrategy    SelectorStrategy = "lru"
)

var (
	ErrServerShuttingDown = errors.New("服务正在停机，请稍后重试")
	ErrShutdownTimeout    = errors.New("等待会话完成超时")
)

// Option 定义用于配置 GatewayManager 的选项函数类型
type Option func(*options)

type options struct {
	selectorStrategy   SelectorStrategy
	healthSyncInterval time.Duration
	logger             *slog.Logger
}

// WithSelectorStrategy 设置选择器策略
func WithSelectorStrategy(strategy SelectorStrategy) Option {
	return func(o *options) {
		o.selectorStrategy = strategy
	}
}

// WithHealthSyncInterval 设置健康检查同步间隔
func WithHealthSyncInterval(interval time.Duration) Option {
	return func(o *options) {
		o.healthSyncInterval = interval
	}
}

// WithLogger 设置日志记录器
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// GatewayManager 是 portal 包的核心协调器
//
// 它负责管理模型、平台和 API 密钥，处理请求路由、负载均衡和健康检查
type GatewayManager struct {
	repo           types.DataRepository
	healthManager  *health.Manager
	selector       types.ChannelSelector
	adapters       map[string]types.Adapter // Key: Platform.Format
	logger         *slog.Logger
	statsManager   *stats.Manager
	isShuttingDown atomic.Bool
	activeSessions sync.WaitGroup
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// New 从配置创建并初始化一个新的 GatewayManager
//
// 该函数会初始化所有适配器并设置日志记录器
func New(ctx context.Context, logger *slog.Logger, repo types.DataRepository, opts ...Option) *GatewayManager {
	// 应用选项
	opt := &options{
		selectorStrategy:   RandomSelectorStrategy,
		healthSyncInterval: time.Minute,
	}

	for _, o := range opts {
		o(opt)
	}

	if logger == nil {
		logger = slog.Default()
	}
	adapterLogger := logger.WithGroup("adapter")

	// 使用 adapter 包中的注册机制初始化适配器
	adapters := adapter.New(adapterLogger)

	// 初始化健康状态管理器
	healthManager, err := health.New(ctx, repo, logger, opt.healthSyncInterval)
	if err != nil {
		logger.Error("failed to create health manager", "error", err)
		// 如果健康管理器创建失败，使用一个默认的
		healthManager = nil
	}

	// 初始化 selector
	var sel types.ChannelSelector
	switch opt.selectorStrategy {
	case LRUSelectorStrategy:
		sel = selector.NewLeastRecentlyUsedSelector(healthManager)
	default:
		sel = selector.NewRandomSelector(healthManager)
	}

	// 初始化统计管理器
	statsManager := stats.NewManager(repo, logger)

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())

	return &GatewayManager{
		repo:           repo,
		healthManager:  healthManager,
		selector:       sel,
		adapters:       adapters,
		logger:         logger,
		statsManager:   statsManager,
		shutdownCtx:    shutdownCtx,
		shutdownCancel: shutdownCancel,
	}
}

// Shutdown 优雅地关闭 GatewayManager
//
// 它会等待所有正在进行的会话完成，然后关闭健康管理器。
// 可以通过可选的 timeout 参数设置最长等待时间。
// 如果等待超时，所有正在进行的会话将被中断。
func (m *GatewayManager) Shutdown(timeout time.Duration) error {
	// 1. 标记服务正在停机，拒绝新请求
	m.isShuttingDown.Store(true)
	m.logger.Info("服务开始停机，不再接受新请求")

	// 2. 等待所有活动会话完成
	done := make(chan struct{})
	go func() {
		m.activeSessions.Wait()
		close(done)
	}()

	var err error
	if timeout > 0 {
		select {
		case <-done:
			m.logger.Info("所有活动会话已正常完成")
		case <-time.After(timeout):
			m.logger.Warn("停机等待超时，正在中断所有剩余会话...")
			m.shutdownCancel() // 触发所有关联上下文的取消
			<-done             // 等待被中断的会话完成清理
			m.logger.Info("所有被中断的会话已结束")
			err = ErrShutdownTimeout
		}
	} else {
		// No timeout, wait indefinitely
		<-done
		m.logger.Info("所有活动会话已正常完成")
	}

	// 3. 关闭健康管理器
	if m.healthManager != nil {
		m.logger.Info("正在关闭健康管理器")
		m.healthManager.Shutdown()
	}

	m.logger.Info("服务已成功停机")
	return err
}

// FindModelsByName 根据名称查找模型
func (m *GatewayManager) FindModelsByName(ctx context.Context, name string) ([]*types.Model, error) {
	return m.repo.FindModelsByName(ctx, name)
}

// buildChannels 从模型列表创建所有可能的通道列表
//
// 该方法会为每个模型获取对应的平台和 API 密钥，并构建通道对象
func (m *GatewayManager) buildChannels(ctx context.Context, models []*types.Model) ([]*types.Channel, error) {
	var channels []*types.Channel
	var errs []error

	for _, model := range models {
		platform, err := m.repo.GetPlatformByID(ctx, model.PlatformID)
		if err != nil {
			m.logger.Error("获取模型平台失败",
				slog.Uint64("模型 ID", uint64(model.ID)),
				slog.Uint64("平台 ID", uint64(model.PlatformID)),
				slog.String("错误", err.Error()))
			errs = append(errs, fmt.Errorf("模型 ID %d: 获取平台失败：%w", model.ID, err))
			continue
		}

		apiKeys, err := m.repo.GetAllAPIKeys(ctx, platform.ID)
		if err != nil {
			m.logger.Error("获取平台 API 密钥失败",
				slog.Uint64("平台 ID", uint64(platform.ID)),
				slog.String("错误", err.Error()))
			errs = append(errs, fmt.Errorf("平台 ID %d: 获取 API 密钥失败：%w", platform.ID, err))
			continue
		}

		if len(apiKeys) == 0 {
			m.logger.Warn("平台没有配置 API 密钥",
				slog.Uint64("平台 ID", uint64(platform.ID)),
				slog.String("平台名称", platform.Name))
			continue
		}

		for _, key := range apiKeys {
			channels = append(channels, &types.Channel{
				Platform: platform,
				Model:    model,
				APIKey:   key,
			})
		}
	}

	// 如果没有成功构建任何通道但有错误，则返回错误
	if len(channels) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("构建通道失败: %v", errs)
	}

	return channels, nil
}

// ChatCompletion 处理聊天完成请求
//
// 该方法会根据请求的模型名称查找匹配的模型，选择可用通道，并执行聊天完成请求
func (m *GatewayManager) ChatCompletion(ctx context.Context, request *types.Request) (*types.Response, error) {
	if m.isShuttingDown.Load() {
		return nil, ErrServerShuttingDown
	}
	m.activeSessions.Add(1)
	defer m.activeSessions.Done()

	// 创建一个与 shutdown 信号关联的新上下文
	reqCtx, reqCancel := context.WithCancel(ctx)
	defer reqCancel()
	go func() {
		select {
		case <-m.shutdownCtx.Done():
			reqCancel()
		case <-reqCtx.Done():
		}
	}()

	processor := NewRequestProcessor(m, reqCtx, request.Model, "non-stream")

	response, err := processor.processChatCompletion(request)

	// 如果 processor 内部没有记录统计信息，则在这里记录
	if err != nil {
		// 记录统计信息
		duration := time.Since(processor.startTime)
		success := err == nil
		processor.recordStat(&stats.RecordOptions{
			Timestamp:   processor.startTime,
			RequestType: processor.requestType,
			ModelName:   request.Model,
			Duration:    duration,
			Success:     success,
			ErrorMsg:    getErrorMsgFromError(err),
		})
	}

	return response, err
}

// Completion 处理文本补全请求
//
// 该方法会根据请求的模型名称查找匹配的模型，选择可用通道，并执行文本补全请求
// func (m *GatewayManager) Completion(ctx context.Context, request *core.CompletionRequest) (*core.CompletionResponse, error) {
// 	processor := &requestProcessor{
// 		manager: m,
// 		ctx:     ctx,
// 		model:   request.Model,
// 		logger:  m.logger,
// 	}

// 	return processor.processCompletion(request)
// }

// ChatCompletionStream 处理流式聊天完成请求
//
// 该方法会根据请求的模型名称查找匹配的模型，选择可用通道，并执行流式聊天完成请求
func (m *GatewayManager) ChatCompletionStream(ctx context.Context, request *types.Request) (<-chan *types.Response, error) {
	if m.isShuttingDown.Load() {
		return nil, ErrServerShuttingDown
	}
	m.activeSessions.Add(1)

	// 创建一个与 shutdown 信号关联的新上下文
	reqCtx, reqCancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-m.shutdownCtx.Done():
			reqCancel()
		case <-reqCtx.Done():
		}
	}()

	processor := &requestProcessor{
		manager:     m,
		ctx:         reqCtx, // 使用新的关联上下文
		model:       request.Model,
		logger:      m.logger,
		requestType: "stream",
		startTime:   time.Now(),
	}

	stream, err := processor.processChatCompletionStream(request)

	// 只有在启动时出错才记录统计信息，否则由 processor 内部处理
	if err != nil {
		reqCancel()             // 清理上下文 goroutine
		m.activeSessions.Done() // 如果启动失败，需要在这里减少计数
		// 如果启动流式传输失败，记录统计信息
		duration := time.Since(processor.startTime)
		success := err == nil
		processor.recordStat(&stats.RecordOptions{
			Timestamp:   processor.startTime,
			RequestType: processor.requestType,
			ModelName:   request.Model,
			Duration:    duration,
			Success:     success,
			ErrorMsg:    getErrorMsgFromError(err),
		})
		return nil, err
	}

	// 对于成功的流式请求，我们包装通道以记录统计信息
	wrappedStream := make(chan *types.Response)

	go func() {
		defer close(wrappedStream)
		defer reqCancel()             // 清理上下文 goroutine
		defer m.activeSessions.Done() // 在流结束后减少计数
		for response := range stream {
			// 转发响应
			wrappedStream <- response
		}
	}()

	return wrappedStream, nil
}

// CompletionStream 处理流式文本补全请求
//
// 该方法会根据请求的模型名称查找匹配的模型，选择可用通道，并执行流式文本补全请求
// func (m *GatewayManager) CompletionStream(ctx context.Context, request *core.CompletionRequest) (<-chan *core.CompletionStreamResponse, error) {
// 	processor := &requestProcessor{
// 		manager: m,
// 		ctx:     ctx,
// 		model:   request.Model,
// 		logger:  m.logger,
// 	}

// 	return processor.processCompletionStream(request)
// }

// QueryStats 查询请求统计列表
func (m *GatewayManager) QueryStats(ctx context.Context, params *types.StatsQueryParams) ([]*types.RequestStat, error) {
	return m.statsManager.QueryStats(ctx, params)
}

// CountStats 统计请求计数
func (m *GatewayManager) CountStats(ctx context.Context, params *types.StatsQueryParams) (*types.StatsSummary, error) {
	return m.statsManager.CountStats(ctx, params)
}

// recordStat 记录请求统计信息
func (p *requestProcessor) recordStat(opts *stats.RecordOptions) {
	// 在后台记录统计信息，避免阻塞主请求处理流程
	go func() {
		// 使用一个新上下文，避免因为主上下文取消而无法记录统计信息
		bgCtx := context.Background()
		if err := p.manager.statsManager.RecordRequestStat(bgCtx, opts); err != nil {
			p.logger.Error("failed to record request stat", "error", err)
		}
	}()
}

// getErrorMsgFromError 从错误中提取错误信息
func getErrorMsgFromError(err error) *string {
	if err != nil {
		msg := err.Error()
		return &msg
	}
	return nil
}

// GetHealthStatus 获取特定资源的健康状态
//
// 该方法根据资源类型和资源 ID 返回对应的健康状态信息
func (m *GatewayManager) GetHealthStatus(resourceType types.ResourceType, resourceID uint) *types.Health {
	return m.healthManager.GetStatus(resourceType, resourceID)
}

// requestProcessor 处理请求的结构体
//
// 该结构体封装了处理各种 AI 请求的通用逻辑和上下文信息
type requestProcessor struct {
	manager *GatewayManager
	ctx     context.Context
	model   string
	logger  *slog.Logger
	// 添加统计相关字段
	requestType  string
	startTime    time.Time
	statsManager *stats.Manager
}

// NewRequestProcessor 创建一个新的请求处理器
func NewRequestProcessor(manager *GatewayManager, ctx context.Context, model, requestType string) *requestProcessor {
	return &requestProcessor{
		manager:      manager,
		ctx:          ctx,
		model:        model,
		logger:       manager.logger,
		requestType:  requestType,
		startTime:    time.Now(),
		statsManager: manager.statsManager,
	}
}

// processChatCompletion 处理聊天完成请求
//
// 该方法实现了完整的请求处理流程：
//
//  1. 查找匹配的模型
//  2. 构建所有可能的通道
//  3. 过滤出健康的通道
//  4. 根据策略选择一个通道
//  5. 获取对应的适配器并执行请求
//  6. 根据执行结果更新健康状态并返回结果
func (p *requestProcessor) processChatCompletion(request *types.Request) (*types.Response, error) {
	// 1. 查找匹配的模型
	models, err := p.manager.FindModelsByName(p.ctx, request.Model)
	if err != nil {
		return nil, fmt.Errorf("查找模型时出错：%w", err)
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("没有找到匹配的模型")
	}
	p.logger.Info("找到匹配的模型", slog.Int("数量", len(models)))

	// 2. 构建所有可能的通道
	allChannels, err := p.manager.buildChannels(p.ctx, models)
	if err != nil {
		return nil, fmt.Errorf("构建通道时出错：%w", err)
	}

	// 如果没有可用通道，直接返回错误
	if len(allChannels) == 0 {
		p.logger.Warn("没有可用的通道")
		return nil, fmt.Errorf("没有可用的通道")
	}

	// 缓存当前时间，避免在 FilterHealthyChannels 中多次调用 time.Now()
	now := time.Now()

	// 循环重试直到成功或没有可用通道
	for {
		// 3. 过滤出健康的通道
		healthyChannels := p.manager.healthManager.FilterHealthyChannelsWithTime(allChannels, now)
		if len(healthyChannels) == 0 {
			p.logger.Warn("没有可用的健康通道")
			return nil, fmt.Errorf("没有可用的通道")
		}
		p.logger.Info("筛选出健康通道", slog.Int("健康数量", len(healthyChannels)), slog.Int("总数", len(allChannels)))

		// 4. 根据策略选择一个通道
		selectedChannel, err := p.manager.selector.Select(p.ctx, healthyChannels)
		if err != nil {
			return nil, fmt.Errorf("通道选择器失败：%w", err)
		}
		p.logger.Info("已选择通道",
			slog.String("平台", selectedChannel.Platform.Name),
			slog.String("模型", selectedChannel.Model.Name))

		// 5. 获取对应的适配器并执行请求
		adapter, ok := p.manager.adapters[selectedChannel.Platform.Format]
		if !ok {
			p.logger.Error("适配器未找到", slog.String("格式", selectedChannel.Platform.Format))
			// 从通道列表中移除无效的通道
			allChannels = removeChannel(allChannels, selectedChannel)

			// 如果没有剩余通道，返回错误
			if len(allChannels) == 0 {
				return nil, fmt.Errorf("适配器未找到：%s", selectedChannel.Platform.Format)
			}

			// 更新时间戳并继续尝试其他通道
			now = time.Now()
			continue
		}

		response, err := adapter.ChatCompletion(p.ctx, request, selectedChannel)
		if err == nil {
			// 6a. 成功时更新健康状态并返回
			p.logger.Info("聊天完成请求成功")
			p.manager.healthManager.UpdateStatusOnSuccess(selectedChannel)

			// 记录统计信息
			p.recordStat(&stats.RecordOptions{
				Timestamp:   p.startTime,
				RequestType: p.requestType,
				ModelName:   request.Model,
				Duration:    time.Since(p.startTime),
				Success:     true,
			})

			return response, nil
		}

		// 检查是否因为上下文取消而失败
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			p.logger.Warn("请求因上下文取消而中断", "error", err)
			// 不再尝试其他通道，直接返回错误
			return nil, err
		}

		// 6b. 失败时更新健康状态并准备重试
		p.logger.Warn("聊天完成请求失败，将尝试其他可用通道", slog.String("错误", err.Error()))
		p.manager.healthManager.UpdateStatusOnFailure(selectedChannel, err)

		// 记录统计信息
		p.recordStat(&stats.RecordOptions{
			Timestamp:   p.startTime,
			RequestType: p.requestType,
			ModelName:   request.Model,
			Duration:    time.Since(p.startTime),
			Success:     false,
			ErrorMsg:    getErrorMsgFromError(err),
		})

		// 从通道列表中移除失败的通道
		allChannels = removeChannel(allChannels, selectedChannel)

		// 如果没有剩余通道，返回错误
		if len(allChannels) == 0 {
			p.logger.Error("所有可用通道都未能处理请求")
			return nil, fmt.Errorf("没有可用的通道")
		}

		// 更新时间戳以用于下一轮健康检查
		now = time.Now()
	}
}

// processChatCompletionStream 处理流式聊天完成请求
//
// 该方法实现了完整的流式请求处理流程：
//
//  1. 查找匹配的模型
//  2. 构建所有可能的通道
//  3. 过滤出健康的通道
//  4. 根据策略选择一个通道
//  5. 获取对应的适配器并执行请求
//  6. 根据执行结果更新健康状态并返回结果
func (p *requestProcessor) processChatCompletionStream(request *types.Request) (<-chan *types.Response, error) {
	// 1. 查找匹配的模型
	models, err := p.manager.FindModelsByName(p.ctx, request.Model)
	if err != nil {
		return nil, fmt.Errorf("查找模型时出错：%w", err)
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("没有找到匹配的模型")
	}
	p.logger.Info("找到匹配的模型", slog.Int("数量", len(models)))

	// 2. 构建所有可能的通道
	allChannels, err := p.manager.buildChannels(p.ctx, models)
	if err != nil {
		return nil, fmt.Errorf("构建通道时出错：%w", err)
	}

	// 如果没有可用通道，直接返回错误
	if len(allChannels) == 0 {
		p.logger.Warn("没有可用的通道")
		return nil, fmt.Errorf("没有可用的通道")
	}

	// 缓存当前时间，避免在 FilterHealthyChannels 中多次调用 time.Now()
	now := time.Now()

	// 循环重试直到成功或没有可用通道
	for {
		// 3. 过滤出健康的通道
		healthyChannels := p.manager.healthManager.FilterHealthyChannelsWithTime(allChannels, now)
		if len(healthyChannels) == 0 {
			p.logger.Warn("没有可用的健康通道")
			return nil, fmt.Errorf("没有可用的通道")
		}
		p.logger.Info("筛选出健康通道", slog.Int("健康数量", len(healthyChannels)), slog.Int("总数", len(allChannels)))

		// 4. 根据策略选择一个通道
		selectedChannel, err := p.manager.selector.Select(p.ctx, healthyChannels)
		if err != nil {
			return nil, fmt.Errorf("通道选择器失败：%w", err)
		}
		p.logger.Info("已选择通道",
			slog.String("平台", selectedChannel.Platform.Name),
			slog.String("模型", selectedChannel.Model.Name))

		// 5. 获取对应的适配器并执行请求
		adapter, ok := p.manager.adapters[selectedChannel.Platform.Format]
		if !ok {
			p.logger.Error("适配器未找到", slog.String("格式", selectedChannel.Platform.Format))
			// 从通道列表中移除无效的通道
			allChannels = removeChannel(allChannels, selectedChannel)

			// 如果没有剩余通道，返回错误
			if len(allChannels) == 0 {
				return nil, fmt.Errorf("适配器未找到: %s", selectedChannel.Platform.Format)
			}

			// 更新时间戳并继续尝试其他通道
			now = time.Now()
			continue
		}

		stream, err := adapter.ChatCompletionStream(p.ctx, request, selectedChannel)
		if err == nil {
			// 6a. 成功时更新健康状态并返回
			p.logger.Info("流式聊天完成请求成功")
			p.manager.healthManager.UpdateStatusOnSuccess(selectedChannel)

			// 记录统计信息
			var firstByteTime *time.Duration
			firstByte := true

			// 包装流以记录首字节时间和最终统计
			wrappedStream := make(chan *types.Response)
			go func() {
				defer close(wrappedStream)
				defer func() {
					// 流结束时记录统计信息
					p.recordStat(&stats.RecordOptions{
						Timestamp:     p.startTime,
						RequestType:   p.requestType,
						ModelName:     request.Model,
						Duration:      time.Since(p.startTime),
						FirstByteTime: firstByteTime,
						Success:       true,
					})
				}()

				for response := range stream {
					// 记录首字时间
					if firstByte && response != nil {
						duration := time.Since(p.startTime)
						firstByteTime = &duration
						firstByte = false
					}

					// 转发响应
					wrappedStream <- response
				}
			}()

			return wrappedStream, nil
		}

		// 检查是否因为上下文取消而失败
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			p.logger.Warn("流式请求因上下文取消而中断", "error", err)
			// 不再尝试其他通道，直接返回错误
			return nil, err
		}

		// 6b. 失败时更新健康状态并准备重试
		p.logger.Warn("流式聊天完成请求失败，将尝试其他可用通道", slog.String("错误", err.Error()))
		p.manager.healthManager.UpdateStatusOnFailure(selectedChannel, err)

		// 从通道列表中移除失败的通道
		allChannels = removeChannel(allChannels, selectedChannel)

		// 如果没有剩余通道，返回错误
		if len(allChannels) == 0 {
			p.logger.Error("所有可用通道都未能处理请求")
			// 记录统计信息
			duration := time.Since(p.startTime)
			p.recordStat(&stats.RecordOptions{
				Timestamp:   p.startTime,
				RequestType: p.requestType,
				ModelName:   request.Model,
				Duration:    duration,
				Success:     false,
				ErrorMsg:    getErrorMsgFromError(err),
			})
			return nil, fmt.Errorf("没有可用的通道")
		}

		// 更新时间戳以用于下一轮健康检查
		now = time.Now()
	}
}

// removeChannel 是从切片中移除特定通道的辅助函数
//
// 该函数会在通道列表中查找指定的通道并将其移除
func removeChannel(channels []*types.Channel, toRemove *types.Channel) []*types.Channel {
	for i, ch := range channels {
		if ch == toRemove {
			return append(channels[:i], channels[i+1:]...)
		}
	}
	return channels
}
