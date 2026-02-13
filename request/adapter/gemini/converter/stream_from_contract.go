package converter

import (
	"encoding/json"
	"fmt"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
)

// StreamEventFormContract 将单个 StreamEventContract 转换为 Gemini 流式响应块。
//
// 根据 plans/gemini-stream-from-contract-plan.md 实现：
//   - message_delta -> 中间块（增量内容）
//   - message_stop -> 结束块（带 FinishReason/FinishMessage）
//   - error -> ErrorResponse（不产生 StreamEvent，通过 error 返回）
//   - 其他类型 -> 返回 nil, nil
func StreamEventFromContract(contract *adapterTypes.StreamEventContract) (*geminiTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	log := logger.Default().WithGroup("stream_converter")

	// 验证事件来源
	if contract.Source != adapterTypes.StreamSourceGemini {
		return nil, errors.New(errors.ErrCodeInvalidArgument,
			fmt.Sprintf("事件来源不匹配，期望 %s，实际 %s", adapterTypes.StreamSourceGemini, contract.Source))
	}

	// 按事件类型处理
	switch contract.Type {
	case adapterTypes.StreamEventMessageDelta, adapterTypes.StreamEventMessageStop:
		event, err := convertMessageEventToGeminiResponse(contract, log)
		if err != nil {
			log.Error("转换消息事件失败", "error", err, "event_type", contract.Type)
			return nil, errors.Wrap(errors.ErrCodeStreamError, "转换消息事件失败", err)
		}
		log.Debug("转换 StreamEventContract 到 Gemini 流式响应完成", "event_type", contract.Type)
		return event, nil
	case adapterTypes.StreamEventError:
		errorResp := convertErrorEventToGeminiError(contract)
		if errorResp != nil {
			// 错误响应不作为 StreamEvent 返回，而是通过 error 返回
			return nil, errors.New(errors.ErrCodeRequestFailed, errorResp.Error.Message).
				WithContext("error_code", errorResp.Error.Code).
				WithContext("error_status", errorResp.Error.Status)
		}
		return nil, nil
	default:
		log.Warn("忽略不支持的事件类型", "event_type", contract.Type)
		return nil, nil
	}
}

// convertMessageEventToGeminiResponse 将消息事件转换为 Gemini Response。
//
// 根据 plans/gemini-stream-from-contract-plan.md 实现：
//   - usageMetadata：在出现该字段的事件块中输出（delta/stop 均可）
//   - promptFeedback/modelVersion/modelStatus：从 extensions.gemini 提取
//   - candidate 扩展字段（safety/grounding/logprobs 等）：从 extensions.gemini 提取
func convertMessageEventToGeminiResponse(contract *adapterTypes.StreamEventContract, log logger.Logger) (*geminiTypes.StreamEvent, error) {
	if contract == nil {
		return nil, nil
	}

	response := &geminiTypes.StreamEvent{
		ResponseID: contract.ResponseID,
	}

	// 从 extensions 提取响应级字段（promptFeedback/modelVersion/modelStatus）
	extractResponseExtensions(response, contract.Extensions)

	// 构建 UsageMetadata（在出现该字段的事件块中输出）
	if contract.Usage != nil {
		response.UsageMetadata = convertStreamUsageToGeminiUsage(contract.Usage)
	}

	// 构建 Candidate（包含扩展字段：safety/grounding/logprobs 等）
	candidate, err := buildGeminiCandidate(contract, log)
	if err != nil {
		return nil, err
	}

	if candidate != nil {
		response.Candidates = []geminiTypes.Candidate{*candidate}
	}

	return response, nil
}

// buildGeminiCandidate 从 StreamEventContract 构建 Gemini Candidate。
//
// 根据 plans/gemini-stream-from-contract-plan.md 实现：
//   - message_delta 事件：构建包含增量 content/parts 的候选
//   - message_stop 事件：构建包含 finishReason/finishMessage 的候选，content 可以为空
//   - 增量原则：parts 不做拼接，保持原样输出
func buildGeminiCandidate(contract *adapterTypes.StreamEventContract, log logger.Logger) (*geminiTypes.Candidate, error) {
	if contract == nil {
		return nil, nil
	}

	candidate := &geminiTypes.Candidate{
		Index: int32(contract.OutputIndex),
	}

	// 从 content.raw 提取 finish_reason（message_stop 事件）
	if contract.Content != nil && contract.Content.Raw != nil {
		if finishReason, ok := contract.Content.Raw["finish_reason"].(string); ok {
			candidate.FinishReason = geminiTypes.FinishReason(finishReason)
		}
	}

	// 从 extensions 提取候选级字段（包括 finish_message）
	extractCandidateExtensions(candidate, contract.Extensions)

	// 构建 Content（message_delta 事件）
	// 注意：message_stop 事件的 content 可以为空，这是允许的
	if contract.Message != nil {
		content, err := buildGeminiContent(contract.Message, log)
		if err != nil {
			return nil, err
		}
		if content != nil {
			candidate.Content = *content
		}
	}

	return candidate, nil
}

