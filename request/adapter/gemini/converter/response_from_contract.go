package converter

import (
	"encoding/json"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
)

// ResponseFromContract 将 ResponseContract 转换为 Gemini 响应
func ResponseFromContract(contract *adapterTypes.ResponseContract) (*geminiTypes.Response, error) {
	if contract == nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "ResponseContract 为空")
	}

	if contract.Source != adapterTypes.VendorSourceGemini {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "ResponseContract 来源不是 Gemini")
	}

	resp := &geminiTypes.Response{
		ResponseID: contract.ID,
	}

	// 映射 Model
	if contract.Model != nil {
		resp.ModelVersion = *contract.Model
	}

	// 映射 Usage
	if contract.Usage != nil {
		resp.UsageMetadata = convertUsageFromContract(contract.Usage)
	}

	// 恢复 PromptFeedback
	if val, ok := contract.Extras["gemini.prompt_feedback"]; ok {
		if feedback, ok := val.(*geminiTypes.PromptFeedback); ok {
			resp.PromptFeedback = feedback
		}
	}

	// 恢复 ModelStatus
	if val, ok := contract.Extras["gemini.model_status"]; ok {
		if status, ok := val.(*geminiTypes.ModelStatus); ok {
			resp.ModelStatus = status
		}
	}

	// 映射 Choices 为 Candidates
	if len(contract.Choices) > 0 {
		resp.Candidates = make([]geminiTypes.Candidate, 0, len(contract.Choices))
		for _, choice := range contract.Choices {
			candidate, err := convertChoiceToCandidate(&choice)
			if err != nil {
				logger.Default().Warn("转换候选响应失败", "error", err)
				continue
			}
			resp.Candidates = append(resp.Candidates, *candidate)
		}
	}

	return resp, nil
}

// convertUsageFromContract 转换使用统计
func convertUsageFromContract(usage *adapterTypes.ResponseUsage) *geminiTypes.UsageMetadata {
	result := &geminiTypes.UsageMetadata{}

	// 基础 token 计数
	if usage.InputTokens != nil {
		result.PromptTokenCount = int32(*usage.InputTokens)
	}
	if usage.OutputTokens != nil {
		result.CandidatesTokenCount = int32(*usage.OutputTokens)
	}
	if usage.TotalTokens != nil {
		result.TotalTokenCount = int32(*usage.TotalTokens)
	}

	// 恢复细分字段
	if val, ok := usage.Extras["gemini.cached_content_token_count"]; ok {
		if count, ok := val.(int32); ok {
			result.CachedContentTokenCount = count
		}
	}
	if val, ok := usage.Extras["gemini.tool_use_prompt_token_count"]; ok {
		if count, ok := val.(int32); ok {
			result.ToolUsePromptTokenCount = count
		}
	}
	if val, ok := usage.Extras["gemini.thoughts_token_count"]; ok {
		if count, ok := val.(int32); ok {
			result.ThoughtsTokenCount = count
		}
	}
	if val, ok := usage.Extras["gemini.prompt_tokens_details"]; ok {
		if details, ok := val.([]geminiTypes.ModalityTokenCount); ok {
			result.PromptTokensDetails = details
		}
	}
	if val, ok := usage.Extras["gemini.cache_tokens_details"]; ok {
		if details, ok := val.([]geminiTypes.ModalityTokenCount); ok {
			result.CacheTokensDetails = details
		}
	}
	if val, ok := usage.Extras["gemini.candidates_tokens_details"]; ok {
		if details, ok := val.([]geminiTypes.ModalityTokenCount); ok {
			result.CandidatesTokensDetails = details
		}
	}
	if val, ok := usage.Extras["gemini.tool_use_prompt_tokens_details"]; ok {
		if details, ok := val.([]geminiTypes.ModalityTokenCount); ok {
			result.ToolUsePromptTokensDetails = details
		}
	}

	return result
}

