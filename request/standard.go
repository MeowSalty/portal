package request

import (
	"context"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/routing"
	"github.com/MeowSalty/portal/types"
)

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
		Timestamp:         now,
		RequestType:       "non-stream",
		ModelName:         channel.ModelName,
		OriginalModelName: request.Model,
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
		p.recordRequestLog(requestLog, nil, false)

		return nil, err
	}

	// 记录成功统计
	requestLog.Success = true
	p.recordRequestLog(requestLog, nil, true)

	return response, nil
}