// buildGeminiContent 从 StreamMessagePayload 构建 Gemini Content。
//
// 根据 plans/gemini-stream-from-contract-plan.md 实现：
//   - 增量原则：parts 不做拼接，保持原样输出
//   - 空 content 允许：message_stop 事件的 content 可以为空
//   - 优先级：message.parts > message.tool_calls > message.content_text
func buildGeminiContent(message *adapterTypes.StreamMessagePayload, log logger.Logger) (*geminiTypes.Content, error) {
	if message == nil {
		return nil, nil
	}

	content := &geminiTypes.Content{
		Role: message.Role,
	}

	// 转换 Parts（增量原则：不做拼接，保持原样）
	parts := make([]geminiTypes.Part, 0)

	// 从 message.parts 转换
	if len(message.Parts) > 0 {
		geminiParts, err := convertStreamPartsToGeminiParts(message.Parts, log)
		if err != nil {
			return nil, err
		}
		parts = append(parts, geminiParts...)
	}

	// 从 message.tool_calls 转换
	if len(message.ToolCalls) > 0 {
		toolParts, err := convertStreamToolCallsToGeminiParts(message.ToolCalls, log)
		if err != nil {
			return nil, err
		}
		parts = append(parts, toolParts...)
	}

	// 从 message.content_text 转换
	if message.ContentText != nil && *message.ContentText != "" {
		parts = append(parts, geminiTypes.Part{
			Text: message.ContentText,
		})
	}

	// 空 content 允许：若 parts 为空，仍返回 content（仅含 role）
	if len(parts) > 0 {
		content.Parts = parts
	}

	return content, nil
}

// convertStreamPartsToGeminiParts 将 StreamContentPart 转换为 Gemini Part。
func convertStreamPartsToGeminiParts(parts []adapterTypes.StreamContentPart, log logger.Logger) ([]geminiTypes.Part, error) {
	if len(parts) == 0 {
		return nil, nil
	}

	geminiParts := make([]geminiTypes.Part, 0, len(parts))
	for _, part := range parts {
		geminiPart := geminiTypes.Part{}

		// 从 raw 提取特有字段
		if part.Raw != nil {
			extractPartRawFields(&geminiPart, part.Raw)
		}

		// 根据 type 设置主要字段
		switch part.Type {
		case "text":
			if part.Text != "" {
				geminiPart.Text = &part.Text
			}
		case "tool_use":
			// tool_use 类型由 tool_calls 处理，这里只处理 raw 中的信息
			if geminiPart.FunctionCall == nil && part.Raw != nil {
				if fc, ok := part.Raw["function_call"].(map[string]interface{}); ok {
					geminiPart.FunctionCall = convertMapToFunctionCall(fc)
				}
			}
		case "tool_result":
			if part.Raw != nil {
				if fr, ok := part.Raw["function_response"].(map[string]interface{}); ok {
					geminiPart.FunctionResponse = convertMapToFunctionResponse(fr)
				}
			}
		case "image", "audio", "video", "file":
			// 从 raw 提取 inline_data 或 file_data
			if part.Raw != nil {
				if id, ok := part.Raw["inline_data"].(map[string]interface{}); ok {
					geminiPart.InlineData = convertMapToInlineData(id)
				} else if fd, ok := part.Raw["file_data"].(map[string]interface{}); ok {
					geminiPart.FileData = convertMapToFileData(fd)
				}
			}
		case "executable_code":
			if part.Raw != nil {
				if ec, ok := part.Raw["executable_code"].(map[string]interface{}); ok {
					geminiPart.ExecutableCode = convertMapToExecutableCode(ec)
				}
			}
		case "code_execution_result":
			if part.Raw != nil {
				if cer, ok := part.Raw["code_execution_result"].(map[string]interface{}); ok {
					geminiPart.CodeExecutionResult = convertMapToCodeExecutionResult(cer)
				}
			}
		default:
			log.Warn("未知的 Part 类型", "type", part.Type)
		}

		geminiParts = append(geminiParts, geminiPart)
	}

	return geminiParts, nil
}

// convertStreamToolCallsToGeminiParts 将 StreamToolCall 转换为 Gemini FunctionCall Parts。
func convertStreamToolCallsToGeminiParts(toolCalls []adapterTypes.StreamToolCall, log logger.Logger) ([]geminiTypes.Part, error) {
	if len(toolCalls) == 0 {
		return nil, nil
	}

	parts := make([]geminiTypes.Part, 0, len(toolCalls))
	for _, tc := range toolCalls {
		part := geminiTypes.Part{
			FunctionCall: &geminiTypes.FunctionCall{
				Name: tc.Name,
			},
		}

		if tc.ID != "" {
			part.FunctionCall.ID = &tc.ID
		}

		if tc.Arguments != "" {
			args := make(map[string]interface{})
			if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
				log.Warn("反序列化工具调用参数失败", "error", err, "tool_name", tc.Name)
			} else {
				part.FunctionCall.Args = args
			}
		}

		// 从 raw 提取额外字段
		if tc.Raw != nil {
			if part.FunctionCall == nil {
				part.FunctionCall = &geminiTypes.FunctionCall{}
			}
			extractFunctionCallRawFields(part.FunctionCall, tc.Raw)
		}

		parts = append(parts, part)
	}

	return parts, nil
}

