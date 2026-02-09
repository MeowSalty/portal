package portal

import (
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/middleware"
	"github.com/MeowSalty/portal/request"
	"github.com/MeowSalty/portal/routing"
	"github.com/MeowSalty/portal/routing/health"
	"github.com/MeowSalty/portal/session"
)

// Portal 是门户结构体，负责协调各个组件
type Portal struct {
	session    *session.Session
	routing    *routing.Routing
	request    *request.Request
	logger     logger.Logger
	middleware *middleware.Chain
}

// Config 是 Portal 的配置结构体
type Config struct {
	PlatformRepo  routing.PlatformRepository
	ModelRepo     routing.ModelRepository
	KeyRepo       routing.KeyRepository
	HealthStorage health.Storage // 健康状态存储
	LogRepo       request.RequestLogRepository
	Logger        logger.Logger           // 可选的日志记录器，如果为 nil 则使用默认的空操作日志记录器
	Middlewares   []middleware.Middleware // 可选的中间件列表
}
