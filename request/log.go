package request

import (
	"context"
	"time"
)

// RequestLog 表示单个请求的统计信息
type RequestLog struct {
	ID uint `json:"id"` // 唯一标识符

	// 请求基本信息
	Timestamp         time.Time `json:"timestamp"`                     // 请求时间
	RequestType       string    `json:"request_type"`                  // 请求类型：stream 或 non-stream
	ModelName         string    `json:"model_name"`                    // 模型名称
	OriginalModelName string    `json:"original_model_name,omitempty"` // 原始模型名称（用户请求中的模型名称）

	// 通道信息
	PlatformID uint `json:"platform_id"` // 平台 ID
	APIKeyID   uint `json:"api_key_id"`  // 密钥 ID
	ModelID    uint `json:"model_id"`    // 模型 ID

	// 耗时信息
	Duration      time.Duration  `json:"duration"`                  // 总用时
	FirstByteTime *time.Duration `json:"first_byte_time,omitempty"` // 首字用时（仅流式）

	// 结果状态
	Success  bool    `json:"success"`             // 是否成功
	ErrorMsg *string `json:"error_msg,omitempty"` // 错误信息（失败时）

	// Token 使用统计
	PromptTokens     *int `json:"prompt_tokens"`     // 提示 Token 数
	CompletionTokens *int `json:"completion_tokens"` // 完成 Token 数
	TotalTokens      *int `json:"total_tokens"`      // 总 Token 数
}

// recordRequestLog 记录请求统计信息
func (p *Request) recordRequestLog(
	requestLog *RequestLog,
	firstByteTime *time.Time,
	success bool,
) {
	// 创建带有请求上下文的日志记录器
	log := p.logger.With(
		"platform_id", requestLog.PlatformID,
		"model_id", requestLog.ModelID,
		"api_key_id", requestLog.APIKeyID,
		"request_type", requestLog.RequestType,
		"model_name", requestLog.ModelName,
	)

	// 计算耗时
	requestDuration := time.Since(requestLog.Timestamp)
	requestLog.Duration = requestDuration

	// 如果记录了首字节时间，则计算首字节耗时
	if firstByteTime != nil && !firstByteTime.IsZero() {
		firstByteDuration := firstByteTime.Sub(requestLog.Timestamp)
		requestLog.FirstByteTime = &firstByteDuration

		log.Debug("记录请求统计信息",
			"duration", requestDuration.String(),
			"first_byte_time", firstByteDuration.String(),
			"success", success,
		)
	} else {
		log.Debug("记录请求统计信息",
			"duration", requestDuration.String(),
			"success", success,
		)
	}

	// 记录 Token 使用情况
	if requestLog.PromptTokens != nil && requestLog.CompletionTokens != nil && requestLog.TotalTokens != nil {
		log.Debug("Token 使用统计",
			"prompt_tokens", *requestLog.PromptTokens,
			"completion_tokens", *requestLog.CompletionTokens,
			"total_tokens", *requestLog.TotalTokens,
		)
	}

	// 记录错误信息
	if !success && requestLog.ErrorMsg != nil {
		log.Error("请求失败",
			"error", *requestLog.ErrorMsg,
		)
	}

	requestLog.Success = success

	// 保存到数据库
	err := p.repo.CreateRequestLog(context.Background(), requestLog)
	if err != nil {
		log.Error("保存请求日志失败", "error", err)
	} else {
		log.Debug("请求日志保存成功")
	}
}
