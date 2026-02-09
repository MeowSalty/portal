package converter

import (
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertUsageToContract 转换 Usage 信息。
func convertUsageToContract(usage *anthropicTypes.Usage) (*types.ResponseUsage, error) {
	if usage == nil {
		return nil, nil
	}

	result := &types.ResponseUsage{
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		Extras:       make(map[string]interface{}),
	}

	// Anthropic 没有 TotalTokens，可以计算
	if usage.InputTokens != nil && usage.OutputTokens != nil {
		total := *usage.InputTokens + *usage.OutputTokens
		result.TotalTokens = &total
	}

	// 将 Anthropic 特有字段放入 Extras
	if usage.CacheCreationInputTokens != nil {
		result.Extras["anthropic.cache_creation_input_tokens"] = *usage.CacheCreationInputTokens
	}
	if usage.CacheReadInputTokens != nil {
		result.Extras["anthropic.cache_read_input_tokens"] = *usage.CacheReadInputTokens
	}
	if usage.CacheCreation != nil {
		result.Extras["anthropic.cache_creation"] = usage.CacheCreation
	}
	if usage.ServerToolUse != nil {
		result.Extras["anthropic.server_tool_use"] = usage.ServerToolUse
	}
	if usage.ServiceTier != nil {
		result.Extras["anthropic.service_tier"] = *usage.ServiceTier
	}

	return result, nil
}

// convertUsageFromContract 从 Contract 转换 Usage。
func convertUsageFromContract(usage *types.ResponseUsage) (*anthropicTypes.Usage, error) {
	if usage == nil {
		return nil, nil
	}

	result := &anthropicTypes.Usage{
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
	}

	// 从 Extras 恢复 Anthropic 特有字段
	if val, ok := usage.Extras["anthropic.cache_creation_input_tokens"].(int); ok {
		result.CacheCreationInputTokens = &val
	} else if val, ok := usage.Extras["anthropic.cache_creation_input_tokens"].(float64); ok {
		intVal := int(val)
		result.CacheCreationInputTokens = &intVal
	}

	if val, ok := usage.Extras["anthropic.cache_read_input_tokens"].(int); ok {
		result.CacheReadInputTokens = &val
	} else if val, ok := usage.Extras["anthropic.cache_read_input_tokens"].(float64); ok {
		intVal := int(val)
		result.CacheReadInputTokens = &intVal
	}

	if val, ok := usage.Extras["anthropic.cache_creation"]; ok {
		if cacheCreation, ok := val.(*anthropicTypes.CacheCreation); ok {
			result.CacheCreation = cacheCreation
		}
	}

	if val, ok := usage.Extras["anthropic.server_tool_use"]; ok {
		if serverToolUse, ok := val.(*anthropicTypes.ServerToolUsage); ok {
			result.ServerToolUse = serverToolUse
		}
	}

	if val, ok := usage.Extras["anthropic.service_tier"].(string); ok {
		serviceTier := anthropicTypes.ResponseServiceTier(val)
		result.ServiceTier = &serviceTier
	}

	return result, nil
}
