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
	cache        Cache
	allowProbing bool // 是否允许对 Unavailable 状态的资源进行探测
}

// NewFilter 创建一个新的过滤器实例
//
// 参数：
//   - cache: 健康状态缓存
//   - allowProbing: 是否允许对 Unavailable 状态的资源进行探测
//     如果为 true，当 Unavailable 状态的资源退避时间结束后，允许进行探测尝试
//     如果为 false，Unavailable 状态的资源将永久不可用，需要手动重置
func NewFilter(cache Cache, allowProbing bool) Filter {
	return &filter{
		cache:        cache,
		allowProbing: allowProbing,
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

	switch status.Status {
	case HealthStatusAvailable, HealthStatusUnknown:
		// 可用或未知状态，认为是健康的
		return true

	case HealthStatusWarning:
		// 警告状态，检查是否已到下次可用时间
		if status.NextAvailableAt != nil && now.After(*status.NextAvailableAt) {
			// 退避时间已过，可以认为是健康的
			return true
		}
		return false

	case HealthStatusUnavailable:
		// 不可用状态
		if f.allowProbing && status.NextAvailableAt != nil && now.After(*status.NextAvailableAt) {
			// 允许探测且退避时间已过，可以进行探测尝试
			return true
		}
		// 不允许探测或退避时间未到，认为是不健康的
		return false

	default:
		// 未知的状态类型，认为是不健康的
		return false
	}
}
