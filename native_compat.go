package portal

import (
	"context"
	"strconv"

	"github.com/MeowSalty/portal/errors"
	anthropicConverter "github.com/MeowSalty/portal/request/adapter/anthropic/converter"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	geminiConverter "github.com/MeowSalty/portal/request/adapter/gemini/converter"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	openaiChatConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/chat"
	openaiResponsesConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/responses"
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
)

// nativeOpenAIChatCompatFallback 在 OpenAI Chat 原生端点不可用时走 Contract 降级（非流式）。
func (p *Portal) nativeOpenAIChatCompatFallback(
	ctx context.Context,
	req *openaiChat.Request,
) (*openaiChat.Response, error) {
	compatLogger := p.logger.WithGroup("native_compat").With(
		"request_mode", "compat",
		"provider", "openai",
		"endpoint_variant", "chat_completions",
	)

	contractReq, err := openaiChatConverter.RequestToContract(req)
	if err != nil {
		return nil, err
	}

	channel, err := p.routing.GetChannel(ctx, req.Model)
	if err != nil {
		return nil, err
	}

	contractResp, err := p.doCompatContractNonStream(ctx, contractReq, channel)
	if err != nil {
		return nil, err
	}

	return openaiChatConverter.ResponseFromContract(contractResp, compatLogger.WithGroup("converter"))
}

// nativeOpenAIResponsesCompatFallback 在 OpenAI Responses 原生端点不可用时走 Contract 降级（非流式）。
func (p *Portal) nativeOpenAIResponsesCompatFallback(
	ctx context.Context,
	req *openaiResponses.Request,
) (*openaiResponses.Response, error) {
	modelName := ""
	if req.Model != nil {
		modelName = *req.Model
	}

	compatLogger := p.logger.WithGroup("native_compat").With(
		"request_mode", "compat",
		"provider", "openai",
		"endpoint_variant", "responses",
	)

	contractReq, err := openaiResponsesConverter.RequestToContract(req)
	if err != nil {
		return nil, err
	}

	channel, err := p.routing.GetChannel(ctx, modelName)
	if err != nil {
		return nil, err
	}

	contractResp, err := p.doCompatContractNonStream(ctx, contractReq, channel)
	if err != nil {
		return nil, err
	}

	return openaiResponsesConverter.ResponseFromContract(contractResp, compatLogger.WithGroup("converter"))
}

// nativeAnthropicCompatFallback 在 Anthropic 原生端点不可用时走 Contract 降级（非流式）。
func (p *Portal) nativeAnthropicCompatFallback(
	ctx context.Context,
	req *anthropicTypes.Request,
) (*anthropicTypes.Response, error) {
	compatLogger := p.logger.WithGroup("native_compat").With(
		"request_mode", "compat",
		"provider", "anthropic",
		"endpoint_variant", "messages",
	)

	contractReq, err := anthropicConverter.RequestToContract(req)
	if err != nil {
		return nil, err
	}

	channel, err := p.routing.GetChannel(ctx, req.Model)
	if err != nil {
		return nil, err
	}

	contractResp, err := p.doCompatContractNonStream(ctx, contractReq, channel)
	if err != nil {
		return nil, err
	}

	return anthropicConverter.ResponseFromContract(contractResp, compatLogger.WithGroup("converter"))
}

// nativeGeminiCompatFallback 在 Gemini 原生端点不可用时走 Contract 降级（非流式）。
func (p *Portal) nativeGeminiCompatFallback(
	ctx context.Context,
	req *geminiTypes.Request,
) (*geminiTypes.Response, error) {
	contractReq, err := geminiConverter.RequestToContract(req)
	if err != nil {
		return nil, err
	}

	channel, err := p.routing.GetChannel(ctx, req.Model)
	if err != nil {
		return nil, err
	}

	contractResp, err := p.doCompatContractNonStream(ctx, contractReq, channel)
	if err != nil {
		return nil, err
	}

	return geminiConverter.ResponseFromContract(contractResp)
}

