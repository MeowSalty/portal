package selector

import (
	"math/rand"
	"time"

	"github.com/MeowSalty/portal/errors"
)

func init() {
	Register(RandomSelector, NewRandomSelector)
}

// randomSelector 实现随机选择策略
type randomSelector struct {
	rng *rand.Rand
}

// NewRandomSelector 创建一个新的随机选择器实例
func NewRandomSelector() Selector {
	return &randomSelector{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 从给定的通道列表中随机选择一个通道
//
// 实现随机选择算法：
// 1. 验证输入参数
// 2. 使用均匀分布随机选择
// 3. 返回选中的通道 ID
func (s *randomSelector) Select(channels []ChannelInfo) (string, error) {
	// 验证输入
	if len(channels) == 0 {
		return "", errors.New(errors.ErrCodeInvalidArgument, "通道列表不能为空")
	}

	// 随机选择一个索引
	index := s.rng.Intn(len(channels))

	// 返回选中的通道 ID
	return channels[index].ID, nil
}

// Name 返回选择器的名称
func (s *randomSelector) Name() string {
	return "Random"
}
