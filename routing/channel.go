package routing

import (
	"context"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/routing/health"
)

// Channel 表示一个完整的通道，包含平台、模型和密钥信息
type Channel struct {
	PlatformID uint
	ModelID    uint
	APIKeyID   uint

	PlatformType string // 平台类型（如 "openai", "anthropic"）
	APIEndpoint  string // API 端点
	ModelName    string // 模型名称
	APIKey       string // API 密钥

	// 健康管理器引用，用于更新状态
	healthService *health.Service
}

// MarkSuccess 标记通道调用成功
func (c *Channel) MarkSuccess(ctx context.Context) {
	if c.healthService == nil {
		return
	}

	// 更新平台级别的健康状态
	c.healthService.UpdateStatus(
		health.ResourceTypePlatform,
		c.PlatformID,
		true, // 成功
		"",   // 无错误信息
		0,    // 无错误码
	)

	// 更新模型级别的健康状态
	c.healthService.UpdateStatus(
		health.ResourceTypeModel,
		c.ModelID,
		true, // 成功
		"",   // 无错误信息
		0,    // 无错误码
	)

	// 更新 API 密钥级别的健康状态
	c.healthService.UpdateStatus(
		health.ResourceTypeAPIKey,
		c.APIKeyID,
		true, // 成功
		"",   // 无错误信息
		0,    // 无错误码
	)
}

// MarkFailure 标记通道调用失败
func (c *Channel) MarkFailure(ctx context.Context, err error) {
	if c.healthService == nil {
		return
	}

	errorCode := errors.GetHTTPStatus(err)
	errorMessage := errors.GetMessage(err)

	// 根据错误类型确定资源类型和资源 ID
	resourceType := health.ResourceTypeModel
	resourceID := c.ModelID

	switch errors.GetCode(err) {
	case errors.ErrCodeUnavailable:
		resourceType = health.ResourceTypePlatform
		resourceID = c.PlatformID
	case errors.ErrCodeAuthenticationFailed:
		resourceType = health.ResourceTypeAPIKey
		resourceID = c.APIKeyID
	}

	// 更新健康状态
	c.healthService.UpdateStatus(
		resourceType,
		resourceID,
		false,        // 失败
		errorMessage, // 错误信息
		errorCode,    // 错误码
	)
}

// IsHealthy 检查通道是否健康
func (c *Channel) IsHealthy() bool {
	if c.healthService == nil {
		return false // 如果没有健康服务，默认认为非健康
	}

	result := c.healthService.CheckChannelHealth(c.PlatformID, c.ModelID, c.APIKeyID)
	return result.Status == health.ChannelStatusAvailable
}

// LastCheckTime 获取最后检查时间
func (c *Channel) LastCheckTime() time.Time {
	if c.healthService == nil {
		return time.Now()
	}

	result := c.healthService.CheckChannelHealth(c.PlatformID, c.ModelID, c.APIKeyID)
	return result.LastCheckAt
}
