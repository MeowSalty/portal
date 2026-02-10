package request

import (
	"context"
	"time"

	"github.com/MeowSalty/portal/errors"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/routing"
)

// RawAnthropicMessages 处理 Anthropic Messages 原生请求（非流式）
//
// 该方法直接发送原生请求到 Anthropic Messages API，复用日志和统计功能。
// 请求体和响应体均为 Anthropic 原生类型。
//
// 参数：
//   - ctx: 上下文
//   - req: Anthropic Messages 原生请求对象
//   - channel: 通道信息
//
// 返回：
//   - *anthropicTypes.Response: Anthropic Messages 原生响应对象
//   - error: 请求失败时返回错误
func (p *Request) RawAnthropicMessages(
	ctx context.Context,
	req *anthropicTypes.Request,
	channel *routing.Channel,
) (*anthropicTypes.Response, error) {
	now := time.Now()

	// 创建带有请求上下文的日志记录器
	log := p.logger.With(
		"platform_type", channel.Provider,
		"platform_id", channel.PlatformID,
		"model_id", channel.ModelID,
		"api_key_id", channel.APIKeyID,
		"model_name", channel.ModelName,
		"original_model", req.Model,
	)

	log.DebugContext(ctx, "开始处理 Anthropic Messages 原生请求")

	// 获取适配器
	log.DebugContext(ctx, "获取适配器", "format", channel.Provider)
	adapter, err := p.getAdapter(channel.Provider)
	if err != nil {
		log.ErrorContext(ctx, "获取适配器失败", "error", err, "format", channel.Provider)
		return nil, errors.Wrap(errors.ErrCodeAdapterNotFound, "获取适配器失败", err).
			WithContext("format", channel.Provider)
	}
	log.DebugContext(ctx, "获取适配器成功", "adapter", adapter.Name())

	// 创建请求日志
	requestLog := &RequestLog{
		Timestamp:         now,
		RequestType:       "non-stream-raw",
		ModelName:         channel.ModelName,
		OriginalModelName: req.Model,
		PlatformID:        channel.PlatformID,
		APIKeyID:          channel.APIKeyID,
		ModelID:           channel.ModelID,
	}
	log.DebugContext(ctx, "创建请求日志")

	// 执行原生请求
	log.DebugContext(ctx, "执行 Anthropic Messages 原生请求")
	response, err := adapter.RawAnthropicMessages(ctx, channel, nil, req)

	// 计算耗时
	requestDuration := time.Since(now)
	requestLog.Duration = requestDuration

	log.DebugContext(ctx, "请求完成",
		"duration", requestDuration.String(),
		"success", err == nil,
	)

	if err != nil {
		// 记录失败统计
		errorMsg := err.Error()
		requestLog.Success = false
		requestLog.ErrorMsg = &errorMsg
		p.recordRequestLog(requestLog, nil, false)

		log.ErrorContext(ctx, "Anthropic Messages 原生请求失败", "error", err)
		return nil, err
	}

	// 记录 Token 用量
	if response.Usage != nil {
		requestLog.PromptTokens = response.Usage.InputTokens
		requestLog.CompletionTokens = response.Usage.OutputTokens
		// Anthropic Usage 没有 TotalTokens 字段，需要计算
		if response.Usage.InputTokens != nil && response.Usage.OutputTokens != nil {
			total := *response.Usage.InputTokens + *response.Usage.OutputTokens
			requestLog.TotalTokens = &total
		}

		log.DebugContext(ctx, "记录 Token 使用情况",
			"prompt_tokens", response.Usage.InputTokens,
			"completion_tokens", response.Usage.OutputTokens,
			"total_tokens", requestLog.TotalTokens,
		)
	}

	// 记录成功统计
	requestLog.Success = true
	p.recordRequestLog(requestLog, nil, true)

	log.InfoContext(ctx, "Anthropic Messages 原生请求成功完成")
	return response, nil
}

