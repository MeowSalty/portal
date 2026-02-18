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
	p.logger.DebugContext(ctx, "开始处理聊天完成请求", "model", request.Model)

	var response *types.ResponseContract
	var channel *routing.Channel
	var err error
	for {
		channel, err = p.routing.GetChannel(ctx, request.Model)
		if err != nil {
			p.logger.ErrorContext(ctx, "获取通道失败", "model", request.Model, "error", err)
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
			response, err = p.request.ChatCompletion(reqCtx, request, channel)
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

	// 通过中间件链处理响应
	if response != nil && p.middleware != nil {
		response, err = p.middleware.Process(ctx, request, response)
		if err != nil {
			p.logger.ErrorContext(ctx, "中间件处理失败", "error", err)
		}
	}

	return response, err
}

// ChatCompletionStream 处理流式聊天完成请求
func (p *Portal) ChatCompletionStream(ctx context.Context, request *types.RequestContract) <-chan *types.StreamEventContract {
	p.logger.DebugContext(ctx, "开始处理流式聊天完成请求", "model", request.Model)

	// 创建内部流（用于接收原始响应）
	internalStream := make(chan *types.StreamEventContract, StreamBufferSize)

	// 启动内部流处理协程
	go func() {
		for {
			channel, err := p.routing.GetChannel(ctx, request.Model)
			if err != nil {
				p.logger.ErrorContext(ctx, "获取通道失败", "model", request.Model, "error", err)
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

			channelLogger.DebugContext(ctx, "获取到通道")

			err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) (err error) {
				defer reqCancel()
				return p.request.ChatCompletionStream(reqCtx, request, internalStream, channel)
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

	// 通过中间件链处理流式响应
	outputStream := p.middleware.ProcessStream(ctx, request, internalStream)

	return outputStream
}
