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
	"github.com/MeowSalty/portal/channel"
	"github.com/MeowSalty/portal/health"
	"github.com/MeowSalty/portal/processor"
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
	repo               types.DataRepository
	selectorStrategy   SelectorStrategy
	healthSyncInterval time.Duration
	logger             *slog.Logger
}

// WithRepository 设置数据仓库
func WithRepository(repo types.DataRepository) Option {
	return func(o *options) {
		o.repo = repo
	}
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
	repo             types.DataRepository
	healthManager    *health.Manager
	selector         types.ChannelSelector
	channelManager   *channel.Manager
	adapters         map[string]types.Adapter // Key: Platform.Format
	adapterLock      sync.RWMutex
	logger           *slog.Logger
	statsManager     *stats.Manager
	requestProcessor *processor.RequestProcessor
	isShuttingDown   atomic.Bool
	activeSessions   sync.WaitGroup
	shutdownCtx      context.Context
	shutdownCancel   context.CancelFunc
}

// New 从配置创建并初始化一个新的 GatewayManager
//
// 该函数会初始化所有适配器并设置日志记录器
func New(ctx context.Context, opts ...Option) (*GatewayManager, error) {
	// 应用选项
	opt := &options{
		selectorStrategy:   RandomSelectorStrategy,
		healthSyncInterval: time.Minute,
		logger:             slog.Default(),
	}

	for _, o := range opts {
		o(opt)
	}

	if opt.repo == nil {
		return nil, errors.New("存储库是必需的")
	}

	adapterLogger := opt.logger.WithGroup("adapter")

	// 使用 adapter 包中的注册机制初始化适配器
	adapters := adapter.NewAdapterRegistry(adapterLogger)

	// 初始化健康状态管理器
	healthManager, err := health.New(ctx, opt.repo, opt.logger, opt.healthSyncInterval)
	if err != nil {
		return nil, fmt.Errorf("创建健康状态管理器失败：%w", err)
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
	statsManager := stats.NewManager(opt.repo, opt.logger)

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	channelManager := channel.NewManager(opt.repo, opt.logger)

	gm := &GatewayManager{
		repo:           opt.repo,
		healthManager:  healthManager,
		selector:       sel,
		channelManager: channelManager,
		adapters:       adapters,
		logger:         opt.logger,
		statsManager:   statsManager,
		shutdownCtx:    shutdownCtx,
		shutdownCancel: shutdownCancel,
	}

	// 初始化请求处理器
	gm.requestProcessor = processor.NewRequestProcessor(
		healthManager,
		sel,
		adapters,
		opt.logger,
		statsManager,
		channelManager.BuildChannels,
		gm.FindModelsByName,
	)

	return gm, nil
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

// ChatCompletion 处理聊天完成请求
//
// 该方法会根据请求的模型名称查找匹配的模型，选择可用通道，并执行聊天完成请求
func (m *GatewayManager) ChatCompletion(ctx context.Context, request *types.Request) (*types.Response, error) {
	var response *types.Response
	err := m.withSession(ctx, func(reqCtx context.Context) (err error) {
		response, err = m.requestProcessor.ProcessChatCompletion(reqCtx, request)
		m.activeSessions.Done()
		return
	})
	return response, err
}

// ChatCompletionStream 处理流式聊天完成请求
//
// 该方法会根据请求的模型名称查找匹配的模型，选择可用通道，并执行流式聊天完成请求
func (m *GatewayManager) ChatCompletionStream(ctx context.Context, request *types.Request) (<-chan *types.Response, error) {
	var stream <-chan *types.Response
	err := m.withSession(ctx, func(reqCtx context.Context) (err error) {
		stream, err = m.requestProcessor.ProcessChatCompletionStream(reqCtx, request, m.activeSessions.Done)
		return
	})
	return stream, err
}

// GetHealthStatus 获取特定资源的健康状态
//
// 该方法根据资源类型和资源 ID 返回对应的健康状态信息
func (m *GatewayManager) GetHealthStatus(resourceType types.ResourceType, resourceID uint) *types.Health {
	return m.healthManager.GetStatus(resourceType, resourceID)
}

// withSession 管理单个请求会话的生命周期
//
// 该函数负责在处理请求时正确管理会话计数和上下文取消。
// 它会在网关关闭时拒绝新请求，并确保在处理过程中正确处理上下文取消。
//
// 参数：
//
//	ctx - 父级上下文
//	fn - 在会话上下文中执行的函数
//
// 返回值：
//
//	error - 执行过程中发生的错误，如果服务正在关闭则返回 ErrServerShuttingDown
func (m *GatewayManager) withSession(ctx context.Context, fn func(reqCtx context.Context) error) error {
	// 检查服务是否正在关闭，如果是则拒绝新请求
	if m.isShuttingDown.Load() {
		return ErrServerShuttingDown
	}
	m.activeSessions.Add(1)

	// 创建可取消的请求上下文
	reqCtx, reqCancel := context.WithCancel(ctx)
	// defer reqCancel()

	// 启动一个 goroutine 监听关闭信号和上下文完成信号
	go func() {
		select {
		case <-m.shutdownCtx.Done():
			reqCancel()
		case <-reqCtx.Done():
		}
	}()

	err := fn(reqCtx)
	if err != nil {
		m.activeSessions.Done()
	}
	return err
}
