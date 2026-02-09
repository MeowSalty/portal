package portal

import (
	"context"
	"time"

	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/middleware"
	"github.com/MeowSalty/portal/request"
	"github.com/MeowSalty/portal/routing"
	"github.com/MeowSalty/portal/routing/selector"
	"github.com/MeowSalty/portal/session"
)

// New 创建一个新的 Portal 实例
func New(cfg Config) (*Portal, error) {
	// 如果未提供日志记录器，使用默认的空操作日志记录器
	log := cfg.Logger
	if log == nil {
		log = logger.NewNopLogger()
	}
	// 初始化全局默认日志记录器
	logger.SetDefault(log)

	routing, err := routing.New(context.TODO(), routing.Config{
		PlatformRepo:  cfg.PlatformRepo,
		ModelRepo:     cfg.ModelRepo,
		KeyRepo:       cfg.KeyRepo,
		HealthStorage: cfg.HealthStorage,
		Selector:      selector.NewLRUSelector(),
	})
	if err != nil {
		return nil, err
	}
	portal := &Portal{
		session:    session.New(),
		routing:    routing,
		request:    request.New(cfg.LogRepo, log.WithGroup("request")),
		logger:     log,
		middleware: middleware.NewChain(cfg.Middlewares...),
	}
	return portal, nil
}

// Close 关闭 Portal 实例，释放资源
func (p *Portal) Close(timeout time.Duration) error {
	return p.session.Shutdown(timeout)
}
