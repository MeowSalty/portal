package responses

import (
	"encoding/json"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/request/adapter/openai/converter/responses/helper"
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ResponseFromContract 将 ResponseContract 转换回 OpenAI Responses 响应。
// 优先使用 contract.Extras["openai.responses.output"] 直接恢复 output。
func ResponseFromContract(contract *types.ResponseContract, log logger.Logger) (*responsesTypes.Response, error) {
	if contract == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	resp := &responsesTypes.Response{
		ID:                contract.ID,
		Object:            "response",
		ParallelToolCalls: false,
		Metadata:          make(map[string]string),
		Output:            []responsesTypes.OutputItem{},
	}

	// 转换 Model
	if contract.Model != nil {
		resp.Model = *contract.Model
	}

	// 转换 CreatedAt
	if contract.CreatedAt != nil {
		resp.CreatedAt = *contract.CreatedAt
	}

	// 转换 Status 和 CompletedAt
	resp.Status = contract.Status
	resp.CompletedAt = contract.CompletedAt

	// 转换 Error
	if contract.Error != nil {
		resp.Error = helper.ConvertErrorFromContract(contract.Error)
	}

	// 转换 Usage
	if contract.Usage != nil {
		resp.Usage = helper.ConvertUsageFromContract(contract.Usage)
	}

	// 优先使用 Extras 中保存的原始 output
	if outputJSON, ok := contract.Extras["openai.responses.output"].(string); ok {
		var output []responsesTypes.OutputItem
		if err := json.Unmarshal([]byte(outputJSON), &output); err != nil {
			log.Warn("反序列化 output 失败，将从 Choices 重建", "error", err)
		} else {
			resp.Output = output
			// 从 Extras 恢复其他字段
			restoreExtrasFields(contract, resp)
			return resp, nil
		}
	}

	// 若缺失原始 output，则从 Choices 重建
	if len(contract.Choices) > 0 {
		output, err := convertChoicesToOutput(contract.Choices, log)
		if err != nil {
			return nil, errors.Wrap(errors.ErrCodeInternal, "从 Choices 重建 Output 失败", err)
		}
		resp.Output = output
	}

	// 从 Extras 恢复其他字段
	restoreExtrasFields(contract, resp)

	return resp, nil
}

// convertChoicesToOutput 从 Choices 重建 Output 数组。
func convertChoicesToOutput(choices []types.ResponseChoice, log logger.Logger) ([]responsesTypes.OutputItem, error) {
	if len(choices) == 0 {
		return nil, nil
	}

	// Responses API 通常只有一个 choice
	choice := choices[0]
	if choice.Message == nil {
		return nil, nil
	}

	var output []responsesTypes.OutputItem

	// 从 Message 重建 OutputItemMessage
	if len(choice.Message.Parts) > 0 || choice.Message.Content != nil || choice.Message.Refusal != nil {
		msgItem, err := convertMessageToOutputItem(choice.Message, log)
		if err != nil {
			return nil, err
		}
		if msgItem != nil {
			output = append(output, *msgItem)
		}
	}

	// 从 ToolCalls 重建工具调用输出项
	for _, toolCall := range choice.Message.ToolCalls {
		item, err := convertToolCallToOutputItem(&toolCall)
		if err != nil {
			log.Warn("转换 ToolCall 失败", "error", err)
			continue
		}
		if item != nil {
			output = append(output, *item)
		}
	}

	return output, nil
}

// convertMessageToOutputItem 将 Message 转换为 OutputItemMessage。
func convertMessageToOutputItem(message *types.ResponseMessage, log logger.Logger) (*responsesTypes.OutputItem, error) {
	outputMsg := responsesTypes.OutputMessage{
		Role:    "assistant",
		Status:  "completed",
		Content: []responsesTypes.OutputMessageContent{},
	}

	if message.Role != nil {
		outputMsg.Role = *message.Role
	}

	// 从 Parts 重建 Content
	for _, part := range message.Parts {
		content, err := convertPartToMessageContent(&part)
		if err != nil {
			log.Warn("转换 Part 失败", "error", err)
			continue
		}
		if content != nil {
			outputMsg.Content = append(outputMsg.Content, *content)
		}
	}

	// 如果有 Refusal，添加 refusal content
	if message.Refusal != nil {
		outputMsg.Content = append(outputMsg.Content, responsesTypes.OutputMessageContent{
			Refusal: &responsesTypes.RefusalContent{
				Type:    responsesTypes.OutputMessageContentTypeRefusal,
				Refusal: *message.Refusal,
			},
		})
	}

	itemID := "msg_default"
	if message.ID != nil {
		itemID = *message.ID
	}

	outputMsg.Type = responsesTypes.OutputItemTypeMessage
	outputMsg.ID = itemID
	return &responsesTypes.OutputItem{
		Message: &outputMsg,
	}, nil
}

// convertPartToMessageContent 将 Part 转换为 OutputMessageContent。
func convertPartToMessageContent(part *types.ResponseContentPart) (*responsesTypes.OutputMessageContent, error) {
	switch part.Type {
	case "output_text", "text":
		if part.Text == nil {
			return nil, nil
		}

		textContent := responsesTypes.OutputTextContent{
			Type: responsesTypes.OutputMessageContentTypeOutputText,
			Text: *part.Text,
		}

		// 转换 Annotations
		if len(part.Annotations) > 0 {
			textContent.Annotations = helper.ConvertAnnotationsFromContract(part.Annotations)
		}

		// 从 Extras 恢复 Logprobs
		if logprobs, ok := part.Extras["openai.responses.logprobs"]; ok {
			if lp, ok := logprobs.([]responsesTypes.LogProb); ok {
				textContent.Logprobs = lp
			}
		}

		return &responsesTypes.OutputMessageContent{
			OutputText: &textContent,
		}, nil

	case "refusal":
		if part.Text == nil {
			return nil, nil
		}

		refusalContent := responsesTypes.RefusalContent{
			Type:    responsesTypes.OutputMessageContentTypeRefusal,
			Refusal: *part.Text,
		}

		return &responsesTypes.OutputMessageContent{
			Refusal: &refusalContent,
		}, nil

	case "thinking":
		// ReasoningItem 从 Extras 恢复
		if _, ok := part.Extras["openai.responses.reasoning"]; ok {
			// 这里不返回 OutputMessageContent，因为 reasoning 是独立的 OutputItem
			return nil, nil
		}
		return nil, nil

	default:
		return nil, nil
	}
}

// convertToolCallToOutputItem 将 ToolCall 转换为 OutputItem。
func convertToolCallToOutputItem(toolCall *types.ResponseToolCall) (*responsesTypes.OutputItem, error) {
	if toolCall.Type == nil {
		return nil, nil
	}

	itemID := "call_default"
	if toolCall.ID != nil {
		itemID = *toolCall.ID
	}

	// 从 Extras 恢复 item_id
	if id, ok := toolCall.Extras["openai.responses.item_id"].(string); ok {
		itemID = id
	}

	switch *toolCall.Type {
	case "function":
		call := responsesTypes.FunctionToolCall{
			Status: "completed",
		}

		if toolCall.ID != nil {
			call.CallID = *toolCall.ID
		}
		if toolCall.Name != nil {
			call.Name = *toolCall.Name
		}
		if toolCall.Arguments != nil {
			call.Arguments = *toolCall.Arguments
		}

		// 从 Extras 恢复 status
		if status, ok := toolCall.Extras["openai.responses.status"].(string); ok {
			call.Status = status
		}

		call.Type = string(responsesTypes.OutputItemTypeFunctionCall)
		call.ID = itemID
		return &responsesTypes.OutputItem{
			FunctionCall: &call,
		}, nil

	case "web_search_call":
		if webSearch, ok := toolCall.Extras["openai.responses.web_search"]; ok {
			if call, ok := webSearch.(responsesTypes.WebSearchToolCall); ok {
				call.Type = string(responsesTypes.OutputItemTypeWebSearchCall)
				call.ID = itemID
				return &responsesTypes.OutputItem{
					WebSearchCall: &call,
				}, nil
			}
		}

	case "file_search_call":
		if fileSearch, ok := toolCall.Extras["openai.responses.file_search"]; ok {
			if call, ok := fileSearch.(responsesTypes.FileSearchToolCall); ok {
				call.Type = string(responsesTypes.OutputItemTypeFileSearchCall)
				call.ID = itemID
				return &responsesTypes.OutputItem{
					FileSearchCall: &call,
				}, nil
			}
		}

	case "computer_call":
		if computer, ok := toolCall.Extras["openai.responses.computer"]; ok {
			if call, ok := computer.(responsesTypes.ComputerToolCall); ok {
				call.Type = string(responsesTypes.OutputItemTypeComputerCall)
				call.ID = itemID
				return &responsesTypes.OutputItem{
					ComputerCall: &call,
				}, nil
			}
		}

	case "code_interpreter_call":
		if codeInterpreter, ok := toolCall.Extras["openai.responses.code_interpreter"]; ok {
			if call, ok := codeInterpreter.(responsesTypes.CodeInterpreterToolCall); ok {
				call.Type = string(responsesTypes.OutputItemTypeCodeInterpreterCall)
				call.ID = itemID
				return &responsesTypes.OutputItem{
					CodeInterpreterCall: &call,
				}, nil
			}
		}

	case "image_generation_call":
		if imageGen, ok := toolCall.Extras["openai.responses.image_gen"]; ok {
			if call, ok := imageGen.(responsesTypes.ImageGenToolCall); ok {
				call.Type = string(responsesTypes.OutputItemTypeImageGenCall)
				call.ID = itemID
				return &responsesTypes.OutputItem{
					ImageGenCall: &call,
				}, nil
			}
		}

	case "local_shell_call":
		if localShell, ok := toolCall.Extras["openai.responses.local_shell"]; ok {
			if call, ok := localShell.(responsesTypes.LocalShellToolCall); ok {
				call.Type = string(responsesTypes.OutputItemTypeLocalShellCall)
				call.ID = itemID
				return &responsesTypes.OutputItem{
					LocalShellCall: &call,
				}, nil
			}
		}

	case "mcp_call":
		if mcp, ok := toolCall.Extras["openai.responses.mcp"]; ok {
			if call, ok := mcp.(responsesTypes.MCPToolCall); ok {
				call.Type = string(responsesTypes.OutputItemTypeMCPCall)
				call.ID = itemID
				return &responsesTypes.OutputItem{
					MCPCall: &call,
				}, nil
			}
		}

	case "custom_tool_call":
		if custom, ok := toolCall.Extras["openai.responses.custom"]; ok {
			if call, ok := custom.(responsesTypes.CustomToolCall); ok {
				call.Type = string(responsesTypes.OutputItemTypeCustomToolCall)
				call.ID = itemID
				return &responsesTypes.OutputItem{
					CustomToolCall: &call,
				}, nil
			}
		}
	}

	return nil, nil
}

// restoreExtrasFields 从 Extras 恢复其他字段。
func restoreExtrasFields(contract *types.ResponseContract, resp *responsesTypes.Response) {
	if val, ok := contract.Extras["openai.responses.parallel_tool_calls"].(bool); ok {
		resp.ParallelToolCalls = val
	}
	if val, ok := contract.Extras["openai.responses.metadata"].(map[string]string); ok {
		resp.Metadata = val
	}
	if val, ok := contract.Extras["openai.responses.tool_choice"]; ok {
		if toolChoice, ok := val.(*shared.ToolChoiceUnion); ok {
			resp.ToolChoice = toolChoice
		}
	}
	if val, ok := contract.Extras["openai.responses.tools"]; ok {
		if tools, ok := val.([]shared.ToolUnion); ok {
			resp.Tools = tools
		}
	}
	if val, ok := contract.Extras["openai.responses.instructions"]; ok {
		if instructions, ok := val.(*responsesTypes.ResponseInstructions); ok {
			resp.Instructions = instructions
		}
	}
	if val, ok := contract.Extras["openai.responses.conversation"]; ok {
		if conv, ok := val.(*responsesTypes.ConversationRef); ok {
			resp.Conversation = conv
		}
	}
	if val, ok := contract.Extras["openai.responses.previous_response_id"].(string); ok {
		resp.PreviousResponseID = &val
	}
	if val, ok := contract.Extras["openai.responses.reasoning"]; ok {
		if reasoning, ok := val.(*responsesTypes.ResponseReasoning); ok {
			resp.Reasoning = reasoning
		}
	}
	if val, ok := contract.Extras["openai.responses.incomplete_details"]; ok {
		if details, ok := val.(*responsesTypes.IncompleteDetails); ok {
			resp.IncompleteDetails = details
		}
	}
	if val, ok := contract.Extras["openai.responses.background"].(bool); ok {
		resp.Background = &val
	}
	if val, ok := contract.Extras["openai.responses.max_output_tokens"].(int); ok {
		resp.MaxOutputTokens = &val
	}
	if val, ok := contract.Extras["openai.responses.max_tool_calls"].(int); ok {
		resp.MaxToolCalls = &val
	}
	if val, ok := contract.Extras["openai.responses.text"]; ok {
		// 类型断言需要根据实际类型处理
		if text, ok := val.(*responsesTypes.TextConfig); ok {
			resp.Text = text
		}
	}
	if val, ok := contract.Extras["openai.responses.top_p"].(float64); ok {
		resp.TopP = &val
	}
	if val, ok := contract.Extras["openai.responses.temperature"].(float64); ok {
		resp.Temperature = &val
	}
	if val, ok := contract.Extras["openai.responses.truncation"].(string); ok {
		resp.Truncation = &val
	}
	if val, ok := contract.Extras["openai.responses.user"].(string); ok {
		resp.User = &val
	}
	if val, ok := contract.Extras["openai.responses.service_tier"].(string); ok {
		resp.ServiceTier = &val
	}
}
