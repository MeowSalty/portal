package responses

// ResponsesStreamEvent 表示 Responses 流式事件
// 通过 Type 区分不同事件。
type ResponsesStreamEvent struct {
	Type       string          `json:"type"`
	Delta      string          `json:"delta,omitempty"`
	Response   *Response       `json:"response,omitempty"`
	ResponseID string          `json:"response_id,omitempty"`
	Error      *ResponsesError `json:"error,omitempty"`
}

// ResponsesError 表示 Responses 流式错误
type ResponsesError struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
}
