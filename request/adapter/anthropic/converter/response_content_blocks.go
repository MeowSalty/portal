package converter

import (
	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertResponseToMessage 转换响应为 Message。
func convertResponseToMessage(resp *anthropicTypes.Response, log logger.Logger) (*types.ResponseMessage, error) {
	message := &types.ResponseMessage{
		Extras: make(map[string]interface{}),
	}

	// 设置 Role
	role := string(resp.Role)
	message.Role = &role

	// 保存原始 Content blocks 到 Extras（用于反向转换）
	if len(resp.Content) > 0 {
		contentBlocksJSON := SafeMarshal(resp.Content)
		if len(contentBlocksJSON) == 0 {
			log.Warn("序列化 content blocks 失败", "error", nil)
		} else if err := SaveVendorExtraRaw("anthropic.content_blocks", contentBlocksJSON, message.Extras); err != nil {
			log.Warn("保存 content blocks 失败", "error", err)
		}
	}

	// 转换 Content blocks
	parts, toolCalls, toolResults, err := convertResponseContentBlocksToContract(resp.Content, log)
	if err != nil {
		return nil, err
	}

	message.Parts = parts
	message.ToolCalls = toolCalls
	message.ToolResults = toolResults

	// 聚合纯文本内容
	var textContent string
	for _, part := range parts {
		if part.Type == "text" && part.Text != nil {
			if textContent != "" {
				textContent += "\n"
			}
			textContent += *part.Text
		}
	}
	if textContent != "" {
		message.Content = &textContent
	}

	return message, nil
}

// convertResponseContentBlocksToContract 转换内容块列表。
func convertResponseContentBlocksToContract(blocks []anthropicTypes.ResponseContentBlock, log logger.Logger) (
	[]types.ResponseContentPart,
	[]types.ResponseToolCall,
	[]types.ResponseToolResult,
	error,
) {
	var parts []types.ResponseContentPart
	var toolCalls []types.ResponseToolCall
	var toolResults []types.ResponseToolResult

	for _, block := range blocks {
		if block.Text != nil {
			part, err := convertResponseTextBlockToContract(block.Text)
			if err != nil {
				log.Error("转换 TextBlock 失败", "error", err)
				return nil, nil, nil, err
			}
			parts = append(parts, *part)
		} else if block.Thinking != nil {
			part := convertResponseThinkingBlockToContract(block.Thinking)
			parts = append(parts, *part)
		} else if block.RedactedThinking != nil {
			part := convertResponseRedactedThinkingBlockToContract(block.RedactedThinking)
			parts = append(parts, *part)
		} else if block.ToolUse != nil {
			toolCall := convertResponseToolUseBlockToContract(block.ToolUse)
			toolCalls = append(toolCalls, *toolCall)
		} else if block.ServerToolUse != nil {
			toolCall := convertResponseServerToolUseBlockToContract(block.ServerToolUse)
			toolCalls = append(toolCalls, *toolCall)
		} else if block.WebSearchToolResult != nil {
			result, err := convertResponseWebSearchToolResultBlockToContract(block.WebSearchToolResult)
			if err != nil {
				log.Error("转换 WebSearchToolResultBlock 失败", "error", err)
				return nil, nil, nil, err
			}
			toolResults = append(toolResults, *result)
		}
	}

	return parts, toolCalls, toolResults, nil
}

// convertResponseTextBlockToContract 转换文本块。
func convertResponseTextBlockToContract(block *anthropicTypes.TextBlock) (*types.ResponseContentPart, error) {
	part := &types.ResponseContentPart{
		Type: "text",
		Text: &block.Text,
	}

	// 转换 Citations
	if len(block.Citations) > 0 {
		annotations, err := convertCitationsToAnnotations(block.Citations)
		if err != nil {
			return nil, err
		}
		part.Annotations = annotations
	}

	return part, nil
}

// convertResponseThinkingBlockToContract 转换思考块。
func convertResponseThinkingBlockToContract(block *anthropicTypes.ThinkingBlock) *types.ResponseContentPart {
	part := &types.ResponseContentPart{
		Type:   "thinking",
		Text:   &block.Thinking,
		Extras: make(map[string]interface{}),
	}

	if err := SaveVendorExtra("anthropic.signature", block.Signature, part.Extras); err != nil {
		logger.Default().Warn("保存思考签名失败", "error", err)
	}

	return part
}

// convertResponseRedactedThinkingBlockToContract 转换脱敏思考块。
func convertResponseRedactedThinkingBlockToContract(block *anthropicTypes.RedactedThinkingBlock) *types.ResponseContentPart {
	part := &types.ResponseContentPart{
		Type:   "thinking",
		Extras: make(map[string]interface{}),
	}

	if err := SaveVendorExtra("anthropic.redacted", true, part.Extras); err != nil {
		logger.Default().Warn("保存脱敏标记失败", "error", err)
	}
	if err := SaveVendorExtra("anthropic.data", block.Data, part.Extras); err != nil {
		logger.Default().Warn("保存脱敏数据失败", "error", err)
	}

	return part
}

// convertResponseToolUseBlockToContract 转换工具使用块。
func convertResponseToolUseBlockToContract(block *anthropicTypes.ToolUseBlock) *types.ResponseToolCall {
	toolCall := &types.ResponseToolCall{
		ID:      &block.ID,
		Name:    &block.Name,
		Payload: block.Input,
		Extras:  make(map[string]interface{}),
	}

	toolType := "tool_use"
	toolCall.Type = &toolType

	return toolCall
}

// convertResponseServerToolUseBlockToContract 转换服务器工具使用块。
func convertResponseServerToolUseBlockToContract(block *anthropicTypes.ServerToolUseBlock) *types.ResponseToolCall {
	toolCall := &types.ResponseToolCall{
		ID:      &block.ID,
		Name:    &block.Name,
		Payload: block.Input,
		Extras:  make(map[string]interface{}),
	}

	toolType := "server_tool_use"
	toolCall.Type = &toolType
	if err := SaveVendorExtra("anthropic.server_tool", true, toolCall.Extras); err != nil {
		logger.Default().Warn("保存 ServerTool 标记失败", "error", err)
	}

	return toolCall
}

// convertResponseWebSearchToolResultBlockToContract 转换 Web 搜索工具结果块。
func convertResponseWebSearchToolResultBlockToContract(block *anthropicTypes.WebSearchToolResultBlock) (*types.ResponseToolResult, error) {
	result := &types.ResponseToolResult{
		ID:      &block.ToolUseID,
		Payload: make(map[string]interface{}),
		Extras:  make(map[string]interface{}),
	}

	// 保存完整的内容结构
	if block.Content.Error != nil {
		result.Payload["error"] = block.Content.Error
		if err := SaveVendorExtra("anthropic.has_error", true, result.Extras); err != nil {
			logger.Default().Warn("保存 WebSearch 错误标记失败", "error", err)
		}
	} else if len(block.Content.Results) > 0 {
		result.Payload["results"] = block.Content.Results
	}

	if err := SaveVendorExtra("anthropic.web_search_result", true, result.Extras); err != nil {
		logger.Default().Warn("保存 WebSearch 标记失败", "error", err)
	}

	return result, nil
}

// convertMessageToResponseContent 从 Message 转换为响应内容块。
func convertMessageToResponseContent(message *types.ResponseMessage, log logger.Logger) ([]anthropicTypes.ResponseContentBlock, error) {
	// 优先使用 Extras 中保存的原始 content_blocks
	if contentBlocksJSON, ok := GetVendorExtraRaw("anthropic.content_blocks", message.Extras); ok {
		var blocks []anthropicTypes.ResponseContentBlock
		if SafeUnmarshal(contentBlocksJSON, &blocks) {
			return blocks, nil
		}
		log.Warn("反序列化原始 content blocks 失败，将从 Parts 重建", "error", nil)
	}

	// 从 Parts、ToolCalls 和 ToolResults 重建 content blocks
	var blocks []anthropicTypes.ResponseContentBlock

	// 转换 Parts
	for _, part := range message.Parts {
		block, err := convertPartToResponseContentBlock(&part)
		if err != nil {
			log.Error("转换 Part 失败", "error", err)
			return nil, err
		}
		if block != nil {
			blocks = append(blocks, *block)
		}
	}

	// 转换 ToolCalls
	for _, toolCall := range message.ToolCalls {
		block, err := convertToolCallToResponseContentBlock(&toolCall)
		if err != nil {
			log.Error("转换 ToolCall 失败", "error", err)
			return nil, err
		}
		if block != nil {
			blocks = append(blocks, *block)
		}
	}

	// 转换 ToolResults
	for _, toolResult := range message.ToolResults {
		block, err := convertToolResultToResponseContentBlock(&toolResult)
		if err != nil {
			log.Error("转换 ToolResult 失败", "error", err)
			return nil, err
		}
		if block != nil {
			blocks = append(blocks, *block)
		}
	}

	return blocks, nil
}

// convertPartToResponseContentBlock 从 Part 转换为响应内容块。
func convertPartToResponseContentBlock(part *types.ResponseContentPart) (*anthropicTypes.ResponseContentBlock, error) {
	switch part.Type {
	case "text":
		if part.Text == nil {
			return nil, nil
		}
		textBlock := &anthropicTypes.TextBlock{
			Type: anthropicTypes.ResponseContentBlockText,
			Text: *part.Text,
		}
		// 转换 Annotations 为 Citations
		if len(part.Annotations) > 0 {
			citations, err := convertAnnotationsToCitations(part.Annotations)
			if err != nil {
				return nil, err
			}
			textBlock.Citations = citations
		}
		return &anthropicTypes.ResponseContentBlock{Text: textBlock}, nil

	case "thinking":
		// 检查是否是 redacted thinking
		var redacted bool
		if found, err := GetVendorExtra("anthropic.redacted", part.Extras, &redacted); err == nil && found && redacted {
			var data string
			if dataFound, dataErr := GetVendorExtra("anthropic.data", part.Extras, &data); dataErr == nil && dataFound {
				return &anthropicTypes.ResponseContentBlock{
					RedactedThinking: &anthropicTypes.RedactedThinkingBlock{
						Type: anthropicTypes.ResponseContentBlockRedactedThinking,
						Data: data,
					},
				}, nil
			}
		}

		// 普通 thinking block
		if part.Text == nil {
			return nil, nil
		}
		thinkingBlock := &anthropicTypes.ThinkingBlock{
			Type:     anthropicTypes.ResponseContentBlockThinking,
			Thinking: *part.Text,
		}
		var signature string
		if found, err := GetVendorExtra("anthropic.signature", part.Extras, &signature); err == nil && found {
			thinkingBlock.Signature = signature
		}
		return &anthropicTypes.ResponseContentBlock{Thinking: thinkingBlock}, nil

	default:
		// 其他类型暂不处理
		return nil, nil
	}
}

// convertToolCallToResponseContentBlock 从 ToolCall 转换为响应内容块。
func convertToolCallToResponseContentBlock(toolCall *types.ResponseToolCall) (*anthropicTypes.ResponseContentBlock, error) {
	if toolCall.Type != nil && *toolCall.Type == "server_tool_use" {
		// ServerToolUse
		if toolCall.ID == nil || toolCall.Name == nil {
			return nil, nil
		}
		return &anthropicTypes.ResponseContentBlock{
			ServerToolUse: &anthropicTypes.ServerToolUseBlock{
				Type:  anthropicTypes.ResponseContentBlockServerToolUse,
				ID:    *toolCall.ID,
				Name:  *toolCall.Name,
				Input: toolCall.Payload,
			},
		}, nil
	}

	// 普通 ToolUse
	if toolCall.ID == nil || toolCall.Name == nil {
		return nil, nil
	}
	return &anthropicTypes.ResponseContentBlock{
		ToolUse: &anthropicTypes.ToolUseBlock{
			Type:  anthropicTypes.ResponseContentBlockToolUse,
			ID:    *toolCall.ID,
			Name:  *toolCall.Name,
			Input: toolCall.Payload,
		},
	}, nil
}

// convertToolResultToResponseContentBlock 从 ToolResult 转换为响应内容块。
func convertToolResultToResponseContentBlock(toolResult *types.ResponseToolResult) (*anthropicTypes.ResponseContentBlock, error) {
	// 检查是否是 WebSearchToolResult
	var isWebSearch bool
	if found, err := GetVendorExtra("anthropic.web_search_result", toolResult.Extras, &isWebSearch); err == nil && found && isWebSearch {
		if toolResult.ID == nil {
			return nil, nil
		}

		content := anthropicTypes.WebSearchToolResultBlockContent{}

		// 从 Payload 恢复内容
		if errData, ok := toolResult.Payload["error"]; ok {
			if webSearchErr, ok := errData.(*anthropicTypes.WebSearchToolResultError); ok {
				content.Error = webSearchErr
			}
		} else if resultsData, ok := toolResult.Payload["results"]; ok {
			if results, ok := resultsData.([]anthropicTypes.WebSearchResultBlock); ok {
				content.Results = results
			}
		}

		return &anthropicTypes.ResponseContentBlock{
			WebSearchToolResult: &anthropicTypes.WebSearchToolResultBlock{
				Type:      anthropicTypes.ResponseContentBlockWebSearchToolResult,
				ToolUseID: *toolResult.ID,
				Content:   content,
			},
		}, nil
	}

	return nil, nil
}
