package chat

import (
	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	chatTypes "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
)

// StreamEventFormContract 将单个 StreamEventContract 转换为 OpenAI Chat 流式响应块。
//
// 根据 plans/stream-event-contract-plan.md 5.3 章节的反向映射实现：
//   - message_start: 输出 role-only chunk（delta.role 仅含角色）
//   - message_delta: 输出内容增量 chunk（delta.content/refusal/tool_calls/logprobs）
//   - message_stop: 输出 finish_reason chunk 和 usage
//   - error: 返回错误，不产生 stream chunk
//   - extensions.openai_chat: 回填 object/service_tier/system_fingerprint
func StreamEventFormContract(contract *adapterTypes.StreamEventContract, log logger.Logger) (*chatTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	log = log.WithGroup("stream_converter")

	// 验证事件来源
	if contract.Source != adapterTypes.StreamSourceOpenAIChat {
		return nil, errors.New(errors.ErrCodeInvalidArgument,
			"事件来源不匹配，期望 openai.chat，实际 "+string(contract.Source))
	}

	switch contract.Type {
	case adapterTypes.StreamEventMessageStart:
		chunk, err := convertMessageStartToChatChunk(contract, log)
		if err != nil {
			log.Error("转换 message_start 失败", "error", err)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 message_start 失败", err)
		}
		return chunk, nil

	case adapterTypes.StreamEventMessageDelta:
		chunk, err := convertMessageDeltaToChatChunk(contract, log)
		if err != nil {
			log.Error("转换 message_delta 失败", "error", err)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 message_delta 失败", err)
		}
		return chunk, nil

	case adapterTypes.StreamEventMessageStop:
		chunk, err := convertMessageStopToChatChunk(contract, log)
		if err != nil {
			log.Error("转换 message_stop 失败", "error", err)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 message_stop 失败", err)
		}
		return chunk, nil

	case adapterTypes.StreamEventError:
		// 错误事件直接返回错误，不产生 stream chunk
		if contract.Error != nil {
			return nil, errors.New(errors.ErrCodeRequestFailed, contract.Error.Message).
				WithContext("error_type", contract.Error.Type).
				WithContext("error_code", contract.Error.Code)
		}
		return nil, nil

	default:
		log.Warn("忽略不支持的事件类型", "event_type", contract.Type)
		return nil, nil
	}
}

// convertMessageStartToChatChunk 将 message_start 事件转换为 role-only chunk。
//
// 根据 OpenAI Chat 流式规范，首帧仅包含 delta.role 字段。
func convertMessageStartToChatChunk(contract *adapterTypes.StreamEventContract, log logger.Logger) (*chatTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	role := chatTypes.ChatStreamMessageRoleAssistant
	if contract.Message != nil && contract.Message.Role != "" {
		role = chatTypes.ChatStreamMessageRole(contract.Message.Role)
	}

	chunk := &chatTypes.StreamEvent{
		ID:      contract.ResponseID,
		Created: contract.CreatedAt,
		Model:   contract.Model,
		Object:  chatTypes.StreamObjectChatCompletionChunk,
		Choices: []chatTypes.StreamChoice{
			{
				Index: contract.OutputIndex,
				Delta: chatTypes.Delta{
					Role: &role,
				},
			},
		},
	}

	// 从 extensions 提取响应级字段
	extractResponseExtensions(chunk, contract.Extensions)

	log.Debug("转换 message_start 完成", "response_id", contract.ResponseID, "role", role)
	return chunk, nil
}

// convertMessageDeltaToChatChunk 将 message_delta 事件转换为内容增量 chunk。
//
// 映射规则：
//   - message.content_text -> delta.content
//   - message.tool_calls -> delta.tool_calls
//   - content.kind=refusal -> delta.refusal
//   - extensions.openai_chat -> logprobs（finish_reason 仅在 message_stop 输出）
//
// 注意：usage 仅在 message_stop 输出，不在 message_delta 输出
func convertMessageDeltaToChatChunk(contract *adapterTypes.StreamEventContract, log logger.Logger) (*chatTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	chunk := &chatTypes.StreamEvent{
		ID:      contract.ResponseID,
		Created: contract.CreatedAt,
		Model:   contract.Model,
		Object:  chatTypes.StreamObjectChatCompletionChunk,
		Choices: []chatTypes.StreamChoice{
			{
				Index: contract.OutputIndex,
				Delta: chatTypes.Delta{},
			},
		},
	}

	choice := &chunk.Choices[0]

	// 转换 message.content_text
	if contract.Message != nil && contract.Message.ContentText != nil {
		choice.Delta.Content = contract.Message.ContentText
	}

	// 转换 message.tool_calls
	if contract.Message != nil && len(contract.Message.ToolCalls) > 0 {
		toolCalls := convertStreamToolCallsToToolCallChunks(contract.Message.ToolCalls)
		choice.Delta.ToolCalls = toolCalls
	}

	// 转换 content.kind=refusal
	if contract.Content != nil && contract.Content.Kind == "refusal" {
		choice.Delta.Refusal = contract.Content.Text
	}

	// 从 extensions 提取响应级字段
	extractResponseExtensions(chunk, contract.Extensions)

	// 从 extensions 提取 choice 级字段（仅 logprobs，finish_reason 仅在 message_stop 输出）
	extractChoiceExtensions(choice, contract.Extensions)

	log.Debug("转换 message_delta 完成", "response_id", contract.ResponseID)
	return chunk, nil
}

