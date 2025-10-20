package converter

import (
	"encoding/json"

	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	coreTypes "github.com/MeowSalty/portal/types"
)

// StreamEventConverter Anthropic 流式事件转换器，维护转换状态
// 用于从核心响应格式转换为 Anthropic 的流式事件，自动处理内容块边界
// 通过缓存角色来检测消息边界，当角色发生变更时表示新消息开始
type StreamEventConverter struct {
	contentBlockStarted bool            // 文本内容块是否已开始
	toolCallsStarted    map[string]bool // 工具调用是否已开始（按 ID 追踪）
	currentIndex        int             // 当前内容块索引
	cachedRole          *string         // 缓存的角色，用于检测消息边界
	cachedFinishReason  *string         // 缓存的完成原因，等待 Usage 信息
}

// NewStreamEventConverter 创建新的流式事件转换器
func NewStreamEventConverter() *StreamEventConverter {
	return &StreamEventConverter{
		contentBlockStarted: false,
		toolCallsStarted:    make(map[string]bool),
		currentIndex:        0,
		cachedRole:          nil,
		cachedFinishReason:  nil,
	}
}

// Reset 重置转换器状态
// 当开始处理新的消息流时调用
func (c *StreamEventConverter) Reset() {
	c.contentBlockStarted = false
	c.toolCallsStarted = make(map[string]bool)
	c.currentIndex = 0
	c.cachedRole = nil
	c.cachedFinishReason = nil
}

