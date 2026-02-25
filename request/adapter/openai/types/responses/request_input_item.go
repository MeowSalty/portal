package responses

import (
	"encoding/json"

	portalErrors "github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
)

// InputItem 表示 Responses API 的输入项
// 使用 input items 形式时，input 为该结构数组。
//
// TODO：后期稳定后需要由多指针 oneof 转为 Type + Item interface
type InputItem struct {
	// ========== 消息类型 ==========
	// Message 表示简化输入消息（type == "message"）
	Message *InputMessage `json:"-"`
	// OutputMessage 表示输出消息（type == "message"），用于输入项中包含模型输出
	OutputMessage *OutputMessage `json:"-"`

	// ========== 引用类型 ==========
	// ItemReference 表示引用项（type == "item_reference"）
	ItemReference *ItemReferenceParam `json:"-"`

	// ========== 工具调用类型 ==========
	// FunctionCall 表示函数工具调用（type == "function_call"）
	FunctionCall *FunctionToolCall `json:"-"`

	// FileSearchCall 表示文件搜索工具调用（type == "file_search_call"）
	FileSearchCall *FileSearchToolCall `json:"-"`
	// WebSearchCall 表示网页搜索工具调用（type == "web_search_call"）
	WebSearchCall *WebSearchToolCall `json:"-"`
	// ComputerCall 表示计算机使用工具调用（type == "computer_call"）
	ComputerCall *ComputerToolCall `json:"-"`
	// CodeInterpreterCall 表示代码解释器工具调用（type == "code_interpreter_call"）
	CodeInterpreterCall *CodeInterpreterToolCall `json:"-"`
	// ImageGenCall 表示图像生成工具调用（type == "image_generation_call"）
	ImageGenCall *ImageGenToolCall `json:"-"`
	// LocalShellCall 表示本地 Shell 工具调用（type == "local_shell_call"）
	LocalShellCall *LocalShellToolCall `json:"-"`
	// FunctionShellCall 表示函数 Shell 调用（type == "function_shell_call"）
	FunctionShellCall *InputFunctionShellToolCall `json:"-"`
	// ApplyPatchCall 表示应用补丁工具调用（type == "apply_patch_call"）
	ApplyPatchCall *InputApplyPatchToolCall `json:"-"`
	// MCPCall 表示 MCP 工具调用（type == "mcp_call"）
	MCPCall *MCPToolCall `json:"-"`
	// CustomToolCall 表示自定义工具调用（type == "custom_tool_call"）
	CustomToolCall *CustomToolCall `json:"-"`
	// MCPListTools 表示 MCP 列出工具（type == "mcp_list_tools"）
	MCPListTools *InputMCPListToolsToolCall `json:"-"`

	// ========== 工具调用输出类型 ==========
	// FunctionCallOutput 表示函数工具调用输出（type == "function_call_output"）
	FunctionCallOutput *InputFunctionToolCallOutput `json:"-"`
	// ComputerCallOutput 表示计算机调用输出（type == "computer_call_output"）
	ComputerCallOutput *InputComputerToolCallOutput `json:"-"`
	// LocalShellCallOutput 表示本地 Shell 调用输出（type == "local_shell_call_output"）
	LocalShellCallOutput *InputLocalShellToolCallOutput `json:"-"`
	// FunctionShellCallOutput 表示函数 Shell 调用输出（type == "function_shell_call_output"）
	FunctionShellCallOutput *InputFunctionShellToolCallOutput `json:"-"`
	// ApplyPatchCallOutput 表示应用补丁调用输出（type == "apply_patch_call_output"）
	ApplyPatchCallOutput *InputApplyPatchToolCallOutput `json:"-"`
	// CustomToolCallOutput 表示自定义工具调用输出（type == "custom_tool_call_output"）
	CustomToolCallOutput *InputCustomToolCallOutput `json:"-"`

	// ========== 其他类型 ==========
	// Reasoning 表示推理项（type == "reasoning"）
	Reasoning *ReasoningItem `json:"-"`
	// Compaction 表示压缩项（type == "compaction"）
	Compaction *InputItemCompaction `json:"-"`
	// MCPApprovalRequest 表示 MCP 审批请求（type == "mcp_approval_request"）
	MCPApprovalRequest *InputItemMCPApprovalRequest `json:"-"`
	// MCPApprovalResponse 表示 MCP 审批响应（type == "mcp_approval_response"）
	MCPApprovalResponse *InputItemMCPApprovalResponse `json:"-"`
}

