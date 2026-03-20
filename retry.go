package portal

import (
	"context"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/routing"
)

// channelFunc 获取通道的函数类型
type channelFunc = func(ctx context.Context) (*routing.Channel, error)

// retryNonStream 非流式重试通用函数
//
// 封装了获取通道 → 创建 logger → 会话包装 → 执行请求 → 重试决策 → 健康标记的完整流程。
//
// 参数：
//   - ctx: 上下文
//   - p: Portal 实例
//   - getChannel: 获取通道的函数
//   - execute: 在会话中执行的请求函数
//   - onChannelErr: 通道获取失败时的回调（nil 表示不处理；返回 result, err, handled）
func retryNonStream[T any](
	ctx context.Context,
	p *Portal,
	getChannel channelFunc,
	execute func(reqCtx context.Context, ch *routing.Channel) (T, error),
	onChannelErr func(err error) (T, error, bool),
) (T, error) {
	var result T
	for {
		if ctx.Err() != nil {
			return result, errors.NormalizeCanceled(ctx.Err())
		}

		channel, err := getChannel(ctx)
		if err != nil {
			if ctx.Err() != nil || errors.IsCanceled(err) {
				return result, errors.NormalizeCanceled(err)
			}

			if onChannelErr != nil {
				if r, e, ok := onChannelErr(err); ok {
					return r, e
				}
			}
			p.logger.ErrorContext(ctx, "获取通道失败", "error", err)
			return result, err
		}

		channelLogger := p.logger.With(
			"platform_id", channel.PlatformID,
			"model_id", channel.ModelID,
			"api_key_id", channel.APIKeyID,
		)
		channelLogger.DebugContext(ctx, "获取到通道")

		err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) error {
			defer reqCancel()
			var callErr error
			result, callErr = execute(reqCtx, channel)
			return callErr
		})

		if err != nil {
			if ctx.Err() != nil || errors.IsCanceled(err) || errors.IsCode(err, errors.ErrCodeAborted) {
				cancelErr := errors.NormalizeCanceled(err)
				channelLogger.InfoContext(ctx, "操作终止", "error", cancelErr)
				return result, cancelErr
			}

			if errors.IsRetryable(err) {
				if ctx.Err() != nil {
					cancelErr := errors.NormalizeCanceled(ctx.Err())
					channelLogger.InfoContext(ctx, "操作终止", "error", cancelErr)
					return result, cancelErr
				}

				channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
				channel.MarkFailure(ctx, err)
				continue
			}

			channelLogger.ErrorContext(ctx, "请求处理失败", "error", err)
			channel.MarkFailure(ctx, err)
			return result, err
		}
		channel.MarkSuccess(ctx)
		channelLogger.InfoContext(ctx, "请求处理成功")
		return result, nil
	}
}

// retryNativeStream 原生流式重试通用函数
//
// 封装了获取通道 → 创建 logger → 会话包装 → 执行流式请求 → 重试决策 → 健康标记 → 类型转发的完整流程。
//
// 参数：
//   - ctx: 上下文
//   - p: Portal 实例
//   - getChannel: 获取通道的函数
//   - execute: 在会话中执行的流式请求函数
//   - onChannelErr: 通道获取失败时的回调（nil 表示不处理；返回 true 表示已处理）
func retryNativeStream[T any](
	ctx context.Context,
	p *Portal,
	getChannel channelFunc,
	execute func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error,
	onChannelErr func(err error, out chan<- T) bool,
) <-chan T {
	out := make(chan T, StreamBufferSize)

	go func() {
		for {
			if ctx.Err() != nil {
				close(out)
				return
			}

			channel, err := getChannel(ctx)
			if err != nil {
				if ctx.Err() != nil || errors.IsCanceled(err) {
					close(out)
					return
				}

				if onChannelErr != nil && onChannelErr(err, out) {
					close(out)
					return
				}
				p.logger.ErrorContext(ctx, "获取通道失败", "error", err)
				close(out)
				return
			}

			channelLogger := p.logger.With(
				"platform_id", channel.PlatformID,
				"model_id", channel.ModelID,
				"api_key_id", channel.APIKeyID,
			)
			channelLogger.DebugContext(ctx, "获取到通道")

			nativeOutput := make(chan any)
			done := make(chan struct{})

			err = p.session.WithSessionStream(ctx, done, func(reqCtx context.Context) error {
				return execute(reqCtx, channel, nativeOutput)
			})

			if err != nil {
				if ctx.Err() != nil || errors.IsCanceled(err) || errors.IsCode(err, errors.ErrCodeAborted) {
					cancelErr := errors.NormalizeCanceled(err)
					channelLogger.InfoContext(ctx, "操作终止", "error", cancelErr)
					close(out)
					return
				}

				if errors.IsRetryable(err) {
					if ctx.Err() != nil {
						channelLogger.InfoContext(ctx, "操作终止", "error", errors.NormalizeCanceled(ctx.Err()))
						close(out)
						return
					}

					channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
					channel.MarkFailure(ctx, err)
					continue
				}

				channelLogger.ErrorContext(ctx, "流处理失败", "error", err)
				channel.MarkFailure(ctx, err)
				close(out)
				return
			}
			channel.MarkSuccess(ctx)
			channelLogger.InfoContext(ctx, "流处理成功")

			go func() {
				defer close(out)
				defer close(done)
				for event := range nativeOutput {
					if evt, ok := event.(T); ok {
						select {
						case <-ctx.Done():
							return
						case out <- evt:
						}
					}
				}
			}()
			return
		}
	}()

	return out
}

// compatFallback 构造非流式 compat 降级回调
//
// 当 options 为 nil 或 compatMode 为 false 时返回 nil。
// 当通道获取失败且错误码匹配时，执行降级回调。
func compatFallback[T any](
	opts *nativeOptions,
	code errors.ErrorCode,
	fallback func() (T, error),
) func(error) (T, error, bool) {
	if opts == nil || !opts.compatMode {
		return nil
	}
	return func(err error) (T, error, bool) {
		if errors.IsCode(err, code) {
			result, err := fallback()
			return result, err, true
		}
		var zero T
		return zero, nil, false
	}
}

// streamCompatFallback 构造流式 compat 降级回调
//
// 当 options 为 nil 或 compatMode 为 false 时返回 nil。
// 当通道获取失败且错误码匹配时，从降级流转发事件到输出通道。
func streamCompatFallback[T any](
	ctx context.Context,
	opts *nativeOptions,
	code errors.ErrorCode,
	fallback func() <-chan T,
) func(error, chan<- T) bool {
	if opts == nil || !opts.compatMode {
		return nil
	}
	return func(err error, out chan<- T) bool {
		if errors.IsCode(err, code) {
			compatStream := fallback()
			for evt := range compatStream {
				select {
				case <-ctx.Done():
					return true
				case out <- evt:
				}
			}
			return true
		}
		return false
	}
}
