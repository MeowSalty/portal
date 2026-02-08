package responses

// ResponseCustomToolCallInputDeltaEvent response.custom_tool_call_input.delta 事件。
type ResponseCustomToolCallInputDeltaEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	Delta          string          `json:"delta"`
	SequenceNumber int             `json:"sequence_number"`
	Obfuscation    *string         `json:"obfuscation,omitempty"`
}

// ResponseCustomToolCallInputDoneEvent response.custom_tool_call_input.done 事件。
type ResponseCustomToolCallInputDoneEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	ItemID         string          `json:"item_id"`
	Input          string          `json:"input"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseErrorEvent error 事件。
type ResponseErrorEvent struct {
	Type           StreamEventType `json:"type"`
	Code           *string         `json:"code,omitempty"`
	Message        string          `json:"message"`
	Param          *string         `json:"param,omitempty"`
	SequenceNumber int             `json:"sequence_number"`
}
