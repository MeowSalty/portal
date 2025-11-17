package request

import (
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/request/adapter"
)

// Request 处理请求的结构体
//
// 该结构体封装了处理各种 AI 请求的通用逻辑和上下文信息
type Request struct {
	adapter *adapter.Adapter
	repo    RequestLogRepository
	logger  logger.Logger
}

// New 创建一个新的请求处理器
//
// 参数：
//   - repo: 请求日志仓库
//   - log: 日志记录器，如果为 nil 则使用默认的 nopLogger
func New(repo RequestLogRepository, log logger.Logger) *Request {
	// 如果未提供 logger，使用 nopLogger
	if log == nil {
		log = logger.NewNopLogger()
	}

	p := &Request{
		repo:   repo,
		logger: log,
	}

	return p
}
