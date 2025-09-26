package converter

import (
	"github.com/MeowSalty/portal/adapter/openai/types"
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

		// 转换消息内容
		if choice.Delta != nil {
			coreChoice.Message = &coreTypes.ResponseMessage{
				Role: choice.Delta.Role,
			}

			// 转换消息内容
			if choice.Delta.Content != nil {
				content := *choice.Delta.Content
				coreChoice.Message.Content = &content
			}

			// 转换工具调用
			if len(choice.Delta.ToolCalls) > 0 {
				coreChoice.Message.ToolCalls = make([]coreTypes.ToolCall, len(choice.Delta.ToolCalls))
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

		// 只有当 choice.Message 不为 nil 时才处理消息相关内容
		if choice.Message != nil {
			openaiChoice.Delta = &types.Delta{
				Role: choice.Message.Role,
			}

			// 转换消息内容
			if choice.Message.Content != nil {
				content := *choice.Message.Content
				openaiChoice.Delta.Content = &content
			}

			// 转换工具调用
			if len(choice.Message.ToolCalls) > 0 {
				openaiChoice.Delta.ToolCalls = make([]types.ToolCall, len(choice.Message.ToolCalls))
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
					openaiChoice.Delta.ToolCalls[j] = openaiToolCall
				}
			}
		}

		result.Choices[i] = openaiChoice
	}

	return result
}
