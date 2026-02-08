package responses

import (
	"encoding/json"
	"fmt"
)

// StreamEvent 表示 Responses 流式事件联合类型。
// 通过自定义 UnmarshalJSON/MarshalJSON 实现根据 type 字段分发到对应的具体事件类型。
type StreamEvent struct {
	// 音频事件
	AudioDelta           *ResponseAudioDeltaEvent
	AudioDone            *ResponseAudioDoneEvent
	AudioTranscriptDelta *ResponseAudioTranscriptDeltaEvent
	AudioTranscriptDone  *ResponseAudioTranscriptDoneEvent

	// Code Interpreter 事件
	CodeInterpreterCallCodeDelta    *ResponseCodeInterpreterCallCodeDeltaEvent
	CodeInterpreterCallCodeDone     *ResponseCodeInterpreterCallCodeDoneEvent
	CodeInterpreterCallCompleted    *ResponseCodeInterpreterCallCompletedEvent
	CodeInterpreterCallInProgress   *ResponseCodeInterpreterCallInProgressEvent
	CodeInterpreterCallInterpreting *ResponseCodeInterpreterCallInterpretingEvent

	// File Search 事件
	FileSearchCallCompleted  *ResponseFileSearchCallCompletedEvent
	FileSearchCallInProgress *ResponseFileSearchCallInProgressEvent
	FileSearchCallSearching  *ResponseFileSearchCallSearchingEvent

	// Web Search 事件
	WebSearchCallCompleted  *ResponseWebSearchCallCompletedEvent
	WebSearchCallInProgress *ResponseWebSearchCallInProgressEvent
	WebSearchCallSearching  *ResponseWebSearchCallSearchingEvent

	// Image Generation 事件
	ImageGenCallCompleted    *ResponseImageGenCallCompletedEvent
	ImageGenCallGenerating   *ResponseImageGenCallGeneratingEvent
	ImageGenCallInProgress   *ResponseImageGenCallInProgressEvent
	ImageGenCallPartialImage *ResponseImageGenCallPartialImageEvent

	// MCP 事件
	MCPCallArgumentsDelta  *ResponseMCPCallArgumentsDeltaEvent
	MCPCallArgumentsDone   *ResponseMCPCallArgumentsDoneEvent
	MCPCallCompleted       *ResponseMCPCallCompletedEvent
	MCPCallFailed          *ResponseMCPCallFailedEvent
	MCPCallInProgress      *ResponseMCPCallInProgressEvent
	MCPListToolsCompleted  *ResponseMCPListToolsCompletedEvent
	MCPListToolsFailed     *ResponseMCPListToolsFailedEvent
	MCPListToolsInProgress *ResponseMCPListToolsInProgressEvent

	// Custom Tool Call 事件
	CustomToolCallInputDelta *ResponseCustomToolCallInputDeltaEvent
	CustomToolCallInputDone  *ResponseCustomToolCallInputDoneEvent

	// Function Call 事件
	FunctionCallArgumentsDelta *ResponseFunctionCallArgumentsDeltaEvent
	FunctionCallArgumentsDone  *ResponseFunctionCallArgumentsDoneEvent

	// Response 生命周期事件
	Created    *ResponseCreatedEvent
	InProgress *ResponseInProgressEvent
	Completed  *ResponseCompletedEvent
	Failed     *ResponseFailedEvent
	Incomplete *ResponseIncompleteEvent
	Queued     *ResponseQueuedEvent

	// Output Item 事件
	OutputItemAdded *ResponseOutputItemAddedEvent
	OutputItemDone  *ResponseOutputItemDoneEvent

	// Content Part 事件
	ContentPartAdded *ResponseContentPartAddedEvent
	ContentPartDone  *ResponseContentPartDoneEvent

	// Output Text 事件
	OutputTextDelta           *ResponseOutputTextDeltaEvent
	OutputTextDone            *ResponseOutputTextDoneEvent
	OutputTextAnnotationAdded *ResponseOutputTextAnnotationAddedEvent

	// Refusal 事件
	RefusalDelta *ResponseRefusalDeltaEvent
	RefusalDone  *ResponseRefusalDoneEvent

	// Reasoning Text 事件
	ReasoningTextDelta *ResponseReasoningTextDeltaEvent
	ReasoningTextDone  *ResponseReasoningTextDoneEvent

	// Reasoning Summary 事件
	ReasoningSummaryPartAdded *ResponseReasoningSummaryPartAddedEvent
	ReasoningSummaryPartDone  *ResponseReasoningSummaryPartDoneEvent
	ReasoningSummaryTextDelta *ResponseReasoningSummaryTextDeltaEvent
	ReasoningSummaryTextDone  *ResponseReasoningSummaryTextDoneEvent

	// Error 事件
	Error *ResponseErrorEvent
}