// convertStreamUsageToGeminiUsage 将 StreamUsagePayload 转换为 Gemini UsageMetadata。
func convertStreamUsageToGeminiUsage(usage *adapterTypes.StreamUsagePayload) *geminiTypes.UsageMetadata {
	if usage == nil {
		return nil
	}

	result := &geminiTypes.UsageMetadata{}

	if usage.InputTokens != nil {
		result.PromptTokenCount = int32(*usage.InputTokens)
	}
	if usage.OutputTokens != nil {
		result.CandidatesTokenCount = int32(*usage.OutputTokens)
	}
	if usage.TotalTokens != nil {
		result.TotalTokenCount = int32(*usage.TotalTokens)
	}

	// 从 raw 提取扩展字段
	if usage.Raw != nil {
		if val, ok := usage.Raw["cached_content_token_count"].(float64); ok {
			result.CachedContentTokenCount = int32(val)
		}
		if val, ok := usage.Raw["tool_use_prompt_token_count"].(float64); ok {
			result.ToolUsePromptTokenCount = int32(val)
		}
		if val, ok := usage.Raw["thoughts_token_count"].(float64); ok {
			result.ThoughtsTokenCount = int32(val)
		}
		if val, ok := usage.Raw["prompt_tokens_details"].([]interface{}); ok {
			result.PromptTokensDetails = convertModalityTokenCounts(val)
		}
		if val, ok := usage.Raw["cache_tokens_details"].([]interface{}); ok {
			result.CacheTokensDetails = convertModalityTokenCounts(val)
		}
		if val, ok := usage.Raw["candidates_tokens_details"].([]interface{}); ok {
			result.CandidatesTokensDetails = convertModalityTokenCounts(val)
		}
		if val, ok := usage.Raw["tool_use_prompt_tokens_details"].([]interface{}); ok {
			result.ToolUsePromptTokensDetails = convertModalityTokenCounts(val)
		}
	}

	return result
}

// convertErrorEventToGeminiError 将错误事件转换为 Gemini ErrorResponse。
func convertErrorEventToGeminiError(contract *adapterTypes.StreamEventContract) *geminiTypes.ErrorResponse {
	if contract == nil || contract.Error == nil {
		return nil
	}

	errorDetail := geminiTypes.ErrorDetail{
		Message: contract.Error.Message,
	}

	if contract.Error.Code != "" {
		// 尝试将 code 转换为 int
		var code int
		if _, err := fmt.Sscanf(contract.Error.Code, "%d", &code); err == nil {
			errorDetail.Code = code
		} else {
			// 如果不是数字，使用默认错误码
			errorDetail.Code = 500
		}
	}

	if contract.Error.Type != "" {
		errorDetail.Status = contract.Error.Type
	}

	if contract.Error.Param != "" {
		errorDetail.Details = []map[string]interface{}{
			{"param": contract.Error.Param},
		}
	}

	// 从 raw 提取额外字段
	if contract.Error.Raw != nil {
		if errorDetail.Details == nil {
			errorDetail.Details = []map[string]interface{}{}
		}
		for k, v := range contract.Error.Raw {
			errorDetail.Details = append(errorDetail.Details, map[string]interface{}{k: v})
		}
	}

	return &geminiTypes.ErrorResponse{
		Error: errorDetail,
	}
}

// extractResponseExtensions 从 extensions 提取响应级字段。
func extractResponseExtensions(response *geminiTypes.StreamEvent, extensions map[string]interface{}) {
	if extensions == nil {
		return
	}

	// 从 gemini 命名空间提取
	if geminiExt, ok := extensions["gemini"].(map[string]interface{}); ok {
		if val, ok := geminiExt["prompt_feedback"].(map[string]interface{}); ok {
			response.PromptFeedback = convertMapToPromptFeedback(val)
		}
		if val, ok := geminiExt["model_version"].(string); ok {
			response.ModelVersion = val
		}
		if val, ok := geminiExt["model_status"].(map[string]interface{}); ok {
			response.ModelStatus = convertMapToModelStatus(val)
		}
	}
}

