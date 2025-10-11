package request

import (
	"sync"

	"github.com/MeowSalty/portal/request/adapter"
	"github.com/MeowSalty/portal/types"
)

// Request 处理请求的结构体
//
// 该结构体封装了处理各种 AI 请求的通用逻辑和上下文信息
type Request struct {
	adapter *adapter.Adapter
	repo    RequestLogRepository
	// 性能优化：预分配的错误响应池
	errorResponsePool sync.Pool
}

// New 创建一个新的请求处理器
func New(repo RequestLogRepository) *Request {
	p := &Request{
		repo: repo,
	}

	// 初始化错误响应池
	p.errorResponsePool.New = func() interface{} {
		return &types.Response{
			Choices: []types.Choice{{Error: &types.ErrorResponse{}}},
		}
	}

	return p
}
