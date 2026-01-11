package responses

import (
	"strings"

	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ConvertCoreResponse 将 Responses 响应转换为核心响应
func ConvertCoreResponse(resp *openaiResponses.Response) *coreTypes.Response {
	if resp == nil {
		return nil
	}

	coreResp := &coreTypes.Response{
		ID:      resp.ID,
		Object:  resp.Object,
		Model:   resp.Model,
		Created: resp.CreatedAt,
	}

	if resp.Usage != nil {
		coreResp.Usage = &coreTypes.ResponseUsage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	// 聚合 output 中的文本
	var contentParts []string
	for _, item := range resp.Output {
		if item.Type != "message" {
			continue
		}
		for _, part := range item.Content {
			if part.Type == "output_text" {
				contentParts = append(contentParts, part.Text)
			}
		}
	}

	var content *string
	if len(contentParts) > 0 {
		joined := strings.Join(contentParts, "")
		content = &joined
	}

	coreResp.Choices = []coreTypes.Choice{
		{
			Message: &coreTypes.ResponseMessage{
				Role:    "assistant",
				Content: content,
			},
		},
	}

	return coreResp
}

// ConvertResponse 将核心响应转换为 Responses 响应
func ConvertResponse(resp *coreTypes.Response) *openaiResponses.Response {
	if resp == nil {
		return nil
	}

	openaiResp := &openaiResponses.Response{
		ID:        resp.ID,
		Object:    resp.Object,
		CreatedAt: resp.Created,
		Model:     resp.Model,
	}

	if resp.Usage != nil {
		openaiResp.Usage = &openaiResponses.Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		}
	}

	if len(resp.Choices) == 0 {
		return openaiResp
	}

	openaiResp.Output = make([]openaiResponses.OutputItem, 0, len(resp.Choices))
	for _, choice := range resp.Choices {
		item := openaiResponses.OutputItem{
			Type: "message",
		}

		var content *string
		if choice.Message != nil {
			item.Role = choice.Message.Role
			content = choice.Message.Content
		} else if choice.Delta != nil {
			if choice.Delta.Role != nil {
				item.Role = *choice.Delta.Role
			}
			content = choice.Delta.Content
		}

		if item.Role == "" {
			item.Role = "assistant"
		}

		if content != nil {
			item.Content = []openaiResponses.OutputPart{
				{
					Type: "output_text",
					Text: *content,
				},
			}
		}

		openaiResp.Output = append(openaiResp.Output, item)
	}

	return openaiResp
}

// ConvertStreamEvent 将 Responses SSE 事件转换为核心响应
func ConvertStreamEvent(event *openaiResponses.ResponsesStreamEvent) *coreTypes.Response {
	if event == nil {
		return nil
	}

	switch event.Type {
	case "response.output_text.delta":
		if event.Delta == "" {
			return nil
		}
		deltaText := event.Delta
		return &coreTypes.Response{
			Choices: []coreTypes.Choice{
				{
					Delta: &coreTypes.Delta{
						Content: &deltaText,
					},
				},
			},
		}
	case "response.completed":
		if event.Response == nil {
			return nil
		}
		coreResp := &coreTypes.Response{
			ID:      event.Response.ID,
			Object:  event.Response.Object,
			Model:   event.Response.Model,
			Created: event.Response.CreatedAt,
		}
		if event.Response.Usage != nil {
			coreResp.Usage = &coreTypes.ResponseUsage{
				PromptTokens:     event.Response.Usage.InputTokens,
				CompletionTokens: event.Response.Usage.OutputTokens,
				TotalTokens:      event.Response.Usage.TotalTokens,
			}
		}
		return coreResp
	case "error":
		if event.Error == nil {
			return nil
		}
		return &coreTypes.Response{
			Choices: []coreTypes.Choice{
				{
					Error: &coreTypes.ErrorResponse{
						Code:    500,
						Message: event.Error.Message,
						Metadata: map[string]interface{}{
							"type": event.Error.Type,
						},
					},
				},
			},
		}
	default:
		return nil
	}
}
