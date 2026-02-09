package converter

import (
	"encoding/json"
	"fmt"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
)

// ResponseToContract 将 Gemini 响应转换为 ResponseContract
func ResponseToContract(resp *geminiTypes.Response, log logger.Logger) (*adapterTypes.ResponseContract, error) {
	if resp == nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "Gemini 响应为空")
	}

	contract := &adapterTypes.ResponseContract{
		Source: adapterTypes.VendorSourceGemini,
		ID:     resp.ResponseID,
		Extras: make(map[string]interface{}),
	}

	// 映射 Model
	if resp.ModelVersion != "" {
		contract.Model = &resp.ModelVersion
	}

	// 映射 Usage
	if resp.UsageMetadata != nil {
		contract.Usage = convertUsageToContract(resp.UsageMetadata)
	}

	// 保存 PromptFeedback 到 Extras
	if resp.PromptFeedback != nil {
		contract.Extras["gemini.prompt_feedback"] = resp.PromptFeedback
	}

	// 保存 ModelStatus 到 Extras
	if resp.ModelStatus != nil {
		contract.Extras["gemini.model_status"] = resp.ModelStatus
	}

	// 映射 Candidates
	if len(resp.Candidates) > 0 {
		contract.Choices = make([]adapterTypes.ResponseChoice, 0, len(resp.Candidates))
		for _, candidate := range resp.Candidates {
			choice, err := convertCandidateToChoice(&candidate)
			if err != nil {
				log.Warn("转换候选响应失败", "error", err)
				continue
			}
			contract.Choices = append(contract.Choices, *choice)
		}
	}

	return contract, nil
}

// convertUsageToContract 转换使用统计
func convertUsageToContract(usage *geminiTypes.UsageMetadata) *adapterTypes.ResponseUsage {
	result := &adapterTypes.ResponseUsage{
		Extras: make(map[string]interface{}),
	}

	// 基础 token 计数
	if usage.PromptTokenCount > 0 {
		inputTokens := int(usage.PromptTokenCount)
		result.InputTokens = &inputTokens
	}
	if usage.CandidatesTokenCount > 0 {
		outputTokens := int(usage.CandidatesTokenCount)
		result.OutputTokens = &outputTokens
	}
	if usage.TotalTokenCount > 0 {
		totalTokens := int(usage.TotalTokenCount)
		result.TotalTokens = &totalTokens
	}

	// 保存细分字段到 Extras
	if usage.CachedContentTokenCount > 0 {
		result.Extras["gemini.cached_content_token_count"] = usage.CachedContentTokenCount
	}
	if usage.ToolUsePromptTokenCount > 0 {
		result.Extras["gemini.tool_use_prompt_token_count"] = usage.ToolUsePromptTokenCount
	}
	if usage.ThoughtsTokenCount > 0 {
		result.Extras["gemini.thoughts_token_count"] = usage.ThoughtsTokenCount
	}
	if len(usage.PromptTokensDetails) > 0 {
		result.Extras["gemini.prompt_tokens_details"] = usage.PromptTokensDetails
	}
	if len(usage.CacheTokensDetails) > 0 {
		result.Extras["gemini.cache_tokens_details"] = usage.CacheTokensDetails
	}
	if len(usage.CandidatesTokensDetails) > 0 {
		result.Extras["gemini.candidates_tokens_details"] = usage.CandidatesTokensDetails
	}
	if len(usage.ToolUsePromptTokensDetails) > 0 {
		result.Extras["gemini.tool_use_prompt_tokens_details"] = usage.ToolUsePromptTokensDetails
	}

	return result
}

