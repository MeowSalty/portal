package responses

import (
	"encoding/json"

	helper "github.com/MeowSalty/portal/request/adapter/openai/converter/responses/helper"
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// RequestToContract 将 OpenAI Responses 请求转换为统一的 RequestContract。
func RequestToContract(req *responsesTypes.Request) (*types.RequestContract, error) {
	if req == nil {
		return nil, nil
	}

	contract := &types.RequestContract{
		Source: types.VendorSourceOpenAIResponse,
	}

	// 转换 Model
	if req.Model != nil {
		contract.Model = *req.Model
	}

	// 转换 Input
	// Input 可以是 string 或 items 数组
	if req.Input != nil {
		if req.Input.StringValue != nil {
			// string -> Prompt
			contract.Prompt = req.Input.StringValue
		} else if len(req.Input.Items) > 0 {
			// items -> Messages
			messages, err := helper.ConvertInputItemsToMessages(req.Input.Items)
			if err != nil {
				return nil, err
			}
			contract.Messages = messages
		}
	}

	// 转换 Instructions -> System
	if req.Instructions != nil {
		contract.System = &types.System{
			Text: req.Instructions,
		}
	}

	// 转换采样参数
	contract.MaxOutputTokens = req.MaxOutputTokens
	contract.Temperature = req.Temperature
	contract.TopP = req.TopP
	contract.TopLogprobs = req.TopLogprobs

	// 转换流式配置
	contract.Stream = req.Stream
	if req.StreamOptions != nil {
		contract.StreamOptions = &types.StreamOption{
			IncludeObfuscation: req.StreamOptions.IncludeObfuscation,
		}
	}

	// 转换 Metadata
	if len(req.Metadata) > 0 {
		contract.Metadata = make(map[string]interface{})
		for k, v := range req.Metadata {
			contract.Metadata[k] = v
		}
	}

	// 转换其他顶层字段
	contract.User = req.User
	if req.ServiceTier != nil {
		serviceTier := string(*req.ServiceTier)
		contract.ServiceTier = &serviceTier
	}
	contract.PromptCacheKey = req.PromptCacheKey
	if req.PromptCacheRetention != nil {
		retention := string(*req.PromptCacheRetention)
		contract.PromptCacheRetention = &retention
	}
	contract.Store = req.Store

	// 转换 Reasoning
	if req.Reasoning != nil {
		contract.Reasoning = &types.Reasoning{}
		if req.Reasoning.Effort != nil {
			effort := string(*req.Reasoning.Effort)
			contract.Reasoning.Effort = &effort
		}
		if req.Reasoning.Summary != nil {
			summary := string(*req.Reasoning.Summary)
			contract.Reasoning.Summary = &summary
		}
		if req.Reasoning.GenerateSummary != nil {
			generateSummary := string(*req.Reasoning.GenerateSummary)
			contract.Reasoning.GenerateSummary = &generateSummary
		}
	}

	// 转换 Text.Format -> ResponseFormat
	if req.Text != nil && req.Text.Format != nil {
		responseFormat, err := helper.ConvertTextFormatToResponseFormat(req.Text.Format)
		if err != nil {
			return nil, err
		}
		contract.ResponseFormat = responseFormat
	}

	// 转换 Tools
	if len(req.Tools) > 0 {
		tools, err := helper.ConvertToolsToContract(req.Tools)
		if err != nil {
			return nil, err
		}
		contract.Tools = tools
	}

	// 转换 ToolChoice
	if req.ToolChoice != nil {
		toolChoice, err := helper.ConvertToolChoiceToContract(req.ToolChoice)
		if err != nil {
			return nil, err
		}
		contract.ToolChoice = toolChoice
	}

	// 转换 ParallelToolCalls
	contract.ParallelToolCalls = req.ParallelToolCalls

	// 初始化 VendorExtras 并存储特有字段
	contract.VendorExtras = make(map[string]interface{})
	source := types.VendorSourceOpenAIResponse
	contract.VendorExtrasSource = &source

	// OpenAI Responses 特有字段放入 VendorExtras
	if len(req.Include) > 0 {
		contract.VendorExtras["include"] = req.Include
	}
	if req.Truncation != nil {
		contract.VendorExtras["truncation"] = string(*req.Truncation)
	}
	if req.Conversation != nil {
		contract.VendorExtras["conversation"] = req.Conversation
	}
	if req.Prompt != nil {
		contract.VendorExtras["prompt"] = req.Prompt
	}
	if req.PreviousResponseID != nil {
		contract.VendorExtras["previous_response_id"] = req.PreviousResponseID
	}
	if req.Background != nil {
		contract.VendorExtras["background"] = req.Background
	}
	if req.MaxToolCalls != nil {
		contract.VendorExtras["max_tool_calls"] = req.MaxToolCalls
	}
	if req.SafetyIdentifier != nil {
		contract.VendorExtras["safety_identifier"] = req.SafetyIdentifier
	}
	if req.Text != nil && req.Text.Verbosity != nil {
		contract.VendorExtras["verbosity"] = string(*req.Text.Verbosity)
	}

	// 合并 ExtraFields
	if len(req.ExtraFields) > 0 {
		for k, v := range req.ExtraFields {
			// 将 json.RawMessage 转换为 interface{} 以兼容 VendorExtras
			var value interface{}
			if err := json.Unmarshal(v, &value); err != nil {
				// 如果反序列化失败，直接保存原始 JSON 字符串
				contract.VendorExtras[k] = string(v)
			} else {
				contract.VendorExtras[k] = value
			}
		}
	}

	return contract, nil
}
