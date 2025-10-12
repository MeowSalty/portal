package request

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// RequestLog 表示单个请求的统计信息
type RequestLog struct {
	ID string `json:"id"` // 唯一标识符

	// 请求基本信息
	Timestamp   time.Time   `json:"timestamp"`    // 请求时间
	RequestType string      `json:"request_type"` // 请求类型：stream 或 non-stream
	ModelName   string      `json:"model_name"`   // 模型名称
	ChannelInfo ChannelInfo `json:"channel_info"` // 通道信息

	// 耗时信息
	Duration      time.Duration  `json:"duration"`                  // 总用时
	FirstByteTime *time.Duration `json:"first_byte_time,omitempty"` // 首字用时（仅流式）

	// 结果状态
	Success  bool    `json:"success"`             // 是否成功
	ErrorMsg *string `json:"error_msg,omitempty"` // 错误信息（失败时）
}

// ChannelInfo 表示通道信息
type ChannelInfo struct {
	PlatformID uint `json:"platform_id"` // 平台 ID
	APIKeyID   uint `json:"api_key_id"`  // 密钥 ID
	ModelID    uint `json:"model_id"`    // 模型 ID
}

// recordRequestLog 记录请求统计信息
func (p *Request) recordRequestLog(
	requestLog *RequestLog,
	firstByteTime *time.Time,
	success bool,
) {
	// 计算耗时
	requestDuration := time.Since(requestLog.Timestamp)
	requestLog.Duration = requestDuration

	// 如果记录了首字节时间，则计算首字节耗时
	if firstByteTime != nil && !firstByteTime.IsZero() {
		firstByteDuration := firstByteTime.Sub(requestLog.Timestamp)
		requestLog.FirstByteTime = &firstByteDuration
	}
	requestLog.Success = success

	p.createRequestLog(context.Background(), requestLog)
}

// createRequestLog 安全地记录统计信息
func (p *Request) createRequestLog(ctx context.Context, log *RequestLog) {
	log.ID = generateID()
	p.repo.CreateRequestLog(ctx, log) // TODO：添加错误处理
}

// generateID 生成唯一的 ID
func generateID() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%d%d", time.Now().UnixNano(), n)
}
