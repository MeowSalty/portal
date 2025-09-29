package processor

import (
	"context"
	"log/slog"

	"github.com/MeowSalty/portal/health"
	"github.com/MeowSalty/portal/stats"
	"github.com/MeowSalty/portal/types"
)

// NewRequestProcessor 创建一个新的请求处理器
func NewRequestProcessor(
	healthManager *health.Manager,
	selector types.ChannelSelector,
	adapters map[string]types.Adapter,
	logger *slog.Logger,
	statsManager *stats.Manager,
	buildChannels func(context.Context, []*types.Model) ([]*types.Channel, error),
	FindModelsByName func(context.Context, string) ([]*types.Model, error),
) *RequestProcessor {
	p := &RequestProcessor{
		HealthManager:    healthManager,
		Selector:         selector,
		Adapters:         adapters,
		Logger:           logger,
		StatsManager:     statsManager,
		BuildChannels:    buildChannels,
		FindModelsByName: FindModelsByName,
	}

	// 初始化错误响应池
	p.errorResponsePool.New = func() interface{} {
		return &types.Response{
			Choices: []types.Choice{{Error: &types.ErrorResponse{}}},
		}
	}

	return p
}
