package converter

import (
	"github.com/MeowSalty/portal/request/adapter/openai/types"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ConvertCoreResponse 将 OpenAI 响应转换为核心响应
func ConvertCoreResponse(openaiResp *types.Response) *coreTypes.Response {
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
					coreToolCall := coreTypes.ToolCall{
						ID:   toolCall.ID,
						Type: toolCall.Type,
					}
					// 转换函数调用
					if toolCall.Function != nil {
						coreToolCall.Function = coreTypes.FunctionCall{
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						}
					}
					coreChoice.Delta.ToolCalls[j] = coreToolCall
				}
			}
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
					coreToolCall := coreTypes.ToolCall{
						ID:   toolCall.ID,
						Type: toolCall.Type,
					}
					// 转换函数调用
					if toolCall.Function != nil {
						coreToolCall.Function = coreTypes.FunctionCall{
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						}
					}
					coreChoice.Message.ToolCalls[j] = coreToolCall
				}
			}
		}

		result.Choices[i] = coreChoice
	}

	return result
}

// ConvertResponse 将核心响应转换为 OpenAI 响应
func ConvertResponse(resp *coreTypes.Response) *types.Response {
	if resp == nil {
		return nil
	}

	result := &types.Response{
		ID:      resp.ID,
		Model:   resp.Model,
		Object:  resp.Object,
		Created: resp.Created,
		Choices: make([]types.Choice, len(resp.Choices)),
	}

	// 转换使用情况
	if resp.Usage != nil {
		result.Usage = &types.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	// 转换每个选择
	for i, choice := range resp.Choices {
		openaiChoice := types.Choice{
			FinishReason: choice.FinishReason,
			Index:        i,
		}

		// 判断是流式响应还是非流式响应
		if choice.Delta != nil {
			// 流式响应：数据在 Delta 中
			openaiChoice.Delta = &types.Delta{
				Role: choice.Delta.Role,
			}

			// 转换消息内容
			if choice.Delta.Content != nil {
				content := *choice.Delta.Content
				openaiChoice.Delta.Content = &content
			}

			// 转换工具调用
			if len(choice.Delta.ToolCalls) > 0 {
				openaiChoice.Delta.ToolCalls = make([]types.ToolCall, len(choice.Delta.ToolCalls))
				for j, toolCall := range choice.Delta.ToolCalls {
					openaiToolCall := types.ToolCall{
						ID:   toolCall.ID,
						Type: toolCall.Type,
					}
					if toolCall.Function.Name != "" {
						openaiToolCall.Function = &types.ToolCallFunction{
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						}
					}
					openaiChoice.Delta.ToolCalls[j] = openaiToolCall
				}
			}
		} else if choice.Message != nil {
			// 非流式响应：数据在 Message 中
			openaiChoice.Message = &types.Message{
				Role: choice.Message.Role,
			}

			// 转换消息内容
			if choice.Message.Content != nil {
				content := *choice.Message.Content
				openaiChoice.Message.Content = &content
			}

			// 转换工具调用
			if len(choice.Message.ToolCalls) > 0 {
				openaiChoice.Message.ToolCalls = make([]types.ToolCall, len(choice.Message.ToolCalls))
				for j, toolCall := range choice.Message.ToolCalls {
					openaiToolCall := types.ToolCall{
						ID:   toolCall.ID,
						Type: toolCall.Type,
					}
					if toolCall.Function.Name != "" {
						openaiToolCall.Function = &types.ToolCallFunction{
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						}
					}
					openaiChoice.Message.ToolCalls[j] = openaiToolCall
				}
			}
		}

		result.Choices[i] = openaiChoice
	}

	return result
}