// nativeOpenAIChatStreamCompatFallback 在 OpenAI Chat 原生端点不可用时走 Contract 降级（流式）。
func (p *Portal) nativeOpenAIChatStreamCompatFallback(
	ctx context.Context,
	req *openaiChat.Request,
) <-chan *openaiChat.StreamEvent {
	outputStream := make(chan *openaiChat.StreamEvent, StreamBufferSize)

	go func() {
		defer close(outputStream)

		compatLogger := p.logger.WithGroup("native_compat").With(
			"request_mode", "compat",
			"provider", "openai",
			"endpoint_variant", "chat_completions",
		)

		contractReq, err := openaiChatConverter.RequestToContract(req)
		if err != nil {
			p.sendNativeCompatOpenAIChatStreamErrorEvent(outputStream, err)
			return
		}

		channel, err := p.routing.GetChannel(ctx, req.Model)
		if err != nil {
			p.sendNativeCompatOpenAIChatStreamErrorEvent(outputStream, err)
			return
		}

		for {
			channelLogger := compatLogger.With(
				"platform_id", channel.PlatformID,
				"model_id", channel.ModelID,
				"api_key_id", channel.APIKeyID,
			)

			contractStream := make(chan *adapterTypes.StreamEventContract, StreamBufferSize)
			done := make(chan struct{})

			err = p.session.WithSessionStream(ctx, done, func(reqCtx context.Context) error {
				return p.request.ChatCompletionStream(reqCtx, contractReq, contractStream, channel)
			})

			if err != nil {
				if errors.IsRetryable(err) {
					channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
					channel.MarkFailure(ctx, err)
					nextChannel, routeErr := p.routing.GetChannel(ctx, contractReq.Model)
					if routeErr != nil {
						p.sendNativeCompatOpenAIChatStreamErrorEvent(outputStream, routeErr)
						return
					}
					channel = nextChannel
					continue
				}
				if errors.IsCode(err, errors.ErrCodeAborted) {
					channelLogger.InfoContext(ctx, "操作终止")
					channel.MarkSuccess(ctx)
				}
				channelLogger.ErrorContext(ctx, "流处理失败", "error", err)
				channel.MarkFailure(ctx, err)
				p.sendNativeCompatOpenAIChatStreamErrorEvent(outputStream, err)
				return
			}

			channel.MarkSuccess(ctx)
			channelLogger.InfoContext(ctx, "流处理成功")

			for contractEvent := range contractStream {
				nativeEvent, convertErr := openaiChatConverter.StreamEventFormContract(contractEvent, compatLogger.WithGroup("converter"))
				if convertErr != nil {
					compatLogger.ErrorContext(ctx, "转换流事件失败", "error", convertErr)
					p.sendNativeCompatOpenAIChatStreamErrorEvent(outputStream, convertErr)
					close(done)
					return
				}
				if nativeEvent == nil {
					continue
				}
				select {
				case <-ctx.Done():
					close(done)
					return
				case outputStream <- nativeEvent:
				}
			}

			close(done)
			return
		}
	}()

	return outputStream
}

