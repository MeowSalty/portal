package responses

import (
	"encoding/json"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/request/adapter/openai/converter/responses/helper"
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ResponseToContract 将 OpenAI Responses 响应转换为统一的 ResponseContract。
func ResponseToContract(resp *responsesTypes.Response, log logger.Logger) (*types.ResponseContract, error) {
	if resp == nil {
		return nil, nil
	}

	if log == nil {
		log = logger.NewNopLogger()
	}

	contract := &types.ResponseContract{
		Source: types.VendorSourceOpenAIResponse,
		ID:     resp.ID,
		Extras: make(map[string]interface{}),
	}

	// 转换顶层字段
	contract.Object = &resp.Object
	contract.Model = &resp.Model
	contract.CreatedAt = &resp.CreatedAt
	contract.Status = resp.Status
	contract.CompletedAt = resp.CompletedAt

	// 转换 Error
	if resp.Error != nil {
		contract.Error = helper.ConvertErrorToContract(resp.Error)
	}

	// 转换 Usage
	if resp.Usage != nil {
		contract.Usage = helper.ConvertUsageToContract(resp.Usage)
	}

	// 保存完整的 output 数组以支持可逆转换
	if len(resp.Output) > 0 {
		outputJSON, err := json.Marshal(resp.Output)
		if err != nil {
			log.Warn("序列化 output 失败", "error", err)
		} else {
			contract.Extras["openai.responses.output"] = string(outputJSON)
		}
	}

	// 转换 Output 为 Choice
	choice, err := convertOutputToChoice(resp, log)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "转换 Output 失败", err)
	}
	if choice != nil {
		contract.Choices = []types.ResponseChoice{*choice}
	}

	// 保存其余请求回显/控制字段到 Extras
	saveExtrasFields(resp, contract)

	return contract, nil
}

// convertOutputToChoice 将 Output 数组转换为单个 Choice。
func convertOutputToChoice(resp *responsesTypes.Response, log logger.Logger) (*types.ResponseChoice, error) {
	if len(resp.Output) == 0 {
		return nil, nil
	}

	choice := &types.ResponseChoice{
		Extras: make(map[string]interface{}),
	}

	// 转换 FinishReason
	finishReason := mapFinishReasonToContract(resp)
	choice.FinishReason = &finishReason

	// 保存原始完成原因
	if resp.IncompleteDetails != nil && resp.IncompleteDetails.Reason != nil {
		choice.NativeFinishReason = resp.IncompleteDetails.Reason
	} else if resp.Status != nil {
		choice.NativeFinishReason = resp.Status
	}

	// 转换 Message
	message := &types.ResponseMessage{
		Extras: make(map[string]interface{}),
	}

	for _, item := range resp.Output {
		if err := convertOutputItem(&item, message, log); err != nil {
			return nil, err
		}
	}

	choice.Message = message
	return choice, nil
}

// mapFinishReasonToContract 映射 OpenAI Responses 完成原因到统一的 FinishReason。
// 规则：
// - incomplete_details.reason = max_output_tokens => length
// - incomplete_details.reason = content_filter => content_filter
// - status = failed => failed
// - 其他 => stop (正常完成) 或 unknown
func mapFinishReasonToContract(resp *responsesTypes.Response) types.ResponseFinishReason {
	if resp.IncompleteDetails != nil && resp.IncompleteDetails.Reason != nil {
		switch *resp.IncompleteDetails.Reason {
		case "max_output_tokens":
			return types.ResponseFinishReasonLength
		case "content_filter":
			return types.ResponseFinishReasonContentFilter
		}
	}

	if resp.Status != nil {
		switch *resp.Status {
		case "failed":
			return types.ResponseFinishReasonFailed
		case "completed":
			return types.ResponseFinishReasonStop
		}
	}

	return types.ResponseFinishReasonUnknown
}

// convertOutputItem 转换单个输出项到 Message。
func convertOutputItem(item *responsesTypes.OutputItem, message *types.ResponseMessage, log logger.Logger) error {
	if item.Message != nil {
		return convertOutputMessage(item.Message, item.Message.ID, message, log)
	}
	if item.FunctionCall != nil {
		var itemID string
		if item.FunctionCall.ID != nil {
			itemID = *item.FunctionCall.ID
		}
		return convertFunctionCall(item.FunctionCall, itemID, message)
	}
	if item.Reasoning != nil {
		return convertReasoning(item.Reasoning, item.Reasoning.ID, message)
	}
	// 其他工具调用类型存入 Extras
	return convertOtherToolCall(item, message, log)
}

// convertOutputMessage 转换输出消息。
func convertOutputMessage(msg *responsesTypes.OutputMessage, itemID string, message *types.ResponseMessage, log logger.Logger) error {
	// 设置 Role
	role := msg.Role
	message.Role = &role
	message.ID = &itemID

	// 转换 Content
	for _, content := range msg.Content {
		part, err := convertMessageContent(&content)
		if err != nil {
			log.Warn("转换消息内容失败", "error", err)
			continue
		}
		message.Parts = append(message.Parts, *part)

		// 提取 refusal
		if content.Refusal != nil {
			message.Refusal = &content.Refusal.Refusal
		}
	}

	return nil
}

