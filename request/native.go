package request

import (
	"context"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/routing"
)

// Native 执行原生 API 请求（非流式）
//
// 该方法通过 Adapter.NativeRequest 统一入口调用，复用请求日志记录功能。
// 请求体和响应体均为原生类型。
//
// 参数：
//   - ctx: 上下文
//   - payload: 原生请求对象
//   - channel: 通道信息
//
// 返回：
//   - any: 原生响应对象
//   - error: 请求失败时返回错误
func (p *Request) Native(
	ctx context.Context,
	payload any,
	channel *routing.Channel,
) (any, error) {
	now := time.Now()

	// 创建带有请求上下文的日志记录器
	log := p.logger.With(
		"platform_type", channel.Provider,
		"platform_id", channel.PlatformID,
		"model_id", channel.ModelID,
		"api_key_id", channel.APIKeyID,
		"model_name", channel.ModelName,
	)

	log.DebugContext(ctx, "开始处理原生请求")

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
		Timestamp:   now,
		RequestType: "non-stream-native",
		ModelName:   channel.ModelName,
		PlatformID:  channel.PlatformID,
		APIKeyID:    channel.APIKeyID,
		ModelID:     channel.ModelID,
	}
	log.DebugContext(ctx, "创建请求日志")

	// 执行原生请求
	log.DebugContext(ctx, "执行原生请求")
	response, err := adapter.Native(ctx, channel, nil, payload)

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

		log.ErrorContext(ctx, "原生请求失败", "error", err)
		return nil, err
	}

	// 记录成功统计
	requestLog.Success = true
	p.recordRequestLog(requestLog, nil, true)

	log.InfoContext(ctx, "原生请求成功")

	return response, nil
}

// NativeStream 执行原生 API 流式请求
//
// 该方法通过 Adapter.NativeStream 统一入口调用，复用请求日志记录功能。
// 请求体为原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - payload: 原生请求对象
//   - channel: 通道信息
//   - output: 原生流事件输出通道
//
// 返回：
//   - error: 请求失败时返回错误
func (p *Request) NativeStream(
	ctx context.Context,
	payload any,
	channel *routing.Channel,
	output chan<- any,
) error {
	now := time.Now()

	// 创建带有请求上下文的日志记录器
	log := p.logger.With(
		"platform_type", channel.Provider,
		"platform_id", channel.PlatformID,
		"model_id", channel.ModelID,
		"api_key_id", channel.APIKeyID,
		"model_name", channel.ModelName,
	)

	log.DebugContext(ctx, "开始处理原生流式请求")

	// 获取适配器
	log.DebugContext(ctx, "获取适配器", "format", channel.Provider)
	adapter, err := p.getAdapter(channel.Provider)
	if err != nil {
		log.ErrorContext(ctx, "获取适配器失败", "error", err, "format", channel.Provider)
		return errors.Wrap(errors.ErrCodeAdapterNotFound, "获取适配器失败", err).
			WithContext("format", channel.Provider)
	}
	log.DebugContext(ctx, "获取适配器成功", "adapter", adapter.Name())

	// 创建请求日志
	requestLog := &RequestLog{
		Timestamp:   now,
		RequestType: "stream-native",
		ModelName:   channel.ModelName,
		PlatformID:  channel.PlatformID,
		APIKeyID:    channel.APIKeyID,
		ModelID:     channel.ModelID,
	}
	log.DebugContext(ctx, "创建请求日志")

	// 创建 RequestLogHooks 实例用于记录流式响应统计
	hooks := &RequestLogHooks{log: requestLog}

	// 执行原生流式请求
	log.DebugContext(ctx, "执行原生流式请求")
	err = adapter.NativeStream(ctx, channel, nil, payload, output, hooks)

	// 计算耗时
	requestDuration := time.Since(now)
	requestLog.Duration = requestDuration

	log.DebugContext(ctx, "流式请求完成",
		"duration", requestDuration.String(),
		"success", err == nil,
	)

	if err != nil {
		// 记录失败统计
		errorMsg := err.Error()
		requestLog.Success = false
		requestLog.ErrorMsg = &errorMsg
		p.recordRequestLog(requestLog, nil, false)

		log.ErrorContext(ctx, "原生流式请求失败", "error", err)
		return err
	}

	// 记录成功统计
	requestLog.Success = true
	p.recordRequestLog(requestLog, nil, true)

	log.InfoContext(ctx, "原生流式请求成功")

	return nil
}
