// Package health 提供了资源健康状态管理功能
package health

import (
	"time"

	"github.com/MeowSalty/portal/errors"
)

// Service 管理所有资源的健康状态
type Service struct {
	storage Storage         // 存储接口
	backoff BackoffStrategy // 退避策略
	filter  Filter          // 健康状态过滤器
	// allowProbing 控制 Unavailable 状态在退避结束后是否允许探测。
	// 该配置与 filter 保持一致，用于避免重复存储读取时在 Service 层复用同一判定语义。
	allowProbing bool
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
		storage:      cfg.Storage,
		backoff:      cfg.Backoff,
		filter:       filter,
		allowProbing: cfg.AllowProbing,
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
//   - snapshot: 错误摘要（失败时写入，成功时忽略）
//
// 返回值：
//   - error: 错误信息
func (m *Service) UpdateStatus(
	resourceType ResourceType,
	resourceID uint,
	success bool,
	snapshot ErrorSnapshot,
) error {
	// 无健康影响的失败不更新健康状态
	if !success && snapshot.Impact == HealthImpactNone {
		return nil
	}

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
		status.LastErrorMessage = ""
		status.LastStructuredErrorCode = ""
		status.LastHTTPStatus = nil
		status.LastErrorFrom = ""
		status.LastCauseMessage = ""
		status.ErrorCount = 0

		// 使用退避策略重置状态
		m.backoff.Reset(status)
	} else if snapshot.Impact == HealthImpactRecoverable {
		// 可恢复失败：记录错误信息但不增加错误计数，仅标记为警告
		status.LastErrorMessage = snapshot.Message
		status.LastStructuredErrorCode = snapshot.Code
		status.LastHTTPStatus = snapshot.HTTPStatus
		status.LastErrorFrom = snapshot.ErrorFrom
		status.LastCauseMessage = snapshot.CauseMessage

		// 历史字段兼容写入
		status.LastError = status.LastErrorMessage
		if status.LastHTTPStatus != nil {
			status.LastErrorCode = *status.LastHTTPStatus
		} else {
			status.LastErrorCode = 0
		}

		// 可恢复失败仅标记为警告，不增加错误计数，不应用退避
		if status.Status != HealthStatusUnavailable {
			status.Status = HealthStatusWarning
		}
	} else {
		// 完全降级失败：计入错误计数并应用退避策略
		status.ErrorCount++
		status.LastErrorMessage = snapshot.Message
		status.LastStructuredErrorCode = snapshot.Code
		status.LastHTTPStatus = snapshot.HTTPStatus
		status.LastErrorFrom = snapshot.ErrorFrom
		status.LastCauseMessage = snapshot.CauseMessage

		// 历史字段兼容写入
		status.LastError = status.LastErrorMessage
		if status.LastHTTPStatus != nil {
			status.LastErrorCode = *status.LastHTTPStatus
		} else {
			status.LastErrorCode = 0
		}

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
	platformStatus, modelStatus, apiKeyStatus := m.getChannelStatuses(platformID, modelID, apiKeyID)
	return m.evaluateChannelHealth(now, platformStatus, modelStatus, apiKeyStatus)
}

// GetChannelHealthAndLastTryTimes 一次性获取通道健康状态与平台/模型/密钥最近尝试时间。
//
// 该方法用于选路热路径，避免先检查健康再二次读取最近尝试时间导致的重复 I/O。
func (m *Service) GetChannelHealthAndLastTryTimes(
	platformID,
	modelID,
	apiKeyID uint,
) (ChannelHealthResult, time.Time, time.Time, time.Time) {
	now := time.Now()
	platformStatus, modelStatus, apiKeyStatus := m.getChannelStatuses(platformID, modelID, apiKeyID)
	result := m.evaluateChannelHealth(now, platformStatus, modelStatus, apiKeyStatus)

	platformLastTry := now
	if platformStatus != nil {
		platformLastTry = platformStatus.LastCheckAt
	}

	modelLastTry := now
	if modelStatus != nil {
		modelLastTry = modelStatus.LastCheckAt
	}

	keyLastTry := now
	if apiKeyStatus != nil {
		keyLastTry = apiKeyStatus.LastCheckAt
	}

	return result, platformLastTry, modelLastTry, keyLastTry
}

// getChannelStatuses 获取平台/模型/密钥状态。
// 如果存储层读取失败，则将对应状态视为 nil（未知）。
func (m *Service) getChannelStatuses(platformID, modelID, apiKeyID uint) (*Health, *Health, *Health) {
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

	return platformStatus, modelStatus, apiKeyStatus
}

// evaluateChannelHealth 基于资源状态计算通道健康结果。
func (m *Service) evaluateChannelHealth(now time.Time, platformStatus, modelStatus, apiKeyStatus *Health) ChannelHealthResult {
	platformHealthy := m.isResourceHealthyByStatus(platformStatus, now)
	modelHealthy := m.isResourceHealthyByStatus(modelStatus, now)
	apiKeyHealthy := m.isResourceHealthyByStatus(apiKeyStatus, now)

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

// isResourceHealthyByStatus 在已获取资源状态的前提下执行健康判定。
func (m *Service) isResourceHealthyByStatus(status *Health, now time.Time) bool {
	if status == nil {
		// 如果存储中不存在或读取失败，视为未知状态（可以尝试）
		return true
	}

	switch status.Status {
	case HealthStatusAvailable, HealthStatusUnknown:
		return true
	case HealthStatusWarning:
		return status.NextAvailableAt != nil && now.After(*status.NextAvailableAt)
	case HealthStatusUnavailable:
		return m.allowProbing && status.NextAvailableAt != nil && now.After(*status.NextAvailableAt)
	default:
		return false
	}
}

// GetLastTryTimes 获取平台/模型/密钥的最近尝试时间
//
// 参数：
//   - platformID: 平台 ID
//   - modelID: 模型 ID
//   - apiKeyID: API 密钥 ID
//
// 返回值：
//   - time.Time: 平台最近尝试时间
//   - time.Time: 模型最近尝试时间
//   - time.Time: 密钥最近尝试时间
//   - error: 错误信息
func (m *Service) GetLastTryTimes(platformID, modelID, apiKeyID uint) (time.Time, time.Time, time.Time, error) {
	platformStatus, err := m.GetStatus(ResourceTypePlatform, platformID)
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, err
	}
	modelStatus, err := m.GetStatus(ResourceTypeModel, modelID)
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, err
	}
	apiKeyStatus, err := m.GetStatus(ResourceTypeAPIKey, apiKeyID)
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, err
	}

	return platformStatus.LastCheckAt, modelStatus.LastCheckAt, apiKeyStatus.LastCheckAt, nil
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

// UpdateLastTry 更新通道的最近尝试时间
//
// 该方法用于在通道被选择后立即更新其平台/模型/密钥的最近尝试时间
//
// 参数：
//   - platformID: 平台 ID
//   - modelID: 模型 ID
//   - apiKeyID: API 密钥 ID
//
// 返回值：
//   - error: 错误信息
func (m *Service) UpdateLastTry(platformID, modelID, apiKeyID uint) error {
	now := time.Now()

	// 更新平台资源的最后使用时间
	platformStatus, err := m.GetStatus(ResourceTypePlatform, platformID)
	if err != nil {
		return err
	}
	platformStatus.LastCheckAt = now
	platformStatus.UpdatedAt = now
	if err := m.storage.Set(platformStatus); err != nil {
		return err
	}

	// 更新模型资源的最后使用时间
	modelStatus, err := m.GetStatus(ResourceTypeModel, modelID)
	if err != nil {
		return err
	}
	modelStatus.LastCheckAt = now
	modelStatus.UpdatedAt = now
	if err := m.storage.Set(modelStatus); err != nil {
		return err
	}

	// 更新 API 密钥资源的最后使用时间
	apiKeyStatus, err := m.GetStatus(ResourceTypeAPIKey, apiKeyID)
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
