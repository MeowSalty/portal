package types

// StreamEventType 表示中间流式事件类型。
// 使用细粒度事件集合以满足往返转换的保真度需求。
type StreamEventType string

const (
	// 消息级事件
	StreamEventMessageStart StreamEventType = "message_start"
	StreamEventMessageDelta StreamEventType = "message_delta"
	StreamEventMessageStop  StreamEventType = "message_stop"

	// 内容块事件（兼容 Anthropic）
	StreamEventContentBlockStart StreamEventType = "content_block_start"
	StreamEventContentBlockDelta StreamEventType = "content_block_delta"
	StreamEventContentBlockStop  StreamEventType = "content_block_stop"

	// 输出项事件（兼容 Responses）
	StreamEventOutputItemAdded StreamEventType = "output_item_added"
	StreamEventOutputItemDone  StreamEventType = "output_item_done"

	// 内容分片事件（兼容 Responses）
	StreamEventContentPartAdded StreamEventType = "content_part_added"
	StreamEventContentPartDone  StreamEventType = "content_part_done"

	// 文本事件（兼容 Responses）
	StreamEventOutputTextDelta           StreamEventType = "output_text_delta"
	StreamEventOutputTextDone            StreamEventType = "output_text_done"
	StreamEventOutputTextAnnotationAdded StreamEventType = "output_text_annotation_added"

	// 响应生命周期事件（兼容 Responses）
	StreamEventResponseCreated    StreamEventType = "response_created"
	StreamEventResponseInProgress StreamEventType = "response_in_progress"
	StreamEventResponseCompleted  StreamEventType = "response_completed"
	StreamEventResponseFailed     StreamEventType = "response_failed"
	StreamEventResponseIncomplete StreamEventType = "response_incomplete"
	StreamEventResponseQueued     StreamEventType = "response_queued"

	// 拒绝事件（兼容 OpenAI Responses）
	StreamEventRefusalDelta StreamEventType = "refusal_delta"
	StreamEventRefusalDone  StreamEventType = "refusal_done"

	// 推理事件（兼容 OpenAI Responses）
	StreamEventReasoningTextDelta        StreamEventType = "reasoning_text_delta"
	StreamEventReasoningTextDone         StreamEventType = "reasoning_text_done"
	StreamEventReasoningSummaryPartAdded StreamEventType = "reasoning_summary_part_added"
	StreamEventReasoningSummaryPartDone  StreamEventType = "reasoning_summary_part_done"
	StreamEventReasoningSummaryTextDelta StreamEventType = "reasoning_summary_text_delta"
	StreamEventReasoningSummaryTextDone  StreamEventType = "reasoning_summary_text_done"

	// 函数调用事件（兼容 OpenAI Responses）
	StreamEventFunctionCallArgumentsDelta StreamEventType = "function_call_arguments_delta"
	StreamEventFunctionCallArgumentsDone  StreamEventType = "function_call_arguments_done"

	// 自定义工具调用事件（兼容 OpenAI Responses）
	StreamEventCustomToolCallInputDelta StreamEventType = "custom_tool_call_input_delta"
	StreamEventCustomToolCallInputDone  StreamEventType = "custom_tool_call_input_done"

	// MCP 调用事件（兼容 OpenAI Responses）
	StreamEventMCPCallArgumentsDelta  StreamEventType = "mcp_call_arguments_delta"
	StreamEventMCPCallArgumentsDone   StreamEventType = "mcp_call_arguments_done"
	StreamEventMCPCallCompleted       StreamEventType = "mcp_call_completed"
	StreamEventMCPCallFailed          StreamEventType = "mcp_call_failed"
	StreamEventMCPCallInProgress      StreamEventType = "mcp_call_in_progress"
	StreamEventMCPListToolsCompleted  StreamEventType = "mcp_list_tools_completed"
	StreamEventMCPListToolsFailed     StreamEventType = "mcp_list_tools_failed"
	StreamEventMCPListToolsInProgress StreamEventType = "mcp_list_tools_in_progress"

	// 音频事件（兼容 OpenAI Responses）
	StreamEventAudioDelta           StreamEventType = "audio_delta"
	StreamEventAudioDone            StreamEventType = "audio_done"
	StreamEventAudioTranscriptDelta StreamEventType = "audio_transcript_delta"
	StreamEventAudioTranscriptDone  StreamEventType = "audio_transcript_done"

	// Code Interpreter 事件（兼容 OpenAI Responses）
	StreamEventCodeInterpreterCallCodeDelta    StreamEventType = "code_interpreter_call_code_delta"
	StreamEventCodeInterpreterCallCodeDone     StreamEventType = "code_interpreter_call_code_done"
	StreamEventCodeInterpreterCallCompleted    StreamEventType = "code_interpreter_call_completed"
	StreamEventCodeInterpreterCallInProgress   StreamEventType = "code_interpreter_call_in_progress"
	StreamEventCodeInterpreterCallInterpreting StreamEventType = "code_interpreter_call_interpreting"

	// File Search 事件（兼容 OpenAI Responses）
	StreamEventFileSearchCallCompleted  StreamEventType = "file_search_call_completed"
	StreamEventFileSearchCallInProgress StreamEventType = "file_search_call_in_progress"
	StreamEventFileSearchCallSearching  StreamEventType = "file_search_call_searching"

	// Web Search 事件（兼容 OpenAI Responses）
	StreamEventWebSearchCallCompleted  StreamEventType = "web_search_call_completed"
	StreamEventWebSearchCallInProgress StreamEventType = "web_search_call_in_progress"
	StreamEventWebSearchCallSearching  StreamEventType = "web_search_call_searching"

	// Image Generation 事件（兼容 OpenAI Responses）
	StreamEventImageGenerationCallCompleted    StreamEventType = "image_generation_call_completed"
	StreamEventImageGenerationCallInProgress   StreamEventType = "image_generation_call_in_progress"
	StreamEventImageGenerationCallGenerating   StreamEventType = "image_generation_call_generating"
	StreamEventImageGenerationCallPartialImage StreamEventType = "image_generation_call_partial_image"

	// 错误事件
	StreamEventError StreamEventType = "error"

	// Ping 事件（兼容 Anthropic）
	StreamEventPing StreamEventType = "ping"
)

