package converter

import (
	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertPingEvent 转换 ping 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.1）：
//   - 简单事件，仅设置 type
func convertPingEvent(log logger.Logger) (*types.StreamEventContract, error) {
	contract := &types.StreamEventContract{
		Type:   types.StreamEventPing,
		Source: types.StreamSourceAnthropic,
	}

	log.Debug("转换 ping 事件完成")
	return contract, nil
}

// convertErrorEvent 转换 error 事件。
//
// 映射规则（plans/stream-event-contract-plan.md 5.1）：
//   - error.message/type/code/param: 来自 ErrorResponse
//   - error.raw: 其余字段
func convertErrorEvent(event *anthropicTypes.ErrorEvent, log logger.Logger) (*types.StreamEventContract, error) {
	contract := &types.StreamEventContract{
		Type:   types.StreamEventError,
		Source: types.StreamSourceAnthropic,
	}

	// 转换 Error
	errorPayload := &types.StreamErrorPayload{
		Message: event.Error.Error.Message,
		Type:    event.Error.Error.Type,
		Raw:     make(map[string]interface{}),
	}

	contract.Error = errorPayload

	log.Error("转换 error 事件完成", "error_type", errorPayload.Type, "error_message", errorPayload.Message)
	return contract, nil
}
