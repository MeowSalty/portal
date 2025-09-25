package adapter

import (
	"log/slog"

	"github.com/MeowSalty/portal/adapter/openai"
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
	converter := &openai.OpenAIRequestConverter{}
	return converter.ConvertRequest(request, channel)
}

// ParseResponse 解析 OpenAI 响应
func (p *OpenAI) ParseResponse(responseData []byte) (*coreTypes.Response, error) {
	converter := &openai.OpenAIResponseConverter{}
	return converter.ConvertResponse(responseData)
}

// ParseStreamResponse 解析 OpenAI 流式响应
func (p *OpenAI) ParseStreamResponse(responseData []byte) (*coreTypes.Response, error) {
	converter := &openai.OpenAIResponseConverter{}
	return converter.ConvertStreamResponse(responseData)
}

// APIEndpoint 返回 API 端点
func (p *OpenAI) APIEndpoint() string {
	return "/v1/chat/completions"
}

// Headers 返回特定头部
// OpenAI-Beta: assistants=v2 用于启用助手 API v2 的 beta 功能
// 这是一个可选头部，仅在需要助手 API 特定功能时使用
func (p *OpenAI) Headers() map[string]string {
	return map[string]string{
		// "OpenAI-Beta": "assistants=v2", // 可选：启用助手 API beta 功能
	}
}

// SupportsStreaming 是否支持流式传输
func (p *OpenAI) SupportsStreaming() bool {
	return true
}
