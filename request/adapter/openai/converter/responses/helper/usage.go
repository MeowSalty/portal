package helper

import (
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ConvertUsageToContract 将 OpenAI Responses Usage 转换为统一 Usage。
func ConvertUsageToContract(usage *responsesTypes.Usage) *types.ResponseUsage {
	if usage == nil {
		return nil
	}

	inputTokens := usage.InputTokens
	outputTokens := usage.OutputTokens
	totalTokens := usage.TotalTokens

	result := &types.ResponseUsage{
		InputTokens:  &inputTokens,
		OutputTokens: &outputTokens,
		TotalTokens:  &totalTokens,
		Extras:       make(map[string]interface{}),
	}

	// 保存细分字段到 Extras
	result.Extras["openai.responses.input_tokens_details"] = usage.InputTokensDetails
	result.Extras["openai.responses.output_tokens_details"] = usage.OutputTokensDetails

	return result
}

// ConvertUsageFromContract 将统一 Usage 转换为 OpenAI Responses Usage。
func ConvertUsageFromContract(usage *types.ResponseUsage) *responsesTypes.Usage {
	if usage == nil {
		return nil
	}

	result := &responsesTypes.Usage{}

	if usage.InputTokens != nil {
		result.InputTokens = *usage.InputTokens
	}
	if usage.OutputTokens != nil {
		result.OutputTokens = *usage.OutputTokens
	}
	if usage.TotalTokens != nil {
		result.TotalTokens = *usage.TotalTokens
	}

	// 从 Extras 恢复细分字段
	if val, ok := usage.Extras["openai.responses.input_tokens_details"]; ok {
		if details, ok := val.(responsesTypes.InputTokensDetails); ok {
			result.InputTokensDetails = details
		}
	}

	if val, ok := usage.Extras["openai.responses.output_tokens_details"]; ok {
		if details, ok := val.(responsesTypes.OutputTokensDetails); ok {
			result.OutputTokensDetails = details
		}
	}

	return result
}

// ConvertUsageToStreamUsage 将 OpenAI Responses Usage 转换为 StreamUsagePayload。
func ConvertUsageToStreamUsage(usage *responsesTypes.Usage) *types.StreamUsagePayload {
	if usage == nil {
		return nil
	}

	result := &types.StreamUsagePayload{
		InputTokens:  &usage.InputTokens,
		OutputTokens: &usage.OutputTokens,
		TotalTokens:  &usage.TotalTokens,
		Raw:          make(map[string]interface{}),
	}

	// 保存特有字段到 raw
	result.Raw["input_tokens_details"] = usage.InputTokensDetails
	result.Raw["output_tokens_details"] = usage.OutputTokensDetails

	return result
}

// ConvertStreamUsageToResponsesUsage 将 StreamUsagePayload 转换为 OpenAI Responses Usage。
func ConvertStreamUsageToResponsesUsage(usage *types.StreamUsagePayload) *responsesTypes.Usage {
	if usage == nil {
		return nil
	}

	result := &responsesTypes.Usage{}

	if usage.InputTokens != nil {
		result.InputTokens = *usage.InputTokens
	}
	if usage.OutputTokens != nil {
		result.OutputTokens = *usage.OutputTokens
	}
	if usage.TotalTokens != nil {
		result.TotalTokens = *usage.TotalTokens
	}

	// 从 Raw 恢复细分字段
	if val, ok := usage.Raw["input_tokens_details"]; ok {
		if details, ok := val.(responsesTypes.InputTokensDetails); ok {
			result.InputTokensDetails = details
		}
	}

	if val, ok := usage.Raw["output_tokens_details"]; ok {
		if details, ok := val.(responsesTypes.OutputTokensDetails); ok {
			result.OutputTokensDetails = details
		}
	}

	return result
}