// convertMessageContent 转换消息内容。
func convertMessageContent(content *responsesTypes.OutputMessageContent) (*types.ResponseContentPart, error) {
	part := &types.ResponseContentPart{
		Extras: make(map[string]interface{}),
	}

	if content.OutputText != nil {
		part.Type = "output_text"
		part.Text = &content.OutputText.Text

		// 转换 Annotations
		if len(content.OutputText.Annotations) > 0 {
			part.Annotations = convertAnnotations(content.OutputText.Annotations)
		}

		// 保存 Logprobs 到 Extras
		if len(content.OutputText.Logprobs) > 0 {
			part.Extras["openai.responses.logprobs"] = content.OutputText.Logprobs
		}
	} else if content.Refusal != nil {
		part.Type = "refusal"
		part.Text = &content.Refusal.Refusal
	}

	return part, nil
}

// convertAnnotations 转换注释。
func convertAnnotations(annotations []responsesTypes.Annotation) []types.ResponseAnnotation {
	result := make([]types.ResponseAnnotation, 0, len(annotations))

	for _, ann := range annotations {
		contractAnn := types.ResponseAnnotation{
			Extras: make(map[string]interface{}),
		}

		switch {
		case ann.FileCitation != nil:
			v := ann.FileCitation
			contractAnn.Type = "file_citation"
			contractAnn.FileID = &v.FileID
			contractAnn.Extras["openai.responses.index"] = v.Index
			contractAnn.Extras["openai.responses.filename"] = v.Filename

		case ann.URLCitation != nil:
			v := ann.URLCitation
			contractAnn.Type = "url_citation"
			contractAnn.URL = &v.URL
			contractAnn.Title = &v.Title
			contractAnn.StartIndex = &v.StartIndex
			contractAnn.EndIndex = &v.EndIndex

		case ann.ContainerFileCitation != nil:
			v := ann.ContainerFileCitation
			contractAnn.Type = "container_file_citation"
			contractAnn.FileID = &v.FileID
			contractAnn.StartIndex = &v.StartIndex
			contractAnn.EndIndex = &v.EndIndex
			contractAnn.Extras["openai.responses.container_id"] = v.ContainerID
			contractAnn.Extras["openai.responses.filename"] = v.Filename

		case ann.FilePath != nil:
			v := ann.FilePath
			contractAnn.Type = "file_path"
			contractAnn.FileID = &v.FileID
			contractAnn.Extras["openai.responses.index"] = v.Index
		}

		result = append(result, contractAnn)
	}

	return result
}

// convertFunctionCall 转换函数工具调用。
func convertFunctionCall(call *responsesTypes.FunctionToolCall, itemID string, message *types.ResponseMessage) error {
	toolCall := types.ResponseToolCall{
		ID:        &call.CallID,
		Name:      &call.Name,
		Arguments: &call.Arguments,
		Extras:    make(map[string]interface{}),
	}

	toolType := "function"
	toolCall.Type = &toolType
	toolCall.Extras["openai.responses.item_id"] = itemID
	toolCall.Extras["openai.responses.status"] = call.Status

	message.ToolCalls = append(message.ToolCalls, toolCall)
	return nil
}

// convertReasoning 转换推理项。
func convertReasoning(item *responsesTypes.ReasoningItem, itemID string, message *types.ResponseMessage) error {
	part := types.ResponseContentPart{
		Type:   "thinking",
		Extras: make(map[string]interface{}),
	}

	// 保存原始结构到 Extras
	part.Extras["openai.responses.item_id"] = itemID
	part.Extras["openai.responses.reasoning"] = item

	// 提取摘要文本
	if len(item.Summary) > 0 {
		var summaryText string
		for _, s := range item.Summary {
			if s.Type == "summary_text" {
				summaryText += s.Text
			}
		}
		if summaryText != "" {
			part.Text = &summaryText
		}
	}

	message.Parts = append(message.Parts, part)
	return nil
}

