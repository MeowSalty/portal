package responses

import (
	"encoding/json"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/request/adapter/openai/converter/responses/helper"
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
)

// marshalContentPart 将 OutputContentPart 序列化为 map[string]interface{}。
// 用于在 contract 中完整保真地存储 part 字段。
func marshalContentPart(part responsesTypes.OutputContentPart) (map[string]interface{}, error) {
	data, err := json.Marshal(part)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// getItemID 从 OutputItem 中获取 ID。
// 由于 OutputItem 移除了顶层 ID 字段，需要从具体结构体中获取。
func getItemID(item *responsesTypes.OutputItem) string {
	if item == nil {
		return ""
	}
	if item.Message != nil {
		return item.Message.ID
	}
	if item.FunctionCall != nil {
		return item.FunctionCall.ID
	}
	if item.FileSearchCall != nil {
		return item.FileSearchCall.ID
	}
	if item.WebSearchCall != nil {
		return item.WebSearchCall.ID
	}
	if item.ComputerCall != nil {
		return item.ComputerCall.ID
	}
	if item.Reasoning != nil {
		return item.Reasoning.ID
	}
	if item.CodeInterpreterCall != nil {
		return item.CodeInterpreterCall.ID
	}
	if item.ImageGenCall != nil {
		return item.ImageGenCall.ID
	}
	if item.LocalShellCall != nil {
		return item.LocalShellCall.ID
	}
	if item.ShellCall != nil {
		return item.ShellCall.ID
	}
	if item.ShellCallOutput != nil {
		return item.ShellCallOutput.ID
	}
	if item.ApplyPatchCall != nil {
		return item.ApplyPatchCall.ID
	}
	if item.ApplyPatchCallOutput != nil {
		return item.ApplyPatchCallOutput.ID
	}
	if item.MCPCall != nil {
		return item.MCPCall.ID
	}
	if item.MCPListTools != nil {
		return item.MCPListTools.ID
	}
	if item.MCPApprovalRequest != nil {
		return item.MCPApprovalRequest.ID
	}
	if item.CustomToolCall != nil {
		return item.CustomToolCall.ID
	}
	if item.Compaction != nil {
		return item.Compaction.ID
	}
	return ""
}

// StreamEventToContract 将 OpenAI Responses 流式事件转换为统一的 StreamEventContract。
//
// 根据 plans/stream-event-contract-plan.md 5.4 章节实现：
//   - response.* 事件：映射到对应的 response_* 事件类型
//   - output_item.* 事件：映射到 output_item_* 事件类型
//   - content_part.* 事件：映射到 content_part_* 事件类型
//   - output_text.* 事件：映射到 output_text_* 事件类型
//   - refusal.* 事件：映射到 refusal_* 事件类型
//   - reasoning_* 事件：映射到对应的 reasoning_* 事件类型
//   - function_call_arguments.* 事件：映射到 function_call_arguments_* 事件类型
//   - custom_tool_call_input.* 事件：映射到 custom_tool_call_input_* 事件类型
//   - mcp_call_arguments.* 事件：映射到 mcp_call_arguments_* 事件类型
//   - mcp_call.* 事件：映射到对应的 mcp_call_* 事件类型
//   - mcp_list_tools.* 事件：映射到对应的 mcp_list_tools_* 事件类型
//   - audio.* 事件：映射到对应的 audio_* 事件类型
//   - code_interpreter_call.* 事件：映射到对应的 code_interpreter_call_* 事件类型
//   - file_search_call.* 事件：映射到对应的 file_search_call_* 事件类型
//   - web_search_call.* 事件：映射到对应的 web_search_call_* 事件类型
//   - image_generation_call.* 事件：映射到对应的 image_generation_call_* 事件类型
//   - error 事件：映射到 error 事件类型
//
// 参数：
//   - event: OpenAI Responses 流式事件
//   - log: 日志记录器（可选，传 nil 时使用 NopLogger）
//
// 返回：
//   - *adapterTypes.StreamEventContract: 转换后的统一流式事件
//   - error: 转换过程中的错误
func StreamEventToContract(event *responsesTypes.StreamEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	if event == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	// 使用 WithGroup 创建子日志记录器
	log = log.WithGroup("stream_converter")

	// Response 生命周期事件
	if event.Created != nil {
		return convertResponseCreatedEvent(event.Created, log)
	}
	if event.InProgress != nil {
		return convertResponseInProgressEvent(event.InProgress, log)
	}
	if event.Completed != nil {
		return convertResponseCompletedEvent(event.Completed, log)
	}
	if event.Failed != nil {
		return convertResponseFailedEvent(event.Failed, log)
	}
	if event.Incomplete != nil {
		return convertResponseIncompleteEvent(event.Incomplete, log)
	}
	if event.Queued != nil {
		return convertResponseQueuedEvent(event.Queued, log)
	}

	// Output Item 事件
	if event.OutputItemAdded != nil {
		return convertOutputItemAddedEvent(event.OutputItemAdded, log)
	}
	if event.OutputItemDone != nil {
		return convertOutputItemDoneEvent(event.OutputItemDone, log)
	}

	// Content Part 事件
	if event.ContentPartAdded != nil {
		return convertContentPartAddedEvent(event.ContentPartAdded, log)
	}
	if event.ContentPartDone != nil {
		return convertContentPartDoneEvent(event.ContentPartDone, log)
	}

	// Output Text 事件
	if event.OutputTextDelta != nil {
		return convertOutputTextDeltaEvent(event.OutputTextDelta, log)
	}
	if event.OutputTextDone != nil {
		return convertOutputTextDoneEvent(event.OutputTextDone, log)
	}
	if event.OutputTextAnnotationAdded != nil {
		return convertOutputTextAnnotationAddedEvent(event.OutputTextAnnotationAdded, log)
	}

	// Refusal 事件
	if event.RefusalDelta != nil {
		return convertRefusalDeltaEvent(event.RefusalDelta, log)
	}
	if event.RefusalDone != nil {
		return convertRefusalDoneEvent(event.RefusalDone, log)
	}

	// Reasoning Text 事件
	if event.ReasoningTextDelta != nil {
		return convertReasoningTextDeltaEvent(event.ReasoningTextDelta, log)
	}
	if event.ReasoningTextDone != nil {
		return convertReasoningTextDoneEvent(event.ReasoningTextDone, log)
	}

	// Reasoning Summary 事件
	if event.ReasoningSummaryPartAdded != nil {
		return convertReasoningSummaryPartAddedEvent(event.ReasoningSummaryPartAdded, log)
	}
	if event.ReasoningSummaryPartDone != nil {
		return convertReasoningSummaryPartDoneEvent(event.ReasoningSummaryPartDone, log)
	}
	if event.ReasoningSummaryTextDelta != nil {
		return convertReasoningSummaryTextDeltaEvent(event.ReasoningSummaryTextDelta, log)
	}
	if event.ReasoningSummaryTextDone != nil {
		return convertReasoningSummaryTextDoneEvent(event.ReasoningSummaryTextDone, log)
	}

	// Function Call 事件
	if event.FunctionCallArgumentsDelta != nil {
		return convertFunctionCallArgumentsDeltaEvent(event.FunctionCallArgumentsDelta, log)
	}
	if event.FunctionCallArgumentsDone != nil {
		return convertFunctionCallArgumentsDoneEvent(event.FunctionCallArgumentsDone, log)
	}

	// Custom Tool Call 事件
	if event.CustomToolCallInputDelta != nil {
		return convertCustomToolCallInputDeltaEvent(event.CustomToolCallInputDelta, log)
	}
	if event.CustomToolCallInputDone != nil {
		return convertCustomToolCallInputDoneEvent(event.CustomToolCallInputDone, log)
	}

	// MCP Call 事件
	if event.MCPCallArgumentsDelta != nil {
		return convertMCPCallArgumentsDeltaEvent(event.MCPCallArgumentsDelta, log)
	}
	if event.MCPCallArgumentsDone != nil {
		return convertMCPCallArgumentsDoneEvent(event.MCPCallArgumentsDone, log)
	}
	if event.MCPCallCompleted != nil {
		return convertMCPCallCompletedEvent(event.MCPCallCompleted, log)
	}
	if event.MCPCallFailed != nil {
		return convertMCPCallFailedEvent(event.MCPCallFailed, log)
	}
	if event.MCPCallInProgress != nil {
		return convertMCPCallInProgressEvent(event.MCPCallInProgress, log)
	}

	// MCP List Tools 事件
	if event.MCPListToolsCompleted != nil {
		return convertMCPListToolsCompletedEvent(event.MCPListToolsCompleted, log)
	}
	if event.MCPListToolsFailed != nil {
		return convertMCPListToolsFailedEvent(event.MCPListToolsFailed, log)
	}
	if event.MCPListToolsInProgress != nil {
		return convertMCPListToolsInProgressEvent(event.MCPListToolsInProgress, log)
	}

	// Audio 事件
	if event.AudioDelta != nil {
		return convertAudioDeltaEvent(event.AudioDelta, log)
	}
	if event.AudioDone != nil {
		return convertAudioDoneEvent(event.AudioDone, log)
	}
	if event.AudioTranscriptDelta != nil {
		return convertAudioTranscriptDeltaEvent(event.AudioTranscriptDelta, log)
	}
	if event.AudioTranscriptDone != nil {
		return convertAudioTranscriptDoneEvent(event.AudioTranscriptDone, log)
	}

	// Code Interpreter 事件
	if event.CodeInterpreterCallCodeDelta != nil {
		return convertCodeInterpreterCallCodeDeltaEvent(event.CodeInterpreterCallCodeDelta, log)
	}
	if event.CodeInterpreterCallCodeDone != nil {
		return convertCodeInterpreterCallCodeDoneEvent(event.CodeInterpreterCallCodeDone, log)
	}
	if event.CodeInterpreterCallCompleted != nil {
		return convertCodeInterpreterCallCompletedEvent(event.CodeInterpreterCallCompleted, log)
	}
	if event.CodeInterpreterCallInProgress != nil {
		return convertCodeInterpreterCallInProgressEvent(event.CodeInterpreterCallInProgress, log)
	}
	if event.CodeInterpreterCallInterpreting != nil {
		return convertCodeInterpreterCallInterpretingEvent(event.CodeInterpreterCallInterpreting, log)
	}

	// File Search 事件
	if event.FileSearchCallCompleted != nil {
		return convertFileSearchCallCompletedEvent(event.FileSearchCallCompleted, log)
	}
	if event.FileSearchCallInProgress != nil {
		return convertFileSearchCallInProgressEvent(event.FileSearchCallInProgress, log)
	}
	if event.FileSearchCallSearching != nil {
		return convertFileSearchCallSearchingEvent(event.FileSearchCallSearching, log)
	}

	// Web Search 事件
	if event.WebSearchCallCompleted != nil {
		return convertWebSearchCallCompletedEvent(event.WebSearchCallCompleted, log)
	}
	if event.WebSearchCallInProgress != nil {
		return convertWebSearchCallInProgressEvent(event.WebSearchCallInProgress, log)
	}
	if event.WebSearchCallSearching != nil {
		return convertWebSearchCallSearchingEvent(event.WebSearchCallSearching, log)
	}

	// Image Generation 事件
	if event.ImageGenCallCompleted != nil {
		return convertImageGenCallCompletedEvent(event.ImageGenCallCompleted, log)
	}
	if event.ImageGenCallGenerating != nil {
		return convertImageGenCallGeneratingEvent(event.ImageGenCallGenerating, log)
	}
	if event.ImageGenCallInProgress != nil {
		return convertImageGenCallInProgressEvent(event.ImageGenCallInProgress, log)
	}
	if event.ImageGenCallPartialImage != nil {
		return convertImageGenCallPartialImageEvent(event.ImageGenCallPartialImage, log)
	}

	// Error 事件
	if event.Error != nil {
		return convertErrorEvent(event.Error, log)
	}

	return nil, errors.New(errors.ErrCodeInvalidArgument, "未知的流式事件类型")
}

// convertResponseCreatedEvent 转换 response.created 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: response_created
//   - response_id: response.id
//   - sequence_number: sequence_number
//   - extensions.openai_responses.response: Response 原文
func convertResponseCreatedEvent(event *responsesTypes.ResponseCreatedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventResponseCreated,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     event.Response.ID,
		SequenceNumber: event.SequenceNumber,
		Extensions:     make(map[string]interface{}),
	}

	// 保存原始 Response 到 extensions
	openaiExt := make(map[string]interface{})
	openaiExt["response"] = event.Response
	contract.Extensions["openai_responses"] = openaiExt

	auditLog := log.WithGroup("audit")
	auditLog.Info("转换 response.created 事件完成", "response_id", contract.ResponseID, "sequence_number", contract.SequenceNumber)

	return contract, nil
}

