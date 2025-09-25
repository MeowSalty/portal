package adapter

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/MeowSalty/portal/types"
)

// ProviderFactory 定义提供商工厂函数的接口
type ProviderFactory func(logger *slog.Logger) Provider

// AdapterFactory 定义适配器工厂函数的接口
type AdapterFactory func(logger *slog.Logger) types.Adapter

// providerFactories 存储所有已注册的提供商工厂
var providerFactories = make(map[string]ProviderFactory)

// adapterCache 缓存已创建的适配器实例
var adapterCache = make(map[string]types.Adapter)
var cacheMutex sync.RWMutex

// RegisterProviderFactory 注册一个新的提供商工厂
func RegisterProviderFactory(name string, factory ProviderFactory) {
	providerFactories[name] = factory
}

// GetRegisteredProviderTypes 返回所有已注册的提供商类型名称
func GetRegisteredProviderTypes() []string {
	types := make([]string, 0, len(providerFactories))
	for name := range providerFactories {
		types = append(types, name)
	}
	return types
}

// NewAdapterRegistry 创建适配器注册表
func NewAdapterRegistry(logger *slog.Logger) map[string]types.Adapter {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if logger == nil {
		logger = slog.Default()
	}

	adapters := make(map[string]types.Adapter)

	// 初始化所有已注册的提供商
	for name, factory := range providerFactories {
		if cachedAdapter, exists := adapterCache[name]; exists {
			adapters[name] = cachedAdapter
			continue
		}

		provider := factory(logger)
		adapter := createAdapterFromProvider(provider, logger)

		adapterCache[name] = adapter
		adapters[name] = adapter
	}

	logger.Info("已初始化适配器",
		slog.Int("数量", len(adapters)),
		slog.Any("提供商", GetRegisteredProviderTypes()))

	return adapters
}

// createAdapterFromProvider 从 Provider 创建适配器
func createAdapterFromProvider(provider Provider, logger *slog.Logger) types.Adapter {
	return &Adapter{
		client:   NewHTTPClient(logger.WithGroup("http")),
		provider: provider,
		logger:   logger.WithGroup(provider.Name()),
	}
}

// GetAdapter 获取指定名称的适配器（线程安全）
func GetAdapter(name string, logger *slog.Logger) (types.Adapter, error) {
	cacheMutex.RLock()
	if adapter, exists := adapterCache[name]; exists {
		cacheMutex.RUnlock()
		return adapter, nil
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// 双重检查
	if adapter, exists := adapterCache[name]; exists {
		return adapter, nil
	}

	// 创建新的适配器
	factory, exists := providerFactories[name]
	if !exists {
		return nil, fmt.Errorf("未找到提供商: %s", name)
	}

	provider := factory(logger)
	adapter := createAdapterFromProvider(provider, logger)
	adapterCache[name] = adapter

	return adapter, nil
}

// ClearCache 清空适配器缓存
func ClearCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	adapterCache = make(map[string]types.Adapter)
}