// convertOtherToolCall 转换其他工具调用类型。
func convertOtherToolCall(item *responsesTypes.OutputItem, message *types.ResponseMessage, log logger.Logger) error {
	toolCall := types.ResponseToolCall{
		Extras: make(map[string]interface{}),
	}

	// 根据类型提取特定字段
	if item.WebSearchCall != nil {
		toolCall.ID = &item.WebSearchCall.ID
		toolType := item.WebSearchCall.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.web_search"] = *item.WebSearchCall
	} else if item.FileSearchCall != nil {
		toolCall.ID = &item.FileSearchCall.ID
		toolType := item.FileSearchCall.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.file_search"] = *item.FileSearchCall
	} else if item.ComputerCall != nil {
		toolCall.ID = &item.ComputerCall.CallID
		toolType := item.ComputerCall.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.computer"] = *item.ComputerCall
	} else if item.CodeInterpreterCall != nil {
		toolCall.ID = &item.CodeInterpreterCall.ID
		toolType := item.CodeInterpreterCall.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.code_interpreter"] = *item.CodeInterpreterCall
	} else if item.ImageGenCall != nil {
		toolCall.ID = &item.ImageGenCall.ID
		toolType := item.ImageGenCall.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.image_gen"] = *item.ImageGenCall
	} else if item.LocalShellCall != nil {
		toolCall.ID = &item.LocalShellCall.CallID
		toolType := item.LocalShellCall.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.local_shell"] = *item.LocalShellCall
	} else if item.MCPCall != nil {
		toolCall.ID = &item.MCPCall.ID
		toolType := item.MCPCall.Type
		toolCall.Type = &toolType
		toolCall.Name = &item.MCPCall.Name
		toolCall.Arguments = &item.MCPCall.Arguments
		toolCall.Extras["openai.responses.mcp"] = *item.MCPCall
	} else if item.CustomToolCall != nil {
		toolCall.ID = &item.CustomToolCall.ID
		toolType := item.CustomToolCall.Type
		toolCall.Type = &toolType
		toolCall.Name = &item.CustomToolCall.Name
		toolCall.Extras["openai.responses.custom"] = *item.CustomToolCall
	} else if item.ShellCall != nil {
		toolCall.ID = &item.ShellCall.CallID
		toolType := item.ShellCall.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.shell_call"] = *item.ShellCall
	} else if item.ShellCallOutput != nil {
		toolCall.ID = &item.ShellCallOutput.CallID
		toolType := item.ShellCallOutput.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.shell_call_output"] = *item.ShellCallOutput
	} else if item.ApplyPatchCall != nil {
		toolCall.ID = &item.ApplyPatchCall.CallID
		toolType := item.ApplyPatchCall.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.apply_patch_call"] = *item.ApplyPatchCall
	} else if item.ApplyPatchCallOutput != nil {
		toolCall.ID = &item.ApplyPatchCallOutput.CallID
		toolType := item.ApplyPatchCallOutput.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.apply_patch_call_output"] = *item.ApplyPatchCallOutput
	} else if item.MCPListTools != nil {
		toolCall.ID = &item.MCPListTools.ID
		toolType := item.MCPListTools.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.mcp_list_tools"] = *item.MCPListTools
	} else if item.MCPApprovalRequest != nil {
		toolCall.ID = &item.MCPApprovalRequest.ID
		toolType := item.MCPApprovalRequest.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.mcp_approval_request"] = *item.MCPApprovalRequest
	} else if item.Compaction != nil {
		toolCall.ID = &item.Compaction.ID
		toolType := item.Compaction.Type
		toolCall.Type = &toolType
		toolCall.Extras["openai.responses.compaction"] = *item.Compaction
	} else {
		log.Warn("未知的输出项类型")
	}

	message.ToolCalls = append(message.ToolCalls, toolCall)
	return nil
}

// saveExtrasFields 保存其余请求回显/控制字段到 Extras。
func saveExtrasFields(resp *responsesTypes.Response, contract *types.ResponseContract) {
	if resp.ParallelToolCalls {
		contract.Extras["openai.responses.parallel_tool_calls"] = resp.ParallelToolCalls
	}
	if len(resp.Metadata) > 0 {
		contract.Extras["openai.responses.metadata"] = resp.Metadata
	}
	if resp.ToolChoice != nil {
		contract.Extras["openai.responses.tool_choice"] = resp.ToolChoice
	}
	if len(resp.Tools) > 0 {
		contract.Extras["openai.responses.tools"] = resp.Tools
	}
	if resp.Instructions != nil {
		contract.Extras["openai.responses.instructions"] = resp.Instructions
	}
	if resp.Conversation != nil {
		contract.Extras["openai.responses.conversation"] = resp.Conversation
	}
	if resp.PreviousResponseID != nil {
		contract.Extras["openai.responses.previous_response_id"] = *resp.PreviousResponseID
	}
	if resp.Reasoning != nil {
		contract.Extras["openai.responses.reasoning"] = resp.Reasoning
	}
	if resp.IncompleteDetails != nil {
		contract.Extras["openai.responses.incomplete_details"] = resp.IncompleteDetails
	}
	if resp.Background != nil {
		contract.Extras["openai.responses.background"] = *resp.Background
	}
	if resp.MaxOutputTokens != nil {
		contract.Extras["openai.responses.max_output_tokens"] = *resp.MaxOutputTokens
	}
	if resp.MaxToolCalls != nil {
		contract.Extras["openai.responses.max_tool_calls"] = *resp.MaxToolCalls
	}
	if resp.Text != nil {
		contract.Extras["openai.responses.text"] = resp.Text
	}
	if resp.TopP != nil {
		contract.Extras["openai.responses.top_p"] = *resp.TopP
	}
	if resp.Temperature != nil {
		contract.Extras["openai.responses.temperature"] = *resp.Temperature
	}
	if resp.Truncation != nil {
		contract.Extras["openai.responses.truncation"] = *resp.Truncation
	}
	if resp.User != nil {
		contract.Extras["openai.responses.user"] = *resp.User
	}
	if resp.ServiceTier != nil {
		contract.Extras["openai.responses.service_tier"] = *resp.ServiceTier
	}
}
