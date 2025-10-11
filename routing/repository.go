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
