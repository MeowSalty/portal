package responses

import (
	"encoding/json"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	helper "github.com/MeowSalty/portal/request/adapter/openai/converter/responses/helper"
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
)

// unmarshalContentPart 从 map[string]interface{} 反序列化为 OutputContentPart。
// 用于从 contract 中恢复 part 字段。
func unmarshalContentPart(data map[string]interface{}) (responsesTypes.OutputContentPart, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return responsesTypes.OutputContentPart{}, err
	}
	var part responsesTypes.OutputContentPart
	if err := json.Unmarshal(jsonData, &part); err != nil {
		return responsesTypes.OutputContentPart{}, err
	}
	return part, nil
}

// ensureIndexFields 确保流事件中的索引字段完整
//
// 根据 plans/stream-index-repair-plan.md C 章节的兜底规则实现：
//   - 仅对非 OpenAI Responses 原生事件应用兜底逻辑（contract.Source != openai.responses）
//   - sequence_number: 若为 0 -> ctx.NextSequence()
//   - item_id: 若为空 -> 优先 MessageID；仍为空 -> ctx.EnsureItemID(BuildStreamIndexKey(...))
//   - output_index: 若为负值（< 0） -> ctx.EnsureOutputIndex(response_id)，0 值保持不变
//   - content_index: 若为负值（< 0）且事件类型为 content_part/output_text -> ctx.EnsureContentIndex(item_id, -1)，0 值保持不变
//   - annotation_index: 若为负值（< 0）且事件类型为 annotation -> ctx.EnsureAnnotationIndex(item_id, -1)，0 值保持不变
//
// 参数：
//   - contract: 流事件合约（不会被修改，返回副本）
//   - ctx: 流索引上下文
//   - log: 日志记录器
//
// 返回：
//   - 补齐后的流事件合约副本
func ensureIndexFields(contract *adapterTypes.StreamEventContract, ctx adapterTypes.StreamIndexContext, log logger.Logger) *adapterTypes.StreamEventContract {
	if contract == nil || ctx == nil {
		return contract
	}

	// 创建副本，避免修改原始 contract
	result := *contract

	// 仅对非 OpenAI Responses 原生事件应用兜底逻辑
	isNativeEvent := result.Source == adapterTypes.StreamSourceOpenAIResponse

	// 补齐 sequence_number
	if result.SequenceNumber == 0 {
		result.SequenceNumber = ctx.NextSequence()
		log.Debug("补齐 sequence_number", "sequence_number", result.SequenceNumber)
	}

	// 补齐 item_id（仅针对非原生事件）
	if result.ItemID == "" && !isNativeEvent {
		// 优先使用 MessageID
		if result.MessageID != "" {
			result.ItemID = result.MessageID
			log.Debug("使用 MessageID 作为 item_id", "item_id", result.ItemID)
		} else {
			// 仍为空则生成新的 item_id
			responseID := result.ResponseID
			if responseID == "" {
				responseID = "unknown"
			}
			key := adapterTypes.BuildStreamIndexKey(responseID, result.OutputIndex, result.ContentIndex)
			result.ItemID = ctx.EnsureItemID(key)
			log.Debug("生成新的 item_id", "item_id", result.ItemID, "key", key)
		}
	}

	// 补齐 output_index（仅针对非原生事件，且仅处理负值）
	if result.OutputIndex < 0 && !isNativeEvent {
		responseID := result.ResponseID
		if responseID == "" {
			responseID = result.ItemID
		}
		if responseID != "" {
			result.OutputIndex = ctx.EnsureOutputIndex(responseID)
			log.Debug("补齐 output_index", "output_index", result.OutputIndex, "response_id", responseID)
		}
	}

	// 补齐 content_index（仅针对非原生事件，且仅处理负值）
	if result.ContentIndex < 0 && !isNativeEvent {
		switch result.Type {
		case adapterTypes.StreamEventContentPartAdded,
			adapterTypes.StreamEventContentPartDone,
			adapterTypes.StreamEventOutputTextDelta,
			adapterTypes.StreamEventOutputTextDone,
			adapterTypes.StreamEventOutputTextAnnotationAdded,
			adapterTypes.StreamEventRefusalDelta,
			adapterTypes.StreamEventRefusalDone,
			adapterTypes.StreamEventReasoningTextDelta,
			adapterTypes.StreamEventReasoningTextDone:
			result.ContentIndex = ctx.EnsureContentIndex(result.ItemID, -1)
			log.Debug("补齐 content_index", "content_index", result.ContentIndex, "item_id", result.ItemID)
		}
	}

	// 补齐 annotation_index（仅针对非原生事件，且仅处理负值）
	if result.AnnotationIndex < 0 && result.Type == adapterTypes.StreamEventOutputTextAnnotationAdded && !isNativeEvent {
		result.AnnotationIndex = ctx.EnsureAnnotationIndex(result.ItemID, -1)
		log.Debug("补齐 annotation_index", "annotation_index", result.AnnotationIndex, "item_id", result.ItemID)
	}

	return &result
}