// extractCandidateExtensions 从 extensions 提取候选级字段。
func extractCandidateExtensions(candidate *geminiTypes.Candidate, extensions map[string]interface{}) {
	if extensions == nil {
		return
	}

	// 从 gemini 命名空间提取
	if geminiExt, ok := extensions["gemini"].(map[string]interface{}); ok {
		if val, ok := geminiExt["finish_message"].(string); ok {
			candidate.FinishMessage = val
		}
		if val, ok := geminiExt["safety_ratings"].([]interface{}); ok {
			candidate.SafetyRatings = convertSafetyRatings(val)
		}
		if val, ok := geminiExt["citation_metadata"].(map[string]interface{}); ok {
			candidate.CitationMetadata = convertMapToCitationMetadata(val)
		}
		if val, ok := geminiExt["grounding_attributions"].([]interface{}); ok {
			candidate.GroundingAttributions = convertGroundingAttributions(val)
		}
		if val, ok := geminiExt["grounding_metadata"].(map[string]interface{}); ok {
			candidate.GroundingMetadata = convertMapToGroundingMetadata(val)
		}
		if val, ok := geminiExt["avg_logprobs"].(float64); ok {
			candidate.AvgLogprobs = &val
		}
		if val, ok := geminiExt["logprobs_result"].(map[string]interface{}); ok {
			candidate.LogprobsResult = convertMapToLogprobsResult(val)
		}
		if val, ok := geminiExt["url_context_metadata"].(map[string]interface{}); ok {
			candidate.URLContextMetadata = convertMapToURLContextMetadata(val)
		}
		if val, ok := geminiExt["token_count"].(float64); ok {
			candidate.TokenCount = int32(val)
		}
	}
}

// extractPartRawFields 从 raw 提取 Part 特有字段。
func extractPartRawFields(part *geminiTypes.Part, raw map[string]interface{}) {
	if val, ok := raw["thought"].(bool); ok {
		part.Thought = &val
	}
	if val, ok := raw["thought_signature"].(string); ok {
		part.ThoughtSignature = &val
	}
	if val, ok := raw["part_metadata"].(map[string]interface{}); ok {
		part.PartMetadata = val
	}
	if val, ok := raw["media_resolution"].(map[string]interface{}); ok {
		part.MediaResolution = convertMapToMediaResolution(val)
	}
}

// extractFunctionCallRawFields 从 raw 提取 FunctionCall 特有字段。
func extractFunctionCallRawFields(fc *geminiTypes.FunctionCall, raw map[string]interface{}) {
	// FunctionCall 的主要字段已经通过 StreamToolCall 设置
	// 这里处理可能的额外字段
}

// convertMapToFunctionCall 将 map 转换为 FunctionCall。
func convertMapToFunctionCall(m map[string]interface{}) *geminiTypes.FunctionCall {
	if m == nil {
		return nil
	}

	fc := &geminiTypes.FunctionCall{
		Args: make(map[string]interface{}),
	}

	if val, ok := m["id"].(string); ok {
		fc.ID = &val
	}
	if val, ok := m["name"].(string); ok {
		fc.Name = val
	}
	if val, ok := m["args"].(map[string]interface{}); ok {
		fc.Args = val
	}

	return fc
}

// convertMapToFunctionResponse 将 map 转换为 FunctionResponse。
func convertMapToFunctionResponse(m map[string]interface{}) *geminiTypes.FunctionResponse {
	if m == nil {
		return nil
	}

	fr := &geminiTypes.FunctionResponse{
		Response: make(map[string]interface{}),
	}

	if val, ok := m["id"].(string); ok {
		fr.ID = &val
	}
	if val, ok := m["name"].(string); ok {
		fr.Name = val
	}
	if val, ok := m["response"].(map[string]interface{}); ok {
		fr.Response = val
	}
	if val, ok := m["parts"].([]interface{}); ok {
		fr.Parts = convertFunctionResponseParts(val)
	}
	if val, ok := m["will_continue"].(bool); ok {
		fr.WillContinue = &val
	}
	if val, ok := m["scheduling"].(string); ok {
		fr.Scheduling = &val
	}

	return fr
}

// convertMapToInlineData 将 map 转换为 InlineData。
func convertMapToInlineData(m map[string]interface{}) *geminiTypes.InlineData {
	if m == nil {
		return nil
	}

	id := &geminiTypes.InlineData{}
	if val, ok := m["mimeType"].(string); ok {
		id.MimeType = val
	}
	if val, ok := m["data"].(string); ok {
		id.Data = val
	}

	return id
}

// convertMapToFileData 将 map 转换为 FileData。
func convertMapToFileData(m map[string]interface{}) *geminiTypes.FileData {
	if m == nil {
		return nil
	}

	fd := &geminiTypes.FileData{}
	if val, ok := m["fileUri"].(string); ok {
		fd.FileURI = val
	}
	if val, ok := m["mimeType"].(string); ok {
		fd.MimeType = &val
	}

	return fd
}

// convertMapToExecutableCode 将 map 转换为 ExecutableCode。
func convertMapToExecutableCode(m map[string]interface{}) *geminiTypes.ExecutableCode {
	if m == nil {
		return nil
	}

	ec := &geminiTypes.ExecutableCode{}
	if val, ok := m["language"].(string); ok {
		ec.Language = val
	}
	if val, ok := m["code"].(string); ok {
		ec.Code = val
	}

	return ec
}

