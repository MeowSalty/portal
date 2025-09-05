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

// New 创建一个新的健康状态管理器
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
func New(
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
			// 收到关闭信号，执行最终同步
			m.logger.Info("健康状态管理器正在关闭，执行最终同步...")
			if err := m.syncDirtyToDB(context.Background()); err != nil {
				m.logger.Error("关闭时健康同步失败", slog.Any("error", err))
			}
			return
		}
	}
}

// Shutdown 关闭健康状态管理器
//
// 此方法会停止后台同步进程并等待所有操作完成
func (m *Manager) Shutdown() {
	close(m.closeChan)
	m.wg.Wait()
	m.logger.Info("健康状态管理器已关闭")
}

// GetStatus 获取指定资源的健康状态
//
// 如果缓存中不存在该资源的健康状态，则会创建一个新的健康状态对象
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//
// 返回值：
//   - *types.Health: 健康状态对象
func (m *Manager) GetStatus(resourceType types.ResourceType, resourceID uint) *types.Health {
	key := m.generateKey(resourceType, resourceID)

	if status, ok := m.cache.Load(key); ok {
		return status.(*types.Health)
	}

	// 如果缓存中不存在，创建一个新的健康状态对象
	status := &types.Health{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Status:       types.HealthStatusUnknown,
		LastCheckAt:  time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 存储到缓存和脏数据中
	m.cache.Store(key, status)
	m.dirty.Store(key, status)

	return status
}

// UpdateStatus 更新指定资源的健康状态
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//   - success: 是否成功
//   - errorMessage: 错误信息（如果失败）
//   - errorCode: 错误代码（如果失败）
func (m *Manager) UpdateStatus(
	resourceType types.ResourceType,
	resourceID uint,
	success bool,
	errorMessage string,
	errorCode int,
) {
	key := m.generateKey(resourceType, resourceID)
	now := time.Now()

	var status *types.Health
	if cachedStatus, ok := m.cache.Load(key); ok {
		status = cachedStatus.(*types.Health)
	} else {
		// 如果缓存中不存在，创建一个新的健康状态对象
		status = &types.Health{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Status:       types.HealthStatusUnknown,
			LastCheckAt:  now,
			CreatedAt:    now,
		}
		m.cache.Store(key, status)
	}

	// 更新状态
	status.LastCheckAt = now
	status.UpdatedAt = now

	if success {
		// 成功情况
		status.SuccessCount++
		status.LastSuccessAt = &now
		status.LastError = ""
		status.LastErrorCode = 0

		// 重置错误计数和退避状态
		status.ErrorCount = 0
		status.RetryCount = 0
		status.NextAvailableAt = nil
		status.BackoffDuration = 0

		// 更新状态为可用
		status.Status = types.HealthStatusAvailable
	} else {
		// 失败情况
		status.ErrorCount++
		status.LastError = errorMessage
		status.LastErrorCode = errorCode

		// 应用指数退避策略
		m.applyBackoff(status)
	}

	// 标记为脏数据以便后续同步
	m.dirty.Store(key, status)
}

// applyBackoff 应用指数退避策略
//
// 参数：
//   - status: 健康状态对象
func (m *Manager) applyBackoff(status *types.Health) {
	// 增加重试次数
	status.RetryCount++

	// 计算退避时长（以秒为单位）
	// 初始退避时长为 1 秒，每次重试翻倍（指数退避）
	backoffSeconds := int64(1 << uint(status.RetryCount-1)) // 1, 2, 4, 8, 16, ...

	// 设置最大退避时长（例如 5 分钟）
	if backoffSeconds > 300 {
		backoffSeconds = 300
	}

	status.BackoffDuration = backoffSeconds

	// 计算下次可用时间
	nextAvailable := time.Now().Add(time.Duration(backoffSeconds) * time.Second)
	status.NextAvailableAt = &nextAvailable

	// 更新状态为警告（使用退避策略）
	status.Status = types.HealthStatusWarning
}

// FilterHealthyChannels 过滤出健康的通道
//
// 该方法会检查通道列表中的每个通道，返回当前可用的通道
// 一个通道要被认为是健康的，其平台、模型和 API 密钥都必须是健康的
//
// 参数：
//   - channels: 通道列表
//   - now: 当前时间
//
// 返回值：
//   - []*types.Channel: 健康的通道列表
func (m *Manager) FilterHealthyChannels(channels []*types.Channel, now time.Time) []*types.Channel {
	var healthyChannels []*types.Channel

	for _, channel := range channels {
		// 检查平台的健康状态
		platformStatus := m.GetStatus(types.ResourceTypePlatform, channel.Platform.ID)
		if !m.isStatusHealthy(platformStatus, now) {
			continue
		}

		// 检查模型的健康状态
		modelStatus := m.GetStatus(types.ResourceTypeModel, channel.Model.ID)
		if !m.isStatusHealthy(modelStatus, now) {
			continue
		}

		// 检查 API 密钥的健康状态
		apiKeyStatus := m.GetStatus(types.ResourceTypeAPIKey, channel.APIKey.ID)
		if !m.isStatusHealthy(apiKeyStatus, now) {
			continue
		}

		// 所有组件都健康，将通道添加到健康通道列表中
		healthyChannels = append(healthyChannels, channel)
	}

	return healthyChannels
}

// isStatusHealthy 检查给定的健康状态是否为健康状态
//
// 参数：
//   - status: 健康状态对象
//   - now: 当前时间
//
// 返回值：
//   - bool: 如果状态健康返回 true，否则返回 false
func (m *Manager) isStatusHealthy(status *types.Health, now time.Time) bool {
	// 检查状态是否为可用或未知（未知状态表示尚未进行健康检查）
	if status.Status == types.HealthStatusAvailable || status.Status == types.HealthStatusUnknown {
		return true
	}

	// 对于警告状态，检查是否已到下次可用时间
	if status.Status == types.HealthStatusWarning && status.NextAvailableAt != nil {
		if now.After(*status.NextAvailableAt) {
			// 退避时间已过，可以认为是健康的
			return true
		}
	}

	return false
}

// generateKey 生成资源的缓存键
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//
// 返回值：
//   - string: 缓存键
func (m *Manager) generateKey(resourceType types.ResourceType, resourceID uint) string {
	return fmt.Sprintf("%d:%d", resourceType, resourceID)
}

// syncDirtyToDB 将脏数据同步到数据库
//
// 参数：
//   - ctx: 上下文
//
// 返回值：
//   - error: 错误信息
func (m *Manager) syncDirtyToDB(ctx context.Context) error {
	var statusesToSync []*types.Health
	m.dirty.Range(func(key, value interface{}) bool {
		status := value.(*types.Health)
		statusesToSync = append(statusesToSync, status)
		return true
	})

	if len(statusesToSync) == 0 {
		return nil
	}

	if err := m.repo.BatchUpdateHealthStatus(ctx, statusesToSync); err != nil {
		return fmt.Errorf("批量更新健康状态失败：%w", err)
	}

	// 清理已同步的脏数据
	for _, status := range statusesToSync {
		key := m.generateKey(status.ResourceType, status.ResourceID)
		m.dirty.Delete(key)
	}

	m.logger.Debug("健康状态同步完成", slog.Int("count", len(statusesToSync)))
	return nil
}