// convertResponseInProgressEvent 转换 response.in_progress 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: response_in_progress
//   - response_id: response.id
//   - sequence_number: sequence_number
//   - extensions.openai_responses.response: Response 原文
func convertResponseInProgressEvent(event *responsesTypes.ResponseInProgressEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventResponseInProgress,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     event.Response.ID,
		SequenceNumber: event.SequenceNumber,
		Extensions:     make(map[string]interface{}),
	}

	// 保存原始 Response 到 extensions
	openaiExt := make(map[string]interface{})
	openaiExt["response"] = event.Response
	contract.Extensions["openai_responses"] = openaiExt

	log.Debug("转换 response.in_progress 事件完成", "response_id", contract.ResponseID, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertResponseCompletedEvent 转换 response.completed 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: response_completed
//   - response_id: response.id
//   - sequence_number: sequence_number
//   - usage: response.usage
//   - extensions.openai_responses.response: Response 原文
func convertResponseCompletedEvent(event *responsesTypes.ResponseCompletedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventResponseCompleted,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     event.Response.ID,
		SequenceNumber: event.SequenceNumber,
		Extensions:     make(map[string]interface{}),
	}

	// 转换 Usage
	if event.Response.Usage != nil {
		contract.Usage = helper.ConvertUsageToStreamUsage(event.Response.Usage)
	}

	// 保存原始 Response 到 extensions
	openaiExt := make(map[string]interface{})
	openaiExt["response"] = event.Response
	contract.Extensions["openai_responses"] = openaiExt

	auditLog := log.WithGroup("audit")
	auditLog.Info("转换 response.completed 事件完成", "response_id", contract.ResponseID, "sequence_number", contract.SequenceNumber)

	return contract, nil
}

