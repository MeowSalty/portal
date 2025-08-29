package selector

import (
	"context"
	"errors"
	"sort"

	"github.com/MeowSalty/portal/health"
	"github.com/MeowSalty/portal/types"
)

// LeastRecentlyUsedSelector 实现了一个选择器，选择最近最少使用的通道
// 它会优先选择健康状态为 Unknown 的通道，其次选择最久未成功使用的通道
type LeastRecentlyUsedSelector struct {
	healthManager *health.Manager // 健康状态管理器
}

// NewLeastRecentlyUsedSelector 创建一个新的 LRU 选择器实例
// 参数：
//
//	healthManager - 用于检查通道健康状态的管理器
func NewLeastRecentlyUsedSelector(healthManager *health.Manager) *LeastRecentlyUsedSelector {
	return &LeastRecentlyUsedSelector{
		healthManager: healthManager,
	}
}

// Select 根据 LRU 策略选择一个通道
// 优先选择状态未知的通道，其次选择最久未成功使用的通道
func (s *LeastRecentlyUsedSelector) Select(_ context.Context, channels []*types.Channel) (*types.Channel, error) {
	if len(channels) == 0 {
		return nil, errors.New("通道列表不能为空")
	}

	unknownChannels := make([]*types.Channel, 0)
	otherChannels := make([]*types.Channel, 0)

	for _, ch := range channels {
		status := s.healthManager.GetStatus(types.ResourceTypeAPIKey, ch.APIKey.ID)
		if status.Status == types.HealthStatusUnknown {
			unknownChannels = append(unknownChannels, ch)
		} else {
			otherChannels = append(otherChannels, ch)
		}
	}

	// 优先选择状态未知的通道。如果存在多个，任选其一即可。
	if len(unknownChannels) > 0 {
		// 优先返回第一个未知状态通道
		return unknownChannels[0], nil
	}

	// 对非未知状态的通道按最后成功时间排序
	// 排序规则：
	// 1. 从未成功过的通道排在最前面
	// 2. 最后成功时间较早的通道排在前面
	sort.Slice(otherChannels, func(i, j int) bool {
		statusI := s.healthManager.GetStatus(types.ResourceTypeAPIKey, otherChannels[i].APIKey.ID)
		statusJ := s.healthManager.GetStatus(types.ResourceTypeAPIKey, otherChannels[j].APIKey.ID)

		// 从未成功过的通道被视为"最久未使用"
		if statusI.LastSuccessAt == nil {
			return true
		}
		if statusJ.LastSuccessAt == nil {
			return false
		}

		// 比较两个通道的最后成功时间
		return statusI.LastSuccessAt.Before(*statusJ.LastSuccessAt)
	})

	return otherChannels[0], nil
}
