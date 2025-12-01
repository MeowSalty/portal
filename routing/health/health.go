// Package health 提供了资源健康状态管理功能
package health

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/MeowSalty/portal/errors"
)

// Service 管理所有资源的健康状态
type Service struct {
	repo    HealthRepository // 数据仓库接口
	cache   Cache            // 健康状态缓存
	syncer  Syncer           // 后台同步器
	backoff BackoffStrategy  // 退避策略
	filter  Filter           // 健康状态过滤器
}

// Config 管理器配置
type Config struct {
	Repo         HealthRepository // 数据仓库接口（必需）
	SyncInterval time.Duration    // 同步间隔（可选，默认 1 分钟）
	Backoff      BackoffStrategy  // 退避策略（可选）
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
func New(ctx context.Context, cfg Config) (*Service, error) {
	// 验证必需参数
	if cfg.Repo == nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "数据仓库不能为空")
	}

	if cfg.SyncInterval <= 0 {
		cfg.SyncInterval = time.Minute
	}
	if cfg.Backoff == nil {
		cfg.Backoff = DefaultBackoffStrategy()
	}

	// 创建各个组件
	cache := NewCache()
	syncer := NewSyncer(cfg.Repo, cache, cfg.SyncInterval)
	filter := NewFilter(cache)

	m := &Service{
		repo:    cfg.Repo,
		cache:   cache,
		syncer:  syncer,
		backoff: cfg.Backoff,
		filter:  filter,
	}

	// 加载初始数据
	if err := m.loadInitialData(ctx); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "加载初始健康数据失败", err)
	}

	// 启动后台同步器
	m.syncer.Start()

	return m, nil
}

// loadInitialData 从仓库获取所有健康状态并填充缓存
//
// 参数：
//   - ctx: 上下文
//
// 返回值：
//   - error: 错误信息
func (m *Service) loadInitialData(ctx context.Context) error {
	statuses, err := m.repo.GetAllHealth(ctx)
	if err != nil {
		return err
	}

	m.cache.LoadAll(statuses)
	return nil
}

// Shutdown 关闭健康状态管理器
//
// 此方法会停止后台同步进程并等待所有操作完成
func (m *Service) Shutdown() {
	m.syncer.Stop()
}

