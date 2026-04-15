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

// HealthImpact 定义失败对健康状态的影响程度。
type HealthImpact int8

const (
	// HealthImpactFull 完全降级：计入错误计数并应用退避策略（默认行为）。
	HealthImpactFull HealthImpact = iota
	// HealthImpactRecoverable 可恢复失败：记录错误信息但不增加错误计数，仅标记为警告。
	HealthImpactRecoverable
	// HealthImpactNone 无健康影响：不更新健康状态。
	HealthImpactNone
)

// Health 健康状态表 (health_status)
type Health struct {
	ResourceType ResourceType // 资源类型
	ResourceID   uint         // 资源 ID

	Status HealthStatus // 健康状态

	// 指数退避相关
	RetryCount      int        // 重试次数
	NextAvailableAt *time.Time // 下次可用时间
	BackoffDuration int64      // 当前退避时长 (秒)

	// 状态详情
	// Deprecated: LastError 为展示型历史字段，语义不稳定；请优先使用 LastErrorMessage。
	LastError string // 最后错误信息（历史兼容）
	// Deprecated: LastErrorCode 为历史兼容字段；存在 HTTP 状态码时写入状态码，否则写入 0。
	LastErrorCode int // 最后错误码（历史兼容）

	LastErrorMessage        string // 最后错误展示消息
	LastStructuredErrorCode string // 最后稳定错误码
	LastHTTPStatus          *int   // 最后 HTTP 状态码（无 HTTP 响应时为空）
	LastErrorFrom           string // 最后错误来源
	LastCauseMessage        string // 最后根因文本

	LastCheckAt   time.Time  // 最后检查时间
	LastSuccessAt *time.Time // 最后成功时间

	// 统计信息
	SuccessCount int // 成功次数
	ErrorCount   int // 错误次数

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ErrorSnapshot 表示健康状态写入所需的轻量错误摘要。
type ErrorSnapshot struct {
	Message      string       // 展示消息
	Code         string       // 稳定错误码
	HTTPStatus   *int         // HTTP 状态码
	ErrorFrom    string       // 错误来源
	CauseMessage string       // 根因文本
	Impact       HealthImpact // 健康影响程度
}
