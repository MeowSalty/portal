package selector

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/MeowSalty/portal/health"
	"github.com/MeowSalty/portal/types"
)

// LeastRecentlyUsedSelector 实现了最近最少使用选择策略的通道选择器
type LeastRecentlyUsedSelector struct {
	healthManager *health.Manager
	lastUsed      sync.Map // key: channel ID, value: time.Time
}

// NewLeastRecentlyUsedSelector 创建一个新的最近最少使用选择器
func NewLeastRecentlyUsedSelector(healthManager *health.Manager) *LeastRecentlyUsedSelector {
	return &LeastRecentlyUsedSelector{
		healthManager: healthManager,
	}
}

// Select 从提供的通道列表中选择最近最少使用的通道
func (s *LeastRecentlyUsedSelector) Select(_ context.Context, channels []*types.Channel) (*types.Channel, error) {
	if len(channels) == 0 {
		return nil, errors.New("通道列表不能为空")
	}

	var selected *types.Channel
	var oldestTime time.Time
	first := true

	for _, ch := range channels {
		// 我们假设 API 密钥的健康状态代表通道的健康状态。
		status := s.healthManager.GetStatus(types.ResourceTypeAPIKey, ch.APIKey.ID)
		if status.Status != types.HealthStatusAvailable && status.Status != types.HealthStatusUnknown {
			continue // 跳过不健康的通道
		}

		if value, ok := s.lastUsed.Load(ch.APIKey.ID); ok {
			lastUsedTime := value.(time.Time)
			if first || lastUsedTime.Before(oldestTime) {
				selected = ch
				oldestTime = lastUsedTime
				first = false
			}
		} else {
			// 如果通道从未被使用过，优先选择它
			selected = ch
			oldestTime = time.Now()
			break
		}
	}

	if selected == nil {
		return nil, errors.New("没有可用的健康通道")
	}

	// 更新选中通道的最后使用时间
	s.lastUsed.Store(selected.APIKey.ID, time.Now())

	return selected, nil
}
