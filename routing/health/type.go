package health

import "time"

// HealthStatus 健康状态枚举
type HealthStatus int8

const (
	HealthStatusUnknown     HealthStatus = iota // 未知
	HealthStatusAvailable                       // 可用
	HealthStatusWarning                         // 警告（使用退避策略）
	HealthStatusUnavailable                     // 不可用
)

// ResourceType 资源类型枚举
type ResourceType int8

const (
	ResourceTypePlatform ResourceType = iota + 1 // 平台级
	ResourceTypeAPIKey                           // 密钥级
	ResourceTypeModel                            // 模型级
)

// Health 健康状态表 (health_status)
type Health struct {
	ID uint

	ResourceType ResourceType // 资源类型
	ResourceID   uint         // 资源 ID

	Status HealthStatus // 健康状态

	// 指数退避相关
	RetryCount      int        // 重试次数
	NextAvailableAt *time.Time // 下次可用时间
	BackoffDuration int64      // 当前退避时长 (秒)

	// 状态详情
	LastError     string     // 最后错误信息
	LastErrorCode int        // 最后错误码
	LastCheckAt   time.Time  // 最后检查时间
	LastSuccessAt *time.Time // 最后成功时间

	// 统计信息
	SuccessCount int // 成功次数
	ErrorCount   int // 错误次数

	CreatedAt time.Time
	UpdatedAt time.Time
}
