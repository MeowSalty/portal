// Package health 提供了资源健康状态管理功能
package health

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/MeowSalty/portal/types"
)

// Manager 管理所有资源的健康状态
type Manager struct {
	repo         types.DataRepository // 数据仓库接口
	logger       *slog.Logger         // 日志记录器
	cache        sync.Map             // 健康状态缓存，key: string (fmt.Sprintf), value: *types.Health
	syncInterval time.Duration        // 同步间隔
	closeChan    chan struct{}        // 关闭信号通道
	wg           sync.WaitGroup       // 等待组，用于等待后台协程结束
	dirty        sync.Map             // 脏数据缓存，key: string, value: *types.Health
}

// NewManager 创建一个新的健康状态管理器
//
// 它从仓库加载初始数据并启动后台同步进程
//
// 参数：
//   - ctx: 上下文
//   - repo: 数据仓库接口
//   - logger: 日志记录器
//   - syncInterval: 同步间隔
//
// 返回值：
//   - *Manager: 健康状态管理器实例
//   - error: 错误信息
func NewManager(
	ctx context.Context,
	repo types.DataRepository,
	logger *slog.Logger,
	syncInterval time.Duration,
) (*Manager, error) {
	if logger == nil {
		logger = slog.Default() // 回退到默认日志记录器
	}
	m := &Manager{
		repo:         repo,
		logger:       logger.WithGroup("health_manager"),
		syncInterval: syncInterval,
		closeChan:    make(chan struct{}),
	}

	if err := m.loadInitialData(ctx); err != nil {
		return nil, fmt.Errorf("加载初始健康数据失败：%w", err)
	}

	m.wg.Add(1)
	go m.run()

	return m, nil
}

// loadInitialData 从仓库获取所有健康状态并填充缓存
//
// 参数：
//   - ctx: 上下文
//
// 返回值：
//   - error: 错误信息
func (m *Manager) loadInitialData(ctx context.Context) error {
	statuses, err := m.repo.GetAllHealthStatus(ctx)
	if err != nil {
		return err
	}

	count := 0
	for _, status := range statuses {
		// 为缓存创建堆分配的副本以避免数据竞争
		s := *status
		key := m.generateKey(s.ResourceType, s.ResourceID)
		m.cache.Store(key, &s)
		count++
	}
	m.logger.Info("成功加载健康状态缓存", slog.Int("count", count))
	return nil
}

// run 是后台进程，定期将缓存与数据库同步
func (m *Manager) run() {
	defer m.wg.Done()
	ticker := time.NewTicker(m.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 定期同步脏数据
			if err := m.syncDirtyToDB(context.Background()); err != nil {
				m.logger.Error("执行定期健康同步失败", slog.Any("error", err))
			}
		case <-m.closeChan:
			m.logger.Info("正在关闭健康状态管理器后台进程")
			// 在关闭时执行最终同步
			if err := m.syncAllToDB(context.Background()); err != nil {
				m.logger.Error("执行最终健康同步失败", slog.Any("error", err))
			}
			return
		}
	}
}

// Shutdown 优雅地停止后台同步进程
func (m *Manager) Shutdown() {
	close(m.closeChan)
	m.wg.Wait()
	m.logger.Info("健康状态管理器已成功关闭")
}

// GetStatus 获取特定资源的健康状态
//
// 如果未找到，则返回默认的'Unknown'状态
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//
// 返回值：
//   - *types.Health: 健康状态
func (m *Manager) GetStatus(resourceType types.ResourceType, resourceID uint) *types.Health {
	key := m.generateKey(resourceType, resourceID)
	if value, ok := m.cache.Load(key); ok {
		// 类型断言并返回存储的健康状态
		if status, ok := value.(*types.Health); ok {
			return status
		}
	}

	// 如果缓存中没有，则返回默认的未知状态
	return &types.Health{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Status:       types.HealthStatusUnknown,
		LastCheckAt:  time.Now(),
	}
}

