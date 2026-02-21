package adapter

import (
	"encoding/json"
	"strings"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	chatConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/chat"
	responsesConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/responses"
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
)

// OpenAI OpenAI 提供商实现
type OpenAI struct {
	apiVariant string
	logger     logger.Logger
}

// init 函数注册 OpenAI 提供商
func init() {
	RegisterProviderFactory("openai", func() Provider {
		return NewOpenAIProvider()
	})
}

// NewOpenAIProvider 创建新的 OpenAI 提供商
func NewOpenAIProvider() *OpenAI {
	return &OpenAI{}
}

// Name 返回提供商名称
func (p *OpenAI) Name() string {
	return "openai"
}

// CreateRequest 创建 OpenAI 请求
func (p *OpenAI) CreateRequest(request *adapterTypes.RequestContract, channel *routing.Channel) (interface{}, error) {
	request.Model = channel.ModelName
	style := p.setAPIStyle(channel)
	if style == "responses" {
		return responsesConverter.RequestFromContract(request)
	}
	return chatConverter.RequestFromContract(request)
}

// ParseResponse 解析 OpenAI 响应
func (p *OpenAI) ParseResponse(responseData []byte) (*adapterTypes.ResponseContract, error) {
	if p.apiVariant == "responses" {
		var response openaiResponses.Response
		if err := json.Unmarshal(responseData, &response); err != nil {
			return nil, err
		}
		return responsesConverter.ResponseToContract(&response, p.logger)
	}

	var response openaiChat.Response
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, err
	}
	return chatConverter.ResponseToContract(&response, p.logger)
}

// ParseStreamResponse 解析 OpenAI 流式响应
func (p *OpenAI) ParseStreamResponse(ctx adapterTypes.StreamIndexContext, responseData []byte) ([]*adapterTypes.StreamEventContract, error) {
	if p.apiVariant == "responses" {
		var event openaiResponses.StreamEvent
		if err := json.Unmarshal(responseData, &event); err != nil {
			return nil, err
		}
		converted, err := responsesConverter.StreamEventToContract(&event, nil)
		if err != nil {
			return nil, err
		}
		if converted == nil {
			return nil, nil
		}
		return []*adapterTypes.StreamEventContract{converted}, nil
	}

	var chunk openaiChat.StreamEvent
	if err := json.Unmarshal(responseData, &chunk); err != nil {
		return nil, err
	}
	return chatConverter.StreamEventToContract(&chunk, nil)
}

// APIEndpoint 返回 API 端点
func (p *OpenAI) APIEndpoint(model string, stream bool, config ...string) string {
	// 默认端点
	var defaultEndpoint string
	if p.apiVariant == "responses" {
		defaultEndpoint = "/v1/responses"
	} else {
		defaultEndpoint = "/v1/chat/completions"
	}

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
func (p *OpenAI) Headers(key string) map[string]string {
	headers := map[string]string{
		"Authorization": "Bearer " + key,
		"Content-Type":  "application/json",
	}

	return headers
}

// SupportsStreaming 是否支持流式传输
func (p *OpenAI) SupportsStreaming() bool {
	return true
}

// SupportsNative 返回是否支持原生 API 调用
func (p *OpenAI) SupportsNative() bool {
	return true
}

// BuildNativeRequest 构建原生请求
func (p *OpenAI) BuildNativeRequest(channel *routing.Channel, payload any) (body any, err error) {
	style := p.setAPIStyle(channel)

	switch style {
	case "chat", "chat_completions":
		if req, ok := payload.(*openaiChat.Request); ok {
			req.Model = channel.ModelName
			return req, nil
		}
		return nil, errors.New(errors.ErrCodeInvalidArgument, "无效的请求类型，期望 openaiChat.Request")

	case "responses":
		if req, ok := payload.(*openaiResponses.Request); ok {
			if req.Model == nil {
				model := channel.ModelName
				req.Model = &model
			}
			return req, nil
		}
		return nil, errors.New(errors.ErrCodeInvalidArgument, "无效的请求类型，期望 openaiResponses.Request")

	default:
		return nil, errors.New(errors.ErrCodeInvalidArgument, "不支持的 API 变体："+style)
	}
}

// ParseNativeResponse 解析原生响应
func (p *OpenAI) ParseNativeResponse(variant string, raw []byte) (any, error) {
	switch variant {
	case "chat", "chat_completions":
		var response openaiChat.Response
		if err := json.Unmarshal(raw, &response); err != nil {
			return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "解析 OpenAI Chat 响应失败", err)
		}
		return &response, nil

	case "responses":
		var response openaiResponses.Response
		if err := json.Unmarshal(raw, &response); err != nil {
			return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "解析 OpenAI Responses 响应失败", err)
		}
		return &response, nil

	default:
		return nil, errors.New(errors.ErrCodeInvalidArgument, "不支持的 API 变体："+variant)
	}
}

// ParseNativeStreamEvent 解析原生流事件
func (p *OpenAI) ParseNativeStreamEvent(variant string, raw []byte) (any, error) {
	switch variant {
	case "chat", "chat_completions":
		var event openaiChat.StreamEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "解析 OpenAI Chat 流事件失败", err)
		}
		return &event, nil

	case "responses":
		var event openaiResponses.StreamEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "解析 OpenAI Responses 流事件失败", err)
		}
		return &event, nil

	default:
		return nil, errors.New(errors.ErrCodeInvalidArgument, "不支持的 API 变体："+variant)
	}
}

// ExtractUsageFromNativeStreamEvent 从原生流事件中提取使用统计信息
func (p *OpenAI) ExtractUsageFromNativeStreamEvent(variant string, event any) *adapterTypes.ResponseUsage {
	switch variant {
	case "chat", "chat_completions":
		chatEvent, ok := event.(*openaiChat.StreamEvent)
		if !ok {
			return nil
		}
		if chatEvent.Usage == nil {
			return nil
		}
		return &adapterTypes.ResponseUsage{
			InputTokens:  &chatEvent.Usage.PromptTokens,
			OutputTokens: &chatEvent.Usage.CompletionTokens,
			TotalTokens:  &chatEvent.Usage.TotalTokens,
		}

	case "responses":
		responsesEvent, ok := event.(*openaiResponses.StreamEvent)
		if !ok {
			return nil
		}
		if responsesEvent.Completed == nil || responsesEvent.Completed.Response.Usage == nil {
			return nil
		}
		inputTokens := responsesEvent.Completed.Response.Usage.InputTokens
		outputTokens := responsesEvent.Completed.Response.Usage.OutputTokens
		totalTokens := responsesEvent.Completed.Response.Usage.TotalTokens
		return &adapterTypes.ResponseUsage{
			InputTokens:  &inputTokens,
			OutputTokens: &outputTokens,
			TotalTokens:  &totalTokens,
		}

	default:
		return nil
	}
}

func (p *OpenAI) setAPIStyle(channel *routing.Channel) string {
	if channel == nil {
		p.apiVariant = "chat_completions"
		return p.apiVariant
	}

	style := strings.ToLower(strings.TrimSpace(channel.APIVariant))
	if style == "" {
		style = "chat_completions"
	}

	p.apiVariant = style
	return p.apiVariant
}
