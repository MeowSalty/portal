package adapter

import (
	"fmt"
	"strings"
	"sync"
)

// ProviderFactory 定义提供商工厂函数的接口
type ProviderFactory func() Provider

// providerFactories 存储所有已注册的提供商工厂
var providerFactories = make(map[string]ProviderFactory)

var providerFactoriesMu sync.RWMutex

// adapterCache 缓存已创建的 Adapter 实例。
// Provider 无状态后，同名 Adapter 可安全复用。
var (
	adapterCache   = make(map[string]*Adapter)
	adapterCacheMu sync.RWMutex
)

// RegisterProviderFactory 注册一个新的提供商工厂
//
// 该函数通常在提供商包的 init() 函数中调用，实现自动注册。
// 注册名称不区分大小写，内部会自动转换为小写存储。
func RegisterProviderFactory(name string, factory ProviderFactory) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		panic("提供商名称不能为空")
	}
	if factory == nil {
		panic(fmt.Sprintf("提供商 %s 的工厂函数不能为空", name))
	}
	providerFactoriesMu.Lock()
	providerFactories[normalizedName] = factory
	providerFactoriesMu.Unlock()

	// 注册同名工厂后使适配器缓存失效，确保后续读取到新工厂。
	adapterCacheMu.Lock()
	delete(adapterCache, normalizedName)
	adapterCacheMu.Unlock()
}

// GetAdapter 根据提供商名称获取适配器实例
//
// Provider 无状态后，同名 Adapter 会被缓存复用，避免每次请求创建新实例。
// 名称匹配不区分大小写。
func GetAdapter(name string) (*Adapter, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return nil, fmt.Errorf("提供商名称不能为空")
	}

	// 快路径：读锁查缓存
	adapterCacheMu.RLock()
	if a, ok := adapterCache[normalizedName]; ok {
		adapterCacheMu.RUnlock()
		return a, nil
	}
	adapterCacheMu.RUnlock()

	providerFactoriesMu.RLock()
	factory, exists := providerFactories[normalizedName]
	providerFactoriesMu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("未注册的提供商: %s", name)
	}

	// 慢路径：写锁创建并缓存
	adapterCacheMu.Lock()
	defer adapterCacheMu.Unlock()

	// double-check
	if a, ok := adapterCache[normalizedName]; ok {
		return a, nil
	}

	a := NewAdapterFromProvider(factory())
	adapterCache[normalizedName] = a
	return a, nil
}

// GetProvider 根据提供商名称获取提供商实例
//
// 该函数用于直接获取提供商实例，不创建适配器。
// 适用于需要直接访问提供商特定功能的场景。
func GetProvider(name string) (Provider, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return nil, fmt.Errorf("提供商名称不能为空")
	}

	providerFactoriesMu.RLock()
	factory, exists := providerFactories[normalizedName]
	providerFactoriesMu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("未注册的提供商: %s", name)
	}

	return factory(), nil
}

// IsProviderRegistered 检查指定名称的提供商是否已注册
func IsProviderRegistered(name string) bool {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	providerFactoriesMu.RLock()
	_, exists := providerFactories[normalizedName]
	providerFactoriesMu.RUnlock()
	return exists
}

// GetRegisteredProviderTypes 返回所有已注册的提供商类型名称
func GetRegisteredProviderTypes() []string {
	providerFactoriesMu.RLock()
	defer providerFactoriesMu.RUnlock()

	types := make([]string, 0, len(providerFactories))
	for name := range providerFactories {
		types = append(types, name)
	}
	return types
}
