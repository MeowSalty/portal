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
) (*geminiTypes.Response, error) {
	// 获取模型名称
	modelName := req.Model

	p.logger.DebugContext(ctx, "开始处理 Gemini GenerateContent 原生请求", "model", modelName)

	var response *geminiTypes.Response
	var channel *routing.Channel
	var err error
	for {
		channel, err = p.routing.GetChannelByProvider(ctx, modelName, "google", "generate")
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

			// 调用 request.Native
			resp, err := p.request.Native(reqCtx, req, channel, modelName)
			if err != nil {
				return err
			}

			if r, ok := resp.(*geminiTypes.Response); ok {
				response = r
			}
			return nil
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
) <-chan *geminiTypes.StreamEvent {
	// 获取模型名称
	modelName := req.Model

	p.logger.DebugContext(ctx, "开始处理 Gemini StreamGenerateContent 原生流式请求", "model", modelName)

	// 创建内部流（用于接收原始响应）
	internalStream := make(chan *geminiTypes.StreamEvent, StreamBufferSize)

	// 启动内部流处理协程
	go func() {
		for {
			channel, err := p.routing.GetChannelByProvider(ctx, modelName, "google", "generate")
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

			// 创建原生事件输出通道
			nativeOutput := make(chan any)
			// 创建流结束信号通道
			done := make(chan struct{})

			err = p.session.WithSessionStream(ctx, done, func(reqCtx context.Context) error {
				// 调用 request.NativeStream
				return p.request.NativeStream(reqCtx, req, channel, modelName, nativeOutput)
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

			// 转换原生事件到指定类型
			go func() {
				defer close(internalStream)
				defer close(done) // 流结束时通知会话管理器
				for event := range nativeOutput {
					if evt, ok := event.(*geminiTypes.StreamEvent); ok {
						select {
						case <-ctx.Done():
							return
						case internalStream <- evt:
						}
					}
				}
			}()

			break
		}
	}()

	return internalStream
}
