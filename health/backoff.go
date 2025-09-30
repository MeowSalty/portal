// Package health 提供了资源健康状态管理功能
package health

import (
	"time"

	"github.com/MeowSalty/portal/types"
)

// BackoffStrategy 定义退避策略接口
type BackoffStrategy interface {
	// Apply 应用退避策略到健康状态
	Apply(status *types.Health)
	// Reset 重置退避状态
	Reset(status *types.Health)
}

// ExponentialBackoff 实现了指数退避策略
type ExponentialBackoff struct {
	initialDelay time.Duration // 初始退避时间
	maxDelay     time.Duration // 最大退避时间
	multiplier   float64       // 退避倍数
}

// NewExponentialBackoff 创建一个新的指数退避策略实例
//
// 参数：
//   - initialDelay: 初始退避时间（默认 1 秒）
//   - maxDelay: 最大退避时间（默认 5 分钟）
//   - multiplier: 退避倍数（默认 2）
func NewExponentialBackoff(initialDelay, maxDelay time.Duration, multiplier float64) BackoffStrategy {
	if initialDelay <= 0 {
		initialDelay = time.Second
	}
	if maxDelay <= 0 {
		maxDelay = 5 * time.Minute
	}
	if multiplier <= 1 {
		multiplier = 2
	}
	return &ExponentialBackoff{
		initialDelay: initialDelay,
		maxDelay:     maxDelay,
		multiplier:   multiplier,
	}
}

// Apply 应用指数退避策略
//
// 每次失败后，退避时间会按照 multiplier 倍数增长，
// 直到达到 maxDelay 的上限
func (b *ExponentialBackoff) Apply(status *types.Health) {
	// 增加重试次数
	status.RetryCount++

	// 计算退避时长
	delay := b.calculateDelay(status.RetryCount)

	// 转换为秒存储
	status.BackoffDuration = int64(delay.Seconds())

	// 计算下次可用时间
	nextAvailable := time.Now().Add(delay)
	status.NextAvailableAt = &nextAvailable

	// 更新状态为警告（使用退避策略）
	status.Status = types.HealthStatusWarning
}

// Reset 重置退避状态
//
// 当操作成功时调用，清除所有退避相关的状态
func (b *ExponentialBackoff) Reset(status *types.Health) {
	status.RetryCount = 0
	status.NextAvailableAt = nil
	status.BackoffDuration = 0
	status.Status = types.HealthStatusAvailable
}

// calculateDelay 计算当前重试次数对应的退避时间
func (b *ExponentialBackoff) calculateDelay(retryCount int) time.Duration {
	if retryCount <= 0 {
		return b.initialDelay
	}

	// 计算指数退避时间
	delay := b.initialDelay
	for i := 1; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * b.multiplier)
		if delay > b.maxDelay {
			return b.maxDelay
		}
	}

	return delay
}

// DefaultBackoffStrategy 返回默认的退避策略
//
// 使用默认参数：
//   - 初始延迟：1 秒
//   - 最大延迟：5 分钟
//   - 倍数：2
func DefaultBackoffStrategy() BackoffStrategy {
	return NewExponentialBackoff(time.Second, 5*time.Minute, 2)
}

// LinearBackoff 实现了线性退避策略
type LinearBackoff struct {
	baseDelay time.Duration // 基础退避时间
	maxDelay  time.Duration // 最大退避时间
}

// NewLinearBackoff 创建一个新的线性退避策略实例
func NewLinearBackoff(baseDelay, maxDelay time.Duration) BackoffStrategy {
	if baseDelay <= 0 {
		baseDelay = time.Second
	}
	if maxDelay <= 0 {
		maxDelay = 5 * time.Minute
	}
	return &LinearBackoff{
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
	}
}

// Apply 应用线性退避策略
func (b *LinearBackoff) Apply(status *types.Health) {
	// 增加重试次数
	status.RetryCount++

	// 计算退避时长（线性增长）
	delay := time.Duration(status.RetryCount) * b.baseDelay
	if delay > b.maxDelay {
		delay = b.maxDelay
	}

	// 转换为秒存储
	status.BackoffDuration = int64(delay.Seconds())

	// 计算下次可用时间
	nextAvailable := time.Now().Add(delay)
	status.NextAvailableAt = &nextAvailable

	// 更新状态为警告
	status.Status = types.HealthStatusWarning
}

// Reset 重置退避状态
func (b *LinearBackoff) Reset(status *types.Health) {
	status.RetryCount = 0
	status.NextAvailableAt = nil
	status.BackoffDuration = 0
	status.Status = types.HealthStatusAvailable
}