// MarshalJSON 实现 InputItem 的自定义 JSON 序列化
// 严格 oneof：只能设置一种类型，否则返回错误
func (i InputItem) MarshalJSON() ([]byte, error) {
	// 统计非空字段数量
	count := 0
	if i.Message != nil {
		count++
	}
	if i.OutputMessage != nil {
		count++
	}
	if i.ItemReference != nil {
		count++
	}
	if i.FunctionCall != nil {
		count++
	}
	if i.FileSearchCall != nil {
		count++
	}
	if i.WebSearchCall != nil {
		count++
	}
	if i.ComputerCall != nil {
		count++
	}
	if i.CodeInterpreterCall != nil {
		count++
	}
	if i.ImageGenCall != nil {
		count++
	}
	if i.LocalShellCall != nil {
		count++
	}
	if i.FunctionShellCall != nil {
		count++
	}
	if i.ApplyPatchCall != nil {
		count++
	}
	if i.MCPCall != nil {
		count++
	}
	if i.CustomToolCall != nil {
		count++
	}
	if i.MCPListTools != nil {
		count++
	}
	if i.FunctionCallOutput != nil {
		count++
	}
	if i.ComputerCallOutput != nil {
		count++
	}
	if i.LocalShellCallOutput != nil {
		count++
	}
	if i.FunctionShellCallOutput != nil {
		count++
	}
	if i.ApplyPatchCallOutput != nil {
		count++
	}
	if i.CustomToolCallOutput != nil {
		count++
	}
	if i.Reasoning != nil {
		count++
	}
	if i.Compaction != nil {
		count++
	}
	if i.MCPApprovalRequest != nil {
		count++
	}
	if i.MCPApprovalResponse != nil {
		count++
	}

	// count==0 序列化为 null
	if count == 0 {
		return json.Marshal(nil)
	}

	// count>1 返回错误
	if count > 1 {
		return nil, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "input item 只能设置一种类型")
	}

	// count==1 直接序列化该字段
	switch {
	case i.Message != nil:
		return json.Marshal(i.Message)
	case i.OutputMessage != nil:
		return json.Marshal(i.OutputMessage)
	case i.ItemReference != nil:
		return json.Marshal(i.ItemReference)
	case i.FunctionCall != nil:
		return json.Marshal(i.FunctionCall)
	case i.FileSearchCall != nil:
		return json.Marshal(i.FileSearchCall)
	case i.WebSearchCall != nil:
		return json.Marshal(i.WebSearchCall)
	case i.ComputerCall != nil:
		return json.Marshal(i.ComputerCall)
	case i.CodeInterpreterCall != nil:
		return json.Marshal(i.CodeInterpreterCall)
	case i.ImageGenCall != nil:
		return json.Marshal(i.ImageGenCall)
	case i.LocalShellCall != nil:
		return json.Marshal(i.LocalShellCall)
	case i.FunctionShellCall != nil:
		return json.Marshal(i.FunctionShellCall)
	case i.ApplyPatchCall != nil:
		return json.Marshal(i.ApplyPatchCall)
	case i.MCPCall != nil:
		return json.Marshal(i.MCPCall)
	case i.CustomToolCall != nil:
		return json.Marshal(i.CustomToolCall)
	case i.MCPListTools != nil:
		return json.Marshal(i.MCPListTools)
	case i.FunctionCallOutput != nil:
		return json.Marshal(i.FunctionCallOutput)
	case i.ComputerCallOutput != nil:
		return json.Marshal(i.ComputerCallOutput)
	case i.LocalShellCallOutput != nil:
		return json.Marshal(i.LocalShellCallOutput)
	case i.FunctionShellCallOutput != nil:
		return json.Marshal(i.FunctionShellCallOutput)
	case i.ApplyPatchCallOutput != nil:
		return json.Marshal(i.ApplyPatchCallOutput)
	case i.CustomToolCallOutput != nil:
		return json.Marshal(i.CustomToolCallOutput)
	case i.Reasoning != nil:
		return json.Marshal(i.Reasoning)
	case i.Compaction != nil:
		return json.Marshal(i.Compaction)
	case i.MCPApprovalRequest != nil:
		return json.Marshal(i.MCPApprovalRequest)
	case i.MCPApprovalResponse != nil:
		return json.Marshal(i.MCPApprovalResponse)
	default:
		return json.Marshal(nil)
	}
}