// UnmarshalJSON 实现 StreamEvent 的反序列化。
// 根据 type 字段将 JSON 数据分发到对应的具体事件类型。
func (e *StreamEvent) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type StreamEventType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("流式事件解析失败：%w", err)
	}

	switch t.Type {
	// 音频事件
	case StreamEventAudioDelta:
		var v ResponseAudioDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.audio.delta 事件解析失败：%w", err)
		}
		e.AudioDelta = &v
	case StreamEventAudioDone:
		var v ResponseAudioDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.audio.done 事件解析失败：%w", err)
		}
		e.AudioDone = &v
	case StreamEventAudioTranscriptDelta:
		var v ResponseAudioTranscriptDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.audio.transcript.delta 事件解析失败：%w", err)
		}
		e.AudioTranscriptDelta = &v
	case StreamEventAudioTranscriptDone:
		var v ResponseAudioTranscriptDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.audio.transcript.done 事件解析失败：%w", err)
		}
		e.AudioTranscriptDone = &v

	// Code Interpreter 事件
	case StreamEventCodeInterpreterCallCodeDelta:
		var v ResponseCodeInterpreterCallCodeDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.code_interpreter_call_code.delta 事件解析失败：%w", err)
		}
		e.CodeInterpreterCallCodeDelta = &v
	case StreamEventCodeInterpreterCallCodeDone:
		var v ResponseCodeInterpreterCallCodeDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.code_interpreter_call_code.done 事件解析失败：%w", err)
		}
		e.CodeInterpreterCallCodeDone = &v
	case StreamEventCodeInterpreterCallCompleted:
		var v ResponseCodeInterpreterCallCompletedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.code_interpreter_call.completed 事件解析失败：%w", err)
		}
		e.CodeInterpreterCallCompleted = &v
	case StreamEventCodeInterpreterCallInProgress:
		var v ResponseCodeInterpreterCallInProgressEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.code_interpreter_call.in_progress 事件解析失败：%w", err)
		}
		e.CodeInterpreterCallInProgress = &v
	case StreamEventCodeInterpreterCallInterpreting:
		var v ResponseCodeInterpreterCallInterpretingEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.code_interpreter_call.interpreting 事件解析失败：%w", err)
		}
		e.CodeInterpreterCallInterpreting = &v

	// File Search 事件
	case StreamEventFileSearchCallCompleted:
		var v ResponseFileSearchCallCompletedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.file_search_call.completed 事件解析失败：%w", err)
		}
		e.FileSearchCallCompleted = &v
	case StreamEventFileSearchCallInProgress:
		var v ResponseFileSearchCallInProgressEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.file_search_call.in_progress 事件解析失败：%w", err)
		}
		e.FileSearchCallInProgress = &v
	case StreamEventFileSearchCallSearching:
		var v ResponseFileSearchCallSearchingEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.file_search_call.searching 事件解析失败：%w", err)
		}
		e.FileSearchCallSearching = &v

	// Web Search 事件
	case StreamEventWebSearchCallCompleted:
		var v ResponseWebSearchCallCompletedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.web_search_call.completed 事件解析失败：%w", err)
		}
		e.WebSearchCallCompleted = &v
	case StreamEventWebSearchCallInProgress:
		var v ResponseWebSearchCallInProgressEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.web_search_call.in_progress 事件解析失败：%w", err)
		}
		e.WebSearchCallInProgress = &v
	case StreamEventWebSearchCallSearching:
		var v ResponseWebSearchCallSearchingEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.web_search_call.searching 事件解析失败：%w", err)
		}
		e.WebSearchCallSearching = &v

	// Image Generation 事件
	case StreamEventImageGenCallCompleted:
		var v ResponseImageGenCallCompletedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.image_generation_call.completed 事件解析失败：%w", err)
		}
		e.ImageGenCallCompleted = &v
	case StreamEventImageGenCallGenerating:
		var v ResponseImageGenCallGeneratingEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.image_generation_call.generating 事件解析失败：%w", err)
		}
		e.ImageGenCallGenerating = &v
	case StreamEventImageGenCallInProgress:
		var v ResponseImageGenCallInProgressEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.image_generation_call.in_progress 事件解析失败：%w", err)
		}
		e.ImageGenCallInProgress = &v
	case StreamEventImageGenCallPartialImage:
		var v ResponseImageGenCallPartialImageEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.image_generation_call.partial_image 事件解析失败：%w", err)
		}
		e.ImageGenCallPartialImage = &v

	// MCP 事件
	case StreamEventMCPCallArgumentsDelta:
		var v ResponseMCPCallArgumentsDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.mcp_call_arguments.delta 事件解析失败：%w", err)
		}
		e.MCPCallArgumentsDelta = &v
	case StreamEventMCPCallArgumentsDone:
		var v ResponseMCPCallArgumentsDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.mcp_call_arguments.done 事件解析失败：%w", err)
		}
		e.MCPCallArgumentsDone = &v
	case StreamEventMCPCallCompleted:
		var v ResponseMCPCallCompletedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.mcp_call.completed 事件解析失败：%w", err)
		}
		e.MCPCallCompleted = &v
	case StreamEventMCPCallFailed:
		var v ResponseMCPCallFailedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.mcp_call.failed 事件解析失败：%w", err)
		}
		e.MCPCallFailed = &v
	case StreamEventMCPCallInProgress:
		var v ResponseMCPCallInProgressEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.mcp_call.in_progress 事件解析失败：%w", err)
		}
		e.MCPCallInProgress = &v
	case StreamEventMCPListToolsCompleted:
		var v ResponseMCPListToolsCompletedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.mcp_list_tools.completed 事件解析失败：%w", err)
		}
		e.MCPListToolsCompleted = &v
	case StreamEventMCPListToolsFailed:
		var v ResponseMCPListToolsFailedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.mcp_list_tools.failed 事件解析失败：%w", err)
		}
		e.MCPListToolsFailed = &v
	case StreamEventMCPListToolsInProgress:
		var v ResponseMCPListToolsInProgressEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.mcp_list_tools.in_progress 事件解析失败：%w", err)
		}
		e.MCPListToolsInProgress = &v

	// Custom Tool Call 事件
	case StreamEventCustomToolCallInputDelta:
		var v ResponseCustomToolCallInputDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.custom_tool_call_input.delta 事件解析失败：%w", err)
		}
		e.CustomToolCallInputDelta = &v
	case StreamEventCustomToolCallInputDone:
		var v ResponseCustomToolCallInputDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.custom_tool_call_input.done 事件解析失败：%w", err)
		}
		e.CustomToolCallInputDone = &v

	// Function Call 事件
	case StreamEventFunctionCallArgumentsDelta:
		var v ResponseFunctionCallArgumentsDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.function_call_arguments.delta 事件解析失败：%w", err)
		}
		e.FunctionCallArgumentsDelta = &v
	case StreamEventFunctionCallArgumentsDone:
		var v ResponseFunctionCallArgumentsDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.function_call_arguments.done 事件解析失败：%w", err)
		}
		e.FunctionCallArgumentsDone = &v

	// Response 生命周期事件
	case StreamEventCreated:
		var v ResponseCreatedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.created 事件解析失败：%w", err)
		}
		e.Created = &v
	case StreamEventInProgress:
		var v ResponseInProgressEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.in_progress 事件解析失败：%w", err)
		}
		e.InProgress = &v
	case StreamEventCompleted:
		var v ResponseCompletedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.completed 事件解析失败：%w", err)
		}
		e.Completed = &v
	case StreamEventFailed:
		var v ResponseFailedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.failed 事件解析失败：%w", err)
		}
		e.Failed = &v
	case StreamEventIncomplete:
		var v ResponseIncompleteEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.incomplete 事件解析失败：%w", err)
		}
		e.Incomplete = &v
	case StreamEventQueued:
		var v ResponseQueuedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.queued 事件解析失败：%w", err)
		}
		e.Queued = &v

	// Output Item 事件
	case StreamEventOutputItemAdded:
		var v ResponseOutputItemAddedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.output_item.added 事件解析失败：%w", err)
		}
		e.OutputItemAdded = &v
	case StreamEventOutputItemDone:
		var v ResponseOutputItemDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.output_item.done 事件解析失败：%w", err)
		}
		e.OutputItemDone = &v

	// Content Part 事件
	case StreamEventContentPartAdded:
		var v ResponseContentPartAddedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.content_part.added 事件解析失败：%w", err)
		}
		e.ContentPartAdded = &v
	case StreamEventContentPartDone:
		var v ResponseContentPartDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.content_part.done 事件解析失败：%w", err)
		}
		e.ContentPartDone = &v

	// Output Text 事件
	case StreamEventOutputTextDelta:
		var v ResponseOutputTextDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.output_text.delta 事件解析失败：%w", err)
		}
		e.OutputTextDelta = &v
	case StreamEventOutputTextDone:
		var v ResponseOutputTextDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.output_text.done 事件解析失败：%w", err)
		}
		e.OutputTextDone = &v
	case StreamEventOutputTextAnnotationAdded:
		var v ResponseOutputTextAnnotationAddedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.output_text.annotation.added 事件解析失败：%w", err)
		}
		e.OutputTextAnnotationAdded = &v

	// Refusal 事件
	case StreamEventRefusalDelta:
		var v ResponseRefusalDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.refusal.delta 事件解析失败：%w", err)
		}
		e.RefusalDelta = &v
	case StreamEventRefusalDone:
		var v ResponseRefusalDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.refusal.done 事件解析失败：%w", err)
		}
		e.RefusalDone = &v

	// Reasoning Text 事件
	case StreamEventReasoningTextDelta:
		var v ResponseReasoningTextDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.reasoning_text.delta 事件解析失败：%w", err)
		}
		e.ReasoningTextDelta = &v
	case StreamEventReasoningTextDone:
		var v ResponseReasoningTextDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.reasoning_text.done 事件解析失败：%w", err)
		}
		e.ReasoningTextDone = &v

	// Reasoning Summary 事件
	case StreamEventReasoningSummaryPartAdded:
		var v ResponseReasoningSummaryPartAddedEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.reasoning_summary_part.added 事件解析失败：%w", err)
		}
		e.ReasoningSummaryPartAdded = &v
	case StreamEventReasoningSummaryPartDone:
		var v ResponseReasoningSummaryPartDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.reasoning_summary_part.done 事件解析失败：%w", err)
		}
		e.ReasoningSummaryPartDone = &v
	case StreamEventReasoningSummaryTextDelta:
		var v ResponseReasoningSummaryTextDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.reasoning_summary_text.delta 事件解析失败：%w", err)
		}
		e.ReasoningSummaryTextDelta = &v
	case StreamEventReasoningSummaryTextDone:
		var v ResponseReasoningSummaryTextDoneEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("response.reasoning_summary_text.done 事件解析失败：%w", err)
		}
		e.ReasoningSummaryTextDone = &v

	// Error 事件
	case StreamEventError:
		var v ResponseErrorEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("error 事件解析失败：%w", err)
		}
		e.Error = &v

	default:
		return fmt.Errorf("不支持的流式事件类型: %s", t.Type)
	}

	return nil
}

