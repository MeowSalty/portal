package routing

import (
	"context"
)

// RateLimitConfig 定义了限流配置
type RateLimitConfig struct {
	RPM int // Requests Per Minute
	TPM int // Tokens Per Minute
}

// Platform 表示一个 AI 平台（例如 OpenAI、Anthropic）
type Platform struct {
	ID            uint
	Name          string
	BaseURL       string
	RateLimit     RateLimitConfig
	CustomHeaders map[string]string // 平台级别的自定义 HTTP 头部
}

// Endpoint 表示平台的端点配置
// 每个平台有一个默认端点，定义了访问模型的具体路径和配置
type Endpoint struct {
	ID              uint              // 端点 ID
	EndpointType    string            // 端点类型（如 "openai", "anthropic"）
	EndpointVariant string            // 端点变体（如 "chat", "responses"）
	Path            string            // API 路径（如 "/v1/chat/completions"）
	CustomHeaders   map[string]string // 端点级别的自定义 HTTP 头部
}

// Model 表示平台上的一个具体模型
type Model struct {
	ID         uint
	PlatformID uint
	Name       string
	Alias      string
	APIKeys    []APIKey // 模型关联的密钥（多对多关系）
}

// APIKey 表示平台的 API 密钥
type APIKey struct {
	ID    uint
	Value string
}

// ModelWithEndpoint 包含模型、平台和端点的完整信息
// 用于构建 Channel 所需的所有配置
type ModelWithEndpoint struct {
	Model    Model    // 模型信息
	Platform Platform // 平台信息
	Endpoint Endpoint // 端点信息（默认端点或指定端点）
}

// PlatformRepository 平台存储接口
type PlatformRepository interface {
	GetPlatformByID(ctx context.Context, id uint) (*Platform, error)
}

// ModelRepository 模型存储接口
type ModelRepository interface {
	// FindModelsWithDefaultEndpoint 通过模型名称查找，返回带有平台和默认端点的完整信息
	// 如果平台没有默认端点，返回错误
	FindModelsWithDefaultEndpoint(ctx context.Context, name string) ([]ModelWithEndpoint, error)

	// FindModelsWithEndpoint 通过模型名称 + 端点类型 + 变体查找
	FindModelsWithEndpoint(ctx context.Context, name, endpointType, endpointVariant string) ([]ModelWithEndpoint, error)
}

type KeyRepository interface {
	GetAllAPIKeysByPlatformID(ctx context.Context, platformID uint) ([]*APIKey, error)
}
