package portal

import (
	"context"

	"github.com/MeowSalty/portal/errors"
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/routing"
)

// RawOpenAIChatCompletion 执行 OpenAI Chat 原生请求（非流式）
//
// 该方法直接发送原生请求到 OpenAI Chat Completions API，不经过 middleware 与统一 contract。
// 请求体和响应体均为 OpenAI Chat 原生类型。
//
// 参数：
//   - ctx: 上下文
//   - req: OpenAI Chat 原生请求对象
//
// 返回：
//   - *openaiChat.Response: OpenAI Chat 原生响应对象
//   - error: 请求失败时返回错误
func (p *Portal) RawOpenAIChatCompletion(
	ctx context.Context,
	req *openaiChat.Request,
) (*openaiChat.Response, error) {
	p.logger.DebugContext(ctx, "开始处理 OpenAI Chat 原生请求", "model", req.Model)

	var response *openaiChat.Response
	var channel *routing.Channel
	var err error
	for {
		channel, err = p.routing.GetChannelByProvider(ctx, req.Model, "openai", "chat_completions")
		if err != nil {
			p.logger.ErrorContext(ctx, "获取通道失败", "model", req.Model, "error", err)
			break
		}

		// 使用 With 创建带有通道上下文的日志记录器
		channelLogger := p.logger.With(
			"platform_id", channel.PlatformID,
			"model_id", channel.ModelID,
			"api_key_id", channel.APIKeyID)

		channelLogger.DebugContext(ctx, "获取到通道")

		err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) (err error) {
			defer reqCancel()
			response, err = p.request.RawOpenAIChatCompletion(reqCtx, req, channel)
			return
		})

		// 检查错误是否可以重试
		if err != nil {
			if errors.IsRetryable(err) {
				channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
				channel.MarkFailure(ctx, err)
				continue
			}
			// 特殊处理：操作终止时标记成功
			if errors.IsCode(err, errors.ErrCodeAborted) {
				channelLogger.InfoContext(ctx, "操作终止")
				channel.MarkSuccess(ctx)
			}
			channelLogger.ErrorContext(ctx, "请求处理失败", "error", err)
			channel.MarkFailure(ctx, err)
			break
		}
		channel.MarkSuccess(ctx)
		channelLogger.InfoContext(ctx, "请求处理成功")
		break
	}

	return response, err
}

// RawOpenAIChatCompletionStream 执行 OpenAI Chat 原生流式请求
//
// 该方法直接发送原生请求到 OpenAI Chat Completions API，不经过 middleware 与统一 contract。
// 请求体为 OpenAI Chat 原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - req: OpenAI Chat 原生请求对象
//
// 返回：
//   - <-chan *openaiChat.StreamEvent: 原生流事件通道
func (p *Portal) RawOpenAIChatCompletionStream(
	ctx context.Context,
	req *openaiChat.Request,
) <-chan *openaiChat.StreamEvent {
	p.logger.DebugContext(ctx, "开始处理 OpenAI Chat 原生流式请求", "model", req.Model)

	// 创建内部流（用于接收原始响应）
	internalStream := make(chan *openaiChat.StreamEvent, 1024)

	// 启动内部流处理协程
	go func() {
		for {
			channel, err := p.routing.GetChannelByProvider(ctx, req.Model, "openai", "chat_completions")
			if err != nil {
				p.logger.ErrorContext(ctx, "获取通道失败", "model", req.Model, "error", err)
				close(internalStream)
				break
			}

			// 使用 With 创建带有通道上下文的日志记录器
			channelLogger := p.logger.With(
				"platform_id", channel.PlatformID,
				"model_id", channel.ModelID,
				"api_key_id", channel.APIKeyID)

			channelLogger.DebugContext(ctx, "获取到通道")

			err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) (err error) {
				defer reqCancel()
				return p.request.RawOpenAIChatCompletionStream(reqCtx, req, internalStream, channel)
			})

			// 检查错误是否可以重试
			if err != nil {
				if errors.IsRetryable(err) {
					channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
					channel.MarkFailure(ctx, err)
					continue
				}
				// 特殊处理：操作终止时标记成功
				if errors.IsCode(err, errors.ErrCodeAborted) {
					channelLogger.InfoContext(ctx, "操作终止")
					channel.MarkSuccess(ctx)
				}
				channelLogger.ErrorContext(ctx, "流处理失败", "error", err)
				channel.MarkFailure(ctx, err)
				close(internalStream)
				break
			}
			channel.MarkSuccess(ctx)
			channelLogger.InfoContext(ctx, "流处理成功")
			break
		}
	}()

	return internalStream
}

