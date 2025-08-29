package types

import "time"

// HealthStatus 健康状态枚举
type HealthStatus string

const (
	HealthStatusUnknown     HealthStatus = "unknown"     // 未知
	HealthStatusAvailable   HealthStatus = "available"   // 可用
	HealthStatusWarning     HealthStatus = "warning"     // 警告（使用退避策略）
	HealthStatusUnavailable HealthStatus = "unavailable" // 不可用
)

// ResourceType 资源类型枚举
type ResourceType string

const (
	ResourceTypePlatform ResourceType = "platform" // 平台级
	ResourceTypeAPIKey   ResourceType = "api_key"  // 密钥级
	ResourceTypeModel    ResourceType = "model"    // 模型级
)

// Health 健康状态表 (health_status)
type Health struct {
	ID           uint
	ResourceType ResourceType // 资源类型
	ResourceID   uint         // 资源 ID

	// 关联资源 ID（作用域隔离）
	RelatedPlatformID *uint // 关联的平台 ID（可选）
	RelatedAPIKeyID   *uint // 关联的密钥 ID（可选）

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

// RateLimitConfig 定义了限流配置
type RateLimitConfig struct {
	RPM int // Requests Per Minute
	TPM int // Tokens Per Minute
}

// Platform 表示一个 AI 平台（例如 OpenAI、Anthropic）
type Platform struct {
	ID        uint
	Name      string
	Format    string
	BaseURL   string
	RateLimit RateLimitConfig
}

// Model 表示平台上的一个具体模型
type Model struct {
	ID         uint
	PlatformID uint
	Name       string
	Alias      string
}

// APIKey 表示平台的 API 密钥
type APIKey struct {
	ID    uint
	Value string
}

// Channel 是一个临时结构体，表示完整的请求路径（平台、模型、API 密钥）
type Channel struct {
	Platform *Platform
	Model    *Model
	APIKey   *APIKey
}