// convertMapToCodeExecutionResult 将 map 转换为 CodeExecutionResult。
func convertMapToCodeExecutionResult(m map[string]interface{}) *geminiTypes.CodeExecutionResult {
	if m == nil {
		return nil
	}

	cer := &geminiTypes.CodeExecutionResult{}
	if val, ok := m["outcome"].(string); ok {
		cer.Outcome = val
	}
	if val, ok := m["output"].(string); ok {
		cer.Output = val
	}

	return cer
}

// convertModalityTokenCounts 将 interface{} 列表转换为 ModalityTokenCount 列表。
func convertModalityTokenCounts(items []interface{}) []geminiTypes.ModalityTokenCount {
	if len(items) == 0 {
		return nil
	}

	result := make([]geminiTypes.ModalityTokenCount, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			mtc := geminiTypes.ModalityTokenCount{}
			if val, ok := m["modality"].(string); ok {
				mtc.Modality = geminiTypes.Modality(val)
			}
			if val, ok := m["tokenCount"].(float64); ok {
				mtc.TokenCount = int32(val)
			}
			result = append(result, mtc)
		}
	}

	return result
}

// convertMapToPromptFeedback 将 map 转换为 PromptFeedback。
func convertMapToPromptFeedback(m map[string]interface{}) *geminiTypes.PromptFeedback {
	if m == nil {
		return nil
	}

	pf := &geminiTypes.PromptFeedback{}
	if val, ok := m["blockReason"].(string); ok {
		pf.BlockReason = geminiTypes.BlockReason(val)
	}
	if val, ok := m["safetyRatings"].([]interface{}); ok {
		pf.SafetyRatings = convertSafetyRatings(val)
	}

	return pf
}

// convertMapToModelStatus 将 map 转换为 ModelStatus。
func convertMapToModelStatus(m map[string]interface{}) *geminiTypes.ModelStatus {
	if m == nil {
		return nil
	}

	ms := &geminiTypes.ModelStatus{}
	if val, ok := m["modelStage"].(string); ok {
		ms.ModelStage = geminiTypes.ModelStage(val)
	}
	if val, ok := m["retirementTime"].(string); ok {
		ms.RetirementTime = val
	}
	if val, ok := m["message"].(string); ok {
		ms.Message = val
	}

	return ms
}

// convertSafetyRatings 将 interface{} 列表转换为 SafetyRating 列表。
func convertSafetyRatings(items []interface{}) []geminiTypes.SafetyRating {
	if len(items) == 0 {
		return nil
	}

	result := make([]geminiTypes.SafetyRating, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			sr := geminiTypes.SafetyRating{}
			if val, ok := m["category"].(string); ok {
				sr.Category = geminiTypes.HarmCategory(val)
			}
			if val, ok := m["probability"].(string); ok {
				sr.Probability = geminiTypes.HarmProbability(val)
			}
			if val, ok := m["blocked"].(bool); ok {
				sr.Blocked = val
			}
			result = append(result, sr)
		}
	}

	return result
}

// convertMapToCitationMetadata 将 map 转换为 CitationMetadata。
func convertMapToCitationMetadata(m map[string]interface{}) *geminiTypes.CitationMetadata {
	if m == nil {
		return nil
	}

	cm := &geminiTypes.CitationMetadata{}
	if val, ok := m["citationSources"].([]interface{}); ok {
		cm.CitationSources = convertCitationSources(val)
	}

	return cm
}

// convertCitationSources 将 interface{} 列表转换为 CitationSource 列表。
func convertCitationSources(items []interface{}) []geminiTypes.CitationSource {
	if len(items) == 0 {
		return nil
	}

	result := make([]geminiTypes.CitationSource, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			cs := geminiTypes.CitationSource{}
			if val, ok := m["startIndex"].(float64); ok {
				cs.StartIndex = int32(val)
			}
			if val, ok := m["endIndex"].(float64); ok {
				cs.EndIndex = int32(val)
			}
			if val, ok := m["uri"].(string); ok {
				cs.URI = val
			}
			if val, ok := m["license"].(string); ok {
				cs.License = val
			}
			result = append(result, cs)
		}
	}

	return result
}

// convertGroundingAttributions 将 interface{} 列表转换为 GroundingAttribution 列表。
func convertGroundingAttributions(items []interface{}) []geminiTypes.GroundingAttribution {
	if len(items) == 0 {
		return nil
	}

	result := make([]geminiTypes.GroundingAttribution, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			ga := geminiTypes.GroundingAttribution{}
			if val, ok := m["sourceId"].(map[string]interface{}); ok {
				ga.SourceID = convertMapToAttributionSourceID(val)
			}
			if val, ok := m["content"].(map[string]interface{}); ok {
				ga.Content = convertMapToContent(val)
			}
			result = append(result, ga)
		}
	}

	return result
}

