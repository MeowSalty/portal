package adapter

import (
	"encoding/json"
	"fmt"

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
	RegisterProviderFactory("Gemini", func() Provider {
		return NewGeminiProvider()
	})
}

// NewGeminiProvider 创建新的 Gemini 提供商
func NewGeminiProvider() *Gemini {
	return &Gemini{}
}

// Name 返回提供商名称
func (p *Gemini) Name() string {
	return "gemini"
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
