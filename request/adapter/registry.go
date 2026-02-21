package adapter

import (
	"fmt"
	"strings"
)

// ProviderFactory 定义提供商工厂函数的接口
type ProviderFactory func() Provider

// providerFactories 存储所有已注册的提供商工厂
var providerFactories = make(map[string]ProviderFactory)

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
	providerFactories[normalizedName] = factory
}

// GetAdapter 根据提供商名称获取适配器实例
//
// 该函数是适配器模式的核心入口，通过提供商名称动态创建对应的适配器。
// 名称匹配不区分大小写。
//
// 参数：
//   - name: 提供商名称（如 "openai", "google"）
//   - logger: 日志记录器，如果为 nil 则使用默认记录器
//
// 返回：
//   - *Adapter: 适配器实例
//   - error: 如果提供商未注册则返回错误
func GetAdapter(name string) (*Adapter, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return nil, fmt.Errorf("提供商名称不能为空")
	}

	factory, exists := providerFactories[normalizedName]
	if !exists {
		return nil, fmt.Errorf("未注册的提供商: %s", name)
	}

	provider := factory()
	return NewAdapterFromProvider(provider), nil
}

// GetProvider 根据提供商名称获取提供商实例
//
// 该函数用于直接获取提供商实例，不创建适配器。
// 适用于需要直接访问提供商特定功能的场景。
//
// 参数：
//   - name: 提供商名称（如 "openai", "google"）
//
// 返回：
//   - Provider: 提供商实例
//   - error: 如果提供商未注册则返回错误
func GetProvider(name string) (Provider, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return nil, fmt.Errorf("提供商名称不能为空")
	}

	factory, exists := providerFactories[normalizedName]
	if !exists {
		return nil, fmt.Errorf("未注册的提供商: %s", name)
	}

	return factory(), nil
}

// IsProviderRegistered 检查指定名称的提供商是否已注册
//
// 名称匹配不区分大小写。
func IsProviderRegistered(name string) bool {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	_, exists := providerFactories[normalizedName]
	return exists
}

// GetRegisteredProviderTypes 返回所有已注册的提供商类型名称
//
// 返回的名称列表按原始注册时的格式返回（小写）。
func GetRegisteredProviderTypes() []string {
	types := make([]string, 0, len(providerFactories))
	for name := range providerFactories {
		types = append(types, name)
	}
	return types
}