// convertMapToAttributionSourceID 将 map 转换为 AttributionSourceID。
func convertMapToAttributionSourceID(m map[string]interface{}) *geminiTypes.AttributionSourceID {
	if m == nil {
		return nil
	}

	asid := &geminiTypes.AttributionSourceID{}
	if val, ok := m["groundingPassage"].(map[string]interface{}); ok {
		asid.GroundingPassage = convertMapToGroundingPassageID(val)
	}
	if val, ok := m["semanticRetrieverChunk"].(map[string]interface{}); ok {
		asid.SemanticRetrieverChunk = convertMapToSemanticRetrieverChunk(val)
	}

	return asid
}

// convertMapToGroundingPassageID 将 map 转换为 GroundingPassageID。
func convertMapToGroundingPassageID(m map[string]interface{}) *geminiTypes.GroundingPassageID {
	if m == nil {
		return nil
	}

	gpid := &geminiTypes.GroundingPassageID{}
	if val, ok := m["passageId"].(string); ok {
		gpid.PassageID = val
	}
	if val, ok := m["partIndex"].(float64); ok {
		gpid.PartIndex = int32(val)
	}

	return gpid
}

// convertMapToSemanticRetrieverChunk 将 map 转换为 SemanticRetrieverChunk。
func convertMapToSemanticRetrieverChunk(m map[string]interface{}) *geminiTypes.SemanticRetrieverChunk {
	if m == nil {
		return nil
	}

	src := &geminiTypes.SemanticRetrieverChunk{}
	if val, ok := m["source"].(string); ok {
		src.Source = val
	}
	if val, ok := m["chunk"].(string); ok {
		src.Chunk = val
	}

	return src
}

// convertMapToContent 将 map 转换为 Content。
func convertMapToContent(m map[string]interface{}) *geminiTypes.Content {
	if m == nil {
		return nil
	}

	content := &geminiTypes.Content{}
	if val, ok := m["role"].(string); ok {
		content.Role = val
	}
	// parts 字段需要递归转换，这里简化处理
	// 实际使用时应该根据具体场景实现完整的转换逻辑

	return content
}

// convertMapToGroundingMetadata 将 map 转换为 GroundingMetadata。
func convertMapToGroundingMetadata(m map[string]interface{}) *geminiTypes.GroundingMetadata {
	if m == nil {
		return nil
	}

	gm := &geminiTypes.GroundingMetadata{}
	if val, ok := m["searchEntryPoint"].(map[string]interface{}); ok {
		gm.SearchEntryPoint = convertMapToSearchEntryPoint(val)
	}
	if val, ok := m["groundingChunks"].([]interface{}); ok {
		gm.GroundingChunks = convertGroundingChunks(val)
	}
	if val, ok := m["groundingSupports"].([]interface{}); ok {
		gm.GroundingSupports = convertGroundingSupports(val)
	}
	if val, ok := m["retrievalMetadata"].(map[string]interface{}); ok {
		gm.RetrievalMetadata = convertMapToRetrievalMetadata(val)
	}
	if val, ok := m["webSearchQueries"].([]interface{}); ok {
		gm.WebSearchQueries = convertStringSlice(val)
	}
	if val, ok := m["googleMapsWidgetContextToken"].(string); ok {
		gm.GoogleMapsWidgetContextToken = val
	}

	return gm
}

// convertMapToSearchEntryPoint 将 map 转换为 SearchEntryPoint。
func convertMapToSearchEntryPoint(m map[string]interface{}) *geminiTypes.SearchEntryPoint {
	if m == nil {
		return nil
	}

	sep := &geminiTypes.SearchEntryPoint{}
	if val, ok := m["renderedContent"].(string); ok {
		sep.RenderedContent = val
	}
	if val, ok := m["sdkBlob"].(string); ok {
		sep.SDKBlob = val
	}

	return sep
}

// convertGroundingChunks 将 interface{} 列表转换为 GroundingChunk 列表。
func convertGroundingChunks(items []interface{}) []geminiTypes.GroundingChunk {
	if len(items) == 0 {
		return nil
	}

	result := make([]geminiTypes.GroundingChunk, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			gc := geminiTypes.GroundingChunk{}
			if val, ok := m["web"].(map[string]interface{}); ok {
				gc.Web = convertMapToWeb(val)
			}
			if val, ok := m["retrievedContext"].(map[string]interface{}); ok {
				gc.RetrievedContext = convertMapToRetrievedContext(val)
			}
			if val, ok := m["maps"].(map[string]interface{}); ok {
				gc.Maps = convertMapToMaps(val)
			}
			result = append(result, gc)
		}
	}

	return result
}

// convertMapToWeb 将 map 转换为 Web。
func convertMapToWeb(m map[string]interface{}) *geminiTypes.Web {
	if m == nil {
		return nil
	}

	web := &geminiTypes.Web{}
	if val, ok := m["uri"].(string); ok {
		web.URI = val
	}
	if val, ok := m["title"].(string); ok {
		web.Title = val
	}

	return web
}

