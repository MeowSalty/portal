package shared

// HTTPError 表示错误响应。
type HTTPError struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 表示错误详情。
type ErrorDetail struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param,omitempty"`
	Code    *string `json:"code,omitempty"`
}