// convertCandidateToChoice 转换候选响应
func convertCandidateToChoice(candidate *geminiTypes.Candidate) (*adapterTypes.ResponseChoice, error) {
	choice := &adapterTypes.ResponseChoice{
		Extras: make(map[string]interface{}),
	}

	// 映射 Index
	index := int(candidate.Index)
	choice.Index = &index

	// 映射 FinishReason
	if candidate.FinishReason != "" {
		finishReason := mapGeminiFinishReason(candidate.FinishReason)
		choice.FinishReason = &finishReason
		choice.NativeFinishReason = &candidate.FinishReason
	}

	// 保存 FinishMessage 到 Extras
	if candidate.FinishMessage != "" {
		choice.Extras["gemini.finish_message"] = candidate.FinishMessage
	}

	// 保存 SafetyRatings 到 Extras
	if len(candidate.SafetyRatings) > 0 {
		choice.Extras["gemini.safety_ratings"] = candidate.SafetyRatings
	}

	// 保存 CitationMetadata 到 Extras
	if candidate.CitationMetadata != nil {
		choice.Extras["gemini.citation_metadata"] = candidate.CitationMetadata
	}

	// 保存 TokenCount 到 Extras
	if candidate.TokenCount > 0 {
		choice.Extras["gemini.token_count"] = candidate.TokenCount
	}

	// 保存 GroundingAttributions 到 Extras
	if len(candidate.GroundingAttributions) > 0 {
		choice.Extras["gemini.grounding_attributions"] = candidate.GroundingAttributions
	}

	// 保存 GroundingMetadata 到 Extras
	if candidate.GroundingMetadata != nil {
		choice.Extras["gemini.grounding_metadata"] = candidate.GroundingMetadata
	}

	// 保存 AvgLogprobs 到 Extras
	if candidate.AvgLogprobs != nil {
		choice.Extras["gemini.avg_logprobs"] = *candidate.AvgLogprobs
	}

	// 保存 LogprobsResult 到 Extras
	if candidate.LogprobsResult != nil {
		choice.Extras["gemini.logprobs_result"] = candidate.LogprobsResult
	}

	// 保存 URLContextMetadata 到 Extras
	if candidate.URLContextMetadata != nil {
		choice.Extras["gemini.url_context_metadata"] = candidate.URLContextMetadata
	}

	// 转换 Content 为 Message
	message, err := convertContentToMessage(&candidate.Content, candidate.CitationMetadata)
	if err != nil {
		return nil, fmt.Errorf("转换内容失败：%w", err)
	}
	choice.Message = message

	return choice, nil
}

// convertContentToMessage 转换 Content 为 ResponseMessage
func convertContentToMessage(content *geminiTypes.Content, citationMeta *geminiTypes.CitationMetadata) (*adapterTypes.ResponseMessage, error) {
	message := &adapterTypes.ResponseMessage{
		Extras: make(map[string]interface{}),
	}

	// 映射 Role
	if content.Role != "" {
		message.Role = &content.Role
	}

	// 保存原始 Parts 到 Extras 以便反向转换
	if len(content.Parts) > 0 {
		message.Extras["gemini.parts"] = content.Parts
	}

	// 处理 Parts
	var textParts []string
	for _, part := range content.Parts {
		// 处理文本
		if part.Text != nil {
			textPart := adapterTypes.ResponseContentPart{
				Type:   "text",
				Text:   part.Text,
				Extras: make(map[string]interface{}),
			}

			// 处理 Citations
			if citationMeta != nil && len(citationMeta.CitationSources) > 0 {
				textPart.Annotations = convertCitationsToAnnotations(citationMeta.CitationSources)
			}

			// 保存 Part 特有字段到 Extras
			if part.Thought != nil {
				textPart.Extras["gemini.thought"] = *part.Thought
			}
			if part.ThoughtSignature != nil {
				textPart.Extras["gemini.thought_signature"] = *part.ThoughtSignature
			}
			if part.PartMetadata != nil {
				textPart.Extras["gemini.part_metadata"] = part.PartMetadata
			}
			if part.MediaResolution != nil {
				textPart.Extras["gemini.media_resolution"] = part.MediaResolution
			}

			message.Parts = append(message.Parts, textPart)
			textParts = append(textParts, *part.Text)
		}

		// 处理 FunctionCall
		if part.FunctionCall != nil {
			toolCall := convertFunctionCallToToolCall(part.FunctionCall)
			message.ToolCalls = append(message.ToolCalls, *toolCall)
		}

		// 处理 FunctionResponse
		if part.FunctionResponse != nil {
			toolResult := convertFunctionResponseToToolResult(part.FunctionResponse)
			message.ToolResults = append(message.ToolResults, *toolResult)
		}
	}

	// 聚合文本内容
	if len(textParts) > 0 {
		aggregated := ""
		for i, text := range textParts {
			if i > 0 {
				aggregated += "\n"
			}
			aggregated += text
		}
		message.Content = &aggregated
	}

	return message, nil
}

