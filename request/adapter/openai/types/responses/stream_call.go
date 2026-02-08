package responses

// ResponseAudioDeltaEvent response.audio.delta 事件。
type ResponseAudioDeltaEvent struct {
	Type           StreamEventType `json:"type"`
	ResponseID     string          `json:"response_id"`
	Delta          string          `json:"delta"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseAudioDoneEvent response.audio.done 事件。
type ResponseAudioDoneEvent struct {
	Type           StreamEventType `json:"type"`
	ResponseID     string          `json:"response_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseAudioTranscriptDeltaEvent response.audio.transcript.delta 事件。
type ResponseAudioTranscriptDeltaEvent struct {
	Type           StreamEventType `json:"type"`
	ResponseID     string          `json:"response_id"`
	Delta          string          `json:"delta"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseAudioTranscriptDoneEvent response.audio.transcript.done 事件。
type ResponseAudioTranscriptDoneEvent struct {
	Type           StreamEventType `json:"type"`
	ResponseID     string          `json:"response_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseCodeInterpreterCallCodeDeltaEvent response.code_interpreter_call_code.delta 事件。
type ResponseCodeInterpreterCallCodeDeltaEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	Delta          string          `json:"delta"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseCodeInterpreterCallCodeDoneEvent response.code_interpreter_call_code.done 事件。
type ResponseCodeInterpreterCallCodeDoneEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	Code           string          `json:"code"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseCodeInterpreterCallCompletedEvent response.code_interpreter_call.completed 事件。
type ResponseCodeInterpreterCallCompletedEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseCodeInterpreterCallInProgressEvent response.code_interpreter_call.in_progress 事件。
type ResponseCodeInterpreterCallInProgressEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseCodeInterpreterCallInterpretingEvent response.code_interpreter_call.interpreting 事件。
type ResponseCodeInterpreterCallInterpretingEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseFileSearchCallCompletedEvent response.file_search_call.completed 事件。
type ResponseFileSearchCallCompletedEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseFileSearchCallInProgressEvent response.file_search_call.in_progress 事件。
type ResponseFileSearchCallInProgressEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseFileSearchCallSearchingEvent response.file_search_call.searching 事件。
type ResponseFileSearchCallSearchingEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseWebSearchCallCompletedEvent response.web_search_call.completed 事件。
type ResponseWebSearchCallCompletedEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseWebSearchCallInProgressEvent response.web_search_call.in_progress 事件。
type ResponseWebSearchCallInProgressEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseWebSearchCallSearchingEvent response.web_search_call.searching 事件。
type ResponseWebSearchCallSearchingEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseImageGenCallCompletedEvent response.image_generation_call.completed 事件。
type ResponseImageGenCallCompletedEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseImageGenCallGeneratingEvent response.image_generation_call.generating 事件。
type ResponseImageGenCallGeneratingEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseImageGenCallInProgressEvent response.image_generation_call.in_progress 事件。
type ResponseImageGenCallInProgressEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseImageGenCallPartialImageEvent response.image_generation_call.partial_image 事件。
type ResponseImageGenCallPartialImageEvent struct {
	Type              StreamEventType `json:"type"`
	OutputIndex       int             `json:"output_index"`
	ItemID            string          `json:"item_id"`
	SequenceNumber    int             `json:"sequence_number"`
	PartialImageIndex int             `json:"partial_image_index"`
	PartialImageB64   string          `json:"partial_image_b64"`
}

// ResponseMCPCallArgumentsDeltaEvent response.mcp_call_arguments.delta 事件。
type ResponseMCPCallArgumentsDeltaEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	Delta          string          `json:"delta"`
	SequenceNumber int             `json:"sequence_number"`
	Obfuscation    *string         `json:"obfuscation,omitempty"`
}

// ResponseMCPCallArgumentsDoneEvent response.mcp_call_arguments.done 事件。
type ResponseMCPCallArgumentsDoneEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	Arguments      string          `json:"arguments"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseMCPCallCompletedEvent response.mcp_call.completed 事件。
type ResponseMCPCallCompletedEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseMCPCallFailedEvent response.mcp_call.failed 事件。
type ResponseMCPCallFailedEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseMCPCallInProgressEvent response.mcp_call.in_progress 事件。
type ResponseMCPCallInProgressEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseMCPListToolsCompletedEvent response.mcp_list_tools.completed 事件。
type ResponseMCPListToolsCompletedEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseMCPListToolsFailedEvent response.mcp_list_tools.failed 事件。
type ResponseMCPListToolsFailedEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseMCPListToolsInProgressEvent response.mcp_list_tools.in_progress 事件。
type ResponseMCPListToolsInProgressEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	SequenceNumber int             `json:"sequence_number"`
}