// UpdateStatusOnSuccess 在 API 调用成功后更新健康状态
//
// 此方法会更新所有相关资源的健康状态：APIKey、Model 和 Platform
//
// 参数：
//   - channel: 通道信息
func (m *Manager) UpdateStatusOnSuccess(channel *types.Channel) {
	now := time.Now()

	// 更新 API Key 的健康状态
	if channel.APIKey != nil {
		key := m.generateKey(types.ResourceTypeAPIKey, channel.APIKey.ID)
		status := m.getOrCreateStatus(key, types.ResourceTypeAPIKey, channel.APIKey.ID)

		// 设置关联资源 ID
		if channel.Model != nil {
			status.RelatedAPIKeyID = &channel.APIKey.ID
		}
		if channel.Platform != nil {
			status.RelatedPlatformID = &channel.Platform.ID
		}

		status.Status = types.HealthStatusAvailable
		status.LastSuccessAt = &now
		status.LastCheckAt = now
		status.SuccessCount++
		status.RetryCount = 0 // 成功时重置重试计数
		status.BackoffDuration = 0
		status.NextAvailableAt = nil

		m.cache.Store(key, status)
		m.dirty.Store(key, status)
		m.logger.Debug("API Key 健康状态已更新为可用", slog.String("key", key))
	}

	// 更新 Model 的健康状态
	if channel.Model != nil {
		key := m.generateKey(types.ResourceTypeModel, channel.Model.ID)
		status := m.getOrCreateStatus(key, types.ResourceTypeModel, channel.Model.ID)

		// 设置关联资源 ID
		if channel.Platform != nil {
			status.RelatedPlatformID = &channel.Platform.ID
		}
		if channel.APIKey != nil {
			status.RelatedAPIKeyID = &channel.APIKey.ID
		}

		status.Status = types.HealthStatusAvailable
		status.LastSuccessAt = &now
		status.LastCheckAt = now
		status.SuccessCount++
		status.RetryCount = 0
		status.BackoffDuration = 0
		status.NextAvailableAt = nil

		m.cache.Store(key, status)
		m.dirty.Store(key, status)
		m.logger.Debug("Model 健康状态已更新为可用", slog.String("key", key))
	}

	// 更新 Platform 的健康状态
	if channel.Platform != nil {
		key := m.generateKey(types.ResourceTypePlatform, channel.Platform.ID)
		status := m.getOrCreateStatus(key, types.ResourceTypePlatform, channel.Platform.ID)

		// 设置关联资源 ID
		if channel.Model != nil {
			status.RelatedAPIKeyID = &channel.Model.ID
		}
		if channel.APIKey != nil {
			status.RelatedAPIKeyID = &channel.APIKey.ID
		}

		status.Status = types.HealthStatusAvailable
		status.LastSuccessAt = &now
		status.LastCheckAt = now
		status.SuccessCount++
		status.RetryCount = 0
		status.BackoffDuration = 0
		status.NextAvailableAt = nil

		m.cache.Store(key, status)
		m.dirty.Store(key, status)
		m.logger.Debug("Platform 健康状态已更新为可用", slog.String("key", key))
	}
}

