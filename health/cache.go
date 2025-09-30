// Package health 提供了资源健康状态管理功能
package health

import (
	"fmt"
	"sync"

	"github.com/MeowSalty/portal/types"
)

// Cache 定义健康状态缓存接口
type Cache interface {
	// Get 获取指定资源的健康状态
	Get(resourceType types.ResourceType, resourceID uint) (*types.Health, bool)
	// Set 设置指定资源的健康状态
	Set(resourceType types.ResourceType, resourceID uint, status *types.Health)
	// LoadAll 批量加载健康状态
	LoadAll(statuses []*types.Health) int
	// ForEach 遍历所有缓存项
	ForEach(fn func(key string, status *types.Health) bool)
}

// cache 实现了线程安全的健康状态缓存
type cache struct {
	data sync.Map
}

// NewCache 创建一个新的缓存实例
func NewCache() Cache {
	return &cache{}
}

// Get 获取指定资源的健康状态
func (c *cache) Get(resourceType types.ResourceType, resourceID uint) (*types.Health, bool) {
	key := generateKey(resourceType, resourceID)
	if value, ok := c.data.Load(key); ok {
		return value.(*types.Health), true
	}
	return nil, false
}

// Set 设置指定资源的健康状态
func (c *cache) Set(resourceType types.ResourceType, resourceID uint, status *types.Health) {
	key := generateKey(resourceType, resourceID)
	c.data.Store(key, status)
}

// LoadAll 批量加载健康状态
//
// 返回加载的健康状态数量
func (c *cache) LoadAll(statuses []*types.Health) int {
	count := 0
	for _, status := range statuses {
		// 为缓存创建堆分配的副本以避免数据竞争
		s := *status
		key := generateKey(s.ResourceType, s.ResourceID)
		c.data.Store(key, &s)
		count++
	}
	return count
}

// ForEach 遍历所有缓存项
//
// 如果 fn 返回 false，则停止遍历
func (c *cache) ForEach(fn func(key string, status *types.Health) bool) {
	c.data.Range(func(key, value interface{}) bool {
		return fn(key.(string), value.(*types.Health))
	})
}

// generateKey 生成资源的缓存键
func generateKey(resourceType types.ResourceType, resourceID uint) string {
	return fmt.Sprintf("%d:%d", resourceType, resourceID)
}
