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

// CreateAdapters 创建指定类型的适配器实例
// 如果 adapterTypes 为空，则创建所有已注册的适配器
func CreateAdapters(logger *slog.Logger, adapterTypes []string) map[string]types.Adapter {
	adapters := make(map[string]types.Adapter)

	// 如果未指定适配器类型，则初始化所有已注册的适配器
	if len(adapterTypes) == 0 {
		for name, factory := range adapterFactories {
			adapters[name] = factory(logger)
		}
	} else {
		// 只初始化指定类型的适配器
		for _, adapterType := range adapterTypes {
			if factory, exists := adapterFactories[adapterType]; exists {
				adapters[adapterType] = factory(logger)
			} else {
				logger.Warn("未找到适配器工厂", slog.String("类型", adapterType))
			}
		}
	}

	logger.Info("已初始化 adapters", slog.Int("数量", len(adapters)))

	return adapters
}
