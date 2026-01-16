package responses

// ResponsesStreamEvent 表示 Responses 流式事件
// 通过 Type 区分不同事件。
type ResponsesStreamEvent struct {
	Type            string          `json:"type"`
	EventID         string          `json:"event_id,omitempty"`
	ResponseID      string          `json:"response_id,omitempty"`
	SequenceNumber  int             `json:"sequence_number,omitempty"`
	ItemID          string          `json:"item_id,omitempty"`
	OutputIndex     int             `json:"output_index,omitempty"`
	ContentIndex    int             `json:"content_index,omitempty"`
	AnnotationIndex int             `json:"annotation_index,omitempty"`
	Delta           string          `json:"delta,omitempty"`
	Text            string          `json:"text,omitempty"`
	Refusal         string          `json:"refusal,omitempty"`
	Arguments       string          `json:"arguments,omitempty"`
	Name            string          `json:"name,omitempty"`
	CallID          string          `json:"call_id,omitempty"`
	Annotation      interface{}     `json:"annotation,omitempty"`
	Logprobs        []interface{}   `json:"logprobs,omitempty"`
	Item            *OutputItem     `json:"item,omitempty"`
	Part            *OutputPart     `json:"part,omitempty"`
	Response        *Response       `json:"response,omitempty"`
	Error           *ResponsesError `json:"error,omitempty"`
	Code            string          `json:"code,omitempty"`
	Message         string          `json:"message,omitempty"`
	Param           string          `json:"param,omitempty"`
}

// ResponsesError 表示 Responses 流式错误
type ResponsesError struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
	Code    string `json:"code,omitempty"`
	Param   string `json:"param,omitempty"`
}
