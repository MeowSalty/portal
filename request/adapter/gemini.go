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
func (p *Gemini) ParseResponse(responseData []byte) (*adapterTypes.ResponseContract, error) {
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
func (p *Gemini) ParseStreamResponse(ctx adapterTypes.StreamIndexContext, responseData []byte) ([]*adapterTypes.StreamEventContract, error) {
	var event geminiTypes.StreamEvent
	if err := json.Unmarshal(responseData, &event); err != nil {
		return nil, err
	}
	return converter.StreamEventToContract(&event, ctx, nil)
}

// APIEndpoint 返回 API 端点
func (p *Gemini) APIEndpoint(model string, stream bool, config ...string) string {
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
