package adapter

import (
	"encoding/json"
	"log/slog"

	converter "github.com/MeowSalty/portal/adapter/openai/converter"
	openaiTypes "github.com/MeowSalty/portal/adapter/openai/types"
	coreTypes "github.com/MeowSalty/portal/types"
)

// OpenAI OpenAI 提供商实现
type OpenAI struct {
	logger *slog.Logger
}

// init 函数注册 OpenAI 提供商
func init() {
	RegisterProviderFactory("OpenAI", func(logger *slog.Logger) Provider {
		return NewOpenAIProvider(logger)
	})
}

// NewOpenAIProvider 创建新的 OpenAI 提供商
func NewOpenAIProvider(logger *slog.Logger) *OpenAI {
	if logger == nil {
		logger = slog.Default()
	}
	return &OpenAI{
		logger: logger.WithGroup("openai"),
	}
}

// Name 返回提供商名称
func (p *OpenAI) Name() string {
	return "openai"
}

// CreateRequest 创建 OpenAI 请求
func (p *OpenAI) CreateRequest(request *coreTypes.Request, channel *coreTypes.Channel) (interface{}, error) {
	return converter.ConvertRequest(request, channel), nil
}

// ParseResponse 解析 OpenAI 响应
func (p *OpenAI) ParseResponse(responseData []byte) (*coreTypes.Response, error) {
	var response openaiTypes.Response
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, err
	}
	return converter.ConvertCoreResponse(&response), nil
}

// ParseStreamResponse 解析 OpenAI 流式响应
func (p *OpenAI) ParseStreamResponse(responseData []byte) (*coreTypes.Response, error) {
	var chunk openaiTypes.Response
	if err := json.Unmarshal(responseData, &chunk); err != nil {
		return nil, err
	}
	return converter.ConvertCoreResponse(&chunk), nil
}

// APIEndpoint 返回 API 端点
func (p *OpenAI) APIEndpoint(model string, stream bool) string {
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
