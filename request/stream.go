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
// - 创建流上下文
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

	// 创建流上下文
	now := time.Now()
	requestLog := &RequestLog{
		Timestamp:   now,
		RequestType: "stream",
		ModelName:   request.Model,
		ChannelInfo: ChannelInfo{
			PlatformID: channel.PlatformID,
			APIKeyID:   channel.APIKeyID,
			ModelID:    channel.ModelID,
		},
	}
	streamCtx := &streamContext{
		requestStart: now,
		requestLog: &RequestLog{
			Timestamp:   now,
			RequestType: "stream",
			ModelName:   request.Model,
			ChannelInfo: ChannelInfo{
				PlatformID: channel.PlatformID,
				APIKeyID:   channel.APIKeyID,
				ModelID:    channel.ModelID,
			},
		},
	}

	// 创建内部流
	internalStream := make(chan *types.Response, 1024)
	err = adapter.ChatCompletionStream(ctx, request, channel, internalStream)
	if err != nil {
		errorMsg := err.(*errors.Error).Error()
		p.recordRequestLog(requestLog, now, nil, false, &errorMsg)
		return err
	}

	// 处理流数据
	return p.handleStreamData(ctx, internalStream, output, streamCtx)
}

// handleStreamData 处理流数据
func (p *Request) handleStreamData(
	ctx context.Context,
	input <-chan *types.Response,
	output chan<- *types.Response,
	streamCtx *streamContext,
) error {
	firstByteRecorded := false

	for response := range input {
		// 检查错误
		if err := p.checkResponseError(response); err != nil {
			msg := err.Error()
			streamCtx.requestLog.ErrorMsg = &msg
			p.recordLog(streamCtx)
			return err
		}

		// 记录首字节时间
		if !firstByteRecorded {
			now := time.Now()
			streamCtx.firstByteTime = &now
			firstByteRecorded = true
		}

		// 发送响应
		if err := p.sendResponse(ctx, output, response, streamCtx); err != nil {
			return err
		}
	}

	// 流成功完成
	close(output)
	p.recordLog(streamCtx)
	return nil
}

// sendResponse 发送响应到输出通道
func (p *Request) sendResponse(
	ctx context.Context,
	output chan<- *types.Response,
	response *types.Response,
	streamCtx *streamContext,
) error {
	select {
	case output <- response:
		return nil
	case <-ctx.Done():
		err := errors.Wrap(errors.ErrCodeAborted, "连接被终止", ctx.Err())
		msg := err.Error()
		streamCtx.requestLog.ErrorMsg = &msg
		p.recordLog(streamCtx)
		return err
	}
}

// recordLog 记录请求日志
func (p *Request) recordLog(streamCtx *streamContext) {
	var errorMsg *string
	isSuccess := true
	if streamCtx.streamError != nil {
		isSuccess = false
		if portalErr, ok := streamCtx.streamError.(*errors.Error); ok {
			if portalErr.Is(errors.ErrAborted) {
				// 忽略因终止操作导致的错误
				isSuccess = true
			}
			msg := portalErr.Error()
			errorMsg = &msg
		}
	}
	p.recordRequestLog(streamCtx.requestLog, streamCtx.requestStart, streamCtx.firstByteTime, isSuccess, errorMsg)
}
