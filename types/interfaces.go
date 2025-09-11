package types

import (
	"context"
)

// DataRepository 定义了网关所需的数据访问方法
type DataRepository interface {
	FindModelsByName(ctx context.Context, name string) ([]*Model, error)
	GetPlatformByID(ctx context.Context, id uint) (*Platform, error)
	GetAllAPIKeys(ctx context.Context, platformID uint) ([]*APIKey, error)
	GetAllHealthStatus(ctx context.Context) ([]*Health, error)
	BatchUpdateHealthStatus(ctx context.Context, statuses []*Health) error

	// 统计相关方法
	SaveRequestStat(ctx context.Context, stat *RequestStat) error
}

// ChannelSelector 定义了从可用通道中选择一个通道的策略
type ChannelSelector interface {
	Select(ctx context.Context, channels []*Channel) (*Channel, error)
}

// Adapter 定义了与特定 AI 平台 API 交互的接口
type Adapter interface {
	ChatCompletion(ctx context.Context, request *Request, channel *Channel) (*Response, error)
	ChatCompletionStream(ctx context.Context, request *Request, channel *Channel) (<-chan *Response, error)
}
