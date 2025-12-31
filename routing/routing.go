package routing

import (
	"context"
	"fmt"
	"sync"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/routing/health"
	"github.com/MeowSalty/portal/routing/selector"
	"github.com/valyala/fasthttp"
)

// Routing 管理通道的获取和状态
type Routing struct {
	selector      selector.Selector
	platformRepo  PlatformRepository
	modelRepo     ModelRepository
	keyRepo       KeyRepository
	healthService *health.Service
	mu            sync.Mutex // 保护并发通道选择的互斥锁
}

// Config 通道服务配置
type Config struct {
	Selector      selector.Selector
	PlatformRepo  PlatformRepository
	ModelRepo     ModelRepository
	KeyRepo       KeyRepository
	HealthStorage health.Storage // 健康状态存储
}

// New 创建一个新的通道服务
func New(ctx context.Context, cfg Config) (*Routing, error) {
	// 验证配置
	if cfg.Selector == nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "选择器不能为空")
	}
	if cfg.PlatformRepo == nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "平台仓库不能为空")
	}
	if cfg.ModelRepo == nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "模型仓库不能为空")
	}
	if cfg.KeyRepo == nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "密钥仓库不能为空")
	}
	if cfg.HealthStorage == nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "健康状态存储不能为空")
	}

	// 创建健康服务
	healthConfig := health.Config{
		Storage: cfg.HealthStorage,
	}
	healthService, err := health.New(healthConfig)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "初始化健康服务失败", err)
	}

	return &Routing{
		selector:      cfg.Selector,
		platformRepo:  cfg.PlatformRepo,
		modelRepo:     cfg.ModelRepo,
		keyRepo:       cfg.KeyRepo,
		healthService: healthService,
	}, nil
}

// GetChannel 根据模型名称获取一个可用的通道
func (r *Routing) GetChannel(ctx context.Context, modelName string) (*Channel, error) {
	if modelName == "" {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "模型名称不能为空").WithHTTPStatus(fasthttp.StatusBadRequest)
	}

	// 1. 通过模型名称或别名查找模型
	models, err := r.modelRepo.FindModelsByNameOrAlias(ctx, modelName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "查询模型失败", err).WithHTTPStatus(fasthttp.StatusInternalServerError)
	}

	if len(models) == 0 {
		// 返回错误
		return nil, errors.New(errors.ErrCodeNotFound, "未找到模型").WithHTTPStatus(fasthttp.StatusNotFound)
	}

	// 2. 为每个模型构建通道
	var availableChannels []*Channel
	var channelInfos []selector.ChannelInfo

	for _, model := range models {
		channels, err := r.buildChannelsForModel(ctx, model)
		if err != nil {
			// 记录错误但继续处理其他模型
			continue
		}

		// 3. 使用 health 验证通道是否可用
		for _, ch := range channels {
			result := r.healthService.CheckChannelHealth(ch.PlatformID, ch.ModelID, ch.APIKeyID)

			switch result.Status {
			case health.ChannelStatusAvailable:
				availableChannels = append(availableChannels, ch)
				channelInfos = append(channelInfos, selector.ChannelInfo{
					ID:           fmt.Sprintf("%d-%d-%d", ch.PlatformID, ch.ModelID, ch.APIKeyID),
					LastUsedTime: result.LastCheckAt,
				})
			case health.ChannelStatusUnknown:
				// 对于未知状态的通道，直接返回它
				return ch, nil
			}
			// 不可用的通道直接跳过
		}
	}

	// 4. 如果没有可用通道，返回错误
	if len(availableChannels) == 0 {
		return nil, errors.New(errors.ErrCodeResourceExhausted, "没有可用的通道").WithHTTPStatus(fasthttp.StatusServiceUnavailable)
	}

	// 5. 使用互斥锁保护通道选择和时间更新操作
	// 确保在并发环境下，选择通道和更新使用时间是原子操作
	r.mu.Lock()
	selectedID, err := r.selector.Select(channelInfos)
	if err != nil {
		r.mu.Unlock()
		return nil, errors.Wrap(errors.ErrCodeInternal, "选择通道失败", err).WithHTTPStatus(fasthttp.StatusInternalServerError)
	}

	// 立即更新选中通道的最后使用时间
	if updateErr := r.healthService.UpdateLastUsed(selectedID); updateErr != nil {
		// 记录错误但不影响通道选择结果
		// TODO: 添加日志记录
		_ = updateErr
	}
	r.mu.Unlock()

	// 找到对应的通道
	for i, info := range channelInfos {
		if info.ID == selectedID {
			return availableChannels[i], nil
		}
	}

	// 不应该到达这里
	return nil, errors.New(errors.ErrCodeInternal, "选择的通道未找到").WithHTTPStatus(fasthttp.StatusInternalServerError)
}

// buildChannelsForModel 为指定模型构建所有可能的通道
func (r *Routing) buildChannelsForModel(ctx context.Context, model Model) ([]*Channel, error) {
	// 获取平台信息
	platform, err := r.platformRepo.GetPlatformByID(ctx, model.PlatformID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "获取平台信息失败", err)
	}

	// 检查模型是否有关联的密钥
	if len(model.APIKeys) == 0 {
		return nil, errors.New(errors.ErrCodeNotFound, "模型没有配置 API 密钥")
	}

	// 为每个密钥创建一个通道
	var channels []*Channel
	for _, apiKey := range model.APIKeys {
		ch := &Channel{
			PlatformID:    platform.ID,
			ModelID:       model.ID,
			APIKeyID:      apiKey.ID,
			PlatformType:  platform.Format,
			APIEndpoint:   platform.BaseURL,
			ModelName:     model.Name,
			APIKey:        apiKey.Value,
			CustomHeaders: platform.CustomHeaders, // 传递平台自定义头部给通道
			healthService: r.healthService,
		}
		channels = append(channels, ch)
	}

	return channels, nil
}
