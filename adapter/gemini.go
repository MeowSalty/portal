package adapter

import (
	"encoding/json"
	"log/slog"

	"github.com/MeowSalty/portal/adapter/gemini/converter"
	geminiTypes "github.com/MeowSalty/portal/adapter/gemini/types"
	coreTypes "github.com/MeowSalty/portal/types"
)

// Gemini Gemini 提供商实现
type Gemini struct {
	logger *slog.Logger
}

// init 函数注册 Gemini 提供商
func init() {
	RegisterProviderFactory("Gemini", func(logger *slog.Logger) Provider {
		return NewGeminiProvider(logger)
	})
}

// NewGeminiProvider 创建新的 Gemini 提供商
func NewGeminiProvider(logger *slog.Logger) *Gemini {
	if logger == nil {
		logger = slog.Default()
	}
	return &Gemini{
		logger: logger.WithGroup("gemini"),
	}
}

// Name 返回提供商名称
func (p *Gemini) Name() string {
	return "gemini"
}

// CreateRequest 创建 Gemini 请求
func (p *Gemini) CreateRequest(request *coreTypes.Request, channel *coreTypes.Channel) (interface{}, error) {
	return converter.ConvertRequest(request, channel), nil
}

// ParseResponse 解析 Gemini 响应
func (p *Gemini) ParseResponse(responseData []byte) (*coreTypes.Response, error) {
	// 首先检查是否是错误响应
	var errorResp geminiTypes.ErrorResponse
	if err := json.Unmarshal(responseData, &errorResp); err == nil && errorResp.Error.Code != 0 {
		return &coreTypes.Response{
			Choices: []coreTypes.Choice{
				{
					Error: &coreTypes.ErrorResponse{
						Code:    errorResp.Error.Code,
						Message: errorResp.Error.Message,
					},
				},
			},
		}, nil
	}

	var response geminiTypes.Response
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, err
	}
	return converter.ConvertCoreResponse(&response), nil
}

// ParseStreamResponse 解析 Gemini 流式响应
func (p *Gemini) ParseStreamResponse(responseData []byte) (*coreTypes.Response, error) {
	// 流式响应与普通响应格式相同
	return p.ParseResponse(responseData)
}

// APIEndpoint 返回 API 端点
func (p *Gemini) APIEndpoint(model string, stream bool) string {
	// 移除模型名的前缀"models/"（如果存在的话）
	if len(model) > 7 && model[:7] == "models/" {
		model = model[7:]
	}

	// 模型名称在 URL 路径中指定
	return "/v1beta/models/" + model + ":" + (func() string {
		if stream {
			return "streamGenerateContent?alt=sse"
		} else {
			return "generateContent"
		}
	})()
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
