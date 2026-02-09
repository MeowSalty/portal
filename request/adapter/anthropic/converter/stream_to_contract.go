package converter

import (
	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// StreamEventToContract 将 Anthropic 流式事件转换为统一的 StreamEventContract。
//
// 根据 plans/stream-event-contract-plan.md 5.1 章节实现：
//   - message_start: 映射 message.id 到 response_id，message.role 和 content 到 message payload
//   - message_delta: 映射 stop_reason/stop_sequence 到 delta.raw，usage 到 usage payload
//   - message_stop: 简单事件，仅设置 type
//   - content_block_start: 映射 index 到 content_index，content_block 到 content payload
//   - content_block_delta: 映射 index 到 content_index，delta 到 delta payload
//   - content_block_stop: 映射 index 到 content_index
//   - ping: 简单事件，仅设置 type
//   - error: 映射 error 到 error payload
//
// 根据 plans/stream-index-repair-plan.md B.2 章节实现索引补齐：
//   - sequence_number：若为 0 -> ctx.NextSequence()
//   - output_index：无显式来源时默认 0（或 ctx.EnsureOutputIndex(message_id)）
//   - item_id：若有 tool_use.id -> 用 tool_use.id；否则 message_id（message_start/stop）；content_block 事件用 message_id:content_index 组合
//   - content_index：来自 event.Index（content_block*）；若缺失 -> ctx.EnsureContentIndex(item_id, -1)
//
// 参数：
//   - event: Anthropic 流式事件
//   - ctx: 流索引上下文，用于生成和维护稳定的索引值
//   - log: 日志记录器（可选，传 nil 时使用 NopLogger）
//
// 返回：
//   - *types.StreamEventContract: 转换后的统一流式事件
//   - error: 转换过程中的错误
func StreamEventToContract(event *anthropicTypes.StreamEvent, ctx types.StreamIndexContext, log logger.Logger) (*types.StreamEventContract, error) {
	if event == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	// 使用 WithGroup 创建子日志记录器
	log = log.WithGroup("stream_converter")

	// 根据事件类型分发转换
	if event.MessageStart != nil {
		return convertMessageStartEvent(event.MessageStart, ctx, log)
	} else if event.MessageDelta != nil {
		return convertMessageDeltaEvent(event.MessageDelta, ctx, log)
	} else if event.MessageStop != nil {
		return convertMessageStopEvent(ctx, log)
	} else if event.ContentBlockStart != nil {
		return convertContentBlockStartEvent(event.ContentBlockStart, ctx, log)
	} else if event.ContentBlockDelta != nil {
		return convertContentBlockDeltaEvent(event.ContentBlockDelta, ctx, log)
	} else if event.ContentBlockStop != nil {
		return convertContentBlockStopEvent(event.ContentBlockStop, ctx, log)
	} else if event.Ping != nil {
		return convertPingEvent(log)
	} else if event.Error != nil {
		return convertErrorEvent(event.Error, log)
	}

	return nil, errors.New(errors.ErrCodeInvalidArgument, "未知的流式事件类型")
}
