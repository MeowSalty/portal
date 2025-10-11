// Package health 提供了资源健康状态管理功能
package health

import (
	"time"
)

// Filter 定义健康状态过滤器接口
type Filter interface {
	// IsHealthy 检查单个资源是否健康
	IsHealthy(resourceType ResourceType, resourceID uint, now time.Time) bool
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

// IsHealthy 检查单个资源是否健康
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//   - now: 当前时间
//
// 返回值：
//   - bool: 如果资源健康返回 true，否则返回 false
func (f *filter) IsHealthy(resourceType ResourceType, resourceID uint, now time.Time) bool {
	status, exists := f.cache.Get(resourceType, resourceID)
	if !exists {
		// 如果缓存中不存在，认为是未知状态（可以尝试）
		return true
	}

	// 检查状态是否为可用或未知（未知状态表示尚未进行健康检查）
	if status.Status == HealthStatusAvailable || status.Status == HealthStatusUnknown {
		return true
	}

	// 对于警告状态，检查是否已到下次可用时间
	if status.Status == HealthStatusWarning && status.NextAvailableAt != nil {
		if now.After(*status.NextAvailableAt) {
			// 退避时间已过，可以认为是健康的
			return true
		}
	}

	// 不可用状态或者还在退避期内
	return false
}