// RawOpenAIResponses 执行 OpenAI Responses 原生请求（非流式）
//
// 该方法直接发送原生请求到 OpenAI Responses API，不经过 middleware 与统一 contract。
// 请求体和响应体均为 OpenAI Responses 原生类型。
//
// 参数：
//   - ctx: 上下文
//   - req: OpenAI Responses 原生请求对象
//
// 返回：
//   - *openaiResponses.Response: OpenAI Responses 原生响应对象
//   - error: 请求失败时返回错误
func (p *Portal) RawOpenAIResponses(
	ctx context.Context,
	req *openaiResponses.Request,
) (*openaiResponses.Response, error) {
	// 获取模型名称
	modelName := ""
	if req.Model != nil {
		modelName = *req.Model
	}

	p.logger.DebugContext(ctx, "开始处理 OpenAI Responses 原生请求", "model", modelName)

	var response *openaiResponses.Response
	var channel *routing.Channel
	var err error
	for {
		channel, err = p.routing.GetChannelByProvider(ctx, modelName, "openai", "responses")
		if err != nil {
			p.logger.ErrorContext(ctx, "获取通道失败", "model", modelName, "error", err)
			break
		}

		// 使用 With 创建带有通道上下文的日志记录器
		channelLogger := p.logger.With(
			"platform_id", channel.PlatformID,
			"model_id", channel.ModelID,
			"api_key_id", channel.APIKeyID)

		channelLogger.DebugContext(ctx, "获取到通道")

		err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) (err error) {
			defer reqCancel()
			response, err = p.request.RawOpenAIResponses(reqCtx, req, channel)
			return
		})

		// 检查错误是否可以重试
		if err != nil {
			if errors.IsRetryable(err) {
				channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
				channel.MarkFailure(ctx, err)
				continue
			}
			// 特殊处理：操作终止时标记成功
			if errors.IsCode(err, errors.ErrCodeAborted) {
				channelLogger.InfoContext(ctx, "操作终止")
				channel.MarkSuccess(ctx)
			}
			channelLogger.ErrorContext(ctx, "请求处理失败", "error", err)
			channel.MarkFailure(ctx, err)
			break
		}
		channel.MarkSuccess(ctx)
		channelLogger.InfoContext(ctx, "请求处理成功")
		break
	}

	return response, err
}

// RawOpenAIResponsesStream 执行 OpenAI Responses 原生流式请求
//
// 该方法直接发送原生请求到 OpenAI Responses API，不经过 middleware 与统一 contract。
// 请求体为 OpenAI Responses 原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - req: OpenAI Responses 原生请求对象
//
// 返回：
//   - <-chan *openaiResponses.StreamEvent: 原生流事件通道
func (p *Portal) RawOpenAIResponsesStream(
	ctx context.Context,
	req *openaiResponses.Request,
) <-chan *openaiResponses.StreamEvent {
	// 获取模型名称
	modelName := ""
	if req.Model != nil {
		modelName = *req.Model
	}

	p.logger.DebugContext(ctx, "开始处理 OpenAI Responses 原生流式请求", "model", modelName)

	// 创建内部流（用于接收原始响应）
	internalStream := make(chan *openaiResponses.StreamEvent, 1024)

	// 启动内部流处理协程
	go func() {
		for {
			channel, err := p.routing.GetChannelByProvider(ctx, modelName, "openai", "responses")
			if err != nil {
				p.logger.ErrorContext(ctx, "获取通道失败", "model", modelName, "error", err)
				close(internalStream)
				break
			}

			// 使用 With 创建带有通道上下文的日志记录器
			channelLogger := p.logger.With(
				"platform_id", channel.PlatformID,
				"model_id", channel.ModelID,
				"api_key_id", channel.APIKeyID)

			channelLogger.DebugContext(ctx, "获取到通道")

			err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) (err error) {
				defer reqCancel()
				return p.request.RawOpenAIResponsesStream(reqCtx, req, internalStream, channel)
			})

			// 检查错误是否可以重试
			if err != nil {
				if errors.IsRetryable(err) {
					channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
					channel.MarkFailure(ctx, err)
					continue
				}
				// 特殊处理：操作终止时标记成功
				if errors.IsCode(err, errors.ErrCodeAborted) {
					channelLogger.InfoContext(ctx, "操作终止")
					channel.MarkSuccess(ctx)
				}
				channelLogger.ErrorContext(ctx, "流处理失败", "error", err)
				channel.MarkFailure(ctx, err)
				close(internalStream)
				break
			}
			channel.MarkSuccess(ctx)
			channelLogger.InfoContext(ctx, "流处理成功")
			break
		}
	}()

	return internalStream
}
