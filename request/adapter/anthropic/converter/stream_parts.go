package converter

import (
	"encoding/json"

	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertResponseContentBlocksToStreamParts 转换响应内容块列表为流式内容分片和工具调用。
//
// 用于 message_start 事件中的 message.parts 和 message.tool_calls 转换。
func convertResponseContentBlocksToStreamParts(blocks []anthropicTypes.ResponseContentBlock, log logger.Logger) ([]types.StreamContentPart, []types.StreamToolCall, error) {
	var parts []types.StreamContentPart
	var toolCalls []types.StreamToolCall

	for _, block := range blocks {
		if block.Text != nil {
			part := types.StreamContentPart{
				Type: "text",
				Text: block.Text.Text,
				Raw:  make(map[string]interface{}),
			}

			// Citations 作为 annotations
			if len(block.Text.Citations) > 0 {
				part.Annotations = make([]interface{}, len(block.Text.Citations))
				for i, citation := range block.Text.Citations {
					part.Annotations[i] = citation
				}
			}

			parts = append(parts, part)
		} else if block.Thinking != nil {
			part := types.StreamContentPart{
				Type: "thinking",
				Text: block.Thinking.Thinking,
				Raw:  make(map[string]interface{}),
			}
			part.Raw["signature"] = block.Thinking.Signature

			parts = append(parts, part)
		} else if block.RedactedThinking != nil {
			part := types.StreamContentPart{
				Type: "redacted_thinking",
				Raw:  make(map[string]interface{}),
			}
			part.Raw["data"] = block.RedactedThinking.Data

			parts = append(parts, part)
		} else if block.ToolUse != nil {
			// ToolUse 转换为 ToolCall
			toolCall := types.StreamToolCall{
				ID:   block.ToolUse.ID,
				Name: block.ToolUse.Name,
				Type: "tool_use",
				Raw:  make(map[string]interface{}),
			}

			// 序列化 Input 为 JSON 字符串
			if block.ToolUse.Input != nil {
				inputJSON, err := json.Marshal(block.ToolUse.Input)
				if err != nil {
					log.Warn("序列化 tool input 失败", "error", err)
				} else {
					toolCall.Arguments = string(inputJSON)
				}
			}

			toolCalls = append(toolCalls, toolCall)
		} else if block.ServerToolUse != nil {
			// ServerToolUse 转换为 ToolCall
			toolCall := types.StreamToolCall{
				ID:   block.ServerToolUse.ID,
				Name: block.ServerToolUse.Name,
				Type: "server_tool_use",
				Raw:  make(map[string]interface{}),
			}

			// 序列化 Input 为 JSON 字符串
			if block.ServerToolUse.Input != nil {
				inputJSON, err := json.Marshal(block.ServerToolUse.Input)
				if err != nil {
					log.Warn("序列化 server tool input 失败", "error", err)
				} else {
					toolCall.Arguments = string(inputJSON)
				}
			}

			toolCalls = append(toolCalls, toolCall)
		} else if block.WebSearchToolResult != nil {
			part := types.StreamContentPart{
				Type: "web_search_tool_result",
				Raw:  make(map[string]interface{}),
			}

			// 保存完整结构到 raw
			resultJSON, err := json.Marshal(block.WebSearchToolResult)
			if err != nil {
				log.Warn("序列化 web_search_tool_result 失败", "error", err)
			} else {
				var rawData map[string]interface{}
				if err := json.Unmarshal(resultJSON, &rawData); err == nil {
					part.Raw = rawData
				}
			}

			parts = append(parts, part)
		}
	}

	return parts, toolCalls, nil
}
