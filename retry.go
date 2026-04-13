package portal

import (
	"context"
	"sync"

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
			return result, normalizeNonStreamCanceledError(ctx, ctx.Err())
		}

		channel, err := getChannel(ctx)
		if err != nil {
			if ctx.Err() != nil || errors.IsCanceled(err) {
				cancelErr := err
				if ctx.Err() != nil {
					cancelErr = ctx.Err()
				}
				return result, normalizeNonStreamCanceledError(ctx, cancelErr)
			}

			if onChannelErr != nil {
				if r, e, ok := onChannelErr(err); ok {
					return r, e
				}
			}
			return result, err
		}

		channelLogger := p.logger.With(
			"platform_id", channel.PlatformID,
			"model_id", channel.ModelID,
			"api_key_id", channel.APIKeyID,
		)
		channelLogger.DebugContext(ctx, "channel_selected")

		err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) error {
			defer reqCancel()
			var callErr error
			result, callErr = execute(reqCtx, channel)
			return callErr
		})

		if err != nil {
			if ctx.Err() != nil || errors.IsCanceled(err) || errors.IsCode(err, errors.ErrCodeAborted) {
				sourceErr := err
				if ctx.Err() != nil {
					sourceErr = ctx.Err()
				}
				cancelErr := normalizeNonStreamCanceledError(ctx, sourceErr)
				channelLogger.InfoContext(ctx, "request_canceled", "error", cancelErr)
				return result, cancelErr
			}

			if errors.IsRetryable(err) {
				if ctx.Err() != nil {
					cancelErr := normalizeNonStreamCanceledError(ctx, ctx.Err())
					channelLogger.InfoContext(ctx, "request_canceled", "error", cancelErr)
					return result, cancelErr
				}

				channelLogger.WarnContext(ctx, "request_retry_scheduled", "error", err)
				channel.MarkFailure(ctx, err)
				continue
			}

			channel.MarkFailure(ctx, err)
			return result, err
		}
		channel.MarkSuccess(ctx)
		channelLogger.InfoContext(ctx, "request_finished", "status", "completed")
		return result, nil
	}
}

