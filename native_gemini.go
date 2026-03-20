package portal

import (
	"context"

	"github.com/MeowSalty/portal/errors"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	"github.com/MeowSalty/portal/routing"
)

// NativeGeminiGenerateContent 执行 Gemini GenerateContent 原生请求（非流式）
//
// 该方法通过 routing 获取通道，使用 retry 机制，调用 request.Native。
// 请求体和响应体均为 Gemini GenerateContent 原生类型。
//
// 参数：
//   - ctx: 上下文
//   - req: Gemini GenerateContent 原生请求对象
//
// 返回：
//   - *geminiTypes.Response: Gemini GenerateContent 原生响应对象
//   - error: 请求失败时返回错误
func (p *Portal) NativeGeminiGenerateContent(
	ctx context.Context,
	req *geminiTypes.Request,
	opts ...NativeOption,
) (*geminiTypes.Response, error) {
	modelName := req.Model
	options := applyNativeOptions(opts)

	p.logger.DebugContext(ctx, "request_started", "model", modelName)

	return retryNonStream(ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			return p.routing.GetChannelByProvider(ctx, modelName, "google", "generate")
		},
		func(reqCtx context.Context, ch *routing.Channel) (*geminiTypes.Response, error) {
			resp, err := p.request.Native(reqCtx, req, ch, modelName)
			if err != nil {
				return nil, err
			}
			r, _ := resp.(*geminiTypes.Response)
			return r, nil
		},
		compatFallback(options, errors.ErrCodeEndpointNotFound, func() (*geminiTypes.Response, error) {
			p.logger.WithGroup("native_compat").InfoContext(ctx, "compat_fallback_applied",
				"request_mode", "compat",
				"model", modelName,
				"provider", "google",
				"endpoint_variant", "generate",
			)
			return p.nativeGeminiCompatFallback(ctx, req)
		}),
	)
}

// NativeGeminiStreamGenerateContent 执行 Gemini StreamGenerateContent 原生流式请求
//
// 该方法通过 routing 获取通道，使用 retry 机制，调用 request.NativeStream。
// 请求体为 Gemini GenerateContent 原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - req: Gemini GenerateContent 原生请求对象
//
// 返回：
//   - <-chan *geminiTypes.StreamEvent: 原生流事件通道
func (p *Portal) NativeGeminiStreamGenerateContent(
	ctx context.Context,
	req *geminiTypes.Request,
	opts ...NativeOption,
) <-chan *geminiTypes.StreamEvent {
	modelName := req.Model
	options := applyNativeOptions(opts)

	p.logger.DebugContext(ctx, "request_started", "model", modelName)

	return retryNativeStream(ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			return p.routing.GetChannelByProvider(ctx, modelName, "google", "generate")
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			return p.request.NativeStream(reqCtx, req, ch, modelName, output)
		},
		streamCompatFallback(ctx, options, errors.ErrCodeEndpointNotFound, func() <-chan *geminiTypes.StreamEvent {
			p.logger.WithGroup("native_compat").InfoContext(ctx, "compat_fallback_applied",
				"request_mode", "compat",
				"model", modelName,
				"provider", "google",
				"endpoint_variant", "generate",
			)
			return p.nativeGeminiStreamCompatFallback(ctx, req)
		}),
	)
}
