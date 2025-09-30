// Package health 提供了资源健康状态管理功能
package health

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/MeowSalty/portal/types"
)

// Manager 管理所有资源的健康状态
type Manager struct {
	repo    types.DataRepository // 数据仓库接口
	logger  *slog.Logger         // 日志记录器
	cache   Cache                // 健康状态缓存
	syncer  Syncer               // 后台同步器
	backoff BackoffStrategy      // 退避策略
	filter  Filter               // 健康状态过滤器
}

// Config 管理器配置
type Config struct {
	Repo         types.DataRepository // 数据仓库接口（必需）
	Logger       *slog.Logger         // 日志记录器（可选）
	SyncInterval time.Duration        // 同步间隔（可选，默认 1 分钟）
	Backoff      BackoffStrategy      // 退避策略（可选）
}

// New 创建一个新的健康状态管理器
//
// 它从仓库加载初始数据并启动后台同步进程
//
// 参数：
//   - ctx: 上下文
//   - cfg: 管理器配置
//
// 返回值：
//   - *Manager: 健康状态管理器实例
//   - error: 错误信息
func New(ctx context.Context, cfg Config) (*Manager, error) {
	// 验证必需参数
	if cfg.Repo == nil {
		return nil, fmt.Errorf("数据仓库不能为空")
	}

	// 设置默认值
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.SyncInterval <= 0 {
		cfg.SyncInterval = time.Minute
	}
	if cfg.Backoff == nil {
		cfg.Backoff = DefaultBackoffStrategy()
	}

	// 创建各个组件
	cache := NewCache()
	syncer := NewSyncer(cfg.Repo, cfg.Logger, cache, cfg.SyncInterval)
	filter := NewFilter(cache)

	m := &Manager{
		repo:    cfg.Repo,
		logger:  cfg.Logger.WithGroup("health_manager"),
		cache:   cache,
		syncer:  syncer,
		backoff: cfg.Backoff,
		filter:  filter,
	}

	// 加载初始数据
	if err := m.loadInitialData(ctx); err != nil {
		return nil, fmt.Errorf("加载初始健康数据失败：%w", err)
	}

	// 启动后台同步器
	m.syncer.Start()

	m.logger.Info("健康状态管理器已初始化",
		slog.Duration("sync_interval", cfg.SyncInterval),
		slog.String("backoff_strategy", fmt.Sprintf("%T", cfg.Backoff)),
	)

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

	count := m.cache.LoadAll(statuses)
	m.logger.Info("成功加载健康状态缓存", slog.Int("count", count))
	return nil
}

// Shutdown 关闭健康状态管理器
//
// 此方法会停止后台同步进程并等待所有操作完成
func (m *Manager) Shutdown() {
	m.logger.Info("正在关闭健康状态管理器...")
	m.syncer.Stop()
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
	if status, exists := m.cache.Get(resourceType, resourceID); exists {
		return status
	}

	// 如果缓存中不存在，创建一个新的健康状态对象
	now := time.Now()
	status := &types.Health{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Status:       types.HealthStatusUnknown,
		LastCheckAt:  now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 存储到缓存并标记为脏数据
	m.cache.Set(resourceType, resourceID, status)
	m.syncer.MarkDirty(resourceType, resourceID, status)

	m.logger.Debug("创建新的健康状态记录",
		slog.String("resource_type", fmt.Sprintf("%v", resourceType)),
		slog.Uint64("resource_id", uint64(resourceID)),
	)

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
	// 获取或创建健康状态
	status := m.GetStatus(resourceType, resourceID)

	// 更新基础信息
	now := time.Now()
	status.LastCheckAt = now
	status.UpdatedAt = now

	if success {
		// 成功情况：重置退避状态
		status.SuccessCount++
		status.LastSuccessAt = &now
		status.LastError = ""
		status.LastErrorCode = 0
		status.ErrorCount = 0

		// 使用退避策略重置状态
		m.backoff.Reset(status)

		m.logger.Debug("资源健康检查成功",
			slog.String("resource_type", fmt.Sprintf("%v", resourceType)),
			slog.Uint64("resource_id", uint64(resourceID)),
			slog.Int("success_count", status.SuccessCount),
		)
	} else {
		// 失败情况：应用退避策略
		status.ErrorCount++
		status.LastError = errorMessage
		status.LastErrorCode = errorCode

		// 使用退避策略更新状态
		m.backoff.Apply(status)

		m.logger.Warn("资源健康检查失败",
			slog.String("resource_type", fmt.Sprintf("%v", resourceType)),
			slog.Uint64("resource_id", uint64(resourceID)),
			slog.String("error", errorMessage),
			slog.Int("error_code", errorCode),
			slog.Int("retry_count", status.RetryCount),
			slog.Duration("backoff", time.Duration(status.BackoffDuration)*time.Second),
		)
	}

	// 标记为脏数据以便后续同步
	m.syncer.MarkDirty(resourceType, resourceID, status)
}

// FilterHealthyChannels 过滤出健康的通道
//
// 该方法会检查通道列表中的每个通道，返回当前可用的通道
// 一个通道要被认为是健康的，其平台、模型和 API 密钥都必须是健康的
//
// 参数：
//   - channels: 通道列表
//   - now: 当前时间（如果为零值，则使用当前时间）
//
// 返回值：
//   - []*types.Channel: 健康的通道列表
func (m *Manager) FilterHealthyChannels(channels []*types.Channel, now time.Time) []*types.Channel {
	if now.IsZero() {
		now = time.Now()
	}
	return m.filter.FilterHealthyChannels(channels, now)
}

// IsHealthy 检查指定资源是否健康
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//   - now: 当前时间（如果为零值，则使用当前时间）
//
// 返回值：
//   - bool: 如果资源健康返回 true，否则返回 false
func (m *Manager) IsHealthy(resourceType types.ResourceType, resourceID uint, now time.Time) bool {
	if now.IsZero() {
		now = time.Now()
	}
	return m.filter.IsHealthy(resourceType, resourceID, now)
}

// ForceSync 强制立即执行同步
//
// 参数：
//   - ctx: 上下文
//
// 返回值：
//   - error: 错误信息
func (m *Manager) ForceSync(ctx context.Context) error {
	m.logger.Info("执行强制同步...")
	return m.syncer.Sync(ctx)
}

// GetHealthStats 获取健康状态统计信息
//
// 返回值：
//   - map[string]int: 各种状态的资源数量
func (m *Manager) GetHealthStats() map[string]int {
	stats := map[string]int{
		"total":       0,
		"available":   0,
		"warning":     0,
		"unavailable": 0,
		"unknown":     0,
	}

	m.cache.ForEach(func(key string, status *types.Health) bool {
		stats["total"]++
		switch status.Status {
		case types.HealthStatusAvailable:
			stats["available"]++
		case types.HealthStatusWarning:
			stats["warning"]++
		case types.HealthStatusUnavailable:
			stats["unavailable"]++
		case types.HealthStatusUnknown:
			stats["unknown"]++
		}
		return true
	})

	return stats
}