// getStatus 获取指定资源的健康状态
//
// 如果缓存中不存在该资源的健康状态，则会创建一个新的健康状态对象
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//
// 返回值：
//   - *Health: 健康状态对象
func (m *Service) getStatus(resourceType ResourceType, resourceID uint) *Health {
	if status, exists := m.cache.Get(resourceType, resourceID); exists {
		return status
	}

	// 如果缓存中不存在，创建一个新的健康状态对象
	now := time.Now()
	status := &Health{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Status:       HealthStatusUnknown,
		LastCheckAt:  now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 存储到缓存并标记为脏数据
	m.cache.Set(resourceType, resourceID, status)
	m.syncer.MarkDirty(resourceType, resourceID, status)

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
func (m *Service) UpdateStatus(
	resourceType ResourceType,
	resourceID uint,
	success bool,
	errorMessage string,
	errorCode int,
) {
	// 获取或创建健康状态
	status := m.getStatus(resourceType, resourceID)

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
	} else {
		// 失败情况：应用退避策略
		status.ErrorCount++
		status.LastError = errorMessage
		status.LastErrorCode = errorCode

		// 使用退避策略更新状态
		m.backoff.Apply(status)
	}

	// 标记为脏数据以便后续同步
	m.syncer.MarkDirty(resourceType, resourceID, status)
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
func (m *Service) IsHealthy(resourceType ResourceType, resourceID uint, now time.Time) bool {
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
func (m *Service) ForceSync(ctx context.Context) error {
	return m.syncer.Sync(ctx)
}

// ChannelStatus 通道健康状态
type ChannelStatus int8

const (
	ChannelStatusUnknown     ChannelStatus = iota // 未知
	ChannelStatusAvailable                        // 可用
	ChannelStatusUnavailable                      // 不可用
)

// ChannelHealthResult 通道健康检查结果
type ChannelHealthResult struct {
	Status      ChannelStatus // 通道状态
	LastCheckAt time.Time     // 最后检查时间（基于模型）
}

// CheckChannelHealth 检查通道的可用性
//
// 使用平台 ID、模型 ID、密钥 ID 来检查通道的可用性
// 返回通道的状态：可用、不可用、未知，以及模型的最后检查时间
//
// 规则：
//   - 如果平台、模型、密钥中有任意一个不健康，则返回不可用
//   - 如果平台、模型、密钥都健康，则返回可用
//   - 如果所有资源状态已知且至少有一个为未知状态，则返回未知
//
// 参数：
//   - platformID: 平台 ID
//   - modelID: 模型 ID
//   - apiKeyID: API 密钥 ID
//
// 返回值：
//   - ChannelHealthResult: 通道健康检查结果，包括状态和最后检查时间
func (m *Service) CheckChannelHealth(platformID, modelID, apiKeyID uint) ChannelHealthResult {
	now := time.Now()

	// 检查所有资源是否都健康
	platformHealthy := m.filter.IsHealthy(ResourceTypePlatform, platformID, now)
	modelHealthy := m.filter.IsHealthy(ResourceTypeModel, modelID, now)
	apiKeyHealthy := m.filter.IsHealthy(ResourceTypeAPIKey, apiKeyID, now)

	// 获取模型的最后检查时间
	modelStatus, modelExists := m.cache.Get(ResourceTypeModel, modelID)
	lastCheckAt := now
	if modelExists {
		lastCheckAt = modelStatus.LastCheckAt
	}

	// 如果任何资源不健康，则通道状态为不可用
	if !platformHealthy || !modelHealthy || !apiKeyHealthy {
		return ChannelHealthResult{
			Status:      ChannelStatusUnavailable,
			LastCheckAt: lastCheckAt,
		}
	}

	// 所有资源都健康，检查是否有未知状态的资源
	platformStatus, platformExists := m.cache.Get(ResourceTypePlatform, platformID)
	apiKeyStatus, apiKeyExists := m.cache.Get(ResourceTypeAPIKey, apiKeyID)

	// 如果所有资源都健康但存在未知状态的资源，则通道状态为未知
	if (!platformExists || platformStatus.Status == HealthStatusUnknown) ||
		(!modelExists || modelStatus.Status == HealthStatusUnknown) ||
		(!apiKeyExists || apiKeyStatus.Status == HealthStatusUnknown) {
		return ChannelHealthResult{
			Status:      ChannelStatusUnknown,
			LastCheckAt: lastCheckAt,
		}
	}

	// 所有资源都健康且已知，通道可用
	return ChannelHealthResult{
		Status:      ChannelStatusAvailable,
		LastCheckAt: lastCheckAt,
	}
}

// UpdateLastUsed 更新通道的最后使用时间
//
// 该方法用于在通道被选择后立即更新其最后使用时间
//
// 参数：
//   - channelID: 通道 ID，格式为 "platformID-modelID-apiKeyID"
//
// 返回值：
//   - error: 错误信息
func (m *Service) UpdateLastUsed(channelID string) error {
	// 解析通道 ID
	parts := strings.Split(channelID, "-")
	if len(parts) != 3 {
		return errors.New(errors.ErrCodeInvalidArgument, "无效的通道 ID 格式")
	}

	// 解析所有资源 ID
	platformID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return errors.Wrap(errors.ErrCodeInvalidArgument, "解析平台 ID 失败", err)
	}
	modelID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return errors.Wrap(errors.ErrCodeInvalidArgument, "解析模型 ID 失败", err)
	}
	apiKeyID, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return errors.Wrap(errors.ErrCodeInvalidArgument, "解析 API 密钥 ID 失败", err)
	}

	now := time.Now()

	// 更新平台资源的最后使用时间
	platformStatus := m.getStatus(ResourceTypePlatform, uint(platformID))
	platformStatus.LastCheckAt = now
	platformStatus.UpdatedAt = now
	m.syncer.MarkDirty(ResourceTypePlatform, uint(platformID), platformStatus)

	// 更新模型资源的最后使用时间
	modelStatus := m.getStatus(ResourceTypeModel, uint(modelID))
	modelStatus.LastCheckAt = now
	modelStatus.UpdatedAt = now
	m.syncer.MarkDirty(ResourceTypeModel, uint(modelID), modelStatus)

	// 更新 API 密钥资源的最后使用时间
	apiKeyStatus := m.getStatus(ResourceTypeAPIKey, uint(apiKeyID))
	apiKeyStatus.LastCheckAt = now
	apiKeyStatus.UpdatedAt = now
	m.syncer.MarkDirty(ResourceTypeAPIKey, uint(apiKeyID), apiKeyStatus)

	return nil
}
