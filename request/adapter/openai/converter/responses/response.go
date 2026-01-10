package responses

import (
	"strings"

	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ConvertResponsesCoreResponse 将 Responses 响应转换为核心响应
func ConvertResponsesCoreResponse(resp *openaiResponses.Response) *coreTypes.Response {
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

// ConvertResponsesStreamEvent 将 Responses SSE 事件转换为核心响应
func ConvertResponsesStreamEvent(event *openaiResponses.ResponsesStreamEvent) *coreTypes.Response {
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
