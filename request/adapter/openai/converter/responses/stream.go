package responses

import (
	"strings"

	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	coreTypes "github.com/MeowSalty/portal/types"
)

// StreamEventMeta 表示 Responses 流式事件的上下文信息
// 用于补齐官方事件所需的序列号和索引字段。
type StreamEventMeta struct {
	SequenceNumber int
	OutputIndex    int
	ContentIndex   int
	ItemID         string
	ResponseID     string
}

// ConvertStreamCreated 将核心响应转换为 response.created 事件
func ConvertStreamCreated(resp *coreTypes.Response, meta *StreamEventMeta) *openaiResponses.ResponsesStreamEvent {
	if resp == nil {
		return nil
	}

	sequenceNumber := 0
	if meta != nil {
		sequenceNumber = meta.SequenceNumber
	}

	return &openaiResponses.ResponsesStreamEvent{
		Type:           "response.created",
		SequenceNumber: sequenceNumber,
		Response: &openaiResponses.Response{
			ID:        resp.ID,
			Object:    resp.Object,
			CreatedAt: resp.Created,
			Model:     resp.Model,
		},
	}
}

// ConvertStreamInProgress 将核心响应转换为 response.in_progress 事件
func ConvertStreamInProgress(resp *coreTypes.Response, meta *StreamEventMeta) *openaiResponses.ResponsesStreamEvent {
	if resp == nil {
		return nil
	}

	sequenceNumber := 0
	if meta != nil {
		sequenceNumber = meta.SequenceNumber
	}

	return &openaiResponses.ResponsesStreamEvent{
		Type:           "response.in_progress",
		SequenceNumber: sequenceNumber,
		Response: &openaiResponses.Response{
			ID:        resp.ID,
			Object:    resp.Object,
			CreatedAt: resp.Created,
			Model:     resp.Model,
		},
	}
}

// ConvertStreamResponse 将核心流式响应转换为 Responses SSE 事件
func ConvertStreamResponse(resp *coreTypes.Response, meta *StreamEventMeta) *openaiResponses.ResponsesStreamEvent {
	if resp == nil {
		return nil
	}

	sequenceNumber := 0
	outputIndex := 0
	contentIndex := 0
	itemID := resp.ID
	responseID := resp.ID
	if meta != nil {
		sequenceNumber = meta.SequenceNumber
		outputIndex = meta.OutputIndex
		contentIndex = meta.ContentIndex
		if meta.ItemID != "" {
			itemID = meta.ItemID
		}
		if meta.ResponseID != "" {
			responseID = meta.ResponseID
		}
	}

	if resp.Usage != nil && len(resp.Choices) == 0 {
		return &openaiResponses.ResponsesStreamEvent{
			Type:           "response.completed",
			SequenceNumber: sequenceNumber,
			Response: &openaiResponses.Response{
				ID:        resp.ID,
				Object:    resp.Object,
				CreatedAt: resp.Created,
				Model:     resp.Model,
				Usage: &openaiResponses.Usage{
					InputTokens:  resp.Usage.PromptTokens,
					OutputTokens: resp.Usage.CompletionTokens,
					TotalTokens:  resp.Usage.TotalTokens,
				},
			},
		}
	}

	for _, choice := range resp.Choices {
		if choice.Error != nil {
			return &openaiResponses.ResponsesStreamEvent{
				Type:           "error",
				SequenceNumber: sequenceNumber,
				ResponseID:     responseID,
				Error: &openaiResponses.ResponsesError{
					Message: choice.Error.Message,
					Type:    extractErrorType(choice.Error.Metadata),
				},
			}
		}

		if resp.Usage != nil {
			return &openaiResponses.ResponsesStreamEvent{
				Type:           "response.completed",
				SequenceNumber: sequenceNumber,
				Response:       ConvertResponse(resp),
			}
		}

		if choice.Delta != nil {
			if choice.Delta.Content == nil || *choice.Delta.Content == "" {
				return nil
			}
			return &openaiResponses.ResponsesStreamEvent{
				Type:           "response.output_text.delta",
				ItemID:         itemID,
				OutputIndex:    outputIndex,
				ContentIndex:   contentIndex,
				SequenceNumber: sequenceNumber,
				Delta:          *choice.Delta.Content,
			}
		}

		if choice.Message != nil {
			msgItemID := itemID
			if choice.Message.ID != nil && *choice.Message.ID != "" {
				msgItemID = *choice.Message.ID
			}
			if len(choice.Message.ContentParts) == 0 && choice.Message.Content == nil {
				return &openaiResponses.ResponsesStreamEvent{
					Type:           "response.output_item.added",
					OutputIndex:    outputIndex,
					SequenceNumber: sequenceNumber,
					Item: &openaiResponses.OutputItem{
						ID:   msgItemID,
						Type: "message",
						Role: choice.Message.Role,
					},
				}
			}

			return &openaiResponses.ResponsesStreamEvent{
				Type:           "response.output_text.done",
				ItemID:         msgItemID,
				OutputIndex:    outputIndex,
				ContentIndex:   contentIndex,
				SequenceNumber: sequenceNumber,
				Text:           extractMessageText(choice.Message),
			}
		}
	}

	return nil
}

