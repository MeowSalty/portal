package chat

import (
	"encoding/json"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	chatTypes "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ResponseToContract 将 OpenAI Chat 响应转换为统一的 ResponseContract。
func ResponseToContract(resp *chatTypes.Response, log logger.Logger) (*types.ResponseContract, error) {
	if resp == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	contract := &types.ResponseContract{
		Source: types.VendorSourceOpenAIChat,
		ID:     resp.ID,
		Extras: make(map[string]interface{}),
	}

	// 转换顶层字段
	contract.Object = &resp.Object
	contract.Model = &resp.Model
	createdAt := resp.Created
	contract.CreatedAt = &createdAt

	// 转换 ServiceTier 和 SystemFingerprint 到 Extras
	if resp.ServiceTier != nil {
		contract.Extras["openai.chat.service_tier"] = *resp.ServiceTier
	}
	if resp.SystemFingerprint != nil {
		contract.Extras["openai.chat.system_fingerprint"] = *resp.SystemFingerprint
	}

	// 转换 Usage
	if resp.Usage != nil {
		usage, err := convertUsageToContract(resp.Usage)
		if err != nil {
			log.Error("转换 Usage 失败", "error", err)
			return nil, errors.Wrap(errors.ErrCodeInternal, "转换 Usage 失败", err)
		}
		contract.Usage = usage
	}

	// 转换 Choices
	if len(resp.Choices) > 0 {
		contract.Choices = make([]types.ResponseChoice, 0, len(resp.Choices))
		for _, choice := range resp.Choices {
			contractChoice, err := convertChoiceToContract(&choice, log)
			if err != nil {
				log.Error("转换 Choice 失败", "error", err)
				return nil, errors.Wrap(errors.ErrCodeInternal, "转换 Choice 失败", err)
			}
			contract.Choices = append(contract.Choices, *contractChoice)
		}
	}

	return contract, nil
}

// convertUsageToContract 转换 Usage 信息。
func convertUsageToContract(usage *chatTypes.Usage) (*types.ResponseUsage, error) {
	if usage == nil {
		return nil, nil
	}

	result := &types.ResponseUsage{
		Extras: make(map[string]interface{}),
	}

	// 基础 token 计数
	inputTokens := usage.PromptTokens
	result.InputTokens = &inputTokens
	outputTokens := usage.CompletionTokens
	result.OutputTokens = &outputTokens
	totalTokens := usage.TotalTokens
	result.TotalTokens = &totalTokens

	// 将细分字段放入 Extras
	if usage.PromptTokensDetails != nil {
		result.Extras["openai.chat.prompt_tokens_details"] = usage.PromptTokensDetails
	}
	if usage.CompletionTokensDetails != nil {
		result.Extras["openai.chat.completion_tokens_details"] = usage.CompletionTokensDetails
	}

	return result, nil
}

// convertChoiceToContract 转换 Choice。
func convertChoiceToContract(choice *chatTypes.Choice, log logger.Logger) (*types.ResponseChoice, error) {
	contractChoice := &types.ResponseChoice{
		Extras: make(map[string]interface{}),
	}

	// 转换 Index
	index := choice.Index
	contractChoice.Index = &index

	// 转换 FinishReason
	finishReason := mapFinishReasonToContract(choice.FinishReason)
	contractChoice.FinishReason = &finishReason
	nativeFinishReason := string(choice.FinishReason)
	contractChoice.NativeFinishReason = &nativeFinishReason

	// 转换 Logprobs
	if choice.Logprobs != nil {
		logprobs, err := convertLogprobsToContract(choice.Logprobs)
		if err != nil {
			log.Error("转换 Logprobs 失败", "error", err)
			return nil, err
		}
		contractChoice.Logprobs = logprobs
	}

	// 转换 Message
	message, err := convertMessageToContract(&choice.Message, log)
	if err != nil {
		log.Error("转换 Message 失败", "error", err)
		return nil, err
	}
	contractChoice.Message = message

	return contractChoice, nil
}

// mapFinishReasonToContract 映射 OpenAI Chat FinishReason 到统一的 FinishReason。
func mapFinishReasonToContract(finishReason chatTypes.FinishReason) types.ResponseFinishReason {
	switch finishReason {
	case chatTypes.FinishReasonStop:
		return types.ResponseFinishReasonStop
	case chatTypes.FinishReasonLength:
		return types.ResponseFinishReasonLength
	case chatTypes.FinishReasonToolCalls:
		return types.ResponseFinishReasonToolCalls
	case chatTypes.FinishReasonContentFilter:
		return types.ResponseFinishReasonContentFilter
	case chatTypes.FinishReasonFunctionCall:
		return types.ResponseFinishReasonToolCalls
	default:
		return types.ResponseFinishReasonUnknown
	}
}

// convertLogprobsToContract 转换 Logprobs。
func convertLogprobsToContract(logprobs *chatTypes.Logprobs) (*types.ResponseLogprobs, error) {
	if logprobs == nil {
		return nil, nil
	}

	result := &types.ResponseLogprobs{
		Extras: make(map[string]interface{}),
	}

	// 转换 Content
	if logprobs.Content != nil {
		result.Content = make([]types.ResponseTokenLogprob, 0, len(*logprobs.Content))
		for _, tokenLogprob := range *logprobs.Content {
			contractToken := types.ResponseTokenLogprob{
				Token:   tokenLogprob.Token,
				Logprob: tokenLogprob.Logprob,
			}
			if tokenLogprob.Bytes != nil {
				contractToken.Bytes = *tokenLogprob.Bytes
			}
			if len(tokenLogprob.TopLogprobs) > 0 {
				contractToken.TopLogprobs = make([]types.ResponseTokenLogprobTop, 0, len(tokenLogprob.TopLogprobs))
				for _, top := range tokenLogprob.TopLogprobs {
					contractTop := types.ResponseTokenLogprobTop{
						Token:   top.Token,
						Logprob: top.Logprob,
					}
					if top.Bytes != nil {
						contractTop.Bytes = *top.Bytes
					}
					contractToken.TopLogprobs = append(contractToken.TopLogprobs, contractTop)
				}
			}
			result.Content = append(result.Content, contractToken)
		}
	}

	// 转换 Refusal
	if logprobs.Refusal != nil {
		result.Refusal = make([]types.ResponseTokenLogprob, 0, len(*logprobs.Refusal))
		for _, tokenLogprob := range *logprobs.Refusal {
			contractToken := types.ResponseTokenLogprob{
				Token:   tokenLogprob.Token,
				Logprob: tokenLogprob.Logprob,
			}
			if tokenLogprob.Bytes != nil {
				contractToken.Bytes = *tokenLogprob.Bytes
			}
			if len(tokenLogprob.TopLogprobs) > 0 {
				contractToken.TopLogprobs = make([]types.ResponseTokenLogprobTop, 0, len(tokenLogprob.TopLogprobs))
				for _, top := range tokenLogprob.TopLogprobs {
					contractTop := types.ResponseTokenLogprobTop{
						Token:   top.Token,
						Logprob: top.Logprob,
					}
					if top.Bytes != nil {
						contractTop.Bytes = *top.Bytes
					}
					contractToken.TopLogprobs = append(contractToken.TopLogprobs, contractTop)
				}
			}
			result.Refusal = append(result.Refusal, contractToken)
		}
	}

	return result, nil
}

// convertMessageToContract 转换 Message。
func convertMessageToContract(message *chatTypes.Message, log logger.Logger) (*types.ResponseMessage, error) {
	contractMessage := &types.ResponseMessage{
		Extras: make(map[string]interface{}),
	}

	// 转换 Role
	role := string(message.Role)
	contractMessage.Role = &role

	// 转换 Content
	contractMessage.Content = message.Content

	// 转换 Refusal
	contractMessage.Refusal = message.Refusal

	// 转换 Audio
	if message.Audio != nil {
		audio := &types.ResponseAudio{
			Extras: make(map[string]interface{}),
		}
		audio.ID = &message.Audio.ID
		audio.Data = &message.Audio.Data
		expiresAt := message.Audio.ExpiresAt
		audio.ExpiresAt = &expiresAt
		audio.Transcript = &message.Audio.Transcript
		contractMessage.Audio = audio
	}

	// 转换 Annotations
	if len(message.Annotations) > 0 {
		annotations, err := convertAnnotationsToContract(message.Annotations)
		if err != nil {
			log.Error("转换 Annotations 失败", "error", err)
			return nil, err
		}
		// 将 annotations 作为 Part 添加
		for _, annotation := range annotations {
			part := types.ResponseContentPart{
				Type:        "text",
				Annotations: []types.ResponseAnnotation{annotation},
				Extras:      make(map[string]interface{}),
			}
			contractMessage.Parts = append(contractMessage.Parts, part)
		}
	}

	// 转换 ToolCalls（包括 FunctionCall）
	if message.FunctionCall != nil {
		// 旧版 function_call 转换为 ToolCall
		toolCall := types.ResponseToolCall{
			Name:      &message.FunctionCall.Name,
			Arguments: &message.FunctionCall.Arguments,
			Extras:    make(map[string]interface{}),
		}
		toolType := "function"
		toolCall.Type = &toolType
		toolCall.Extras["openai.chat.legacy_function_call"] = true
		contractMessage.ToolCalls = append(contractMessage.ToolCalls, toolCall)
	}

	if len(message.ToolCalls) > 0 {
		for _, toolCall := range message.ToolCalls {
			contractToolCall, err := convertToolCallToContract(&toolCall)
			if err != nil {
				log.Error("转换 ToolCall 失败", "error", err)
				return nil, err
			}
			contractMessage.ToolCalls = append(contractMessage.ToolCalls, *contractToolCall)
		}
	}

	// 保存 ExtraFields 到 Extras
	if len(message.ExtraFields) > 0 {
		extraFieldsJSON, err := json.Marshal(message.ExtraFields)
		if err != nil {
			log.Warn("序列化 ExtraFields 失败", "error", err)
		} else {
			contractMessage.Extras["openai.chat.extra_fields"] = string(extraFieldsJSON)
		}
	}

	return contractMessage, nil
}

// convertAnnotationsToContract 转换 Annotations。
func convertAnnotationsToContract(annotations []chatTypes.MessageAnnotation) ([]types.ResponseAnnotation, error) {
	result := make([]types.ResponseAnnotation, 0, len(annotations))

	for _, annotation := range annotations {
		contractAnnotation := types.ResponseAnnotation{
			Type:   annotation.Type,
			Extras: make(map[string]interface{}),
		}

		if annotation.URLCitation != nil {
			contractAnnotation.StartIndex = &annotation.URLCitation.StartIndex
			contractAnnotation.EndIndex = &annotation.URLCitation.EndIndex
			contractAnnotation.URL = &annotation.URLCitation.URL
			contractAnnotation.Title = &annotation.URLCitation.Title
		}

		result = append(result, contractAnnotation)
	}

	return result, nil
}

// convertToolCallToContract 转换 ToolCall。
func convertToolCallToContract(toolCall *chatTypes.MessageToolCall) (*types.ResponseToolCall, error) {
	contractToolCall := &types.ResponseToolCall{
		ID:     &toolCall.ID,
		Extras: make(map[string]interface{}),
	}

	toolType := string(toolCall.Type)
	contractToolCall.Type = &toolType

	if toolCall.Function != nil {
		contractToolCall.Name = &toolCall.Function.Name
		contractToolCall.Arguments = &toolCall.Function.Arguments
	}

	if toolCall.Custom != nil {
		contractToolCall.Name = &toolCall.Custom.Name
		// Custom 的 Input 是字符串，存入 Payload
		if contractToolCall.Payload == nil {
			contractToolCall.Payload = make(map[string]interface{})
		}
		contractToolCall.Payload["input"] = toolCall.Custom.Input
		contractToolCall.Extras["openai.chat.custom_tool"] = true
	}

	return contractToolCall, nil
}
