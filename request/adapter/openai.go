package adapter

import (
	"encoding/json"
	"strings"

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
	RegisterProviderFactory("OpenAI", func() Provider {
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
func (p *OpenAI) APIEndpoint(model string, stream bool) string {
	if p.apiVariant == "responses" {
		return "/v1/responses"
	}
	return "/v1/chat/completions"
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