// RawAnthropicMessagesStream 处理 Anthropic Messages 原生流式请求
//
// 该方法直接发送原生请求到 Anthropic Messages API，复用日志和统计功能。
// 请求体为 Anthropic Messages 原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - req: Anthropic Messages 原生请求对象
//   - output: 原生流事件输出通道
//   - channel: 通道信息
//
// 返回：
//   - error: 请求失败时返回错误
func (p *Request) RawAnthropicMessagesStream(
	ctx context.Context,
	req *anthropicTypes.Request,
	output chan<- *anthropicTypes.StreamEvent,
	channel *routing.Channel,
) error {
	// 创建带有请求上下文的日志记录器
	log := p.logger.With(
		"platform_type", channel.Provider,
		"platform_id", channel.PlatformID,
		"model_id", channel.ModelID,
		"api_key_id", channel.APIKeyID,
		"model_name", channel.ModelName,
		"original_model", req.Model,
	)

	log.DebugContext(ctx, "开始处理 Anthropic Messages 原生流式请求")

	// 获取适配器
	log.DebugContext(ctx, "获取适配器", "format", channel.Provider)
	adapter, err := p.getAdapter(channel.Provider)
	if err != nil {
		log.ErrorContext(ctx, "获取适配器失败", "error", err, "format", channel.Provider)
		return errors.Wrap(errors.ErrCodeAdapterNotFound, "获取适配器失败", err).
			WithContext("format", channel.Provider)
	}
	log.DebugContext(ctx, "获取适配器成功", "adapter", adapter.Name())

	// 创建请求日志
	now := time.Now()
	requestLog := &RequestLog{
		Timestamp:         now,
		RequestType:       "stream-raw",
		ModelName:         channel.ModelName,
		OriginalModelName: req.Model,
		PlatformID:        channel.PlatformID,
		APIKeyID:          channel.APIKeyID,
		ModelID:           channel.ModelID,
	}
	log.DebugContext(ctx, "创建请求日志")

	// 创建内部流
	log.DebugContext(ctx, "创建内部流通道")
	internalStream := make(chan *anthropicTypes.StreamEvent, 1024)

	log.DebugContext(ctx, "执行 Anthropic Messages 原生流式请求")
	err = adapter.RawAnthropicMessagesStream(ctx, channel, nil, req, internalStream)
	if err != nil {
		errorMsg := err.Error()
		requestLog.ErrorMsg = &errorMsg
		p.recordRequestLog(requestLog, nil, false)

		log.ErrorContext(ctx, "Anthropic Messages 原生流式请求失败", "error", err)
		return err
	}

	// 处理流数据
	log.DebugContext(ctx, "开始处理流数据")
	return p.handleAnthropicRawStreamData(ctx, internalStream, output, requestLog)
}

// handleAnthropicRawStreamData 处理 Anthropic 原生流数据
func (p *Request) handleAnthropicRawStreamData(
	ctx context.Context,
	input <-chan *anthropicTypes.StreamEvent,
	output chan<- *anthropicTypes.StreamEvent,
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

		// 记录 Token 用量（在 message_delta 事件中）
		if response.MessageDelta != nil && response.MessageDelta.Usage != nil {
			requestLog.CompletionTokens = response.MessageDelta.Usage.OutputTokens
			requestLog.PromptTokens = response.MessageDelta.Usage.InputTokens
			// Anthropic 没有 TotalTokens，需要计算
			if response.MessageDelta.Usage.InputTokens != nil && response.MessageDelta.Usage.OutputTokens != nil {
				total := *response.MessageDelta.Usage.InputTokens + *response.MessageDelta.Usage.OutputTokens
				requestLog.TotalTokens = &total
			}

			log.DebugContext(ctx, "更新 Token 使用情况",
				"prompt_tokens", response.MessageDelta.Usage.InputTokens,
				"completion_tokens", response.MessageDelta.Usage.OutputTokens,
				"total_tokens", requestLog.TotalTokens,
			)
		}

		// 发送响应
		if err := p.sendAnthropicRawResponse(ctx, output, response, requestLog, firstByteTime); err != nil {
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

// sendAnthropicRawResponse 发送 Anthropic 原生响应到输出通道
func (p *Request) sendAnthropicRawResponse(
	ctx context.Context,
	output chan<- *anthropicTypes.StreamEvent,
	response *anthropicTypes.StreamEvent,
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
