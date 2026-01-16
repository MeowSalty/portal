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

	// 聚合 output 中的文本，并记录结构化内容
	var contentParts []string
	var contentItems []coreTypes.ResponseContentPart
	var messageID *string
	var messageRole string
	for _, item := range resp.Output {
		if item.Type != "message" {
			continue
		}
		if item.ID != "" {
			id := item.ID
			messageID = &id
		}
		if item.Role != "" {
			messageRole = item.Role
		}
		for _, part := range item.Content {
			if part.Type == "output_text" {
				contentParts = append(contentParts, part.Text)
			}
			contentItems = append(contentItems, coreTypes.ResponseContentPart{
				Type:        part.Type,
				Text:        part.Text,
				Annotations: part.Annotations,
			})
		}
	}

	var content *string
	if len(contentParts) > 0 {
		joined := strings.Join(contentParts, "")
		content = &joined
	}

	if messageRole == "" {
		messageRole = "assistant"
	}

	coreResp.Choices = []coreTypes.Choice{
		{
			Message: &coreTypes.ResponseMessage{
				ID:           messageID,
				Role:         messageRole,
				Content:      content,
				ContentParts: contentItems,
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
		var contentParts []coreTypes.ResponseContentPart
		if choice.Message != nil {
			item.Role = choice.Message.Role
			content = choice.Message.Content
			contentParts = choice.Message.ContentParts
			if choice.Message.ID != nil {
				item.ID = *choice.Message.ID
			}
		} else if choice.Delta != nil {
			if choice.Delta.Role != nil {
				item.Role = *choice.Delta.Role
			}
			content = choice.Delta.Content
		}

		if item.Role == "" {
			item.Role = "assistant"
		}

		if len(contentParts) > 0 {
			item.Content = make([]openaiResponses.OutputPart, 0, len(contentParts))
			for _, part := range contentParts {
				item.Content = append(item.Content, openaiResponses.OutputPart{
					Type:        part.Type,
					Text:        part.Text,
					Annotations: part.Annotations,
				})
			}
		} else if content != nil {
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
