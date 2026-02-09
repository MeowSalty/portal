package chat

import (
	"github.com/MeowSalty/portal/logger"
	chatTypes "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
)

// StreamEventToContract 将 OpenAI Chat 流式响应块转换为统一的 StreamEventContract 列表。
//
// 根据 plans/stream-event-contract-plan.md 5.3 章节实现：
//   - 每个 chunk 默认映射为 message_delta
//   - delta.role 且无内容时可补发 message_start
//   - finish_reason 存在时追加 message_stop
//   - refusal/tool_calls/usage/extensions 按规则映射
func StreamEventToContract(event *chatTypes.StreamEvent, log logger.Logger) ([]*adapterTypes.StreamEventContract, error) {
	if event == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	log = log.WithGroup("stream_converter")

	usage := convertUsageToStreamUsage(event.Usage)
	baseExtensions := buildOpenAIChatBaseExtensions(event)

	if len(event.Choices) == 0 {
		contract := &adapterTypes.StreamEventContract{
			Type:       adapterTypes.StreamEventMessageDelta,
			Source:     adapterTypes.StreamSourceOpenAIChat,
			ResponseID: event.ID,
			MessageID:  event.ID,
			CreatedAt:  event.Created,
			Model:      event.Model,
			Usage:      usage,
			Extensions: cloneExtensions(baseExtensions),
		}
		log.Warn("OpenAI Chat 流式响应缺少 choices", "response_id", event.ID)
		return []*adapterTypes.StreamEventContract{contract}, nil
	}

	events := make([]*adapterTypes.StreamEventContract, 0, len(event.Choices)*2)
	for _, choice := range event.Choices {
		extensions := cloneExtensions(baseExtensions)
		mergeOpenAIChatChoiceExtensions(extensions, &choice)

		if shouldEmitMessageStart(&choice.Delta) {
			startEvent := buildMessageStartEvent(event, &choice, extensions)
			events = append(events, startEvent)
		}

		deltaEvent := buildMessageDeltaEvent(event, &choice, usage, extensions)
		events = append(events, deltaEvent)

		if choice.FinishReason != nil {
			stopEvent := buildMessageStopEvent(event, &choice, extensions)
			auditLog := log.WithGroup("audit")
			auditLog.Info("追加 OpenAI Chat message_stop 事件", "response_id", stopEvent.ResponseID, "output_index", stopEvent.OutputIndex, "finish_reason", *choice.FinishReason)
			events = append(events, stopEvent)
		}
	}

	return events, nil
}

// buildMessageStartEvent 构建 OpenAI Chat 的 message_start 事件。
func buildMessageStartEvent(event *chatTypes.StreamEvent, choice *chatTypes.StreamChoice, extensions map[string]interface{}) *adapterTypes.StreamEventContract {
	message := buildStreamMessagePayload(&choice.Delta)

	contract := &adapterTypes.StreamEventContract{
		Type:        adapterTypes.StreamEventMessageStart,
		Source:      adapterTypes.StreamSourceOpenAIChat,
		ResponseID:  event.ID,
		MessageID:   event.ID,
		OutputIndex: choice.Index,
		CreatedAt:   event.Created,
		Model:       event.Model,
		Message:     message,
		Extensions:  cloneExtensions(extensions),
	}

	return contract
}

// buildMessageDeltaEvent 构建 OpenAI Chat 的 message_delta 事件。
func buildMessageDeltaEvent(event *chatTypes.StreamEvent, choice *chatTypes.StreamChoice, usage *adapterTypes.StreamUsagePayload, extensions map[string]interface{}) *adapterTypes.StreamEventContract {
	contract := &adapterTypes.StreamEventContract{
		Type:        adapterTypes.StreamEventMessageDelta,
		Source:      adapterTypes.StreamSourceOpenAIChat,
		ResponseID:  event.ID,
		MessageID:   event.ID,
		OutputIndex: choice.Index,
		CreatedAt:   event.Created,
		Model:       event.Model,
		Usage:       usage,
		Extensions:  cloneExtensions(extensions),
	}

	message := buildStreamMessagePayload(&choice.Delta)
	if message != nil {
		contract.Message = message
	}

	content := buildRefusalContent(&choice.Delta)
	if content != nil {
		contract.Content = content
	}

	return contract
}

// buildMessageStopEvent 构建 OpenAI Chat 的 message_stop 事件。
func buildMessageStopEvent(event *chatTypes.StreamEvent, choice *chatTypes.StreamChoice, extensions map[string]interface{}) *adapterTypes.StreamEventContract {
	contract := &adapterTypes.StreamEventContract{
		Type:        adapterTypes.StreamEventMessageStop,
		Source:      adapterTypes.StreamSourceOpenAIChat,
		ResponseID:  event.ID,
		MessageID:   event.ID,
		OutputIndex: choice.Index,
		CreatedAt:   event.Created,
		Model:       event.Model,
		Extensions:  cloneExtensions(extensions),
	}

	return contract
}

// shouldEmitMessageStart 判断是否需要补发 message_start 事件。
func shouldEmitMessageStart(delta *chatTypes.Delta) bool {
	if delta == nil || delta.Role == nil {
		return false
	}

	return delta.Content == nil && delta.Refusal == nil && len(delta.ToolCalls) == 0 && delta.FunctionCall == nil
}

