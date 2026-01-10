package responses

import (
	openaiChatConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/chat"
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ConvertResponsesRequest 将核心请求转换为 Responses API 请求
func ConvertResponsesRequest(request *coreTypes.Request, channel *routing.Channel) interface{} {
	if request == nil {
		return &openaiResponses.Request{}
	}

	respReq := &openaiResponses.Request{
		Model: channel.ModelName,
	}

	// 处理 input
	respReq.Input = convertResponsesInput(request)

	// 处理流参数
	if request.Stream != nil {
		respReq.Stream = request.Stream
	}

	if request.MaxTokens != nil {
		respReq.MaxOutputTokens = request.MaxTokens
	}

	if request.Temperature != nil {
		respReq.Temperature = request.Temperature
	}

	if request.TopP != nil {
		respReq.TopP = request.TopP
	}

	// 复用 OpenAI 的工具字段结构
	if len(request.Tools) > 0 || request.ToolChoice.Mode != nil || request.ToolChoice.Function != nil {
		openaiReq := openaiChatConverter.ConvertRequest(request, channel).(*openaiChat.Request)
		respReq.Tools = openaiReq.Tools
		respReq.ToolChoice = openaiReq.ToolChoice
	}

	// 透传额外字段
	if request.ExtraFieldsFormat != nil && *request.ExtraFieldsFormat == "openai" && len(request.ExtraFields) > 0 {
		respReq.ExtraFields = make(map[string]interface{})
		for key, value := range request.ExtraFields {
			respReq.ExtraFields[key] = value
		}
	}

	return respReq
}

func convertResponsesInput(request *coreTypes.Request) interface{} {
	// 如果是 prompt 形式，直接传字符串
	if request.Prompt != nil {
		return *request.Prompt
	}

	// 否则转换 messages
	if len(request.Messages) == 0 {
		return []openaiResponses.InputItem{}
	}

	items := make([]openaiResponses.InputItem, 0, len(request.Messages))
	for _, msg := range request.Messages {
		item := openaiResponses.InputItem{
			Role:    msg.Role,
			Content: make([]openaiResponses.InputPart, 0),
		}

		if msg.Content.StringValue != nil {
			item.Content = append(item.Content, openaiResponses.InputPart{
				Type: "input_text",
				Text: *msg.Content.StringValue,
			})
		} else if len(msg.Content.ContentParts) > 0 {
			for _, part := range msg.Content.ContentParts {
				if part.Text != nil {
					item.Content = append(item.Content, openaiResponses.InputPart{
						Type: "input_text",
						Text: *part.Text,
					})
				} else if part.ImageURL != nil {
					item.Content = append(item.Content, openaiResponses.InputPart{
						Type: "input_image",
						ImageURL: &openaiResponses.ImageURL{
							URL:    part.ImageURL.URL,
							Detail: part.ImageURL.Detail,
						},
					})
				}
			}
		}

		items = append(items, item)
	}

	return items
}
