package converter

import (
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertReasoningFromContract 从 Contract 转换推理配置。
func convertReasoningFromContract(reasoning *types.Reasoning) (*anthropicTypes.ThinkingConfigParam, error) {
	if reasoning == nil {
		return nil, nil
	}

	result := &anthropicTypes.ThinkingConfigParam{}

	if reasoning.Mode != nil {
		switch *reasoning.Mode {
		case "enabled":
			result.Enabled = &anthropicTypes.ThinkingConfigEnabled{
				Type: anthropicTypes.ThinkingConfigTypeEnabled,
			}
			if reasoning.Budget != nil {
				result.Enabled.BudgetTokens = *reasoning.Budget
			}
		case "disabled":
			result.Disabled = &anthropicTypes.ThinkingConfigDisabled{
				Type: anthropicTypes.ThinkingConfigTypeDisabled,
			}
		}
	}

	return result, nil
}

// convertThinkingToContract 转换思考配置。
func convertThinkingToContract(thinking *anthropicTypes.ThinkingConfigParam) (*types.Reasoning, error) {
	if thinking == nil {
		return nil, nil
	}

	result := &types.Reasoning{}

	if thinking.Enabled != nil {
		mode := "enabled"
		result.Mode = &mode
		result.Budget = &thinking.Enabled.BudgetTokens
	} else if thinking.Disabled != nil {
		mode := "disabled"
		result.Mode = &mode
	}

	return result, nil
}
