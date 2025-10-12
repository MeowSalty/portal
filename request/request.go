package request

import (
	"github.com/MeowSalty/portal/request/adapter"
)

// Request 处理请求的结构体
//
// 该结构体封装了处理各种 AI 请求的通用逻辑和上下文信息
type Request struct {
	adapter *adapter.Adapter
	repo    RequestLogRepository
}

// New 创建一个新的请求处理器
func New(repo RequestLogRepository) *Request {
	p := &Request{
		repo: repo,
	}

	return p
}