// buildStreamMessagePayload 将 Chat delta 转换为 StreamMessagePayload。
func buildStreamMessagePayload(delta *chatTypes.Delta) *adapterTypes.StreamMessagePayload {
	if delta == nil {
		return nil
	}

	message := &adapterTypes.StreamMessagePayload{}

	if delta.Role != nil {
		message.Role = string(*delta.Role)
	}
	if delta.Content != nil {
		message.ContentText = delta.Content
	}

	toolCalls := convertToolCallChunks(delta.ToolCalls)
	if delta.FunctionCall != nil {
		legacyCall := adapterTypes.StreamToolCall{
			Type:      "function",
			Name:      delta.FunctionCall.Name,
			Arguments: delta.FunctionCall.Arguments,
			Raw: map[string]interface{}{
				"legacy_function_call": true,
			},
		}
		toolCalls = append(toolCalls, legacyCall)
	}
	if len(toolCalls) > 0 {
		message.ToolCalls = toolCalls
	}

	if message.Role == "" && message.ContentText == nil && len(message.ToolCalls) == 0 {
		return nil
	}

	return message
}

// buildRefusalContent 将 refusal 映射到 StreamContentPayload。
func buildRefusalContent(delta *chatTypes.Delta) *adapterTypes.StreamContentPayload {
	if delta == nil || delta.Refusal == nil {
		return nil
	}

	return &adapterTypes.StreamContentPayload{
		Kind: "refusal",
		Text: delta.Refusal,
	}
}

// convertToolCallChunks 转换流式工具调用增量。
func convertToolCallChunks(chunks []chatTypes.ToolCallChunk) []adapterTypes.StreamToolCall {
	if len(chunks) == 0 {
		return nil
	}

	calls := make([]adapterTypes.StreamToolCall, 0, len(chunks))
	for _, chunk := range chunks {
		call := adapterTypes.StreamToolCall{
			Raw: map[string]interface{}{
				"index": chunk.Index,
			},
		}

		if chunk.ID != nil {
			call.ID = *chunk.ID
		}
		if chunk.Type != nil {
			call.Type = string(*chunk.Type)
		} else {
			call.Type = "function"
		}
		if chunk.Function != nil {
			if chunk.Function.Name != nil {
				call.Name = *chunk.Function.Name
			}
			if chunk.Function.Arguments != nil {
				call.Arguments = *chunk.Function.Arguments
			}
		}

		if len(call.Raw) == 0 {
			call.Raw = nil
		}
		calls = append(calls, call)
	}

	return calls
}

// convertUsageToStreamUsage 转换 OpenAI Chat usage 到 StreamUsagePayload。
func convertUsageToStreamUsage(usage *chatTypes.Usage) *adapterTypes.StreamUsagePayload {
	if usage == nil {
		return nil
	}

	result := &adapterTypes.StreamUsagePayload{
		Raw: make(map[string]interface{}),
	}

	if usage.PromptTokens > 0 {
		inputTokens := usage.PromptTokens
		result.InputTokens = &inputTokens
	}
	if usage.CompletionTokens > 0 {
		outputTokens := usage.CompletionTokens
		result.OutputTokens = &outputTokens
	}
	if usage.TotalTokens > 0 {
		totalTokens := usage.TotalTokens
		result.TotalTokens = &totalTokens
	} else if result.InputTokens != nil && result.OutputTokens != nil {
		totalTokens := *result.InputTokens + *result.OutputTokens
		result.TotalTokens = &totalTokens
	}

	if usage.PromptTokensDetails != nil {
		result.Raw["prompt_tokens_details"] = usage.PromptTokensDetails
	}
	if usage.CompletionTokensDetails != nil {
		result.Raw["completion_tokens_details"] = usage.CompletionTokensDetails
	}
	if len(result.Raw) == 0 {
		result.Raw = nil
	}

	return result
}

// buildOpenAIChatBaseExtensions 构建 OpenAI Chat 响应级扩展字段。
func buildOpenAIChatBaseExtensions(event *chatTypes.StreamEvent) map[string]interface{} {
	if event == nil {
		return nil
	}

	openaiExt := make(map[string]interface{})
	if event.Object != "" {
		openaiExt["object"] = event.Object
	}
	if event.ServiceTier != nil {
		openaiExt["service_tier"] = *event.ServiceTier
	}
	if event.SystemFingerprint != nil {
		openaiExt["system_fingerprint"] = *event.SystemFingerprint
	}

	if len(openaiExt) == 0 {
		return nil
	}

	return map[string]interface{}{
		"openai_chat": openaiExt,
	}
}

// mergeOpenAIChatChoiceExtensions 合并 OpenAI Chat choice 级扩展字段。
func mergeOpenAIChatChoiceExtensions(extensions map[string]interface{}, choice *chatTypes.StreamChoice) {
	if choice == nil {
		return
	}

	openaiExt := ensureOpenAIChatExtensions(extensions)
	if choice.FinishReason != nil {
		openaiExt["finish_reason"] = *choice.FinishReason
	}
	if choice.Logprobs != nil {
		openaiExt["logprobs"] = choice.Logprobs
	}
	if len(choice.Delta.ExtraFields) > 0 {
		openaiExt["delta_extra_fields"] = choice.Delta.ExtraFields
	}
}

// ensureOpenAIChatExtensions 获取或创建 OpenAI Chat 扩展命名空间。
func ensureOpenAIChatExtensions(extensions map[string]interface{}) map[string]interface{} {
	if extensions == nil {
		extensions = make(map[string]interface{})
	}

	if ext, ok := extensions["openai_chat"]; ok {
		if openaiExt, ok := ext.(map[string]interface{}); ok {
			return openaiExt
		}
	}

	openaiExt := make(map[string]interface{})
	extensions["openai_chat"] = openaiExt
	return openaiExt
}

// cloneExtensions 复制扩展字段，避免复用引用。
func cloneExtensions(extensions map[string]interface{}) map[string]interface{} {
	if len(extensions) == 0 {
		return nil
	}

	clone := make(map[string]interface{}, len(extensions))
	for key, value := range extensions {
		clone[key] = value
	}
	return clone
}
