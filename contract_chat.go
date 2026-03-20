package portal

import (
	"context"
	"strconv"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
)

// ChatCompletion 处理聊天完成请求（非流式）
func (p *Portal) ChatCompletion(ctx context.Context, request *types.RequestContract) (*types.ResponseContract, error) {
	p.logger.DebugContext(ctx, "request_started", "model", request.Model)

	response, err := retryNonStream(ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			return p.routing.GetChannel(ctx, request.Model)
		},
		func(reqCtx context.Context, ch *routing.Channel) (*types.ResponseContract, error) {
			return p.request.ChatCompletion(reqCtx, request, ch)
		},
		nil,
	)

	// 通过中间件链处理响应
	if response != nil && p.middleware != nil {
		response, err = p.middleware.Process(ctx, request, response)
	}

	if err != nil {
		p.logger.ErrorContext(ctx, "request_failed", "model", request.Model, "error", err)
	} else {
		p.logger.InfoContext(ctx, "request_finished", "model", request.Model)
	}

	return response, err
}

// ChatCompletionStream 处理流式聊天完成请求
func (p *Portal) ChatCompletionStream(ctx context.Context, request *types.RequestContract) <-chan *types.StreamEventContract {
	p.logger.DebugContext(ctx, "request_started", "model", request.Model)

	// 创建内部流（用于接收原始响应）
	internalStream := make(chan *types.StreamEventContract, StreamBufferSize)

	// 启动内部流处理协程
	go func() {
		for {
			channel, err := p.routing.GetChannel(ctx, request.Model)
			if err != nil {
				if errors.IsCode(err, errors.ErrCodeAborted) || errors.IsCanceled(err) || errors.IsCanceled(ctx.Err()) {
					p.logger.InfoContext(ctx, "stream_finished", "model", request.Model, "status", "canceled", "error", errors.NormalizeCanceled(err))
				} else {
					p.logger.ErrorContext(ctx, "request_failed", "model", request.Model, "error", err)
				}
				message := errors.GetMessage(err)
				if message == "" {
					message = err.Error()
				}
				statusCode := errors.GetHTTPStatus(err)
				code := ""
				if statusCode > 0 {
					code = strconv.Itoa(statusCode)
				}
				// 创建错误响应并发送到流中
				errorResponse := &types.StreamEventContract{
					Type: types.StreamEventError,
					Error: &types.StreamErrorPayload{
						Message: message,
						Type:    "stream_error",
						Code:    code,
					},
				}
				select {
				case <-ctx.Done():
				default:
					select {
					case internalStream <- errorResponse:
					default:
					}
				}
				close(internalStream)
				break
			}

			// 使用 With 创建带有通道上下文的日志记录器
			channelLogger := p.logger.With(
				"platform_id", channel.PlatformID,
				"model_id", channel.ModelID,
				"api_key_id", channel.APIKeyID)

			err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) (err error) {
				defer reqCancel()
				return p.request.ChatCompletionStream(reqCtx, request, internalStream, channel)
			})

			// 检查错误是否可以重试
			if err != nil {
				if errors.IsRetryable(err) {
					channelLogger.WarnContext(ctx, "request_retry_scheduled", "error", err)
					channel.MarkFailure(ctx, err)
					continue
				}
				// 特殊处理：主动取消/操作终止不视为失败噪音
				if errors.IsCode(err, errors.ErrCodeAborted) || errors.IsCanceled(err) || errors.IsCanceled(ctx.Err()) {
					channelLogger.InfoContext(ctx, "stream_finished", "status", "canceled")
					channel.MarkSuccess(ctx)
					close(internalStream)
					break
				}
				p.logger.ErrorContext(ctx, "request_failed", "model", request.Model, "error", err)
				channel.MarkFailure(ctx, err)
				close(internalStream)
				break
			}
			channel.MarkSuccess(ctx)
			p.logger.InfoContext(ctx, "stream_finished", "model", request.Model, "status", "completed")
			break
		}
	}()

	// 通过中间件链处理流式响应
	outputStream := p.middleware.ProcessStream(ctx, request, internalStream)

	return outputStream
}