// UpdateStatusOnFailure 在 API 调用失败后更新健康状态
//
// 此方法会更新所有相关资源的健康状态：APIKey、Model 和 Platform
//
// 参数：
//   - channel: 通道信息
//   - err: 错误信息
func (m *Manager) UpdateStatusOnFailure(channel *types.Channel, err error) {
	now := time.Now()

	// 更新 API Key 的健康状态
	if channel.APIKey != nil {
		key := m.generateKey(types.ResourceTypeAPIKey, channel.APIKey.ID)
		status := m.getOrCreateStatus(key, types.ResourceTypeAPIKey, channel.APIKey.ID)

		// 设置关联资源 ID
		if channel.Model != nil {
			status.RelatedAPIKeyID = &channel.Model.ID
		}
		if channel.Platform != nil {
			status.RelatedPlatformID = &channel.Platform.ID
		}

		status.Status = types.HealthStatusWarning
		status.LastError = err.Error()
		status.LastCheckAt = now
		status.ErrorCount++
		status.RetryCount++

		// 实现指数退避
		// 示例：base_duration * 2^(retry_count-1)
		baseDuration := int64(60) // 1 分钟基础时间
		backoff := baseDuration * (1 << (status.RetryCount - 1))
		if backoff > 3600 { // 最多 1 小时
			backoff = 3600
		}
		status.BackoffDuration = backoff
		nextAvailable := now.Add(time.Duration(backoff) * time.Second)
		status.NextAvailableAt = &nextAvailable

		m.cache.Store(key, status)
		m.dirty.Store(key, status)
		m.logger.Warn("API Key 健康状态已更新为警告",
			slog.String("key", key),
			slog.String("error", err.Error()),
			slog.Int64("backoff_seconds", backoff))
	}

	// 更新 Model 的健康状态
	if channel.Model != nil {
		key := m.generateKey(types.ResourceTypeModel, channel.Model.ID)
		status := m.getOrCreateStatus(key, types.ResourceTypeModel, channel.Model.ID)

		// 设置关联资源 ID
		if channel.Platform != nil {
			status.RelatedPlatformID = &channel.Platform.ID
		}
		if channel.APIKey != nil {
			status.RelatedAPIKeyID = &channel.APIKey.ID
		}

		status.Status = types.HealthStatusWarning
		status.LastError = err.Error()
		status.LastCheckAt = now
		status.ErrorCount++
		status.RetryCount++

		// 实现指数退避
		baseDuration := int64(60) // 1 分钟基础时间
		backoff := baseDuration * (1 << (status.RetryCount - 1))
		if backoff > 3600 { // 最多 1 小时
			backoff = 3600
		}
		status.BackoffDuration = backoff
		nextAvailable := now.Add(time.Duration(backoff) * time.Second)
		status.NextAvailableAt = &nextAvailable

		m.cache.Store(key, status)
		m.dirty.Store(key, status)
		m.logger.Warn("Model 健康状态已更新为警告",
			slog.String("key", key),
			slog.String("error", err.Error()),
			slog.Int64("backoff_seconds", backoff))
	}

	// 更新 Platform 的健康状态
	if channel.Platform != nil {
		key := m.generateKey(types.ResourceTypePlatform, channel.Platform.ID)
		status := m.getOrCreateStatus(key, types.ResourceTypePlatform, channel.Platform.ID)

		// 设置关联资源 ID
		if channel.Model != nil {
			status.RelatedAPIKeyID = &channel.Model.ID
		}
		if channel.APIKey != nil {
			status.RelatedAPIKeyID = &channel.APIKey.ID
		}

		status.Status = types.HealthStatusWarning
		status.LastError = err.Error()
		status.LastCheckAt = now
		status.ErrorCount++
		status.RetryCount++

		// 实现指数退避
		baseDuration := int64(60) // 1 分钟基础时间
		backoff := baseDuration * (1 << (status.RetryCount - 1))
		if backoff > 3600 { // 最多 1 小时
			backoff = 3600
		}
		status.BackoffDuration = backoff
		nextAvailable := now.Add(time.Duration(backoff) * time.Second)
		status.NextAvailableAt = &nextAvailable

		m.cache.Store(key, status)
		m.dirty.Store(key, status)
		m.logger.Warn("Platform 健康状态已更新为警告",
			slog.String("key", key),
			slog.String("error", err.Error()),
			slog.Int64("backoff_seconds", backoff))
	}
}

// FilterHealthyChannels 过滤出健康的通道
func (m *Manager) FilterHealthyChannels(channels []*types.Channel) []*types.Channel {
	return m.FilterHealthyChannelsWithTime(channels, time.Now())
}

