package converter

import (
	"encoding/json"
	"strings"

	"github.com/MeowSalty/portal/logger"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
)

// StreamEventToContract 将 Gemini 流式响应块转换为统一的 StreamEventContract 列表。
//
// 根据 plans/stream-event-contract-plan.md 5.2 章节实现：
//   - 每个流式 Response 块 -> message_delta
//   - candidates.finishReason 存在时追加 message_stop
//
// 根据 plans/stream-index-repair-plan.md B.1 章节实现索引补齐：
//   - sequence_number：如果 contract.SequenceNumber == 0 -> ctx.NextSequence()
//   - output_index：candidate.Index -> contract.OutputIndex；若缺失 -> ctx.EnsureOutputIndex(responseID)
//   - item_id：优先 tool_call id（如存在）；否则 response_id:output_index 组合；若仍缺失 -> ctx.EnsureItemID(BuildStreamIndexKey(...))
//   - content_index：Gemini parts 拆分时以 part 顺序作为 content_index；若无 parts -> ctx.EnsureContentIndex(item_id, -1)
func StreamEventToContract(event *geminiTypes.StreamEvent, ctx adapterTypes.StreamIndexContext, log logger.Logger) ([]*adapterTypes.StreamEventContract, error) {
	if event == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	log = log.WithGroup("stream_converter")

	usage := convertUsageMetadataToStreamUsage(event.UsageMetadata)
	responseExtensions := buildGeminiResponseExtensions(event)

	if len(event.Candidates) == 0 {
		contract := &adapterTypes.StreamEventContract{
			Type:         adapterTypes.StreamEventMessageDelta,
			Source:       adapterTypes.StreamSourceGemini,
			ResponseID:   event.ResponseID,
			MessageID:    event.ResponseID,
			OutputIndex:  -1, // 缺失索引使用 -1 哨兵
			ContentIndex: -1, // 缺失索引使用 -1 哨兵
			Usage:        usage,
			Extensions:   responseExtensions,
		}

		// 补齐索引字段
		fillMissingIndices(contract, event.ResponseID, 0, nil, -1, ctx)

		log.Warn("Gemini 流式响应缺少 candidates")
		return []*adapterTypes.StreamEventContract{contract}, nil
	}

	events := make([]*adapterTypes.StreamEventContract, 0, len(event.Candidates)*2)
	for _, candidate := range event.Candidates {
		contract := buildMessageDeltaEvent(event, &candidate, usage, responseExtensions, ctx, log)
		events = append(events, contract)

		if candidate.FinishReason != "" {
			stopEvent := buildMessageStopEvent(event, &candidate, responseExtensions, ctx)
			auditLog := log.WithGroup("audit")
			auditLog.Info("追加 Gemini message_stop 事件", "response_id", stopEvent.ResponseID, "output_index", stopEvent.OutputIndex, "finish_reason", candidate.FinishReason)
			events = append(events, stopEvent)
		}
	}

	return events, nil
}

// buildMessageDeltaEvent 构建 Gemini 的 message_delta 事件。
func buildMessageDeltaEvent(event *geminiTypes.StreamEvent, candidate *geminiTypes.Candidate, usage *adapterTypes.StreamUsagePayload, responseExtensions map[string]interface{}, ctx adapterTypes.StreamIndexContext, log logger.Logger) *adapterTypes.StreamEventContract {
	contract := &adapterTypes.StreamEventContract{
		Type:         adapterTypes.StreamEventMessageDelta,
		Source:       adapterTypes.StreamSourceGemini,
		ResponseID:   event.ResponseID,
		MessageID:    event.ResponseID,
		OutputIndex:  int(candidate.Index),
		ContentIndex: -1, // 缺失索引使用 -1 哨兵
		Usage:        usage,
		Extensions:   cloneExtensions(responseExtensions),
	}

	message := &adapterTypes.StreamMessagePayload{}
	if candidate.Content.Role != "" {
		message.Role = candidate.Content.Role
	}

	parts, toolCalls, contentText, contentIndex := convertPartsToStreamMessage(candidate.Content.Parts, ctx, log)
	if len(parts) > 0 {
		message.Parts = parts
	}
	if len(toolCalls) > 0 {
		message.ToolCalls = toolCalls
	}
	if contentText != nil {
		message.ContentText = contentText
	}

	if message.Role != "" || len(message.Parts) > 0 || len(message.ToolCalls) > 0 || message.ContentText != nil {
		contract.Message = message
	}

	mergeGeminiCandidateExtensions(contract, candidate)

	// 补齐索引字段
	fillMissingIndices(contract, event.ResponseID, int(candidate.Index), toolCalls, contentIndex, ctx)

	log.Debug("转换 Gemini message_delta 事件完成", "response_id", contract.ResponseID, "output_index", contract.OutputIndex)
	return contract
}

