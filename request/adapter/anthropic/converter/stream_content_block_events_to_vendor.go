package converter

import (
	"encoding/json"

	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertContractToContentBlockStart 转换 content_block_start 事件为 Anthropic 格式。
//
// 反向映射规则：
//   - index: ContentIndex
//   - content_block: 从 Content.Kind + Text/Tool + Raw 构建，优先使用 raw.content_block
func convertContractToContentBlockStart(event *types.StreamEventContract, log logger.Logger) (*anthropicTypes.StreamEvent, error) {
	var contentBlock anthropicTypes.ResponseContentBlock
	useRawBlock := false

	// 优先尝试从 raw 中恢复原始 content_block
	if event.Content != nil && event.Content.Raw != nil {
		if rawBlock, ok := event.Content.Raw["content_block"]; ok {
			blockJSON, err := json.Marshal(rawBlock)
			if err == nil {
				if err := json.Unmarshal(blockJSON, &contentBlock); err == nil {
					log.Debug("使用 content.raw.content_block 原文", "content_index", event.ContentIndex)
					useRawBlock = true
				} else {
					log.Warn("反序列化原始 content_block 失败", "error", err)
				}
			}
		}
	}

	// 若没有原始 block，则构造
	if !useRawBlock && event.Content != nil {
		switch event.Content.Kind {
		case "text":
			textBlock := &anthropicTypes.TextBlock{
				Type: anthropicTypes.ResponseContentBlockText,
			}
			if event.Content.Text != nil {
				textBlock.Text = *event.Content.Text
			}
			// 转换 annotations 为 citations
			if len(event.Content.Annotations) > 0 {
				annotations := make([]types.ResponseAnnotation, 0, len(event.Content.Annotations))
				for _, ann := range event.Content.Annotations {
					switch converted := ann.(type) {
					case types.ResponseAnnotation:
						annotations = append(annotations, converted)
					case *types.ResponseAnnotation:
						if converted != nil {
							annotations = append(annotations, *converted)
						}
					default:
						// 兼容历史 annotations 以 JSON 方式转换
						citeJSON, err := json.Marshal(ann)
						if err != nil {
							continue
						}
						var annotation types.ResponseAnnotation
						if err := json.Unmarshal(citeJSON, &annotation); err == nil {
							annotations = append(annotations, annotation)
						}
					}
				}
				if len(annotations) > 0 {
					citations, err := convertAnnotationsToCitations(annotations)
					if err != nil {
						return nil, err
					}
					textBlock.Citations = append(textBlock.Citations, citations...)
				}
			}
			contentBlock.Text = textBlock

		case "tool_use":
			if event.Content.Tool != nil {
				toolBlock := &anthropicTypes.ToolUseBlock{
					Type: anthropicTypes.ResponseContentBlockToolUse,
					ID:   event.Content.Tool.ID,
					Name: event.Content.Tool.Name,
				}
				// 解析 arguments JSON
				if event.Content.Tool.Arguments != "" {
					var input map[string]interface{}
					if err := json.Unmarshal([]byte(event.Content.Tool.Arguments), &input); err == nil {
						toolBlock.Input = input
					} else {
						log.Warn("解析工具参数失败", "error", err, "arguments", event.Content.Tool.Arguments)
					}
				}
				contentBlock.ToolUse = toolBlock
			}

		case "server_tool_use":
			if event.Content.Tool != nil {
				toolBlock := &anthropicTypes.ServerToolUseBlock{
					Type: anthropicTypes.ResponseContentBlockServerToolUse,
					ID:   event.Content.Tool.ID,
					Name: event.Content.Tool.Name,
				}
				if event.Content.Tool.Arguments != "" {
					var input map[string]interface{}
					if err := json.Unmarshal([]byte(event.Content.Tool.Arguments), &input); err == nil {
						toolBlock.Input = input
					} else {
						log.Warn("解析服务器工具参数失败", "error", err)
					}
				}
				contentBlock.ServerToolUse = toolBlock
			}

		case "thinking":
			if event.Content.Text != nil {
				thinkingBlock := &anthropicTypes.ThinkingBlock{
					Type:     anthropicTypes.ResponseContentBlockThinking,
					Thinking: *event.Content.Text,
				}
				// 从 raw 中提取 signature
				if event.Content.Raw != nil {
					if sig, ok := event.Content.Raw["signature"].(string); ok {
						thinkingBlock.Signature = sig
					}
				}
				contentBlock.Thinking = thinkingBlock
			}

		case "redacted_thinking":
			redactedBlock := &anthropicTypes.RedactedThinkingBlock{
				Type: anthropicTypes.ResponseContentBlockRedactedThinking,
			}
			// 从 raw 中提取 data
			if event.Content.Raw != nil {
				if data, ok := event.Content.Raw["data"].(string); ok {
					redactedBlock.Data = data
				}
			}
			contentBlock.RedactedThinking = redactedBlock

		case "web_search_tool_result":
			// 尝试从 raw 中恢复完整结构
			if event.Content.Raw != nil {
				resultJSON, err := json.Marshal(event.Content.Raw)
				if err == nil {
					var resultBlock anthropicTypes.WebSearchToolResultBlock
					if err := json.Unmarshal(resultJSON, &resultBlock); err == nil {
						contentBlock.WebSearchToolResult = &resultBlock
					}
				}
			}

		default:
			// 其他类型尝试从 raw 恢复
			if event.Content.Raw != nil {
				blockJSON, err := json.Marshal(event.Content.Raw)
				if err == nil {
					var block anthropicTypes.ResponseContentBlock
					if err := json.Unmarshal(blockJSON, &block); err == nil {
						contentBlock = block
					}
				}
			}
		}
	}

	startEvent := &anthropicTypes.ContentBlockStartEvent{
		Type:         anthropicTypes.StreamEventContentBlockStart,
		Index:        event.ContentIndex,
		ContentBlock: contentBlock,
	}

	log.Debug("转换 content_block_start 事件完成", "content_index", event.ContentIndex, "kind", event.Content.Kind)
	return &anthropicTypes.StreamEvent{ContentBlockStart: startEvent}, nil
}

// convertContractToContentBlockDelta 转换 content_block_delta 事件为 Anthropic 格式。
//
// 反向映射规则：
//   - index: ContentIndex
//   - delta: 根据 DeltaType 构造对应的增量类型
func convertContractToContentBlockDelta(event *types.StreamEventContract, log logger.Logger) (*anthropicTypes.StreamEvent, error) {
	delta := anthropicTypes.ContentBlockDelta{}

	if event.Delta != nil {
		switch event.Delta.DeltaType {
		case string(anthropicTypes.DeltaTypeText):
			if event.Delta.Text != nil {
				delta.Text = &anthropicTypes.TextDelta{
					Type: anthropicTypes.DeltaTypeText,
					Text: *event.Delta.Text,
				}
			}

		case string(anthropicTypes.DeltaTypeInputJSON):
			if event.Delta.PartialJSON != nil {
				delta.InputJSON = &anthropicTypes.InputJSONDelta{
					Type:        anthropicTypes.DeltaTypeInputJSON,
					PartialJSON: *event.Delta.PartialJSON,
				}
			}

		case string(anthropicTypes.DeltaTypeThinking):
			if event.Delta.Thinking != nil {
				delta.Thinking = &anthropicTypes.ThinkingDelta{
					Type:     anthropicTypes.DeltaTypeThinking,
					Thinking: *event.Delta.Thinking,
				}
			}

		case string(anthropicTypes.DeltaTypeSignature):
			if event.Delta.Signature != nil {
				delta.Signature = &anthropicTypes.SignatureDelta{
					Type:      anthropicTypes.DeltaTypeSignature,
					Signature: *event.Delta.Signature,
				}
			}

		case string(anthropicTypes.DeltaTypeCitations):
			if event.Delta.Citation != nil {
				// 尝试将 Citation 转换为 TextCitation
				citation, ok := event.Delta.Citation.(anthropicTypes.TextCitation)
				if ok {
					delta.Citations = &anthropicTypes.CitationsDelta{
						Type:     anthropicTypes.DeltaTypeCitations,
						Citation: citation,
					}
				} else {
					// 尝试通过 JSON 转换
					citeJSON, err := json.Marshal(event.Delta.Citation)
					if err == nil {
						var cite anthropicTypes.TextCitation
						if err := json.Unmarshal(citeJSON, &cite); err == nil {
							delta.Citations = &anthropicTypes.CitationsDelta{
								Type:     anthropicTypes.DeltaTypeCitations,
								Citation: cite,
							}
						}
					}
				}
			}

		default:
			log.Warn("不支持的 delta_type，尝试从 raw 恢复", "delta_type", event.Delta.DeltaType)
			// 尝试从 raw 恢复
			if event.Delta.Raw != nil {
				deltaJSON, err := json.Marshal(event.Delta.Raw)
				if err == nil {
					var deltaBlock anthropicTypes.ContentBlockDelta
					if err := json.Unmarshal(deltaJSON, &deltaBlock); err == nil {
						delta = deltaBlock
					}
				}
			}
		}
	}

	deltaEvent := &anthropicTypes.ContentBlockDeltaEvent{
		Type:  anthropicTypes.StreamEventContentBlockDelta,
		Index: event.ContentIndex,
		Delta: delta,
	}

	log.Debug("转换 content_block_delta 事件完成", "content_index", event.ContentIndex, "delta_type", event.Delta.DeltaType)
	return &anthropicTypes.StreamEvent{ContentBlockDelta: deltaEvent}, nil
}

// convertContractToContentBlockStop 转换 content_block_stop 事件为 Anthropic 格式。
func convertContractToContentBlockStop(event *types.StreamEventContract, log logger.Logger) (*anthropicTypes.StreamEvent, error) {
	stopEvent := &anthropicTypes.ContentBlockStopEvent{
		Type:  anthropicTypes.StreamEventContentBlockStop,
		Index: event.ContentIndex,
	}

	log.Debug("转换 content_block_stop 事件完成", "content_index", event.ContentIndex)
	return &anthropicTypes.StreamEvent{ContentBlockStop: stopEvent}, nil
}
