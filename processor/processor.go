package processor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/MeowSalty/portal/health"
	"github.com/MeowSalty/portal/stats"
	"github.com/MeowSalty/portal/types"
	"github.com/valyala/fasthttp"
)

// RequestProcessor 处理请求的结构体
//
// 该结构体封装了处理各种 AI 请求的通用逻辑和上下文信息
type RequestProcessor struct {
	HealthManager *health.Manager
	Selector      types.ChannelSelector
	Adapters      map[string]types.Adapter
	Logger        *slog.Logger
	StatsManager  *stats.Manager
	BuildChannels func(context.Context, []*types.Model) ([]*types.Channel, error)
	FindModels    func(context.Context, string) ([]*types.Model, error)
}

// NewRequestProcessor 创建一个新的请求处理器
func NewRequestProcessor(
	healthManager *health.Manager,
	selector types.ChannelSelector,
	adapters map[string]types.Adapter,
	logger *slog.Logger,
	statsManager *stats.Manager,
	buildChannels func(context.Context, []*types.Model) ([]*types.Channel, error),
	findModels func(context.Context, string) ([]*types.Model, error),
) *RequestProcessor {
	return &RequestProcessor{
		HealthManager: healthManager,
		Selector:      selector,
		Adapters:      adapters,
		Logger:        logger,
		StatsManager:  statsManager,
		BuildChannels: buildChannels,
		FindModels:    findModels,
	}
}

// ProcessChatCompletion 处理聊天完成请求
//
// 该方法实现了完整的请求处理流程：
//
//  1. 查找匹配的模型
//  2. 构建所有可能的通道
//  3. 过滤出健康的通道
//  4. 根据策略选择一个通道
//  5. 获取对应的适配器并执行请求
//  6. 根据执行结果更新健康状态并返回结果
func (p *RequestProcessor) ProcessChatCompletion(ctx context.Context, request *types.Request) (*types.Response, error) {
	requestStart := time.Now()
	allChannels, err := p.prepareChannels(ctx, request.Model)
	if err != nil {
		return nil, err
	}

	// 循环重试直到成功或没有可用通道
	for {
		selectedChannel, err := p.selectChannel(ctx, allChannels, requestStart)
		if err != nil {
			return nil, err
		}

		// 获取对应的适配器并执行请求
		adapter, ok := p.Adapters[selectedChannel.Platform.Format]
		if !ok {
			p.Logger.Error("适配器未找到", slog.String("格式", selectedChannel.Platform.Format))
			// 从通道列表中移除无效的通道
			allChannels = removeChannel(allChannels, selectedChannel)

			// 如果没有剩余通道，返回错误
			if len(allChannels) == 0 {
				return nil, fmt.Errorf("适配器未找到：%s", selectedChannel.Platform.Format)
			}

			// 更新时间戳并继续尝试其他通道
			requestStart = time.Now()
			continue
		}

		// 记录统计信息
		statOptions := &types.RequestStat{
			Timestamp:   requestStart,
			RequestType: "non-stream",
			ModelName:   request.Model,
			ChannelInfo: types.ChannelInfo{
				PlatformID: selectedChannel.Platform.ID,
				APIKeyID:   selectedChannel.APIKey.ID,
				ModelID:    selectedChannel.Model.ID,
			},
		}

		response, err := adapter.ChatCompletion(ctx, request, selectedChannel)

		// 计算耗时
		requestDuration := time.Since(requestStart)

		// 更新统计信息
		statOptions.Duration = requestDuration

		if err != nil {
			p.Logger.Error("请求执行失败",
				slog.String("model", request.Model),
				slog.String("platform", selectedChannel.Platform.Name),
				slog.Any("error", err))

			// 更新健康状态为失败
			p.HealthManager.UpdateStatus(
				types.ResourceTypeAPIKey,
				selectedChannel.APIKey.ID,
				false,
				err.Error(),
				0,
			)

			// 记录失败的统计信息
			errorMsg := err.Error()
			statOptions.Success = false
			statOptions.ErrorMsg = &errorMsg

			if recordErr := p.StatsManager.RecordRequestStat(ctx, statOptions); recordErr != nil {
				p.Logger.Error("记录统计信息失败", slog.Any("error", recordErr))
			}

			// 从通道列表中移除失败的通道
			allChannels = removeChannel(allChannels, selectedChannel)

			// 如果没有剩余通道，返回错误
			if len(allChannels) == 0 {
				return nil, fmt.Errorf("所有通道都已尝试且失败：%w", err)
			}

			// 更新时间戳并继续尝试其他通道
			requestStart = time.Now()
			continue
		}

		// 请求成功，更新健康状态
		p.HealthManager.UpdateStatus(
			types.ResourceTypeAPIKey,
			selectedChannel.APIKey.ID,
			true,
			"",
			0,
		)

		// 记录成功的统计信息
		statOptions.Success = true
		if recordErr := p.StatsManager.RecordRequestStat(ctx, statOptions); recordErr != nil {
			p.Logger.Error("记录统计信息失败", slog.Any("error", recordErr))
		}

		return response, nil
	}
}

