package converter

import (
	"encoding/json"

	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertContentBlockStartEvent 转换 content_block_start 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.1）：
//   - content_index: index
//   - content.kind: 依据 content_block.type 映射 text/tool_use/other
//   - content.text: content_block.text（若为 text）
//   - content.tool: content_block.tool_use（若为 tool_use）
//   - content.raw: content_block 原文
//
// 索引补齐规则（plans/stream-index-repair-plan.md B.2）：
//   - sequence_number：若为 0 -> ctx.NextSequence()
//   - output_index：默认 0
//   - item_id：若有 tool_use.id -> 用 tool_use.id；否则 message_id:content_index 组合
//   - content_index：来自 event.Index；若缺失 -> ctx.EnsureContentIndex(item_id, -1)
func convertContentBlockStartEvent(event *anthropicTypes.ContentBlockStartEvent, ctx types.StreamIndexContext, log logger.Logger) (*types.StreamEventContract, error) {
	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockStart,
		Source:       types.StreamSourceAnthropic,
		OutputIndex:  -1, // 缺失索引使用 -1 哨兵
		ContentIndex: event.Index,
	}

	// 从上下文获取 message_id
	messageID := ctx.GetMessageID()

	// 转换 ContentBlock
	content := &types.StreamContentPayload{
		Raw: make(map[string]interface{}),
	}

	var toolCalls []types.StreamToolCall

	// 根据 content_block 类型设置 kind 和相关字段
	if event.ContentBlock.Text != nil {
		content.Kind = "text"
		content.Text = &event.ContentBlock.Text.Text

		// 保存原始结构到 raw
		blockJSON, err := json.Marshal(event.ContentBlock.Text)
		if err != nil {
			log.Warn("序列化 text block 失败", "error", err)
		} else {
			content.Raw["content_block"] = json.RawMessage(blockJSON)
		}

		// Citations 作为 annotations
		if len(event.ContentBlock.Text.Citations) > 0 {
			annotations, err := convertCitationsToAnnotations(event.ContentBlock.Text.Citations)
			if err != nil {
				return nil, err
			}
			content.Annotations = make([]interface{}, len(annotations))
			for i, annotation := range annotations {
				content.Annotations[i] = annotation
			}
		}
	} else if event.ContentBlock.ToolUse != nil {
		content.Kind = "tool_use"

		// 转换 ToolUse 为 StreamToolCall
		tool := &types.StreamToolCall{
			ID:   event.ContentBlock.ToolUse.ID,
			Name: event.ContentBlock.ToolUse.Name,
			Type: "tool_use",
			Raw:  make(map[string]interface{}),
		}

		// 序列化 Input 为 JSON 字符串
		if event.ContentBlock.ToolUse.Input != nil {
			inputJSON, err := SerializeToolInput(event.ContentBlock.ToolUse.Input, nil)
			if err == nil && inputJSON != "" {
				tool.Arguments = inputJSON
			}
		}

		content.Tool = tool
		toolCalls = append(toolCalls, *tool)

		// 保存原始结构到 raw
		blockJSON, err := json.Marshal(event.ContentBlock.ToolUse)
		if err != nil {
			log.Warn("序列化 tool_use block 失败", "error", err)
		} else {
			content.Raw["content_block"] = json.RawMessage(blockJSON)
		}
	} else if event.ContentBlock.Thinking != nil {
		content.Kind = "thinking"
		content.Text = &event.ContentBlock.Thinking.Thinking

		// 保存原始结构到 raw
		blockJSON, err := json.Marshal(event.ContentBlock.Thinking)
		if err != nil {
			log.Warn("序列化 thinking block 失败", "error", err)
		} else {
			content.Raw["content_block"] = json.RawMessage(blockJSON)
		}
	} else if event.ContentBlock.RedactedThinking != nil {
		content.Kind = "redacted_thinking"

		// 保存原始结构到 raw
		blockJSON, err := json.Marshal(event.ContentBlock.RedactedThinking)
		if err != nil {
			log.Warn("序列化 redacted_thinking block 失败", "error", err)
		} else {
			content.Raw["content_block"] = json.RawMessage(blockJSON)
		}
	} else if event.ContentBlock.ServerToolUse != nil {
		content.Kind = "server_tool_use"

		// 转换 ServerToolUse 为 StreamToolCall
		tool := &types.StreamToolCall{
			ID:   event.ContentBlock.ServerToolUse.ID,
			Name: event.ContentBlock.ServerToolUse.Name,
			Type: "server_tool_use",
			Raw:  make(map[string]interface{}),
		}

		// 序列化 Input 为 JSON 字符串
		if event.ContentBlock.ServerToolUse.Input != nil {
			inputJSON, err := SerializeToolInput(event.ContentBlock.ServerToolUse.Input, nil)
			if err == nil && inputJSON != "" {
				tool.Arguments = inputJSON
			}
		}

		content.Tool = tool
		toolCalls = append(toolCalls, *tool)

		// 保存原始结构到 raw
		blockJSON, err := json.Marshal(event.ContentBlock.ServerToolUse)
		if err != nil {
			log.Warn("序列化 server_tool_use block 失败", "error", err)
		} else {
			content.Raw["content_block"] = json.RawMessage(blockJSON)
		}
	} else if event.ContentBlock.WebSearchToolResult != nil {
		content.Kind = "web_search_tool_result"

		// 保存原始结构到 raw
		blockJSON, err := json.Marshal(event.ContentBlock.WebSearchToolResult)
		if err != nil {
			log.Warn("序列化 web_search_tool_result block 失败", "error", err)
		} else {
			content.Raw["content_block"] = json.RawMessage(blockJSON)
		}
	} else {
		content.Kind = "other"
	}

	contract.Content = content

	// 补齐索引字段，传入 toolCalls 以便使用 tool_use.id 作为 item_id
	fillMissingIndices(contract, messageID, 0, toolCalls, event.Index, ctx)

	// 保存 item_id 到上下文，供后续事件使用
	if contract.ItemID != "" {
		ctx.SetItemID(contract.ItemID)
	}

	log.Debug("转换 content_block_start 事件完成", "content_index", event.Index, "kind", content.Kind)
	return contract, nil
}