// buildMessageStopEvent 构建 Gemini 的 message_stop 事件。
func buildMessageStopEvent(event *geminiTypes.StreamEvent, candidate *geminiTypes.Candidate, responseExtensions map[string]interface{}, ctx adapterTypes.StreamIndexContext) *adapterTypes.StreamEventContract {
	contract := &adapterTypes.StreamEventContract{
		Type:         adapterTypes.StreamEventMessageStop,
		Source:       adapterTypes.StreamSourceGemini,
		ResponseID:   event.ResponseID,
		MessageID:    event.ResponseID,
		OutputIndex:  int(candidate.Index),
		ContentIndex: -1, // 缺失索引使用 -1 哨兵
		Extensions:   cloneExtensions(responseExtensions),
		Content: &adapterTypes.StreamContentPayload{
			Kind: "other",
			Raw: map[string]interface{}{
				"finish_reason": candidate.FinishReason,
			},
		},
	}

	if candidate.FinishMessage != "" {
		if contract.Extensions == nil {
			contract.Extensions = make(map[string]interface{})
		}
		geminiExt := ensureGeminiExtensions(contract.Extensions)
		geminiExt["finish_message"] = candidate.FinishMessage
	}

	// 补齐索引字段
	fillMissingIndices(contract, event.ResponseID, int(candidate.Index), nil, -1, ctx)

	return contract
}

// convertUsageMetadataToStreamUsage 转换 Gemini 使用统计到 StreamUsagePayload。
func convertUsageMetadataToStreamUsage(usage *geminiTypes.UsageMetadata) *adapterTypes.StreamUsagePayload {
	if usage == nil {
		return nil
	}

	result := &adapterTypes.StreamUsagePayload{
		Raw: make(map[string]interface{}),
	}

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
	} else if result.InputTokens != nil && result.OutputTokens != nil {
		totalTokens := *result.InputTokens + *result.OutputTokens
		result.TotalTokens = &totalTokens
	}

	if usage.CachedContentTokenCount > 0 {
		result.Raw["cached_content_token_count"] = usage.CachedContentTokenCount
	}
	if usage.ToolUsePromptTokenCount > 0 {
		result.Raw["tool_use_prompt_token_count"] = usage.ToolUsePromptTokenCount
	}
	if usage.ThoughtsTokenCount > 0 {
		result.Raw["thoughts_token_count"] = usage.ThoughtsTokenCount
	}
	if len(usage.PromptTokensDetails) > 0 {
		result.Raw["prompt_tokens_details"] = usage.PromptTokensDetails
	}
	if len(usage.CacheTokensDetails) > 0 {
		result.Raw["cache_tokens_details"] = usage.CacheTokensDetails
	}
	if len(usage.CandidatesTokensDetails) > 0 {
		result.Raw["candidates_tokens_details"] = usage.CandidatesTokensDetails
	}
	if len(usage.ToolUsePromptTokensDetails) > 0 {
		result.Raw["tool_use_prompt_tokens_details"] = usage.ToolUsePromptTokensDetails
	}

	if len(result.Raw) == 0 {
		result.Raw = nil
	}

	return result
}

// buildGeminiResponseExtensions 构建 Gemini 响应级扩展字段。
func buildGeminiResponseExtensions(event *geminiTypes.StreamEvent) map[string]interface{} {
	if event == nil {
		return nil
	}

	geminiExt := make(map[string]interface{})
	if event.PromptFeedback != nil {
		geminiExt["prompt_feedback"] = event.PromptFeedback
	}
	if event.ModelVersion != "" {
		geminiExt["model_version"] = event.ModelVersion
	}
	if event.ModelStatus != nil {
		geminiExt["model_status"] = event.ModelStatus
	}

	if len(geminiExt) == 0 {
		return nil
	}

	return map[string]interface{}{
		"gemini": geminiExt,
	}
}

