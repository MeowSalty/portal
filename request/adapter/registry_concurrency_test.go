package adapter

import (
	"sync"
	"testing"
)

func TestRegistry_ConcurrentRegisterReadAndEnumerate(t *testing.T) {
	const providerName = "registry-concurrency-provider"

	// 清理可能存在的历史状态，避免测试间相互影响。
	providerFactoriesMu.Lock()
	delete(providerFactories, providerName)
	providerFactoriesMu.Unlock()
	adapterCacheMu.Lock()
	delete(adapterCache, providerName)
	adapterCacheMu.Unlock()

	var wg sync.WaitGroup

	// 并发注册同名工厂，模拟热更新场景。
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			RegisterProviderFactory(providerName, func() Provider {
				return NewOpenAIProvider()
			})
		}()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = GetAdapter(providerName)
			_, _ = GetProvider(providerName)
			_ = IsProviderRegistered(providerName)
			_ = GetRegisteredProviderTypes()
		}()
	}

	wg.Wait()

	if !IsProviderRegistered(providerName) {
		t.Fatalf("提供商应已注册")
	}
}

func TestRegisterProviderFactory_InvalidatesAdapterCache(t *testing.T) {
	const name = "registry-cache-reset-provider"

	providerFactoriesMu.Lock()
	delete(providerFactories, name)
	providerFactoriesMu.Unlock()
	adapterCacheMu.Lock()
	delete(adapterCache, name)
	adapterCacheMu.Unlock()

	RegisterProviderFactory(name, func() Provider { return NewOpenAIProvider() })
	a1, err := GetAdapter(name)
	if err != nil {
		t.Fatalf("首次获取适配器失败：%v", err)
	}

	RegisterProviderFactory(name, func() Provider { return NewGeminiProvider() })
	a2, err := GetAdapter(name)
	if err != nil {
		t.Fatalf("重注册后获取适配器失败：%v", err)
	}

	if a1 == a2 {
		t.Fatalf("重注册后适配器缓存应失效并重建")
	}
}
