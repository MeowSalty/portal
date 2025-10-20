package types

import (
	"fmt"

	coreTypes "github.com/MeowSalty/portal/types"
)

// Response Anthropic API 响应结构
type Response struct {
	ID           string            `json:"id"`              // 消息 ID
	Type         string            `json:"type"`            // "message"
	Role         string            `json:"role"`            // "assistant"
	Content      []ResponseContent `json:"content"`         // 内容块数组
	Model        string            `json:"model"`           // 使用的模型
	StopReason   *string           `json:"stop_reason"`     // 停止原因
	StopSequence *string           `json:"stop_sequence"`   // 停止序列
	Usage        *Usage            `json:"usage,omitempty"` // 使用统计
}

// ResponseContent 响应内容块
type ResponseContent struct {
	Type     string                `json:"type"`               // "text", "tool_use", "thinking", "server_tool_use", "web_search_tool_result"
	Text     *string               `json:"text,omitempty"`     // 文本内容
	ID       *string               `json:"id,omitempty"`       // 工具使用 ID
	Name     *string               `json:"name,omitempty"`     // 工具名称
	Input    interface{}           `json:"input,omitempty"`    // 工具输入
	Thinking *string               `json:"thinking,omitempty"` // 思考内容
	Content  []WebSearchToolResult `json:"content,omitempty"`  // web search 结果
}

// Usage 使用统计
type Usage struct {
	InputTokens  int `json:"input_tokens"`  // 输入 token 数
	OutputTokens int `json:"output_tokens"` // 输出 token 数
}

// StreamEvent 流式事件
type StreamEvent struct {
	Type         string           `json:"type"` // 事件类型
	Message      *Response        `json:"message,omitempty"`
	Index        *int             `json:"index,omitempty"`
	ContentBlock *ResponseContent `json:"content_block,omitempty"`
	Delta        *Delta           `json:"delta,omitempty"`
	Usage        *Usage           `json:"usage,omitempty"`
	Error        *Error           `json:"error,omitempty"` // 错误事件
}

// Delta 增量更新
type Delta struct {
	Type         string  `json:"type,omitempty"`          // "text_delta", "input_json_delta", "thinking_delta", "signature_delta"
	Text         *string `json:"text,omitempty"`          // 文本增量
	PartialJSON  *string `json:"partial_json,omitempty"`  // 部分 JSON
	Thinking     *string `json:"thinking,omitempty"`      // 思考内容增量
	Signature    *string `json:"signature,omitempty"`     // 签名
	StopReason   *string `json:"stop_reason,omitempty"`   // 停止原因
	StopSequence *string `json:"stop_sequence,omitempty"` // 停止序列
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Type  string `json:"type"`  // "error"
	Error Error  `json:"error"` // 错误详情
}

// Error 错误详情
type Error struct {
	Type    string `json:"type"`    // 错误类型
	Message string `json:"message"` // 错误消息
}

// WebSearchToolResult 网络搜索工具结果
type WebSearchToolResult struct {
	Type             string `json:"type"`               // "web_search_result"
	Title            string `json:"title"`              // 结果标题
	URL              string `json:"url"`                // 结果链接
	EncryptedContent string `json:"encrypted_content"`  // 加密内容
	PageAge          *int   `json:"page_age,omitempty"` // 页面年龄
}

// ConvertCoreResponse 将 Anthropic 响应转换为核心响应
func (resp Response) ConvertCoreResponse() *coreTypes.Response {
	response := &coreTypes.Response{
		ID:      resp.ID,
		Model:   resp.Model,
		Choices: make([]coreTypes.Choice, 1),
	}

	// 转换内容
	var contentStr *string
	if resp.Content != nil {
		content := convertResponseContent(resp.Content)
		if content.StringValue != nil {
			contentStr = content.StringValue
		}
	}

	choice := coreTypes.Choice{
		Message: &coreTypes.ResponseMessage{
			Role:    resp.Role,
			Content: contentStr,
		},
	}

	// 设置停止原因
	if resp.StopReason != nil {
		choice.FinishReason = resp.StopReason
	}

	response.Choices[0] = choice

	// 转换使用统计
	if resp.Usage.InputTokens > 0 || resp.Usage.OutputTokens > 0 {
		response.Usage = &coreTypes.ResponseUsage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		}
	}

	return response
}

// convertResponseContent 转换响应内容
func convertResponseContent(contents []ResponseContent) coreTypes.MessageContent {
	// 如果只有一个文本块，返回字符串
	if len(contents) == 1 && contents[0].Type == "text" && contents[0].Text != nil {
		return coreTypes.MessageContent{
			StringValue: contents[0].Text,
		}
	}

	// 否则返回内容部分数组
	parts := make([]coreTypes.ContentPart, 0, len(contents))

	for _, content := range contents {
		part := coreTypes.ContentPart{
			Type: content.Type,
		}

		switch content.Type {
		case "text":
			if content.Text != nil {
				part.Text = content.Text
			}
		case "tool_use", "server_tool_use":
			// 转换工具使用（包括服务器工具）
			part.Type = "tool_use"
			// 工具使用的详细信息需要序列化到适当的格式
			// 这里简化处理，实际应用中可能需要更复杂的转换
		case "thinking":
			// 转换思考内容为文本
			if content.Thinking != nil {
				part.Type = "text"
				part.Text = content.Thinking
			}
		case "web_search_tool_result":
			// 转换 Web 搜索结果为文本格式
			if len(content.Content) > 0 {
				part.Type = "text"
				resultText := ""
				for i, result := range content.Content {
					if i > 0 {
						resultText += "\n\n"
					}
					resultText += result.Title + "\n" + result.URL
				}
				part.Text = &resultText
			}
		}

		parts = append(parts, part)
	}

	return coreTypes.MessageContent{
		ContentParts: parts,
	}
}

