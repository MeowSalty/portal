package selector

import "time"

// ChannelInfo 包含通道的基本信息和选择策略所需的元数据
type ChannelInfo struct {
	ID              string    // 通道唯一标识符
	PlatformID      uint      // 平台 ID
	ModelID         uint      // 模型 ID
	APIKeyID        uint      // API 密钥 ID
	LastTryPlatform time.Time // 平台最近尝试时间
	LastTryModel    time.Time // 模型最近尝试时间
	LastTryKey      time.Time // 密钥最近尝试时间
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

	// LRUSelector 多维 LRU 选择器
	LRUSelector SelectorType = "multi_dim_lru"
)

// SelectorFactory 选择器工厂函数类型
type SelectorFactory func() Selector