// convertCitationsToAnnotations 转换引用为注释
func convertCitationsToAnnotations(sources []geminiTypes.CitationSource) []adapterTypes.ResponseAnnotation {
	annotations := make([]adapterTypes.ResponseAnnotation, 0, len(sources))
	for _, source := range sources {
		annotation := adapterTypes.ResponseAnnotation{
			Type:   "url_citation",
			Extras: make(map[string]interface{}),
		}

		if source.StartIndex > 0 {
			startIndex := int(source.StartIndex)
			annotation.StartIndex = &startIndex
		}
		if source.EndIndex > 0 {
			endIndex := int(source.EndIndex)
			annotation.EndIndex = &endIndex
		}
		if source.URI != "" {
			annotation.URL = &source.URI
		}

		// License 不映射到 Title，放入 Extras
		if source.License != "" {
			annotation.Extras["gemini.license"] = source.License
		}

		// 标注索引单位为字节
		annotation.Extras["gemini.index_unit"] = "byte"

		annotations = append(annotations, annotation)
	}
	return annotations
}

// convertFunctionCallToToolCall 转换函数调用为工具调用
func convertFunctionCallToToolCall(fc *geminiTypes.FunctionCall) *adapterTypes.ResponseToolCall {
	toolCall := &adapterTypes.ResponseToolCall{
		Name:    &fc.Name,
		Payload: fc.Args,
		Extras:  make(map[string]interface{}),
	}

	if fc.ID != nil {
		toolCall.ID = fc.ID
	}

	// Gemini 没有 type 字段，默认为 function
	funcType := "function"
	toolCall.Type = &funcType

	return toolCall
}

// convertFunctionResponseToToolResult 转换函数响应为工具结果
func convertFunctionResponseToToolResult(fr *geminiTypes.FunctionResponse) *adapterTypes.ResponseToolResult {
	toolResult := &adapterTypes.ResponseToolResult{
		Name:    &fr.Name,
		Payload: fr.Response,
		Extras:  make(map[string]interface{}),
	}

	if fr.ID != nil {
		toolResult.ID = fr.ID
	}

	// 尝试将 Response 转换为可读文本
	if fr.Response != nil {
		if jsonBytes, err := json.Marshal(fr.Response); err == nil {
			content := string(jsonBytes)
			toolResult.Content = &content
		}
	}

	// 保存其他字段到 Extras
	if len(fr.Parts) > 0 {
		toolResult.Extras["gemini.parts"] = fr.Parts
	}
	if fr.WillContinue != nil {
		toolResult.Extras["gemini.will_continue"] = *fr.WillContinue
	}
	if fr.Scheduling != nil {
		toolResult.Extras["gemini.scheduling"] = *fr.Scheduling
	}

	return toolResult
}

// mapGeminiFinishReason 映射 Gemini FinishReason 到统一格式
func mapGeminiFinishReason(reason geminiTypes.FinishReason) adapterTypes.ResponseFinishReason {
	switch reason {
	case geminiTypes.FinishReasonStop:
		return adapterTypes.ResponseFinishReasonStop
	case geminiTypes.FinishReasonMaxTokens:
		return adapterTypes.ResponseFinishReasonLength
	case geminiTypes.FinishReasonSafety,
		geminiTypes.FinishReasonBlocklist,
		geminiTypes.FinishReasonProhibitedContent,
		geminiTypes.FinishReasonImageProhibited,
		geminiTypes.FinishReasonImageSafety:
		return adapterTypes.ResponseFinishReasonContentFilter
	case geminiTypes.FinishReasonRecitation,
		geminiTypes.FinishReasonImageRecitation:
		return adapterTypes.ResponseFinishReasonRecitation
	case geminiTypes.FinishReasonLanguage:
		return adapterTypes.ResponseFinishReasonLanguage
	case geminiTypes.FinishReasonMalformedFunction:
		return adapterTypes.ResponseFinishReasonToolCallMalformed
	case geminiTypes.FinishReasonUnexpectedToolCall:
		return adapterTypes.ResponseFinishReasonToolCallUnexpected
	case geminiTypes.FinishReasonTooManyToolCalls:
		return adapterTypes.ResponseFinishReasonToolCallLimit
	case geminiTypes.FinishReasonMissingThoughtSig:
		return adapterTypes.ResponseFinishReasonThoughtSignatureMissing
	default:
		return adapterTypes.ResponseFinishReasonUnknown
	}
}
