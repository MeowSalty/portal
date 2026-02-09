package request

import (
	"context"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
)

// ChatCompletion 处理聊天完成请求
func (p *Request) ChatCompletion(
	ctx context.Context,
	request *types.RequestContract,
	channel *routing.Channel,
) (*types.ResponseContract, error) {
	now := time.Now()

	// 创建带有请求上下文的日志记录器
	log := p.logger.With(
		"platform_type", channel.Provider,
		"platform_id", channel.PlatformID,
		"model_id", channel.ModelID,
		"api_key_id", channel.APIKeyID,
		"model_name", channel.ModelName,
		"original_model", request.Model,
	)

	log.DebugContext(ctx, "开始处理聊天完成请求")

	// 获取适配器
	log.DebugContext(ctx, "获取适配器", "format", channel.Provider)
	adapter, err := p.getAdapter(channel.Provider)
	if err != nil {
		log.ErrorContext(ctx, "获取适配器失败", "error", err, "format", channel.Provider)
		return nil, errors.Wrap(errors.ErrCodeAdapterNotFound, "获取适配器失败", err).
			WithContext("format", channel.Provider)
	}
	log.DebugContext(ctx, "获取适配器成功", "adapter", adapter.Name())

	// 创建请求日志
	requestLog := &RequestLog{
		Timestamp:         now,
		RequestType:       "non-stream",
		ModelName:         channel.ModelName,
		OriginalModelName: request.Model,
		PlatformID:        channel.PlatformID,
		APIKeyID:          channel.APIKeyID,
		ModelID:           channel.ModelID,
	}
	log.DebugContext(ctx, "创建请求日志")

	// 执行请求
	log.DebugContext(ctx, "执行聊天完成请求")
	response, err := adapter.ChatCompletion(ctx, request, channel)

	// 计算耗时
	requestDuration := time.Since(now)
	requestLog.Duration = requestDuration

	log.DebugContext(ctx, "请求完成",
		"duration", requestDuration.String(),
		"success", err == nil,
	)

	if err != nil {
		// 记录失败统计
		errorMsg := err.Error()
		requestLog.Success = false
		requestLog.ErrorMsg = &errorMsg
		p.recordRequestLog(requestLog, nil, false)

		log.ErrorContext(ctx, "聊天完成请求失败", "error", err)
		return nil, err
	}

	// 记录 Token 用量
	if response.Usage != nil {
		requestLog.PromptTokens = response.Usage.InputTokens
		requestLog.CompletionTokens = response.Usage.OutputTokens
		requestLog.TotalTokens = response.Usage.TotalTokens

		log.DebugContext(ctx, "记录 Token 使用情况",
			"prompt_tokens", response.Usage.InputTokens,
			"completion_tokens", response.Usage.OutputTokens,
			"total_tokens", response.Usage.TotalTokens,
		)
	}

	// 记录成功统计
	requestLog.Success = true
	p.recordRequestLog(requestLog, nil, true)

	log.InfoContext(ctx, "聊天完成请求成功完成")
	return response, nil
}
