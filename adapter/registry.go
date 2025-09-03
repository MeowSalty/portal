package adapter

import (
	"log/slog"

	"github.com/MeowSalty/portal/types"
)

// AdapterFactory 定义 adapter 工厂函数的接口
type AdapterFactory func(logger *slog.Logger) types.Adapter

// adapterFactories 存储所有已注册的 adapter 工厂
var adapterFactories = make(map[string]AdapterFactory)

// RegisterAdapterFactory 注册一个新的 adapter 工厂
func RegisterAdapterFactory(name string, factory AdapterFactory) {
	adapterFactories[name] = factory
}

// GetRegisteredAdapterTypes 返回所有已注册的适配器类型名称
func GetRegisteredAdapterTypes() []string {
	types := make([]string, 0, len(adapterFactories))
	for name := range adapterFactories {
		types = append(types, name)
	}
	return types
}

// New 创建适配器实例
func New(logger *slog.Logger) map[string]types.Adapter {
	adapters := make(map[string]types.Adapter)

	// 初始化所有已注册的适配器
	for name, factory := range adapterFactories {
		adapters[name] = factory(logger)
	}

	logger.Info("已初始化 adapters", slog.Int("数量", len(adapters)))

	return adapters
}
