package helper

import (
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// ConvertMessagesToInputItems 从 Contract 转换消息列表为 input items。
// 只输出 InputMessage.content 列表；文本消息转为单一 input_text 片段。
func ConvertMessagesToInputItems(messages []types.Message) ([]responsesTypes.InputItem, error) {
	result := make([]responsesTypes.InputItem, 0, len(messages))

	for _, msg := range messages {
		item := responsesTypes.InputItem{}

		// 构造 InputMessage
		easyMsg := responsesTypes.InputMessage{
			Type: responsesTypes.InputItemTypeMessage,
			Role: responsesTypes.ResponseMessageRole(msg.Role),
		}

		// 转换内容为 InputContent 列表
		if msg.Content.Text != nil {
			// 文本消息转为单一 input_text 片段
			parts := []responsesTypes.InputContent{
				{
					Text: &responsesTypes.InputTextContent{
						Type: responsesTypes.InputContentTypeText,
						Text: *msg.Content.Text,
					},
				},
			}
			easyMsg.Content = responsesTypes.NewInputMessageContentFromList(parts)
		} else if len(msg.Content.Parts) > 0 {
			parts, err := convertContentPartsToInputContent(msg.Content.Parts)
			if err != nil {
				return nil, err
			}
			easyMsg.Content = responsesTypes.NewInputMessageContentFromList(parts)
		}

		item.Message = &easyMsg
		result = append(result, item)
	}

	return result, nil
}

// convertInputItemsToMessages 转换 input items 为消息列表。
// 从 InputMessage.content 列表生成 types.Content；单文本片段映射为 Content.Text。
func ConvertInputItemsToMessages(items []responsesTypes.InputItem) ([]types.Message, error) {
	result := make([]types.Message, 0, len(items))

	for _, item := range items {
		// 跳过 item_reference 类型
		if item.ItemReference != nil {
			continue
		}

		// 检查是否为消息类型
		var isMessage bool
		var role string
		var content []responsesTypes.InputContent

		if item.Message != nil {
			isMessage = true
			role = string(item.Message.Role)
			// 从 InputMessageContent 中提取 List
			if item.Message.Content.List != nil {
				content = *item.Message.Content.List
			} else if item.Message.Content.String != nil {
				// 字符串内容转换为 InputContent 列表
				content = []responsesTypes.InputContent{
					{
						Text: &responsesTypes.InputTextContent{
							Type: responsesTypes.InputContentTypeText,
							Text: *item.Message.Content.String,
						},
					},
				}
			}
		}

		if isMessage {
			if role == "" {
				continue
			}

			// 构造消息
			msg := types.Message{
				Role: role,
			}

			// 转换内容
			if len(content) == 0 {
				msg.Content = types.Content{}
			} else if len(content) == 1 && content[0].Text != nil {
				// 单个文本内容
				msg.Content = types.Content{
					Text: &content[0].Text.Text,
				}
			} else {
				// 多模态内容
				parts, err := convertInputContentToContentParts(content)
				if err != nil {
					return nil, err
				}
				msg.Content = types.Content{
					Parts: parts,
				}
			}

			result = append(result, msg)
		} else {
			// 处理非消息类型的 input items（工具调用/输出）
			part := convertInputItemToContentPart(item)

			// 根据内容部分类型确定消息角色
			role := "assistant"
			if part.Type == "tool_result" {
				role = "tool"
			}

			// 构造消息
			msg := types.Message{
				Role: role,
				Content: types.Content{
					Parts: []types.ContentPart{part},
				},
			}

			result = append(result, msg)
		}
	}

	return result, nil
}

// convertContentPartsToInputContent 从 Contract 转换内容片段为输入内容。
func convertContentPartsToInputContent(parts []types.ContentPart) ([]responsesTypes.InputContent, error) {
	result := make([]responsesTypes.InputContent, 0, len(parts))

	for _, part := range parts {
		var inputContent responsesTypes.InputContent

		switch part.Type {
		case "text", string(responsesTypes.InputContentTypeText):
			if part.Text != nil {
				inputContent.Text = &responsesTypes.InputTextContent{
					Type: responsesTypes.InputContentTypeText,
					Text: *part.Text,
				}
			}

		case "image", string(responsesTypes.InputContentTypeImage):
			if part.Image != nil {
				imageContent := responsesTypes.InputImageContent{
					Type:     responsesTypes.InputContentTypeImage,
					ImageURL: part.Image.URL,
				}
				if part.Image.Detail != nil {
					detail := shared.ImageDetail(*part.Image.Detail)
					imageContent.Detail = &detail
				}

				// 从 VendorExtras 恢复 FileID
				if part.VendorExtras != nil {
					if fileID, ok := part.VendorExtras["file_id"].(string); ok {
						imageContent.FileID = &fileID
					}
				}
				inputContent.Image = &imageContent
			}

		case "file", string(responsesTypes.InputContentTypeFile):
			if part.File != nil {
				inputContent.File = &responsesTypes.InputFileContent{
					Type:     responsesTypes.InputContentTypeFile,
					FileID:   part.File.ID,
					FileData: part.File.Data,
					Filename: part.File.Filename,
					FileURL:  part.File.URL,
				}
			}
		}

		result = append(result, inputContent)
	}

	return result, nil
}

// convertInputContentToContentParts 转换输入内容为统一内容片段。
func convertInputContentToContentParts(contents []responsesTypes.InputContent) ([]types.ContentPart, error) {
	result := make([]types.ContentPart, 0, len(contents))

	for _, content := range contents {
		var contractPart types.ContentPart

		switch {
		case content.Text != nil:
			contractPart = types.ContentPart{
				Type: string(responsesTypes.InputContentTypeText),
				Text: &content.Text.Text,
			}

		case content.Image != nil:
			contractPart = types.ContentPart{
				Type:  string(responsesTypes.InputContentTypeImage),
				Image: &types.Image{},
			}
			if content.Image.ImageURL != nil {
				contractPart.Image.URL = content.Image.ImageURL
			}
			if content.Image.FileID != nil {
				// FileID 放入 VendorExtras
				contractPart.VendorExtras = make(map[string]interface{})
				source := types.VendorSourceOpenAIResponse
				contractPart.VendorExtrasSource = &source
				contractPart.VendorExtras["file_id"] = *content.Image.FileID
			}
			if content.Image.Detail != nil {
				detail := string(*content.Image.Detail)
				contractPart.Image.Detail = &detail
			}

		case content.File != nil:
			contractPart = types.ContentPart{
				Type: string(responsesTypes.InputContentTypeFile),
				File: &types.File{
					ID:       content.File.FileID,
					Data:     content.File.FileData,
					Filename: content.File.Filename,
				},
			}
			if content.File.FileURL != nil {
				contractPart.File.URL = content.File.FileURL
			}
		}

		result = append(result, contractPart)
	}

	return result, nil
}

// convertInputItemToContentPart 将非消息类型的 input item 转换为 ContentPart。
// 根据 plans/responses-input-items-tool-mapping.md 的映射规则实现。
func convertInputItemToContentPart(item responsesTypes.InputItem) types.ContentPart {
	part := types.ContentPart{
		VendorExtras:       make(map[string]interface{}),
		VendorExtrasSource: nil,
	}
	source := types.VendorSourceOpenAIResponse
	part.VendorExtrasSource = &source

	// ========== 工具调用类型 ==========
	if item.FunctionCall != nil {
		v := item.FunctionCall
		// function_call -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "function"
		part.ToolCall = &types.ToolCall{
			Type:      &typeStr,
			ID:        &v.CallID,
			Name:      &v.Name,
			Arguments: &v.Arguments,
			Payload: map[string]interface{}{
				"status": v.Status,
			},
		}
	} else if item.FileSearchCall != nil {
		v := item.FileSearchCall
		// file_search_call -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "file_search"
		payload := map[string]interface{}{
			"status": v.Status,
		}
		if v.Queries != nil {
			payload["queries"] = v.Queries
		}
		if v.Results != nil {
			payload["results"] = v.Results
		}
		part.ToolCall = &types.ToolCall{
			Type:    &typeStr,
			ID:      nil,
			Name:    nil,
			Payload: payload,
		}
	} else if item.WebSearchCall != nil {
		v := item.WebSearchCall
		// web_search_call -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "web_search"
		payload := map[string]interface{}{
			"status": v.Status,
		}
		if v.Action != nil {
			payload["action"] = v.Action
		}
		part.ToolCall = &types.ToolCall{
			Type:    &typeStr,
			ID:      nil,
			Name:    nil,
			Payload: payload,
		}
	} else if item.ComputerCall != nil {
		v := item.ComputerCall
		// computer_call -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "computer"
		payload := map[string]interface{}{
			"status": v.Status,
		}
		if v.Action != nil {
			payload["action"] = v.Action
		}
		if v.PendingSafetyChecks != nil {
			payload["pending_safety_checks"] = v.PendingSafetyChecks
		}
		part.ToolCall = &types.ToolCall{
			Type:    &typeStr,
			ID:      &v.CallID,
			Name:    nil,
			Payload: payload,
		}
	} else if item.CodeInterpreterCall != nil {
		v := item.CodeInterpreterCall
		// code_interpreter_call -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "code_interpreter"
		payload := map[string]interface{}{
			"status": v.Status,
		}
		if v.ContainerID != "" {
			payload["container_id"] = v.ContainerID
		}
		if v.Code != nil {
			payload["code"] = *v.Code
		}
		if v.Outputs != nil {
			payload["outputs"] = v.Outputs
		}
		part.ToolCall = &types.ToolCall{
			Type:    &typeStr,
			ID:      nil,
			Name:    nil,
			Payload: payload,
		}
	} else if item.ImageGenCall != nil {
		v := item.ImageGenCall
		// image_generation_call -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "image_generation"
		payload := map[string]interface{}{
			"status": v.Status,
		}
		if v.Result != nil {
			payload["result"] = *v.Result
		}
		part.ToolCall = &types.ToolCall{
			Type:    &typeStr,
			ID:      nil,
			Name:    nil,
			Payload: payload,
		}
	} else if item.LocalShellCall != nil {
		v := item.LocalShellCall
		// local_shell_call -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "local_shell"
		payload := map[string]interface{}{
			"status": v.Status,
		}
		if v.Action != nil {
			payload["action"] = v.Action
		}
		part.ToolCall = &types.ToolCall{
			Type:    &typeStr,
			ID:      &v.CallID,
			Name:    nil,
			Payload: payload,
		}
	} else if item.FunctionShellCall != nil {
		v := item.FunctionShellCall
		// function_shell_call -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "function_shell"
		payload := map[string]interface{}{
			"status": v.Status,
		}
		if v.Action != nil {
			payload["action"] = v.Action
		}
		part.ToolCall = &types.ToolCall{
			Type:    &typeStr,
			ID:      &v.CallID,
			Name:    nil,
			Payload: payload,
		}
	} else if item.ApplyPatchCall != nil {
		v := item.ApplyPatchCall
		// apply_patch_call -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "apply_patch"
		payload := map[string]interface{}{
			"status": v.Status,
		}
		if v.Operation != nil {
			payload["operation"] = v.Operation
		}
		part.ToolCall = &types.ToolCall{
			Type:    &typeStr,
			ID:      &v.CallID,
			Name:    nil,
			Payload: payload,
		}
	} else if item.MCPCall != nil {
		v := item.MCPCall
		// mcp_call -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "mcp"
		payload := map[string]interface{}{}
		if v.ServerLabel != "" {
			payload["server_label"] = v.ServerLabel
		}
		if v.ApprovalRequestID != nil {
			payload["approval_request_id"] = *v.ApprovalRequestID
		}
		part.ToolCall = &types.ToolCall{
			Type:      &typeStr,
			ID:        nil,
			Name:      &v.Name,
			Arguments: &v.Arguments,
			Payload:   payload,
		}
	} else if item.CustomToolCall != nil {
		v := item.CustomToolCall
		// custom_tool_call -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "custom"
		part.ToolCall = &types.ToolCall{
			Type:      &typeStr,
			ID:        &v.CallID,
			Name:      &v.Name,
			Arguments: &v.Input,
			Payload:   map[string]interface{}{},
		}
	} else if item.MCPListTools != nil {
		v := item.MCPListTools
		// mcp_list_tools -> ContentPart.ToolCall
		part.Type = "tool_call"
		typeStr := "mcp_list_tools"
		payload := map[string]interface{}{}
		if v.ServerLabel != "" {
			payload["server_label"] = v.ServerLabel
		}
		part.ToolCall = &types.ToolCall{
			Type:    &typeStr,
			ID:      nil,
			Name:    nil,
			Payload: payload,
		}
	} else if item.FunctionCallOutput != nil {
		v := item.FunctionCallOutput
		// function_call_output -> ContentPart.ToolResult
		part.Type = "tool_result"
		part.ToolResult = convertToolOutput(&v.CallID, nil, "", v.Output)
	} else if item.ComputerCallOutput != nil {
		v := item.ComputerCallOutput
		// computer_call_output -> ContentPart.ToolResult
		part.Type = "tool_result"
		part.ToolResult = convertToolOutput(&v.CallID, nil, "", v.Output)
	} else if item.LocalShellCallOutput != nil {
		v := item.LocalShellCallOutput
		// local_shell_call_output -> ContentPart.ToolResult
		part.Type = "tool_result"
		part.ToolResult = convertToolOutput(&v.CallID, nil, v.Status, v.Output)
	} else if item.FunctionShellCallOutput != nil {
		v := item.FunctionShellCallOutput
		// function_shell_call_output -> ContentPart.ToolResult
		part.Type = "tool_result"
		part.ToolResult = convertToolOutput(&v.CallID, nil, "", v.Output)
	} else if item.ApplyPatchCallOutput != nil {
		v := item.ApplyPatchCallOutput
		// apply_patch_call_output -> ContentPart.ToolResult
		part.Type = "tool_result"
		part.ToolResult = convertToolOutput(&v.CallID, nil, v.Status, nil)
	} else if item.CustomToolCallOutput != nil {
		v := item.CustomToolCallOutput
		// custom_tool_call_output -> ContentPart.ToolResult
		part.Type = "tool_result"
		part.ToolResult = convertToolOutput(&v.CallID, nil, "", v.Output)
	} else if item.Reasoning != nil {
		// reasoning -> VendorExtras
		part.Type = "text"
		part.VendorExtras["raw_input_item"] = item.Reasoning
	} else if item.Compaction != nil {
		// compaction -> VendorExtras
		part.Type = "text"
		part.VendorExtras["raw_input_item"] = item.Compaction
	} else if item.MCPApprovalRequest != nil {
		// mcp_approval_request -> VendorExtras
		part.Type = "text"
		part.VendorExtras["raw_input_item"] = item.MCPApprovalRequest
	} else if item.MCPApprovalResponse != nil {
		// mcp_approval_response -> VendorExtras
		part.Type = "text"
		part.VendorExtras["raw_input_item"] = item.MCPApprovalResponse
	} else {
		// 未知类型 -> VendorExtras
		part.Type = "text"
		part.VendorExtras["raw_input_item"] = item
	}

	// 如果 Payload 为空，则省略它
	if part.ToolCall != nil && len(part.ToolCall.Payload) == 0 {
		part.ToolCall.Payload = nil
	}
	if part.ToolResult != nil && len(part.ToolResult.Payload) == 0 {
		part.ToolResult.Payload = nil
	}

	// 如果 VendorExtras 为空，则省略它
	if len(part.VendorExtras) == 0 {
		part.VendorExtras = nil
		part.VendorExtrasSource = nil
	}

	return part
}