func extractMessageText(message *coreTypes.ResponseMessage) string {
	if message == nil {
		return ""
	}
	if message.Content != nil {
		return *message.Content
	}
	if len(message.ContentParts) == 0 {
		return ""
	}
	parts := make([]string, 0, len(message.ContentParts))
	for _, part := range message.ContentParts {
		if part.Type == "output_text" && part.Text != "" {
			parts = append(parts, part.Text)
		}
	}
	return strings.Join(parts, "")
}

func extractErrorType(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	if value, ok := metadata["type"]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// ConvertStreamEvent 将 Responses SSE 事件转换为核心响应
func ConvertStreamEvent(event *openaiResponses.ResponsesStreamEvent) *coreTypes.Response {
	if event == nil {
		return nil
	}

	switch event.Type {
	case "response.created":
		if event.Response == nil {
			return nil
		}
		return &coreTypes.Response{
			ID:      event.Response.ID,
			Object:  event.Response.Object,
			Model:   event.Response.Model,
			Created: event.Response.CreatedAt,
		}
	case "response.in_progress":
		if event.Response == nil {
			return nil
		}
		coreResp := &coreTypes.Response{
			ID:      event.Response.ID,
			Object:  event.Response.Object,
			Model:   event.Response.Model,
			Created: event.Response.CreatedAt,
		}
		// 将 status 存入 ExtraFields（通过空 Message 传递）
		message := &coreTypes.ResponseMessage{
			Role: "assistant",
			ExtraFields: map[string]interface{}{
				"status": "in_progress",
			},
			ExtraFieldsFormat: "openai",
		}
		coreResp.Choices = []coreTypes.Choice{
			{
				Message: message,
			},
		}
		return coreResp
	case "response.output_item.added":
		message := buildStreamMessageFromItem(event.Item)
		attachStreamMetaToMessage(message, event)
		if message == nil {
			return nil
		}
		return &coreTypes.Response{
			ID: event.ResponseID,
			Choices: []coreTypes.Choice{
				{
					Message: message,
				},
			},
		}
	case "response.output_item.done":
		message := buildStreamMessageFromItem(event.Item)
		if message == nil {
			return nil
		}
		if message.ExtraFields == nil {
			message.ExtraFields = map[string]interface{}{}
		}
		message.ExtraFields["status"] = "completed"
		attachStreamMetaToMessage(message, event)
		return &coreTypes.Response{
			ID: event.ResponseID,
			Choices: []coreTypes.Choice{
				{
					Message: message,
				},
			},
		}
	case "response.content_part.added":
		message := buildStreamMessageFromPart(event.Part)
		if message == nil {
			return nil
		}
		attachStreamMetaToMessage(message, event)
		return &coreTypes.Response{
			ID: event.ResponseID,
			Choices: []coreTypes.Choice{
				{
					Message: message,
				},
			},
		}
	case "response.content_part.done":
		message := buildStreamMessageFromPart(event.Part)
		if message == nil {
			return nil
		}
		if message.ExtraFields == nil {
			message.ExtraFields = map[string]interface{}{}
		}
		message.ExtraFields["status"] = "completed"
		attachStreamMetaToMessage(message, event)
		return &coreTypes.Response{
			ID: event.ResponseID,
			Choices: []coreTypes.Choice{
				{
					Message: message,
				},
			},
		}
	case "response.output_text.delta":
		if event.Delta == "" {
			return nil
		}
		deltaText := event.Delta
		delta := &coreTypes.Delta{
			Content: &deltaText,
		}
		attachStreamMetaToDelta(delta, event)
		return &coreTypes.Response{
			ID: event.ResponseID,
			Choices: []coreTypes.Choice{
				{
					Delta: delta,
				},
			},
		}
	case "response.output_text.done":
		if event.Text == "" {
			return nil
		}
		message := &coreTypes.ResponseMessage{
			Role:    "assistant",
			Content: &event.Text,
			ContentParts: []coreTypes.ResponseContentPart{
				{
					Type: "output_text",
					Text: event.Text,
				},
			},
		}
		attachStreamMetaToMessage(message, event)
		return &coreTypes.Response{
			ID: event.ResponseID,
			Choices: []coreTypes.Choice{
				{
					Message: message,
				},
			},
		}
	case "response.output_text.annotation.added":
		if event.Annotation == nil {
			return nil
		}
		message := &coreTypes.ResponseMessage{
			Role: "assistant",
			ContentParts: []coreTypes.ResponseContentPart{
				{
					Type:        "output_text",
					Annotations: []interface{}{event.Annotation},
				},
			},
		}
		attachStreamMetaToMessage(message, event)
		return &coreTypes.Response{
			ID: event.ResponseID,
			Choices: []coreTypes.Choice{
				{
					Message: message,
				},
			},
		}
	case "response.completed":
		if event.Response == nil {
			return nil
		}
		if len(event.Response.Output) == 0 {
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
		}
		return ConvertCoreResponse(event.Response)
	case "error":
		message := ""
		errType := ""
		if event.Error != nil {
			message = event.Error.Message
			errType = event.Error.Type
		} else {
			message = event.Message
		}
		if message == "" {
			return nil
		}
		metadata := map[string]interface{}{}
		if errType != "" {
			metadata["type"] = errType
		}
		if event.Code != "" {
			metadata["code"] = event.Code
		}
		if event.Param != "" {
			metadata["param"] = event.Param
		}
		return &coreTypes.Response{
			Choices: []coreTypes.Choice{
				{
					Error: &coreTypes.ErrorResponse{
						Code:     500,
						Message:  message,
						Metadata: metadata,
					},
				},
			},
		}
	default:
		return nil
	}
}

func buildStreamMessageFromPart(part *openaiResponses.OutputPart) *coreTypes.ResponseMessage {
	if part == nil {
		return nil
	}

	message := &coreTypes.ResponseMessage{
		Role: "assistant",
	}

	message.ContentParts = []coreTypes.ResponseContentPart{
		{
			Type:        part.Type,
			Text:        part.Text,
			Annotations: part.Annotations,
		},
	}

	if part.Type == "output_text" && part.Text != "" {
		message.Content = &part.Text
	}

	return message
}

func buildStreamMessageFromItem(item *openaiResponses.OutputItem) *coreTypes.ResponseMessage {
	if item == nil {
		return nil
	}

	message := &coreTypes.ResponseMessage{
		Role: item.Role,
	}
	if item.ID != "" {
		id := item.ID
		message.ID = &id
	}
	if message.Role == "" {
		message.Role = "assistant"
	}

	if len(item.Content) == 0 {
		return message
	}

	contentParts := make([]string, 0, len(item.Content))
	message.ContentParts = make([]coreTypes.ResponseContentPart, 0, len(item.Content))
	for _, part := range item.Content {
		if part.Type == "output_text" {
			contentParts = append(contentParts, part.Text)
		}
		message.ContentParts = append(message.ContentParts, coreTypes.ResponseContentPart{
			Type:        part.Type,
			Text:        part.Text,
			Annotations: part.Annotations,
		})
	}
	if len(contentParts) > 0 {
		joined := strings.Join(contentParts, "")
		message.Content = &joined
	}
	return message
}

func attachStreamMetaToMessage(message *coreTypes.ResponseMessage, event *openaiResponses.ResponsesStreamEvent) {
	if message == nil || event == nil {
		return
	}
	if message.ExtraFields == nil {
		message.ExtraFields = map[string]interface{}{}
	}
	if event.ResponseID != "" {
		message.ExtraFields["response_id"] = event.ResponseID
	}
	if event.ItemID != "" {
		if message.ID == nil {
			id := event.ItemID
			message.ID = &id
		}
		message.ExtraFields["item_id"] = event.ItemID
	}
	message.ExtraFields["output_index"] = event.OutputIndex
	message.ExtraFields["content_index"] = event.ContentIndex
	message.ExtraFields["annotation_index"] = event.AnnotationIndex
	if len(event.Logprobs) > 0 {
		message.ExtraFields["logprobs"] = event.Logprobs
	}
	message.ExtraFieldsFormat = "openai"
}

func attachStreamMetaToDelta(delta *coreTypes.Delta, event *openaiResponses.ResponsesStreamEvent) {
	if delta == nil || event == nil {
		return
	}
	if delta.ExtraFields == nil {
		delta.ExtraFields = map[string]interface{}{}
	}
	if event.ResponseID != "" {
		delta.ExtraFields["response_id"] = event.ResponseID
	}
	if event.ItemID != "" {
		delta.ExtraFields["item_id"] = event.ItemID
	}
	delta.ExtraFields["output_index"] = event.OutputIndex
	delta.ExtraFields["content_index"] = event.ContentIndex
	if len(event.Logprobs) > 0 {
		delta.ExtraFields["logprobs"] = event.Logprobs
	}
	delta.ExtraFieldsFormat = "openai"
}