// mergeGeminiCandidateExtensions 合并 Gemini 候选响应的特有字段。
func mergeGeminiCandidateExtensions(contract *adapterTypes.StreamEventContract, candidate *geminiTypes.Candidate) {
	if contract == nil || candidate == nil {
		return
	}

	if contract.Extensions == nil {
		contract.Extensions = make(map[string]interface{})
	}
	geminiExt := ensureGeminiExtensions(contract.Extensions)

	if len(candidate.SafetyRatings) > 0 {
		geminiExt["safety_ratings"] = candidate.SafetyRatings
	}
	if candidate.CitationMetadata != nil {
		geminiExt["citation_metadata"] = candidate.CitationMetadata
	}
	if len(candidate.GroundingAttributions) > 0 {
		geminiExt["grounding_attributions"] = candidate.GroundingAttributions
	}
	if candidate.GroundingMetadata != nil {
		geminiExt["grounding_metadata"] = candidate.GroundingMetadata
	}
	if candidate.AvgLogprobs != nil {
		geminiExt["avg_logprobs"] = *candidate.AvgLogprobs
	}
	if candidate.LogprobsResult != nil {
		geminiExt["logprobs_result"] = candidate.LogprobsResult
	}
	if candidate.URLContextMetadata != nil {
		geminiExt["url_context_metadata"] = candidate.URLContextMetadata
	}
	if candidate.TokenCount > 0 {
		geminiExt["token_count"] = candidate.TokenCount
	}
}

// convertPartsToStreamMessage 转换 Gemini 内容分片到 StreamMessagePayload 所需字段。
// 返回：parts, toolCalls, contentText, contentIndex（最后一个 part 的索引）
func convertPartsToStreamMessage(parts []geminiTypes.Part, ctx adapterTypes.StreamIndexContext, log logger.Logger) ([]adapterTypes.StreamContentPart, []adapterTypes.StreamToolCall, *string, int) {
	if len(parts) == 0 {
		return nil, nil, nil, -1
	}

	streamParts := make([]adapterTypes.StreamContentPart, 0, len(parts))
	toolCalls := make([]adapterTypes.StreamToolCall, 0)
	textSegments := make([]string, 0)

	for _, part := range parts {
		streamPart := adapterTypes.StreamContentPart{
			Raw: buildPartRaw(&part),
		}

		switch {
		case part.Text != nil:
			streamPart.Type = "text"
			streamPart.Text = *part.Text
			textSegments = append(textSegments, *part.Text)
		case part.FunctionCall != nil:
			streamPart.Type = "tool_use"
			streamPart.Raw["function_call"] = part.FunctionCall
			toolCall := convertFunctionCallToStreamToolCall(part.FunctionCall, log)
			if toolCall != nil {
				toolCalls = append(toolCalls, *toolCall)
			}
		case part.FunctionResponse != nil:
			streamPart.Type = "tool_result"
			streamPart.Raw["function_response"] = part.FunctionResponse
		case part.InlineData != nil:
			streamPart.Type = guessInlineDataType(part.InlineData.MimeType)
			streamPart.Raw["inline_data"] = part.InlineData
		case part.FileData != nil:
			streamPart.Type = "file"
			streamPart.Raw["file_data"] = part.FileData
		case part.ExecutableCode != nil:
			streamPart.Type = "executable_code"
			streamPart.Raw["executable_code"] = part.ExecutableCode
		case part.CodeExecutionResult != nil:
			streamPart.Type = "code_execution_result"
			streamPart.Raw["code_execution_result"] = part.CodeExecutionResult
		default:
			streamPart.Type = "other"
		}

		if len(streamPart.Raw) == 0 {
			streamPart.Raw = nil
		}
		streamParts = append(streamParts, streamPart)
	}

	if len(textSegments) == 0 {
		return streamParts, toolCalls, nil, len(parts) - 1
	}

	contentText := strings.Join(textSegments, "\n")
	return streamParts, toolCalls, &contentText, len(parts) - 1
}

// convertFunctionCallToStreamToolCall 转换 Gemini FunctionCall 到 StreamToolCall。
func convertFunctionCallToStreamToolCall(fc *geminiTypes.FunctionCall, log logger.Logger) *adapterTypes.StreamToolCall {
	if fc == nil {
		return nil
	}

	toolCall := &adapterTypes.StreamToolCall{
		Type: "function",
		Name: fc.Name,
		Raw:  map[string]interface{}{},
	}

	if fc.ID != nil {
		toolCall.ID = *fc.ID
	}

	if fc.Args != nil {
		argsJSON, err := json.Marshal(fc.Args)
		if err != nil {
			log.Warn("序列化 Gemini FunctionCall Args 失败", "error", err)
		} else {
			toolCall.Arguments = string(argsJSON)
		}
	}

	if len(toolCall.Raw) == 0 {
		toolCall.Raw = nil
	}

	return toolCall
}

