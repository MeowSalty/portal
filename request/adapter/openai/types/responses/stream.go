package responses

// ResponsesStreamEvent 表示 Responses 流式事件
// 通过 Type 区分不同事件。
type ResponsesStreamEvent struct {
	Type            string          `json:"type"`
	Delta           string          `json:"delta,omitempty"`
	Text            string          `json:"text,omitempty"`
	ItemID          string          `json:"item_id,omitempty"`
	OutputIndex     int             `json:"output_index"`
	ContentIndex    int             `json:"content_index"`
	AnnotationIndex int             `json:"annotation_index"`
	Annotation      interface{}     `json:"annotation,omitempty"`
	SequenceNumber  int             `json:"sequence_number,omitempty"`
	Response        *Response       `json:"response,omitempty"`
	ResponseID      string          `json:"response_id,omitempty"`
	Error           *ResponsesError `json:"error,omitempty"`
	Item            *OutputItem     `json:"item,omitempty"`
	Part            *OutputPart     `json:"part,omitempty"`
	Logprobs        []interface{}   `json:"logprobs,omitempty"`
	Code            string          `json:"code,omitempty"`
	Message         string          `json:"message,omitempty"`
	Param           string          `json:"param,omitempty"`
}

// ResponsesError 表示 Responses 流式错误
type ResponsesError struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
}
