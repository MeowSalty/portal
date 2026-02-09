package converter

import (
	"encoding/json"

	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertMessagesFromContract 从 Contract 转换消息列表。
func convertMessagesFromContract(messages []types.Message) ([]anthropicTypes.Message, error) {
	result := make([]anthropicTypes.Message, 0, len(messages))

	for _, msg := range messages {
		inputMsg := anthropicTypes.Message{
			Role: anthropicTypes.Role(msg.Role),
		}

		// 转换 Content
		if msg.Content.Text != nil {
			inputMsg.Content = anthropicTypes.MessageContentParam{
				StringValue: msg.Content.Text,
			}
		} else if len(msg.Content.Parts) > 0 || len(msg.ToolCalls) > 0 {
			blocks, err := convertContentPartsFromContract(msg.Content.Parts, msg.ToolCalls)
			if err != nil {
				return nil, err
			}
			inputMsg.Content = anthropicTypes.MessageContentParam{
				Blocks: blocks,
			}
		}

		result = append(result, inputMsg)
	}

	return result, nil
}

// convertContentPartsFromContract 从 Contract 转换内容块列表。
func convertContentPartsFromContract(parts []types.ContentPart, toolCalls []types.ToolCall) ([]anthropicTypes.ContentBlockParam, error) {
	blocks := make([]anthropicTypes.ContentBlockParam, 0, len(parts)+len(toolCalls))

	for _, part := range parts {
		block, err := convertContentPartFromContract(&part)
		if err != nil {
			return nil, err
		}
		if block != nil {
			blocks = append(blocks, *block)
		}
	}

	// 转换 ToolCalls
	for _, tc := range toolCalls {
		block, err := convertToolCallFromContract(&tc)
		if err != nil {
			return nil, err
		}
		if block != nil {
			blocks = append(blocks, *block)
		}
	}

	return blocks, nil
}

// convertContentPartFromContract 从 Contract 转换单个内容块。
func convertContentPartFromContract(part *types.ContentPart) (*anthropicTypes.ContentBlockParam, error) {
	if part == nil {
		return nil, nil
	}

	block := &anthropicTypes.ContentBlockParam{}

	switch part.Type {
	case "text":
		if part.Text != nil {
			textBlock := &anthropicTypes.TextBlockParam{
				Type: anthropicTypes.ContentBlockTypeText,
				Text: *part.Text,
			}

			// 从 VendorExtras 恢复 CacheControl 和 Citations
			if part.VendorExtras != nil {
				if cc, ok := part.VendorExtras["cache_control"].(*anthropicTypes.CacheControlEphemeral); ok {
					textBlock.CacheControl = cc
				}
				if citations, ok := part.VendorExtras["citations"].([]anthropicTypes.TextCitationParam); ok {
					textBlock.Citations = citations
				}
			}

			block.Text = textBlock
		}

	case "image":
		if part.Image != nil {
			imageBlock := &anthropicTypes.ImageBlockParam{
				Type: anthropicTypes.ContentBlockTypeImage,
			}

			if part.Image.Data != nil {
				imageBlock.Source = anthropicTypes.ImageSource{
					Base64: &anthropicTypes.Base64ImageSource{
						Type: anthropicTypes.ImageSourceTypeBase64,
						Data: *part.Image.Data,
					},
				}
				if part.Image.MIME != nil {
					imageBlock.Source.Base64.MediaType = anthropicTypes.ImageMediaType(*part.Image.MIME)
				}
			} else if part.Image.URL != nil {
				imageBlock.Source = anthropicTypes.ImageSource{
					URL: &anthropicTypes.URLImageSource{
						Type: anthropicTypes.ImageSourceTypeURL,
						URL:  *part.Image.URL,
					},
				}
			}

			// 从 VendorExtras 恢复 CacheControl
			if part.VendorExtras != nil {
				if cc, ok := part.VendorExtras["cache_control"].(*anthropicTypes.CacheControlEphemeral); ok {
					imageBlock.CacheControl = cc
				}
			}

			block.Image = imageBlock
		}

	case "document":
		if part.File != nil {
			docBlock := &anthropicTypes.DocumentBlockParam{
				Type: anthropicTypes.ContentBlockTypeDocument,
			}

			if part.File.Data != nil {
				if part.File.MIME != nil && *part.File.MIME == "application/pdf" {
					docBlock.Source = anthropicTypes.DocumentSource{
						Base64: &anthropicTypes.Base64PDFSource{
							Type:      anthropicTypes.DocumentSourceTypeBase64,
							MediaType: anthropicTypes.DocumentMediaTypePDF,
							Data:      *part.File.Data,
						},
					}
				} else {
					docBlock.Source = anthropicTypes.DocumentSource{
						Text: &anthropicTypes.PlainTextSource{
							Type:      anthropicTypes.DocumentSourceTypeText,
							MediaType: anthropicTypes.DocumentMediaTypeTextPlain,
							Data:      *part.File.Data,
						},
					}
				}
			} else if part.File.URL != nil {
				docBlock.Source = anthropicTypes.DocumentSource{
					URL: &anthropicTypes.URLPDFSource{
						Type: anthropicTypes.DocumentSourceTypeURL,
						URL:  *part.File.URL,
					},
				}
			}

			// 从 VendorExtras 恢复其他字段
			if part.VendorExtras != nil {
				if cc, ok := part.VendorExtras["cache_control"].(*anthropicTypes.CacheControlEphemeral); ok {
					docBlock.CacheControl = cc
				}
				if citations, ok := part.VendorExtras["citations"].(*anthropicTypes.CitationsConfigParam); ok {
					docBlock.Citations = citations
				}
				if context, ok := part.VendorExtras["context"].(*string); ok {
					docBlock.Context = context
				}
			}

			if part.File.Filename != nil {
				docBlock.Title = part.File.Filename
			}

			block.Document = docBlock
		}

	case "tool_result":
		if part.ToolResult != nil {
			toolResultBlock := &anthropicTypes.ToolResultBlockParam{
				Type:      anthropicTypes.ContentBlockTypeToolResult,
				ToolUseID: *part.ToolResult.ID,
			}

			if part.ToolResult.Content != nil {
				toolResultBlock.Content = anthropicTypes.ToolResultContentParam{
					StringValue: part.ToolResult.Content,
				}
			}

			// 从 VendorExtras 恢复其他字段
			if part.VendorExtras != nil {
				if cc, ok := part.VendorExtras["cache_control"].(*anthropicTypes.CacheControlEphemeral); ok {
					toolResultBlock.CacheControl = cc
				}
				if isError, ok := part.VendorExtras["is_error"].(*bool); ok {
					toolResultBlock.IsError = isError
				}
			}

			block.ToolResult = toolResultBlock
		}

	case "thinking":
		// 从 VendorExtras 恢复思考块
		if part.VendorExtras != nil {
			thinkingBlock := &anthropicTypes.ThinkingBlockParam{
				Type: anthropicTypes.ContentBlockTypeThinking,
			}
			if sig, ok := part.VendorExtras["signature"].(string); ok {
				thinkingBlock.Signature = sig
			}
			if thinking, ok := part.VendorExtras["thinking"].(string); ok {
				thinkingBlock.Thinking = thinking
			}
			block.Thinking = thinkingBlock
		}

	case "redacted_thinking":
		// 从 VendorExtras 恢复脱敏思考块
		if part.VendorExtras != nil {
			redactedBlock := &anthropicTypes.RedactedThinkingBlockParam{
				Type: anthropicTypes.ContentBlockTypeRedactedThinking,
			}
			if data, ok := part.VendorExtras["data"].(string); ok {
				redactedBlock.Data = data
			}
			block.RedactedThinking = redactedBlock
		}

	default:
		return nil, nil
	}

	return block, nil
}

// convertMessagesToContract 转换消息列表。
func convertMessagesToContract(messages []anthropicTypes.Message) ([]types.Message, error) {
	result := make([]types.Message, 0, len(messages))

	for _, msg := range messages {
		contractMsg := types.Message{
			Role: string(msg.Role),
		}

		// 转换 Content
		if msg.Content.StringValue != nil {
			contractMsg.Content = types.Content{
				Text: msg.Content.StringValue,
			}
		} else if len(msg.Content.Blocks) > 0 {
			parts, toolCalls, err := convertContentBlocksToContract(msg.Content.Blocks)
			if err != nil {
				return nil, err
			}
			contractMsg.Content = types.Content{
				Parts: parts,
			}
			if len(toolCalls) > 0 {
				contractMsg.ToolCalls = toolCalls
			}
		}

		result = append(result, contractMsg)
	}

	return result, nil
}

// convertContentBlocksToContract 转换内容块列表。
func convertContentBlocksToContract(blocks []anthropicTypes.ContentBlockParam) ([]types.ContentPart, []types.ToolCall, error) {
	parts := make([]types.ContentPart, 0, len(blocks))
	toolCalls := make([]types.ToolCall, 0)

	for _, block := range blocks {
		if block.Text != nil {
			part, err := convertTextBlockToContract(block.Text)
			if err != nil {
				return nil, nil, err
			}
			parts = append(parts, *part)
		} else if block.Image != nil {
			part, err := convertImageBlockToContract(block.Image)
			if err != nil {
				return nil, nil, err
			}
			parts = append(parts, *part)
		} else if block.Document != nil {
			part, err := convertDocumentBlockToContract(block.Document)
			if err != nil {
				return nil, nil, err
			}
			parts = append(parts, *part)
		} else if block.ToolUse != nil {
			toolCall, err := convertToolUseToContract(block.ToolUse)
			if err != nil {
				return nil, nil, err
			}
			toolCalls = append(toolCalls, *toolCall)
		} else if block.ToolResult != nil {
			part, err := convertToolResultToContract(block.ToolResult)
			if err != nil {
				return nil, nil, err
			}
			parts = append(parts, *part)
		} else if block.SearchResult != nil {
			part, err := convertSearchResultToContract(block.SearchResult)
			if err != nil {
				return nil, nil, err
			}
			parts = append(parts, *part)
		} else if block.Thinking != nil {
			part := convertThinkingBlockToContract(block.Thinking)
			parts = append(parts, *part)
		} else if block.RedactedThinking != nil {
			part := convertRedactedThinkingBlockToContract(block.RedactedThinking)
			parts = append(parts, *part)
		} else if block.ServerToolUse != nil {
			part := convertServerToolUseToContract(block.ServerToolUse)
			parts = append(parts, *part)
		} else if block.WebSearchToolResult != nil {
			part := convertWebSearchToolResultToContract(block.WebSearchToolResult)
			parts = append(parts, *part)
		}
	}

	return parts, toolCalls, nil
}

// convertTextBlockToContract 转换文本块。
func convertTextBlockToContract(block *anthropicTypes.TextBlockParam) (*types.ContentPart, error) {
	part := &types.ContentPart{
		Type: "text",
		Text: &block.Text,
	}

	// Citations 和 CacheControl 放入 VendorExtras
	if block.Citations != nil || block.CacheControl != nil {
		part.VendorExtras = make(map[string]interface{})
		source := types.VendorSourceAnthropic
		part.VendorExtrasSource = &source

		if block.Citations != nil {
			part.VendorExtras["citations"] = block.Citations
		}
		if block.CacheControl != nil {
			part.VendorExtras["cache_control"] = block.CacheControl
		}
	}

	return part, nil
}

// convertImageBlockToContract 转换图片块。
func convertImageBlockToContract(block *anthropicTypes.ImageBlockParam) (*types.ContentPart, error) {
	part := &types.ContentPart{
		Type:  "image",
		Image: &types.Image{},
	}

	// 转换图片来源
	if block.Source.Base64 != nil {
		part.Image.Data = &block.Source.Base64.Data
		mime := string(block.Source.Base64.MediaType)
		part.Image.MIME = &mime
	} else if block.Source.URL != nil {
		part.Image.URL = &block.Source.URL.URL
	}

	// CacheControl 放入 VendorExtras
	if block.CacheControl != nil {
		part.VendorExtras = make(map[string]interface{})
		source := types.VendorSourceAnthropic
		part.VendorExtrasSource = &source
		part.VendorExtras["cache_control"] = block.CacheControl
	}

	return part, nil
}

// convertDocumentBlockToContract 转换文档块。
func convertDocumentBlockToContract(block *anthropicTypes.DocumentBlockParam) (*types.ContentPart, error) {
	part := &types.ContentPart{
		Type: "document",
		File: &types.File{},
	}

	// 转换文档来源
	if block.Source.Base64 != nil {
		part.File.Data = &block.Source.Base64.Data
		mime := string(block.Source.Base64.MediaType)
		part.File.MIME = &mime
	} else if block.Source.Text != nil {
		part.File.Data = &block.Source.Text.Data
		mime := string(block.Source.Text.MediaType)
		part.File.MIME = &mime
	} else if block.Source.URL != nil {
		part.File.URL = &block.Source.URL.URL
	} else if block.Source.Content != nil {
		// Content 类型的文档来源放入 VendorExtras
		if part.VendorExtras == nil {
			part.VendorExtras = make(map[string]interface{})
		}
		part.VendorExtras["document_source_content"] = block.Source.Content
	}

	// 其他字段放入 VendorExtras
	if block.CacheControl != nil || block.Citations != nil || block.Context != nil || block.Title != nil {
		if part.VendorExtras == nil {
			part.VendorExtras = make(map[string]interface{})
		}
		source := types.VendorSourceAnthropic
		part.VendorExtrasSource = &source

		if block.CacheControl != nil {
			part.VendorExtras["cache_control"] = block.CacheControl
		}
		if block.Citations != nil {
			part.VendorExtras["citations"] = block.Citations
		}
		if block.Context != nil {
			part.VendorExtras["context"] = block.Context
		}
		if block.Title != nil {
			part.File.Filename = block.Title
		}
	}

	return part, nil
}

// convertToolUseToContract 转换工具使用块。
func convertToolUseToContract(block *anthropicTypes.ToolUseBlockParam) (*types.ToolCall, error) {
	toolCall := &types.ToolCall{
		ID:   &block.ID,
		Name: &block.Name,
	}

	// 序列化 Input 为 JSON 字符串
	if block.Input != nil {
		inputJSON, err := json.Marshal(block.Input)
		if err != nil {
			return nil, err
		}
		inputStr := string(inputJSON)
		toolCall.Arguments = &inputStr
	}

	toolType := "function"
	toolCall.Type = &toolType

	return toolCall, nil
}

// convertToolResultToContract 转换工具结果块。
func convertToolResultToContract(block *anthropicTypes.ToolResultBlockParam) (*types.ContentPart, error) {
	part := &types.ContentPart{
		Type: "tool_result",
		ToolResult: &types.ToolResult{
			ID: &block.ToolUseID,
		},
	}

	// 转换 Content
	if block.Content.StringValue != nil {
		part.ToolResult.Content = block.Content.StringValue
	} else if len(block.Content.Blocks) > 0 {
		// 将复杂内容序列化为 JSON 放入 Payload
		part.ToolResult.Payload = make(map[string]interface{})
		part.ToolResult.Payload["blocks"] = block.Content.Blocks
	}

	// 其他字段放入 VendorExtras
	if block.CacheControl != nil || block.IsError != nil {
		part.VendorExtras = make(map[string]interface{})
		source := types.VendorSourceAnthropic
		part.VendorExtrasSource = &source

		if block.CacheControl != nil {
			part.VendorExtras["cache_control"] = block.CacheControl
		}
		if block.IsError != nil {
			part.VendorExtras["is_error"] = block.IsError
		}
	}

	return part, nil
}

// convertSearchResultToContract 转换搜索结果块。
func convertSearchResultToContract(block *anthropicTypes.SearchResultBlockParam) (*types.ContentPart, error) {
	part := &types.ContentPart{
		Type: "search_result",
	}

	// 将搜索结果的详细信息放入 VendorExtras
	part.VendorExtras = make(map[string]interface{})
	source := types.VendorSourceAnthropic
	part.VendorExtrasSource = &source

	part.VendorExtras["content"] = block.Content
	part.VendorExtras["source"] = block.Source
	part.VendorExtras["title"] = block.Title

	if block.CacheControl != nil {
		part.VendorExtras["cache_control"] = block.CacheControl
	}
	if block.Citations != nil {
		part.VendorExtras["citations"] = block.Citations
	}

	return part, nil
}

// convertThinkingBlockToContract 转换思考块。
func convertThinkingBlockToContract(block *anthropicTypes.ThinkingBlockParam) *types.ContentPart {
	part := &types.ContentPart{
		Type: "thinking",
	}

	part.VendorExtras = make(map[string]interface{})
	source := types.VendorSourceAnthropic
	part.VendorExtrasSource = &source

	part.VendorExtras["signature"] = block.Signature
	part.VendorExtras["thinking"] = block.Thinking

	return part
}

// convertRedactedThinkingBlockToContract 转换脱敏思考块。
func convertRedactedThinkingBlockToContract(block *anthropicTypes.RedactedThinkingBlockParam) *types.ContentPart {
	part := &types.ContentPart{
		Type: "redacted_thinking",
	}

	part.VendorExtras = make(map[string]interface{})
	source := types.VendorSourceAnthropic
	part.VendorExtrasSource = &source

	part.VendorExtras["data"] = block.Data

	return part
}

// convertServerToolUseToContract 转换服务器工具使用块。
func convertServerToolUseToContract(block *anthropicTypes.ServerToolUseBlockParam) *types.ContentPart {
	part := &types.ContentPart{
		Type: "server_tool_use",
	}

	part.VendorExtras = make(map[string]interface{})
	source := types.VendorSourceAnthropic
	part.VendorExtrasSource = &source

	part.VendorExtras["id"] = block.ID
	part.VendorExtras["name"] = block.Name
	part.VendorExtras["input"] = block.Input

	if block.CacheControl != nil {
		part.VendorExtras["cache_control"] = block.CacheControl
	}

	return part
}

// convertWebSearchToolResultToContract 转换 Web 搜索工具结果块。
func convertWebSearchToolResultToContract(block *anthropicTypes.WebSearchToolResultBlockParam) *types.ContentPart {
	part := &types.ContentPart{
		Type: "web_search_tool_result",
	}

	part.VendorExtras = make(map[string]interface{})
	source := types.VendorSourceAnthropic
	part.VendorExtrasSource = &source

	part.VendorExtras["tool_use_id"] = block.ToolUseID
	part.VendorExtras["content"] = block.Content

	if block.CacheControl != nil {
		part.VendorExtras["cache_control"] = block.CacheControl
	}

	return part
}

// convertStreamMessageToContentBlocks 将 StreamMessagePayload 转换为 Anthropic ResponseContentBlock 列表。
//
// 用于 message_start 事件中的 message.content 构造。
func convertStreamMessageToContentBlocks(msg *types.StreamMessagePayload, log logger.Logger) ([]anthropicTypes.ResponseContentBlock, error) {
	var blocks []anthropicTypes.ResponseContentBlock

	// 转换 Parts 为 content blocks
	for _, part := range msg.Parts {
		switch part.Type {
		case "text":
			textBlock := &anthropicTypes.TextBlock{
				Type: anthropicTypes.ResponseContentBlockText,
				Text: part.Text,
			}
			// 转换 annotations 为 citations
			for _, ann := range part.Annotations {
				if citation, ok := ann.(anthropicTypes.TextCitation); ok {
					textBlock.Citations = append(textBlock.Citations, citation)
				} else {
					citeJSON, err := json.Marshal(ann)
					if err == nil {
						var cite anthropicTypes.TextCitation
						if err := json.Unmarshal(citeJSON, &cite); err == nil {
							textBlock.Citations = append(textBlock.Citations, cite)
						}
					}
				}
			}
			blocks = append(blocks, anthropicTypes.ResponseContentBlock{Text: textBlock})

		case "thinking":
			thinkingBlock := &anthropicTypes.ThinkingBlock{
				Type:     anthropicTypes.ResponseContentBlockThinking,
				Thinking: part.Text,
			}
			// 从 raw 中提取 signature
			if part.Raw != nil {
				if sig, ok := part.Raw["signature"].(string); ok {
					thinkingBlock.Signature = sig
				}
			}
			blocks = append(blocks, anthropicTypes.ResponseContentBlock{Thinking: thinkingBlock})

		case "redacted_thinking":
			redactedBlock := &anthropicTypes.RedactedThinkingBlock{
				Type: anthropicTypes.ResponseContentBlockRedactedThinking,
			}
			if part.Raw != nil {
				if data, ok := part.Raw["data"].(string); ok {
					redactedBlock.Data = data
				}
			}
			blocks = append(blocks, anthropicTypes.ResponseContentBlock{RedactedThinking: redactedBlock})

		case "web_search_tool_result":
			if part.Raw != nil {
				blockJSON, err := json.Marshal(part.Raw)
				if err == nil {
					var block anthropicTypes.WebSearchToolResultBlock
					if err := json.Unmarshal(blockJSON, &block); err == nil {
						blocks = append(blocks, anthropicTypes.ResponseContentBlock{WebSearchToolResult: &block})
					}
				}
			}

		default:
			// 其他类型尝试从 raw 恢复
			if part.Raw != nil {
				blockJSON, err := json.Marshal(part.Raw)
				if err == nil {
					var block anthropicTypes.ResponseContentBlock
					if err := json.Unmarshal(blockJSON, &block); err == nil {
						blocks = append(blocks, block)
					}
				}
			}
		}
	}

	// 转换 ToolCalls 为 tool_use/server_tool_use blocks
	for _, tool := range msg.ToolCalls {
		switch tool.Type {
		case "tool_use":
			toolBlock := &anthropicTypes.ToolUseBlock{
				Type: anthropicTypes.ResponseContentBlockToolUse,
				ID:   tool.ID,
				Name: tool.Name,
			}
			if tool.Arguments != "" {
				var input map[string]interface{}
				if err := json.Unmarshal([]byte(tool.Arguments), &input); err == nil {
					toolBlock.Input = input
				} else {
					log.Warn("解析工具参数失败", "error", err, "tool_id", tool.ID)
				}
			}
			blocks = append(blocks, anthropicTypes.ResponseContentBlock{ToolUse: toolBlock})

		case "server_tool_use":
			toolBlock := &anthropicTypes.ServerToolUseBlock{
				Type: anthropicTypes.ResponseContentBlockServerToolUse,
				ID:   tool.ID,
				Name: tool.Name,
			}
			if tool.Arguments != "" {
				var input map[string]interface{}
				if err := json.Unmarshal([]byte(tool.Arguments), &input); err == nil {
					toolBlock.Input = input
				} else {
					log.Warn("解析服务器工具参数失败", "error", err, "tool_id", tool.ID)
				}
			}
			blocks = append(blocks, anthropicTypes.ResponseContentBlock{ServerToolUse: toolBlock})
		}
	}

	return blocks, nil
}
