package processor

import (
	"context"
	"log/slog"
	"time"

	"github.com/MeowSalty/portal/types"
)

// recordStats 记录请求统计信息
func (p *RequestProcessor) recordStats(
	options *types.RequestStat,
	requestStart time.Time,
	firstByteTime *time.Time,
	success bool,
	errMsg *string,
	selectedChannel *types.Channel,
) {
	// 计算耗时
	requestDuration := time.Since(requestStart)
	options.Duration = requestDuration

	// 如果记录了首字节时间，则计算首字节耗时
	if firstByteTime != nil && !firstByteTime.IsZero() {
		firstByteDuration := firstByteTime.Sub(requestStart)
		options.FirstByteTime = &firstByteDuration
	}

	if errMsg != nil {
		p.Logger.Error("流式传输过程中发生错误",
			slog.String("model", options.ModelName),
			slog.String("platform", selectedChannel.Platform.Name),
			slog.String("error", *errMsg))

		// 更新健康状态为失败
		p.updateHealthStatus(selectedChannel, false, *errMsg)

		options.Success = false
		options.ErrorMsg = errMsg
	} else if success {
		// 流正常结束，更新健康状态为成功
		p.updateHealthStatus(selectedChannel, true, "")

		// 记录成功的统计信息
		options.Success = true
	}

	p.recordStatsSafe(context.Background(), options)
}

// recordStatsSafe 安全地记录统计信息
func (p *RequestProcessor) recordStatsSafe(ctx context.Context, stats *types.RequestStat) {
	if err := p.StatsManager.RecordRequestStat(ctx, stats); err != nil {
		p.Logger.Error("记录统计信息失败", slog.Any("error", err))
	}
}

// updateHealthStatus 更新健康状态
func (p *RequestProcessor) updateHealthStatus(channel *types.Channel, success bool, errorMsg string) {
	p.HealthManager.UpdateStatus(
		types.ResourceTypeAPIKey,
		channel.APIKey.ID,
		success,
		errorMsg,
		0,
	)
}
