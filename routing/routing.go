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

// GetChannel 根据模型名称获取一个可用的通道（使用默认端点）
func (r *Routing) GetChannel(ctx context.Context, modelName string) (*Channel, error) {
	if modelName == "" {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "模型名称不能为空").WithHTTPStatus(fasthttp.StatusBadRequest)
	}

	// 通过模型名称查找，返回带有平台和默认端点的完整信息
	modelsWithEndpoint, err := r.modelRepo.FindModelsWithDefaultEndpoint(ctx, modelName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "查询模型失败", err).WithHTTPStatus(fasthttp.StatusInternalServerError)
	}

	if len(modelsWithEndpoint) == 0 {
		return nil, errors.New(errors.ErrCodeNotFound, "未找到模型或平台未配置默认端点").WithHTTPStatus(fasthttp.StatusNotFound)
	}

	return r.selectChannelFromModelsWithEndpoint(modelsWithEndpoint)
}

// GetChannelByProvider 根据模型名称、端点类型和变体获取一个可用的通道
func (r *Routing) GetChannelByProvider(ctx context.Context, modelName, endpointType, endpointVariant string) (*Channel, error) {
	// 参数校验
	if modelName == "" {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "模型名称不能为空").WithHTTPStatus(fasthttp.StatusBadRequest)
	}
	if endpointType == "" {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "端点类型不能为空").WithHTTPStatus(fasthttp.StatusBadRequest)
	}
	if endpointVariant == "" {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "端点变体不能为空").WithHTTPStatus(fasthttp.StatusBadRequest)
	}

	// 通过模型名称 + 端点类型 + 变体查找
	modelsWithEndpoint, err := r.modelRepo.FindModelsWithEndpoint(ctx, modelName, endpointType, endpointVariant)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "查询模型失败", err).WithHTTPStatus(fasthttp.StatusInternalServerError)
	}

	if len(modelsWithEndpoint) == 0 {
		return nil, errors.New(errors.ErrCodeEndpointNotFound, "未找到匹配的端点").WithHTTPStatus(fasthttp.StatusNotFound)
	}

	return r.selectChannelFromModelsWithEndpoint(modelsWithEndpoint)
}

// selectChannelFromModelsWithEndpoint 从模型列表中选择一个可用的通道
func (r *Routing) selectChannelFromModelsWithEndpoint(modelsWithEndpoint []ModelWithEndpoint) (*Channel, error) {
	// 为每个模型构建通道
	var availableChannels []*Channel
	var channelInfos []selector.ChannelInfo

	for _, mwe := range modelsWithEndpoint {
		channels := r.buildChannelsForModelWithEndpoint(mwe)

		// 使用 health 验证通道是否可用
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

	// 如果没有可用通道，返回错误
	if len(availableChannels) == 0 {
		return nil, errors.New(errors.ErrCodeResourceExhausted, "没有可用的通道").WithHTTPStatus(fasthttp.StatusServiceUnavailable)
	}

	// 使用互斥锁保护通道选择和时间更新操作
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

// buildChannelsForModelWithEndpoint 为指定模型构建所有可能的通道
// 从 ModelWithEndpoint 中获取所有需要的信息
func (r *Routing) buildChannelsForModelWithEndpoint(mwe ModelWithEndpoint) []*Channel {
	model := mwe.Model
	platform := mwe.Platform
	endpoint := mwe.Endpoint

	// 合并 CustomHeaders（Platform 级别 + Endpoint 级别，Endpoint 覆盖同名）
	customHeaders := mergeCustomHeaders(platform.CustomHeaders, endpoint.CustomHeaders)

	// 检查模型是否有关联的密钥
	if len(model.APIKeys) == 0 {
		return nil
	}

	// 为每个 APIKey 创建一个 Channel
	var channels []*Channel
	for _, key := range model.APIKeys {
		channel := &Channel{
			PlatformID:        platform.ID,
			ModelID:           model.ID,
			APIKeyID:          key.ID,
			Provider:          endpoint.EndpointType, // 从 Endpoint 获取
			BaseURL:           platform.BaseURL,      // 从 Platform 获取
			ModelName:         model.Name,
			APIKey:            key.Value,
			APIVariant:        endpoint.EndpointVariant, // 从 Endpoint 获取
			APIEndpointConfig: endpoint.Path,            // 从 Endpoint 获取
			CustomHeaders:     customHeaders,
			healthService:     r.healthService,
		}
		channels = append(channels, channel)
	}
	return channels
}

// mergeCustomHeaders 合并 Platform 和 Endpoint 的 CustomHeaders
// Endpoint 的头部优先级更高，会覆盖 Platform 同名头部
func mergeCustomHeaders(platform, endpoint map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range platform {
		result[k] = v
	}
	for k, v := range endpoint {
		result[k] = v
	}
	return result
}
