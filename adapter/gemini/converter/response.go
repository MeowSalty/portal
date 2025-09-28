package converter

import (
	"github.com/MeowSalty/portal/adapter/gemini/types"
	coreTypes "github.com/MeowSalty/portal/types"
)

// ConvertCoreResponse 将 Gemini 响应转换为核心响应
func ConvertCoreResponse(geminiResp *types.Response) *coreTypes.Response {
	if geminiResp == nil {
		return nil
	}

	result := &coreTypes.Response{
		ID:      geminiResp.ResponseID,
		Model:   geminiResp.ModelVersion,
		Object:  "chat.completion",
		Choices: make([]coreTypes.Choice, len(geminiResp.Candidates)),
	}

	// 转换使用情况
	if geminiResp.UsageMetadata != nil {
		result.Usage = &coreTypes.ResponseUsage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		}
	}

	// 转换每个候选
	for i, candidate := range geminiResp.Candidates {
		coreChoice := coreTypes.Choice{}

		// 转换完成原因
		if candidate.FinishReason != "" {
			coreChoice.FinishReason = &candidate.FinishReason
		}

		// 转换消息内容
		if len(candidate.Content.Parts) > 0 {
			coreChoice.Message = &coreTypes.ResponseMessage{
				Role: "assistant",
			}

			// 获取第一个部分的文本内容
			for _, part := range candidate.Content.Parts {
				if part.Text != nil {
					content := *part.Text
					coreChoice.Message.Content = &content
					break
				}
			}

			// 检查函数调用
			for _, part := range candidate.Content.Parts {
				if part.FunctionCall != nil {
					// 处理函数调用
					if coreChoice.Message.ToolCalls == nil {
						coreChoice.Message.ToolCalls = []coreTypes.ToolCall{}
					}

					toolCall := coreTypes.ToolCall{
						ID:   "", // Gemini 不提供工具调用 ID
						Type: "function",
						Function: coreTypes.FunctionCall{
							Name:      part.FunctionCall.Name,
							Arguments: convertArgsToString(part.FunctionCall.Args),
						},
					}
					coreChoice.Message.ToolCalls = append(coreChoice.Message.ToolCalls, toolCall)
				}
			}
		}

		result.Choices[i] = coreChoice
	}

	return result
}

// convertArgsToString 将参数 map 转换为 JSON 字符串
func convertArgsToString(args map[string]interface{}) string {
	if args == nil {
		return "{}"
	}

	// 在实际实现中，这里应该使用 JSON 序列化
	// 这里简化为返回空对象
	return "{}"
}

// ConvertResponse 将核心响应转换为 Gemini 响应（反向转换）
func ConvertResponse(resp *coreTypes.Response) *types.Response {
	if resp == nil {
		return nil
	}

	result := &types.Response{
		ResponseID:   resp.ID,
		ModelVersion: resp.Model,
		Candidates:   make([]types.Candidate, len(resp.Choices)),
	}

	// 转换使用情况
	if resp.Usage != nil {
		result.UsageMetadata = &types.UsageMetadata{
			PromptTokenCount:     resp.Usage.PromptTokens,
			CandidatesTokenCount: resp.Usage.CompletionTokens,
			TotalTokenCount:      resp.Usage.TotalTokens,
		}
	}

	// 转换每个选择
	for i, choice := range resp.Choices {
		candidate := types.Candidate{
			Index: i,
		}

		// 转换完成原因
		if choice.FinishReason != nil {
			candidate.FinishReason = *choice.FinishReason
		}

		// 转换消息内容
		if choice.Message != nil {
			candidate.Content = types.Content{
				Role:  "model",
				Parts: []types.Part{},
			}

			// 转换文本内容
			if choice.Message.Content != nil {
				text := *choice.Message.Content
				candidate.Content.Parts = append(candidate.Content.Parts, types.Part{
					Text: &text,
				})
			}

			// 转换工具调用
			if len(choice.Message.ToolCalls) > 0 {
				for _, toolCall := range choice.Message.ToolCalls {
					if toolCall.Function.Name != "" {
						candidate.Content.Parts = append(candidate.Content.Parts, types.Part{
							FunctionCall: &types.FunctionCall{
								Name: toolCall.Function.Name,
								Args: map[string]interface{}{},
							},
						})
					}
				}
			}
		}

		result.Candidates[i] = candidate
	}

	return result
}