// StreamEventFormContract 将单个 StreamEventContract 转换为 OpenAI Responses 流式事件列表。
//
// 根据 plans/stream-event-contract-plan.md 5.4 章节的反向映射实现：
//   - response.* 事件：直接映射到对应的 response.* 事件类型
//   - output_item.* 事件：直接映射到对应的 output_item.* 事件类型
//   - content_part.* 事件：直接映射到对应的 content_part.* 事件类型
//   - output_text.* 事件：直接映射到对应的 output_text.* 事件类型
//   - refusal.* 事件：直接映射到对应的 refusal.* 事件类型
//   - reasoning_* 事件：直接映射到对应的 reasoning_* 事件类型
//   - function_call_arguments.* 事件：直接映射到对应的 function_call_arguments.* 事件类型
//   - custom_tool_call_input.* 事件：直接映射到对应的 custom_tool_call_input.* 事件类型
//   - mcp_call_arguments.* 事件：直接映射到对应的 mcp_call_arguments.* 事件类型
//   - mcp_call.* 事件：直接映射到对应的 mcp_call.* 事件类型
//   - mcp_list_tools.* 事件：直接映射到对应的 mcp_list_tools.* 事件类型
//   - audio.* 事件：直接映射到对应的 audio.* 事件类型
//   - code_interpreter_call.* 事件：直接映射到对应的 code_interpreter_call.* 事件类型
//   - file_search_call.* 事件：直接映射到对应的 file_search_call.* 事件类型
//   - web_search_call.* 事件：直接映射到对应的 web_search_call.* 事件类型
//   - image_generation_call.* 事件：直接映射到对应的 image_generation_call.* 事件类型
//   - error 事件：映射到 error 事件类型
//
// 降级映射规则（非 Responses 事件）：
//   - message_start: 降级为 response.created
//   - message_delta: 降级为 response.output_text.delta（提取文本内容）
//   - message_stop: 降级为 response.output_text.done + response.completed
//   - content_block_start: 降级为 response.output_item.added + response.content_part.added
//   - content_block_delta: 降级为 response.output_text.delta（提取文本内容）
//   - content_block_stop: 降级为 response.content_part.done + response.output_item.done
//   - ping: 忽略（不产生输出事件）
//
// 返回值：
//   - 对于大多数事件，返回包含 1 个事件的切片
//   - 对于需要降级为多个事件的情况（如 message_stop, content_block_start, content_block_stop），
//     返回包含 2 个事件的切片
//   - 对于 ping 事件，返回空切片
//   - 对于 nil 输入，返回 nil
func StreamEventFormContract(contract *adapterTypes.StreamEventContract, log logger.Logger, indexCtx adapterTypes.StreamIndexContext) ([]*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	log = log.WithGroup("stream_converter")

	// 保存原始 sequence_number，用于判断是否由补齐生成
	seqProvided := contract.SequenceNumber

	// 补齐缺失的索引字段（兜底逻辑）
	// 仅对缺失或零值进行补齐，已存在的非零值保持不变
	contract = ensureIndexFields(contract, indexCtx, log)

	// 转换事件
	var events []*responsesTypes.StreamEvent

	switch contract.Type {
	// Response 生命周期事件
	case adapterTypes.StreamEventResponseCreated:
		event, convertErr := convertResponseCreatedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 response.created 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 response.created 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventResponseInProgress:
		event, convertErr := convertResponseInProgressFromContract(contract)
		if convertErr != nil {
			log.Error("转换 response.in_progress 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 response.in_progress 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventResponseCompleted:
		event, convertErr := convertResponseCompletedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 response.completed 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 response.completed 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventResponseFailed:
		event, convertErr := convertResponseFailedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 response.failed 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 response.failed 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventResponseIncomplete:
		event, convertErr := convertResponseIncompleteFromContract(contract)
		if convertErr != nil {
			log.Error("转换 response.incomplete 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 response.incomplete 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventResponseQueued:
		event, convertErr := convertResponseQueuedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 response.queued 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 response.queued 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Output Item 事件
	case adapterTypes.StreamEventOutputItemAdded:
		event, convertErr := convertOutputItemAddedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 output_item.added 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 output_item.added 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventOutputItemDone:
		event, convertErr := convertOutputItemDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 output_item.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 output_item.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Content Part 事件
	case adapterTypes.StreamEventContentPartAdded:
		event, convertErr := convertContentPartAddedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 content_part.added 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 content_part.added 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventContentPartDone:
		event, convertErr := convertContentPartDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 content_part.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 content_part.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Output Text 事件
	case adapterTypes.StreamEventOutputTextDelta:
		event, convertErr := convertOutputTextDeltaFromContract(contract)
		if convertErr != nil {
			log.Error("转换 output_text.delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 output_text.delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventOutputTextDone:
		event, convertErr := convertOutputTextDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 output_text.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 output_text.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventOutputTextAnnotationAdded:
		event, convertErr := convertOutputTextAnnotationAddedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 output_text.annotation.added 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 output_text.annotation.added 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Refusal 事件
	case adapterTypes.StreamEventRefusalDelta:
		event, convertErr := convertRefusalDeltaFromContract(contract)
		if convertErr != nil {
			log.Error("转换 refusal.delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 refusal.delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventRefusalDone:
		event, convertErr := convertRefusalDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 refusal.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 refusal.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Reasoning Text 事件
	case adapterTypes.StreamEventReasoningTextDelta:
		event, convertErr := convertReasoningTextDeltaFromContract(contract)
		if convertErr != nil {
			log.Error("转换 reasoning_text.delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 reasoning_text.delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventReasoningTextDone:
		event, convertErr := convertReasoningTextDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 reasoning_text.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 reasoning_text.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Reasoning Summary 事件
	case adapterTypes.StreamEventReasoningSummaryPartAdded:
		event, convertErr := convertReasoningSummaryPartAddedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 reasoning_summary_part.added 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 reasoning_summary_part.added 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventReasoningSummaryPartDone:
		event, convertErr := convertReasoningSummaryPartDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 reasoning_summary_part.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 reasoning_summary_part.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventReasoningSummaryTextDelta:
		event, convertErr := convertReasoningSummaryTextDeltaFromContract(contract)
		if convertErr != nil {
			log.Error("转换 reasoning_summary_text.delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 reasoning_summary_text.delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventReasoningSummaryTextDone:
		event, convertErr := convertReasoningSummaryTextDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 reasoning_summary_text.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 reasoning_summary_text.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Function Call 事件
	case adapterTypes.StreamEventFunctionCallArgumentsDelta:
		event, convertErr := convertFunctionCallArgumentsDeltaFromContract(contract)
		if convertErr != nil {
			log.Error("转换 function_call_arguments.delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 function_call_arguments.delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventFunctionCallArgumentsDone:
		event, convertErr := convertFunctionCallArgumentsDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 function_call_arguments.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 function_call_arguments.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Custom Tool Call 事件
	case adapterTypes.StreamEventCustomToolCallInputDelta:
		event, convertErr := convertCustomToolCallInputDeltaFromContract(contract)
		if convertErr != nil {
			log.Error("转换 custom_tool_call_input.delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 custom_tool_call_input.delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventCustomToolCallInputDone:
		event, convertErr := convertCustomToolCallInputDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 custom_tool_call_input.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 custom_tool_call_input.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// MCP Call 事件
	case adapterTypes.StreamEventMCPCallArgumentsDelta:
		event, convertErr := convertMCPCallArgumentsDeltaFromContract(contract)
		if convertErr != nil {
			log.Error("转换 mcp_call_arguments.delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 mcp_call_arguments.delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventMCPCallArgumentsDone:
		event, convertErr := convertMCPCallArgumentsDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 mcp_call_arguments.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 mcp_call_arguments.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventMCPCallCompleted:
		event, convertErr := convertMCPCallCompletedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 mcp_call.completed 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 mcp_call.completed 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventMCPCallFailed:
		event, convertErr := convertMCPCallFailedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 mcp_call.failed 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 mcp_call.failed 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventMCPCallInProgress:
		event, convertErr := convertMCPCallInProgressFromContract(contract)
		if convertErr != nil {
			log.Error("转换 mcp_call.in_progress 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 mcp_call.in_progress 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventMCPListToolsCompleted:
		event, convertErr := convertMCPListToolsCompletedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 mcp_list_tools.completed 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 mcp_list_tools.completed 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventMCPListToolsFailed:
		event, convertErr := convertMCPListToolsFailedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 mcp_list_tools.failed 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 mcp_list_tools.failed 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventMCPListToolsInProgress:
		event, convertErr := convertMCPListToolsInProgressFromContract(contract)
		if convertErr != nil {
			log.Error("转换 mcp_list_tools.in_progress 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 mcp_list_tools.in_progress 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Audio 事件
	case adapterTypes.StreamEventAudioDelta:
		event, convertErr := convertAudioDeltaFromContract(contract)
		if convertErr != nil {
			log.Error("转换 audio.delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 audio.delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventAudioDone:
		event, convertErr := convertAudioDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 audio.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 audio.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventAudioTranscriptDelta:
		event, convertErr := convertAudioTranscriptDeltaFromContract(contract)
		if convertErr != nil {
			log.Error("转换 audio.transcript.delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 audio.transcript.delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventAudioTranscriptDone:
		event, convertErr := convertAudioTranscriptDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 audio.transcript.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 audio.transcript.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Code Interpreter 事件
	case adapterTypes.StreamEventCodeInterpreterCallCodeDelta:
		event, convertErr := convertCodeInterpreterCallCodeDeltaFromContract(contract)
		if convertErr != nil {
			log.Error("转换 code_interpreter_call_code.delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 code_interpreter_call_code.delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventCodeInterpreterCallCodeDone:
		event, convertErr := convertCodeInterpreterCallCodeDoneFromContract(contract)
		if convertErr != nil {
			log.Error("转换 code_interpreter_call_code.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 code_interpreter_call_code.done 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventCodeInterpreterCallCompleted:
		event, convertErr := convertCodeInterpreterCallCompletedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 code_interpreter_call.completed 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 code_interpreter_call.completed 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventCodeInterpreterCallInProgress:
		event, convertErr := convertCodeInterpreterCallInProgressFromContract(contract)
		if convertErr != nil {
			log.Error("转换 code_interpreter_call.in_progress 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 code_interpreter_call.in_progress 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventCodeInterpreterCallInterpreting:
		event, convertErr := convertCodeInterpreterCallInterpretingFromContract(contract)
		if convertErr != nil {
			log.Error("转换 code_interpreter_call.interpreting 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 code_interpreter_call.interpreting 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// File Search 事件
	case adapterTypes.StreamEventFileSearchCallCompleted:
		event, convertErr := convertFileSearchCallCompletedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 file_search_call.completed 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 file_search_call.completed 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventFileSearchCallInProgress:
		event, convertErr := convertFileSearchCallInProgressFromContract(contract)
		if convertErr != nil {
			log.Error("转换 file_search_call.in_progress 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 file_search_call.in_progress 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventFileSearchCallSearching:
		event, convertErr := convertFileSearchCallSearchingFromContract(contract)
		if convertErr != nil {
			log.Error("转换 file_search_call.searching 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 file_search_call.searching 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Web Search 事件
	case adapterTypes.StreamEventWebSearchCallCompleted:
		event, convertErr := convertWebSearchCallCompletedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 web_search_call.completed 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 web_search_call.completed 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventWebSearchCallInProgress:
		event, convertErr := convertWebSearchCallInProgressFromContract(contract)
		if convertErr != nil {
			log.Error("转换 web_search_call.in_progress 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 web_search_call.in_progress 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventWebSearchCallSearching:
		event, convertErr := convertWebSearchCallSearchingFromContract(contract)
		if convertErr != nil {
			log.Error("转换 web_search_call.searching 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 web_search_call.searching 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Image Generation 事件
	case adapterTypes.StreamEventImageGenerationCallCompleted:
		event, convertErr := convertImageGenCallCompletedFromContract(contract)
		if convertErr != nil {
			log.Error("转换 image_generation_call.completed 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 image_generation_call.completed 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventImageGenerationCallInProgress:
		event, convertErr := convertImageGenCallInProgressFromContract(contract)
		if convertErr != nil {
			log.Error("转换 image_generation_call.in_progress 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 image_generation_call.in_progress 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventImageGenerationCallGenerating:
		event, convertErr := convertImageGenCallGeneratingFromContract(contract)
		if convertErr != nil {
			log.Error("转换 image_generation_call.generating 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 image_generation_call.generating 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventImageGenerationCallPartialImage:
		event, convertErr := convertImageGenCallPartialImageFromContract(contract)
		if convertErr != nil {
			log.Error("转换 image_generation_call.partial_image 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 image_generation_call.partial_image 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// Error 事件
	case adapterTypes.StreamEventError:
		event, convertErr := convertStreamErrorFromContract(contract)
		if convertErr != nil {
			log.Error("转换 error 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换 error 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	// 降级映射事件（非 Responses 原生事件）
	case adapterTypes.StreamEventMessageStart:
		// 降级为 response.created
		event, convertErr := degradeMessageStartToResponseCreated(contract)
		if convertErr != nil {
			log.Error("降级 message_start 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "降级 message_start 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventMessageDelta:
		// 降级为 response.output_text.delta
		event, convertErr := degradeMessageDeltaToOutputTextDelta(contract)
		if convertErr != nil {
			log.Error("降级 message_delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "降级 message_delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventMessageStop:
		// 降级为 response.output_text.done + response.completed
		doneEvent, convertErr := degradeMessageStopToOutputTextDone(contract)
		if convertErr != nil {
			log.Error("降级 message_stop 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "降级 message_stop 失败", convertErr)
		}
		if doneEvent != nil {
			events = append(events, doneEvent)
		}

		// 添加 response.completed
		completedEvent, convertErr := degradeMessageStopToResponseCompleted(contract)
		if convertErr != nil {
			log.Error("降级 message_stop 到 response.completed 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "降级 message_stop 到 response.completed 失败", convertErr)
		}
		if completedEvent != nil {
			events = append(events, completedEvent)
		}

	case adapterTypes.StreamEventContentBlockStart:
		// 降级为 response.output_item.added + response.content_part.added
		itemEvent, convertErr := degradeContentBlockStartToOutputItemAdded(contract)
		if convertErr != nil {
			log.Error("降级 content_block_start 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "降级 content_block_start 失败", convertErr)
		}
		if itemEvent != nil {
			events = append(events, itemEvent)
		}

		partEvent, convertErr := degradeContentBlockStartToContentPartAdded(contract)
		if convertErr != nil {
			log.Error("降级 content_block_start 到 content_part.added 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "降级 content_block_start 到 content_part.added 失败", convertErr)
		}
		if partEvent != nil {
			events = append(events, partEvent)
		}

	case adapterTypes.StreamEventContentBlockDelta:
		// 降级为 response.output_text.delta
		event, convertErr := degradeContentBlockDeltaToOutputTextDelta(contract)
		if convertErr != nil {
			log.Error("降级 content_block_delta 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "降级 content_block_delta 失败", convertErr)
		}
		if event != nil {
			events = append(events, event)
		}

	case adapterTypes.StreamEventContentBlockStop:
		// 降级为 response.content_part.done + response.output_item.done
		partEvent, convertErr := degradeContentBlockStopToContentPartDone(contract)
		if convertErr != nil {
			log.Error("降级 content_block_stop 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "降级 content_block_stop 失败", convertErr)
		}
		if partEvent != nil {
			events = append(events, partEvent)
		}

		itemEvent, convertErr := degradeContentBlockStopToOutputItemDone(contract)
		if convertErr != nil {
			log.Error("降级 content_block_stop 到 output_item.done 失败", "error", convertErr)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "降级 content_block_stop 到 output_item.done 失败", convertErr)
		}
		if itemEvent != nil {
			events = append(events, itemEvent)
		}

	case adapterTypes.StreamEventPing:
		// 忽略 ping 事件
		log.Debug("忽略 ping 事件")
		return nil, nil

	default:
		log.Warn("忽略不支持的事件类型", "event_type", contract.Type)
		return nil, nil
	}

	// 修正同一输入拆分出的多个事件的 sequence_number，确保严格递增
	// 仅影响 len(events) > 1 的情况（降级拆分路径）
	if len(events) > 1 {
		for i := 1; i < len(events); i++ {
			prevSeq := helper.GetSequenceNumber(events[i-1])
			if seqProvided == 0 && indexCtx != nil {
				// 原始序号为 0，使用 indexCtx 递增
				helper.SetSequenceNumber(events[i], indexCtx.NextSequence())
			} else {
				// 原始序号非零，使用前一个序号 +1
				helper.SetSequenceNumber(events[i], prevSeq+1)
			}
		}
	}

	return events, nil
}

// LifecycleState 生命周期状态追踪器，用于在流式传输过程中补发生命周期事件。
type LifecycleState struct {
	// HasCreated 是否已发送 response.created 事件
	HasCreated bool
	// HasInProgress 是否已发送 response.in_progress 事件
	HasInProgress bool
	// HasCompleted 是否已发送 response.completed/failed/incomplete 事件
	HasCompleted bool
	// ResponseID 从事件中提取的 response_id
	ResponseID string
	// SequenceNumber 从事件中提取的基础序列号
	SequenceNumber int
}

// EnsureLifecycleEventsForStream 确保生命周期事件存在，返回需要补发的事件。
//
// 该函数在接收到每个事件后调用，检查是否需要补发 response.created
// 或 response.in_progress 事件。
//
// 参数：
//   - event: 当前收到的事件（不会被修改）
//   - state: 生命周期状态（会被更新）
//   - log: 日志记录器
//
// 返回：
//   - 需要在 event 之前补发的事件列表（0~2 个）
//   - 错误（如有）
func EnsureLifecycleEventsForStream(event *responsesTypes.StreamEvent, state *LifecycleState, log logger.Logger) ([]*responsesTypes.StreamEvent, error) {
	if event == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	var result []*responsesTypes.StreamEvent

	// 提取 response_id 和 sequence_number
	if state.ResponseID == "" || state.SequenceNumber == 0 {
		if event.Created != nil {
			state.ResponseID = event.Created.Response.ID
			state.SequenceNumber = event.Created.SequenceNumber
			state.HasCreated = true
		} else if event.InProgress != nil {
			state.ResponseID = event.InProgress.Response.ID
			state.SequenceNumber = event.InProgress.SequenceNumber
			state.HasInProgress = true
		} else if event.OutputTextDelta != nil {
			state.ResponseID = event.OutputTextDelta.ItemID
			state.SequenceNumber = event.OutputTextDelta.SequenceNumber
		}
	}

	// 如果不存在 response.created，在第一个事件前插入
	if !state.HasCreated && state.ResponseID != "" {
		createdEvent := &responsesTypes.StreamEvent{
			Created: &responsesTypes.ResponseCreatedEvent{
				Type:           responsesTypes.StreamEventCreated,
				Response:       responsesTypes.Response{ID: state.ResponseID},
				SequenceNumber: state.SequenceNumber,
			},
		}
		result = append(result, createdEvent)
		state.HasCreated = true
		log.Debug("补发 response.created 事件", "response_id", state.ResponseID)

		// 如果不存在 response.in_progress，在 response.created 后插入
		if !state.HasInProgress {
			inProgressEvent := &responsesTypes.StreamEvent{
				InProgress: &responsesTypes.ResponseInProgressEvent{
					Type:           responsesTypes.StreamEventInProgress,
					Response:       responsesTypes.Response{ID: state.ResponseID},
					SequenceNumber: state.SequenceNumber + 1,
				},
			}
			result = append(result, inProgressEvent)
			state.HasInProgress = true
			log.Debug("补发 response.in_progress 事件", "response_id", state.ResponseID)
		}
	}

	// 检查是否收到完成事件
	if event.Completed != nil || event.Failed != nil || event.Incomplete != nil {
		state.HasCompleted = true
	}

	return result, nil
}

// EnsureLifecycleEventsOnEnd 确保在流结束时补发 response.completed 事件。
//
// 该函数在流式传输结束时调用，检查是否需要补发 response.completed 事件。
//
// 参数：
//   - state: 生命周期状态
//   - log: 日志记录器
//
// 返回：
//   - 需要补发的事件列表（0 或 1 个）
func EnsureLifecycleEventsOnEnd(state *LifecycleState, log logger.Logger) []*responsesTypes.StreamEvent {
	if log == nil {
		log = logger.NewNopLogger()
	}

	// 如果不存在 response.completed，在末尾插入
	if !state.HasCompleted && state.ResponseID != "" {
		completedEvent := &responsesTypes.StreamEvent{
			Completed: &responsesTypes.ResponseCompletedEvent{
				Type:           responsesTypes.StreamEventCompleted,
				Response:       responsesTypes.Response{ID: state.ResponseID},
				SequenceNumber: state.SequenceNumber + 2,
			},
		}
		log.Debug("补发 response.completed 事件", "response_id", state.ResponseID)
		return []*responsesTypes.StreamEvent{completedEvent}
	}

	return nil
}

// 以下是各个事件的转换函数

// Response 生命周期事件转换函数

func convertResponseCreatedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	// 从 extensions 提取原始 Response
	var response responsesTypes.Response
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if resp, ok := openaiExt["response"].(responsesTypes.Response); ok {
				response = resp
			}
		}
	}

	// 如果没有原始 Response，创建一个基本的
	if response.ID == "" {
		response.ID = contract.ResponseID
	}

	return &responsesTypes.StreamEvent{
		Created: &responsesTypes.ResponseCreatedEvent{
			Type:           responsesTypes.StreamEventCreated,
			Response:       response,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertResponseInProgressFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	// 从 extensions 提取原始 Response
	var response responsesTypes.Response
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if resp, ok := openaiExt["response"].(responsesTypes.Response); ok {
				response = resp
			}
		}
	}

	// 如果没有原始 Response，创建一个基本的
	if response.ID == "" {
		response.ID = contract.ResponseID
	}

	return &responsesTypes.StreamEvent{
		InProgress: &responsesTypes.ResponseInProgressEvent{
			Type:           responsesTypes.StreamEventInProgress,
			Response:       response,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertResponseCompletedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	// 从 extensions 提取原始 Response
	var response responsesTypes.Response
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if resp, ok := openaiExt["response"].(responsesTypes.Response); ok {
				response = resp
			}
		}
	}

	// 如果没有原始 Response，创建一个基本的
	if response.ID == "" {
		response.ID = contract.ResponseID
	}

	// 转换 Usage
	if contract.Usage != nil {
		response.Usage = helper.ConvertStreamUsageToResponsesUsage(contract.Usage)
	}

	return &responsesTypes.StreamEvent{
		Completed: &responsesTypes.ResponseCompletedEvent{
			Type:           responsesTypes.StreamEventCompleted,
			Response:       response,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertResponseFailedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	// 从 extensions 提取原始 Response
	var response responsesTypes.Response
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if resp, ok := openaiExt["response"].(responsesTypes.Response); ok {
				response = resp
			}
		}
	}

	// 如果没有原始 Response，创建一个基本的
	if response.ID == "" {
		response.ID = contract.ResponseID
	}

	// 转换 Error
	if contract.Error != nil {
		response.Error = helper.ConvertStreamErrorToResponseError(contract.Error)
	}

	return &responsesTypes.StreamEvent{
		Failed: &responsesTypes.ResponseFailedEvent{
			Type:           responsesTypes.StreamEventFailed,
			Response:       response,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertResponseIncompleteFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	// 从 extensions 提取原始 Response
	var response responsesTypes.Response
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if resp, ok := openaiExt["response"].(responsesTypes.Response); ok {
				response = resp
			}
			if incompleteDetails, ok := openaiExt["incomplete_details"]; ok {
				if details, ok := incompleteDetails.(*responsesTypes.IncompleteDetails); ok {
					response.IncompleteDetails = details
				}
			}
		}
	}

	// 如果没有原始 Response，创建一个基本的
	if response.ID == "" {
		response.ID = contract.ResponseID
	}

	return &responsesTypes.StreamEvent{
		Incomplete: &responsesTypes.ResponseIncompleteEvent{
			Type:           responsesTypes.StreamEventIncomplete,
			Response:       response,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertResponseQueuedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	// 从 extensions 提取原始 Response
	var response responsesTypes.Response
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if resp, ok := openaiExt["response"].(responsesTypes.Response); ok {
				response = resp
			}
		}
	}

	// 如果没有原始 Response，创建一个基本的
	if response.ID == "" {
		response.ID = contract.ResponseID
	}

	return &responsesTypes.StreamEvent{
		Queued: &responsesTypes.ResponseQueuedEvent{
			Type:           responsesTypes.StreamEventQueued,
			Response:       response,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// Output Item 事件转换函数

func convertOutputItemAddedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	// 从 content.raw 提取 item
	var item responsesTypes.OutputItem
	if contract.Content != nil && contract.Content.Raw != nil {
		if i, ok := contract.Content.Raw["item"].(responsesTypes.OutputItem); ok {
			item = i
		}
	}

	// 如果没有原始 item 或 item.Message 为空，创建一个基本的 message
	if item.Message == nil {
		item.Message = &responsesTypes.OutputMessage{
			ID:   contract.ItemID,
			Type: responsesTypes.OutputItemTypeMessage,
		}
	} else if item.Message.ID == "" && contract.ItemID != "" {
		// Message 存在但 ID 为空时，回填 contract.ItemID
		item.Message.ID = contract.ItemID
	}

	return &responsesTypes.StreamEvent{
		OutputItemAdded: &responsesTypes.ResponseOutputItemAddedEvent{
			Type:           responsesTypes.StreamEventOutputItemAdded,
			OutputIndex:    contract.OutputIndex,
			Item:           item,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertOutputItemDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	// 从 content.raw 提取 item
	var item responsesTypes.OutputItem
	if contract.Content != nil && contract.Content.Raw != nil {
		if i, ok := contract.Content.Raw["item"].(responsesTypes.OutputItem); ok {
			item = i
		}
	}

	// 如果没有原始 item 或 item.Message 为空，创建一个基本的 message
	if item.Message == nil {
		item.Message = &responsesTypes.OutputMessage{
			ID:   contract.ItemID,
			Type: responsesTypes.OutputItemTypeMessage,
		}
	} else if item.Message.ID == "" && contract.ItemID != "" {
		// Message 存在但 ID 为空时，回填 contract.ItemID
		item.Message.ID = contract.ItemID
	}

	return &responsesTypes.StreamEvent{
		OutputItemDone: &responsesTypes.ResponseOutputItemDoneEvent{
			Type:           responsesTypes.StreamEventOutputItemDone,
			OutputIndex:    contract.OutputIndex,
			Item:           item,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// Content Part 事件转换函数

func convertContentPartAddedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	// 从 content.raw.part 中恢复 part 字段
	var part responsesTypes.OutputContentPart
	if contract.Content != nil && contract.Content.Raw != nil {
		if partData, ok := contract.Content.Raw["part"].(map[string]interface{}); ok {
			var err error
			part, err = unmarshalContentPart(partData)
			if err != nil {
				return nil, errors.Wrap(errors.ErrCodeStreamError, "反序列化 part 字段失败", err)
			}
		}
	}

	return &responsesTypes.StreamEvent{
		ContentPartAdded: &responsesTypes.ResponseContentPartAddedEvent{
			Type:           responsesTypes.StreamEventContentPartAdded,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			Part:           part,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertContentPartDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	// 从 content.raw.part 中恢复 part 字段
	var part responsesTypes.OutputContentPart
	if contract.Content != nil && contract.Content.Raw != nil {
		if partData, ok := contract.Content.Raw["part"].(map[string]interface{}); ok {
			var err error
			part, err = unmarshalContentPart(partData)
			if err != nil {
				return nil, errors.Wrap(errors.ErrCodeStreamError, "反序列化 part 字段失败", err)
			}
		}
	}

	return &responsesTypes.StreamEvent{
		ContentPartDone: &responsesTypes.ResponseContentPartDoneEvent{
			Type:           responsesTypes.StreamEventContentPartDone,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			Part:           part,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// Output Text 事件转换函数

func convertOutputTextDeltaFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	if contract.Delta != nil && contract.Delta.Text != nil {
		delta = *contract.Delta.Text
	}

	// 从 extensions 提取 logprobs 和 obfuscation
	// 注意：logprobs 必须始终输出为数组（即使是空数组），不能为 null
	var logprobs []responsesTypes.ResponseLogProb
	var obfuscation *string
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			// 始终提取 logprobs，即使是空数组
			if lp, ok := openaiExt["logprobs"]; ok {
				// 尝试类型断言
				if lpArray, ok := lp.([]responsesTypes.ResponseLogProb); ok {
					logprobs = lpArray
				} else if lpArray, ok := lp.([]interface{}); ok {
					// 处理从 JSON 反序列化后的 []interface{} 类型
					logprobs = make([]responsesTypes.ResponseLogProb, 0, len(lpArray))
					for _, item := range lpArray {
						if itemMap, ok := item.(map[string]interface{}); ok {
							jsonData, _ := json.Marshal(itemMap)
							var logProb responsesTypes.ResponseLogProb
							if err := json.Unmarshal(jsonData, &logProb); err == nil {
								logprobs = append(logprobs, logProb)
							}
						}
					}
				}
			}
			if obf, ok := openaiExt["obfuscation"].(string); ok {
				obfuscation = &obf
			}
		}
	}
	// 确保 logprobs 不为 nil，始终输出空数组
	if logprobs == nil {
		logprobs = []responsesTypes.ResponseLogProb{}
	}

	return &responsesTypes.StreamEvent{
		OutputTextDelta: &responsesTypes.ResponseOutputTextDeltaEvent{
			Type:           responsesTypes.StreamEventOutputTextDelta,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
			Logprobs:       logprobs,
			Obfuscation:    obfuscation,
		},
	}, nil
}

func convertOutputTextDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	text := ""
	if contract.Content != nil && contract.Content.Text != nil {
		text = *contract.Content.Text
	}

	// 从 extensions 提取 logprobs
	// 注意：logprobs 必须始终输出为数组（即使是空数组），不能为 null
	var logprobs []responsesTypes.ResponseLogProb
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			// 始终提取 logprobs，即使是空数组
			if lp, ok := openaiExt["logprobs"]; ok {
				// 尝试类型断言
				if lpArray, ok := lp.([]responsesTypes.ResponseLogProb); ok {
					logprobs = lpArray
				} else if lpArray, ok := lp.([]interface{}); ok {
					// 处理从 JSON 反序列化后的 []interface{} 类型
					logprobs = make([]responsesTypes.ResponseLogProb, 0, len(lpArray))
					for _, item := range lpArray {
						if itemMap, ok := item.(map[string]interface{}); ok {
							jsonData, _ := json.Marshal(itemMap)
							var logProb responsesTypes.ResponseLogProb
							if err := json.Unmarshal(jsonData, &logProb); err == nil {
								logprobs = append(logprobs, logProb)
							}
						}
					}
				}
			}
		}
	}
	// 确保 logprobs 不为 nil，始终输出空数组
	if logprobs == nil {
		logprobs = []responsesTypes.ResponseLogProb{}
	}

	return &responsesTypes.StreamEvent{
		OutputTextDone: &responsesTypes.ResponseOutputTextDoneEvent{
			Type:           responsesTypes.StreamEventOutputTextDone,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			Text:           text,
			SequenceNumber: contract.SequenceNumber,
			Logprobs:       logprobs,
		},
	}, nil
}

func convertOutputTextAnnotationAddedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	var annotation responsesTypes.Annotation
	if contract.Content != nil && len(contract.Content.Annotations) > 0 {
		if ann, ok := contract.Content.Annotations[0].(responsesTypes.Annotation); ok {
			annotation = ann
		}
	}

	return &responsesTypes.StreamEvent{
		OutputTextAnnotationAdded: &responsesTypes.ResponseOutputTextAnnotationAddedEvent{
			Type:            responsesTypes.StreamEventOutputTextAnnotationAdded,
			ItemID:          contract.ItemID,
			OutputIndex:     contract.OutputIndex,
			ContentIndex:    contract.ContentIndex,
			AnnotationIndex: contract.AnnotationIndex,
			Annotation:      annotation,
			SequenceNumber:  contract.SequenceNumber,
		},
	}, nil
}

// Refusal 事件转换函数

func convertRefusalDeltaFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	if contract.Delta != nil && contract.Delta.Text != nil {
		delta = *contract.Delta.Text
	}

	// 从 extensions 提取 obfuscation
	var obfuscation *string
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if obf, ok := openaiExt["obfuscation"].(string); ok {
				obfuscation = &obf
			}
		}
	}

	return &responsesTypes.StreamEvent{
		RefusalDelta: &responsesTypes.ResponseRefusalDeltaEvent{
			Type:           responsesTypes.StreamEventRefusalDelta,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
			Obfuscation:    obfuscation,
		},
	}, nil
}

func convertRefusalDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	refusal := ""
	if contract.Content != nil && contract.Content.Text != nil {
		refusal = *contract.Content.Text
	}

	return &responsesTypes.StreamEvent{
		RefusalDone: &responsesTypes.ResponseRefusalDoneEvent{
			Type:           responsesTypes.StreamEventRefusalDone,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			Refusal:        refusal,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// Reasoning Text 事件转换函数

func convertReasoningTextDeltaFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	if contract.Delta != nil && contract.Delta.Text != nil {
		delta = *contract.Delta.Text
	}

	// 从 extensions 提取 obfuscation
	var obfuscation *string
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if obf, ok := openaiExt["obfuscation"].(string); ok {
				obfuscation = &obf
			}
		}
	}

	return &responsesTypes.StreamEvent{
		ReasoningTextDelta: &responsesTypes.ResponseReasoningTextDeltaEvent{
			Type:           responsesTypes.StreamEventReasoningTextDelta,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
			Obfuscation:    obfuscation,
		},
	}, nil
}

func convertReasoningTextDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	text := ""
	if contract.Content != nil && contract.Content.Text != nil {
		text = *contract.Content.Text
	}

	return &responsesTypes.StreamEvent{
		ReasoningTextDone: &responsesTypes.ResponseReasoningTextDoneEvent{
			Type:           responsesTypes.StreamEventReasoningTextDone,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			Text:           text,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// Reasoning Summary 事件转换函数

func convertReasoningSummaryPartAddedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	var part responsesTypes.OutputSummaryPart
	if contract.Content != nil && contract.Content.Raw != nil {
		if p, ok := contract.Content.Raw["part"].(responsesTypes.OutputSummaryPart); ok {
			part = p
		}
	}

	summaryIndex := 0
	if contract.Content != nil && contract.Content.Raw != nil {
		if si, ok := contract.Content.Raw["summary_index"].(float64); ok {
			summaryIndex = int(si)
		}
	}

	return &responsesTypes.StreamEvent{
		ReasoningSummaryPartAdded: &responsesTypes.ResponseReasoningSummaryPartAddedEvent{
			Type:           responsesTypes.StreamEventReasoningSummaryPartAdded,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			SummaryIndex:   summaryIndex,
			Part:           part,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertReasoningSummaryPartDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	var part responsesTypes.OutputSummaryPart
	if contract.Content != nil && contract.Content.Raw != nil {
		if p, ok := contract.Content.Raw["part"].(responsesTypes.OutputSummaryPart); ok {
			part = p
		}
	}

	summaryIndex := 0
	if contract.Content != nil && contract.Content.Raw != nil {
		if si, ok := contract.Content.Raw["summary_index"].(float64); ok {
			summaryIndex = int(si)
		}
	}

	return &responsesTypes.StreamEvent{
		ReasoningSummaryPartDone: &responsesTypes.ResponseReasoningSummaryPartDoneEvent{
			Type:           responsesTypes.StreamEventReasoningSummaryPartDone,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			SummaryIndex:   summaryIndex,
			Part:           part,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertReasoningSummaryTextDeltaFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	if contract.Delta != nil && contract.Delta.Text != nil {
		delta = *contract.Delta.Text
	}

	// 从 extensions 提取 obfuscation
	var obfuscation *string
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if obf, ok := openaiExt["obfuscation"].(string); ok {
				obfuscation = &obf
			}
		}
	}

	summaryIndex := 0
	if contract.Content != nil && contract.Content.Raw != nil {
		if si, ok := contract.Content.Raw["summary_index"].(float64); ok {
			summaryIndex = int(si)
		}
	}

	return &responsesTypes.StreamEvent{
		ReasoningSummaryTextDelta: &responsesTypes.ResponseReasoningSummaryTextDeltaEvent{
			Type:           responsesTypes.StreamEventReasoningSummaryTextDelta,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			SummaryIndex:   summaryIndex,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
			Obfuscation:    obfuscation,
		},
	}, nil
}

func convertReasoningSummaryTextDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	text := ""
	if contract.Content != nil && contract.Content.Text != nil {
		text = *contract.Content.Text
	}

	summaryIndex := 0
	if contract.Content != nil && contract.Content.Raw != nil {
		if si, ok := contract.Content.Raw["summary_index"].(float64); ok {
			summaryIndex = int(si)
		}
	}

	return &responsesTypes.StreamEvent{
		ReasoningSummaryTextDone: &responsesTypes.ResponseReasoningSummaryTextDoneEvent{
			Type:           responsesTypes.StreamEventReasoningSummaryTextDone,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			SummaryIndex:   summaryIndex,
			Text:           text,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// Function Call 事件转换函数

func convertFunctionCallArgumentsDeltaFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	if contract.Delta != nil && contract.Delta.PartialJSON != nil {
		delta = *contract.Delta.PartialJSON
	}

	// 从 extensions 提取 obfuscation
	var obfuscation *string
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if obf, ok := openaiExt["obfuscation"].(string); ok {
				obfuscation = &obf
			}
		}
	}

	return &responsesTypes.StreamEvent{
		FunctionCallArgumentsDelta: &responsesTypes.ResponseFunctionCallArgumentsDeltaEvent{
			Type:           responsesTypes.StreamEventFunctionCallArgumentsDelta,
			ItemID:         contract.ItemID,
			OutputIndex:    contract.OutputIndex,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
			Obfuscation:    obfuscation,
		},
	}, nil
}

func convertFunctionCallArgumentsDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	name := ""
	arguments := ""
	if contract.Content != nil && contract.Content.Tool != nil {
		name = contract.Content.Tool.Name
		arguments = contract.Content.Tool.Arguments
	}

	return &responsesTypes.StreamEvent{
		FunctionCallArgumentsDone: &responsesTypes.ResponseFunctionCallArgumentsDoneEvent{
			Type:           responsesTypes.StreamEventFunctionCallArgumentsDone,
			ItemID:         contract.ItemID,
			Name:           name,
			OutputIndex:    contract.OutputIndex,
			Arguments:      arguments,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// Custom Tool Call 事件转换函数

func convertCustomToolCallInputDeltaFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	if contract.Delta != nil && contract.Delta.PartialJSON != nil {
		delta = *contract.Delta.PartialJSON
	}

	// 从 extensions 提取 obfuscation
	var obfuscation *string
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if obf, ok := openaiExt["obfuscation"].(string); ok {
				obfuscation = &obf
			}
		}
	}

	return &responsesTypes.StreamEvent{
		CustomToolCallInputDelta: &responsesTypes.ResponseCustomToolCallInputDeltaEvent{
			Type:           responsesTypes.StreamEventCustomToolCallInputDelta,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
			Obfuscation:    obfuscation,
		},
	}, nil
}

func convertCustomToolCallInputDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	input := ""
	if contract.Content != nil && contract.Content.Tool != nil {
		input = contract.Content.Tool.Arguments
	}

	return &responsesTypes.StreamEvent{
		CustomToolCallInputDone: &responsesTypes.ResponseCustomToolCallInputDoneEvent{
			Type:           responsesTypes.StreamEventCustomToolCallInputDone,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			Input:          input,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// MCP Call 事件转换函数

func convertMCPCallArgumentsDeltaFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	if contract.Delta != nil && contract.Delta.PartialJSON != nil {
		delta = *contract.Delta.PartialJSON
	}

	// 从 extensions 提取 obfuscation
	var obfuscation *string
	if contract.Extensions != nil {
		if openaiExt, ok := contract.Extensions["openai_responses"].(map[string]interface{}); ok {
			if obf, ok := openaiExt["obfuscation"].(string); ok {
				obfuscation = &obf
			}
		}
	}

	return &responsesTypes.StreamEvent{
		MCPCallArgumentsDelta: &responsesTypes.ResponseMCPCallArgumentsDeltaEvent{
			Type:           responsesTypes.StreamEventMCPCallArgumentsDelta,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
			Obfuscation:    obfuscation,
		},
	}, nil
}

func convertMCPCallArgumentsDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	arguments := ""
	if contract.Content != nil && contract.Content.Tool != nil {
		arguments = contract.Content.Tool.Arguments
	}

	return &responsesTypes.StreamEvent{
		MCPCallArgumentsDone: &responsesTypes.ResponseMCPCallArgumentsDoneEvent{
			Type:           responsesTypes.StreamEventMCPCallArgumentsDone,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			Arguments:      arguments,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertMCPCallCompletedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		MCPCallCompleted: &responsesTypes.ResponseMCPCallCompletedEvent{
			Type:           responsesTypes.StreamEventMCPCallCompleted,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertMCPCallFailedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		MCPCallFailed: &responsesTypes.ResponseMCPCallFailedEvent{
			Type:           responsesTypes.StreamEventMCPCallFailed,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertMCPCallInProgressFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		MCPCallInProgress: &responsesTypes.ResponseMCPCallInProgressEvent{
			Type:           responsesTypes.StreamEventMCPCallInProgress,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertMCPListToolsCompletedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		MCPListToolsCompleted: &responsesTypes.ResponseMCPListToolsCompletedEvent{
			Type:           responsesTypes.StreamEventMCPListToolsCompleted,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertMCPListToolsFailedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		MCPListToolsFailed: &responsesTypes.ResponseMCPListToolsFailedEvent{
			Type:           responsesTypes.StreamEventMCPListToolsFailed,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertMCPListToolsInProgressFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		MCPListToolsInProgress: &responsesTypes.ResponseMCPListToolsInProgressEvent{
			Type:           responsesTypes.StreamEventMCPListToolsInProgress,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// Audio 事件转换函数

func convertAudioDeltaFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	if contract.Delta != nil && contract.Delta.Text != nil {
		delta = *contract.Delta.Text
	}

	return &responsesTypes.StreamEvent{
		AudioDelta: &responsesTypes.ResponseAudioDeltaEvent{
			Type:           responsesTypes.StreamEventAudioDelta,
			ResponseID:     contract.ResponseID,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertAudioDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		AudioDone: &responsesTypes.ResponseAudioDoneEvent{
			Type:           responsesTypes.StreamEventAudioDone,
			ResponseID:     contract.ResponseID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertAudioTranscriptDeltaFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	if contract.Delta != nil && contract.Delta.Text != nil {
		delta = *contract.Delta.Text
	}

	return &responsesTypes.StreamEvent{
		AudioTranscriptDelta: &responsesTypes.ResponseAudioTranscriptDeltaEvent{
			Type:           responsesTypes.StreamEventAudioTranscriptDelta,
			ResponseID:     contract.ResponseID,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertAudioTranscriptDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		AudioTranscriptDone: &responsesTypes.ResponseAudioTranscriptDoneEvent{
			Type:           responsesTypes.StreamEventAudioTranscriptDone,
			ResponseID:     contract.ResponseID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// Code Interpreter 事件转换函数

func convertCodeInterpreterCallCodeDeltaFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	if contract.Delta != nil && contract.Delta.Text != nil {
		delta = *contract.Delta.Text
	}

	return &responsesTypes.StreamEvent{
		CodeInterpreterCallCodeDelta: &responsesTypes.ResponseCodeInterpreterCallCodeDeltaEvent{
			Type:           responsesTypes.StreamEventCodeInterpreterCallCodeDelta,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertCodeInterpreterCallCodeDoneFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	code := ""
	if contract.Content != nil && contract.Content.Text != nil {
		code = *contract.Content.Text
	}

	return &responsesTypes.StreamEvent{
		CodeInterpreterCallCodeDone: &responsesTypes.ResponseCodeInterpreterCallCodeDoneEvent{
			Type:           responsesTypes.StreamEventCodeInterpreterCallCodeDone,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			Code:           code,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertCodeInterpreterCallCompletedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		CodeInterpreterCallCompleted: &responsesTypes.ResponseCodeInterpreterCallCompletedEvent{
			Type:           responsesTypes.StreamEventCodeInterpreterCallCompleted,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertCodeInterpreterCallInProgressFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		CodeInterpreterCallInProgress: &responsesTypes.ResponseCodeInterpreterCallInProgressEvent{
			Type:           responsesTypes.StreamEventCodeInterpreterCallInProgress,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertCodeInterpreterCallInterpretingFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		CodeInterpreterCallInterpreting: &responsesTypes.ResponseCodeInterpreterCallInterpretingEvent{
			Type:           responsesTypes.StreamEventCodeInterpreterCallInterpreting,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// File Search 事件转换函数

func convertFileSearchCallCompletedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		FileSearchCallCompleted: &responsesTypes.ResponseFileSearchCallCompletedEvent{
			Type:           responsesTypes.StreamEventFileSearchCallCompleted,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertFileSearchCallInProgressFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		FileSearchCallInProgress: &responsesTypes.ResponseFileSearchCallInProgressEvent{
			Type:           responsesTypes.StreamEventFileSearchCallInProgress,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertFileSearchCallSearchingFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		FileSearchCallSearching: &responsesTypes.ResponseFileSearchCallSearchingEvent{
			Type:           responsesTypes.StreamEventFileSearchCallSearching,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// Web Search 事件转换函数

func convertWebSearchCallCompletedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		WebSearchCallCompleted: &responsesTypes.ResponseWebSearchCallCompletedEvent{
			Type:           responsesTypes.StreamEventWebSearchCallCompleted,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertWebSearchCallInProgressFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		WebSearchCallInProgress: &responsesTypes.ResponseWebSearchCallInProgressEvent{
			Type:           responsesTypes.StreamEventWebSearchCallInProgress,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertWebSearchCallSearchingFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		WebSearchCallSearching: &responsesTypes.ResponseWebSearchCallSearchingEvent{
			Type:           responsesTypes.StreamEventWebSearchCallSearching,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

// Image Generation 事件转换函数

func convertImageGenCallCompletedFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		ImageGenCallCompleted: &responsesTypes.ResponseImageGenCallCompletedEvent{
			Type:           responsesTypes.StreamEventImageGenCallCompleted,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertImageGenCallInProgressFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		ImageGenCallInProgress: &responsesTypes.ResponseImageGenCallInProgressEvent{
			Type:           responsesTypes.StreamEventImageGenCallInProgress,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertImageGenCallGeneratingFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	return &responsesTypes.StreamEvent{
		ImageGenCallGenerating: &responsesTypes.ResponseImageGenCallGeneratingEvent{
			Type:           responsesTypes.StreamEventImageGenCallGenerating,
			OutputIndex:    contract.OutputIndex,
			ItemID:         contract.ItemID,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func convertImageGenCallPartialImageFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	partialImageIndex := 0
	partialImageB64 := ""
	if contract.Content != nil && contract.Content.Raw != nil {
		if idx, ok := contract.Content.Raw["partial_image_index"].(float64); ok {
			partialImageIndex = int(idx)
		}
		if b64, ok := contract.Content.Raw["partial_image_b64"].(string); ok {
			partialImageB64 = b64
		}
	}

	return &responsesTypes.StreamEvent{
		ImageGenCallPartialImage: &responsesTypes.ResponseImageGenCallPartialImageEvent{
			Type:              responsesTypes.StreamEventImageGenCallPartialImage,
			OutputIndex:       contract.OutputIndex,
			ItemID:            contract.ItemID,
			SequenceNumber:    contract.SequenceNumber,
			PartialImageIndex: partialImageIndex,
			PartialImageB64:   partialImageB64,
		},
	}, nil
}

// Error 事件转换函数

func convertStreamErrorFromContract(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil || contract.Error == nil {
		return nil, nil
	}

	event := &responsesTypes.ResponseErrorEvent{
		Type:           responsesTypes.StreamEventError,
		Message:        contract.Error.Message,
		SequenceNumber: contract.SequenceNumber,
	}

	if contract.Error.Code != "" {
		event.Code = &contract.Error.Code
	}
	if contract.Error.Param != "" {
		event.Param = &contract.Error.Param
	}

	return &responsesTypes.StreamEvent{
		Error: event,
	}, nil
}

// 降级映射函数

func degradeMessageStartToResponseCreated(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	responseID := contract.ResponseID
	if responseID == "" {
		responseID = contract.MessageID
	}

	return &responsesTypes.StreamEvent{
		Created: &responsesTypes.ResponseCreatedEvent{
			Type:           responsesTypes.StreamEventCreated,
			Response:       responsesTypes.Response{ID: responseID},
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func degradeMessageDeltaToOutputTextDelta(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	// 从 message.content_text 提取
	if contract.Message != nil && contract.Message.ContentText != nil {
		delta = *contract.Message.ContentText
	}
	// 从 delta.text 提取
	if delta == "" && contract.Delta != nil && contract.Delta.Text != nil {
		delta = *contract.Delta.Text
	}

	itemID := contract.ItemID
	if itemID == "" {
		itemID = contract.MessageID
	}

	return &responsesTypes.StreamEvent{
		OutputTextDelta: &responsesTypes.ResponseOutputTextDeltaEvent{
			Type:           responsesTypes.StreamEventOutputTextDelta,
			ItemID:         itemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
			Logprobs:       []responsesTypes.ResponseLogProb{},
		},
	}, nil
}

func degradeMessageStopToOutputTextDone(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	text := ""
	// 从 message.content_text 提取
	if contract.Message != nil && contract.Message.ContentText != nil {
		text = *contract.Message.ContentText
	}

	itemID := contract.ItemID
	if itemID == "" {
		itemID = contract.MessageID
	}

	return &responsesTypes.StreamEvent{
		OutputTextDone: &responsesTypes.ResponseOutputTextDoneEvent{
			Type:           responsesTypes.StreamEventOutputTextDone,
			ItemID:         itemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			Text:           text,
			SequenceNumber: contract.SequenceNumber,
			Logprobs:       []responsesTypes.ResponseLogProb{},
		},
	}, nil
}

func degradeMessageStopToResponseCompleted(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	responseID := contract.ResponseID
	if responseID == "" {
		responseID = contract.MessageID
	}

	return &responsesTypes.StreamEvent{
		Completed: &responsesTypes.ResponseCompletedEvent{
			Type:           responsesTypes.StreamEventCompleted,
			Response:       responsesTypes.Response{ID: responseID},
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func degradeContentBlockStartToOutputItemAdded(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	itemID := contract.ItemID
	if itemID == "" {
		itemID = contract.MessageID
	}

	item := responsesTypes.OutputItem{
		Message: &responsesTypes.OutputMessage{
			ID:   itemID,
			Type: responsesTypes.OutputItemTypeMessage,
		},
	}

	return &responsesTypes.StreamEvent{
		OutputItemAdded: &responsesTypes.ResponseOutputItemAddedEvent{
			Type:           responsesTypes.StreamEventOutputItemAdded,
			OutputIndex:    contract.OutputIndex,
			Item:           item,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func degradeContentBlockStartToContentPartAdded(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	itemID := contract.ItemID
	if itemID == "" {
		itemID = contract.MessageID
	}

	return &responsesTypes.StreamEvent{
		ContentPartAdded: &responsesTypes.ResponseContentPartAddedEvent{
			Type:           responsesTypes.StreamEventContentPartAdded,
			ItemID:         itemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func degradeContentBlockDeltaToOutputTextDelta(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	delta := ""
	// 从 delta.text 提取
	if contract.Delta != nil && contract.Delta.Text != nil {
		delta = *contract.Delta.Text
	}
	// 从 content.text 提取
	if delta == "" && contract.Content != nil && contract.Content.Text != nil {
		delta = *contract.Content.Text
	}

	itemID := contract.ItemID
	if itemID == "" {
		itemID = contract.MessageID
	}

	return &responsesTypes.StreamEvent{
		OutputTextDelta: &responsesTypes.ResponseOutputTextDeltaEvent{
			Type:           responsesTypes.StreamEventOutputTextDelta,
			ItemID:         itemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			Delta:          delta,
			SequenceNumber: contract.SequenceNumber,
			Logprobs:       []responsesTypes.ResponseLogProb{},
		},
	}, nil
}

func degradeContentBlockStopToContentPartDone(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	itemID := contract.ItemID
	if itemID == "" {
		itemID = contract.MessageID
	}

	return &responsesTypes.StreamEvent{
		ContentPartDone: &responsesTypes.ResponseContentPartDoneEvent{
			Type:           responsesTypes.StreamEventContentPartDone,
			ItemID:         itemID,
			OutputIndex:    contract.OutputIndex,
			ContentIndex:   contract.ContentIndex,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}

func degradeContentBlockStopToOutputItemDone(contract *adapterTypes.StreamEventContract) (*responsesTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	itemID := contract.ItemID
	if itemID == "" {
		itemID = contract.MessageID
	}

	item := responsesTypes.OutputItem{
		Message: &responsesTypes.OutputMessage{
			ID:   itemID,
			Type: responsesTypes.OutputItemTypeMessage,
		},
	}

	return &responsesTypes.StreamEvent{
		OutputItemDone: &responsesTypes.ResponseOutputItemDoneEvent{
			Type:           responsesTypes.StreamEventOutputItemDone,
			OutputIndex:    contract.OutputIndex,
			Item:           item,
			SequenceNumber: contract.SequenceNumber,
		},
	}, nil
}