// ProcessChatCompletionStream 处理流式聊天完成请求
//
// 该方法实现了完整的流式请求处理流程：
// 1. 查找匹配的模型
// 2. 构建所有可能的通道
// 3. 过滤出健康的通道
// 4. 根据策略选择一个通道
// 5. 获取对应的适配器并执行流式请求
// 6. 根据执行结果更新健康状态并返回结果流
func (p *RequestProcessor) ProcessChatCompletionStream(ctx context.Context, request *types.Request, doneFunc func()) (<-chan *types.Response, error) {
	startTime := time.Now()
	allChannels, err := p.prepareChannels(ctx, request.Model)
	if err != nil {
		return nil, err
	}

	// 循环重试直到成功或没有可用通道
	for {
		// 缓存当前时间，避免在 selectChannel 中多次调用 time.Now()
		now := time.Now()
		selectedChannel, err := p.selectChannel(ctx, allChannels, now)
		if err != nil {
			return nil, err
		}

		// 获取对应的适配器并执行流式请求
		adapter, ok := p.Adapters[selectedChannel.Platform.Format]
		if !ok {
			p.Logger.Error("适配器未找到", slog.String("格式", selectedChannel.Platform.Format))
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

		// 记录统计信息
		statOptions := &types.RequestStat{
			Timestamp:   startTime,
			RequestType: "stream",
			ModelName:   request.Model,
			ChannelInfo: types.ChannelInfo{
				PlatformID: selectedChannel.Platform.ID,
				APIKeyID:   selectedChannel.APIKey.ID,
				ModelID:    selectedChannel.Model.ID,
			},
		}

		// 执行请求前记录开始时间
		requestStart := time.Now()

		stream, err := adapter.ChatCompletionStream(ctx, request, selectedChannel)
		if err != nil {
			p.Logger.Error("流式请求初始化失败",
				slog.String("model", request.Model),
				slog.String("platform", selectedChannel.Platform.Name),
				slog.Any("error", err),
				slog.String("request_type", "stream"))

			// 更新健康状态为失败
			p.HealthManager.UpdateStatus(
				types.ResourceTypeAPIKey,
				selectedChannel.APIKey.ID,
				false,
				err.Error(),
				0,
			)

			// 记录失败的统计信息
			errorMsg := err.Error()
			statOptions.Success = false
			statOptions.ErrorMsg = &errorMsg

			if recordErr := p.StatsManager.RecordRequestStat(ctx, statOptions); recordErr != nil {
				p.Logger.Error("记录统计信息失败",
					slog.Any("error", recordErr),
					slog.String("operation", "record_request_stat"))
			}

			// 从通道列表中移除失败的通道
			allChannels = removeChannel(allChannels, selectedChannel)

			// 如果没有剩余通道，返回错误
			if len(allChannels) == 0 {
				return nil, fmt.Errorf("所有通道都已尝试且失败：%w", err)
			}

			// 更新时间戳并继续尝试其他通道
			now = time.Now()
			continue
		}

		// 创建包装的流，用于处理统计信息记录
		wrappedStream := make(chan *types.Response)

		go func() {
			defer close(wrappedStream)
			// 确保在流处理完成后调用 doneFunc
			defer doneFunc()

			firstByteRecorded := false
			var firstByteTime time.Time
			var streamErr error

			// 用于确保只处理一次流结束后的状态更新
			statsRecorded := false
			defer func() {
				if !statsRecorded {
					// 计算耗时
					requestDuration := time.Since(requestStart)

					// 更新统计信息
					statOptions.Duration = requestDuration

					// 如果记录了首字节时间，则计算首字节耗时
					if firstByteRecorded {
						firstByteDuration := firstByteTime.Sub(requestStart)
						statOptions.FirstByteTime = &firstByteDuration
					}

					// 内联 recordStreamStats 函数的实现
					if streamErr != nil {
						p.Logger.Error("流式传输过程中发生错误",
							slog.String("model", statOptions.ModelName),
							slog.String("platform", selectedChannel.Platform.Name),
							slog.Any("error", streamErr))

						// 更新健康状态为失败
						p.HealthManager.UpdateStatus(
							types.ResourceTypeAPIKey,
							selectedChannel.APIKey.ID,
							false,
							streamErr.Error(),
							0,
						)

						// 记录失败的统计信息
						errorMsg := streamErr.Error()
						statOptions.Success = false
						statOptions.ErrorMsg = &errorMsg
					} else {
						// 流正常结束，更新健康状态为成功
						p.HealthManager.UpdateStatus(
							types.ResourceTypeAPIKey,
							selectedChannel.APIKey.ID,
							true,
							"",
							0,
						)

						// 记录成功的统计信息
						statOptions.Success = true
					}

					if recordErr := p.StatsManager.RecordRequestStat(ctx, statOptions); recordErr != nil {
						p.Logger.Error("记录统计信息失败", slog.Any("error", recordErr))
					}
					statsRecorded = true
				}
			}()

			// 从原始流中读取数据
			for response := range stream {
				// 检查是否有错误信息
				for _, choice := range response.Choices {
					if choice.Error != nil && choice.Error.Code != fasthttp.StatusOK {
						streamErr = fmt.Errorf("stream error: code=%d, message=%s", choice.Error.Code, choice.Error.Message)
						// 记录错误日志
						p.Logger.Error("流式传输过程中发生错误",
							slog.String("model", request.Model),
							slog.String("platform", selectedChannel.Platform.Name),
							slog.Int("code", choice.Error.Code),
							slog.String("message", choice.Error.Message),
							slog.String("request_type", "stream"))

						// 发送错误信息到包装流
						wrappedStream <- response
						return
					}
				}

				// 记录第一个响应的到达时间（只记录一次）
				if !firstByteRecorded {
					firstByteTime = time.Now()
					firstByteRecorded = true
				}

				select {
				case wrappedStream <- response:
				case <-ctx.Done():
					streamErr = ctx.Err()
					p.Logger.Error("客户端断开连接",
						slog.String("model", request.Model),
						slog.String("platform", selectedChannel.Platform.Name),
						slog.Any("error", streamErr))
					return
				}
			}

			// 流正常结束，不需要单独更新健康状态
			// defer 会处理统计信息记录
		}()

		return wrappedStream, nil
	}
}

// prepareChannels 准备通道
//
// 该方法实现了通道准备流程：
//
// 1. 根据模型名称查找所有匹配的模型
// 2. 为这些模型构建所有可能的通道
func (p *RequestProcessor) prepareChannels(ctx context.Context, modelName string) ([]*types.Channel, error) {
	// 1. 根据模型名称查找所有匹配的模型
	models, err := p.FindModels(ctx, modelName)
	if err != nil {
		p.Logger.Error("查找模型失败", slog.String("模型名称", modelName), slog.Any("error", err))
		return nil, fmt.Errorf("查找模型失败：%w", err)
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("未找到模型: %s", modelName)
	}

	// 2. 为这些模型构建所有可能的通道
	channels, err := p.BuildChannels(ctx, models)
	if err != nil {
		p.Logger.Error("构建通道失败", slog.String("模型名称", modelName), slog.Any("error", err))
		return nil, fmt.Errorf("构建通道失败：%w", err)
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("未找到可用通道: %s", modelName)
	}

	return channels, nil
}

// selectChannel 选择通道
//
// 该方法实现了通道选择流程：
//
// 1. 过滤出当前健康的通道
// 2. 使用选择器从健康通道中选择一个
func (p *RequestProcessor) selectChannel(ctx context.Context, channels []*types.Channel, now time.Time) (*types.Channel, error) {
	// 过滤出当前健康的通道
	healthyChannels := p.HealthManager.FilterHealthyChannels(channels, now)
	if len(healthyChannels) == 0 {
		return nil, errors.New("没有可用的健康通道")
	}

	// 使用选择器从健康通道中选择一个
	selectedChannel, err := p.Selector.Select(ctx, healthyChannels)
	if err != nil {
		return nil, fmt.Errorf("选择通道失败：%w", err)
	}

	p.Logger.Info("通道选择成功",
		slog.String("model", selectedChannel.Model.Name),
		slog.String("platform", selectedChannel.Platform.Name))

	return selectedChannel, nil
}

// removeChannel 从通道列表中移除指定通道
func removeChannel(channels []*types.Channel, target *types.Channel) []*types.Channel {
	for i, ch := range channels {
		if ch.Platform.ID == target.Platform.ID &&
			ch.Model.ID == target.Model.ID &&
			ch.APIKey.ID == target.APIKey.ID {
			// 找到目标通道，移除它
			return append(channels[:i], channels[i+1:]...)
		}
	}
	return channels
}
