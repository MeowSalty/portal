// Package health 提供了资源健康状态管理功能
package health

import (
	"time"

	"github.com/MeowSalty/portal/types"
)

// Filter 定义健康状态过滤器接口
type Filter interface {
	// FilterHealthyChannels 过滤出健康的通道
	FilterHealthyChannels(channels []*types.Channel, now time.Time) []*types.Channel
	// IsHealthy 检查单个资源是否健康
	IsHealthy(resourceType types.ResourceType, resourceID uint, now time.Time) bool
}

// filter 实现了健康状态过滤器
type filter struct {
	cache Cache
}

// NewFilter 创建一个新的过滤器实例
func NewFilter(cache Cache) Filter {
	return &filter{
		cache: cache,
	}
}

// FilterHealthyChannels 过滤出健康的通道
//
// 该方法会检查通道列表中的每个通道，返回当前可用的通道
// 一个通道要被认为是健康的，其平台、模型和 API 密钥都必须是健康的
//
// 参数：
//   - channels: 通道列表
//   - now: 当前时间
//
// 返回值：
//   - []*types.Channel: 健康的通道列表
func (f *filter) FilterHealthyChannels(channels []*types.Channel, now time.Time) []*types.Channel {
	var healthyChannels []*types.Channel

	for _, channel := range channels {
		// 检查所有组件的健康状态
		if f.isChannelHealthy(channel, now) {
			healthyChannels = append(healthyChannels, channel)
		}
	}

	return healthyChannels
}

// IsHealthy 检查单个资源是否健康
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//   - now: 当前时间
//
// 返回值：
//   - bool: 如果资源健康返回 true，否则返回 false
func (f *filter) IsHealthy(resourceType types.ResourceType, resourceID uint, now time.Time) bool {
	status, exists := f.cache.Get(resourceType, resourceID)
	if !exists {
		// 如果缓存中不存在，认为是未知状态（可以尝试）
		return true
	}

	return f.isStatusHealthy(status, now)
}

// isChannelHealthy 检查通道是否健康
func (f *filter) isChannelHealthy(channel *types.Channel, now time.Time) bool {
	// 检查平台的健康状态
	if !f.IsHealthy(types.ResourceTypePlatform, channel.Platform.ID, now) {
		return false
	}

	// 检查模型的健康状态
	if !f.IsHealthy(types.ResourceTypeModel, channel.Model.ID, now) {
		return false
	}

	// 检查 API 密钥的健康状态
	if !f.IsHealthy(types.ResourceTypeAPIKey, channel.APIKey.ID, now) {
		return false
	}

	return true
}

// isStatusHealthy 检查给定的健康状态是否为健康状态
//
// 参数：
//   - status: 健康状态对象
//   - now: 当前时间
//
// 返回值：
//   - bool: 如果状态健康返回 true，否则返回 false
func (f *filter) isStatusHealthy(status *types.Health, now time.Time) bool {
	// 检查状态是否为可用或未知（未知状态表示尚未进行健康检查）
	if status.Status == types.HealthStatusAvailable || status.Status == types.HealthStatusUnknown {
		return true
	}

	// 对于警告状态，检查是否已到下次可用时间
	if status.Status == types.HealthStatusWarning && status.NextAvailableAt != nil {
		if now.After(*status.NextAvailableAt) {
			// 退避时间已过，可以认为是健康的
			return true
		}
	}

	// 不可用状态或者还在退避期内
	return false
}

// HealthChecker 提供更高级的健康检查功能
type HealthChecker struct {
	filter Filter
}

// NewHealthChecker 创建一个新的健康检查器
func NewHealthChecker(filter Filter) *HealthChecker {
	return &HealthChecker{
		filter: filter,
	}
}

// GetHealthyChannelCount 获取健康通道的数量
//
// 参数：
//   - channels: 通道列表
//   - now: 当前时间
//
// 返回值：
//   - int: 健康通道的数量
func (h *HealthChecker) GetHealthyChannelCount(channels []*types.Channel, now time.Time) int {
	healthyChannels := h.filter.FilterHealthyChannels(channels, now)
	return len(healthyChannels)
}

// GetHealthyResourceIDs 获取指定类型的所有健康资源 ID
//
// 参数：
//   - resourceType: 资源类型
//   - resourceIDs: 资源 ID 列表
//   - now: 当前时间
//
// 返回值：
//   - []uint: 健康资源 ID 列表
func (h *HealthChecker) GetHealthyResourceIDs(resourceType types.ResourceType, resourceIDs []uint, now time.Time) []uint {
	var healthyIDs []uint

	for _, id := range resourceIDs {
		if h.filter.IsHealthy(resourceType, id, now) {
			healthyIDs = append(healthyIDs, id)
		}
	}

	return healthyIDs
}
