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
	// 创建带有请求上下文的日志记录器
	log := p.logger.With(
		"platform_type", channel.PlatformType,
		"platform_id", channel.PlatformID,
		"model_id", channel.ModelID,
		"api_key_id", channel.APIKeyID,
		"model_name", channel.ModelName,
		"original_model", request.Model,
	)

	log.DebugContext(ctx, "开始处理流式聊天完成请求")

	// 获取适配器
	log.DebugContext(ctx, "获取适配器", "format", channel.PlatformType)
	adapter, err := p.getAdapter(channel.PlatformType)
	if err != nil {
		log.ErrorContext(ctx, "获取适配器失败", "error", err, "format", channel.PlatformType)
		return errors.Wrap(errors.ErrCodeAdapterNotFound, "获取适配器失败", err).
			WithHTTPStatus(fasthttp.StatusInternalServerError).
			WithContext("format", channel.PlatformType)
	}
	log.DebugContext(ctx, "获取适配器成功", "adapter", adapter.Name())

	// 创建请求日志
	now := time.Now()
	requestLog := &RequestLog{
		Timestamp:         now,
		RequestType:       "stream",
		ModelName:         channel.ModelName,
		OriginalModelName: request.Model,
		PlatformID:        channel.PlatformID,
		APIKeyID:          channel.APIKeyID,
		ModelID:           channel.ModelID,
	}
	log.DebugContext(ctx, "创建请求日志")

	// 创建内部流
	log.DebugContext(ctx, "创建内部流通道")
	internalStream := make(chan *types.Response, 1024)

	log.DebugContext(ctx, "执行流式聊天完成请求")
	err = adapter.ChatCompletionStream(ctx, request, channel, internalStream)
	if err != nil {
		errorMsg := err.(*errors.Error).Error()
		requestLog.ErrorMsg = &errorMsg
		p.recordRequestLog(requestLog, nil, false)

		log.ErrorContext(ctx, "流式聊天完成请求失败", "error", err)
		return err
	}

	// 处理流数据
	log.DebugContext(ctx, "开始处理流数据")
	return p.handleStreamData(ctx, internalStream, output, requestLog)
}

// handleStreamData 处理流数据
func (p *Request) handleStreamData(
	ctx context.Context,
	input <-chan *types.Response,
	output chan<- *types.Response,
	requestLog *RequestLog,
) error {
	log := p.logger.With(
		"platform_id", requestLog.PlatformID,
		"model_id", requestLog.ModelID,
		"api_key_id", requestLog.APIKeyID,
	)

	log.DebugContext(ctx, "开始处理流数据")

	firstByteRecorded := false
	var firstByteTime *time.Time
	messageCount := 0

	for response := range input {
		messageCount++

		// 检查错误
		if err := p.checkResponseError(response); err != nil {
			var portalError *errors.Error
			if errors.As(err, &portalError) {
				switch errors.GetCode(portalError) {
				case errors.ErrCodeEmptyResponse:
					log.ErrorContext(ctx, "响应中 Choices 为空", "error", portalError.Error())
				case errors.ErrCodeStreamError:
					log.ErrorContext(ctx, "流处理错误", "error", portalError.Error())
				default:
					// 兜底：处理其他错误码
					log.ErrorContext(ctx, "流数据错误", "error", portalError.Error(), "code", portalError.Code)
				}
			}

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

			firstByteDuration := now.Sub(requestLog.Timestamp)
			log.DebugContext(ctx, "收到首字节",
				"first_byte_time", firstByteDuration.String(),
			)
		}

		// 记录 Token 用量
		if response.Usage != nil {
			requestLog.CompletionTokens = &response.Usage.CompletionTokens
			requestLog.PromptTokens = &response.Usage.PromptTokens
			requestLog.TotalTokens = &response.Usage.TotalTokens

			log.DebugContext(ctx, "更新 Token 使用情况",
				"prompt_tokens", response.Usage.PromptTokens,
				"completion_tokens", response.Usage.CompletionTokens,
				"total_tokens", response.Usage.TotalTokens,
			)
		}

		// 发送响应
		if err := p.sendResponse(ctx, output, response, requestLog, firstByteTime); err != nil {
			log.ErrorContext(ctx, "发送响应失败",
				"error", err,
				"message_count", messageCount,
			)
			return err
		}

	}

	// 流成功完成
	log.DebugContext(ctx, "流数据处理完成",
		"total_messages", messageCount,
		"duration", time.Since(requestLog.Timestamp).String(),
	)

	close(output)
	p.recordRequestLog(requestLog, firstByteTime, true)

	log.InfoContext(ctx, "流式请求成功完成")
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
	log := p.logger.With(
		"platform_id", requestLog.PlatformID,
		"model_id", requestLog.ModelID,
		"api_key_id", requestLog.APIKeyID,
	)

	select {
	case output <- response:
		log.DebugContext(ctx, "响应发送成功")
		return nil
	case <-ctx.Done():
		err := errors.Wrap(errors.ErrCodeAborted, "连接被终止", ctx.Err())
		msg := err.Error()
		requestLog.ErrorMsg = &msg
		p.recordRequestLog(requestLog, firstByteTime, true)

		log.WarnContext(ctx, "连接被终止", "error", ctx.Err())
		return err
	}
}
