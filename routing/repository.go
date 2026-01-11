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
	Provider      string
	Variant       string
	BaseURL       string
	RateLimit     RateLimitConfig
	CustomHeaders map[string]string // 平台级别的自定义 HTTP 头部
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

type PlatformRepository interface {
	GetPlatformByID(ctx context.Context, id uint) (*Platform, error)
}

type ModelRepository interface {
	GetModelByID(ctx context.Context, id uint) (Model, error)
	FindModelsByNameOrAlias(ctx context.Context, name string) ([]Model, error)
}

type KeyRepository interface {
	GetAllAPIKeysByPlatformID(ctx context.Context, platformID uint) ([]*APIKey, error)
}