// nativeOpenAIResponsesStreamCompatFallback 在 OpenAI Responses 原生端点不可用时走 Contract 降级（流式）。
func (p *Portal) nativeOpenAIResponsesStreamCompatFallback(
	ctx context.Context,
	req *openaiResponses.Request,
) <-chan *openaiResponses.StreamEvent {
	outputStream := make(chan *openaiResponses.StreamEvent, StreamBufferSize)

	go func() {
		defer close(outputStream)

		modelName := ""
		if req.Model != nil {
			modelName = *req.Model
		}

		compatLogger := p.logger.WithGroup("native_compat").With(
			"request_mode", "compat",
			"provider", "openai",
			"endpoint_variant", "responses",
		)

		contractReq, err := openaiResponsesConverter.RequestToContract(req)
		if err != nil {
			p.sendNativeCompatOpenAIResponsesStreamErrorEvent(outputStream, err)
			return
		}

		channel, err := p.routing.GetChannel(ctx, modelName)
		if err != nil {
			p.sendNativeCompatOpenAIResponsesStreamErrorEvent(outputStream, err)
			return
		}

		for {
			channelLogger := compatLogger.With(
				"platform_id", channel.PlatformID,
				"model_id", channel.ModelID,
				"api_key_id", channel.APIKeyID,
			)

			contractStream := make(chan *adapterTypes.StreamEventContract, StreamBufferSize)
			done := make(chan struct{})

			err = p.session.WithSessionStream(ctx, done, func(reqCtx context.Context) error {
				return p.request.ChatCompletionStream(reqCtx, contractReq, contractStream, channel)
			})

			if err != nil {
				if errors.IsRetryable(err) {
					channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
					channel.MarkFailure(ctx, err)
					nextChannel, routeErr := p.routing.GetChannel(ctx, contractReq.Model)
					if routeErr != nil {
						p.sendNativeCompatOpenAIResponsesStreamErrorEvent(outputStream, routeErr)
						return
					}
					channel = nextChannel
					continue
				}
				if errors.IsCode(err, errors.ErrCodeAborted) {
					channelLogger.InfoContext(ctx, "操作终止")
					channel.MarkSuccess(ctx)
				}
				channelLogger.ErrorContext(ctx, "流处理失败", "error", err)
				channel.MarkFailure(ctx, err)
				p.sendNativeCompatOpenAIResponsesStreamErrorEvent(outputStream, err)
				return
			}

			channel.MarkSuccess(ctx)
			channelLogger.InfoContext(ctx, "流处理成功")

			indexCtx := adapterTypes.NewStreamIndexContext()
			for contractEvent := range contractStream {
				nativeEvents, convertErr := openaiResponsesConverter.StreamEventFormContract(contractEvent, compatLogger.WithGroup("converter"), indexCtx)
				if convertErr != nil {
					compatLogger.ErrorContext(ctx, "转换流事件失败", "error", convertErr)
					p.sendNativeCompatOpenAIResponsesStreamErrorEvent(outputStream, convertErr)
					close(done)
					return
				}
				for _, nativeEvent := range nativeEvents {
					if nativeEvent == nil {
						continue
					}
					select {
					case <-ctx.Done():
						close(done)
						return
					case outputStream <- nativeEvent:
					}
				}
			}

			close(done)
			return
		}
	}()

	return outputStream
}

// nativeAnthropicStreamCompatFallback 在 Anthropic 原生端点不可用时走 Contract 降级（流式）。
func (p *Portal) nativeAnthropicStreamCompatFallback(
	ctx context.Context,
	req *anthropicTypes.Request,
) <-chan *anthropicTypes.StreamEvent {
	outputStream := make(chan *anthropicTypes.StreamEvent, StreamBufferSize)

	go func() {
		defer close(outputStream)

		compatLogger := p.logger.WithGroup("native_compat").With(
			"request_mode", "compat",
			"provider", "anthropic",
			"endpoint_variant", "messages",
		)

		contractReq, err := anthropicConverter.RequestToContract(req)
		if err != nil {
			p.sendNativeCompatAnthropicStreamErrorEvent(outputStream, err)
			return
		}

		channel, err := p.routing.GetChannel(ctx, req.Model)
		if err != nil {
			p.sendNativeCompatAnthropicStreamErrorEvent(outputStream, err)
			return
		}

		for {
			channelLogger := compatLogger.With(
				"platform_id", channel.PlatformID,
				"model_id", channel.ModelID,
				"api_key_id", channel.APIKeyID,
			)

			contractStream := make(chan *adapterTypes.StreamEventContract, StreamBufferSize)
			done := make(chan struct{})

			err = p.session.WithSessionStream(ctx, done, func(reqCtx context.Context) error {
				return p.request.ChatCompletionStream(reqCtx, contractReq, contractStream, channel)
			})

			if err != nil {
				if errors.IsRetryable(err) {
					channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
					channel.MarkFailure(ctx, err)
					nextChannel, routeErr := p.routing.GetChannel(ctx, contractReq.Model)
					if routeErr != nil {
						p.sendNativeCompatAnthropicStreamErrorEvent(outputStream, routeErr)
						return
					}
					channel = nextChannel
					continue
				}
				if errors.IsCode(err, errors.ErrCodeAborted) {
					channelLogger.InfoContext(ctx, "操作终止")
					channel.MarkSuccess(ctx)
				}
				channelLogger.ErrorContext(ctx, "流处理失败", "error", err)
				channel.MarkFailure(ctx, err)
				p.sendNativeCompatAnthropicStreamErrorEvent(outputStream, err)
				return
			}

			channel.MarkSuccess(ctx)
			channelLogger.InfoContext(ctx, "流处理成功")

			for contractEvent := range contractStream {
				nativeEvent, convertErr := anthropicConverter.StreamEventFromContract(contractEvent, compatLogger.WithGroup("converter"))
				if convertErr != nil {
					compatLogger.ErrorContext(ctx, "转换流事件失败", "error", convertErr)
					p.sendNativeCompatAnthropicStreamErrorEvent(outputStream, convertErr)
					close(done)
					return
				}
				if nativeEvent == nil {
					continue
				}
				select {
				case <-ctx.Done():
					close(done)
					return
				case outputStream <- nativeEvent:
				}
			}

			close(done)
			return
		}
	}()

	return outputStream
}