// convertMapToRetrievedContext 将 map 转换为 RetrievedContext。
func convertMapToRetrievedContext(m map[string]interface{}) *geminiTypes.RetrievedContext {
	if m == nil {
		return nil
	}

	rc := &geminiTypes.RetrievedContext{}
	if val, ok := m["uri"].(string); ok {
		rc.URI = val
	}
	if val, ok := m["title"].(string); ok {
		rc.Title = val
	}
	if val, ok := m["text"].(string); ok {
		rc.Text = val
	}
	if val, ok := m["fileSearchStore"].(string); ok {
		rc.FileSearchStore = val
	}

	return rc
}

// convertMapToMaps 将 map 转换为 Maps。
func convertMapToMaps(m map[string]interface{}) *geminiTypes.Maps {
	if m == nil {
		return nil
	}

	maps := &geminiTypes.Maps{}
	if val, ok := m["uri"].(string); ok {
		maps.URI = val
	}
	if val, ok := m["title"].(string); ok {
		maps.Title = val
	}
	if val, ok := m["text"].(string); ok {
		maps.Text = val
	}
	if val, ok := m["placeId"].(string); ok {
		maps.PlaceID = val
	}
	if val, ok := m["placeAnswerSources"].(map[string]interface{}); ok {
		maps.PlaceAnswerSources = convertMapToPlaceAnswerSources(val)
	}

	return maps
}

// convertMapToPlaceAnswerSources 将 map 转换为 PlaceAnswerSources。
func convertMapToPlaceAnswerSources(m map[string]interface{}) *geminiTypes.PlaceAnswerSources {
	if m == nil {
		return nil
	}

	pas := &geminiTypes.PlaceAnswerSources{}
	if val, ok := m["reviewSnippets"].([]interface{}); ok {
		pas.ReviewSnippets = convertReviewSnippets(val)
	}

	return pas
}

// convertReviewSnippets 将 interface{} 列表转换为 ReviewSnippet 列表。
func convertReviewSnippets(items []interface{}) []geminiTypes.ReviewSnippet {
	if len(items) == 0 {
		return nil
	}

	result := make([]geminiTypes.ReviewSnippet, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			rs := geminiTypes.ReviewSnippet{}
			if val, ok := m["reviewId"].(string); ok {
				rs.ReviewID = val
			}
			if val, ok := m["googleMapsUri"].(string); ok {
				rs.GoogleMapsURI = val
			}
			if val, ok := m["title"].(string); ok {
				rs.Title = val
			}
			result = append(result, rs)
		}
	}

	return result
}

// convertGroundingSupports 将 interface{} 列表转换为 GroundingSupport 列表。
func convertGroundingSupports(items []interface{}) []geminiTypes.GroundingSupport {
	if len(items) == 0 {
		return nil
	}

	result := make([]geminiTypes.GroundingSupport, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			gs := geminiTypes.GroundingSupport{}
			if val, ok := m["segment"].(map[string]interface{}); ok {
				gs.Segment = convertMapToSegment(val)
			}
			if val, ok := m["groundingChunkIndices"].([]interface{}); ok {
				gs.GroundingChunkIndices = convertInt32Slice(val)
			}
			if val, ok := m["confidenceScores"].([]interface{}); ok {
				gs.ConfidenceScores = convertFloat32Slice(val)
			}
			result = append(result, gs)
		}
	}

	return result
}

// convertMapToSegment 将 map 转换为 Segment。
func convertMapToSegment(m map[string]interface{}) *geminiTypes.Segment {
	if m == nil {
		return nil
	}

	segment := &geminiTypes.Segment{}
	if val, ok := m["partIndex"].(float64); ok {
		segment.PartIndex = int32(val)
	}
	if val, ok := m["startIndex"].(float64); ok {
		segment.StartIndex = int32(val)
	}
	if val, ok := m["endIndex"].(float64); ok {
		segment.EndIndex = int32(val)
	}
	if val, ok := m["text"].(string); ok {
		segment.Text = val
	}

	return segment
}

// convertMapToRetrievalMetadata 将 map 转换为 RetrievalMetadata。
func convertMapToRetrievalMetadata(m map[string]interface{}) *geminiTypes.RetrievalMetadata {
	if m == nil {
		return nil
	}

	rm := &geminiTypes.RetrievalMetadata{}
	if val, ok := m["googleSearchDynamicRetrievalScore"].(float64); ok {
		rm.GoogleSearchDynamicRetrievalScore = float32(val)
	}

	return rm
}

// convertMapToLogprobsResult 将 map 转换为 LogprobsResult。
func convertMapToLogprobsResult(m map[string]interface{}) *geminiTypes.LogprobsResult {
	if m == nil {
		return nil
	}

	lr := &geminiTypes.LogprobsResult{}
	if val, ok := m["logProbabilitySum"].(float64); ok {
		lr.LogProbabilitySum = float32(val)
	}
	if val, ok := m["topCandidates"].([]interface{}); ok {
		// TopCandidates 是一个切片，不是单个对象
		for _, item := range val {
			if m, ok := item.(map[string]interface{}); ok {
				tc := convertMapToTopCandidates(m)
				if tc != nil {
					lr.TopCandidates = append(lr.TopCandidates, *tc)
				}
			}
		}
	}
	if val, ok := m["chosenCandidates"].([]interface{}); ok {
		lr.ChosenCandidates = convertLogprobsResultCandidates(val)
	}

	return lr
}