// FilterHealthyChannelsWithTime 使用指定的时间过滤出健康的通道，避免重复调用 time.Now()
func (m *Manager) FilterHealthyChannelsWithTime(channels []*types.Channel, now time.Time) []*types.Channel {
	healthyChannels := make([]*types.Channel, 0, len(channels))

	for _, ch := range channels {
		// 检查平台健康状态
		platformHealthy := true
		if ch.Platform != nil {
			status := m.GetStatus(types.ResourceTypePlatform, ch.Platform.ID)
			if status.Status == types.HealthStatusUnavailable ||
				(status.Status == types.HealthStatusWarning && status.NextAvailableAt != nil && now.Before(*status.NextAvailableAt)) {
				platformHealthy = false
			}
		}

		// 检查模型健康状态
		modelHealthy := true
		if ch.Model != nil {
			status := m.GetStatus(types.ResourceTypeModel, ch.Model.ID)
			if status.Status == types.HealthStatusUnavailable ||
				(status.Status == types.HealthStatusWarning && status.NextAvailableAt != nil && now.Before(*status.NextAvailableAt)) {
				modelHealthy = false
			}
		}

		// 检查 API 密钥健康状态
		apiKeyHealthy := true
		if ch.APIKey != nil {
			status := m.GetStatus(types.ResourceTypeAPIKey, ch.APIKey.ID)
			if status.Status == types.HealthStatusUnavailable ||
				(status.Status == types.HealthStatusWarning && status.NextAvailableAt != nil && now.Before(*status.NextAvailableAt)) {
				apiKeyHealthy = false
			}
		}

		// 只有当所有相关资源都健康时，通道才被认为是健康的
		if platformHealthy && modelHealthy && apiKeyHealthy {
			healthyChannels = append(healthyChannels, ch)
		}
	}
	return healthyChannels
}

// getOrCreateStatus 获取或创建健康状态
//
// 参数：
//   - key: 缓存键
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//
// 返回值：
//   - *types.Health: 健康状态
func (m *Manager) getOrCreateStatus(key string, resourceType types.ResourceType, resourceID uint) *types.Health {
	if value, ok := m.cache.Load(key); ok {
		// 类型断言并返回存储的健康状态
		if status, ok := value.(*types.Health); ok {
			return status
		}
	}

	// 如果未找到，则创建一个新的健康状态对象
	return &types.Health{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Status:       types.HealthStatusUnknown,
		LastCheckAt:  time.Now(),
	}
}

// generateKey 生成资源的键
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//
// 返回值：
//   - string: 生成的键
func (m *Manager) generateKey(resourceType types.ResourceType, resourceID uint) string {
	return fmt.Sprintf("%s-%d", resourceType, resourceID)
}

// syncDirtyToDB 将缓存中的脏数据同步到数据库
func (m *Manager) syncDirtyToDB(ctx context.Context) error {
	var dirtyStatuses []*types.Health
	m.dirty.Range(func(key, value interface{}) bool {
		if status, ok := value.(*types.Health); ok {
			dirtyStatuses = append(dirtyStatuses, status)
		}
		return true
	})

	if len(dirtyStatuses) == 0 {
		m.logger.Debug("没有需要同步的健康状态")
		return nil
	}

	if err := m.repo.BatchUpdateHealthStatus(ctx, dirtyStatuses); err != nil {
		return fmt.Errorf("批量更新健康状态失败：%w", err)
	}

	// 清理脏数据标记
	m.dirty.Range(func(key, value interface{}) bool {
		m.dirty.Delete(key)
		return true
	})

	m.logger.Info("成功同步脏健康状态到数据库", slog.Int("count", len(dirtyStatuses)))
	return nil
}

// syncAllToDB 将缓存中的所有数据同步到数据库
func (m *Manager) syncAllToDB(ctx context.Context) error {
	var allStatuses []*types.Health
	m.cache.Range(func(key, value interface{}) bool {
		if status, ok := value.(*types.Health); ok {
			allStatuses = append(allStatuses, status)
		}
		return true
	})

	if len(allStatuses) == 0 {
		m.logger.Info("缓存中没有需要同步的健康状态")
		return nil
	}

	if err := m.repo.BatchUpdateHealthStatus(ctx, allStatuses); err != nil {
		return fmt.Errorf("批量更新所有健康状态失败：%w", err)
	}

	m.logger.Info("成功同步所有健康状态到数据库", slog.Int("count", len(allStatuses)))
	return nil
}
