package adapter

import (
	"encoding/json"
	"fmt"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/request/adapter/gemini/converter"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
)

// Gemini Gemini 提供商实现
type Gemini struct {
	logger logger.Logger
}

// init 函数注册 Gemini 提供商
func init() {
	RegisterProviderFactory("google", func() Provider {
		return NewGeminiProvider()
	})
}

// NewGeminiProvider 创建新的 Gemini 提供商
func NewGeminiProvider() *Gemini {
	return &Gemini{}
}

// Name 返回提供商名称
func (p *Gemini) Name() string {
	return "google"
}

// CreateRequest 创建 Gemini 请求
func (p *Gemini) CreateRequest(request *adapterTypes.RequestContract, channel *routing.Channel) (interface{}, error) {
	return converter.FromContract(request)
}

// ParseResponse 解析 Gemini 响应
func (p *Gemini) ParseResponse(variant string, responseData []byte) (*adapterTypes.ResponseContract, error) {
	// 首先检查是否是错误响应
	var errorResp geminiTypes.ErrorResponse
	if err := json.Unmarshal(responseData, &errorResp); err == nil && errorResp.Error.Code != 0 {
		code := fmt.Sprintf("%d", errorResp.Error.Code)
		respErr := &adapterTypes.ResponseError{
			Code:    &code,
			Message: &errorResp.Error.Message,
			Type:    &errorResp.Error.Status,
		}
		if len(errorResp.Error.Details) > 0 {
			respErr.Extras = map[string]interface{}{
				"details": errorResp.Error.Details,
			}
		}
		return &adapterTypes.ResponseContract{Error: respErr}, nil
	}

	var response geminiTypes.Response
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, err
	}
	return converter.ResponseToContract(&response, p.logger)
}

// ParseStreamResponse 解析 Gemini 流式响应
func (p *Gemini) ParseStreamResponse(variant string, ctx adapterTypes.StreamIndexContext, responseData []byte) ([]*adapterTypes.StreamEventContract, error) {
	var event geminiTypes.StreamEvent
	if err := json.Unmarshal(responseData, &event); err != nil {
		return nil, err
	}
	return converter.StreamEventToContract(&event, ctx, nil)
}

// APIEndpoint 返回 API 端点
func (p *Gemini) APIEndpoint(variant string, model string, stream bool, config ...string) string {
	// 移除模型名的前缀"models/"（如果存在的话）
	if len(model) > 7 && model[:7] == "models/" {
		model = model[7:]
	}

	// 默认端点
	defaultEndpoint := "/v1beta/models/" + model + ":" + (func() string {
		if stream {
			return "streamGenerateContent?alt=sse"
		} else {
			return "generateContent"
		}
	})()

	// 如果没有提供 config，使用默认端点
	if len(config) == 0 || config[0] == "" {
		return defaultEndpoint
	}

	c := config[0]

	// 如果 config 以 "/" 结尾，视为前缀，拼接默认端点
	if len(c) > 0 && c[len(c)-1] == '/' {
		return c + defaultEndpoint
	}

	// 其他情况，视为完整路径
	return c
}

// Headers 返回特定头部
func (p *Gemini) Headers(key string) map[string]string {
	headers := map[string]string{
		"x-goog-api-key": key,
		"Content-Type":   "application/json",
	}

	return headers
}

// SupportsStreaming 是否支持流式传输
func (p *Gemini) SupportsStreaming() bool {
	return true
}

// SupportsNative 返回是否支持原生 API 调用
func (p *Gemini) SupportsNative() bool {
	return true
}

// BuildNativeRequest 构建原生请求
func (p *Gemini) BuildNativeRequest(channel *routing.Channel, payload any) (body any, err error) {
	if req, ok := payload.(*geminiTypes.Request); ok {
		// 确保模型名称设置正确
		if req.Model == "" {
			req.Model = channel.ModelName
		}
		return req, nil
	}
	return nil, errors.New(errors.ErrCodeInvalidArgument, "无效的请求类型，期望 geminiTypes.Request")
}

// ParseNativeResponse 解析原生响应
func (p *Gemini) ParseNativeResponse(variant string, raw []byte) (any, error) {
	var response geminiTypes.Response
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "解析 Gemini 响应失败", err)
	}
	return &response, nil
}