// convertResponseFailedEvent 转换 response.failed 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: response_failed
//   - response_id: response.id
//   - sequence_number: sequence_number
//   - error: response.error
//   - extensions.openai_responses.response: Response 原文
func convertResponseFailedEvent(event *responsesTypes.ResponseFailedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventResponseFailed,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     event.Response.ID,
		SequenceNumber: event.SequenceNumber,
		Extensions:     make(map[string]interface{}),
	}

	// 转换 Error
	if event.Response.Error != nil {
		contract.Error = helper.ConvertResponseErrorToStreamError(event.Response.Error)
	}

	// 保存原始 Response 到 extensions
	openaiExt := make(map[string]interface{})
	openaiExt["response"] = event.Response
	contract.Extensions["openai_responses"] = openaiExt

	auditLog := log.WithGroup("audit")
	auditLog.Error("转换 response.failed 事件完成", "response_id", contract.ResponseID, "sequence_number", contract.SequenceNumber, "error", contract.Error)

	return contract, nil
}

// convertResponseIncompleteEvent 转换 response.incomplete 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: response_incomplete
//   - response_id: response.id
//   - sequence_number: sequence_number
//   - extensions.openai_responses.incomplete_details: incomplete_details
//   - extensions.openai_responses.response: Response 原文
func convertResponseIncompleteEvent(event *responsesTypes.ResponseIncompleteEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventResponseIncomplete,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     event.Response.ID,
		SequenceNumber: event.SequenceNumber,
		Extensions:     make(map[string]interface{}),
	}

	// 保存 incomplete_details 到 extensions
	openaiExt := make(map[string]interface{})
	if event.Response.IncompleteDetails != nil {
		openaiExt["incomplete_details"] = event.Response.IncompleteDetails
	}
	openaiExt["response"] = event.Response
	contract.Extensions["openai_responses"] = openaiExt

	auditLog := log.WithGroup("audit")
	auditLog.Warn("转换 response.incomplete 事件完成", "response_id", contract.ResponseID, "sequence_number", contract.SequenceNumber)

	return contract, nil
}

