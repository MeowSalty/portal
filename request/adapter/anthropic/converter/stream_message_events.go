package converter

import (
	"encoding/json"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertContractToMessageStart 转换 message_start 事件为 Anthropic 格式。
//
// 反向映射规则：
//   - message.id: 优先 MessageID，fallback ResponseID
//   - message.role: Message.Role
//   - message.content: 从 Message.Parts + ToolCalls 构造，优先使用 extensions.anthropic.message
func convertContractToMessageStart(event *types.StreamEventContract, log logger.Logger) (*anthropicTypes.StreamEvent, error) {
	// 优先尝试从 extensions 中恢复原始 message
	var message anthropicTypes.Response
	useRawMessage := false

	if event.Extensions != nil {
		if anthropicExt, ok := event.Extensions["anthropic"].(map[string]interface{}); ok {
			if rawMsg, ok := anthropicExt["message"]; ok {
				msgJSON, err := json.Marshal(rawMsg)
				if err == nil {
					if err := json.Unmarshal(msgJSON, &message); err == nil {
						log.Debug("使用 extensions.anthropic.message 原文")
						useRawMessage = true
					} else {
						log.Warn("反序列化原始 message 失败", "error", err)
					}
				}
			}
		}
	}

	// 若没有原始 message，则构造
	if !useRawMessage {
		// 获取 message ID
		messageID := event.MessageID
		if messageID == "" {
			messageID = event.ResponseID
		}

		// 获取 role
		role := "assistant"
		if event.Message != nil && event.Message.Role != "" {
			role = event.Message.Role
		}

		message = anthropicTypes.Response{
			ID:   messageID,
			Type: anthropicTypes.ResponseTypeMessage,
			Role: anthropicTypes.Role(role),
		}

		// 构造 content blocks
		var blocks []anthropicTypes.ResponseContentBlock
		var err error

		if event.Message != nil {
			blocks, err = convertStreamMessageToContentBlocks(event.Message, log)
			if err != nil {
				log.Error("转换 message payload 失败", "error", err)
				return nil, errors.Wrap(errors.ErrCodeInternal, "转换 message payload 失败", err)
			}
		}

		message.Content = blocks
	}

	// 构造事件
	startEvent := &anthropicTypes.MessageStartEvent{
		Type:    anthropicTypes.StreamEventMessageStart,
		Message: message,
	}

	log.Debug("转换 message_start 事件完成", "message_id", message.ID)
	return &anthropicTypes.StreamEvent{MessageStart: startEvent}, nil
}

// convertContractToMessageDelta 转换 message_delta 事件为 Anthropic 格式。
//
// 反向映射规则：
//   - delta.stop_reason/stop_sequence: 从 Delta.Raw 中提取
//   - usage: 从 Usage 构造，包含 raw 中的特有字段
func convertContractToMessageDelta(event *types.StreamEventContract, log logger.Logger) (*anthropicTypes.StreamEvent, error) {
	deltaEvent := &anthropicTypes.MessageDeltaEvent{
		Type:  anthropicTypes.StreamEventMessageDelta,
		Delta: anthropicTypes.MessageDelta{},
		Usage: nil,
	}

	// 转换 delta
	if event.Delta != nil && len(event.Delta.Raw) > 0 {
		// 提取 stop_reason
		if stopReasonRaw, ok := event.Delta.Raw["stop_reason"].(string); ok && stopReasonRaw != "" {
			stopReason := anthropicTypes.StopReason(stopReasonRaw)
			deltaEvent.Delta.StopReason = &stopReason
		}

		// 提取 stop_sequence
		if stopSeqRaw, ok := event.Delta.Raw["stop_sequence"].(string); ok && stopSeqRaw != "" {
			deltaEvent.Delta.StopSequence = &stopSeqRaw
		}
	}

	// 转换 usage
	if event.Usage != nil {
		usage := &anthropicTypes.MessageDeltaUsage{
			InputTokens:  event.Usage.InputTokens,
			OutputTokens: event.Usage.OutputTokens,
		}

		// 提取 raw 中的特有字段
		if event.Usage.Raw != nil {
			if cacheCreation, ok := event.Usage.Raw["cache_creation_input_tokens"].(float64); ok {
				v := int(cacheCreation)
				usage.CacheCreationInputTokens = &v
			}
			if cacheRead, ok := event.Usage.Raw["cache_read_input_tokens"].(float64); ok {
				v := int(cacheRead)
				usage.CacheReadInputTokens = &v
			}
			// server_tool_use 可能是对象或数字
			if serverToolRaw, ok := event.Usage.Raw["server_tool_use"]; ok {
				switch v := serverToolRaw.(type) {
				case map[string]interface{}:
					serverToolUsage := &anthropicTypes.ServerToolUsage{}
					if webSearch, ok := v["web_search_requests"].(float64); ok {
						reqs := int(webSearch)
						serverToolUsage.WebSearchRequests = &reqs
					}
					usage.ServerToolUse = serverToolUsage
				case float64:
					// 旧格式：直接是数字
					// 这种情况较少见，暂不做处理
				}
			}
		}

		if usage.InputTokens != nil || usage.OutputTokens != nil ||
			usage.CacheCreationInputTokens != nil || usage.CacheReadInputTokens != nil {
			deltaEvent.Usage = usage
		}
	}

	log.Debug("转换 message_delta 事件完成")
	return &anthropicTypes.StreamEvent{MessageDelta: deltaEvent}, nil
}

// convertContractToMessageStop 转换 message_stop 事件为 Anthropic 格式。
func convertContractToMessageStop(log logger.Logger) (*anthropicTypes.StreamEvent, error) {
	stopEvent := &anthropicTypes.MessageStopEvent{
		Type: anthropicTypes.StreamEventMessageStop,
	}

	log.Debug("转换 message_stop 事件完成")
	return &anthropicTypes.StreamEvent{MessageStop: stopEvent}, nil
}

// convertMessageStartEvent 转换 message_start 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.1）：
//   - response_id: message.id
//   - message.role: message.role
//   - message.parts: message.content（按 text/tool_use 分类）
//   - extensions.anthropic.message: Message 原始结构
//
// 索引补齐规则（plans/stream-index-repair-plan.md B.2）：
//   - sequence_number：若为 0 -> ctx.NextSequence()
//   - output_index：默认 0
//   - item_id：若有 tool_use.id -> 用 tool_use.id；否则 message_id
//   - content_index：不适用
func convertMessageStartEvent(event *anthropicTypes.MessageStartEvent, ctx types.StreamIndexContext, log logger.Logger) (*types.StreamEventContract, error) {
	contract := &types.StreamEventContract{
		Type:         types.StreamEventMessageStart,
		Source:       types.StreamSourceAnthropic,
		ResponseID:   event.Message.ID,
		MessageID:    event.Message.ID,
		OutputIndex:  -1, // 缺失索引使用 -1 哨兵
		ContentIndex: -1, // 缺失索引使用 -1 哨兵
		Extensions:   make(map[string]interface{}),
	}

	// 保存 message_id 到上下文，供后续事件使用
	ctx.SetMessageID(event.Message.ID)

	// 转换 Message
	message := &types.StreamMessagePayload{
		Role: string(event.Message.Role),
	}

	// 转换 Content blocks 为 Parts 和 ToolCalls
	if len(event.Message.Content) > 0 {
		parts, toolCalls, err := convertResponseContentBlocksToStreamParts(event.Message.Content, log)
		if err != nil {
			log.Error("转换 content blocks 失败", "error", err)
			return nil, errors.Wrap(errors.ErrCodeInternal, "转换 content blocks 失败", err)
		}
		message.Parts = parts
		message.ToolCalls = toolCalls
	}

	contract.Message = message

	// 保存原始 Message 到 extensions
	messageJSON, err := json.Marshal(event.Message)
	if err != nil {
		log.Warn("序列化原始 message 失败", "error", err)
	} else {
		anthropicExt := make(map[string]interface{})
		anthropicExt["message"] = json.RawMessage(messageJSON)
		contract.Extensions["anthropic"] = anthropicExt
	}

	// 补齐索引字段
	fillMissingIndices(contract, contract.MessageID, 0, message.ToolCalls, -1, ctx)

	// 保存 item_id 到上下文，供后续事件使用
	if contract.ItemID != "" {
		ctx.SetItemID(contract.ItemID)
	}

	log.Debug("转换 message_start 事件完成", "response_id", contract.ResponseID)
	return contract, nil
}

// convertMessageDeltaEvent 转换 message_delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.1）：
//   - delta.stop_reason: delta.raw.stop_reason
//   - delta.stop_sequence: delta.raw.stop_sequence
//   - usage: usage.input_tokens/output_tokens，原样进入 usage.raw
//
// 索引补齐规则（plans/stream-index-repair-plan.md B.2）：
//   - sequence_number：若为 0 -> ctx.NextSequence()
//   - output_index：默认 0
//   - item_id：不适用（message_delta 不包含 item 信息）
//   - content_index：不适用
func convertMessageDeltaEvent(event *anthropicTypes.MessageDeltaEvent, ctx types.StreamIndexContext, log logger.Logger) (*types.StreamEventContract, error) {
	contract := &types.StreamEventContract{
		Type:         types.StreamEventMessageDelta,
		Source:       types.StreamSourceAnthropic,
		OutputIndex:  -1, // 缺失索引使用 -1 哨兵
		ContentIndex: -1, // 缺失索引使用 -1 哨兵
	}

	// 转换 Delta
	if event.Delta.StopReason != nil || event.Delta.StopSequence != nil {
		delta := &types.StreamDeltaPayload{
			DeltaType: "other",
			Raw:       make(map[string]interface{}),
		}

		if event.Delta.StopReason != nil {
			delta.Raw["stop_reason"] = *event.Delta.StopReason
		}
		if event.Delta.StopSequence != nil {
			delta.Raw["stop_sequence"] = *event.Delta.StopSequence
		}

		contract.Delta = delta
	}

	// 转换 Usage
	if event.Usage != nil {
		usage := &types.StreamUsagePayload{
			InputTokens:  event.Usage.InputTokens,
			OutputTokens: event.Usage.OutputTokens,
			Raw:          make(map[string]interface{}),
		}

		// 计算 TotalTokens
		if event.Usage.InputTokens != nil && event.Usage.OutputTokens != nil {
			total := *event.Usage.InputTokens + *event.Usage.OutputTokens
			usage.TotalTokens = &total
		}

		// Anthropic 特有字段放入 raw
		if event.Usage.CacheCreationInputTokens != nil {
			usage.Raw["cache_creation_input_tokens"] = *event.Usage.CacheCreationInputTokens
		}
		if event.Usage.CacheReadInputTokens != nil {
			usage.Raw["cache_read_input_tokens"] = *event.Usage.CacheReadInputTokens
		}
		if event.Usage.ServerToolUse != nil {
			usage.Raw["server_tool_use"] = event.Usage.ServerToolUse
		}

		contract.Usage = usage
	}

	// 从上下文获取 message_id，确保 item_id 一致性
	messageID := ctx.GetMessageID()

	// 补齐索引字段
	fillMissingIndices(contract, messageID, 0, nil, -1, ctx)

	log.Debug("转换 message_delta 事件完成")
	return contract, nil
}

// convertMessageStopEvent 转换 message_stop 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.1）：
//   - 简单事件，仅设置 type
//
// 索引补齐规则（plans/stream-index-repair-plan.md B.2）：
//   - sequence_number：若为 0 -> ctx.NextSequence()
//   - output_index：默认 0
//   - item_id：不适用
//   - content_index：不适用
func convertMessageStopEvent(ctx types.StreamIndexContext, log logger.Logger) (*types.StreamEventContract, error) {
	contract := &types.StreamEventContract{
		Type:         types.StreamEventMessageStop,
		Source:       types.StreamSourceAnthropic,
		OutputIndex:  -1, // 缺失索引使用 -1 哨兵
		ContentIndex: -1, // 缺失索引使用 -1 哨兵
	}

	// 从上下文获取 message_id，确保 item_id 一致性
	messageID := ctx.GetMessageID()

	// 补齐索引字段
	fillMissingIndices(contract, messageID, 0, nil, -1, ctx)

	log.Debug("转换 message_stop 事件完成")
	return contract, nil
}
