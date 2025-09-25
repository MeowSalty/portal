package openai

import (
	"encoding/json"

	"github.com/MeowSalty/portal/adapter/openai/types"
	coreTypes "github.com/MeowSalty/portal/types"
)

// OpenAIRequestConverter 将核心请求转换为 OpenAI 请求
type OpenAIRequestConverter struct{}

// ConvertRequest 将核心请求转换为 OpenAI 请求
func (c *OpenAIRequestConverter) ConvertRequest(request *coreTypes.Request, channel *coreTypes.Channel) (interface{}, error) {
	openAIReq := &types.ChatCompletionRequest{
		Model:    channel.Model.Name,
		Messages: make([]types.RequestMessage, len(request.Messages)),
	}

	// 处理流参数
	if request.Stream != nil {
		openAIReq.Stream = *request.Stream
	}

	// 处理温度参数
	if request.Temperature != nil {
		openAIReq.Temperature = *request.Temperature
	}

	// 处理 TopP 参数
	if request.TopP != nil {
		openAIReq.TopP = *request.TopP
	}

	for i, msg := range request.Messages {
		// 直接使用 msg.Content，可以是字符串或数组
		openAIReq.Messages[i] = types.RequestMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return openAIReq, nil
}

// SetStream 在 OpenAI 请求上设置流标志
func (c *OpenAIRequestConverter) SetStream(req interface{}, stream bool) {
	if openAIReq, ok := req.(*types.ChatCompletionRequest); ok {
		openAIReq.Stream = stream
	}
}

// OpenAIResponseConverter 将 OpenAI 响应转换为核心响应
type OpenAIResponseConverter struct{}

// ConvertResponse 将 OpenAI 响应转换为核心响应
func (c *OpenAIResponseConverter) ConvertResponse(responseData []byte) (*coreTypes.Response, error) {
	var openAIResp types.ChatCompletionResponse
	if err := json.Unmarshal(responseData, &openAIResp); err != nil {
		return nil, err
	}

	// 将 OpenAI 响应转换为核心响应格式
	coreResp := ChatCompletionResponseToResponse(&openAIResp)
	return coreResp, nil
}

// ConvertStreamResponse 将 OpenAI 流响应转换为核心响应
func (c *OpenAIResponseConverter) ConvertStreamResponse(responseData []byte) (*coreTypes.Response, error) {
	var chunk types.ChatCompletionResponse
	if err := json.Unmarshal(responseData, &chunk); err != nil {
		return nil, err
	}

	// 将数据块转换为核心格式
	coreChunk := ChatCompletionResponseToResponse(&chunk)
	return coreChunk, nil
}