// nativeGeminiStreamCompatFallback 在 Gemini 原生端点不可用时走 Contract 降级（流式）。
func (p *Portal) nativeGeminiStreamCompatFallback(
	ctx context.Context,
	req *geminiTypes.Request,
) <-chan *geminiTypes.StreamEvent {
	outputStream := make(chan *geminiTypes.StreamEvent, StreamBufferSize)

	go func() {
		defer close(outputStream)

		compatLogger := p.logger.WithGroup("native_compat").With(
			"request_mode", "compat",
			"provider", "google",
			"endpoint_variant", "generate",
		)

		contractReq, err := geminiConverter.RequestToContract(req)
		if err != nil {
			p.sendNativeCompatGeminiStreamErrorEvent(outputStream, err)
			return
		}

		channel, err := p.routing.GetChannel(ctx, req.Model)
		if err != nil {
			p.sendNativeCompatGeminiStreamErrorEvent(outputStream, err)
			return
		}

		for {
			channelLogger := compatLogger.With(
				"platform_id", channel.PlatformID,
				"model_id", channel.ModelID,
				"api_key_id", channel.APIKeyID,
			)

			contractStream := make(chan *adapterTypes.StreamEventContract, StreamBufferSize)
			done := make(chan struct{})

			err = p.session.WithSessionStream(ctx, done, func(reqCtx context.Context) error {
				return p.request.ChatCompletionStream(reqCtx, contractReq, contractStream, channel)
			})

			if err != nil {
				if errors.IsRetryable(err) {
					channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
					channel.MarkFailure(ctx, err)
					nextChannel, routeErr := p.routing.GetChannel(ctx, contractReq.Model)
					if routeErr != nil {
						p.sendNativeCompatGeminiStreamErrorEvent(outputStream, routeErr)
						return
					}
					channel = nextChannel
					continue
				}
				if errors.IsCode(err, errors.ErrCodeAborted) {
					channelLogger.InfoContext(ctx, "操作终止")
					channel.MarkSuccess(ctx)
				}
				channelLogger.ErrorContext(ctx, "流处理失败", "error", err)
				channel.MarkFailure(ctx, err)
				p.sendNativeCompatGeminiStreamErrorEvent(outputStream, err)
				return
			}

			channel.MarkSuccess(ctx)
			channelLogger.InfoContext(ctx, "流处理成功")

			for contractEvent := range contractStream {
				nativeEvent, convertErr := geminiConverter.StreamEventFromContract(contractEvent)
				if convertErr != nil {
					compatLogger.ErrorContext(ctx, "转换流事件失败", "error", convertErr)
					p.sendNativeCompatGeminiStreamErrorEvent(outputStream, convertErr)
					close(done)
					return
				}
				if nativeEvent == nil {
					continue
				}
				select {
				case <-ctx.Done():
					close(done)
					return
				case outputStream <- nativeEvent:
				}
			}

			close(done)
			return
		}
	}()

	return outputStream
}

