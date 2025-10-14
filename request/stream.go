package request

import (
	"context"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/routing"
	"github.com/MeowSalty/portal/types"
	"github.com/valyala/fasthttp"
)

// ChatCompletionStream 处理流式聊天完成请求
//
// 该方法负责处理单个通道的流式请求，包括：
// - 获取并验证适配器
// - 创建请求日志
// - 初始化流连接
// - 处理流数据
func (p *Request) ChatCompletionStream(
	ctx context.Context,
	request *types.Request,
	output chan<- *types.Response,
	channel *routing.Channel,
) error {
	// 获取适配器
	adapter, err := p.getAdapter(channel.PlatformType)
	if err != nil {
		return errors.Wrap(errors.ErrCodeAdapterNotFound, "获取适配器失败", err).
			WithHTTPStatus(fasthttp.StatusInternalServerError).
			WithContext("format", channel.PlatformType)
	}

	// 创建请求日志
	now := time.Now()
	requestLog := &RequestLog{
		Timestamp:         now,
		RequestType:       "stream",
		ModelName:         channel.ModelName,
		OriginalModelName: request.Model,
		ChannelInfo: ChannelInfo{
			PlatformID: channel.PlatformID,
			APIKeyID:   channel.APIKeyID,
			ModelID:    channel.ModelID,
		},
	}

	// 创建内部流
	internalStream := make(chan *types.Response, 1024)
	err = adapter.ChatCompletionStream(ctx, request, channel, internalStream)
	if err != nil {
		errorMsg := err.(*errors.Error).Error()
		requestLog.ErrorMsg = &errorMsg
		p.recordRequestLog(requestLog, nil, false)
		return err
	}

	// 处理流数据
	return p.handleStreamData(ctx, internalStream, output, requestLog)
}

// handleStreamData 处理流数据
func (p *Request) handleStreamData(
	ctx context.Context,
	input <-chan *types.Response,
	output chan<- *types.Response,
	requestLog *RequestLog,
) error {
	firstByteRecorded := false
	var firstByteTime *time.Time
	for response := range input {
		// 检查错误
		if err := p.checkResponseError(response); err != nil {
			msg := err.Error()
			requestLog.ErrorMsg = &msg
			p.recordRequestLog(requestLog, nil, false)
			return err
		}

		// 记录首字节时间
		if !firstByteRecorded {
			now := time.Now()
			firstByteTime = &now
			firstByteRecorded = true
		}

		// 发送响应
		if err := p.sendResponse(ctx, output, response, requestLog, firstByteTime); err != nil {
			return err
		}

	}

	// 流成功完成
	close(output)
	p.recordRequestLog(requestLog, firstByteTime, true)
	return nil
}

// sendResponse 发送响应到输出通道
func (p *Request) sendResponse(
	ctx context.Context,
	output chan<- *types.Response,
	response *types.Response,
	requestLog *RequestLog,
	firstByteTime *time.Time,
) error {
	select {
	case output <- response:
		return nil
	case <-ctx.Done():
		err := errors.Wrap(errors.ErrCodeAborted, "连接被终止", ctx.Err())
		msg := err.Error()
		requestLog.ErrorMsg = &msg
		p.recordRequestLog(requestLog, firstByteTime, true)
		return err
	}
}
