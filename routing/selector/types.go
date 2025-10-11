package selector

import "time"

// ChannelInfo 包含通道的基本信息和选择策略所需的元数据
type ChannelInfo struct {
	ID           string    // 通道唯一标识符
	LastUsedTime time.Time // 最近使用时间
}

// Selector 定义了通道选择器接口
type Selector interface {
	// Select 从给定的通道列表中选择一个通道
	// 返回选中的通道 ID，如果没有可用通道则返回错误
	Select(channels []ChannelInfo) (string, error)

	// Name 返回选择器的名称
	Name() string
}

// SelectorType 定义选择器类型
type SelectorType string

const (
	// RandomSelector 随机选择器
	RandomSelector SelectorType = "random"

	// LRUSelector 最近最少使用选择器
	LRUSelector SelectorType = "lru"
)

// SelectorFactory 选择器工厂函数类型
type SelectorFactory func() Selector
