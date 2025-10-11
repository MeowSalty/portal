// Package health 提供了资源健康状态管理功能
package health

import (
	"fmt"
	"sync"
)

// Cache 定义健康状态缓存接口
type Cache interface {
	// Get 获取指定资源的健康状态
	Get(resourceType ResourceType, resourceID uint) (*Health, bool)
	// Set 设置指定资源的健康状态
	Set(resourceType ResourceType, resourceID uint, status *Health)
	// LoadAll 批量加载健康状态
	LoadAll(statuses []*Health)
	// ForEach 遍历所有缓存项
	ForEach(fn func(key string, status *Health) bool)
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
func (c *cache) Get(resourceType ResourceType, resourceID uint) (*Health, bool) {
	key := generateKey(resourceType, resourceID)
	value, ok := c.data.Load(key)
	if !ok {
		return nil, false
	}

	status, ok := value.(*Health)
	if !ok {
		// 类型不匹配，清理无效数据
		c.data.Delete(key)
		return nil, false
	}
	return status, true
}

// Set 设置指定资源的健康状态
func (c *cache) Set(resourceType ResourceType, resourceID uint, status *Health) {
	key := generateKey(resourceType, resourceID)
	c.data.Store(key, status)
}

// LoadAll 批量加载健康状态
func (c *cache) LoadAll(statuses []*Health) {
	for _, status := range statuses {
		if status == nil {
			continue
		}
		// 为缓存创建堆分配的副本以避免数据竞争
		s := *status
		key := generateKey(s.ResourceType, s.ResourceID)
		c.data.Store(key, &s)
	}
}

// ForEach 遍历所有缓存项
//
// 如果 fn 返回 false，则停止遍历
func (c *cache) ForEach(fn func(key string, status *Health) bool) {
	c.data.Range(func(key, value interface{}) bool {
		k, ok1 := key.(string)
		v, ok2 := value.(*Health)
		if !ok1 || !ok2 {
			// 类型不匹配，清理无效数据
			c.data.Delete(key)
			return true
		}
		return fn(k, v)
	})
}

// generateKey 生成资源的缓存键
func generateKey(resourceType ResourceType, resourceID uint) string {
	return fmt.Sprintf("%d:%d", resourceType, resourceID)
}
