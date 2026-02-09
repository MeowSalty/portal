package chat

import (
	"fmt"

	"github.com/MeowSalty/portal/errors"
	chatTypes "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// RequestFromContract 将 RequestContract 转换为 OpenAI Chat 请求。
func RequestFromContract(contract *types.RequestContract) (*chatTypes.Request, error) {
	if contract == nil {
		return nil, nil
	}

	req := &chatTypes.Request{
		Model: contract.Model,
	}

	// 转换 Messages
	// OpenAI Chat 中 system 消息需要插入到 messages 数组中
	if len(contract.Messages) > 0 || contract.System != nil {
		messages, err := convertMessagesFromContract(contract.Messages, contract.System)
		if err != nil {
			return nil, err
		}
		req.Messages = messages
	} else if contract.Prompt != nil {
		// 若 Prompt 不为空且 Messages 为空，构造单条 user message
		req.Messages = []chatTypes.RequestMessage{
			{
				Role: chatTypes.MessageRoleUser,
				Content: chatTypes.MessageContent{
					StringValue: contract.Prompt,
				},
			},
		}
	}

	// 转换采样参数
	// MaxOutputTokens -> MaxCompletionTokens (优先)
	if contract.MaxOutputTokens != nil {
		req.MaxCompletionTokens = contract.MaxOutputTokens
	}

	req.Temperature = contract.Temperature
	req.TopP = contract.TopP
	req.PresencePenalty = contract.PresencePenalty
	req.FrequencyPenalty = contract.FrequencyPenalty
	req.Seed = contract.Seed

	// 转换 Stop
	if contract.Stop != nil {
		req.Stop = convertStopFromContract(contract.Stop)
	}

	// 转换 CandidateCount -> N
	req.N = contract.CandidateCount

	// 转换 Logprobs
	req.Logprobs = contract.Logprobs
	req.TopLogprobs = contract.TopLogprobs

	// 转换流式配置
	req.Stream = contract.Stream
	if contract.StreamOptions != nil {
		req.StreamOptions = &chatTypes.StreamOptions{
			IncludeUsage:       contract.StreamOptions.IncludeUsage,
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
	req.PromptCacheRetention = contract.PromptCacheRetention
	req.Store = contract.Store

	// 转换 Reasoning -> ReasoningEffort
	if contract.Reasoning != nil && contract.Reasoning.Effort != nil {
		effort := shared.ReasoningEffort(*contract.Reasoning.Effort)
		req.ReasoningEffort = &effort
	}

	// 转换 ResponseFormat
	if contract.ResponseFormat != nil {
		responseFormat, err := convertResponseFormatFromContract(contract.ResponseFormat)
		if err != nil {
			return nil, err
		}
		req.ResponseFormat = responseFormat
	}

	// 转换 Modalities
	if len(contract.Modalities) > 0 {
		req.Modalities = make([]chatTypes.ChatModalities, len(contract.Modalities))
		for i, m := range contract.Modalities {
			req.Modalities[i] = chatTypes.ChatModalities(m)
		}
	}

	// 转换 Tools
	if len(contract.Tools) > 0 {
		tools, err := convertToolsFromContract(contract.Tools)
		if err != nil {
			return nil, err
		}
		req.Tools = tools
	}

	// 转换 ToolChoice
	if contract.ToolChoice != nil {
		toolChoice, err := convertToolChoiceFromContract(contract.ToolChoice)
		if err != nil {
			return nil, err
		}
		req.ToolChoice = toolChoice
	}

	// 转换 ParallelToolCalls
	req.ParallelToolCalls = contract.ParallelToolCalls

	// 从 VendorExtras 恢复特有字段
	if contract.VendorExtras != nil {
		if audio, ok := contract.VendorExtras["audio"].(*chatTypes.RequestAudio); ok {
			req.Audio = audio
		}
		if prediction, ok := contract.VendorExtras["prediction"].(*chatTypes.PredictionContent); ok {
			req.Prediction = prediction
		}
		if webSearch, ok := contract.VendorExtras["web_search_options"].(*chatTypes.WebSearchOptions); ok {
			req.WebSearchOptions = webSearch
		}
		if verbosity, ok := contract.VendorExtras["verbosity"].(string); ok {
			v := shared.VerbosityLevel(verbosity)
			req.Verbosity = &v
		}
		if safetyID, ok := contract.VendorExtras["safety_identifier"].(*string); ok {
			req.SafetyIdentifier = safetyID
		}
		if logitBias, ok := contract.VendorExtras["logit_bias"].(map[string]int); ok {
			req.LogitBias = logitBias
		}
		if functionCall, ok := contract.VendorExtras["function_call"].(*chatTypes.FunctionCallUnion); ok {
			req.FunctionCall = functionCall
		}
		if functions, ok := contract.VendorExtras["functions"].([]shared.FunctionDefinition); ok {
			req.Functions = functions
		}

		// 恢复 ExtraFields
		req.ExtraFields = make(map[string]interface{})
		knownVendorFields := map[string]bool{
			"audio":              true,
			"prediction":         true,
			"web_search_options": true,
			"verbosity":          true,
			"safety_identifier":  true,
			"logit_bias":         true,
			"function_call":      true,
			"functions":          true,
		}
		for k, v := range contract.VendorExtras {
			if !knownVendorFields[k] {
				req.ExtraFields[k] = v
			}
		}
	}

	return req, nil
}

// convertMessagesFromContract 从 Contract 转换消息列表，插入 system 消息。
func convertMessagesFromContract(messages []types.Message, system *types.System) ([]chatTypes.RequestMessage, error) {
	result := make([]chatTypes.RequestMessage, 0, len(messages)+1)

	// 如果有 system 指令，插入到最前面
	if system != nil {
		systemMsg := chatTypes.RequestMessage{
			Role: chatTypes.MessageRoleSystem,
		}

		if system.Text != nil {
			systemMsg.Content = chatTypes.MessageContent{
				StringValue: system.Text,
			}
		} else if len(system.Parts) > 0 {
			parts, err := convertContentPartsFromContract(system.Parts)
			if err != nil {
				return nil, err
			}
			systemMsg.Content = chatTypes.MessageContent{
				ContentParts: parts,
			}
		}

		// 从 VendorExtras 恢复 system 消息的额外字段
		if system.VendorExtras != nil {
			systemMsg.ExtraFields = make(map[string]interface{})
			for k, v := range system.VendorExtras {
				systemMsg.ExtraFields[k] = v
			}
		}

		result = append(result, systemMsg)
	}

	// 转换普通消息
	for _, msg := range messages {
		reqMsg := chatTypes.RequestMessage{
			Role:       msg.Role,
			Name:       msg.Name,
			ToolCallID: msg.ToolCallID,
		}

		// 转换 Content
		if msg.Content.Text != nil {
			reqMsg.Content = chatTypes.MessageContent{
				StringValue: msg.Content.Text,
			}
		} else if len(msg.Content.Parts) > 0 {
			parts, err := convertContentPartsFromContract(msg.Content.Parts)
			if err != nil {
				return nil, err
			}
			reqMsg.Content = chatTypes.MessageContent{
				ContentParts: parts,
			}
		}

		// 转换 ToolCalls
		if len(msg.ToolCalls) > 0 {
			toolCalls, err := convertToolCallsFromContract(msg.ToolCalls)
			if err != nil {
				return nil, err
			}
			reqMsg.ToolCalls = toolCalls
		}

		// 从 VendorExtras 恢复消息级别的特有字段
		if msg.VendorExtras != nil {
			if functionCall, ok := msg.VendorExtras["function_call"].(*chatTypes.RequestFunctionCall); ok {
				reqMsg.FunctionCall = functionCall
			}
			if refusal, ok := msg.VendorExtras["refusal"].(*string); ok {
				reqMsg.Refusal = refusal
			}
			if audio, ok := msg.VendorExtras["audio"].(*chatTypes.AssistantAudio); ok {
				reqMsg.Audio = audio
			}

			// 恢复其他 ExtraFields
			reqMsg.ExtraFields = make(map[string]interface{})
			knownFields := map[string]bool{
				"function_call": true,
				"refusal":       true,
				"audio":         true,
			}
			for k, v := range msg.VendorExtras {
				if !knownFields[k] {
					reqMsg.ExtraFields[k] = v
				}
			}
		}

		result = append(result, reqMsg)
	}

	return result, nil
}

// convertContentPartsFromContract 从 Contract 转换内容片段列表。
func convertContentPartsFromContract(parts []types.ContentPart) ([]chatTypes.ContentPart, error) {
	result := make([]chatTypes.ContentPart, 0, len(parts))

	for _, part := range parts {
		reqPart := chatTypes.ContentPart{
			Type: part.Type,
		}

		switch part.Type {
		case "text":
			reqPart.Text = part.Text

		case "image", chatTypes.ContentPartTypeImageURL:
			if part.Image != nil {
				reqPart.Type = chatTypes.ContentPartTypeImageURL
				reqPart.ImageURL = &chatTypes.ImageURL{}

				if part.Image.URL != nil {
					reqPart.ImageURL.URL = *part.Image.URL
				}
				if part.Image.Detail != nil {
					detail := shared.ImageDetail(*part.Image.Detail)
					reqPart.ImageURL.Detail = &detail
				}
			}

		case "audio", chatTypes.ContentPartTypeInputAudio:
			if part.Audio != nil {
				reqPart.Type = chatTypes.ContentPartTypeInputAudio
				if part.Audio.Data != nil && part.Audio.Format != nil {
					reqPart.InputAudio = &chatTypes.InputAudio{
						Data:   *part.Audio.Data,
						Format: chatTypes.AudioFormat(*part.Audio.Format),
					}
				}
			}

		case "file":
			if part.File != nil {
				reqPart.Type = chatTypes.ContentPartTypeFile
				reqPart.File = &chatTypes.InputFile{
					FileID:   part.File.ID,
					FileData: part.File.Data,
					Filename: part.File.Filename,
				}
			}

		case chatTypes.ContentPartTypeRefusal:
			// 从 VendorExtras 恢复 refusal
			if part.VendorExtras != nil {
				if refusal, ok := part.VendorExtras["refusal"].(*string); ok {
					reqPart.Refusal = refusal
				}
			}

		default:
			// 未知类型从 VendorExtras 恢复
			if part.VendorExtras != nil {
				if originalPart, ok := part.VendorExtras["original_part"].(chatTypes.ContentPart); ok {
					reqPart = originalPart
				}
			}
		}

		result = append(result, reqPart)
	}

	return result, nil
}

// convertToolCallsFromContract 从 Contract 转换工具调用列表。
func convertToolCallsFromContract(toolCalls []types.ToolCall) ([]chatTypes.RequestToolCall, error) {
	result := make([]chatTypes.RequestToolCall, 0, len(toolCalls))

	for _, tc := range toolCalls {
		if tc.ID == nil || tc.Name == nil {
			continue
		}

		reqToolCall := chatTypes.RequestToolCall{
			ID:   *tc.ID,
			Type: chatTypes.RequestToolCallTypeFunction,
			Function: chatTypes.RequestToolCallFunction{
				Name: *tc.Name,
			},
		}

		if tc.Arguments != nil {
			reqToolCall.Function.Arguments = *tc.Arguments
		}

		result = append(result, reqToolCall)
	}

	return result, nil
}

// convertStopFromContract 从 Contract 转换停止条件。
func convertStopFromContract(stop *types.Stop) *chatTypes.StopConfiguration {
	if stop == nil {
		return nil
	}

	result := &chatTypes.StopConfiguration{}

	// 优先使用 List
	if len(stop.List) > 0 {
		result.StringArray = stop.List
	} else if stop.Text != nil {
		result.StringValue = stop.Text
	}

	return result
}

// convertResponseFormatFromContract 从 Contract 转换响应格式。
func convertResponseFormatFromContract(format *types.ResponseFormat) (*chatTypes.FormatUnion, error) {
	if format == nil {
		return nil, nil
	}

	result := &chatTypes.FormatUnion{}

	switch format.Type {
	case string(chatTypes.ResponseFormatTypeText):
		result.Text = &chatTypes.FormatText{
			Type: chatTypes.ResponseFormatTypeText,
		}

	case string(chatTypes.ResponseFormatTypeJSONSchema):
		if format.JSONSchema != nil {
			jsonSchemaSpec := chatTypes.FormatJSONSchemaSpec{}

			// 从 JSONSchema 中提取字段
			if schemaMap, ok := format.JSONSchema.(map[string]interface{}); ok {
				if name, ok := schemaMap["name"].(string); ok {
					jsonSchemaSpec.Name = name
				}
				if desc, ok := schemaMap["description"].(string); ok {
					jsonSchemaSpec.Description = &desc
				}
				if schema, ok := schemaMap["schema"].(map[string]interface{}); ok {
					jsonSchemaSpec.Schema = schema
				}
				if strict, ok := schemaMap["strict"].(bool); ok {
					jsonSchemaSpec.Strict = &strict
				}
			}

			result.JSONSchema = &chatTypes.FormatJSONSchema{
				Type:       chatTypes.ResponseFormatTypeJSONSchema,
				JSONSchema: jsonSchemaSpec,
			}
		}

	case string(chatTypes.ResponseFormatTypeJSONObject):
		result.JSONObject = &chatTypes.FormatJSONObject{
			Type: chatTypes.ResponseFormatTypeJSONObject,
		}
	}

	return result, nil
}

// convertToolsFromContract 从 Contract 转换工具列表。
// Chat Completions 仅支持 function 和 custom 两种工具类型，其他类型将返回错误。
func convertToolsFromContract(tools []types.Tool) ([]chatTypes.ChatToolUnion, error) {
	result := make([]chatTypes.ChatToolUnion, 0, len(tools))

	for _, tool := range tools {
		toolUnion := chatTypes.ChatToolUnion{}

		if tool.Type == "function" && tool.Function != nil {
			// 标准函数工具
			toolFunc := &chatTypes.ChatToolFunction{
				Type: "function",
				Function: shared.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}

			// 从 VendorExtras 恢复 Strict
			if tool.VendorExtras != nil {
				if strict, ok := tool.VendorExtras["strict"].(*bool); ok {
					toolFunc.Function.Strict = strict
				}
			}

			toolUnion.Function = toolFunc
		} else if tool.Type == "custom" {
			// 自定义工具
			if tool.VendorExtras != nil {
				if t, ok := tool.VendorExtras["tool"].(*shared.ToolCustom); ok {
					// 将 shared.ToolCustom 转换为 chatTypes.ChatToolCustom
					toolUnion.Custom = &chatTypes.ChatToolCustom{
						Type:        t.Type,
						Custom:      t.Custom,
						Name:        t.Name,
						Description: t.Description,
						Format:      t.Format,
					}
				}
			}
		} else {
			// 不支持的类型，返回错误
			return nil, errors.New(errors.ErrCodeInvalidArgument,
				fmt.Sprintf("Chat Completions 不支持的工具类型: %s，仅支持 function 和 custom", tool.Type),
			)
		}

		result = append(result, toolUnion)
	}

	return result, nil
}

// convertToolChoiceFromContract 从 Contract 转换工具选择。
func convertToolChoiceFromContract(toolChoice *types.ToolChoice) (*shared.ToolChoiceUnion, error) {
	if toolChoice == nil {
		return nil, nil
	}

	result := &shared.ToolChoiceUnion{}

	if toolChoice.Mode != nil {
		mode := *toolChoice.Mode

		// 简单模式字符串（auto/none/required）
		if mode == "auto" || mode == "none" || mode == "required" {
			result.Auto = &mode
		} else if mode == "function" && toolChoice.Function != nil {
			// 命名函数选择
			result.Named = &shared.ToolChoiceNamed{
				Type: "function",
			}
			result.Named.Function.Name = *toolChoice.Function
		} else if len(toolChoice.Allowed) > 0 {
			// 允许的工具列表
			result.Allowed = &shared.ToolChoiceAllowed{
				Type:  "allowed_tools",
				Mode:  mode,
				Tools: make([]map[string]interface{}, len(toolChoice.Allowed)),
			}
			for i, name := range toolChoice.Allowed {
				result.Allowed.Tools[i] = map[string]interface{}{
					"type": "function",
					"name": name,
				}
			}
		} else {
			// 其他类型（hosted/custom/mcp 等）
			// 尝试从原始数据恢复
			switch mode {
			case "file_search", "web_search_preview", "computer_use_preview",
				"web_search_preview_2025_03_11", "image_generation", "code_interpreter":
				result.Hosted = &shared.ToolChoiceHosted{
					Type: mode,
				}
			case "custom":
				result.NamedCustom = &shared.ToolChoiceNamedCustom{
					Type: mode,
				}
				if toolChoice.Function != nil {
					result.NamedCustom.Custom.Name = *toolChoice.Function
				}
			case "mcp":
				result.NamedMCP = &shared.ToolChoiceNamedMCP{
					Type: mode,
				}
			case "apply_patch":
				result.ApplyPatch = &shared.ToolChoiceApplyPatch{
					Type: mode,
				}
			case "shell":
				result.Shell = &shared.ToolChoiceShell{
					Type: mode,
				}
			default:
				// 默认作为 auto 模式
				result.Auto = &mode
			}
		}
	}

	return result, nil
}