// buildPartRaw 构建 Gemini Part 的原始字段映射。
func buildPartRaw(part *geminiTypes.Part) map[string]interface{} {
	if part == nil {
		return nil
	}

	raw := make(map[string]interface{})
	if part.Thought != nil {
		raw["thought"] = *part.Thought
	}
	if part.ThoughtSignature != nil {
		raw["thought_signature"] = *part.ThoughtSignature
	}
	if part.PartMetadata != nil {
		raw["part_metadata"] = part.PartMetadata
	}
	if part.MediaResolution != nil {
		raw["media_resolution"] = part.MediaResolution
	}
	if part.VideoMetadata != nil {
		raw["video_metadata"] = part.VideoMetadata
	}

	return raw
}

// guessInlineDataType 根据 MIME 类型推断 Part 类型。
func guessInlineDataType(mimeType string) string {
	lower := strings.ToLower(mimeType)
	if strings.HasPrefix(lower, "image/") {
		return "image"
	}
	if strings.HasPrefix(lower, "audio/") {
		return "audio"
	}
	if strings.HasPrefix(lower, "video/") {
		return "video"
	}
	return "file"
}

// ensureGeminiExtensions 获取或创建 Gemini 扩展命名空间。
func ensureGeminiExtensions(extensions map[string]interface{}) map[string]interface{} {
	if extensions == nil {
		return map[string]interface{}{}
	}

	if ext, ok := extensions["gemini"]; ok {
		if geminiExt, ok := ext.(map[string]interface{}); ok {
			return geminiExt
		}
	}

	geminiExt := make(map[string]interface{})
	extensions["gemini"] = geminiExt
	return geminiExt
}

// cloneExtensions 复制扩展字段，避免复用引用。
func cloneExtensions(extensions map[string]interface{}) map[string]interface{} {
	if len(extensions) == 0 {
		return nil
	}

	clone := make(map[string]interface{}, len(extensions))
	for key, value := range extensions {
		clone[key] = value
	}
	return clone
}

// fillMissingIndices 补齐缺失的索引字段
//
// 根据 plans/stream-index-repair-plan.md B.1 章节实现：
//   - sequence_number：如果 contract.SequenceNumber == 0 -> ctx.NextSequence()
//   - output_index：若缺失（< 0）-> ctx.EnsureOutputIndex(responseID)，0 值保持不变
//   - item_id：优先 tool_call id（如存在）；否则 response_id:output_index 组合；若仍缺失 -> ctx.EnsureItemID(BuildStreamIndexKey(...))
//   - content_index：Gemini parts 拆分时以 part 顺序作为 content_index；若无 parts -> ctx.EnsureContentIndex(item_id, -1)
//   - 缺失索引统一使用 -1 作为哨兵值，与合法 0 区分
func fillMissingIndices(contract *adapterTypes.StreamEventContract, responseID string, candidateIndex int, toolCalls []adapterTypes.StreamToolCall, contentIndex int, ctx adapterTypes.StreamIndexContext) {
	// 补齐 sequence_number
	if contract.SequenceNumber == 0 {
		contract.SequenceNumber = ctx.NextSequence()
	}

	// 补齐 output_index（仅处理负值，0 值保持不变）
	if contract.OutputIndex < 0 {
		contract.OutputIndex = ctx.EnsureOutputIndex(responseID)
	}

	// 补齐 item_id
	if contract.ItemID == "" {
		// 优先使用 tool_call id
		if len(toolCalls) > 0 && toolCalls[0].ID != "" {
			contract.ItemID = toolCalls[0].ID
		} else {
			// 使用 EnsureItemID 生成稳定的 item_id
			key := adapterTypes.BuildStreamIndexKey(responseID, contract.OutputIndex, -1)
			contract.ItemID = ctx.EnsureItemID(key)
		}
	}

	// 补齐 content_index（仅处理负值，0 值保持不变）
	if contract.ContentIndex < 0 {
		if contentIndex >= 0 {
			// 若有 parts，使用 part 顺序作为 content_index
			contract.ContentIndex = ctx.EnsureContentIndex(contract.ItemID, contentIndex)
		} else {
			// 若无 parts，使用 EnsureContentIndex
			contract.ContentIndex = ctx.EnsureContentIndex(contract.ItemID, -1)
		}
	}
}
