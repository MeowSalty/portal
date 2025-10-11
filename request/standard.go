package request

import (
	"context"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/routing"
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
// func (p *Request) ChatCompletion(ctx context.Context, request *types.Request, channel *types.Channel) (*types.Response, error) {
// 	requestStart := time.Now()
// 	// allChannels, err := p.prepareChannels(ctx, request.Model)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// 创建重试上下文
// 	// retryCtx := p.createRetryContext(allChannels, requestStart)

// 	// 循环重试直到成功或没有可用通道
// 	// for len(retryCtx.allChannels) > 0 {
// 	// selectedChannel, err := p.selectChannel(ctx, retryCtx.allChannels, time.Now())
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// 尝试执行请求
// 	response, err := p.executeRequest(ctx, request, channel, requestStart)
// 	if err == nil {
// 		return response, nil
// 	}

// 	// // 请求失败，移除通道并继续
// 	// retryCtx.allChannels = p.removeChannel(retryCtx, selectedChannel)

// 	// // 如果没有剩余通道，返回错误
// 	// if len(retryCtx.allChannels) == 0 {
// 	// 	return nil, fmt.Errorf("所有通道都已尝试且失败：%w", err)
// 	// }
// 	// }

// 	return nil, errors.New("没有可用通道")
// }

// ChatCompletion 处理聊天完成请求
func (p *Request) ChatCompletion(
	ctx context.Context,
	request *types.Request,
	channel *routing.Channel,
) (*types.Response, error) {
	now := time.Now()
	// 获取适配器
	adapter, err := p.getAdapter(channel.PlatformType)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeAdapterNotFound, "获取适配器失败", err).
			WithContext("format", channel.PlatformType)
	}

	// 创建请求日志
	requestLog := &RequestLog{
		Timestamp:   now,
		RequestType: "non-stream",
		ModelName:   request.Model,
		ChannelInfo: ChannelInfo{
			PlatformID: channel.PlatformID,
			APIKeyID:   channel.APIKeyID,
			ModelID:    channel.ModelID,
		},
	}

	// 执行请求
	response, err := adapter.ChatCompletion(ctx, request, channel)

	// 计算耗时
	requestDuration := time.Since(now)
	requestLog.Duration = requestDuration

	if err != nil {
		// 记录失败统计
		errorMsg := err.Error()
		requestLog.Success = false
		requestLog.ErrorMsg = &errorMsg
		p.recordRequestLog(requestLog, now, nil, false, &errorMsg)

		return nil, err
	}

	// 记录成功统计
	requestLog.Success = true
	p.recordRequestLog(requestLog, now, nil, true, nil)

	return response, nil
}
