package responses

import (
	"encoding/json"

	openaiChatConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/chat"
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ConvertRequest 将核心请求转换为 Responses API 请求
func ConvertRequest(request *coreTypes.Request, channel *routing.Channel) interface{} {
	if request == nil {
		return &openaiResponses.Request{}
	}

	respReq := &openaiResponses.Request{
		Model: channel.ModelName,
	}

	// 处理 input
	respReq.Input = convertInput(request)

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

// ConvertCoreRequest 将 Responses API 请求转换为核心请求
func ConvertCoreRequest(openaiReq *openaiResponses.Request) *coreTypes.Request {
	if openaiReq == nil {
		return nil
	}

	coreReq := &coreTypes.Request{
		Model: openaiReq.Model,
	}

	switch input := openaiReq.Input.(type) {
	case string:
		coreReq.Prompt = &input
	case []openaiResponses.InputItem:
		coreReq.Messages = convertInputItemsToCoreMessages(input)
	case []interface{}:
		coreReq.Messages = convertRawInputItemsToCoreMessages(input)
	}

	if openaiReq.Stream != nil {
		coreReq.Stream = openaiReq.Stream
	}

	if openaiReq.MaxOutputTokens != nil {
		coreReq.MaxTokens = openaiReq.MaxOutputTokens
	}

	if openaiReq.Temperature != nil {
		coreReq.Temperature = openaiReq.Temperature
	}

	if openaiReq.TopP != nil {
		coreReq.TopP = openaiReq.TopP
	}

	if len(openaiReq.Tools) > 0 {
		coreReq.Tools = make([]coreTypes.Tool, len(openaiReq.Tools))
		for i, tool := range openaiReq.Tools {
			if tool.Function == nil {
				continue
			}

			var parameters interface{}
			if tool.Function.Function.Parameters != nil {
				switch params := tool.Function.Function.Parameters.(type) {
				case json.RawMessage:
					var paramMap map[string]interface{}
					if err := json.Unmarshal(params, &paramMap); err == nil {
						parameters = paramMap
					}
				case []byte:
					var paramMap map[string]interface{}
					if err := json.Unmarshal(params, &paramMap); err == nil {
						parameters = paramMap
					}
				case map[string]interface{}:
					parameters = params
				default:
					parameters = params
				}
			}

			coreReq.Tools[i] = coreTypes.Tool{
				Type: "function",
				Function: coreTypes.FunctionDescription{
					Name:        tool.Function.Function.Name,
					Description: tool.Function.Function.Description,
					Parameters:  parameters,
				},
			}
		}
	}

	if openaiReq.ToolChoice != nil {
		if openaiReq.ToolChoice.Auto != nil {
			mode := *openaiReq.ToolChoice.Auto
			coreReq.ToolChoice.Mode = &mode
		} else if openaiReq.ToolChoice.Named != nil {
			coreReq.ToolChoice.Function = &coreTypes.FunctionToolChoice{
				Type: "function",
				Function: coreTypes.FunctionNameChoice{
					Name: openaiReq.ToolChoice.Named.Function.Name,
				},
			}
		}
	}

	if len(openaiReq.ExtraFields) > 0 {
		coreReq.ExtraFields = make(map[string]interface{})
		for key, value := range openaiReq.ExtraFields {
			coreReq.ExtraFields[key] = value
		}
		format := "openai"
		coreReq.ExtraFieldsFormat = &format
	}

	return coreReq
}

func convertInput(request *coreTypes.Request) interface{} {
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
					continue
				}
				if part.ImageURL != nil {
					item.Content = append(item.Content, openaiResponses.InputPart{
						Type: "input_image",
						ImageURL: &openaiResponses.ImageURL{
							URL:    part.ImageURL.URL,
							Detail: part.ImageURL.Detail,
						},
					})
					continue
				}
				if len(part.ExtraFields) > 0 {
					item.Content = append(item.Content, openaiResponses.InputPart{
						Raw: part.ExtraFields,
					})
				}
			}
		}

		items = append(items, item)
	}

	return items
}

func convertInputItemsToCoreMessages(items []openaiResponses.InputItem) []coreTypes.Message {
	messages := make([]coreTypes.Message, 0, len(items))
	for _, item := range items {
		msg := coreTypes.Message{
			Role: item.Role,
		}

		if len(item.Content) == 0 {
			messages = append(messages, msg)
			continue
		}

		if len(item.Content) == 1 && item.Content[0].Type == "input_text" && item.Content[0].ImageURL == nil {
			text := item.Content[0].Text
			msg.Content.StringValue = &text
			messages = append(messages, msg)
			continue
		}

		parts := make([]coreTypes.ContentPart, 0, len(item.Content))
		for _, part := range item.Content {
			switch part.Type {
			case "input_text":
				text := part.Text
				parts = append(parts, coreTypes.ContentPart{
					Type:        "text",
					Text:        &text,
					ExtraFields: part.Raw,
				})
			case "input_image":
				if part.ImageURL == nil {
					continue
				}
				parts = append(parts, coreTypes.ContentPart{
					Type: "image_url",
					ImageURL: &coreTypes.ImageURL{
						URL:    part.ImageURL.URL,
						Detail: part.ImageURL.Detail,
					},
					ExtraFields: part.Raw,
				})
			default:
				parts = append(parts, coreTypes.ContentPart{
					Type:        part.Type,
					ExtraFields: part.Raw,
				})
			}
		}

		msg.Content.ContentParts = parts
		messages = append(messages, msg)
	}

	return messages
}

func convertRawInputItemsToCoreMessages(items []interface{}) []coreTypes.Message {
	messages := make([]coreTypes.Message, 0, len(items))
	for _, rawItem := range items {
		itemMap, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}

		msg := coreTypes.Message{}
		if role, ok := itemMap["role"].(string); ok {
			msg.Role = role
		}

		content := itemMap["content"]
		switch typed := content.(type) {
		case string:
			text := typed
			msg.Content.StringValue = &text
		case []interface{}:
			parts := make([]coreTypes.ContentPart, 0, len(typed))
			for _, rawPart := range typed {
				partMap, ok := rawPart.(map[string]interface{})
				if !ok {
					continue
				}
				partType, _ := partMap["type"].(string)
				switch partType {
				case "input_text":
					text, _ := partMap["text"].(string)
					parts = append(parts, coreTypes.ContentPart{
						Type:        "text",
						Text:        &text,
						ExtraFields: partMap,
					})
				case "input_image":
					imageURL, ok := partMap["image_url"].(map[string]interface{})
					if !ok {
						continue
					}
					url, _ := imageURL["url"].(string)
					var detail *string
					if detailValue, ok := imageURL["detail"].(string); ok {
						detail = &detailValue
					}
					parts = append(parts, coreTypes.ContentPart{
						Type: "image_url",
						ImageURL: &coreTypes.ImageURL{
							URL:    url,
							Detail: detail,
						},
						ExtraFields: partMap,
					})
				default:
					parts = append(parts, coreTypes.ContentPart{
						Type:        partType,
						ExtraFields: partMap,
					})
				}
			}

			if len(parts) == 1 && parts[0].Type == "text" && parts[0].ImageURL == nil && parts[0].Text != nil {
				msg.Content.StringValue = parts[0].Text
			} else {
				msg.Content.ContentParts = parts
			}
		}

		messages = append(messages, msg)
	}

	return messages
}
