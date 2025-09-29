package processor

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/MeowSalty/portal/health"
	"github.com/MeowSalty/portal/stats"
	"github.com/MeowSalty/portal/types"
)

// RequestProcessor 处理请求的结构体
//
// 该结构体封装了处理各种 AI 请求的通用逻辑和上下文信息
type RequestProcessor struct {
	HealthManager    *health.Manager
	Selector         types.ChannelSelector
	Adapters         map[string]types.Adapter
	Logger           *slog.Logger
	StatsManager     *stats.Manager
	BuildChannels    func(context.Context, []*types.Model) ([]*types.Channel, error)
	FindModelsByName func(context.Context, string) ([]*types.Model, error)

	// 性能优化：预分配的错误响应池
	errorResponsePool sync.Pool
}

// streamContext 封装流处理的上下文信息
type streamContext struct {
	request         *types.Request
	selectedChannel *types.Channel
	statOptions     *types.RequestStat
	requestStart    time.Time
	firstByteTime   *time.Time
	streamErr       error
}

// retryContext 封装重试相关的上下文
type retryContext struct {
	allChannels  []*types.Channel
	startTime    time.Time
	channelIndex map[string]int // 优化通道移除性能
}

// errorResponse 错误响应的可重用结构
type errorResponse struct {
	code    int
	message string
}