// ParseNativeStreamEvent 解析原生流事件
func (p *Gemini) ParseNativeStreamEvent(variant string, raw []byte) (any, error) {
	var event geminiTypes.StreamEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "解析 Gemini 流事件失败", err)
	}
	return &event, nil
}

// ExtractUsageFromNativeStreamEvent 从原生流事件中提取使用统计信息
func (p *Gemini) ExtractUsageFromNativeStreamEvent(variant string, event any) *adapterTypes.ResponseUsage {
	streamEvent, ok := event.(*geminiTypes.StreamEvent)
	if !ok {
		return nil
	}
	if streamEvent.UsageMetadata == nil {
		return nil
	}
	inputTokens := int(streamEvent.UsageMetadata.PromptTokenCount)
	outputTokens := int(streamEvent.UsageMetadata.CandidatesTokenCount)
	totalTokens := int(streamEvent.UsageMetadata.TotalTokenCount)
	return &adapterTypes.ResponseUsage{
		InputTokens:  &inputTokens,
		OutputTokens: &outputTokens,
		TotalTokens:  &totalTokens,
	}
}

// IdentifyStreamEventSignal 识别 Gemini 原生流事件的信号类型。
//
// Gemini 的完成信号识别规则：
//   - Candidate.FinishReason 非空且非 FINISH_REASON_UNSPECIFIED 时为完成信号
//   - FinishReason 为 "STOP" 时为正常完成
//   - FinishReason 为 "MAX_TOKENS" 时为输出截断（异常终止）
//   - FinishReason 为 "SAFETY"、"RECITATION"、"BLOCKLIST"、"PROHIBITED_CONTENT"、
//     "SPII"、"IMAGE_SAFETY"、"IMAGE_PROHIBITED_CONTENT"、"IMAGE_OTHER"、
//     "IMAGE_RECITATION"、"NO_IMAGE" 时为安全拦截（异常终止）
//   - FinishReason 为 "OTHER"、"LANGUAGE"、"MALFORMED_FUNCTION_CALL"、
//     "UNEXPECTED_TOOL_CALL"、"TOO_MANY_TOOL_CALLS"、"MISSING_THOUGHT_SIGNATURE" 时为其他异常终止
//   - PromptFeedback.BlockReason 非空时为提示级安全拦截（异常终止），
//     表示请求在生成前即被阻止
//   - Candidate.Content.Parts 非空时为有效输出
//   - UsageMetadata 与 FinishReason 同时出现时确认流完成，但 FinishReason 单独即可判定
func (p *Gemini) IdentifyStreamEventSignal(variant string, event any) StreamEventSignal {
	signal := StreamEventSignal{}

	streamEvent, ok := event.(*geminiTypes.StreamEvent)
	if !ok {
		return signal
	}

	// 检查是否有有效输出
	for _, candidate := range streamEvent.Candidates {
		// 检查内容部分
		if len(candidate.Content.Parts) > 0 {
			for _, part := range candidate.Content.Parts {
				// 文本内容
				if part.Text != nil && *part.Text != "" {
					signal.HasValidOutput = true
				}
				// 函数调用
				if part.FunctionCall != nil {
					signal.HasValidOutput = true
				}
			}
		}

		// 检查完成信号
		// FINISH_REASON_UNSPECIFIED 是默认/未知值，不视为有效完成信号。
		// 只有明确的完成原因才标记为 IsCompletionSignal。
		if candidate.FinishReason != "" && candidate.FinishReason != geminiTypes.FinishReasonUnspecified {
			signal.IsCompletionSignal = true
			signal.IsTerminalEvent = true
			signal.FinishReason = candidate.FinishReason
		}
	}

	// 检查提示反馈（安全拦截情况）
	// PromptFeedback.BlockReason 非空表示请求在生成候选之前即被阻止，
	// 这是提示级安全拦截，属于异常终止。
	// 常见值：SAFETY、BLOCKLIST、PROHIBITED_CONTENT、IMAGE_SAFETY 等。
	if streamEvent.PromptFeedback != nil && streamEvent.PromptFeedback.BlockReason != "" &&
		streamEvent.PromptFeedback.BlockReason != geminiTypes.BlockReasonUnspecified {
		signal.IsCompletionSignal = true
		signal.IsTerminalEvent = true
		signal.FinishReason = streamEvent.PromptFeedback.BlockReason
	}

	return signal
}
