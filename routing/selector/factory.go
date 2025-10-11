package selector

import (
	"sync"

	"github.com/MeowSalty/portal/errors"
)

var (
	// factories 存储所有已注册的选择器工厂
	factories = make(map[SelectorType]SelectorFactory)

	// mu 保护 factories map 的并发访问
	mu sync.RWMutex
)

// Register 注册一个新的选择器工厂
//
// 如果选择器类型已存在，将会覆盖原有的工厂函数。
// 此函数是并发安全的。
func Register(selectorType SelectorType, factory SelectorFactory) {
	mu.Lock()
	defer mu.Unlock()
	factories[selectorType] = factory
}

// Unregister 注销一个选择器工厂
//
// 此函数是并发安全的。
func Unregister(selectorType SelectorType) {
	mu.Lock()
	defer mu.Unlock()
	delete(factories, selectorType)
}

// Create 根据选择器类型创建选择器实例
//
// 如果选择器类型不存在，返回错误。
// 此函数是并发安全的。
func Create(selectorType SelectorType) (Selector, error) {
	mu.RLock()
	factory, exists := factories[selectorType]
	mu.RUnlock()

	if !exists {
		return nil, errors.New(errors.ErrCodeNotFound, "未找到指定的选择器类型").
			WithContext("selector_type", string(selectorType))
	}

	return factory(), nil
}

// IsRegistered 检查指定的选择器类型是否已注册
//
// 此函数是并发安全的。
func IsRegistered(selectorType SelectorType) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, exists := factories[selectorType]
	return exists
}