// convertChoiceToCandidate 转换候选响应
func convertChoiceToCandidate(choice *adapterTypes.ResponseChoice) (*geminiTypes.Candidate, error) {
	candidate := &geminiTypes.Candidate{}

	// 映射 Index
	if choice.Index != nil {
		candidate.Index = int32(*choice.Index)
	}

	// 映射 FinishReason
	if choice.NativeFinishReason != nil {
		candidate.FinishReason = *choice.NativeFinishReason
	} else if choice.FinishReason != nil {
		candidate.FinishReason = mapFinishReasonToGemini(*choice.FinishReason)
	}

	// 恢复 FinishMessage
	if val, ok := choice.Extras["gemini.finish_message"]; ok {
		if msg, ok := val.(string); ok {
			candidate.FinishMessage = msg
		}
	}

	// 恢复 SafetyRatings
	if val, ok := choice.Extras["gemini.safety_ratings"]; ok {
		if ratings, ok := val.([]geminiTypes.SafetyRating); ok {
			candidate.SafetyRatings = ratings
		}
	}

	// 恢复 CitationMetadata
	if val, ok := choice.Extras["gemini.citation_metadata"]; ok {
		if metadata, ok := val.(*geminiTypes.CitationMetadata); ok {
			candidate.CitationMetadata = metadata
		}
	}

	// 恢复 TokenCount
	if val, ok := choice.Extras["gemini.token_count"]; ok {
		if count, ok := val.(int32); ok {
			candidate.TokenCount = count
		}
	}

	// 恢复 GroundingAttributions
	if val, ok := choice.Extras["gemini.grounding_attributions"]; ok {
		if attrs, ok := val.([]geminiTypes.GroundingAttribution); ok {
			candidate.GroundingAttributions = attrs
		}
	}

	// 恢复 GroundingMetadata
	if val, ok := choice.Extras["gemini.grounding_metadata"]; ok {
		if metadata, ok := val.(*geminiTypes.GroundingMetadata); ok {
			candidate.GroundingMetadata = metadata
		}
	}

	// 恢复 AvgLogprobs
	if val, ok := choice.Extras["gemini.avg_logprobs"]; ok {
		if logprobs, ok := val.(float64); ok {
			candidate.AvgLogprobs = &logprobs
		}
	}

	// 恢复 LogprobsResult
	if val, ok := choice.Extras["gemini.logprobs_result"]; ok {
		if result, ok := val.(*geminiTypes.LogprobsResult); ok {
			candidate.LogprobsResult = result
		}
	}

	// 恢复 URLContextMetadata
	if val, ok := choice.Extras["gemini.url_context_metadata"]; ok {
		if metadata, ok := val.(*geminiTypes.URLContextMetadata); ok {
			candidate.URLContextMetadata = metadata
		}
	}

	// 转换 Message 为 Content
	if choice.Message != nil {
		content, err := convertMessageToContent(choice.Message)
		if err != nil {
			return nil, err
		}
		candidate.Content = *content
	}

	return candidate, nil
}

// convertMessageToContent 转换 ResponseMessage 为 Content
func convertMessageToContent(message *adapterTypes.ResponseMessage) (*geminiTypes.Content, error) {
	content := &geminiTypes.Content{}

	// 映射 Role
	if message.Role != nil {
		content.Role = *message.Role
	} else {
		// 默认为 model
		content.Role = "model"
	}

	// 优先使用原始 Parts
	if val, ok := message.Extras["gemini.parts"]; ok {
		if parts, ok := val.([]geminiTypes.Part); ok {
			content.Parts = parts
			return content, nil
		}
	}

	// 否则从 Parts、ToolCalls、ToolResults 重建
	content.Parts = make([]geminiTypes.Part, 0)

	// 处理 Parts
	for _, part := range message.Parts {
		if part.Type == "text" && part.Text != nil {
			gPart := geminiTypes.Part{
				Text: part.Text,
			}

			// 恢复 Part 特有字段
			if val, ok := part.Extras["gemini.thought"]; ok {
				if thought, ok := val.(bool); ok {
					gPart.Thought = &thought
				}
			}
			if val, ok := part.Extras["gemini.thought_signature"]; ok {
				if sig, ok := val.(string); ok {
					gPart.ThoughtSignature = &sig
				}
			}
			if val, ok := part.Extras["gemini.part_metadata"]; ok {
				if metadata, ok := val.(map[string]interface{}); ok {
					gPart.PartMetadata = metadata
				}
			}
			if val, ok := part.Extras["gemini.media_resolution"]; ok {
				if resolution, ok := val.(*geminiTypes.MediaResolution); ok {
					gPart.MediaResolution = resolution
				}
			}

			content.Parts = append(content.Parts, gPart)
		}
	}

	// 处理 ToolCalls
	for _, toolCall := range message.ToolCalls {
		fc := convertToolCallToFunctionCall(&toolCall)
		content.Parts = append(content.Parts, geminiTypes.Part{
			FunctionCall: fc,
		})
	}

	// 处理 ToolResults
	for _, toolResult := range message.ToolResults {
		fr := convertToolResultToFunctionResponse(&toolResult)
		content.Parts = append(content.Parts, geminiTypes.Part{
			FunctionResponse: fr,
		})
	}

	// 如果没有 Parts，但有 Content，创建一个文本 Part
	if len(content.Parts) == 0 && message.Content != nil {
		content.Parts = append(content.Parts, geminiTypes.Part{
			Text: message.Content,
		})
	}

	return content, nil
}

