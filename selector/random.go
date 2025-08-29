package selector

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/MeowSalty/portal/health"
	"github.com/MeowSalty/portal/types"
)

// RandomSelector 实现了随机选择策略的通道选择器
// 它会优先选择健康状态为 Unknown 的通道
type RandomSelector struct {
	healthManager *health.Manager // 健康状态管理器
	rand          *rand.Rand      // 随机数生成器
}

// NewRandomSelector 创建一个新的随机选择器
// 参数：
//
//	healthManager - 用于检查通道健康状态的管理器
func NewRandomSelector(healthManager *health.Manager) *RandomSelector {
	return &RandomSelector{
		healthManager: healthManager,
		// 初始化随机数生成器，使用当前时间作为种子
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 从提供的通道列表中随机选择一个通道
// 优先从状态未知的通道中随机选择
// 如果没有未知状态的通道，则从所有健康通道中随机选择
func (s *RandomSelector) Select(_ context.Context, channels []*types.Channel) (*types.Channel, error) {
	if len(channels) == 0 {
		return nil, errors.New("通道列表不能为空")
	}

	unknownChannels := make([]*types.Channel, 0)
	for _, ch := range channels {
		// 我们假设 API 密钥的健康状态代表通道的健康状态。
		status := s.healthManager.GetStatus(types.ResourceTypeAPIKey, ch.APIKey.ID)
		if status.Status == types.HealthStatusUnknown {
			unknownChannels = append(unknownChannels, ch)
		}
	}

	// 优先从未知状态通道中随机选择
	if len(unknownChannels) > 0 {
		selectedIndex := s.rand.Intn(len(unknownChannels))
		return unknownChannels[selectedIndex], nil
	}

	// 没有未知状态通道，则从所有通道中随机选择
	selectedIndex := s.rand.Intn(len(channels))
	return channels[selectedIndex], nil
}