// MarshalJSON 实现 StreamEvent 的序列化。
// 将已设置的具体事件类型序列化为 JSON。
func (e StreamEvent) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}

	// 音频事件
	if e.AudioDelta != nil {
		set(e.AudioDelta)
	}
	if e.AudioDone != nil {
		set(e.AudioDone)
	}
	if e.AudioTranscriptDelta != nil {
		set(e.AudioTranscriptDelta)
	}
	if e.AudioTranscriptDone != nil {
		set(e.AudioTranscriptDone)
	}

	// Code Interpreter 事件
	if e.CodeInterpreterCallCodeDelta != nil {
		set(e.CodeInterpreterCallCodeDelta)
	}
	if e.CodeInterpreterCallCodeDone != nil {
		set(e.CodeInterpreterCallCodeDone)
	}
	if e.CodeInterpreterCallCompleted != nil {
		set(e.CodeInterpreterCallCompleted)
	}
	if e.CodeInterpreterCallInProgress != nil {
		set(e.CodeInterpreterCallInProgress)
	}
	if e.CodeInterpreterCallInterpreting != nil {
		set(e.CodeInterpreterCallInterpreting)
	}

	// File Search 事件
	if e.FileSearchCallCompleted != nil {
		set(e.FileSearchCallCompleted)
	}
	if e.FileSearchCallInProgress != nil {
		set(e.FileSearchCallInProgress)
	}
	if e.FileSearchCallSearching != nil {
		set(e.FileSearchCallSearching)
	}

	// Web Search 事件
	if e.WebSearchCallCompleted != nil {
		set(e.WebSearchCallCompleted)
	}
	if e.WebSearchCallInProgress != nil {
		set(e.WebSearchCallInProgress)
	}
	if e.WebSearchCallSearching != nil {
		set(e.WebSearchCallSearching)
	}

	// Image Generation 事件
	if e.ImageGenCallCompleted != nil {
		set(e.ImageGenCallCompleted)
	}
	if e.ImageGenCallGenerating != nil {
		set(e.ImageGenCallGenerating)
	}
	if e.ImageGenCallInProgress != nil {
		set(e.ImageGenCallInProgress)
	}
	if e.ImageGenCallPartialImage != nil {
		set(e.ImageGenCallPartialImage)
	}

	// MCP 事件
	if e.MCPCallArgumentsDelta != nil {
		set(e.MCPCallArgumentsDelta)
	}
	if e.MCPCallArgumentsDone != nil {
		set(e.MCPCallArgumentsDone)
	}
	if e.MCPCallCompleted != nil {
		set(e.MCPCallCompleted)
	}
	if e.MCPCallFailed != nil {
		set(e.MCPCallFailed)
	}
	if e.MCPCallInProgress != nil {
		set(e.MCPCallInProgress)
	}
	if e.MCPListToolsCompleted != nil {
		set(e.MCPListToolsCompleted)
	}
	if e.MCPListToolsFailed != nil {
		set(e.MCPListToolsFailed)
	}
	if e.MCPListToolsInProgress != nil {
		set(e.MCPListToolsInProgress)
	}

	// Custom Tool Call 事件
	if e.CustomToolCallInputDelta != nil {
		set(e.CustomToolCallInputDelta)
	}
	if e.CustomToolCallInputDone != nil {
		set(e.CustomToolCallInputDone)
	}

	// Function Call 事件
	if e.FunctionCallArgumentsDelta != nil {
		set(e.FunctionCallArgumentsDelta)
	}
	if e.FunctionCallArgumentsDone != nil {
		set(e.FunctionCallArgumentsDone)
	}

	// Response 生命周期事件
	if e.Created != nil {
		set(e.Created)
	}
	if e.InProgress != nil {
		set(e.InProgress)
	}
	if e.Completed != nil {
		set(e.Completed)
	}
	if e.Failed != nil {
		set(e.Failed)
	}
	if e.Incomplete != nil {
		set(e.Incomplete)
	}
	if e.Queued != nil {
		set(e.Queued)
	}

	// Output Item 事件
	if e.OutputItemAdded != nil {
		set(e.OutputItemAdded)
	}
	if e.OutputItemDone != nil {
		set(e.OutputItemDone)
	}

	// Content Part 事件
	if e.ContentPartAdded != nil {
		set(e.ContentPartAdded)
	}
	if e.ContentPartDone != nil {
		set(e.ContentPartDone)
	}

	// Output Text 事件
	if e.OutputTextDelta != nil {
		set(e.OutputTextDelta)
	}
	if e.OutputTextDone != nil {
		set(e.OutputTextDone)
	}
	if e.OutputTextAnnotationAdded != nil {
		set(e.OutputTextAnnotationAdded)
	}

	// Refusal 事件
	if e.RefusalDelta != nil {
		set(e.RefusalDelta)
	}
	if e.RefusalDone != nil {
		set(e.RefusalDone)
	}

	// Reasoning Text 事件
	if e.ReasoningTextDelta != nil {
		set(e.ReasoningTextDelta)
	}
	if e.ReasoningTextDone != nil {
		set(e.ReasoningTextDone)
	}

	// Reasoning Summary 事件
	if e.ReasoningSummaryPartAdded != nil {
		set(e.ReasoningSummaryPartAdded)
	}
	if e.ReasoningSummaryPartDone != nil {
		set(e.ReasoningSummaryPartDone)
	}
	if e.ReasoningSummaryTextDelta != nil {
		set(e.ReasoningSummaryTextDelta)
	}
	if e.ReasoningSummaryTextDone != nil {
		set(e.ReasoningSummaryTextDone)
	}

	// Error 事件
	if e.Error != nil {
		set(e.Error)
	}

	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("流式事件只能设置一种类型")
	}
	return json.Marshal(payload)
}
