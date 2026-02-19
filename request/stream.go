package request

import (
	"context"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
	"github.com/valyala/fasthttp"
)

// RequestLogHooks 实现 StreamHooks 接口，用于记录流式响应的统计指标到 RequestLog。
//
// 该结构体持有对 RequestLog 的引用，在流式响应的生命周期各个关键时机更新统计字段。
type RequestLogHooks struct {
	// log 指向需要更新的请求日志记录
	log *RequestLog
}

// OnFirstChunk 在第一次解析出有效事件时触发。
//
// 该方法计算并记录首字时间（Time to First Token, TTFT），这是衡量流式响应性能的重要指标。
//
// 参数 t 表示首次接收到有效内容的时间戳。
func (h *RequestLogHooks) OnFirstChunk(t time.Time) {
	if h.log == nil {
		return
	}

	// 计算首字时间
	elapsed := t.Sub(h.log.Timestamp)
	h.log.FirstByteTime = &elapsed
}

// OnUsage 当流中出现 usage 信息时触发。
//
// 该方法记录 token 使用量统计到 RequestLog 的对应字段。
//
// 参数 u 包含输入、输出和总 token 数量。
func (h *RequestLogHooks) OnUsage(u types.Usage) {
	if h.log == nil {
		return
	}

	// 记录 token 用量
	h.log.PromptTokens = &u.InputTokens
	h.log.CompletionTokens = &u.OutputTokens
	h.log.TotalTokens = &u.TotalTokens
}

// OnComplete 在流正常结束时触发。
//
// 该方法计算并记录流的总耗时。
//
// 参数 end 表示流结束的时间戳。
func (h *RequestLogHooks) OnComplete(end time.Time) {
	if h.log == nil {
		return
	}

	// 计算总耗时
	h.log.Duration = end.Sub(h.log.Timestamp)
}

// OnError 在流异常结束时触发。
//
// 该方法作为异常结束的兜底处理，计算总耗时并记录错误信息。
//
// 参数 err 表示导致流异常的错误信息。
func (h *RequestLogHooks) OnError(err error) {
	if h.log == nil {
		return
	}

	// 计算总耗时
	h.log.Duration = time.Since(h.log.Timestamp)

	// 记录错误信息
	if err != nil {
		errMsg := err.Error()
		h.log.ErrorMsg = &errMsg
	}
}

// ChatCompletionStream 处理流式聊天完成请求
//
// 该方法负责处理单个通道的流式请求，包括：
// - 获取并验证适配器
// - 创建请求日志
// - 初始化流连接
// - 处理流数据
func (p *Request) ChatCompletionStream(
	ctx context.Context,
	request *types.RequestContract,
	output chan<- *types.StreamEventContract,
	channel *routing.Channel,
) error {
	// 创建带有请求上下文的日志记录器
	log := p.logger.With(
		"platform_type", channel.Provider,
		"platform_id", channel.PlatformID,
		"model_id", channel.ModelID,
		"api_key_id", channel.APIKeyID,
		"model_name", channel.ModelName,
		"original_model", request.Model,
	)

	log.DebugContext(ctx, "开始处理流式聊天完成请求")

	// 获取适配器
	log.DebugContext(ctx, "获取适配器", "format", channel.Provider)
	adapter, err := p.getAdapter(channel.Provider)
	if err != nil {
		log.ErrorContext(ctx, "获取适配器失败", "error", err, "format", channel.Provider)
		return errors.Wrap(errors.ErrCodeAdapterNotFound, "获取适配器失败", err).
			WithHTTPStatus(fasthttp.StatusInternalServerError).
			WithContext("format", channel.Provider)
	}
	log.DebugContext(ctx, "获取适配器成功", "adapter", adapter.Name())

	// 创建请求日志
	now := time.Now()
	requestLog := &RequestLog{
		Timestamp:         now,
		RequestType:       "stream",
		ModelName:         channel.ModelName,
		OriginalModelName: request.Model,
		PlatformID:        channel.PlatformID,
		APIKeyID:          channel.APIKeyID,
		ModelID:           channel.ModelID,
	}
	log.DebugContext(ctx, "创建请求日志")

	// 创建 RequestLogHooks 实例用于记录流式响应统计
	hooks := &RequestLogHooks{log: requestLog}

	// 创建内部流
	log.DebugContext(ctx, "创建内部流通道")
	internalStream := make(chan *types.StreamEventContract, 1024)

	log.DebugContext(ctx, "执行流式聊天完成请求")
	err = adapter.ChatCompletionStream(ctx, request, channel, internalStream)
	if err != nil {
		errorMsg := err.(*errors.Error).Error()
		requestLog.ErrorMsg = &errorMsg
		p.recordRequestLog(requestLog, nil, false)

		log.ErrorContext(ctx, "流式聊天完成请求失败", "error", err)
		return err
	}

	// 处理流数据
	log.DebugContext(ctx, "开始处理流数据")
	return p.handleStreamData(ctx, internalStream, output, requestLog, hooks)
}

