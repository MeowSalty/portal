package adapter

import (
	"encoding/json"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/request/adapter/anthropic/converter"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
)

// Anthropic Anthropic 提供商实现
type Anthropic struct {
	logger logger.Logger
}

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
func (p *Anthropic) CreateRequest(request *adapterTypes.RequestContract, channel *routing.Channel) (interface{}, error) {
	request.Model = channel.ModelName
	return converter.RequestFromContract(request)
}

// ParseResponse 解析 Anthropic 响应
func (p *Anthropic) ParseResponse(responseData []byte) (*adapterTypes.ResponseContract, error) {
	// 首先检查是否是错误响应
	var errorResp anthropicTypes.ErrorResponse
	if err := json.Unmarshal(responseData, &errorResp); err == nil && errorResp.Type == "error" {
		return &adapterTypes.ResponseContract{
			Error: &adapterTypes.ResponseError{
				Code:    &errorResp.Error.Type,
				Message: &errorResp.Error.Message,
				Type:    &errorResp.Error.Type,
			},
		}, nil
	}

	var response anthropicTypes.Response
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, err
	}
	return converter.ResponseToContract(&response, p.logger)
}

// ParseStreamResponse 解析 Anthropic 流式响应
func (p *Anthropic) ParseStreamResponse(ctx adapterTypes.StreamIndexContext, responseData []byte) ([]*adapterTypes.StreamEventContract, error) {
	var event anthropicTypes.StreamEvent
	if err := json.Unmarshal(responseData, &event); err != nil {
		return nil, err
	}

	converted, err := converter.StreamEventToContract(&event, ctx, p.logger)
	if err != nil {
		return nil, err
	}
	if converted == nil {
		return nil, nil
	}
	return []*adapterTypes.StreamEventContract{converted}, nil
}

// APIEndpoint 返回 API 端点
func (p *Anthropic) APIEndpoint(model string, stream bool, config ...string) string {
	defaultEndpoint := "/v1/messages"

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

// SupportsNative 返回是否支持原生 API 调用
func (p *Anthropic) SupportsNative() bool {
	return true
}

// BuildNativeRequest 构建原生请求
func (p *Anthropic) BuildNativeRequest(channel *routing.Channel, payload any) (body any, err error) {
	if req, ok := payload.(*anthropicTypes.Request); ok {
		req.Model = channel.ModelName
		return req, nil
	}
	return nil, errors.New(errors.ErrCodeInvalidArgument, "无效的请求类型，期望 anthropicTypes.Request")
}

// ParseNativeResponse 解析原生响应
func (p *Anthropic) ParseNativeResponse(variant string, raw []byte) (any, error) {
	var response anthropicTypes.Response
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "解析 Anthropic 响应失败", err)
	}
	return &response, nil
}

// ParseNativeStreamEvent 解析原生流事件
func (p *Anthropic) ParseNativeStreamEvent(variant string, raw []byte) (any, error) {
	var event anthropicTypes.StreamEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "解析 Anthropic 流事件失败", err)
	}
	return &event, nil
}

// ExtractUsageFromNativeStreamEvent 从原生流事件中提取使用统计信息
func (p *Anthropic) ExtractUsageFromNativeStreamEvent(variant string, event any) *adapterTypes.ResponseUsage {
	streamEvent, ok := event.(*anthropicTypes.StreamEvent)
	if !ok {
		return nil
	}
	if streamEvent.MessageDelta == nil || streamEvent.MessageDelta.Usage == nil {
		return nil
	}
	usage := streamEvent.MessageDelta.Usage
	if usage.InputTokens == nil || usage.OutputTokens == nil {
		return nil
	}
	totalTokens := *usage.InputTokens + *usage.OutputTokens
	return &adapterTypes.ResponseUsage{
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		TotalTokens:  &totalTokens,
	}
}
