package converter

import (
	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// StreamEventFromContract 将统一的 StreamEventContract 转换为 Anthropic 流式事件。
//
// 根据 plans/stream-event-contract-plan.md 第 5.1 章节的反向映射实现：
//   - message_start: 从 message.payload + extensions.anthropic.message 构建 MessageStartEvent
//   - message_delta: 从 delta.raw + usage 构建 MessageDeltaEvent
//   - message_stop: 构建 MessageStopEvent
//   - content_block_start: 从 content.kind + raw 构建 ContentBlockStartEvent
//   - content_block_delta: 从 delta.payload 构建 ContentBlockDeltaEvent
//   - content_block_stop: 构建 ContentBlockStopEvent
//   - ping: 构建 PingEvent
//   - error: 从 error.payload 构建 ErrorEvent
//
// Raw 优先回写策略：
//   - 若 extensions.anthropic.message 存在，优先反序列化为原始 Message
//   - 若 content.raw.content_block 存在，优先反序列化为原始 ContentBlock
//
// 参数：
//   - event: 统一流式事件合约
//   - log: 日志记录器（可选，传 nil 时使用 NopLogger）
//
// 返回：
//   - *anthropicTypes.StreamEvent: 转换后的 Anthropic 流式事件
//   - error: 转换过程中的错误
func StreamEventFromContract(event *types.StreamEventContract, log logger.Logger) (*anthropicTypes.StreamEvent, error) {
	if event == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	// 使用 WithGroup 创建子日志记录器
	log = log.WithGroup("contract_converter")

	// 根据事件类型分发转换
	switch event.Type {
	case types.StreamEventMessageStart:
		return convertContractToMessageStart(event, log)
	case types.StreamEventMessageDelta:
		return convertContractToMessageDelta(event, log)
	case types.StreamEventMessageStop:
		return convertContractToMessageStop(log)
	case types.StreamEventContentBlockStart:
		return convertContractToContentBlockStart(event, log)
	case types.StreamEventContentBlockDelta:
		return convertContractToContentBlockDelta(event, log)
	case types.StreamEventContentBlockStop:
		return convertContractToContentBlockStop(event, log)
	case types.StreamEventPing:
		return convertContractToPing(log)
	case types.StreamEventError:
		return convertContractToError(event, log)
	default:
		return nil, errors.New(errors.ErrCodeInvalidArgument, "不支持的流式事件类型："+string(event.Type))
	}
}
