// Package health 提供了资源健康状态管理功能
package health

import (
	"math"
	"time"
)

// BackoffStrategy 定义退避策略接口
type BackoffStrategy interface {
	// Apply 应用退避策略到健康状态
	Apply(status *Health)
	// Reset 重置退避状态
	Reset(status *Health)
}

// baseBackoff 包含退避策略的公共逻辑
type baseBackoff struct {
	maxDelay time.Duration // 最大退避时间
}

// applyBackoff 应用退避策略的公共逻辑
func (b *baseBackoff) applyBackoff(status *Health, delay time.Duration) {
	// 增加重试次数
	status.RetryCount++

	// 限制最大延迟
	if delay > b.maxDelay {
		delay = b.maxDelay
	}

	// 转换为秒存储
	status.BackoffDuration = int64(delay.Seconds())

	// 计算下次可用时间
	nextAvailable := time.Now().Add(delay)
	status.NextAvailableAt = &nextAvailable

	// 更新状态为警告（使用退避策略）
	status.Status = HealthStatusWarning
}

// resetBackoff 重置退避状态的公共逻辑
func (b *baseBackoff) resetBackoff(status *Health) {
	status.RetryCount = 0
	status.NextAvailableAt = nil
	status.BackoffDuration = 0
	status.Status = HealthStatusAvailable
}

// ExponentialBackoff 实现了指数退避策略
type ExponentialBackoff struct {
	baseBackoff
	initialDelay time.Duration // 初始退避时间
	multiplier   float64       // 退避倍数
}

// NewExponentialBackoff 创建一个新的指数退避策略实例
//
// 参数：
//   - initialDelay: 初始退避时间（默认 30 秒）
//   - maxDelay: 最大退避时间（默认 24 小时）
//   - multiplier: 退避倍数（默认 2）
func NewExponentialBackoff(initialDelay, maxDelay time.Duration, multiplier float64) BackoffStrategy {
	if initialDelay <= 0 {
		initialDelay = 30 * time.Second
	}
	if maxDelay <= 0 {
		maxDelay = 24 * time.Hour
	}
	if multiplier <= 1 {
		multiplier = 2
	}
	return &ExponentialBackoff{
		baseBackoff:  baseBackoff{maxDelay: maxDelay},
		initialDelay: initialDelay,
		multiplier:   multiplier,
	}
}

// Apply 应用指数退避策略
func (b *ExponentialBackoff) Apply(status *Health) {
	delay := b.calculateDelay(status.RetryCount + 1)
	b.applyBackoff(status, delay)
}

// Reset 重置退避状态
func (b *ExponentialBackoff) Reset(status *Health) {
	b.resetBackoff(status)
}

// calculateDelay 计算当前重试次数对应的退避时间
func (b *ExponentialBackoff) calculateDelay(retryCount int) time.Duration {
	if retryCount <= 1 {
		return b.initialDelay
	}
	// 使用 math.Pow 计算指数退避时间
	delay := time.Duration(float64(b.initialDelay) * math.Pow(b.multiplier, float64(retryCount-1)))
	return delay
}

// LinearBackoff 实现了线性退避策略
type LinearBackoff struct {
	baseBackoff
	baseDelay time.Duration // 基础退避时间
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
		baseBackoff: baseBackoff{maxDelay: maxDelay},
		baseDelay:   baseDelay,
	}
}

// Apply 应用线性退避策略
func (b *LinearBackoff) Apply(status *Health) {
	delay := time.Duration(status.RetryCount+1) * b.baseDelay
	b.applyBackoff(status, delay)
}

// Reset 重置退避状态
func (b *LinearBackoff) Reset(status *Health) {
	b.resetBackoff(status)
}

// DefaultBackoffStrategy 返回默认的退避策略
//
// 使用默认参数：
//   - 初始延迟：1 秒
//   - 最大延迟：5 分钟
//   - 倍数：2
func DefaultBackoffStrategy() BackoffStrategy {
	return NewExponentialBackoff(30*time.Second, 24*time.Hour, 2)
}
