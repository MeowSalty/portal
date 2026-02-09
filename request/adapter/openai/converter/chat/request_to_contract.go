package chat

import (
	"encoding/json"

	chatTypes "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// RequestToContract 将 OpenAI Chat 请求转换为统一的 RequestContract。
func RequestToContract(req *chatTypes.Request) (*types.RequestContract, error) {
	if req == nil {
		return nil, nil
	}

	contract := &types.RequestContract{
		Source: types.VendorSourceOpenAIChat,
		Model:  req.Model,
	}

	// 转换 Messages
	// OpenAI Chat 中 system 消息在 messages 数组中，需要提取到 System 字段
	if len(req.Messages) > 0 {
		messages, system, err := convertMessagesToContract(req.Messages)
		if err != nil {
			return nil, err
		}
		contract.Messages = messages
		contract.System = system
	}

	// 转换采样参数
	// MaxCompletionTokens 优先于 MaxTokens
	if req.MaxCompletionTokens != nil {
		contract.MaxOutputTokens = req.MaxCompletionTokens
	} else if req.MaxTokens != nil {
		contract.MaxOutputTokens = req.MaxTokens
	}

	contract.Temperature = req.Temperature
	contract.TopP = req.TopP
	contract.PresencePenalty = req.PresencePenalty
	contract.FrequencyPenalty = req.FrequencyPenalty
	contract.Seed = req.Seed

	// 转换 Stop
	if req.Stop != nil {
		contract.Stop = convertStopToContract(req.Stop)
	}

	// 转换 N (候选数)
	contract.CandidateCount = req.N

	// 转换 Logprobs
	contract.Logprobs = req.Logprobs
	contract.TopLogprobs = req.TopLogprobs

	// 转换流式配置
	contract.Stream = req.Stream
	if req.StreamOptions != nil {
		contract.StreamOptions = &types.StreamOption{
			IncludeUsage:       req.StreamOptions.IncludeUsage,
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
	contract.PromptCacheRetention = req.PromptCacheRetention
	contract.Store = req.Store

	// 转换 ReasoningEffort
	if req.ReasoningEffort != nil {
		effort := string(*req.ReasoningEffort)
		contract.Reasoning = &types.Reasoning{
			Effort: &effort,
		}
	}

	// 转换 ResponseFormat
	if req.ResponseFormat != nil {
		responseFormat, err := convertResponseFormatToContract(req.ResponseFormat)
		if err != nil {
			return nil, err
		}
		contract.ResponseFormat = responseFormat
	}

	// 转换 Modalities
	if len(req.Modalities) > 0 {
		contract.Modalities = make([]string, len(req.Modalities))
		for i, m := range req.Modalities {
			contract.Modalities[i] = string(m)
		}
	}

	// 转换 Tools
	if len(req.Tools) > 0 {
		tools, err := convertToolsToContract(req.Tools)
		if err != nil {
			return nil, err
		}
		contract.Tools = tools
	}

	// 转换 ToolChoice
	if req.ToolChoice != nil {
		toolChoice, err := convertToolChoiceToContract(req.ToolChoice)
		if err != nil {
			return nil, err
		}
		contract.ToolChoice = toolChoice
	}

	// 转换 ParallelToolCalls
	contract.ParallelToolCalls = req.ParallelToolCalls

	// 初始化 VendorExtras 并存储特有字段
	contract.VendorExtras = make(map[string]interface{})
	source := types.VendorSourceOpenAIChat
	contract.VendorExtrasSource = &source

	// OpenAI Chat 特有字段放入 VendorExtras
	if req.Audio != nil {
		contract.VendorExtras["audio"] = req.Audio
	}
	if req.Prediction != nil {
		contract.VendorExtras["prediction"] = req.Prediction
	}
	if req.WebSearchOptions != nil {
		contract.VendorExtras["web_search_options"] = req.WebSearchOptions
	}
	if req.Verbosity != nil {
		contract.VendorExtras["verbosity"] = string(*req.Verbosity)
	}
	if req.SafetyIdentifier != nil {
		contract.VendorExtras["safety_identifier"] = req.SafetyIdentifier
	}
	if len(req.LogitBias) > 0 {
		contract.VendorExtras["logit_bias"] = req.LogitBias
	}

	// 转换 FunctionCall (deprecated)
	if req.FunctionCall != nil {
		contract.VendorExtras["function_call"] = req.FunctionCall
	}

	// 转换 Functions (deprecated)
	if len(req.Functions) > 0 {
		contract.VendorExtras["functions"] = req.Functions
	}

	// 合并 ExtraFields
	if len(req.ExtraFields) > 0 {
		for k, v := range req.ExtraFields {
			contract.VendorExtras[k] = v
		}
	}

	return contract, nil
}

// convertMessagesToContract 转换消息列表，提取 system 消息。
func convertMessagesToContract(messages []chatTypes.RequestMessage) ([]types.Message, *types.System, error) {
	result := make([]types.Message, 0, len(messages))
	var system *types.System

	for _, msg := range messages {
		// 提取 system 消息
		if msg.Role == chatTypes.MessageRoleSystem {
			if system == nil {
				system = &types.System{}
			}

			// 转换 system 消息内容
			if msg.Content.StringValue != nil {
				system.Text = msg.Content.StringValue
			} else if len(msg.Content.ContentParts) > 0 {
				parts, err := convertContentPartsToContract(msg.Content.ContentParts)
				if err != nil {
					return nil, nil, err
				}
				system.Parts = parts
			}

			// system 消息的 ExtraFields 放入 VendorExtras
			if len(msg.ExtraFields) > 0 {
				if system.VendorExtras == nil {
					system.VendorExtras = make(map[string]interface{})
				}
				source := types.VendorSourceOpenAIChat
				system.VendorExtrasSource = &source
				for k, v := range msg.ExtraFields {
					system.VendorExtras[k] = v
				}
			}

			continue
		}

		// 转换普通消息
		contractMsg := types.Message{
			Role:       msg.Role,
			Name:       msg.Name,
			ToolCallID: msg.ToolCallID,
		}

		// 转换 Content
		if msg.Content.StringValue != nil {
			contractMsg.Content = types.Content{
				Text: msg.Content.StringValue,
			}
		} else if len(msg.Content.ContentParts) > 0 {
			parts, err := convertContentPartsToContract(msg.Content.ContentParts)
			if err != nil {
				return nil, nil, err
			}
			contractMsg.Content = types.Content{
				Parts: parts,
			}
		}

		// 转换 ToolCalls
		if len(msg.ToolCalls) > 0 {
			toolCalls, err := convertToolCallsToContract(msg.ToolCalls)
			if err != nil {
				return nil, nil, err
			}
			contractMsg.ToolCalls = toolCalls
		}

		// 消息级别的特有字段放入 VendorExtras
		if msg.FunctionCall != nil || msg.Refusal != nil || msg.Audio != nil || len(msg.ExtraFields) > 0 {
			contractMsg.VendorExtras = make(map[string]interface{})
			source := types.VendorSourceOpenAIChat
			contractMsg.VendorExtrasSource = &source

			if msg.FunctionCall != nil {
				contractMsg.VendorExtras["function_call"] = msg.FunctionCall
			}
			if msg.Refusal != nil {
				contractMsg.VendorExtras["refusal"] = msg.Refusal
			}
			if msg.Audio != nil {
				contractMsg.VendorExtras["audio"] = msg.Audio
			}
			for k, v := range msg.ExtraFields {
				contractMsg.VendorExtras[k] = v
			}
		}

		result = append(result, contractMsg)
	}

	return result, system, nil
}

// convertContentPartsToContract 转换内容片段列表。
func convertContentPartsToContract(parts []chatTypes.ContentPart) ([]types.ContentPart, error) {
	result := make([]types.ContentPart, 0, len(parts))

	for _, part := range parts {
		contractPart := types.ContentPart{
			Type: part.Type,
		}

		switch part.Type {
		case chatTypes.ContentPartTypeText:
			contractPart.Text = part.Text

		case chatTypes.ContentPartTypeImageURL:
			if part.ImageURL != nil {
				contractPart.Image = &types.Image{
					URL: &part.ImageURL.URL,
				}
				if part.ImageURL.Detail != nil {
					detail := string(*part.ImageURL.Detail)
					contractPart.Image.Detail = &detail
				}
			}

		case chatTypes.ContentPartTypeInputAudio:
			if part.InputAudio != nil {
				format := string(part.InputAudio.Format)
				contractPart.Audio = &types.Audio{
					Data:   &part.InputAudio.Data,
					Format: &format,
				}
			}

		case chatTypes.ContentPartTypeFile:
			if part.File != nil {
				contractPart.File = &types.File{
					ID:       part.File.FileID,
					Data:     part.File.FileData,
					Filename: part.File.Filename,
				}
			}

		case chatTypes.ContentPartTypeRefusal:
			// Refusal 类型放入 VendorExtras
			contractPart.VendorExtras = make(map[string]interface{})
			source := types.VendorSourceOpenAIChat
			contractPart.VendorExtrasSource = &source
			contractPart.VendorExtras["refusal"] = part.Refusal

		default:
			// 未知类型放入 VendorExtras
			contractPart.VendorExtras = make(map[string]interface{})
			source := types.VendorSourceOpenAIChat
			contractPart.VendorExtrasSource = &source
			contractPart.VendorExtras["original_part"] = part
		}

		result = append(result, contractPart)
	}

	return result, nil
}

// convertToolCallsToContract 转换工具调用列表。
func convertToolCallsToContract(toolCalls []chatTypes.RequestToolCall) ([]types.ToolCall, error) {
	result := make([]types.ToolCall, 0, len(toolCalls))

	for _, tc := range toolCalls {
		toolType := string(tc.Type)
		contractToolCall := types.ToolCall{
			ID:        &tc.ID,
			Type:      &toolType,
			Name:      &tc.Function.Name,
			Arguments: &tc.Function.Arguments,
		}

		result = append(result, contractToolCall)
	}

	return result, nil
}

// convertStopToContract 转换停止条件。
func convertStopToContract(stop *chatTypes.StopConfiguration) *types.Stop {
	result := &types.Stop{}

	if stop.StringValue != nil {
		result.Text = stop.StringValue
	} else if len(stop.StringArray) > 0 {
		result.List = stop.StringArray
	}

	return result
}

// convertResponseFormatToContract 转换响应格式。
func convertResponseFormatToContract(format *chatTypes.FormatUnion) (*types.ResponseFormat, error) {
	if format == nil {
		return nil, nil
	}

	result := &types.ResponseFormat{}

	if format.Text != nil {
		result.Type = string(format.Text.Type)
	} else if format.JSONSchema != nil {
		result.Type = string(format.JSONSchema.Type)
		// 构造 JSONSchema
		schema := map[string]interface{}{
			"name":   format.JSONSchema.JSONSchema.Name,
			"schema": format.JSONSchema.JSONSchema.Schema,
		}
		if format.JSONSchema.JSONSchema.Description != nil {
			schema["description"] = *format.JSONSchema.JSONSchema.Description
		}
		if format.JSONSchema.JSONSchema.Strict != nil {
			schema["strict"] = *format.JSONSchema.JSONSchema.Strict
		}
		result.JSONSchema = schema
	} else if format.JSONObject != nil {
		result.Type = string(format.JSONObject.Type)
	}

	return result, nil
}

// convertToolsToContract 转换工具列表。
// Chat 专用类型，仅支持 function 和 custom 工具类型。
func convertToolsToContract(tools []chatTypes.ChatToolUnion) ([]types.Tool, error) {
	result := make([]types.Tool, 0, len(tools))

	for _, tool := range tools {
		if tool.Function != nil {
			// 标准函数工具
			contractTool := types.Tool{
				Type: "function",
				Function: &types.Function{
					Name:        tool.Function.Function.Name,
					Description: tool.Function.Function.Description,
					Parameters:  tool.Function.Function.Parameters,
				},
			}

			// Strict 放入 VendorExtras
			if tool.Function.Function.Strict != nil {
				contractTool.VendorExtras = make(map[string]interface{})
				source := types.VendorSourceOpenAIChat
				contractTool.VendorExtrasSource = &source
				contractTool.VendorExtras["strict"] = *tool.Function.Function.Strict
			}

			result = append(result, contractTool)
		} else if tool.Custom != nil {
			// 自定义工具
			contractTool := types.Tool{
				Type: "custom",
				VendorExtras: map[string]interface{}{
					"tool": tool.Custom,
				},
			}
			source := types.VendorSourceOpenAIChat
			contractTool.VendorExtrasSource = &source
			result = append(result, contractTool)
		}
		// ChatToolUnion 仅支持 function 和 custom，其他类型不会存在
	}

	return result, nil
}

// convertToolChoiceToContract 转换工具选择。
func convertToolChoiceToContract(toolChoice *shared.ToolChoiceUnion) (*types.ToolChoice, error) {
	if toolChoice == nil {
		return nil, nil
	}

	result := &types.ToolChoice{}

	if toolChoice.Auto != nil {
		// 字符串模式（auto/none/required）
		mode := *toolChoice.Auto
		result.Mode = &mode
	} else if toolChoice.Named != nil {
		// 命名函数选择
		mode := "function"
		result.Mode = &mode
		result.Function = &toolChoice.Named.Function.Name
	} else if toolChoice.Allowed != nil {
		// 允许的工具列表
		mode := toolChoice.Allowed.Mode
		result.Mode = &mode

		// 提取允许的工具名称
		if len(toolChoice.Allowed.Tools) > 0 {
			allowed := make([]string, 0, len(toolChoice.Allowed.Tools))
			for _, t := range toolChoice.Allowed.Tools {
				if name, ok := t["name"].(string); ok {
					allowed = append(allowed, name)
				}
			}
			if len(allowed) > 0 {
				result.Allowed = allowed
			}
		}
	} else {
		// 其他类型放入 VendorExtras（NamedCustom/NamedMCP/Hosted 等）
		// 这些类型在统一格式中没有直接对应，需要序列化保存
		toolChoiceJSON, err := json.Marshal(toolChoice)
		if err != nil {
			return nil, err
		}

		var toolChoiceMap map[string]interface{}
		if err := json.Unmarshal(toolChoiceJSON, &toolChoiceMap); err != nil {
			return nil, err
		}

		// 提取 type 作为 mode
		if typeVal, ok := toolChoiceMap["type"].(string); ok {
			result.Mode = &typeVal
		}
	}

	return result, nil
}
