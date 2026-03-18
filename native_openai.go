package portal

import (
	"context"

	"github.com/MeowSalty/portal/errors"
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/routing"
)

// NativeOpenAIChatCompletion 执行 OpenAI Chat 原生请求（非流式）
//
// 该方法通过 routing 获取通道，使用 retry 机制，调用 request.Native。
// 请求体和响应体均为 OpenAI Chat 原生类型。
//
// 参数：
//   - ctx: 上下文
//   - req: OpenAI Chat 原生请求对象
//
// 返回：
//   - *openaiChat.Response: OpenAI Chat 原生响应对象
//   - error: 请求失败时返回错误
func (p *Portal) NativeOpenAIChatCompletion(
	ctx context.Context,
	req *openaiChat.Request,
	opts ...NativeOption,
) (*openaiChat.Response, error) {
	p.logger.DebugContext(ctx, "开始处理 OpenAI Chat 原生请求", "model", req.Model)
	options := applyNativeOptions(opts)

	return retryNonStream(ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			return p.routing.GetChannelByProvider(ctx, req.Model, "openai", "chat_completions")
		},
		func(reqCtx context.Context, ch *routing.Channel) (*openaiChat.Response, error) {
			resp, err := p.request.Native(reqCtx, req, ch, req.Model)
			if err != nil {
				return nil, err
			}
			r, _ := resp.(*openaiChat.Response)
			return r, nil
		},
		compatFallback(options, errors.ErrCodeEndpointNotFound, func() (*openaiChat.Response, error) {
			p.logger.WithGroup("native_compat").WarnContext(ctx, "原生端点未找到，降级到默认端点",
				"request_mode", "compat",
				"model", req.Model,
				"provider", "openai",
				"endpoint_variant", "chat_completions",
			)
			return p.nativeOpenAIChatCompatFallback(ctx, req)
		}),
	)
}

// NativeOpenAIChatCompletionStream 执行 OpenAI Chat 原生流式请求
//
// 该方法通过 routing 获取通道，使用 retry 机制，调用 request.NativeStream。
// 请求体为 OpenAI Chat 原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - req: OpenAI Chat 原生请求对象
//
// 返回：
//   - <-chan *openaiChat.StreamEvent: 原生流事件通道
func (p *Portal) NativeOpenAIChatCompletionStream(
	ctx context.Context,
	req *openaiChat.Request,
	opts ...NativeOption,
) <-chan *openaiChat.StreamEvent {
	p.logger.DebugContext(ctx, "开始处理 OpenAI Chat 原生流式请求", "model", req.Model)
	options := applyNativeOptions(opts)

	return retryNativeStream(ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			return p.routing.GetChannelByProvider(ctx, req.Model, "openai", "chat_completions")
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			return p.request.NativeStream(reqCtx, req, ch, req.Model, output)
		},
		streamCompatFallback(ctx, options, errors.ErrCodeEndpointNotFound, func() <-chan *openaiChat.StreamEvent {
			p.logger.WithGroup("native_compat").WarnContext(ctx, "原生端点未找到，降级到默认端点",
				"request_mode", "compat",
				"model", req.Model,
				"provider", "openai",
				"endpoint_variant", "chat_completions",
			)
			return p.nativeOpenAIChatStreamCompatFallback(ctx, req)
		}),
	)
}

// NativeOpenAIResponses 执行 OpenAI Responses 原生请求（非流式）
//
// 该方法通过 routing 获取通道，使用 retry 机制，调用 request.Native。
// 请求体和响应体均为 OpenAI Responses 原生类型。
//
// 参数：
//   - ctx: 上下文
//   - req: OpenAI Responses 原生请求对象
//
// 返回：
//   - *openaiResponses.Response: OpenAI Responses 原生响应对象
//   - error: 请求失败时返回错误
func (p *Portal) NativeOpenAIResponses(
	ctx context.Context,
	req *openaiResponses.Request,
	opts ...NativeOption,
) (*openaiResponses.Response, error) {
	// 获取模型名称
	modelName := ""
	if req.Model != nil {
		modelName = *req.Model
	}

	p.logger.DebugContext(ctx, "开始处理 OpenAI Responses 原生请求", "model", modelName)
	options := applyNativeOptions(opts)

	return retryNonStream(ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			return p.routing.GetChannelByProvider(ctx, modelName, "openai", "responses")
		},
		func(reqCtx context.Context, ch *routing.Channel) (*openaiResponses.Response, error) {
			resp, err := p.request.Native(reqCtx, req, ch, modelName)
			if err != nil {
				return nil, err
			}
			r, _ := resp.(*openaiResponses.Response)
			return r, nil
		},
		compatFallback(options, errors.ErrCodeEndpointNotFound, func() (*openaiResponses.Response, error) {
			p.logger.WithGroup("native_compat").WarnContext(ctx, "原生端点未找到，降级到默认端点",
				"request_mode", "compat",
				"model", modelName,
				"provider", "openai",
				"endpoint_variant", "responses",
			)
			return p.nativeOpenAIResponsesCompatFallback(ctx, req)
		}),
	)
}

// NativeOpenAIResponsesStream 执行 OpenAI Responses 原生流式请求
//
// 该方法通过 routing 获取通道，使用 retry 机制，调用 request.NativeStream。
// 请求体为 OpenAI Responses 原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - req: OpenAI Responses 原生请求对象
//
// 返回：
//   - <-chan *openaiResponses.StreamEvent: 原生流事件通道
func (p *Portal) NativeOpenAIResponsesStream(
	ctx context.Context,
	req *openaiResponses.Request,
	opts ...NativeOption,
) <-chan *openaiResponses.StreamEvent {
	// 获取模型名称
	modelName := ""
	if req.Model != nil {
		modelName = *req.Model
	}

	p.logger.DebugContext(ctx, "开始处理 OpenAI Responses 原生流式请求", "model", modelName)
	options := applyNativeOptions(opts)

	return retryNativeStream(ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			return p.routing.GetChannelByProvider(ctx, modelName, "openai", "responses")
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			return p.request.NativeStream(reqCtx, req, ch, modelName, output)
		},
		streamCompatFallback(ctx, options, errors.ErrCodeEndpointNotFound, func() <-chan *openaiResponses.StreamEvent {
			p.logger.WithGroup("native_compat").WarnContext(ctx, "原生端点未找到，降级到默认端点",
				"request_mode", "compat",
				"model", modelName,
				"provider", "openai",
				"endpoint_variant", "responses",
			)
			return p.nativeOpenAIResponsesStreamCompatFallback(ctx, req)
		}),
	)
}
