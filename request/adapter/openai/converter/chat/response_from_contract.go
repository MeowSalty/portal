package chat

import (
	"encoding/json"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	chatTypes "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ResponseFromContract 将 ResponseContract 转换回 OpenAI Chat 响应。
func ResponseFromContract(contract *types.ResponseContract, log logger.Logger) (*chatTypes.Response, error) {
	if contract == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	resp := &chatTypes.Response{
		ID:      contract.ID,
		Object:  "chat.completion",
		Choices: []chatTypes.Choice{},
	}

	// 转换 Model
	if contract.Model != nil {
		resp.Model = *contract.Model
	}

	// 转换 CreatedAt
	if contract.CreatedAt != nil {
		resp.Created = *contract.CreatedAt
	}

	// 从 Extras 恢复 Object
	if object, ok := contract.Extras["openai.chat.object"].(string); ok {
		resp.Object = object
	}

	// 从 Extras 恢复 ServiceTier
	if serviceTier, ok := contract.Extras["openai.chat.service_tier"].(string); ok {
		resp.ServiceTier = &serviceTier
	}

	// 从 Extras 恢复 SystemFingerprint
	if systemFingerprint, ok := contract.Extras["openai.chat.system_fingerprint"].(string); ok {
		resp.SystemFingerprint = &systemFingerprint
	}

	// 转换 Usage
	if contract.Usage != nil {
		usage, err := convertUsageFromContract(contract.Usage)
		if err != nil {
			log.Error("转换 Usage 失败", "error", err)
			return nil, errors.Wrap(errors.ErrCodeInternal, "转换 Usage 失败", err)
		}
		resp.Usage = usage
	}

	// 转换 Choices
	if len(contract.Choices) > 0 {
		resp.Choices = make([]chatTypes.Choice, 0, len(contract.Choices))
		for _, choice := range contract.Choices {
			chatChoice, err := convertChoiceFromContract(&choice, log)
			if err != nil {
				log.Error("转换 Choice 失败", "error", err)
				return nil, errors.Wrap(errors.ErrCodeInternal, "转换 Choice 失败", err)
			}
			resp.Choices = append(resp.Choices, *chatChoice)
		}
	}

	return resp, nil
}

// convertUsageFromContract 从 Contract 转换 Usage。
func convertUsageFromContract(usage *types.ResponseUsage) (*chatTypes.Usage, error) {
	if usage == nil {
		return nil, nil
	}

	result := &chatTypes.Usage{}

	// 基础 token 计数
	if usage.InputTokens != nil {
		result.PromptTokens = *usage.InputTokens
	}
	if usage.OutputTokens != nil {
		result.CompletionTokens = *usage.OutputTokens
	}
	if usage.TotalTokens != nil {
		result.TotalTokens = *usage.TotalTokens
	}

	// 从 Extras 恢复细分字段
	if val, ok := usage.Extras["openai.chat.prompt_tokens_details"]; ok {
		if promptDetails, ok := val.(*chatTypes.PromptTokensDetails); ok {
			result.PromptTokensDetails = promptDetails
		}
	}

	if val, ok := usage.Extras["openai.chat.completion_tokens_details"]; ok {
		if completionDetails, ok := val.(*chatTypes.CompletionTokensDetails); ok {
			result.CompletionTokensDetails = completionDetails
		}
	}

	return result, nil
}

// convertChoiceFromContract 从 Contract 转换 Choice。
func convertChoiceFromContract(choice *types.ResponseChoice, log logger.Logger) (*chatTypes.Choice, error) {
	chatChoice := &chatTypes.Choice{
		FinishReason: chatTypes.FinishReasonStop,
		Logprobs:     nil,
	}

	// 转换 Index
	if choice.Index != nil {
		chatChoice.Index = *choice.Index
	}

	// 转换 FinishReason
	if choice.FinishReason != nil {
		chatChoice.FinishReason = mapFinishReasonFromContract(*choice.FinishReason)
	}

	// 转换 Logprobs
	if choice.Logprobs != nil {
		logprobs := convertLogprobsFromContract(choice.Logprobs)
		chatChoice.Logprobs = &logprobs
	}

	// 转换 Message
	if choice.Message != nil {
		message, err := convertMessageFromContract(choice.Message, log)
		if err != nil {
			log.Error("转换 Message 失败", "error", err)
			return nil, err
		}
		chatChoice.Message = *message
	}

	return chatChoice, nil
}

// mapFinishReasonFromContract 映射统一的 FinishReason 到 OpenAI Chat FinishReason。
func mapFinishReasonFromContract(finishReason types.ResponseFinishReason) chatTypes.FinishReason {
	switch finishReason {
	case types.ResponseFinishReasonStop:
		return chatTypes.FinishReasonStop
	case types.ResponseFinishReasonLength:
		return chatTypes.FinishReasonLength
	case types.ResponseFinishReasonToolCalls:
		return chatTypes.FinishReasonToolCalls
	case types.ResponseFinishReasonContentFilter:
		return chatTypes.FinishReasonContentFilter
	default:
		return chatTypes.FinishReasonStop
	}
}

// convertLogprobsFromContract 从 Contract 转换 Logprobs。
func convertLogprobsFromContract(logprobs *types.ResponseLogprobs) chatTypes.Logprobs {
	result := chatTypes.Logprobs{}

	// 转换 Content
	if len(logprobs.Content) > 0 {
		content := make([]chatTypes.TokenLogprob, 0, len(logprobs.Content))
		for _, token := range logprobs.Content {
			chatToken := chatTypes.TokenLogprob{
				Token:   token.Token,
				Logprob: token.Logprob,
			}
			if len(token.Bytes) > 0 {
				chatToken.Bytes = &token.Bytes
			}
			if len(token.TopLogprobs) > 0 {
				chatToken.TopLogprobs = make([]chatTypes.TokenLogprobTopLogprob, 0, len(token.TopLogprobs))
				for _, top := range token.TopLogprobs {
					chatTop := chatTypes.TokenLogprobTopLogprob{
						Token:   top.Token,
						Logprob: top.Logprob,
					}
					if len(top.Bytes) > 0 {
						chatTop.Bytes = &top.Bytes
					}
					chatToken.TopLogprobs = append(chatToken.TopLogprobs, chatTop)
				}
			}
			content = append(content, chatToken)
		}
		result.Content = &content
	}

	// 转换 Refusal
	if len(logprobs.Refusal) > 0 {
		refusal := make([]chatTypes.TokenLogprob, 0, len(logprobs.Refusal))
		for _, token := range logprobs.Refusal {
			chatToken := chatTypes.TokenLogprob{
				Token:   token.Token,
				Logprob: token.Logprob,
			}
			if len(token.Bytes) > 0 {
				chatToken.Bytes = &token.Bytes
			}
			if len(token.TopLogprobs) > 0 {
				chatToken.TopLogprobs = make([]chatTypes.TokenLogprobTopLogprob, 0, len(token.TopLogprobs))
				for _, top := range token.TopLogprobs {
					chatTop := chatTypes.TokenLogprobTopLogprob{
						Token:   top.Token,
						Logprob: top.Logprob,
					}
					if len(top.Bytes) > 0 {
						chatTop.Bytes = &top.Bytes
					}
					chatToken.TopLogprobs = append(chatToken.TopLogprobs, chatTop)
				}
			}
			refusal = append(refusal, chatToken)
		}
		result.Refusal = &refusal
	}

	return result
}

// convertMessageFromContract 从 Contract 转换 Message。
func convertMessageFromContract(message *types.ResponseMessage, log logger.Logger) (*chatTypes.Message, error) {
	chatMessage := &chatTypes.Message{
		Role:        chatTypes.ChatResponseMessageRoleAssistant,
		ExtraFields: make(map[string]interface{}),
	}

	// 转换 Role
	if message.Role != nil {
		chatMessage.Role = chatTypes.ChatResponseMessageRole(*message.Role)
	}

	// 优先使用 Content
	chatMessage.Content = message.Content

	// 若 Parts 仅包含 text 类型，可合并回 content
	if message.Content == nil && len(message.Parts) > 0 {
		var textContent string
		allText := true
		for _, part := range message.Parts {
			if part.Type == "text" && part.Text != nil {
				if textContent != "" {
					textContent += "\n"
				}
				textContent += *part.Text
			} else if part.Type != "text" {
				allText = false
			}
		}
		if allText && textContent != "" {
			chatMessage.Content = &textContent
		}
	}

	// 转换 Refusal
	chatMessage.Refusal = message.Refusal

	// 转换 Audio
	if message.Audio != nil {
		audio := chatTypes.ResponseAudio{}
		if message.Audio.ID != nil {
			audio.ID = *message.Audio.ID
		}
		if message.Audio.Data != nil {
			audio.Data = *message.Audio.Data
		}
		if message.Audio.ExpiresAt != nil {
			audio.ExpiresAt = *message.Audio.ExpiresAt
		}
		if message.Audio.Transcript != nil {
			audio.Transcript = *message.Audio.Transcript
		}
		chatMessage.Audio = &audio
	}

	// 从 Parts 中提取 Annotations
	if len(message.Parts) > 0 {
		annotations := convertAnnotationsFromContract(message.Parts)
		if len(annotations) > 0 {
			chatMessage.Annotations = annotations
		}
	}

	// 转换 ToolCalls
	if len(message.ToolCalls) > 0 {
		toolCalls := make([]chatTypes.MessageToolCall, 0, len(message.ToolCalls))
		for _, toolCall := range message.ToolCalls {
			chatToolCall, err := convertToolCallFromContract(&toolCall)
			if err != nil {
				log.Error("转换 ToolCall 失败", "error", err)
				return nil, err
			}
			toolCalls = append(toolCalls, *chatToolCall)
		}
		chatMessage.ToolCalls = toolCalls
	}

	// 从 Extras 恢复 ExtraFields
	if extraFieldsJSON, ok := message.Extras["openai.chat.extra_fields"].(string); ok {
		var extraFields map[string]interface{}
		if err := json.Unmarshal([]byte(extraFieldsJSON), &extraFields); err == nil {
			chatMessage.ExtraFields = extraFields
		} else {
			log.Warn("反序列化 ExtraFields 失败", "error", err)
		}
	}

	return chatMessage, nil
}

// convertAnnotationsFromContract 从 Parts 转换 Annotations。
func convertAnnotationsFromContract(parts []types.ResponseContentPart) []chatTypes.MessageAnnotation {
	var annotations []chatTypes.MessageAnnotation

	for _, part := range parts {
		if len(part.Annotations) > 0 {
			for _, annotation := range part.Annotations {
				chatAnnotation := chatTypes.MessageAnnotation{
					Type: annotation.Type,
				}

				// 仅处理 URL 引用类型
				if annotation.Type == "url_citation" {
					urlCitation := &chatTypes.URLCitation{}
					if annotation.StartIndex != nil {
						urlCitation.StartIndex = *annotation.StartIndex
					}
					if annotation.EndIndex != nil {
						urlCitation.EndIndex = *annotation.EndIndex
					}
					if annotation.URL != nil {
						urlCitation.URL = *annotation.URL
					}
					if annotation.Title != nil {
						urlCitation.Title = *annotation.Title
					}
					chatAnnotation.URLCitation = urlCitation
				}

				annotations = append(annotations, chatAnnotation)
			}
		}
	}

	return annotations
}

// convertToolCallFromContract 从 Contract 转换 ToolCall。
func convertToolCallFromContract(toolCall *types.ResponseToolCall) (*chatTypes.MessageToolCall, error) {
	chatToolCall := &chatTypes.MessageToolCall{}

	// 检查是否是旧版 function_call
	if isLegacy, ok := toolCall.Extras["openai.chat.legacy_function_call"].(bool); ok && isLegacy {
		// 旧版格式，使用 FunctionCall 字段
		if toolCall.Name != nil && toolCall.Arguments != nil {
			chatToolCall.Type = chatTypes.ToolCallTypeFunction
			chatToolCall.Function = &chatTypes.ToolCallFunction{
				Name:      *toolCall.Name,
				Arguments: *toolCall.Arguments,
			}
			return chatToolCall, nil
		}
	}

	// 新版格式
	if toolCall.ID != nil {
		chatToolCall.ID = *toolCall.ID
	}

	if toolCall.Type != nil {
		chatToolCall.Type = chatTypes.ToolCallType(*toolCall.Type)
	} else {
		chatToolCall.Type = chatTypes.ToolCallTypeFunction
	}

	// 检查是否是自定义工具
	if isCustom, ok := toolCall.Extras["openai.chat.custom_tool"].(bool); ok && isCustom {
		if toolCall.Name != nil {
			chatToolCall.Custom = &chatTypes.ToolCallCustom{
				Name: *toolCall.Name,
			}
		}
		if input, ok := toolCall.Payload["input"].(string); ok {
			if chatToolCall.Custom != nil {
				chatToolCall.Custom.Input = input
			}
		}
		return chatToolCall, nil
	}

	// 普通函数工具
	if toolCall.Name != nil || toolCall.Arguments != nil {
		function := &chatTypes.ToolCallFunction{}
		if toolCall.Name != nil {
			function.Name = *toolCall.Name
		}
		if toolCall.Arguments != nil {
			function.Arguments = *toolCall.Arguments
		}
		chatToolCall.Function = function
	}

	return chatToolCall, nil
}
