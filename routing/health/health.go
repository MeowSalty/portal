// Package health 提供了资源健康状态管理功能
package health

import (
	"strconv"
	"strings"
	"time"

	"github.com/MeowSalty/portal/errors"
)

// Service 管理所有资源的健康状态
type Service struct {
	storage Storage         // 存储接口
	backoff BackoffStrategy // 退避策略
	filter  Filter          // 健康状态过滤器
}

// Config 管理器配置
type Config struct {
	Storage      Storage         // 存储接口（必需）
	Backoff      BackoffStrategy // 退避策略（可选）
	AllowProbing bool            // 是否允许对 Unavailable 状态的资源进行探测（可选，默认 false）
}

// New 创建一个新的健康状态管理器
//
// 参数：
//   - cfg: 管理器配置
//
// 返回值：
//   - *Service: 健康状态管理器实例
//   - error: 错误信息
func New(cfg Config) (*Service, error) {
	// 验证必需参数
	if cfg.Storage == nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "存储接口不能为空")
	}

	if cfg.Backoff == nil {
		cfg.Backoff = DefaultBackoffStrategy()
	}

	// 创建过滤器
	filter := NewFilter(cfg.Storage, cfg.AllowProbing)

	m := &Service{
		storage: cfg.Storage,
		backoff: cfg.Backoff,
		filter:  filter,
	}

	return m, nil
}

// GetStatus 获取指定资源的健康状态
//
// 如果存储中不存在该资源的健康状态，则会创建一个新的健康状态对象
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//
// 返回值：
//   - *Health: 健康状态对象
//   - error: 错误信息
func (m *Service) GetStatus(resourceType ResourceType, resourceID uint) (*Health, error) {
	status, err := m.storage.Get(resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	if status != nil {
		return status, nil
	}

	// 如果存储中不存在，创建一个新的健康状态对象
	now := time.Now()
	status = &Health{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Status:       HealthStatusUnknown,
		LastCheckAt:  now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 存储到存储接口
	if err := m.storage.Set(status); err != nil {
		return nil, err
	}

	return status, nil
}

// UpdateStatus 更新指定资源的健康状态
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//   - success: 是否成功
//   - errorMessage: 错误信息（如果失败）
//   - errorCode: 错误代码（如果失败）
//
// 返回值：
//   - error: 错误信息
func (m *Service) UpdateStatus(
	resourceType ResourceType,
	resourceID uint,
	success bool,
	errorMessage string,
	errorCode int,
) error {
	// 获取或创建健康状态
	status, err := m.GetStatus(resourceType, resourceID)
	if err != nil {
		return err
	}

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

	// 保存到存储
	return m.storage.Set(status)
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
	LastCheckAt time.Time     // 最后检查时间（从平台、密钥、模型中取最新值）
}

// CheckChannelHealth 检查通道的可用性
//
// 使用平台 ID、模型 ID、密钥 ID 来检查通道的可用性
// 返回通道的状态：可用、不可用、未知，以及最后检查时间（从平台、密钥、模型中取最新值）
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

	// 获取所有资源的健康状态
	// 如果存储层出现错误，将状态视为 nil（未知）
	platformStatus, err := m.storage.Get(ResourceTypePlatform, platformID)
	if err != nil {
		platformStatus = nil
	}
	modelStatus, err := m.storage.Get(ResourceTypeModel, modelID)
	if err != nil {
		modelStatus = nil
	}
	apiKeyStatus, err := m.storage.Get(ResourceTypeAPIKey, apiKeyID)
	if err != nil {
		apiKeyStatus = nil
	}

	// 计算最后检查时间：从平台、密钥、模型中取最新的值
	lastCheckAt := getLatestCheckTime(now, platformStatus, modelStatus, apiKeyStatus)

	// 如果任何资源不健康，则通道状态为不可用
	if !platformHealthy || !modelHealthy || !apiKeyHealthy {
		return ChannelHealthResult{
			Status:      ChannelStatusUnavailable,
			LastCheckAt: lastCheckAt,
		}
	}

	// 所有资源都健康，检查是否有未知状态的资源
	if isUnknownStatus(platformStatus) || isUnknownStatus(modelStatus) || isUnknownStatus(apiKeyStatus) {
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

// getLatestCheckTime 从多个健康状态中获取最新的检查时间
//
// 如果所有状态都为 nil 或未知，则返回当前时间
//
// 参数：
//   - now: 当前时间
//   - statuses: 健康状态列表
//
// 返回值：
//   - time.Time: 最新的检查时间
func getLatestCheckTime(now time.Time, statuses ...*Health) time.Time {
	var latest time.Time
	hasKnownStatus := false

	for _, status := range statuses {
		if status != nil && status.Status != HealthStatusUnknown {
			hasKnownStatus = true
			if status.LastCheckAt.After(latest) {
				latest = status.LastCheckAt
			}
		}
	}

	// 如果所有状态都是未知或不存在，返回当前时间
	if !hasKnownStatus {
		return now
	}

	return latest
}

// isUnknownStatus 检查健康状态是否为未知
//
// 参数：
//   - status: 健康状态
//
// 返回值：
//   - bool: 如果状态为 nil 或状态为未知，返回 true
func isUnknownStatus(status *Health) bool {
	return status == nil || status.Status == HealthStatusUnknown
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
	platformStatus, err := m.GetStatus(ResourceTypePlatform, uint(platformID))
	if err != nil {
		return err
	}
	platformStatus.LastCheckAt = now
	platformStatus.UpdatedAt = now
	if err := m.storage.Set(platformStatus); err != nil {
		return err
	}

	// 更新模型资源的最后使用时间
	modelStatus, err := m.GetStatus(ResourceTypeModel, uint(modelID))
	if err != nil {
		return err
	}
	modelStatus.LastCheckAt = now
	modelStatus.UpdatedAt = now
	if err := m.storage.Set(modelStatus); err != nil {
		return err
	}

	// 更新 API 密钥资源的最后使用时间
	apiKeyStatus, err := m.GetStatus(ResourceTypeAPIKey, uint(apiKeyID))
	if err != nil {
		return err
	}
	apiKeyStatus.LastCheckAt = now
	apiKeyStatus.UpdatedAt = now
	return m.storage.Set(apiKeyStatus)
}

// ResetHealth 手动重置指定资源的健康状态
//
// 该方法用于将处于 Unavailable 状态的资源重置为可用状态
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//
// 返回值：
//   - error: 错误信息
func (m *Service) ResetHealth(resourceType ResourceType, resourceID uint) error {
	status, err := m.GetStatus(resourceType, resourceID)
	if err != nil {
		return err
	}
	now := time.Now()
	status.UpdatedAt = now
	m.backoff.Reset(status)
	return m.storage.Set(status)
}

// DisableHealth 手动将指定资源设置为不可用状态
//
// 该方法用于手动禁用资源，将其状态设置为 Unavailable
//
// 参数：
//   - resourceType: 资源类型
//   - resourceID: 资源 ID
//   - reason: 禁用原因
//
// 返回值：
//   - error: 错误信息
func (m *Service) DisableHealth(resourceType ResourceType, resourceID uint, reason string) error {
	status, err := m.GetStatus(resourceType, resourceID)
	if err != nil {
		return err
	}
	now := time.Now()
	status.Status = HealthStatusUnavailable
	status.LastError = reason
	status.LastCheckAt = now
	status.UpdatedAt = now
	status.NextAvailableAt = nil // 手动禁用不设置自动恢复时间
	return m.storage.Set(status)
}
