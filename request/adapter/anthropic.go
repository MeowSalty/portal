package adapter

import (
	"encoding/json"

	"github.com/MeowSalty/portal/request/adapter/anthropic/converter"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
)

// Anthropic Anthropic 提供商实现
type Anthropic struct{}

// init 函数注册 Anthropic 提供商
func init() {
	RegisterProviderFactory("Anthropic", func() Provider {
		return NewAnthropicProvider()
	})
}

// NewAnthropicProvider 创建新的 Anthropic 提供商
func NewAnthropicProvider() *Anthropic {
	return &Anthropic{}
}

// Name 返回提供商名称
func (p *Anthropic) Name() string {
	return "anthropic"
}

// CreateRequest 创建 Anthropic 请求
func (p *Anthropic) CreateRequest(request *coreTypes.Request, channel *routing.Channel) (interface{}, error) {
	return converter.ConvertRequest(request, channel), nil
}

// ParseResponse 解析 Anthropic 响应
func (p *Anthropic) ParseResponse(responseData []byte) (*coreTypes.Response, error) {
	// 首先检查是否是错误响应
	var errorResp anthropicTypes.ErrorResponse
	if err := json.Unmarshal(responseData, &errorResp); err == nil && errorResp.Type == "error" {
		return &coreTypes.Response{
			Choices: []coreTypes.Choice{
				{
					Error: &coreTypes.ErrorResponse{
						Code:    500,
						Message: errorResp.Error.Message,
						Metadata: map[string]interface{}{
							"type": errorResp.Error.Type,
						},
					},
				},
			},
		}, nil
	}

	var response anthropicTypes.Response
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, err
	}
	return response.ConvertCoreResponse(), nil
}

// ParseStreamResponse 解析 Anthropic 流式响应
func (p *Anthropic) ParseStreamResponse(responseData []byte) (*coreTypes.Response, error) {
	var event anthropicTypes.StreamEvent
	if err := json.Unmarshal(responseData, &event); err != nil {
		return nil, err
	}

	// 检查是否是错误事件
	if event.Type == "error" {
		var errorResp anthropicTypes.ErrorResponse
		if err := json.Unmarshal(responseData, &errorResp); err == nil {
			return &coreTypes.Response{
				Choices: []coreTypes.Choice{
					{
						Error: &coreTypes.ErrorResponse{
							Code:    500,
							Message: errorResp.Error.Message,
							Metadata: map[string]interface{}{
								"type": errorResp.Error.Type,
							},
						},
					},
				},
			}, nil
		}
	}

	return event.ConvertCoreResponse(), nil
}

// APIEndpoint 返回 API 端点
func (p *Anthropic) APIEndpoint(model string, stream bool) string {
	return "/v1/messages"
}

// Headers 返回特定头部
func (p *Anthropic) Headers(key string) map[string]string {
	headers := map[string]string{
		"x-api-key":         key,
		"anthropic-version": "2023-06-01",
		"Content-Type":      "application/json",
	}

	return headers
}

// SupportsStreaming 是否支持流式传输
func (p *Anthropic) SupportsStreaming() bool {
	return true
}