// convertToolCallToFunctionCall 转换工具调用为函数调用
func convertToolCallToFunctionCall(toolCall *adapterTypes.ResponseToolCall) *geminiTypes.FunctionCall {
	fc := &geminiTypes.FunctionCall{
		Args: toolCall.Payload,
	}

	if toolCall.ID != nil {
		fc.ID = toolCall.ID
	}

	if toolCall.Name != nil {
		fc.Name = *toolCall.Name
	}

	// 如果有 Arguments 字符串，尝试解析为 Args
	if toolCall.Arguments != nil && len(fc.Args) == 0 {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(*toolCall.Arguments), &args); err == nil {
			fc.Args = args
		}
	}

	return fc
}

// convertToolResultToFunctionResponse 转换工具结果为函数响应
func convertToolResultToFunctionResponse(toolResult *adapterTypes.ResponseToolResult) *geminiTypes.FunctionResponse {
	fr := &geminiTypes.FunctionResponse{
		Response: toolResult.Payload,
	}

	if toolResult.ID != nil {
		fr.ID = toolResult.ID
	}

	if toolResult.Name != nil {
		fr.Name = *toolResult.Name
	}

	// 恢复其他字段
	if val, ok := toolResult.Extras["gemini.parts"]; ok {
		if parts, ok := val.([]geminiTypes.FunctionResponsePart); ok {
			fr.Parts = parts
		}
	}
	if val, ok := toolResult.Extras["gemini.will_continue"]; ok {
		if willContinue, ok := val.(bool); ok {
			fr.WillContinue = &willContinue
		}
	}
	if val, ok := toolResult.Extras["gemini.scheduling"]; ok {
		if scheduling, ok := val.(string); ok {
			fr.Scheduling = &scheduling
		}
	}

	return fr
}

// mapFinishReasonToGemini 映射统一格式到 Gemini FinishReason
func mapFinishReasonToGemini(reason adapterTypes.ResponseFinishReason) geminiTypes.FinishReason {
	switch reason {
	case adapterTypes.ResponseFinishReasonStop:
		return geminiTypes.FinishReasonStop
	case adapterTypes.ResponseFinishReasonLength:
		return geminiTypes.FinishReasonMaxTokens
	case adapterTypes.ResponseFinishReasonContentFilter:
		return geminiTypes.FinishReasonSafety
	case adapterTypes.ResponseFinishReasonRecitation:
		return geminiTypes.FinishReasonRecitation
	case adapterTypes.ResponseFinishReasonLanguage:
		return geminiTypes.FinishReasonLanguage
	case adapterTypes.ResponseFinishReasonToolCallMalformed:
		return geminiTypes.FinishReasonMalformedFunction
	case adapterTypes.ResponseFinishReasonToolCallUnexpected:
		return geminiTypes.FinishReasonUnexpectedToolCall
	case adapterTypes.ResponseFinishReasonToolCallLimit:
		return geminiTypes.FinishReasonTooManyToolCalls
	case adapterTypes.ResponseFinishReasonThoughtSignatureMissing:
		return geminiTypes.FinishReasonMissingThoughtSig
	default:
		return geminiTypes.FinishReasonOther
	}
}