// ConvertCoreResponse 转换流式事件为核心响应
// 这个函数处理 Anthropic 的 StreamEvent 并将其转换为统一的核心响应格式
func (event *StreamEvent) ConvertCoreResponse() *coreTypes.Response {
	response := &coreTypes.Response{
		Object:  "chat.completion.chunk",
		Choices: make([]coreTypes.Choice, 1),
	}

	choice := coreTypes.Choice{}

	switch event.Type {
	case "message_start":
		// 消息开始事件 - 包含初始的消息元数据
		if event.Message != nil {
			response.ID = event.Message.ID
			response.Model = event.Message.Model
			choice.Delta = &coreTypes.Delta{
				Role: &event.Message.Role,
			}
		}

	case "content_block_start":
		// 内容块开始 - 开始一个新的内容块（文本、工具使用、思考等）
		if event.ContentBlock != nil {
			switch event.ContentBlock.Type {
			case "text":
				emptyText := ""
				choice.Delta = &coreTypes.Delta{
					Content: &emptyText,
				}
			case "tool_use", "server_tool_use":
				// 工具使用块开始（包括服务器工具）
				if event.ContentBlock.ID != nil && event.ContentBlock.Name != nil {
					choice.Delta = &coreTypes.Delta{
						ToolCalls: []coreTypes.ToolCall{
							{
								ID:   *event.ContentBlock.ID,
								Type: "function",
								Function: coreTypes.FunctionCall{
									Name:      *event.ContentBlock.Name,
									Arguments: "",
								},
							},
						},
					}
				}
			case "thinking":
				// 思考内容块开始
				emptyText := ""
				choice.Delta = &coreTypes.Delta{
					Content: &emptyText,
				}
			case "web_search_tool_result":
				// Web 搜索结果块 - 通常不需要特殊处理
			}
		}

	case "content_block_delta":
		// 内容块增量更新 - 接收内容的增量部分
		if event.Delta != nil {
			switch event.Delta.Type {
			case "text_delta":
				if event.Delta.Text != nil {
					choice.Delta = &coreTypes.Delta{
						Content: event.Delta.Text,
					}
				}
			case "input_json_delta":
				if event.Delta.PartialJSON != nil {
					// 工具输入的增量 JSON
					var toolID string
					if event.Index != nil {
						toolID = fmt.Sprintf("%d", *event.Index)
					}
					choice.Delta = &coreTypes.Delta{
						ToolCalls: []coreTypes.ToolCall{
							{
								ID:   toolID,
								Type: "function",
								Function: coreTypes.FunctionCall{
									Arguments: *event.Delta.PartialJSON,
								},
							},
						},
					}
				}
			case "thinking_delta":
				if event.Delta.Thinking != nil {
					// 思考内容增量
					choice.Delta = &coreTypes.Delta{
						Content: event.Delta.Thinking,
					}
				}
			case "signature_delta":
				// 签名增量 - 用于验证思考块的完整性
				// 通常不需要发送到客户端，可以在内部验证
			}
		}

	case "content_block_stop":
		// 内容块结束 - 一个内容块完成
		// 通常不需要特殊处理，可以发送空的 delta

	case "message_delta":
		// 消息增量更新 - 包含停止原因和使用统计的更新
		if event.Delta != nil {
			if event.Delta.StopReason != nil {
				// 将 Anthropic 的停止原因转换为核心格式
				finishReason := convertAnthropicStopReason(*event.Delta.StopReason)
				choice.FinishReason = &finishReason
			}
		}
		// 使用统计在这个事件中更新
		if event.Usage != nil {
			response.Usage = &coreTypes.ResponseUsage{
				PromptTokens:     event.Usage.InputTokens,
				CompletionTokens: event.Usage.OutputTokens,
				TotalTokens:      event.Usage.InputTokens + event.Usage.OutputTokens,
			}
		}

	case "message_stop":
		// 消息结束 - 流的最后一个事件
		// 通常在 message_delta 中已经设置了 finish_reason
		// 如果没有设置，使用默认值
		if choice.FinishReason == nil {
			finishReason := "stop"
			choice.FinishReason = &finishReason
		}

	case "ping":
		// ping 事件 - 保持连接活跃，不需要特殊处理
		// 返回空的响应即可

	case "error":
		// 错误事件 - 从事件中提取错误信息
		errorMsg := "流式响应错误"
		errorType := "unknown"
		if event.Error != nil {
			errorMsg = event.Error.Message
			errorType = event.Error.Type
		}
		choice.Error = &coreTypes.ErrorResponse{
			Code:    500,
			Message: errorMsg,
			Metadata: map[string]interface{}{
				"provider":   "anthropic",
				"error_type": errorType,
			},
		}
	}

	response.Choices[0] = choice
	return response
}

// convertAnthropicStopReason 将 Anthropic 的停止原因转换为核心格式
func convertAnthropicStopReason(anthropicReason string) string {
	// Anthropic 停止原因："end_turn", "max_tokens", "stop_sequence", "tool_use"
	// 核心格式："stop", "length", "tool_calls", "content_filter"
	switch anthropicReason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "tool_use":
		return "tool_calls"
	case "stop_sequence":
		return "stop"
	default:
		return "stop"
	}
}
