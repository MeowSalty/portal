package converter

import (
	"encoding/json"

	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ConvertCoreResponse 将 Anthropic 响应转换为核心响应
func ConvertCoreResponse(anthropicResp *anthropicTypes.AnthropicResponse) *coreTypes.Response {
	response := &coreTypes.Response{
		ID:      anthropicResp.ID,
		Model:   anthropicResp.Model,
		Choices: make([]coreTypes.Choice, 1),
	}

	// 转换内容
	content := convertResponseContent(anthropicResp.Content)
	var contentStr *string
	if content.StringValue != nil {
		contentStr = content.StringValue
	}

	choice := coreTypes.Choice{
		Message: &coreTypes.ResponseMessage{
			Role:    anthropicResp.Role,
			Content: contentStr,
		},
	}

	// 设置停止原因
	if anthropicResp.StopReason != nil {
		choice.FinishReason = anthropicResp.StopReason
	}

	response.Choices[0] = choice

	// 转换使用统计
	if anthropicResp.Usage.InputTokens > 0 || anthropicResp.Usage.OutputTokens > 0 {
		response.Usage = &coreTypes.ResponseUsage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		}
	}

	return response
}

// convertResponseContent 转换响应内容
func convertResponseContent(contents []anthropicTypes.ResponseContent) coreTypes.MessageContent {
	// 如果只有一个文本块，返回字符串
	if len(contents) == 1 && contents[0].Type == "text" && contents[0].Text != nil {
		return coreTypes.MessageContent{
			StringValue: contents[0].Text,
		}
	}

	// 否则返回内容部分数组
	parts := make([]coreTypes.ContentPart, 0, len(contents))

	for _, content := range contents {
		part := coreTypes.ContentPart{
			Type: content.Type,
		}

		if content.Type == "text" && content.Text != nil {
			part.Text = content.Text
		} else if content.Type == "tool_use" {
			// 转换工具使用
			part.Type = "tool_use"
			// 工具使用的详细信息需要序列化到适当的格式
			// 这里简化处理，实际应用中可能需要更复杂的转换
		}

		parts = append(parts, part)
	}

	return coreTypes.MessageContent{
		ContentParts: parts,
	}
}

// ConvertStreamEvent 转换流式事件为核心响应
func ConvertStreamEvent(event *anthropicTypes.StreamEvent) *coreTypes.Response {
	response := &coreTypes.Response{
		Choices: make([]coreTypes.Choice, 1),
	}

	choice := coreTypes.Choice{}

	switch event.Type {
	case "message_start":
		// 消息开始事件
		if event.Message != nil {
			response.ID = event.Message.ID
			response.Model = event.Message.Model
			choice.Delta = &coreTypes.Delta{
				Role: &event.Message.Role,
			}
		}

	case "content_block_start":
		// 内容块开始
		if event.ContentBlock != nil {
			if event.ContentBlock.Type == "text" {
				emptyText := ""
				choice.Delta = &coreTypes.Delta{
					Content: &emptyText,
				}
			}
		}

	case "content_block_delta":
		// 内容块增量更新
		if event.Delta != nil {
			if event.Delta.Type == "text_delta" && event.Delta.Text != nil {
				choice.Delta = &coreTypes.Delta{
					Content: event.Delta.Text,
				}
			}
		}

	case "content_block_stop":
		// 内容块结束
		// 不需要特殊处理

	case "message_delta":
		// 消息增量更新
		if event.Delta != nil {
			if event.Delta.StopReason != nil {
				choice.FinishReason = event.Delta.StopReason
			}
		}

	case "message_stop":
		// 消息结束
		choice.FinishReason = stringPtr("stop")

	case "error":
		// 错误事件
		// 需要从原始数据中解析错误
	}

	// 设置使用统计
	if event.Usage != nil {
		response.Usage = &coreTypes.ResponseUsage{
			PromptTokens:     event.Usage.InputTokens,
			CompletionTokens: event.Usage.OutputTokens,
			TotalTokens:      event.Usage.InputTokens + event.Usage.OutputTokens,
		}
	}

	response.Choices[0] = choice
	return response
}

// ParseStreamLine 解析流式响应行
func ParseStreamLine(line []byte) (*anthropicTypes.StreamEvent, error) {
	var event anthropicTypes.StreamEvent
	if err := json.Unmarshal(line, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}
