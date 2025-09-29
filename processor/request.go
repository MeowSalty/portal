package processor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/MeowSalty/portal/types"
)

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

	// 创建重试上下文
	retryCtx := p.createRetryContext(allChannels, requestStart)

	// 循环重试直到成功或没有可用通道
	for len(retryCtx.allChannels) > 0 {
		selectedChannel, err := p.selectChannel(ctx, retryCtx.allChannels, time.Now())
		if err != nil {
			return nil, err
		}

		// 尝试执行请求
		response, err := p.executeRequest(ctx, request, selectedChannel, requestStart)
		if err == nil {
			return response, nil
		}

		// 请求失败，移除通道并继续
		retryCtx.allChannels = p.removeChannel(retryCtx, selectedChannel)

		// 如果没有剩余通道，返回错误
		if len(retryCtx.allChannels) == 0 {
			return nil, fmt.Errorf("所有通道都已尝试且失败：%w", err)
		}
	}

	return nil, errors.New("没有可用通道")
}

// executeRequest 执行单个请求
func (p *RequestProcessor) executeRequest(
	ctx context.Context,
	request *types.Request,
	selectedChannel *types.Channel,
	requestStart time.Time,
) (*types.Response, error) {
	// 获取适配器
	adapter, err := p.getAdapter(selectedChannel.Platform.Format)
	if err != nil {
		p.Logger.Error("适配器未找到", slog.String("格式", selectedChannel.Platform.Format))
		return nil, err
	}

	// 创建统计信息
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

	// 执行请求
	response, err := adapter.ChatCompletion(ctx, request, selectedChannel)

	// 计算耗时
	requestDuration := time.Since(requestStart)
	statOptions.Duration = requestDuration

	if err != nil {
		err = sanitizeError(err)
		p.Logger.Error("请求执行失败",
			slog.String("model", request.Model),
			slog.String("platform", selectedChannel.Platform.Name),
			slog.Any("error", err))

		// 更新健康状态
		p.updateHealthStatus(selectedChannel, false, err.Error())

		// 记录失败统计
		errorMsg := err.Error()
		statOptions.Success = false
		statOptions.ErrorMsg = &errorMsg
		p.recordStatsSafe(ctx, statOptions)

		return nil, err
	}

	// 请求成功
	p.updateHealthStatus(selectedChannel, true, "")

	// 记录成功统计
	statOptions.Success = true
	p.recordStatsSafe(ctx, statOptions)

	return response, nil
}

// prepareChannels 准备通道
//
// 该方法实现了通道准备流程：
//
// 1. 根据模型名称查找所有匹配的模型
// 2. 为这些模型构建所有可能的通道
func (p *RequestProcessor) prepareChannels(ctx context.Context, modelName string) ([]*types.Channel, error) {
	// 1. 根据模型名称查找所有匹配的模型
	models, err := p.FindModelsByName(ctx, modelName)
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
