package types

// ErrorResponse 错误响应
type ErrorResponse struct {
	Type  string `json:"type"`  // "error"
	Error Error  `json:"error"` // 错误详情
}

// Error 错误详情
type Error struct {
	Type    string `json:"type"`    // 错误类型
	Message string `json:"message"` // 错误消息
}
