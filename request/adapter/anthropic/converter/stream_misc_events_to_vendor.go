package converter

import (
	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertContractToPing 转换 ping 事件为 Anthropic 格式。
func convertContractToPing(log logger.Logger) (*anthropicTypes.StreamEvent, error) {
	pingEvent := &anthropicTypes.PingEvent{
		Type: anthropicTypes.StreamEventPing,
	}

	log.Debug("转换 ping 事件完成")
	return &anthropicTypes.StreamEvent{Ping: pingEvent}, nil
}

// convertContractToError 转换 error 事件为 Anthropic 格式。
//
// 反向映射规则：
//   - error.message/type/code/param: 从 Error Payload 提取
func convertContractToError(event *types.StreamEventContract, log logger.Logger) (*anthropicTypes.StreamEvent, error) {
	if event.Error == nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "error payload 不能为空")
	}

	errorEvent := &anthropicTypes.ErrorEvent{
		Type: anthropicTypes.StreamEventError,
		Error: anthropicTypes.ErrorResponse{
			Type: "error",
			Error: anthropicTypes.Error{
				Type:    event.Error.Type,
				Message: event.Error.Message,
			},
		},
	}

	log.Error("转换 error 事件完成", "error_type", event.Error.Type, "error_message", event.Error.Message)
	return &anthropicTypes.StreamEvent{Error: errorEvent}, nil
}
