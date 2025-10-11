package selector

import (
	"time"

	"github.com/MeowSalty/portal/errors"
)

func init() {
	Register(LRUSelector, NewLRUSelector)
}

// lruSelector 实现最近最少使用 (LRU) 选择策略
type lruSelector struct{}

// NewLRUSelector 创建一个新的 LRU 选择器实例
func NewLRUSelector() Selector {
	return &lruSelector{}
}

// Select 从给定的通道列表中选择最近最少使用的通道
//
// 实现 LRU 选择算法：
// 1. 验证输入参数
// 2. 遍历所有通道，找出最早使用的通道（LastUsedTime 最小）
// 3. 如果多个通道的 LastUsedTime 相同，选择第一个
// 4. 返回选中的通道 ID
func (s *lruSelector) Select(channels []ChannelInfo) (string, error) {
	// 验证输入
	if len(channels) == 0 {
		return "", errors.New(errors.ErrCodeInvalidArgument, "通道列表不能为空")
	}

	// 只有一个通道时直接返回
	if len(channels) == 1 {
		return channels[0].ID, nil
	}

	// 找出最近最少使用的通道（LastUsedTime 最小的通道）
	minIndex := 0
	minTime := channels[0].LastUsedTime

	for i := 1; i < len(channels); i++ {
		// 如果当前通道的最近使用时间更早（更旧），则选择它
		if channels[i].LastUsedTime.Before(minTime) {
			minIndex = i
			minTime = channels[i].LastUsedTime
		}
	}

	// 返回选中的通道 ID
	return channels[minIndex].ID, nil
}

// Name 返回选择器的名称
func (s *lruSelector) Name() string {
	return "LRU"
}

// SelectWithUpdate 选择通道并返回更新后的 LastUsedTime
//
// 这是一个辅助方法，用于在选择通道后更新其使用时间。
// 调用者需要自行将更新后的时间保存到存储中。
func (s *lruSelector) SelectWithUpdate(channels []ChannelInfo) (channelID string, updatedTime time.Time, err error) {
	channelID, err = s.Select(channels)
	if err != nil {
		return "", time.Time{}, err
	}

	updatedTime = time.Now()
	return channelID, updatedTime, nil
}
