package responses

import (
	"encoding/json"

	"github.com/MeowSalty/portal/request/adapter/openai/converter/responses/helper"
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// RequestFromContract 将 RequestContract 转换为 OpenAI Responses 请求。
func RequestFromContract(contract *types.RequestContract) (*responsesTypes.Request, error) {
	if contract == nil {
		return nil, nil
	}

	req := &responsesTypes.Request{}

	// 转换 Model
	if contract.Model != "" {
		model := contract.Model
		req.Model = &model
	}

	// 转换 Input
	// Prompt 优先，否则使用 Messages
	if contract.Prompt != nil {
		req.Input = &responsesTypes.InputUnion{
			StringValue: contract.Prompt,
		}
	} else if len(contract.Messages) > 0 {
		items, err := helper.ConvertMessagesToInputItems(contract.Messages)
		if err != nil {
			return nil, err
		}
		req.Input = &responsesTypes.InputUnion{
			Items: items,
		}
	}

	// 转换 System -> Instructions
	if contract.System != nil && contract.System.Text != nil {
		req.Instructions = contract.System.Text
	}

	// 转换采样参数
	req.MaxOutputTokens = contract.MaxOutputTokens
	req.Temperature = contract.Temperature
	req.TopP = contract.TopP
	req.TopLogprobs = contract.TopLogprobs

	// 转换流式配置
	req.Stream = contract.Stream
	if contract.StreamOptions != nil {
		req.StreamOptions = &responsesTypes.StreamOptions{
			IncludeObfuscation: contract.StreamOptions.IncludeObfuscation,
		}
	}

	// 转换 Metadata
	if len(contract.Metadata) > 0 {
		req.Metadata = make(map[string]string)
		for k, v := range contract.Metadata {
			if strVal, ok := v.(string); ok {
				req.Metadata[k] = strVal
			}
		}
	}

	// 转换其他顶层字段
	req.User = contract.User
	if contract.ServiceTier != nil {
		serviceTier := shared.ServiceTier(*contract.ServiceTier)
		req.ServiceTier = &serviceTier
	}
	req.PromptCacheKey = contract.PromptCacheKey
	if contract.PromptCacheRetention != nil {
		retention := responsesTypes.PromptCacheRetention(*contract.PromptCacheRetention)
		req.PromptCacheRetention = &retention
	}
	req.Store = contract.Store

	// 转换 Reasoning
	if contract.Reasoning != nil {
		req.Reasoning = &responsesTypes.Reasoning{}
		if contract.Reasoning.Effort != nil {
			effort := shared.ReasoningEffort(*contract.Reasoning.Effort)
			req.Reasoning.Effort = &effort
		}
		if contract.Reasoning.Summary != nil {
			summary := responsesTypes.ReasoningSummary(*contract.Reasoning.Summary)
			req.Reasoning.Summary = &summary
		}
		if contract.Reasoning.GenerateSummary != nil {
			generateSummary := responsesTypes.ReasoningSummary(*contract.Reasoning.GenerateSummary)
			req.Reasoning.GenerateSummary = &generateSummary
		}
	}

	// 转换 ResponseFormat -> Text.Format
	if contract.ResponseFormat != nil {
		textFormat, verbosity, err := helper.ConvertResponseFormatToTextFormat(contract.ResponseFormat)
		if err != nil {
			return nil, err
		}
		req.Text = &responsesTypes.TextConfig{
			Format: textFormat,
		}
		if verbosity != nil {
			req.Text.Verbosity = verbosity
		}
	}

	// 转换 Tools
	if len(contract.Tools) > 0 {
		tools, err := helper.ConvertToolsFromContract(contract.Tools)
		if err != nil {
			return nil, err
		}
		req.Tools = tools
	}

	// 转换 ToolChoice
	if contract.ToolChoice != nil {
		toolChoice, err := helper.ConvertToolChoiceFromContract(contract.ToolChoice)
		if err != nil {
			return nil, err
		}
		req.ToolChoice = toolChoice
	}

	// 转换 ParallelToolCalls
	req.ParallelToolCalls = contract.ParallelToolCalls

	// 从 VendorExtras 恢复特有字段
	if contract.VendorExtras != nil {
		if include, ok := contract.VendorExtras["include"].([]string); ok {
			req.Include = include
		}
		if truncation, ok := contract.VendorExtras["truncation"].(string); ok {
			t := responsesTypes.TruncationStrategy(truncation)
			req.Truncation = &t
		}
		if conversation, ok := contract.VendorExtras["conversation"].(*responsesTypes.ConversationUnion); ok {
			req.Conversation = conversation
		}
		if prompt, ok := contract.VendorExtras["prompt"].(*responsesTypes.PromptTemplate); ok {
			req.Prompt = prompt
		}
		if previousResponseID, ok := contract.VendorExtras["previous_response_id"].(*string); ok {
			req.PreviousResponseID = previousResponseID
		}
		if background, ok := contract.VendorExtras["background"].(*bool); ok {
			req.Background = background
		}
		if maxToolCalls, ok := contract.VendorExtras["max_tool_calls"].(*int); ok {
			req.MaxToolCalls = maxToolCalls
		}
		if safetyID, ok := contract.VendorExtras["safety_identifier"].(*string); ok {
			req.SafetyIdentifier = safetyID
		}
		if verbosity, ok := contract.VendorExtras["verbosity"].(string); ok {
			v := shared.VerbosityLevel(verbosity)
			if req.Text == nil {
				req.Text = &responsesTypes.TextConfig{}
			}
			req.Text.Verbosity = &v
		}

		// 恢复 ExtraFields
		var unknownFields map[string]json.RawMessage
		knownVendorFields := map[string]bool{
			"include":              true,
			"truncation":           true,
			"conversation":         true,
			"prompt":               true,
			"previous_response_id": true,
			"background":           true,
			"max_tool_calls":       true,
			"safety_identifier":    true,
			"verbosity":            true,
		}
		for k, v := range contract.VendorExtras {
			if !knownVendorFields[k] {
				// 将 interface{} 转换为 json.RawMessage
				raw, err := json.Marshal(v)
				if err != nil {
					continue // 跳过无法序列化的字段
				}
				if unknownFields == nil {
					unknownFields = make(map[string]json.RawMessage)
				}
				unknownFields[k] = raw
			}
		}
		// 只有存在未知字段时才设置 ExtraFields
		if len(unknownFields) > 0 {
			req.ExtraFields = unknownFields
		}
	}

	return req, nil
}
