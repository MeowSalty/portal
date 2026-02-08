package responses

// ResponseOutputTextDeltaEvent response.output_text.delta 事件。
type ResponseOutputTextDeltaEvent struct {
	Type           StreamEventType   `json:"type"`
	ItemID         string            `json:"item_id"`
	OutputIndex    int               `json:"output_index"`
	ContentIndex   int               `json:"content_index"`
	Delta          string            `json:"delta"`
	SequenceNumber int               `json:"sequence_number"`
	Logprobs       []ResponseLogProb `json:"logprobs"`
	Obfuscation    *string           `json:"obfuscation,omitempty"`
}

// ResponseOutputTextDoneEvent response.output_text.done 事件。
type ResponseOutputTextDoneEvent struct {
	Type           StreamEventType   `json:"type"`
	ItemID         string            `json:"item_id"`
	OutputIndex    int               `json:"output_index"`
	ContentIndex   int               `json:"content_index"`
	Text           string            `json:"text"`
	SequenceNumber int               `json:"sequence_number"`
	Logprobs       []ResponseLogProb `json:"logprobs"`
}

// ResponseOutputTextAnnotationAddedEvent response.output_text.annotation.added 事件。
type ResponseOutputTextAnnotationAddedEvent struct {
	Type            StreamEventType `json:"type"`
	ItemID          string          `json:"item_id"`
	OutputIndex     int             `json:"output_index"`
	ContentIndex    int             `json:"content_index"`
	AnnotationIndex int             `json:"annotation_index"`
	Annotation      Annotation      `json:"annotation"`
	SequenceNumber  int             `json:"sequence_number"`
}

// ResponseRefusalDeltaEvent response.refusal.delta 事件。
type ResponseRefusalDeltaEvent struct {
	Type           StreamEventType `json:"type"`
	ItemID         string          `json:"item_id"`
	OutputIndex    int             `json:"output_index"`
	ContentIndex   int             `json:"content_index"`
	Delta          string          `json:"delta"`
	SequenceNumber int             `json:"sequence_number"`
	Obfuscation    *string         `json:"obfuscation,omitempty"`
}

// ResponseRefusalDoneEvent response.refusal.done 事件。
type ResponseRefusalDoneEvent struct {
	Type           StreamEventType `json:"type"`
	ItemID         string          `json:"item_id"`
	OutputIndex    int             `json:"output_index"`
	ContentIndex   int             `json:"content_index"`
	Refusal        string          `json:"refusal"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseReasoningTextDeltaEvent response.reasoning_text.delta 事件。
type ResponseReasoningTextDeltaEvent struct {
	Type           StreamEventType `json:"type"`
	ItemID         string          `json:"item_id"`
	OutputIndex    int             `json:"output_index"`
	ContentIndex   int             `json:"content_index"`
	Delta          string          `json:"delta"`
	SequenceNumber int             `json:"sequence_number"`
	Obfuscation    *string         `json:"obfuscation,omitempty"`
}

// ResponseReasoningTextDoneEvent response.reasoning_text.done 事件。
type ResponseReasoningTextDoneEvent struct {
	Type           StreamEventType `json:"type"`
	ItemID         string          `json:"item_id"`
	OutputIndex    int             `json:"output_index"`
	ContentIndex   int             `json:"content_index"`
	Text           string          `json:"text"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseReasoningSummaryPartAddedEvent response.reasoning_summary_part.added 事件。
type ResponseReasoningSummaryPartAddedEvent struct {
	Type           StreamEventType   `json:"type"`
	ItemID         string            `json:"item_id"`
	OutputIndex    int               `json:"output_index"`
	SummaryIndex   int               `json:"summary_index"`
	Part           OutputSummaryPart `json:"part"`
	SequenceNumber int               `json:"sequence_number"`
}

// ResponseReasoningSummaryPartDoneEvent response.reasoning_summary_part.done 事件。
type ResponseReasoningSummaryPartDoneEvent struct {
	Type           StreamEventType   `json:"type"`
	ItemID         string            `json:"item_id"`
	OutputIndex    int               `json:"output_index"`
	SummaryIndex   int               `json:"summary_index"`
	Part           OutputSummaryPart `json:"part"`
	SequenceNumber int               `json:"sequence_number"`
}

// ResponseReasoningSummaryTextDeltaEvent response.reasoning_summary_text.delta 事件。
type ResponseReasoningSummaryTextDeltaEvent struct {
	Type           StreamEventType `json:"type"`
	ItemID         string          `json:"item_id"`
	OutputIndex    int             `json:"output_index"`
	SummaryIndex   int             `json:"summary_index"`
	Delta          string          `json:"delta"`
	SequenceNumber int             `json:"sequence_number"`
	Obfuscation    *string         `json:"obfuscation,omitempty"`
}

// ResponseReasoningSummaryTextDoneEvent response.reasoning_summary_text.done 事件。
type ResponseReasoningSummaryTextDoneEvent struct {
	Type           StreamEventType `json:"type"`
	ItemID         string          `json:"item_id"`
	OutputIndex    int             `json:"output_index"`
	SummaryIndex   int             `json:"summary_index"`
	Text           string          `json:"text"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseFunctionCallArgumentsDeltaEvent response.function_call_arguments.delta 事件。
type ResponseFunctionCallArgumentsDeltaEvent struct {
	Type           StreamEventType `json:"type"`
	ItemID         string          `json:"item_id"`
	OutputIndex    int             `json:"output_index"`
	Delta          string          `json:"delta"`
	SequenceNumber int             `json:"sequence_number"`
	Obfuscation    *string         `json:"obfuscation,omitempty"`
}

// ResponseFunctionCallArgumentsDoneEvent response.function_call_arguments.done 事件。
type ResponseFunctionCallArgumentsDoneEvent struct {
	Type           StreamEventType `json:"type"`
	ItemID         string          `json:"item_id"`
	Name           string          `json:"name"`
	OutputIndex    int             `json:"output_index"`
	Arguments      string          `json:"arguments"`
	SequenceNumber int             `json:"sequence_number"`
}