// convertContentBlockDeltaEvent 转换 content_block_delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.1）：
//   - content_index: index
//   - delta.delta_type: type
//   - delta.text: text_delta.text
//   - delta.partial_json: input_json_delta.partial_json
//   - delta.thinking: thinking_delta.thinking
//   - delta.signature: signature_delta.signature
//   - delta.citation: citations_delta.citation
//
// 索引补齐规则（plans/stream-index-repair-plan.md B.2）：
//   - sequence_number：若为 0 -> ctx.NextSequence()
//   - output_index：默认 0
//   - item_id：不适用（delta 事件不包含 item 信息）
//   - content_index：来自 event.Index；若缺失 -> ctx.EnsureContentIndex(item_id, -1)
func convertContentBlockDeltaEvent(event *anthropicTypes.ContentBlockDeltaEvent, ctx types.StreamIndexContext, log logger.Logger) (*types.StreamEventContract, error) {
	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockDelta,
		Source:       types.StreamSourceAnthropic,
		OutputIndex:  -1, // 缺失索引使用 -1 哨兵
		ContentIndex: event.Index,
	}

	// 从上下文获取 message_id
	messageID := ctx.GetMessageID()

	// 转换 Delta
	delta := &types.StreamDeltaPayload{
		Raw: make(map[string]interface{}),
	}

	if event.Delta.Text != nil {
		delta.DeltaType = string(event.Delta.Text.Type)
		delta.Text = &event.Delta.Text.Text
	} else if event.Delta.InputJSON != nil {
		delta.DeltaType = string(event.Delta.InputJSON.Type)
		delta.PartialJSON = &event.Delta.InputJSON.PartialJSON
	} else if event.Delta.Thinking != nil {
		delta.DeltaType = string(event.Delta.Thinking.Type)
		delta.Thinking = &event.Delta.Thinking.Thinking
	} else if event.Delta.Signature != nil {
		delta.DeltaType = string(event.Delta.Signature.Type)
		delta.Signature = &event.Delta.Signature.Signature
	} else if event.Delta.Citations != nil {
		delta.DeltaType = string(event.Delta.Citations.Type)
		delta.Citation = event.Delta.Citations.Citation
	} else {
		delta.DeltaType = "other"
	}

	contract.Delta = delta

	// 补齐索引字段
	fillMissingIndices(contract, messageID, 0, nil, event.Index, ctx)

	log.Debug("转换 content_block_delta 事件完成", "content_index", event.Index, "delta_type", delta.DeltaType)
	return contract, nil
}

// convertContentBlockStopEvent 转换 content_block_stop 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.1）：
//   - content_index: index
//
// 索引补齐规则（plans/stream-index-repair-plan.md B.2）：
//   - sequence_number：若为 0 -> ctx.NextSequence()
//   - output_index：默认 0
//   - item_id：不适用
//   - content_index：来自 event.Index；若缺失 -> ctx.EnsureContentIndex(item_id, -1)
func convertContentBlockStopEvent(event *anthropicTypes.ContentBlockStopEvent, ctx types.StreamIndexContext, log logger.Logger) (*types.StreamEventContract, error) {
	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockStop,
		Source:       types.StreamSourceAnthropic,
		OutputIndex:  -1, // 缺失索引使用 -1 哨兵
		ContentIndex: event.Index,
	}

	// 从上下文获取 message_id
	messageID := ctx.GetMessageID()

	// 补齐索引字段
	fillMissingIndices(contract, messageID, 0, nil, event.Index, ctx)

	log.Debug("转换 content_block_stop 事件完成", "content_index", event.Index)
	return contract, nil
}