// UnmarshalJSON 实现 InputItem 的自定义 JSON 反序列化
// 严格 oneof：必须有 type 字段，按 type 分支反序列化到对应字段
func (i *InputItem) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	typeVal, ok := raw["type"].(string)
	if !ok || typeVal == "" {
		// 根据 OpenAI 文档，EasyInputMessage 的 type 是常量 message 但不是 required
		// 当缺少 type 时，默认按 message 类型处理
		typeVal = InputItemTypeMessage
	}

	*i = InputItem{}

	switch typeVal {
	case InputItemTypeMessage:
		// 区分 InputMessage 和 OutputMessage
		// OutputMessage 有 status 字段，InputMessage 没有
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}

		// 检查是否有 status 字段来判断是 OutputMessage
		if _, hasStatus := raw["status"]; hasStatus {
			var outputMsg OutputMessage
			if err := json.Unmarshal(data, &outputMsg); err != nil {
				return err
			}
			i.OutputMessage = &outputMsg
		} else {
			if shouldNormalizeToOutputMessage(raw) {
				logger.Default().WithGroup("openai").Warn("检测到非标准 input message，已自动转换为 OutputMessage，请尽快升级为 OpenAI 最新格式")
				normalized := normalizeOutputMessageRaw(raw)
				payload, err := json.Marshal(normalized)
				if err != nil {
					return err
				}
				var outputMsg OutputMessage
				if err := json.Unmarshal(payload, &outputMsg); err != nil {
					return err
				}
				i.OutputMessage = &outputMsg
				return nil
			}

			var msg InputMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				return err
			}
			i.Message = &msg
		}
	case InputItemTypeItemReference:
		var ref ItemReferenceParam
		if err := json.Unmarshal(data, &ref); err != nil {
			return err
		}
		i.ItemReference = &ref
	case InputItemTypeFunctionCall:
		var call FunctionToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.FunctionCall = &call
	case InputItemTypeFileSearchCall:
		var call FileSearchToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.FileSearchCall = &call
	case InputItemTypeWebSearchCall:
		var call WebSearchToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.WebSearchCall = &call
	case InputItemTypeComputerCall:
		var call ComputerToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.ComputerCall = &call
	case InputItemTypeCodeInterpreterCall:
		var call CodeInterpreterToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.CodeInterpreterCall = &call
	case InputItemTypeImageGenCall:
		var call ImageGenToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.ImageGenCall = &call
	case InputItemTypeLocalShellCall:
		var call LocalShellToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.LocalShellCall = &call
	case InputItemTypeFunctionShellCall:
		var call InputFunctionShellToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.FunctionShellCall = &call
	case InputItemTypeApplyPatchCall:
		var call InputApplyPatchToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.ApplyPatchCall = &call
	case InputItemTypeMCPCall:
		var call MCPToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.MCPCall = &call
	case InputItemTypeCustomToolCall:
		var call CustomToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.CustomToolCall = &call
	case InputItemTypeMCPListTools:
		var call InputMCPListToolsToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		i.MCPListTools = &call
	case InputItemTypeFunctionCallOutput:
		var output InputFunctionToolCallOutput
		if err := json.Unmarshal(data, &output); err != nil {
			return err
		}
		i.FunctionCallOutput = &output
	case InputItemTypeComputerCallOutput:
		var output InputComputerToolCallOutput
		if err := json.Unmarshal(data, &output); err != nil {
			return err
		}
		i.ComputerCallOutput = &output
	case InputItemTypeLocalShellCallOutput:
		var output InputLocalShellToolCallOutput
		if err := json.Unmarshal(data, &output); err != nil {
			return err
		}
		i.LocalShellCallOutput = &output
	case InputItemTypeFunctionShellCallOutput:
		var output InputFunctionShellToolCallOutput
		if err := json.Unmarshal(data, &output); err != nil {
			return err
		}
		i.FunctionShellCallOutput = &output
	case InputItemTypeApplyPatchCallOutput:
		var output InputApplyPatchToolCallOutput
		if err := json.Unmarshal(data, &output); err != nil {
			return err
		}
		i.ApplyPatchCallOutput = &output
	case InputItemTypeCustomToolCallOutput:
		var output InputCustomToolCallOutput
		if err := json.Unmarshal(data, &output); err != nil {
			return err
		}
		i.CustomToolCallOutput = &output
	case InputItemTypeReasoning:
		var item ReasoningItem
		if err := json.Unmarshal(data, &item); err != nil {
			return err
		}
		i.Reasoning = &item
	case InputItemTypeCompaction:
		var item InputItemCompaction
		if err := json.Unmarshal(data, &item); err != nil {
			return err
		}
		i.Compaction = &item
	case InputItemTypeMCPApprovalRequest:
		var item InputItemMCPApprovalRequest
		if err := json.Unmarshal(data, &item); err != nil {
			return err
		}
		i.MCPApprovalRequest = &item
	case InputItemTypeMCPApprovalResponse:
		var item InputItemMCPApprovalResponse
		if err := json.Unmarshal(data, &item); err != nil {
			return err
		}
		i.MCPApprovalResponse = &item
	default:
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "input item 类型不支持："+typeVal)
	}

	return nil
}
