package types

import (
	"time"
)

// RequestStat 表示单个请求的统计信息
type RequestStat struct {
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

// StatsQueryParams 表示统计查询参数
type StatsQueryParams struct {
	// 时间范围
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`

	// 可选过滤条件
	ModelName  *string `json:"model_name,omitempty"`
	Success    *bool   `json:"success,omitempty"` // 成功状态过滤
	PlatformID *uint   `json:"platform_id,omitempty"`
}

// StatsSummary 表示统计摘要
type StatsSummary struct {
	TotalRequests   int64 `json:"total_requests"`   // 总请求数
	SuccessRequests int64 `json:"success_requests"` // 成功请求数
}