// convertMapToTopCandidates 将 map 转换为 TopCandidates。
func convertMapToTopCandidates(m map[string]interface{}) *geminiTypes.TopCandidates {
	if m == nil {
		return nil
	}

	tc := &geminiTypes.TopCandidates{}
	if val, ok := m["candidates"].([]interface{}); ok {
		tc.Candidates = convertLogprobsResultCandidates(val)
	}

	return tc
}

// convertLogprobsResultCandidates 将 interface{} 列表转换为 LogprobsResultCandidate 列表。
func convertLogprobsResultCandidates(items []interface{}) []geminiTypes.LogprobsResultCandidate {
	if len(items) == 0 {
		return nil
	}

	result := make([]geminiTypes.LogprobsResultCandidate, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			lrc := geminiTypes.LogprobsResultCandidate{}
			if val, ok := m["token"].(string); ok {
				lrc.Token = val
			}
			if val, ok := m["tokenId"].(float64); ok {
				lrc.TokenID = int32(val)
			}
			if val, ok := m["logProbability"].(float64); ok {
				lrc.LogProbability = float32(val)
			}
			result = append(result, lrc)
		}
	}

	return result
}

// convertMapToURLContextMetadata 将 map 转换为 URLContextMetadata。
func convertMapToURLContextMetadata(m map[string]interface{}) *geminiTypes.URLContextMetadata {
	if m == nil {
		return nil
	}

	ucm := &geminiTypes.URLContextMetadata{}
	if val, ok := m["urlMetadata"].([]interface{}); ok {
		ucm.URLMetadata = convertURLMetadataList(val)
	}

	return ucm
}

// convertURLMetadataList 将 interface{} 列表转换为 URLMetadata 列表。
func convertURLMetadataList(items []interface{}) []geminiTypes.URLMetadata {
	if len(items) == 0 {
		return nil
	}

	result := make([]geminiTypes.URLMetadata, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			um := geminiTypes.URLMetadata{}
			if val, ok := m["retrievedUrl"].(string); ok {
				um.RetrievedURL = val
			}
			if val, ok := m["urlRetrievalStatus"].(string); ok {
				um.URLRetrievalStatus = geminiTypes.URLRetrievalStatus(val)
			}
			result = append(result, um)
		}
	}

	return result
}

// convertMapToMediaResolution 将 map 转换为 MediaResolution。
func convertMapToMediaResolution(m map[string]interface{}) *geminiTypes.MediaResolution {
	if m == nil {
		return nil
	}

	mr := &geminiTypes.MediaResolution{}
	if val, ok := m["level"].(string); ok {
		mr.Level = val
	}

	return mr
}

// convertFunctionResponseParts 将 interface{} 列表转换为 FunctionResponsePart 列表。
func convertFunctionResponseParts(items []interface{}) []geminiTypes.FunctionResponsePart {
	if len(items) == 0 {
		return nil
	}

	result := make([]geminiTypes.FunctionResponsePart, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			frp := geminiTypes.FunctionResponsePart{}
			if val, ok := m["inlineData"].(map[string]interface{}); ok {
				frp.InlineData = convertMapToFunctionResponseBlob(val)
			}
			result = append(result, frp)
		}
	}

	return result
}

// convertMapToFunctionResponseBlob 将 map 转换为 FunctionResponseBlob。
func convertMapToFunctionResponseBlob(m map[string]interface{}) *geminiTypes.FunctionResponseBlob {
	if m == nil {
		return nil
	}

	frb := &geminiTypes.FunctionResponseBlob{}
	if val, ok := m["mimeType"].(string); ok {
		frb.MimeType = val
	}
	if val, ok := m["data"].(string); ok {
		frb.Data = val
	}

	return frb
}

// convertStringSlice 将 interface{} 列表转换为 string 列表。
func convertStringSlice(items []interface{}) []string {
	if len(items) == 0 {
		return nil
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		if val, ok := item.(string); ok {
			result = append(result, val)
		}
	}

	return result
}

// convertInt32Slice 将 interface{} 列表转换为 int32 列表。
func convertInt32Slice(items []interface{}) []int32 {
	if len(items) == 0 {
		return nil
	}

	result := make([]int32, 0, len(items))
	for _, item := range items {
		if val, ok := item.(float64); ok {
			result = append(result, int32(val))
		}
	}

	return result
}

// convertFloat32Slice 将 interface{} 列表转换为 float32 列表。
func convertFloat32Slice(items []interface{}) []float32 {
	if len(items) == 0 {
		return nil
	}

	result := make([]float32, 0, len(items))
	for _, item := range items {
		if val, ok := item.(float64); ok {
			result = append(result, float32(val))
		}
	}

	return result
}