// handleStreamData 处理流数据
func (p *Request) handleStreamData(
	ctx context.Context,
	input <-chan *types.StreamEventContract,
	output chan<- *types.StreamEventContract,
	requestLog *RequestLog,
	hooks types.StreamHooks,
) error {
	log := p.logger.With(
		"platform_id", requestLog.PlatformID,
		"model_id", requestLog.ModelID,
		"api_key_id", requestLog.APIKeyID,
	)

	log.DebugContext(ctx, "开始处理流数据")

	firstByteRecorded := false
	var firstByteTime *time.Time
	messageCount := 0

	for response := range input {
		messageCount++

		// 检查错误
		if err := p.checkResponseError(response); err != nil {
			var portalError *errors.Error
			if errors.As(err, &portalError) {
				switch errors.GetCode(portalError) {
				case errors.ErrCodeEmptyResponse:
					log.ErrorContext(ctx, "响应中 Choices 为空", "error", portalError.Error())
				case errors.ErrCodeStreamError:
					log.ErrorContext(ctx, "流处理错误", "error", portalError.Error())
				default:
					// 兜底：处理其他错误码
					log.ErrorContext(ctx, "流数据错误", "error", portalError.Error(), "code", portalError.Code)
				}
			}

			msg := err.Error()
			requestLog.ErrorMsg = &msg
			p.recordRequestLog(requestLog, nil, false)

			return err
		}

		// 记录首字节时间并触发 OnFirstChunk Hook
		if !firstByteRecorded {
			now := time.Now()
			firstByteTime = &now
			firstByteRecorded = true

			firstByteDuration := now.Sub(requestLog.Timestamp)
			log.DebugContext(ctx, "收到首字节",
				"first_byte_time", firstByteDuration.String(),
			)

			// 触发 OnFirstChunk Hook
			if hooks != nil {
				hooks.OnFirstChunk(now)
			}
		}

		// 记录 Token 用量并触发 OnUsage Hook
		if response.Usage != nil {
			requestLog.CompletionTokens = response.Usage.OutputTokens
			requestLog.PromptTokens = response.Usage.InputTokens
			requestLog.TotalTokens = response.Usage.TotalTokens

			log.DebugContext(ctx, "更新 Token 使用情况",
				"prompt_tokens", response.Usage.InputTokens,
				"completion_tokens", response.Usage.OutputTokens,
				"total_tokens", response.Usage.TotalTokens,
			)

			// 触发 OnUsage Hook
			if hooks != nil && response.Usage.TotalTokens != nil {
				hooks.OnUsage(types.Usage{
					InputTokens:  getIntValue(response.Usage.InputTokens),
					OutputTokens: getIntValue(response.Usage.OutputTokens),
					TotalTokens:  *response.Usage.TotalTokens,
				})
			}
		}

		// 发送响应
		if err := p.sendResponse(ctx, output, response, requestLog); err != nil {
			log.ErrorContext(ctx, "发送响应失败",
				"error", err,
				"message_count", messageCount,
			)
			return err
		}

		// 检查错误 - 触发 OnError Hook
		if err := p.checkResponseError(response); err != nil {
			var portalError *errors.Error
			if errors.As(err, &portalError) {
				switch errors.GetCode(portalError) {
				case errors.ErrCodeEmptyResponse:
					log.ErrorContext(ctx, "响应中 Choices 为空", "error", portalError.Error())
				case errors.ErrCodeStreamError:
					log.ErrorContext(ctx, "流处理错误", "error", portalError.Error())
				default:
					// 兜底：处理其他错误码
					log.ErrorContext(ctx, "流数据错误", "error", portalError.Error(), "code", portalError.Code)
				}
			}

			// 触发 OnError Hook
			if hooks != nil {
				hooks.OnError(err)
			}

			msg := err.Error()
			requestLog.ErrorMsg = &msg
			p.recordRequestLog(requestLog, nil, false)

			return err
		}
	}

	// 流成功完成 - 触发 OnComplete Hook
	log.DebugContext(ctx, "流数据处理完成",
		"total_messages", messageCount,
		"duration", time.Since(requestLog.Timestamp).String(),
	)

	// 触发 OnComplete Hook
	if hooks != nil {
		hooks.OnComplete(time.Now())
	}

	close(output)
	p.recordRequestLog(requestLog, firstByteTime, true)

	log.InfoContext(ctx, "流式请求成功完成")
	return nil
}

// getIntValue 从指针获取 int 值，若为 nil 则返回 0
func getIntValue(ptr *int) int {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// sendResponse 发送响应到输出通道
func (p *Request) sendResponse(
	ctx context.Context,
	output chan<- *types.StreamEventContract,
	response *types.StreamEventContract,
	requestLog *RequestLog,
) error {
	log := p.logger.With(
		"platform_id", requestLog.PlatformID,
		"model_id", requestLog.ModelID,
		"api_key_id", requestLog.APIKeyID,
	)

	select {
	case output <- response:
		log.DebugContext(ctx, "响应发送成功")
		return nil
	case <-ctx.Done():
		err := errors.Wrap(errors.ErrCodeAborted, "连接被终止", ctx.Err())

		log.WarnContext(ctx, "连接被终止", "error", ctx.Err())
		return err
	}
}
