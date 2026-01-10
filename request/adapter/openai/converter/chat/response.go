package chat

import (
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ConvertCoreResponse 将 OpenAI 响应转换为核心响应
func ConvertCoreResponse(openaiResp *openaiChat.Response) *coreTypes.Response {
	if openaiResp == nil {
		return nil
	}

	result := &coreTypes.Response{
		ID:      openaiResp.ID,
		Model:   openaiResp.Model,
		Object:  openaiResp.Object,
		Choices: make([]coreTypes.Choice, len(openaiResp.Choices)),
		Created: openaiResp.Created,
	}

	// 转换使用情况
	if openaiResp.Usage != nil {
		result.Usage = &coreTypes.ResponseUsage{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
		}
	}

	// 转换每个选择
	for i, choice := range openaiResp.Choices {
		coreChoice := coreTypes.Choice{
			FinishReason: choice.FinishReason,
		}

		// 判断是流式响应还是非流式响应
		if choice.Delta != nil {
			// 流式响应：数据应放在 Delta 中
			coreChoice.Delta = &coreTypes.Delta{
				Role: choice.Delta.Role,
			}

			// 转换消息内容
			if choice.Delta.Content != nil {
				content := *choice.Delta.Content
				coreChoice.Delta.Content = &content
			}

			// 转换工具调用
			if len(choice.Delta.ToolCalls) > 0 {
				coreChoice.Delta.ToolCalls = make([]coreTypes.ToolCall, len(choice.Delta.ToolCalls))
				for j, toolCall := range choice.Delta.ToolCalls {
					coreToolCall := coreTypes.ToolCall{}

					// 转换 ID（指针转字符串）
					if toolCall.ID != nil {
						coreToolCall.ID = *toolCall.ID
					}

					// 转换 Type（指针转字符串）
					if toolCall.Type != nil {
						coreToolCall.Type = *toolCall.Type
					}

					// 转换函数调用
					if toolCall.Function != nil {
						if toolCall.Function.Name != nil {
							coreToolCall.Function.Name = *toolCall.Function.Name
						}
						if toolCall.Function.Arguments != nil {
							coreToolCall.Function.Arguments = *toolCall.Function.Arguments
						}
					}
					coreChoice.Delta.ToolCalls[j] = coreToolCall
				}
			}

			// 传递额外字段和来源格式
			if choice.Delta.ExtraFields != nil {
				coreChoice.Delta.ExtraFields = choice.Delta.ExtraFields
			}
			coreChoice.Delta.ExtraFieldsFormat = "openai"
		} else if choice.Message != nil {
			// 非流式响应：数据应放在 Message 中
			coreChoice.Message = &coreTypes.ResponseMessage{
				Role: choice.Message.Role,
			}

			// 转换消息内容
			if choice.Message.Content != nil {
				content := *choice.Message.Content
				coreChoice.Message.Content = &content
			}

			// 转换工具调用
			if len(choice.Message.ToolCalls) > 0 {
				coreChoice.Message.ToolCalls = make([]coreTypes.ToolCall, len(choice.Message.ToolCalls))
				for j, toolCall := range choice.Message.ToolCalls {
					coreToolCall := coreTypes.ToolCall{}

					// 转换 ID（指针转字符串）
					if toolCall.ID != nil {
						coreToolCall.ID = *toolCall.ID
					}

					// 转换 Type（指针转字符串）
					if toolCall.Type != nil {
						coreToolCall.Type = *toolCall.Type
					}

					// 转换函数调用
					if toolCall.Function != nil {
						if toolCall.Function.Name != nil {
							coreToolCall.Function.Name = *toolCall.Function.Name
						}
						if toolCall.Function.Arguments != nil {
							coreToolCall.Function.Arguments = *toolCall.Function.Arguments
						}
					}
					coreChoice.Message.ToolCalls[j] = coreToolCall
				}
			}

			// 传递额外字段和来源格式
			if choice.Message.ExtraFields != nil {
				coreChoice.Message.ExtraFields = choice.Message.ExtraFields
			}
			coreChoice.Message.ExtraFieldsFormat = "openai"
		}

		result.Choices[i] = coreChoice
	}

	return result
}

// ConvertResponse 将核心响应转换为 OpenAI 响应
func ConvertResponse(resp *coreTypes.Response) *openaiChat.Response {
	if resp == nil {
		return nil
	}

	result := &openaiChat.Response{
		ID:      resp.ID,
		Model:   resp.Model,
		Object:  resp.Object,
		Created: resp.Created,
		Choices: make([]openaiChat.Choice, len(resp.Choices)),
	}

	// 转换使用情况
	if resp.Usage != nil {
		result.Usage = &openaiChat.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	// 转换每个选择
	for i, choice := range resp.Choices {
		openaiChoice := openaiChat.Choice{
			FinishReason: choice.FinishReason,
			Index:        i,
		}

		// 判断是流式响应还是非流式响应
		if choice.Delta != nil {
			// 流式响应：数据在 Delta 中
			openaiChoice.Delta = &openaiChat.Delta{
				Role: choice.Delta.Role,
			}

			// 转换消息内容
			if choice.Delta.Content != nil {
				content := *choice.Delta.Content
				openaiChoice.Delta.Content = &content
			}

			// 转换工具调用
			if len(choice.Delta.ToolCalls) > 0 {
				openaiChoice.Delta.ToolCalls = make([]openaiChat.ToolCall, len(choice.Delta.ToolCalls))
				for j, toolCall := range choice.Delta.ToolCalls {
					openaiToolCall := openaiChat.ToolCall{}

					// 转换 ID（字符串转指针）
					if toolCall.ID != "" {
						id := toolCall.ID
						openaiToolCall.ID = &id
					}

					// 转换 Type（字符串转指针）
					if toolCall.Type != "" {
						typ := toolCall.Type
						openaiToolCall.Type = &typ
					}

					// 转换函数调用
					if toolCall.Function.Name != "" || toolCall.Function.Arguments != "" {
						openaiToolCall.Function = &openaiChat.ToolCallFunction{}
						if toolCall.Function.Name != "" {
							name := toolCall.Function.Name
							openaiToolCall.Function.Name = &name
						}
						if toolCall.Function.Arguments != "" {
							args := toolCall.Function.Arguments
							openaiToolCall.Function.Arguments = &args
						}
					}
					openaiChoice.Delta.ToolCalls[j] = openaiToolCall
				}
			}
		} else if choice.Message != nil {
			// 非流式响应：数据在 Message 中
			openaiChoice.Message = &openaiChat.Message{
				Role: choice.Message.Role,
			}

			// 转换消息内容
			if choice.Message.Content != nil {
				content := *choice.Message.Content
				openaiChoice.Message.Content = &content
			}

			// 转换工具调用
			if len(choice.Message.ToolCalls) > 0 {
				openaiChoice.Message.ToolCalls = make([]openaiChat.ToolCall, len(choice.Message.ToolCalls))
				for j, toolCall := range choice.Message.ToolCalls {
					openaiToolCall := openaiChat.ToolCall{}

					// 转换 ID（字符串转指针）
					if toolCall.ID != "" {
						id := toolCall.ID
						openaiToolCall.ID = &id
					}

					// 转换 Type（字符串转指针）
					if toolCall.Type != "" {
						typ := toolCall.Type
						openaiToolCall.Type = &typ
					}

					// 转换函数调用
					if toolCall.Function.Name != "" || toolCall.Function.Arguments != "" {
						openaiToolCall.Function = &openaiChat.ToolCallFunction{}
						if toolCall.Function.Name != "" {
							name := toolCall.Function.Name
							openaiToolCall.Function.Name = &name
						}
						if toolCall.Function.Arguments != "" {
							args := toolCall.Function.Arguments
							openaiToolCall.Function.Arguments = &args
						}
					}
					openaiChoice.Message.ToolCalls[j] = openaiToolCall
				}
			}
		}

		// 传递额外字段
		if choice.Delta != nil && choice.Delta.ExtraFields != nil {
			openaiChoice.Delta.ExtraFields = choice.Delta.ExtraFields
		} else if choice.Message != nil && choice.Message.ExtraFields != nil {
			openaiChoice.Message.ExtraFields = choice.Message.ExtraFields
		}

		result.Choices[i] = openaiChoice
	}

	return result
}