// ParseStreamLine 解析流式响应行
func ParseStreamLine(line []byte) (*anthropicTypes.StreamEvent, error) {
	var event anthropicTypes.StreamEvent
	if err := json.Unmarshal(line, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// ConvertResponse 将核心响应转换为 Anthropic 响应（非流式）
func ConvertResponse(coreResp *coreTypes.Response) *anthropicTypes.Response {
	anthropicResp := &anthropicTypes.Response{
		ID:    coreResp.ID,
		Type:  "message",
		Model: coreResp.Model,
	}

	// 转换第一个 choice（Anthropic 只支持单个响应）
	if len(coreResp.Choices) > 0 {
		choice := coreResp.Choices[0]

		// 设置角色
		if choice.Message != nil {
			anthropicResp.Role = choice.Message.Role

			// 转换消息内容
			if choice.Message.Content != nil {
				anthropicResp.Content = []anthropicTypes.ResponseContent{
					{
						Type: "text",
						Text: choice.Message.Content,
					},
				}
			}

			// 转换工具调用
			if len(choice.Message.ToolCalls) > 0 {
				for _, toolCall := range choice.Message.ToolCalls {
					// 解析工具调用参数
					var input interface{}
					if toolCall.Function.Arguments != "" {
						json.Unmarshal([]byte(toolCall.Function.Arguments), &input)
					}

					anthropicResp.Content = append(anthropicResp.Content, anthropicTypes.ResponseContent{
						Type:  "tool_use",
						ID:    &toolCall.ID,
						Name:  &toolCall.Function.Name,
						Input: input,
					})
				}
			}
		} else {
			// 如果没有消息，设置默认角色
			anthropicResp.Role = "assistant"
			anthropicResp.Content = []anthropicTypes.ResponseContent{}
		}

		// 转换停止原因
		if choice.FinishReason != nil {
			anthropicResp.StopReason = convertFinishReason(*choice.FinishReason)
		}
	}

	// 转换使用统计
	if coreResp.Usage != nil {
		anthropicResp.Usage = &anthropicTypes.Usage{
			InputTokens:  coreResp.Usage.PromptTokens,
			OutputTokens: coreResp.Usage.CompletionTokens,
		}
	}

	return anthropicResp
}

// ConvertStreamEvent 将核心流式响应转换为 StreamEvent
// 用于将统一的核心响应格式转换回 Anthropic 的流式事件格式
// 注意：此函数不会自动生成 content_block_start 事件，仅用于 Anthropic 原生响应
// 如果需要从其他提供商（如 OpenAI）转换，请使用 StreamEventConverter
func ConvertStreamEvent(coreResp *coreTypes.Response) *anthropicTypes.StreamEvent {
	event := &anthropicTypes.StreamEvent{}

	// 根据响应的内容判断事件类型
	if len(coreResp.Choices) > 0 {
		choice := coreResp.Choices[0]

		// 如果有 Delta，说明是流式响应
		if choice.Delta != nil {
			// 如果有角色信息，这是 message_start 事件
			if choice.Delta.Role != nil {
				event.Type = "message_start"
				event.Message = &anthropicTypes.Response{
					ID:    coreResp.ID,
					Type:  "message",
					Model: coreResp.Model,
					Role:  *choice.Delta.Role,
				}
			} else if choice.Delta.Content != nil {
				// 如果有内容，这是 content_block_delta 事件
				event.Type = "content_block_delta"
				event.Delta = &anthropicTypes.Delta{
					Type: "text_delta",
					Text: choice.Delta.Content,
				}
			} else if len(choice.Delta.ToolCalls) > 0 {
				// 如果有工具调用，根据内容判断是开始还是增量
				toolCall := choice.Delta.ToolCalls[0]
				if toolCall.Function.Name != "" {
					// 工具使用块开始
					event.Type = "content_block_start"
					event.ContentBlock = &anthropicTypes.ResponseContent{
						Type: "tool_use",
						ID:   &toolCall.ID,
						Name: &toolCall.Function.Name,
					}
				} else if toolCall.Function.Arguments != "" {
					// 工具输入增量
					event.Type = "content_block_delta"
					event.Delta = &anthropicTypes.Delta{
						Type:        "input_json_delta",
						PartialJSON: &toolCall.Function.Arguments,
					}
				}
			}
		}

		// 如果有完成原因，这是 message_delta 事件
		if choice.FinishReason != nil {
			event.Type = "message_delta"
			anthropicStopReason := convertFinishReason(*choice.FinishReason)
			event.Delta = &anthropicTypes.Delta{
				StopReason: anthropicStopReason,
			}
		}
	}

	// 如果有使用统计，添加到事件中
	if coreResp.Usage != nil {
		event.Usage = &anthropicTypes.Usage{
			InputTokens:  coreResp.Usage.PromptTokens,
			OutputTokens: coreResp.Usage.CompletionTokens,
		}
	}

	return event
}

// ConvertStreamEvents 将核心流式响应转换为 Anthropic StreamEvent 列表
// 此方法会自动处理 content_block_start 和 content_block_stop 事件的生成
// 适用于从其他提供商（如 OpenAI）转换到 Anthropic 格式
// 通过检测角色变更来识别消息边界，角色变更时会自动生成新的 message_start 事件
func (c *StreamEventConverter) ConvertStreamEvents(coreResp *coreTypes.Response) []*anthropicTypes.StreamEvent {
	events := make([]*anthropicTypes.StreamEvent, 0)

	if len(coreResp.Choices) == 0 {
		return events
	}

	choice := coreResp.Choices[0]

	// 处理 Delta
	if choice.Delta != nil {
		// 1. 检测角色变更 - 判断是否为新消息
		var isNewMessage bool
		if choice.Delta.Role != nil && *choice.Delta.Role != "" {
			// 如果角色与缓存的角色不同，说明是新消息
			if c.cachedRole == nil || *c.cachedRole != *choice.Delta.Role {
				isNewMessage = true
				c.cachedRole = choice.Delta.Role
			}
		}

		// 2. 处理新消息 - message_start 事件
		if isNewMessage {
			// 如果有正在进行的内容块，先结束它
			if c.contentBlockStarted || len(c.toolCallsStarted) > 0 {
				indexCopy := c.currentIndex
				events = append(events, &anthropicTypes.StreamEvent{
					Type:  "content_block_stop",
					Index: &indexCopy,
				})
			}

			// 重置状态（除了角色缓存）
			c.contentBlockStarted = false
			c.toolCallsStarted = make(map[string]bool)
			c.currentIndex = 0

			// 生成 message_start 事件
			events = append(events, &anthropicTypes.StreamEvent{
				Type: "message_start",
				Message: &anthropicTypes.Response{
					ID:      coreResp.ID,
					Type:    "message",
					Model:   coreResp.Model,
					Role:    *choice.Delta.Role,
					Content: []anthropicTypes.ResponseContent{},
				},
			})
		}

		// 3. 处理文本内容
		if choice.Delta.Content != nil && *choice.Delta.Content != "" {
			// 如果内容块还没开始，先发送 content_block_start
			if !c.contentBlockStarted {
				emptyText := ""
				indexCopy := c.currentIndex
				events = append(events, &anthropicTypes.StreamEvent{
					Type:  "content_block_start",
					Index: &indexCopy,
					ContentBlock: &anthropicTypes.ResponseContent{
						Type: "text",
						Text: &emptyText,
					},
				})
				c.contentBlockStarted = true
			}

			// 发送 content_block_delta
			indexCopy := c.currentIndex
			events = append(events, &anthropicTypes.StreamEvent{
				Type:  "content_block_delta",
				Index: &indexCopy,
				Delta: &anthropicTypes.Delta{
					Type: "text_delta",
					Text: choice.Delta.Content,
				},
			})
		}

		// 4. 处理工具调用
		if len(choice.Delta.ToolCalls) > 0 {
			for _, toolCall := range choice.Delta.ToolCalls {
				// 工具调用开始
				if toolCall.Function.Name != "" && !c.toolCallsStarted[toolCall.ID] {
					// 如果之前有文本内容块，先结束它
					if c.contentBlockStarted {
						indexCopy := c.currentIndex
						events = append(events, &anthropicTypes.StreamEvent{
							Type:  "content_block_stop",
							Index: &indexCopy,
						})
						c.currentIndex++
						c.contentBlockStarted = false
					}

					indexCopy := c.currentIndex
					events = append(events, &anthropicTypes.StreamEvent{
						Type:  "content_block_start",
						Index: &indexCopy,
						ContentBlock: &anthropicTypes.ResponseContent{
							Type: "tool_use",
							ID:   &toolCall.ID,
							Name: &toolCall.Function.Name,
						},
					})
					c.toolCallsStarted[toolCall.ID] = true
				}

				// 工具调用参数增量
				if toolCall.Function.Arguments != "" {
					indexCopy := c.currentIndex
					events = append(events, &anthropicTypes.StreamEvent{
						Type:  "content_block_delta",
						Index: &indexCopy,
						Delta: &anthropicTypes.Delta{
							Type:        "input_json_delta",
							PartialJSON: &toolCall.Function.Arguments,
						},
					})
				}
			}
		}
	}

	// 5. 处理完成原因 - 缓存 FinishReason，等待 Usage 信息
	if choice.FinishReason != nil {
		// 结束当前内容块
		if c.contentBlockStarted || len(c.toolCallsStarted) > 0 {
			indexCopy := c.currentIndex
			events = append(events, &anthropicTypes.StreamEvent{
				Type:  "content_block_stop",
				Index: &indexCopy,
			})
		}

		// 缓存 FinishReason，等待 Usage 信息
		c.cachedFinishReason = choice.FinishReason
	}

	// 6. 处理使用统计 - 当收到 Usage 时发送 message_delta 和 message_stop
	if coreResp.Usage != nil && c.cachedFinishReason != nil {
		// 发送 message_delta 事件（包含 stop_reason 和 usage）
		anthropicStopReason := convertFinishReason(*c.cachedFinishReason)
		messageDeltaEvent := &anthropicTypes.StreamEvent{
			Type: "message_delta",
			Delta: &anthropicTypes.Delta{
				StopReason: anthropicStopReason,
			},
			Usage: &anthropicTypes.Usage{
				InputTokens:  coreResp.Usage.PromptTokens,
				OutputTokens: coreResp.Usage.CompletionTokens,
			},
		}
		events = append(events, messageDeltaEvent)

		// 发送 message_stop 事件
		events = append(events, &anthropicTypes.StreamEvent{
			Type: "message_stop",
		})

		// 清除缓存
		c.cachedFinishReason = nil
	}

	return events
}

// convertFinishReason 转换完成原因为 Anthropic 格式
func convertFinishReason(finishReason string) *string {
	// 映射核心的完成原因到 Anthropic 的停止原因
	// Anthropic 支持："end_turn", "max_tokens", "stop_sequence", "tool_use"
	var anthropicReason string

	switch finishReason {
	case "stop":
		anthropicReason = "end_turn"
	case "length":
		anthropicReason = "max_tokens"
	case "tool_calls":
		anthropicReason = "tool_use"
	case "content_filter":
		anthropicReason = "end_turn"
	default:
		anthropicReason = "end_turn"
	}

	return &anthropicReason
}
