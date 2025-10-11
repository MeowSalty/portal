package routing

import (
	"context"
	"fmt"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/routing/health"
	"github.com/MeowSalty/portal/routing/selector"
)

// Routing 管理通道的获取和状态
type Routing struct {
	selector      selector.Selector
	platformRepo  PlatformRepository
	modelRepo     ModelRepository
	keyRepo       KeyRepository
	healthService *health.Service
}

// Config 通道服务配置
type Config struct {
	Selector     selector.Selector
	PlatformRepo PlatformRepository
	ModelRepo    ModelRepository
	KeyRepo      KeyRepository
	HealthRepo   health.HealthRepository // 健康状态仓库
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
	if cfg.HealthRepo == nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "健康状态仓库不能为空")
	}

	// 创建健康服务
	healthConfig := health.Config{
		Repo: cfg.HealthRepo,
	}
	healthService, err := health.New(ctx, healthConfig)
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
		return nil, errors.New(errors.ErrCodeInvalidArgument, "模型名称不能为空")
	}

	// 1. 通过模型名称或别名查找模型
	models, err := r.modelRepo.FindModelsByNameOrAlias(ctx, modelName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "查询模型失败", err)
	}

	if len(models) == 0 {
		// 返回错误
		return nil, errors.New(errors.ErrCodeNotFound, "未找到模型")
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
		return nil, errors.New(errors.ErrCodeNotFound, "没有可用的通道")
	}

	// 5. 使用 selector 选择一个通道
	selectedID, err := r.selector.Select(channelInfos)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "选择通道失败", err)
	}

	// 找到对应的通道
	for i, info := range channelInfos {
		if info.ID == selectedID {
			return availableChannels[i], nil
		}
	}

	// 不应该到达这里
	return nil, errors.New(errors.ErrCodeInternal, "选择的通道未找到")
}

// buildChannelsForModel 为指定模型构建所有可能的通道
func (r *Routing) buildChannelsForModel(ctx context.Context, model Model) ([]*Channel, error) {
	// 获取平台信息
	platform, err := r.platformRepo.GetPlatformByID(ctx, model.PlatformID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "获取平台信息失败", err)
	}

	// 获取平台的所有密钥
	apiKeys, err := r.keyRepo.GetAllAPIKeysByPlatformID(ctx, platform.ID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "获取 API 密钥失败", err)
	}

	if len(apiKeys) == 0 {
		return nil, errors.New(errors.ErrCodeNotFound, "平台没有配置 API 密钥")
	}

	// 为每个密钥创建一个通道
	var channels []*Channel
	for _, apiKey := range apiKeys {
		ch := &Channel{
			PlatformID:    platform.ID,
			ModelID:       model.ID,
			APIKeyID:      apiKey.ID,
			PlatformType:  platform.Format,
			APIEndpoint:   platform.BaseURL,
			ModelName:     model.Name,
			APIKey:        apiKey.Value,
			healthService: r.healthService,
		}
		channels = append(channels, ch)
	}

	return channels, nil
}

// Shutdown 关闭服务
func (r *Routing) Shutdown() {
	if r.healthService != nil {
		r.healthService.Shutdown()
	}
}