// doCompatContractNonStream 在默认通道上执行 Contract 非流式请求，并复用重试逻辑。
func (p *Portal) doCompatContractNonStream(
	ctx context.Context,
	contractReq *adapterTypes.RequestContract,
	channel *routing.Channel,
) (*adapterTypes.ResponseContract, error) {
	var response *adapterTypes.ResponseContract

	for {
		channelLogger := p.logger.WithGroup("native_compat").With(
			"request_mode", "compat",
			"platform_id", channel.PlatformID,
			"model_id", channel.ModelID,
			"api_key_id", channel.APIKeyID,
		)

		err := p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) error {
			defer reqCancel()
			var callErr error
			response, callErr = p.request.ChatCompletion(reqCtx, contractReq, channel)
			return callErr
		})

		if err != nil {
			if errors.IsRetryable(err) {
				channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
				channel.MarkFailure(ctx, err)
				nextChannel, routeErr := p.routing.GetChannel(ctx, contractReq.Model)
				if routeErr != nil {
					return nil, routeErr
				}
				channel = nextChannel
				continue
			}
			if errors.IsCode(err, errors.ErrCodeAborted) {
				channelLogger.InfoContext(ctx, "操作终止")
				channel.MarkSuccess(ctx)
			}
			channelLogger.ErrorContext(ctx, "请求处理失败", "error", err)
			channel.MarkFailure(ctx, err)
			return nil, err
		}

		channel.MarkSuccess(ctx)
		channelLogger.InfoContext(ctx, "请求处理成功")
		return response, nil
	}
}

// sendNativeCompatOpenAIChatStreamErrorEvent 发送 OpenAI Chat 兼容流式错误事件。
func (p *Portal) sendNativeCompatOpenAIChatStreamErrorEvent(outputStream chan<- *openaiChat.StreamEvent, err error) {
	message := errors.GetMessage(err)
	if message == "" && err != nil {
		message = err.Error()
	}

	errMsg := message
	event := &openaiChat.StreamEvent{
		Choices: []openaiChat.StreamChoice{
			{
				Index: 0,
				Delta: openaiChat.Delta{Refusal: &errMsg},
			},
		},
	}

	select {
	case outputStream <- event:
	default:
	}
}

// sendNativeCompatOpenAIResponsesStreamErrorEvent 发送 OpenAI Responses 兼容流式错误事件。
func (p *Portal) sendNativeCompatOpenAIResponsesStreamErrorEvent(outputStream chan<- *openaiResponses.StreamEvent, err error) {
	message := errors.GetMessage(err)
	if message == "" && err != nil {
		message = err.Error()
	}

	statusCode := errors.GetHTTPStatus(err)
	codeValue := ""
	if statusCode > 0 {
		codeValue = strconv.Itoa(statusCode)
	}

	event := &openaiResponses.StreamEvent{
		Error: &openaiResponses.ResponseErrorEvent{
			Type:    openaiResponses.StreamEventError,
			Message: message,
		},
	}
	if codeValue != "" {
		event.Error.Code = &codeValue
	}

	select {
	case outputStream <- event:
	default:
	}
}

// sendNativeCompatAnthropicStreamErrorEvent 发送 Anthropic 兼容流式错误事件。
func (p *Portal) sendNativeCompatAnthropicStreamErrorEvent(outputStream chan<- *anthropicTypes.StreamEvent, err error) {
	message := errors.GetMessage(err)
	if message == "" && err != nil {
		message = err.Error()
	}

	code := string(errors.GetCode(err))
	if code == "" {
		code = "stream_error"
	}

	event := &anthropicTypes.StreamEvent{
		Error: &anthropicTypes.ErrorEvent{
			Type: anthropicTypes.StreamEventError,
			Error: anthropicTypes.ErrorResponse{
				Type: "error",
				Error: anthropicTypes.Error{
					Type:    code,
					Message: message,
				},
			},
		},
	}

	select {
	case outputStream <- event:
	default:
	}
}

// sendNativeCompatGeminiStreamErrorEvent 发送 Gemini 兼容流式错误事件。
func (p *Portal) sendNativeCompatGeminiStreamErrorEvent(outputStream chan<- *geminiTypes.StreamEvent, err error) {
	message := errors.GetMessage(err)
	if message == "" && err != nil {
		message = err.Error()
	}

	code := errors.GetHTTPStatus(err)
	if code <= 0 {
		code = 500
	}

	status := string(errors.GetCode(err))
	if status == "" {
		status = "STREAM_ERROR"
	}

	p.logger.WithGroup("native_compat").Error("Gemini 兼容流式请求失败", "message", message, "code", code, "status", status)

	select {
	case outputStream <- &geminiTypes.StreamEvent{}:
	default:
	}
}
