package types

// ErrorResponse 表示错误响应。
type ErrorResponse struct {
	Error ErrorDetail `json:"error"` // 错误详情
}

// ErrorDetail 表示错误详情。
type ErrorDetail struct {
	Code    int                      `json:"code"`              // 错误代码
	Message string                   `json:"message"`           // 错误消息
	Status  string                   `json:"status"`            // 状态
	Details []map[string]interface{} `json:"details,omitempty"` // 详细信息
}