// StreamEventSource 表示事件来源。
type StreamEventSource string

const (
	StreamSourceAnthropic      StreamEventSource = "anthropic"
	StreamSourceGemini         StreamEventSource = "gemini"
	StreamSourceOpenAIChat     StreamEventSource = "openai.chat"
	StreamSourceOpenAIResponse StreamEventSource = "openai.responses"
)

// StreamEventContract 表示中间流式事件。
// 通过 Source + Extensions 保留供应商特有字段，避免往返转换丢失信息。
type StreamEventContract struct {
	Type            StreamEventType `json:"type"`
	EventID         string          `json:"event_id,omitempty"`
	ResponseID      string          `json:"response_id,omitempty"`
	MessageID       string          `json:"message_id,omitempty"`
	ItemID          string          `json:"item_id,omitempty"`
	SequenceNumber  int             `json:"sequence_number,omitempty"`
	OutputIndex     int             `json:"output_index,omitempty"`
	ContentIndex    int             `json:"content_index,omitempty"`
	AnnotationIndex int             `json:"annotation_index,omitempty"`
	CreatedAt       int64           `json:"created_at,omitempty"`
	Model           string          `json:"model,omitempty"`

	Message *StreamMessagePayload `json:"message,omitempty"`
	Content *StreamContentPayload `json:"content,omitempty"`
	Delta   *StreamDeltaPayload   `json:"delta,omitempty"`
	Usage   *StreamUsagePayload   `json:"usage,omitempty"`
	Error   *StreamErrorPayload   `json:"error,omitempty"`

	Source     StreamEventSource      `json:"source"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// StreamMessagePayload 表示消息实体（适用于 message_start 与 output_item 事件）。
type StreamMessagePayload struct {
	Role        string              `json:"role"`
	ContentText *string             `json:"content_text,omitempty"`
	Parts       []StreamContentPart `json:"parts,omitempty"`
	ToolCalls   []StreamToolCall    `json:"tool_calls,omitempty"`
}

// StreamContentPart 表示结构化内容分片。
type StreamContentPart struct {
	Type        string                 `json:"type"`
	Text        string                 `json:"text,omitempty"`
	Annotations []interface{}          `json:"annotations,omitempty"`
	Raw         map[string]interface{} `json:"raw,omitempty"`
}

// StreamContentPayload 表示内容块或分片。
// Kind 建议取值：text、tool_use、output_text、refusal、reasoning_text、reasoning_summary_text、audio、image、call_code、other。
type StreamContentPayload struct {
	Kind        string                 `json:"kind"`
	Text        *string                `json:"text,omitempty"`
	Tool        *StreamToolCall        `json:"tool,omitempty"`
	Annotations []interface{}          `json:"annotations,omitempty"`
	Raw         map[string]interface{} `json:"raw,omitempty"`
}

// StreamDeltaPayload 表示增量内容。
// DeltaType 建议取值：text_delta、input_json_delta、thinking_delta、signature_delta、citations_delta、audio_delta、other。
type StreamDeltaPayload struct {
	DeltaType   string                 `json:"delta_type"`
	Text        *string                `json:"text,omitempty"`
	PartialJSON *string                `json:"partial_json,omitempty"`
	Thinking    *string                `json:"thinking,omitempty"`
	Signature   *string                `json:"signature,omitempty"`
	Citation    interface{}            `json:"citation,omitempty"`
	Raw         map[string]interface{} `json:"raw,omitempty"`
}

// StreamUsagePayload 表示使用量统计。
type StreamUsagePayload struct {
	InputTokens  *int                   `json:"input_tokens,omitempty"`
	OutputTokens *int                   `json:"output_tokens,omitempty"`
	TotalTokens  *int                   `json:"total_tokens,omitempty"`
	Raw          map[string]interface{} `json:"raw,omitempty"`
}

// StreamErrorPayload 表示错误信息。
type StreamErrorPayload struct {
	Message string                 `json:"message"`
	Type    string                 `json:"type,omitempty"`
	Code    string                 `json:"code,omitempty"`
	Param   string                 `json:"param,omitempty"`
	Raw     map[string]interface{} `json:"raw,omitempty"`
}

// StreamToolCall 表示工具调用。
type StreamToolCall struct {
	ID        string                 `json:"id,omitempty"`
	Type      string                 `json:"type,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Arguments string                 `json:"arguments,omitempty"`
	Raw       map[string]interface{} `json:"raw,omitempty"`
}
