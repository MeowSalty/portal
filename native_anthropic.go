package portal

import (
	"context"

	"github.com/MeowSalty/portal/errors"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/routing"
)

// NativeAnthropicMessages 执行 Anthropic Messages 原生请求（非流式）
//
// 该方法通过 routing 获取通道，使用 retry 机制，调用 request.Native。
// 请求体和响应体均为 Anthropic Messages 原生类型。
//
// 参数：
//   - ctx: 上下文
//   - req: Anthropic Messages 原生请求对象
//
// 返回：
//   - *anthropicTypes.Response: Anthropic Messages 原生响应对象
//   - error: 请求失败时返回错误
func (p *Portal) NativeAnthropicMessages(
	ctx context.Context,
	req *anthropicTypes.Request,
	opts ...NativeOption,
) (*anthropicTypes.Response, error) {
	p.logger.DebugContext(ctx, "开始处理 Anthropic Messages 原生请求", "model", req.Model)
	options := applyNativeOptions(opts)

	return retryNonStream(ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			return p.routing.GetChannelByProvider(ctx, req.Model, "anthropic", "messages")
		},
		func(reqCtx context.Context, ch *routing.Channel) (*anthropicTypes.Response, error) {
			resp, err := p.request.Native(reqCtx, req, ch, req.Model)
			if err != nil {
				return nil, err
			}
			r, _ := resp.(*anthropicTypes.Response)
			return r, nil
		},
		compatFallback(options, errors.ErrCodeEndpointNotFound, func() (*anthropicTypes.Response, error) {
			p.logger.WithGroup("native_compat").WarnContext(ctx, "原生端点未找到，降级到默认端点",
				"request_mode", "compat",
				"model", req.Model,
				"provider", "anthropic",
				"endpoint_variant", "messages",
			)
			return p.nativeAnthropicCompatFallback(ctx, req)
		}),
	)
}

// NativeAnthropicMessagesStream 执行 Anthropic Messages 原生流式请求
//
// 该方法通过 routing 获取通道，使用 retry 机制，调用 request.NativeStream。
// 请求体为 Anthropic Messages 原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - req: Anthropic Messages 原生请求对象
//
// 返回：
//   - <-chan *anthropicTypes.StreamEvent: 原生流事件通道
func (p *Portal) NativeAnthropicMessagesStream(
	ctx context.Context,
	req *anthropicTypes.Request,
	opts ...NativeOption,
) <-chan *anthropicTypes.StreamEvent {
	p.logger.DebugContext(ctx, "开始处理 Anthropic Messages 原生流式请求", "model", req.Model)
	options := applyNativeOptions(opts)

	return retryNativeStream[*anthropicTypes.StreamEvent](ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			return p.routing.GetChannelByProvider(ctx, req.Model, "anthropic", "messages")
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			return p.request.NativeStream(reqCtx, req, ch, req.Model, output)
		},
		streamCompatFallback(ctx, options, errors.ErrCodeEndpointNotFound, func() <-chan *anthropicTypes.StreamEvent {
			p.logger.WithGroup("native_compat").WarnContext(ctx, "原生端点未找到，降级到默认端点",
				"request_mode", "compat",
				"model", req.Model,
				"provider", "anthropic",
				"endpoint_variant", "messages",
			)
			return p.nativeAnthropicStreamCompatFallback(ctx, req)
		}),
	)
}
