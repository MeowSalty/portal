package converter

import (
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// mapFinishReasonToStopReason 映射统一的 FinishReason 到 Anthropic StopReason。
func mapFinishReasonToStopReason(finishReason types.ResponseFinishReason) anthropicTypes.StopReason {
	switch finishReason {
	case types.ResponseFinishReasonStop:
		return anthropicTypes.StopReasonEndTurn
	case types.ResponseFinishReasonLength:
		return anthropicTypes.StopReasonMaxTokens
	case types.ResponseFinishReasonToolCalls:
		return anthropicTypes.StopReasonToolUse
	case types.ResponseFinishReasonContentFilter:
		return anthropicTypes.StopReasonRefusal
	default:
		return anthropicTypes.StopReasonEndTurn
	}
}

// mapStopReasonToFinishReason 映射 Anthropic StopReason 到统一的 FinishReason。
func mapStopReasonToFinishReason(stopReason anthropicTypes.StopReason) types.ResponseFinishReason {
	switch stopReason {
	case anthropicTypes.StopReasonEndTurn, anthropicTypes.StopReasonStopSeq:
		return types.ResponseFinishReasonStop
	case anthropicTypes.StopReasonMaxTokens:
		return types.ResponseFinishReasonLength
	case anthropicTypes.StopReasonToolUse:
		return types.ResponseFinishReasonToolCalls
	case anthropicTypes.StopReasonRefusal:
		return types.ResponseFinishReasonContentFilter
	case anthropicTypes.StopReasonPauseTurn:
		return types.ResponseFinishReasonUnknown
	default:
		return types.ResponseFinishReasonUnknown
	}
}
