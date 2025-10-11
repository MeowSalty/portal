package request

import (
	"time"
)

// streamContext 封装流处理的上下文信息
type streamContext struct {
	requestLog    *RequestLog
	requestStart  time.Time
	firstByteTime *time.Time
	streamError   error
}

// errorResponse 错误响应的可重用结构
type errorResponse struct {
	code    int
	message string
}