// convertMessageStopToChatChunk 将 message_stop 事件转换为 finish_reason chunk。
//
// 映射规则：
//   - extensions.openai_chat.finish_reason -> choice.finish_reason
//   - usage -> chunk.usage（usage 仅在终止帧输出）
func convertMessageStopToChatChunk(contract *adapterTypes.StreamEventContract, log logger.Logger) (*chatTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	chunk := &chatTypes.StreamEvent{
		ID:      contract.ResponseID,
		Created: contract.CreatedAt,
		Model:   contract.Model,
		Object:  chatTypes.StreamObjectChatCompletionChunk,
		Choices: []chatTypes.StreamChoice{
			{
				Index: contract.OutputIndex,
			},
		},
	}

	choice := &chunk.Choices[0]

	// 从 extensions 提取 finish_reason
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_chat"].(map[string]interface{}); ok {
			if finishReason, ok := openaiExt["finish_reason"].(string); ok {
				fr := chatTypes.FinishReason(finishReason)
				choice.FinishReason = &fr
			}
		}
	}

	// 转换 usage（usage 仅在终止帧输出）
	if contract.Usage != nil {
		chunk.Usage = convertStreamUsageToChatUsage(contract.Usage)
	}

	// 从 extensions 提取响应级字段
	extractResponseExtensions(chunk, contract.Extensions)

	log.Debug("转换 message_stop 完成", "response_id", contract.ResponseID, "finish_reason", choice.FinishReason)
	return chunk, nil
}

// convertStreamToolCallsToToolCallChunks 将 StreamToolCall 转换为 ToolCallChunk 列表。
func convertStreamToolCallsToToolCallChunks(toolCalls []adapterTypes.StreamToolCall) []chatTypes.ToolCallChunk {
	if len(toolCalls) == 0 {
		return nil
	}

	chunks := make([]chatTypes.ToolCallChunk, 0, len(toolCalls))
	for _, tc := range toolCalls {
		chunk := chatTypes.ToolCallChunk{}

		// 从 raw 提取 index
		if tc.Raw != nil {
			if idx, ok := tc.Raw["index"].(float64); ok {
				chunk.Index = int(idx)
			}
		}

		// 设置 ID
		if tc.ID != "" {
			chunk.ID = &tc.ID
		}

		// 设置 Type
		if tc.Type != "" {
			t := chatTypes.ToolCallType(tc.Type)
			chunk.Type = &t
		} else {
			t := chatTypes.ToolCallTypeFunction
			chunk.Type = &t
		}

		// 设置 Function
		if tc.Name != "" || tc.Arguments != "" {
			chunk.Function = &chatTypes.ToolCallChunkFunction{}
			if tc.Name != "" {
				chunk.Function.Name = &tc.Name
			}
			if tc.Arguments != "" {
				chunk.Function.Arguments = &tc.Arguments
			}
		}

		chunks = append(chunks, chunk)
	}

	return chunks
}

// convertStreamUsageToChatUsage 将 StreamUsagePayload 转换为 Chat Usage。
func convertStreamUsageToChatUsage(usage *adapterTypes.StreamUsagePayload) *chatTypes.Usage {
	if usage == nil {
		return nil
	}

	result := &chatTypes.Usage{}

	if usage.InputTokens != nil {
		result.PromptTokens = *usage.InputTokens
	}
	if usage.OutputTokens != nil {
		result.CompletionTokens = *usage.OutputTokens
	}
	if usage.TotalTokens != nil {
		result.TotalTokens = *usage.TotalTokens
	}

	// 从 raw 提取扩展字段
	if usage.Raw != nil {
		if val, ok := usage.Raw["prompt_tokens_details"]; ok {
			if details, ok := val.(*chatTypes.PromptTokensDetails); ok {
				result.PromptTokensDetails = details
			}
		}
		if val, ok := usage.Raw["completion_tokens_details"]; ok {
			if details, ok := val.(*chatTypes.CompletionTokensDetails); ok {
				result.CompletionTokensDetails = details
			}
		}
	}

	return result
}

// extractResponseExtensions 从 extensions 提取响应级字段。
func extractResponseExtensions(chunk *chatTypes.StreamEvent, extensions map[string]interface{}) {
	if extensions == nil {
		return
	}

	// 从 openai_chat 命名空间提取
	if openaiExt, ok := extensions["openai_chat"].(map[string]interface{}); ok {
		if val, ok := openaiExt["object"].(string); ok {
			chunk.Object = chatTypes.StreamObject(val)
		}
		if val, ok := openaiExt["service_tier"].(string); ok {
			st := shared.ServiceTier(val)
			chunk.ServiceTier = &st
		}
		if val, ok := openaiExt["system_fingerprint"].(string); ok {
			chunk.SystemFingerprint = &val
		}
	}
}

// extractChoiceExtensions 从 extensions 提取 choice 级字段。
func extractChoiceExtensions(choice *chatTypes.StreamChoice, extensions map[string]interface{}) {
	if extensions == nil {
		return
	}

	// 从 openai_chat 命名空间提取
	if openaiExt, ok := extensions["openai_chat"].(map[string]interface{}); ok {
		if val, ok := openaiExt["logprobs"]; ok {
			if logprobs, ok := val.(*chatTypes.Logprobs); ok {
				choice.Logprobs = logprobs
			}
		}
	}
}