// normalizeNonStreamCanceledError 归一化非流式取消类错误。
//
// 语义规则：
//   - deadline 统一映射为 DEADLINE_EXCEEDED/504/gateway
//   - 已带明确来源的取消错误保持原来源
//   - 外层 ctx 已取消时默认视为客户端取消
//   - 外层 ctx 未取消但出现取消错误时视为服务端取消
func normalizeNonStreamCanceledError(ctx context.Context, err error) error {
	if !errors.IsCanceled(err) {
		return err
	}

	switch errors.GetErrorFrom(err) {
	case errors.ErrorFromServer:
		return errors.NormalizeCanceledWithSource(err, false)
	case errors.ErrorFromClient:
		return errors.NormalizeCanceledWithSource(err, true)
	}

	if errors.IsCode(err, errors.ErrCodeCanceled) {
		return errors.NormalizeCanceledWithSource(err, false)
	}

	if errors.IsCode(err, errors.ErrCodeAborted) {
		return errors.NormalizeCanceledWithSource(err, true)
	}

	if errors.IsDeadlineExceeded(err) || errors.IsCode(err, errors.ErrCodeDeadlineExceeded) {
		return errors.NormalizeCanceledWithSource(err, false)
	}

	if ctx != nil && ctx.Err() != nil {
		return errors.NormalizeCanceledWithSource(ctx.Err(), true)
	}

	return errors.NormalizeCanceledWithSource(err, false)
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
		defer close(out)

		for {
			if ctx.Err() != nil {
				return
			}

			channel, err := getChannel(ctx)
			if err != nil {
				if ctx.Err() != nil || errors.IsCanceled(err) {
					return
				}

				if onChannelErr != nil && onChannelErr(err, out) {
					return
				}
				return
			}

			channelLogger := p.logger.With(
				"platform_id", channel.PlatformID,
				"model_id", channel.ModelID,
				"api_key_id", channel.APIKeyID,
			)
			channelLogger.DebugContext(ctx, "channel_selected")

			nativeOutput := make(chan any)
			done := make(chan struct{})
			var doneOnce sync.Once
			closeDone := func() {
				doneOnce.Do(func() {
					close(done)
				})
			}

			err = p.session.WithSessionStream(ctx, done, func(reqCtx context.Context) error {
				return execute(reqCtx, channel, nativeOutput)
			})

			if err != nil {
				closeDone()
				if ctx.Err() != nil || errors.IsCanceled(err) || errors.IsCode(err, errors.ErrCodeAborted) {
					cancelErr := normalizeStreamCanceledError(ctx, err)
					status, cancelSource := streamCanceledStatus(cancelErr)
					channelLogger.InfoContext(ctx, "stream_finished",
						"status", status,
						"completion_state", "not_completed",
						"connection_status", "disconnected",
						"cancel_source", cancelSource,
						"error", cancelErr,
					)
					return
				}

				if errors.IsRetryable(err) {
					if ctx.Err() != nil {
						cancelErr := normalizeStreamCanceledError(ctx, ctx.Err())
						status, cancelSource := streamCanceledStatus(cancelErr)
						channelLogger.InfoContext(ctx, "stream_finished",
							"status", status,
							"completion_state", "not_completed",
							"connection_status", "disconnected",
							"cancel_source", cancelSource,
							"error", cancelErr,
						)
						return
					}

					channelLogger.WarnContext(ctx, "request_retry_scheduled", "error", err)
					channel.MarkFailure(ctx, err)
					continue
				}

				channel.MarkFailure(ctx, err)
				channelLogger.WarnContext(ctx, "stream_finished",
					"status", "failed",
					"completion_state", "not_completed",
					"connection_status", "disconnected",
					"error", err,
				)
				return
			}

			streamDrained := false
			hasForwardedOutput := false
			for {
				var (
					event any
					ok    bool
				)

				// 优先尝试消费 nativeOutput，避免在“nativeOutput 已关闭”与“ctx.Done 可读”同时发生时误判为 canceled。
				select {
				case event, ok = <-nativeOutput:
				default:
					select {
					case event, ok = <-nativeOutput:
					case <-ctx.Done():
						closeDone()
						cancelErr := normalizeStreamCanceledError(ctx, ctx.Err())
						status, cancelSource := streamCanceledStatus(cancelErr)
						completionState := "not_completed"
						if hasForwardedOutput {
							completionState = "partial"
						}

						select {
						case _, drained := <-nativeOutput:
							if !drained {
								channel.MarkSuccess(ctx)
								channelLogger.InfoContext(ctx, "stream_finished",
									"status", "completed_then_disconnected",
									"completion_state", "completed",
									"connection_status", "completed_then_disconnected",
									"cancel_source", cancelSource,
									"termination_phase", "forwarding",
									"before_drain", true,
								)
								return
							}
						default:
						}

						channelLogger.InfoContext(ctx, "stream_finished",
							"status", status,
							"completion_state", completionState,
							"connection_status", "disconnected",
							"cancel_source", cancelSource,
							"termination_phase", "forwarding",
							"before_drain", true,
							"error", cancelErr,
						)
						return
					}
				}

				if !ok {
					streamDrained = true
					break
				}

				evt, ok := event.(T)
				if !ok {
					continue
				}

				select {
				case <-ctx.Done():
					closeDone()
					cancelErr := normalizeStreamCanceledError(ctx, ctx.Err())
					status, cancelSource := streamCanceledStatus(cancelErr)
					completionState := "not_completed"
					if hasForwardedOutput {
						completionState = "partial"
					}

					select {
					case _, drained := <-nativeOutput:
						if !drained {
							channel.MarkSuccess(ctx)
							channelLogger.InfoContext(ctx, "stream_finished",
								"status", "completed_then_disconnected",
								"completion_state", "completed",
								"connection_status", "completed_then_disconnected",
								"cancel_source", cancelSource,
								"termination_phase", "forwarding",
								"before_drain", true,
							)
							return
						}
					default:
					}

					if completionState == "completed" {
						channel.MarkSuccess(ctx)
						channelLogger.InfoContext(ctx, "stream_finished",
							"status", "completed_then_disconnected",
							"completion_state", completionState,
							"connection_status", "completed_then_disconnected",
							"cancel_source", cancelSource,
							"termination_phase", "forwarding",
							"before_drain", true,
						)
						return
					}
					channelLogger.InfoContext(ctx, "stream_finished",
						"status", status,
						"completion_state", completionState,
						"connection_status", "disconnected",
						"cancel_source", cancelSource,
						"termination_phase", "forwarding",
						"before_drain", true,
						"error", cancelErr,
					)
					return
				case out <- evt:
					hasForwardedOutput = true
				}
			}

			closeDone()

			cleanupCanceled := streamDrained && ctx.Err() != nil
			if cleanupCanceled {
				cancelErr := normalizeStreamCanceledError(ctx, ctx.Err())
				status, cancelSource := streamCanceledStatus(cancelErr)
				if hasForwardedOutput {
					channelLogger.DebugContext(ctx, "stream_cleanup_canceled",
						"status", status,
						"after_drain", true,
						"cancel_source", cancelSource,
					)
				} else {
					channelLogger.InfoContext(ctx, "stream_finished",
						"status", status,
						"completion_state", "not_completed",
						"connection_status", "disconnected",
						"cancel_source", cancelSource,
						"termination_phase", "drain",
						"after_drain", true,
						"error", cancelErr,
					)
					return
				}
			}

			channel.MarkSuccess(ctx)
			channelLogger.InfoContext(ctx, "stream_finished",
				"status", "completed",
				"completion_state", "completed",
				"connection_status", "disconnected",
			)
			return
		}
	}()

	return out
}

func normalizeStreamCanceledError(ctx context.Context, err error) error {
	if ctx != nil && ctx.Err() != nil {
		source := errors.GetErrorFrom(err)
		isClient := source != errors.ErrorFromServer
		return errors.NormalizeCanceledWithSource(ctx.Err(), isClient)
	}

	if err == nil {
		return errors.NormalizeCanceled(context.Canceled)
	}

	if errors.IsCode(err, errors.ErrCodeCanceled) {
		return errors.NormalizeCanceledWithSource(err, false)
	}

	if errors.IsCode(err, errors.ErrCodeAborted) {
		return errors.NormalizeCanceledWithSource(err, true)
	}

	if errors.IsDeadlineExceeded(err) || errors.IsCode(err, errors.ErrCodeDeadlineExceeded) {
		return errors.NormalizeCanceledWithSource(err, false)
	}

	return errors.NormalizeCanceled(err)
}

func streamCanceledStatus(err error) (status string, cancelSource string) {
	if errors.IsDeadlineExceeded(err) || errors.IsCode(err, errors.ErrCodeDeadlineExceeded) {
		return "timed_out", "deadline"
	}

	if errors.IsCode(err, errors.ErrCodeCanceled) || errors.GetErrorFrom(err) == errors.ErrorFromServer {
		return "canceled", "server"
	}

	if errors.IsCode(err, errors.ErrCodeAborted) || errors.GetErrorFrom(err) == errors.ErrorFromClient {
		return "canceled", "client"
	}

	return "canceled", "unknown"
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