// convertResponseQueuedEvent 转换 response.queued 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: response_queued
//   - response_id: response.id
//   - sequence_number: sequence_number
//   - extensions.openai_responses.response: Response 原文
func convertResponseQueuedEvent(event *responsesTypes.ResponseQueuedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventResponseQueued,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     event.Response.ID,
		SequenceNumber: event.SequenceNumber,
		Extensions:     make(map[string]interface{}),
	}

	// 保存原始 Response 到 extensions
	openaiExt := make(map[string]interface{})
	openaiExt["response"] = event.Response
	contract.Extensions["openai_responses"] = openaiExt

	log.Debug("转换 response.queued 事件完成", "response_id", contract.ResponseID, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertOutputItemAddedEvent 转换 response.output_item.added 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: output_item_added
//   - output_index: output_index
//   - item_id: item.id
//   - sequence_number: sequence_number
//   - content.raw.item: item
func convertOutputItemAddedEvent(event *responsesTypes.ResponseOutputItemAddedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	// 从具体结构体中获取 ID
	itemID := getItemID(&event.Item)

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputItemAdded,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         itemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "other",
			Raw:  make(map[string]interface{}),
		},
	}

	// 保存 item 到 content.raw
	contract.Content.Raw["item"] = event.Item

	log.Debug("转换 output_item.added 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertOutputItemDoneEvent 转换 response.output_item.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: output_item_done
//   - output_index: output_index
//   - item_id: item.id
//   - sequence_number: sequence_number
//   - content.raw.item: item
func convertOutputItemDoneEvent(event *responsesTypes.ResponseOutputItemDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	// 从具体结构体中获取 ID
	itemID := getItemID(&event.Item)

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputItemDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         itemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "other",
			Raw:  make(map[string]interface{}),
		},
	}

	// 保存 item 到 content.raw
	contract.Content.Raw["item"] = event.Item

	log.Debug("转换 output_item.done 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertContentPartAddedEvent 转换 response.content_part.added 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: content_part_added
//   - output_index: output_index
//   - item_id: item_id
//   - content_index: content_index
//   - sequence_number: sequence_number
//   - content.raw.part: part 字段（完整保真）
func convertContentPartAddedEvent(event *responsesTypes.ResponseContentPartAddedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventContentPartAdded,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		ContentIndex:   event.ContentIndex,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Raw: make(map[string]interface{}),
		},
	}

	// 将 part 字段存储到 content.raw.part 中，确保完整保真
	partData, err := marshalContentPart(event.Part)
	if err != nil {
		log.Error("序列化 part 字段失败", "error", err)
		return nil, errors.Wrap(errors.ErrCodeStreamError, "序列化 part 字段失败", err)
	}
	contract.Content.Raw["part"] = partData

	log.Debug("转换 content_part.added 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "content_index", contract.ContentIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertContentPartDoneEvent 转换 response.content_part.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: content_part_done
//   - output_index: output_index
//   - item_id: item_id
//   - content_index: content_index
//   - sequence_number: sequence_number
//   - content.raw.part: part 字段（完整保真）
func convertContentPartDoneEvent(event *responsesTypes.ResponseContentPartDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventContentPartDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		ContentIndex:   event.ContentIndex,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Raw: make(map[string]interface{}),
		},
	}

	// 将 part 字段存储到 content.raw.part 中，确保完整保真
	partData, err := marshalContentPart(event.Part)
	if err != nil {
		log.Error("序列化 part 字段失败", "error", err)
		return nil, errors.Wrap(errors.ErrCodeStreamError, "序列化 part 字段失败", err)
	}
	contract.Content.Raw["part"] = partData

	log.Debug("转换 content_part.done 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "content_index", contract.ContentIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertOutputTextDeltaEvent 转换 response.output_text.delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: output_text_delta
//   - output_index: output_index
//   - item_id: item_id
//   - content_index: content_index
//   - sequence_number: sequence_number
//   - delta.delta_type: text_delta
//   - delta.text: delta
//   - content.kind: output_text
//   - extensions.openai_responses.logprobs: logprobs
//   - extensions.openai_responses.obfuscation: obfuscation
func convertOutputTextDeltaEvent(event *responsesTypes.ResponseOutputTextDeltaEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputTextDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		ContentIndex:   event.ContentIndex,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "output_text",
		},
		Delta: &adapterTypes.StreamDeltaPayload{
			DeltaType: "text_delta",
			Text:      &event.Delta,
		},
		Extensions: make(map[string]interface{}),
	}

	// 保存 logprobs 和 obfuscation 到 extensions
	// 注意：logprobs 必须始终输出为数组（即使是空数组），不能为 null
	openaiExt := make(map[string]interface{})
	// 始终保存 logprobs，即使是空数组
	openaiExt["logprobs"] = event.Logprobs
	if event.Obfuscation != nil {
		openaiExt["obfuscation"] = *event.Obfuscation
	}
	contract.Extensions["openai_responses"] = openaiExt

	log.Debug("转换 output_text.delta 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "content_index", contract.ContentIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertOutputTextDoneEvent 转换 response.output_text.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: output_text_done
//   - output_index: output_index
//   - item_id: item_id
//   - content_index: content_index
//   - sequence_number: sequence_number
//   - content.kind: output_text
//   - content.text: text
//   - extensions.openai_responses.logprobs: logprobs
func convertOutputTextDoneEvent(event *responsesTypes.ResponseOutputTextDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputTextDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		ContentIndex:   event.ContentIndex,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "output_text",
			Text: &event.Text,
		},
		Extensions: make(map[string]interface{}),
	}

	// 保存 logprobs 到 extensions
	// 注意：logprobs 必须始终输出为数组（即使是空数组），不能为 null
	openaiExt := make(map[string]interface{})
	// 始终保存 logprobs，即使是空数组
	openaiExt["logprobs"] = event.Logprobs
	contract.Extensions["openai_responses"] = openaiExt

	log.Debug("转换 output_text.done 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "content_index", contract.ContentIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertOutputTextAnnotationAddedEvent 转换 response.output_text.annotation.added 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: output_text_annotation_added
//   - output_index: output_index
//   - item_id: item_id
//   - content_index: content_index
//   - annotation_index: annotation_index
//   - sequence_number: sequence_number
//   - content.annotations: [annotation]
func convertOutputTextAnnotationAddedEvent(event *responsesTypes.ResponseOutputTextAnnotationAddedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:            adapterTypes.StreamEventOutputTextAnnotationAdded,
		Source:          adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:     event.OutputIndex,
		ItemID:          event.ItemID,
		ContentIndex:    event.ContentIndex,
		AnnotationIndex: event.AnnotationIndex,
		SequenceNumber:  event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind:        "output_text",
			Annotations: []interface{}{event.Annotation},
		},
	}

	log.Debug("转换 output_text.annotation.added 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "content_index", contract.ContentIndex, "annotation_index", contract.AnnotationIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertRefusalDeltaEvent 转换 response.refusal.delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: refusal_delta
//   - output_index: output_index
//   - item_id: item_id
//   - content_index: content_index
//   - sequence_number: sequence_number
//   - content.kind: refusal
//   - delta.text: delta
//   - extensions.openai_responses.obfuscation: obfuscation
func convertRefusalDeltaEvent(event *responsesTypes.ResponseRefusalDeltaEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventRefusalDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		ContentIndex:   event.ContentIndex,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "refusal",
		},
		Delta: &adapterTypes.StreamDeltaPayload{
			DeltaType: "text_delta",
			Text:      &event.Delta,
		},
		Extensions: make(map[string]interface{}),
	}

	// 保存 obfuscation 到 extensions
	if event.Obfuscation != nil {
		openaiExt := make(map[string]interface{})
		openaiExt["obfuscation"] = *event.Obfuscation
		contract.Extensions["openai_responses"] = openaiExt
	}

	log.Debug("转换 refusal.delta 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "content_index", contract.ContentIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertRefusalDoneEvent 转换 response.refusal.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: refusal_done
//   - output_index: output_index
//   - item_id: item_id
//   - content_index: content_index
//   - sequence_number: sequence_number
//   - content.kind: refusal
//   - content.text: refusal
func convertRefusalDoneEvent(event *responsesTypes.ResponseRefusalDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventRefusalDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		ContentIndex:   event.ContentIndex,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "refusal",
			Text: &event.Refusal,
		},
	}

	log.Debug("转换 refusal.done 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "content_index", contract.ContentIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertReasoningTextDeltaEvent 转换 response.reasoning_text.delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: reasoning_text_delta
//   - output_index: output_index
//   - item_id: item_id
//   - content_index: content_index
//   - sequence_number: sequence_number
//   - content.kind: reasoning_text
//   - delta.text: delta
//   - extensions.openai_responses.obfuscation: obfuscation
func convertReasoningTextDeltaEvent(event *responsesTypes.ResponseReasoningTextDeltaEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventReasoningTextDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		ContentIndex:   event.ContentIndex,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "reasoning_text",
		},
		Delta: &adapterTypes.StreamDeltaPayload{
			DeltaType: "text_delta",
			Text:      &event.Delta,
		},
		Extensions: make(map[string]interface{}),
	}

	// 保存 obfuscation 到 extensions
	if event.Obfuscation != nil {
		openaiExt := make(map[string]interface{})
		openaiExt["obfuscation"] = *event.Obfuscation
		contract.Extensions["openai_responses"] = openaiExt
	}

	log.Debug("转换 reasoning_text.delta 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "content_index", contract.ContentIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertReasoningTextDoneEvent 转换 response.reasoning_text.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: reasoning_text_done
//   - output_index: output_index
//   - item_id: item_id
//   - content_index: content_index
//   - sequence_number: sequence_number
//   - content.kind: reasoning_text
//   - content.text: text
func convertReasoningTextDoneEvent(event *responsesTypes.ResponseReasoningTextDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventReasoningTextDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		ContentIndex:   event.ContentIndex,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "reasoning_text",
			Text: &event.Text,
		},
	}

	log.Debug("转换 reasoning_text.done 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "content_index", contract.ContentIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertReasoningSummaryPartAddedEvent 转换 response.reasoning_summary_part.added 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: reasoning_summary_part_added
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: reasoning_summary_text
//   - content.raw.part: part
func convertReasoningSummaryPartAddedEvent(event *responsesTypes.ResponseReasoningSummaryPartAddedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventReasoningSummaryPartAdded,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "reasoning_summary_text",
			Raw:  make(map[string]interface{}),
		},
	}

	// 保存 part 到 content.raw
	contract.Content.Raw["part"] = event.Part

	log.Debug("转换 reasoning_summary_part.added 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "summary_index", event.SummaryIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertReasoningSummaryPartDoneEvent 转换 response.reasoning_summary_part.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: reasoning_summary_part_done
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: reasoning_summary_text
//   - content.raw.part: part
func convertReasoningSummaryPartDoneEvent(event *responsesTypes.ResponseReasoningSummaryPartDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventReasoningSummaryPartDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "reasoning_summary_text",
			Raw:  make(map[string]interface{}),
		},
	}

	// 保存 part 到 content.raw
	contract.Content.Raw["part"] = event.Part

	log.Debug("转换 reasoning_summary_part.done 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "summary_index", event.SummaryIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertReasoningSummaryTextDeltaEvent 转换 response.reasoning_summary_text.delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: reasoning_summary_text_delta
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: reasoning_summary_text
//   - delta.text: delta
//   - extensions.openai_responses.obfuscation: obfuscation
func convertReasoningSummaryTextDeltaEvent(event *responsesTypes.ResponseReasoningSummaryTextDeltaEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventReasoningSummaryTextDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "reasoning_summary_text",
		},
		Delta: &adapterTypes.StreamDeltaPayload{
			DeltaType: "text_delta",
			Text:      &event.Delta,
		},
		Extensions: make(map[string]interface{}),
	}

	// 保存 obfuscation 到 extensions
	if event.Obfuscation != nil {
		openaiExt := make(map[string]interface{})
		openaiExt["obfuscation"] = *event.Obfuscation
		contract.Extensions["openai_responses"] = openaiExt
	}

	log.Debug("转换 reasoning_summary_text.delta 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "summary_index", event.SummaryIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertReasoningSummaryTextDoneEvent 转换 response.reasoning_summary_text.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: reasoning_summary_text_done
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: reasoning_summary_text
//   - content.text: text
func convertReasoningSummaryTextDoneEvent(event *responsesTypes.ResponseReasoningSummaryTextDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventReasoningSummaryTextDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "reasoning_summary_text",
			Text: &event.Text,
		},
	}

	log.Debug("转换 reasoning_summary_text.done 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "summary_index", event.SummaryIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertFunctionCallArgumentsDeltaEvent 转换 response.function_call_arguments.delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: function_call_arguments_delta
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: tool_use
//   - delta.partial_json: delta
//   - extensions.openai_responses.obfuscation: obfuscation
func convertFunctionCallArgumentsDeltaEvent(event *responsesTypes.ResponseFunctionCallArgumentsDeltaEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventFunctionCallArgumentsDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "tool_use",
		},
		Delta: &adapterTypes.StreamDeltaPayload{
			DeltaType:   "input_json_delta",
			PartialJSON: &event.Delta,
		},
		Extensions: make(map[string]interface{}),
	}

	// 保存 obfuscation 到 extensions
	if event.Obfuscation != nil {
		openaiExt := make(map[string]interface{})
		openaiExt["obfuscation"] = *event.Obfuscation
		contract.Extensions["openai_responses"] = openaiExt
	}

	log.Debug("转换 function_call_arguments.delta 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertFunctionCallArgumentsDoneEvent 转换 response.function_call_arguments.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: function_call_arguments_done
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: tool_use
//   - tool.name: name
//   - tool.arguments: arguments
func convertFunctionCallArgumentsDoneEvent(event *responsesTypes.ResponseFunctionCallArgumentsDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventFunctionCallArgumentsDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "tool_use",
			Tool: &adapterTypes.StreamToolCall{
				Name:      event.Name,
				Arguments: event.Arguments,
			},
		},
	}

	log.Debug("转换 function_call_arguments.done 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertCustomToolCallInputDeltaEvent 转换 response.custom_tool_call_input.delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: custom_tool_call_input_delta
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: tool_use
//   - delta.partial_json: delta
//   - extensions.openai_responses.obfuscation: obfuscation
func convertCustomToolCallInputDeltaEvent(event *responsesTypes.ResponseCustomToolCallInputDeltaEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventCustomToolCallInputDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "tool_use",
		},
		Delta: &adapterTypes.StreamDeltaPayload{
			DeltaType:   "input_json_delta",
			PartialJSON: &event.Delta,
		},
		Extensions: make(map[string]interface{}),
	}

	// 保存 obfuscation 到 extensions
	if event.Obfuscation != nil {
		openaiExt := make(map[string]interface{})
		openaiExt["obfuscation"] = *event.Obfuscation
		contract.Extensions["openai_responses"] = openaiExt
	}

	log.Debug("转换 custom_tool_call_input.delta 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertCustomToolCallInputDoneEvent 转换 response.custom_tool_call_input.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: custom_tool_call_input_done
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: tool_use
//   - tool.arguments: input
func convertCustomToolCallInputDoneEvent(event *responsesTypes.ResponseCustomToolCallInputDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventCustomToolCallInputDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "tool_use",
			Tool: &adapterTypes.StreamToolCall{
				Arguments: event.Input,
			},
		},
	}

	log.Debug("转换 custom_tool_call_input.done 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertMCPCallArgumentsDeltaEvent 转换 response.mcp_call_arguments.delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: mcp_call_arguments_delta
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: tool_use
//   - delta.partial_json: delta
//   - extensions.openai_responses.obfuscation: obfuscation
func convertMCPCallArgumentsDeltaEvent(event *responsesTypes.ResponseMCPCallArgumentsDeltaEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventMCPCallArgumentsDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "tool_use",
		},
		Delta: &adapterTypes.StreamDeltaPayload{
			DeltaType:   "input_json_delta",
			PartialJSON: &event.Delta,
		},
		Extensions: make(map[string]interface{}),
	}

	// 保存 obfuscation 到 extensions
	if event.Obfuscation != nil {
		openaiExt := make(map[string]interface{})
		openaiExt["obfuscation"] = *event.Obfuscation
		contract.Extensions["openai_responses"] = openaiExt
	}

	log.Debug("转换 mcp_call_arguments.delta 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertMCPCallArgumentsDoneEvent 转换 response.mcp_call_arguments.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: mcp_call_arguments_done
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: tool_use
//   - tool.arguments: arguments
func convertMCPCallArgumentsDoneEvent(event *responsesTypes.ResponseMCPCallArgumentsDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventMCPCallArgumentsDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "tool_use",
			Tool: &adapterTypes.StreamToolCall{
				Arguments: event.Arguments,
			},
		},
	}

	log.Debug("转换 mcp_call_arguments.done 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertMCPCallCompletedEvent 转换 response.mcp_call.completed 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: mcp_call_completed
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertMCPCallCompletedEvent(event *responsesTypes.ResponseMCPCallCompletedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventMCPCallCompleted,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 mcp_call.completed 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertMCPCallFailedEvent 转换 response.mcp_call.failed 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: mcp_call_failed
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertMCPCallFailedEvent(event *responsesTypes.ResponseMCPCallFailedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventMCPCallFailed,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 mcp_call.failed 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertMCPCallInProgressEvent 转换 response.mcp_call.in_progress 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: mcp_call_in_progress
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertMCPCallInProgressEvent(event *responsesTypes.ResponseMCPCallInProgressEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventMCPCallInProgress,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 mcp_call.in_progress 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertMCPListToolsCompletedEvent 转换 response.mcp_list_tools.completed 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: mcp_list_tools_completed
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertMCPListToolsCompletedEvent(event *responsesTypes.ResponseMCPListToolsCompletedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventMCPListToolsCompleted,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 mcp_list_tools.completed 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertMCPListToolsFailedEvent 转换 response.mcp_list_tools.failed 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: mcp_list_tools_failed
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertMCPListToolsFailedEvent(event *responsesTypes.ResponseMCPListToolsFailedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventMCPListToolsFailed,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 mcp_list_tools.failed 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertMCPListToolsInProgressEvent 转换 response.mcp_list_tools.in_progress 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: mcp_list_tools_in_progress
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertMCPListToolsInProgressEvent(event *responsesTypes.ResponseMCPListToolsInProgressEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventMCPListToolsInProgress,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 mcp_list_tools.in_progress 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertAudioDeltaEvent 转换 response.audio.delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: audio_delta
//   - response_id: response_id
//   - sequence_number: sequence_number
//   - content.kind: audio
//   - delta.delta_type: audio_delta
//   - delta.text: delta
func convertAudioDeltaEvent(event *responsesTypes.ResponseAudioDeltaEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventAudioDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     event.ResponseID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "audio",
		},
		Delta: &adapterTypes.StreamDeltaPayload{
			DeltaType: "audio_delta",
			Text:      &event.Delta,
		},
	}

	log.Debug("转换 audio.delta 事件完成", "response_id", contract.ResponseID, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertAudioDoneEvent 转换 response.audio.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: audio_done
//   - response_id: response_id
//   - sequence_number: sequence_number
func convertAudioDoneEvent(event *responsesTypes.ResponseAudioDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventAudioDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     event.ResponseID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 audio.done 事件完成", "response_id", contract.ResponseID, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertAudioTranscriptDeltaEvent 转换 response.audio.transcript.delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: audio_transcript_delta
//   - response_id: response_id
//   - sequence_number: sequence_number
//   - content.kind: audio
//   - delta.text: delta
func convertAudioTranscriptDeltaEvent(event *responsesTypes.ResponseAudioTranscriptDeltaEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventAudioTranscriptDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     event.ResponseID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "audio",
		},
		Delta: &adapterTypes.StreamDeltaPayload{
			DeltaType: "text_delta",
			Text:      &event.Delta,
		},
	}

	log.Debug("转换 audio.transcript.delta 事件完成", "response_id", contract.ResponseID, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertAudioTranscriptDoneEvent 转换 response.audio.transcript.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: audio_transcript_done
//   - response_id: response_id
//   - sequence_number: sequence_number
func convertAudioTranscriptDoneEvent(event *responsesTypes.ResponseAudioTranscriptDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventAudioTranscriptDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     event.ResponseID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 audio.transcript.done 事件完成", "response_id", contract.ResponseID, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertCodeInterpreterCallCodeDeltaEvent 转换 response.code_interpreter_call_code.delta 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: code_interpreter_call_code_delta
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: call_code
//   - delta.text: delta
func convertCodeInterpreterCallCodeDeltaEvent(event *responsesTypes.ResponseCodeInterpreterCallCodeDeltaEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventCodeInterpreterCallCodeDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "call_code",
		},
		Delta: &adapterTypes.StreamDeltaPayload{
			DeltaType: "text_delta",
			Text:      &event.Delta,
		},
	}

	log.Debug("转换 code_interpreter_call_code.delta 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertCodeInterpreterCallCodeDoneEvent 转换 response.code_interpreter_call_code.done 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: code_interpreter_call_code_done
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: call_code
//   - content.text: code
func convertCodeInterpreterCallCodeDoneEvent(event *responsesTypes.ResponseCodeInterpreterCallCodeDoneEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventCodeInterpreterCallCodeDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "call_code",
			Text: &event.Code,
		},
	}

	log.Debug("转换 code_interpreter_call_code.done 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertCodeInterpreterCallCompletedEvent 转换 response.code_interpreter_call.completed 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: code_interpreter_call_completed
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertCodeInterpreterCallCompletedEvent(event *responsesTypes.ResponseCodeInterpreterCallCompletedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventCodeInterpreterCallCompleted,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 code_interpreter_call.completed 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertCodeInterpreterCallInProgressEvent 转换 response.code_interpreter_call.in_progress 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: code_interpreter_call_in_progress
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertCodeInterpreterCallInProgressEvent(event *responsesTypes.ResponseCodeInterpreterCallInProgressEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventCodeInterpreterCallInProgress,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 code_interpreter_call.in_progress 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertCodeInterpreterCallInterpretingEvent 转换 response.code_interpreter_call.interpreting 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: code_interpreter_call_interpreting
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertCodeInterpreterCallInterpretingEvent(event *responsesTypes.ResponseCodeInterpreterCallInterpretingEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventCodeInterpreterCallInterpreting,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 code_interpreter_call.interpreting 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertFileSearchCallCompletedEvent 转换 response.file_search_call.completed 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: file_search_call_completed
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertFileSearchCallCompletedEvent(event *responsesTypes.ResponseFileSearchCallCompletedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventFileSearchCallCompleted,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 file_search_call.completed 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertFileSearchCallInProgressEvent 转换 response.file_search_call.in_progress 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: file_search_call_in_progress
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertFileSearchCallInProgressEvent(event *responsesTypes.ResponseFileSearchCallInProgressEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventFileSearchCallInProgress,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 file_search_call.in_progress 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertFileSearchCallSearchingEvent 转换 response.file_search_call.searching 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: file_search_call_searching
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertFileSearchCallSearchingEvent(event *responsesTypes.ResponseFileSearchCallSearchingEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventFileSearchCallSearching,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 file_search_call.searching 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertWebSearchCallCompletedEvent 转换 response.web_search_call.completed 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: web_search_call_completed
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertWebSearchCallCompletedEvent(event *responsesTypes.ResponseWebSearchCallCompletedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventWebSearchCallCompleted,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 web_search_call.completed 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertWebSearchCallInProgressEvent 转换 response.web_search_call.in_progress 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: web_search_call_in_progress
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertWebSearchCallInProgressEvent(event *responsesTypes.ResponseWebSearchCallInProgressEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventWebSearchCallInProgress,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 web_search_call.in_progress 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertWebSearchCallSearchingEvent 转换 response.web_search_call.searching 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: web_search_call_searching
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertWebSearchCallSearchingEvent(event *responsesTypes.ResponseWebSearchCallSearchingEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventWebSearchCallSearching,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 web_search_call.searching 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertImageGenCallCompletedEvent 转换 response.image_generation_call.completed 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: image_generation_call_completed
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertImageGenCallCompletedEvent(event *responsesTypes.ResponseImageGenCallCompletedEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventImageGenerationCallCompleted,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 image_generation_call.completed 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertImageGenCallGeneratingEvent 转换 response.image_generation_call.generating 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: image_generation_call_generating
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertImageGenCallGeneratingEvent(event *responsesTypes.ResponseImageGenCallGeneratingEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventImageGenerationCallGenerating,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 image_generation_call.generating 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertImageGenCallInProgressEvent 转换 response.image_generation_call.in_progress 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: image_generation_call_in_progress
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
func convertImageGenCallInProgressEvent(event *responsesTypes.ResponseImageGenCallInProgressEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventImageGenerationCallInProgress,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
	}

	log.Debug("转换 image_generation_call.in_progress 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber)
	return contract, nil
}

// convertImageGenCallPartialImageEvent 转换 response.image_generation_call.partial_image 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: image_generation_call_partial_image
//   - output_index: output_index
//   - item_id: item_id
//   - sequence_number: sequence_number
//   - content.kind: image
//   - content.raw.partial_image_b64: partial_image_b64
//   - content.raw.partial_image_index: partial_image_index
func convertImageGenCallPartialImageEvent(event *responsesTypes.ResponseImageGenCallPartialImageEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventImageGenerationCallPartialImage,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		OutputIndex:    event.OutputIndex,
		ItemID:         event.ItemID,
		SequenceNumber: event.SequenceNumber,
		Content: &adapterTypes.StreamContentPayload{
			Kind: "image",
			Raw:  make(map[string]interface{}),
		},
	}

	// 保存 partial_image_b64 和 partial_image_index 到 content.raw
	contract.Content.Raw["partial_image_b64"] = event.PartialImageB64
	contract.Content.Raw["partial_image_index"] = event.PartialImageIndex

	log.Debug("转换 image_generation_call.partial_image 事件完成", "item_id", contract.ItemID, "output_index", contract.OutputIndex, "sequence_number", contract.SequenceNumber, "partial_image_index", event.PartialImageIndex)
	return contract, nil
}

// convertErrorEvent 转换 error 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.4）：
//   - type: error
//   - error.message: message
//   - error.code: code
//   - error.param: param
//   - sequence_number: sequence_number
func convertErrorEvent(event *responsesTypes.ResponseErrorEvent, log logger.Logger) (*adapterTypes.StreamEventContract, error) {
	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventError,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		SequenceNumber: event.SequenceNumber,
		Error: &adapterTypes.StreamErrorPayload{
			Message: event.Message,
			Raw:     make(map[string]interface{}),
		},
	}

	// 映射 code 和 param
	if event.Code != nil {
		contract.Error.Code = *event.Code
	}
	if event.Param != nil {
		contract.Error.Param = *event.Param
	}

	log.Error("转换 error 事件完成", "error_code", contract.Error.Code, "error_message", contract.Error.Message, "error_param", contract.Error.Param)
	return contract, nil
}
