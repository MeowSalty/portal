package responses

import (
	"encoding/json"

	portalErrors "github.com/MeowSalty/portal/errors"
)

// OutputItemType 表示 Responses 输出项类型。
type OutputItemType string

const (
	OutputItemTypeMessage              OutputItemType = "message"
	OutputItemTypeFunctionCall         OutputItemType = "function_call"
	OutputItemTypeFileSearchCall       OutputItemType = "file_search_call"
	OutputItemTypeWebSearchCall        OutputItemType = "web_search_call"
	OutputItemTypeComputerCall         OutputItemType = "computer_call"
	OutputItemTypeReasoning            OutputItemType = "reasoning"
	OutputItemTypeCodeInterpreterCall  OutputItemType = "code_interpreter_call"
	OutputItemTypeImageGenCall         OutputItemType = "image_generation_call"
	OutputItemTypeLocalShellCall       OutputItemType = "local_shell_call"
	OutputItemTypeShellCall            OutputItemType = "shell_call"
	OutputItemTypeShellCallOutput      OutputItemType = "shell_call_output"
	OutputItemTypeApplyPatchCall       OutputItemType = "apply_patch_call"
	OutputItemTypeApplyPatchCallOutput OutputItemType = "apply_patch_call_output"
	OutputItemTypeMCPCall              OutputItemType = "mcp_call"
	OutputItemTypeMCPListTools         OutputItemType = "mcp_list_tools"
	OutputItemTypeMCPApprovalRequest   OutputItemType = "mcp_approval_request"
	OutputItemTypeCustomToolCall       OutputItemType = "custom_tool_call"
	OutputItemTypeCompaction           OutputItemType = "compaction"
)

// OutputItem 表示 Responses 输出项联合类型。
// 使用 type 字段进行类型判断，根据不同的 type 值解析不同的字段。
// 参考：https://platform.openai.com/docs/api-reference/responses/object#responses/object-output
//
// 一致性规则：
//   - oneof 约束：只允许一个非空指针字段
//   - 非空字段的 Type 必须与预期常量一致
//   - 多指针冲突或 Type 与预期不一致时返回错误
type OutputItem struct {
	Message              *OutputMessage            `json:"-"`
	FunctionCall         *FunctionToolCall         `json:"-"`
	FileSearchCall       *FileSearchToolCall       `json:"-"`
	WebSearchCall        *WebSearchToolCall        `json:"-"`
	ComputerCall         *ComputerToolCall         `json:"-"`
	Reasoning            *ReasoningItem            `json:"-"`
	CodeInterpreterCall  *CodeInterpreterToolCall  `json:"-"`
	ImageGenCall         *ImageGenToolCall         `json:"-"`
	LocalShellCall       *LocalShellToolCall       `json:"-"`
	ShellCall            *FunctionShellCall        `json:"-"`
	ShellCallOutput      *FunctionShellCallOutput  `json:"-"`
	ApplyPatchCall       *ApplyPatchToolCall       `json:"-"`
	ApplyPatchCallOutput *ApplyPatchToolCallOutput `json:"-"`
	MCPCall              *MCPToolCall              `json:"-"`
	MCPListTools         *MCPListTools             `json:"-"`
	MCPApprovalRequest   *MCPApprovalRequest       `json:"-"`
	CustomToolCall       *CustomToolCall           `json:"-"`
	Compaction           *CompactionBody           `json:"-"`
}

// validateOneof 验证 oneof 约束：只允许一个非空字段，且该字段的 Type 必须与预期常量一致。
func (o *OutputItem) validateOneof() error {
	var nonEmptyCount int

	if o.Message != nil {
		nonEmptyCount++
		if o.Message.Type != OutputItemTypeMessage {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.FunctionCall != nil {
		nonEmptyCount++
		if o.FunctionCall.Type != string(OutputItemTypeFunctionCall) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.FileSearchCall != nil {
		nonEmptyCount++
		if o.FileSearchCall.Type != string(OutputItemTypeFileSearchCall) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.WebSearchCall != nil {
		nonEmptyCount++
		if o.WebSearchCall.Type != string(OutputItemTypeWebSearchCall) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.ComputerCall != nil {
		nonEmptyCount++
		if o.ComputerCall.Type != string(OutputItemTypeComputerCall) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.Reasoning != nil {
		nonEmptyCount++
		if o.Reasoning.Type != string(OutputItemTypeReasoning) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.CodeInterpreterCall != nil {
		nonEmptyCount++
		if o.CodeInterpreterCall.Type != string(OutputItemTypeCodeInterpreterCall) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.ImageGenCall != nil {
		nonEmptyCount++
		if o.ImageGenCall.Type != string(OutputItemTypeImageGenCall) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.LocalShellCall != nil {
		nonEmptyCount++
		if o.LocalShellCall.Type != string(OutputItemTypeLocalShellCall) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.ShellCall != nil {
		nonEmptyCount++
		if o.ShellCall.Type != string(OutputItemTypeShellCall) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.ShellCallOutput != nil {
		nonEmptyCount++
		if o.ShellCallOutput.Type != string(OutputItemTypeShellCallOutput) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.ApplyPatchCall != nil {
		nonEmptyCount++
		if o.ApplyPatchCall.Type != string(OutputItemTypeApplyPatchCall) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.ApplyPatchCallOutput != nil {
		nonEmptyCount++
		if o.ApplyPatchCallOutput.Type != string(OutputItemTypeApplyPatchCallOutput) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.MCPCall != nil {
		nonEmptyCount++
		if o.MCPCall.Type != string(OutputItemTypeMCPCall) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.MCPListTools != nil {
		nonEmptyCount++
		if o.MCPListTools.Type != string(OutputItemTypeMCPListTools) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.MCPApprovalRequest != nil {
		nonEmptyCount++
		if o.MCPApprovalRequest.Type != string(OutputItemTypeMCPApprovalRequest) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.CustomToolCall != nil {
		nonEmptyCount++
		if o.CustomToolCall.Type != string(OutputItemTypeCustomToolCall) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}
	if o.Compaction != nil {
		nonEmptyCount++
		if o.Compaction.Type != string(OutputItemTypeCompaction) {
			return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型与内容字段不匹配")
		}
	}

	if nonEmptyCount > 1 {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项存在多个非空字段，违反 oneof 约束")
	}

	return nil
}

// UnmarshalJSON 实现 OutputItem 的自定义反序列化。
func (o *OutputItem) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var base struct {
		Type OutputItemType `json:"type"`
		ID   string         `json:"id"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	if base.Type == "" {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型为空")
	}

	*o = OutputItem{}

	switch base.Type {
	case OutputItemTypeMessage:
		var msg OutputMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}
		o.Message = &msg
	case OutputItemTypeFunctionCall:
		var call FunctionToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		o.FunctionCall = &call
	case OutputItemTypeFileSearchCall:
		var call FileSearchToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		o.FileSearchCall = &call
	case OutputItemTypeWebSearchCall:
		var call WebSearchToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		o.WebSearchCall = &call
	case OutputItemTypeComputerCall:
		var call ComputerToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		o.ComputerCall = &call
	case OutputItemTypeReasoning:
		var item ReasoningItem
		if err := json.Unmarshal(data, &item); err != nil {
			return err
		}
		o.Reasoning = &item
	case OutputItemTypeCodeInterpreterCall:
		var call CodeInterpreterToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		o.CodeInterpreterCall = &call
	case OutputItemTypeImageGenCall:
		var call ImageGenToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		o.ImageGenCall = &call
	case OutputItemTypeLocalShellCall:
		var call LocalShellToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		o.LocalShellCall = &call
	case OutputItemTypeMCPCall:
		var call MCPToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		o.MCPCall = &call
	case OutputItemTypeCustomToolCall:
		var call CustomToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		o.CustomToolCall = &call
	case OutputItemTypeShellCall:
		var call FunctionShellCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		o.ShellCall = &call
	case OutputItemTypeShellCallOutput:
		var output FunctionShellCallOutput
		if err := json.Unmarshal(data, &output); err != nil {
			return err
		}
		o.ShellCallOutput = &output
	case OutputItemTypeApplyPatchCall:
		var call ApplyPatchToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		o.ApplyPatchCall = &call
	case OutputItemTypeApplyPatchCallOutput:
		var output ApplyPatchToolCallOutput
		if err := json.Unmarshal(data, &output); err != nil {
			return err
		}
		o.ApplyPatchCallOutput = &output
	case OutputItemTypeMCPListTools:
		var list MCPListTools
		if err := json.Unmarshal(data, &list); err != nil {
			return err
		}
		o.MCPListTools = &list
	case OutputItemTypeMCPApprovalRequest:
		var request MCPApprovalRequest
		if err := json.Unmarshal(data, &request); err != nil {
			return err
		}
		o.MCPApprovalRequest = &request
	case OutputItemTypeCompaction:
		var compaction CompactionBody
		if err := json.Unmarshal(data, &compaction); err != nil {
			return err
		}
		o.Compaction = &compaction
	default:
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项类型不支持")
	}

	return nil
}

// MarshalJSON 实现 OutputItem 的自定义序列化。
// 直接序列化具体结构体指针内容，禁止兜底输出顶层 type/id。
func (o OutputItem) MarshalJSON() ([]byte, error) {
	// 验证 oneof 约束和 Type 一致性
	if err := o.validateOneof(); err != nil {
		return nil, err
	}

	switch {
	case o.Message != nil:
		return json.Marshal(o.Message)
	case o.FunctionCall != nil:
		return json.Marshal(o.FunctionCall)
	case o.FileSearchCall != nil:
		return json.Marshal(o.FileSearchCall)
	case o.WebSearchCall != nil:
		return json.Marshal(o.WebSearchCall)
	case o.ComputerCall != nil:
		return json.Marshal(o.ComputerCall)
	case o.Reasoning != nil:
		return json.Marshal(o.Reasoning)
	case o.CodeInterpreterCall != nil:
		return json.Marshal(o.CodeInterpreterCall)
	case o.ImageGenCall != nil:
		return json.Marshal(o.ImageGenCall)
	case o.LocalShellCall != nil:
		return json.Marshal(o.LocalShellCall)
	case o.MCPCall != nil:
		return json.Marshal(o.MCPCall)
	case o.CustomToolCall != nil:
		return json.Marshal(o.CustomToolCall)
	case o.ShellCall != nil:
		return json.Marshal(o.ShellCall)
	case o.ShellCallOutput != nil:
		return json.Marshal(o.ShellCallOutput)
	case o.ApplyPatchCall != nil:
		return json.Marshal(o.ApplyPatchCall)
	case o.ApplyPatchCallOutput != nil:
		return json.Marshal(o.ApplyPatchCallOutput)
	case o.MCPListTools != nil:
		return json.Marshal(o.MCPListTools)
	case o.MCPApprovalRequest != nil:
		return json.Marshal(o.MCPApprovalRequest)
	case o.Compaction != nil:
		return json.Marshal(o.Compaction)
	default:
		return nil, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出项没有内容字段，无法序列化")
	}
}

// OutputMessage 表示输出消息（type == "message"）。
type OutputMessage struct {
	Type    OutputItemType         `json:"type"`    // 输出项类型
	ID      string                 `json:"id"`      // 输出项的唯一 ID
	Role    string                 `json:"role"`    // 角色：assistant
	Content []OutputMessageContent `json:"content"` // 消息内容
	Status  string                 `json:"status"`  // 状态：in_progress, completed, incomplete
}

// FunctionToolCall 表示函数工具调用（type == "function_call"）。
type FunctionToolCall struct {
	Type      string  `json:"type"`
	ID        *string `json:"id,omitempty"`
	CallID    string  `json:"call_id"`   // 调用 ID
	Name      string  `json:"name"`      // 函数名称
	Arguments string  `json:"arguments"` // JSON 格式的参数
	Status    string  `json:"status"`    // 状态：in_progress, completed, incomplete
}

// FileSearchToolCall 表示文件搜索工具调用（type == "file_search_call"）。
type FileSearchToolCall struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Status  string      `json:"status"`            // 状态：in_progress, searching, completed, incomplete, failed
	Queries []string    `json:"queries"`           // 搜索查询
	Results interface{} `json:"results,omitempty"` // 搜索结果
}

// WebSearchToolCall 表示网页搜索工具调用（type == "web_search_call"）。
type WebSearchToolCall struct {
	Type   string      `json:"type"`
	ID     string      `json:"id"`
	Status string      `json:"status"` // 状态：in_progress, searching, completed, failed
	Action interface{} `json:"action"` // 搜索动作
}

// ComputerToolCall 表示计算机使用工具调用（type == "computer_call"）。
type ComputerToolCall struct {
	Type                string      `json:"type"`
	ID                  string      `json:"id"`
	CallID              string      `json:"call_id"`                         // 调用 ID
	Action              interface{} `json:"action"`                          // 计算机动作
	PendingSafetyChecks interface{} `json:"pending_safety_checks,omitempty"` // 待处理的安全检查
	Status              string      `json:"status"`                          // 状态：in_progress, completed, incomplete
}

// ReasoningItem 表示推理项（type == "reasoning"）。
type ReasoningItem struct {
	Type             string                 `json:"type"`
	ID               string                 `json:"id"`
	EncryptedContent *string                `json:"encrypted_content,omitempty"` // 加密内容
	Summary          []OutputSummaryPart    `json:"summary"`                     // 推理摘要
	Content          []ReasoningTextContent `json:"content,omitempty"`           // 推理文本内容
	Status           string                 `json:"status,omitempty"`            // 状态
}

// CodeInterpreterToolCall 表示代码解释器工具调用（type == "code_interpreter_call"）。
type CodeInterpreterToolCall struct {
	Type        string                  `json:"type"`
	ID          string                  `json:"id"`
	Status      string                  `json:"status"`                 // 状态：in_progress, completed, incomplete, interpreting, failed
	ContainerID string                  `json:"container_id,omitempty"` // 容器 ID
	Code        *string                 `json:"code,omitempty"`         // 代码
	Outputs     []CodeInterpreterOutput `json:"outputs,omitempty"`      // 输出
}

// ImageGenToolCall 表示图像生成工具调用（type == "image_generation_call"）。
type ImageGenToolCall struct {
	Type   string  `json:"type"`
	ID     string  `json:"id"`
	Status string  `json:"status"` // 状态：in_progress, completed, generating, failed
	Result *string `json:"result"` // Base64 编码的图像
}

// LocalShellToolCall 表示本地 Shell 工具调用（type == "local_shell_call"）。
type LocalShellToolCall struct {
	Type   string      `json:"type"`
	ID     string      `json:"id"`
	CallID string      `json:"call_id"` // 调用 ID
	Action interface{} `json:"action"`  // Shell 动作
	Status string      `json:"status"`  // 状态：in_progress, completed, incomplete
}

// MCPToolCall 表示 MCP 工具调用（type == "mcp_call"）。
type MCPToolCall struct {
	Type              string  `json:"type"`
	ID                string  `json:"id"`
	ServerLabel       string  `json:"server_label"`                  // MCP 服务器标签
	Name              string  `json:"name"`                          // 工具名称
	Arguments         string  `json:"arguments"`                     // JSON 格式的参数
	Status            string  `json:"status"`                        // 状态
	ApprovalRequestID *string `json:"approval_request_id,omitempty"` // 审批请求 ID
	Output            *string `json:"output,omitempty"`              // 输出（可选）
	Error             *string `json:"error,omitempty"`               // 错误信息（可选）
}

// CustomToolCall 表示自定义工具调用（type == "custom_tool_call"）。
type CustomToolCall struct {
	Type   string `json:"type"`
	ID     string `json:"id,omitempty"` // 唯一 ID（可选）
	CallID string `json:"call_id"`      // 调用 ID
	Name   string `json:"name"`         // 工具名称
	Input  string `json:"input"`        // 输入
}

// FunctionShellCall 表示函数 Shell 工具调用（type == "shell_call"）。
// 参考：https://platform.openai.com/docs/api-reference/responses/object#responses/object-output
type FunctionShellCall struct {
	Type      string              `json:"type"`                 // 类型：shell_call
	ID        string              `json:"id"`                   // 唯一 ID
	CallID    string              `json:"call_id"`              // 调用 ID
	Action    FunctionShellAction `json:"action"`               // Shell 动作
	Status    string              `json:"status"`               // 状态：in_progress, completed, incomplete
	CreatedBy string              `json:"created_by,omitempty"` // 创建者 ID（可选）
}

// FunctionShellCallOutput 表示函数 Shell 工具调用输出（type == "shell_call_output"）。
// 参考：https://platform.openai.com/docs/api-reference/responses/object#responses/object-output
type FunctionShellCallOutput struct {
	Type            string                           `json:"type"`                 // 类型：shell_call_output
	ID              string                           `json:"id"`                   // 唯一 ID
	CallID          string                           `json:"call_id"`              // 调用 ID
	Output          []FunctionShellCallOutputContent `json:"output"`               // 输出内容
	MaxOutputLength *int                             `json:"max_output_length"`    // 最大输出长度
	CreatedBy       string                           `json:"created_by,omitempty"` // 创建者 ID（可选）
}

// ApplyPatchToolCall 表示应用补丁工具调用（type == "apply_patch_call"）。
// 参考：https://platform.openai.com/docs/api-reference/responses/object#responses/object-output
type ApplyPatchToolCall struct {
	Type      string                      `json:"type"`                 // 类型：apply_patch_call
	ID        string                      `json:"id"`                   // 唯一 ID
	CallID    string                      `json:"call_id"`              // 调用 ID
	Status    string                      `json:"status"`               // 状态：in_progress, completed
	Operation ApplyPatchToolCallOperation `json:"operation"`            // 补丁操作
	CreatedBy string                      `json:"created_by,omitempty"` // 创建者 ID（可选）
}

// ApplyPatchToolCallOutput 表示应用补丁工具调用输出（type == "apply_patch_call_output"）。
// 参考：https://platform.openai.com/docs/api-reference/responses/object#responses/object-output
type ApplyPatchToolCallOutput struct {
	Type      string  `json:"type"`                 // 类型：apply_patch_call_output
	ID        string  `json:"id"`                   // 唯一 ID
	CallID    string  `json:"call_id"`              // 调用 ID
	Status    string  `json:"status"`               // 状态：completed, failed
	Output    *string `json:"output"`               // 可选的文本输出
	CreatedBy string  `json:"created_by,omitempty"` // 创建者 ID（可选）
}

// MCPListTools 表示 MCP 列出工具（type == "mcp_list_tools"）。
// 参考：https://platform.openai.com/docs/api-reference/responses/object#responses/object-output
type MCPListTools struct {
	Type        string             `json:"type"`         // 类型：mcp_list_tools
	ID          string             `json:"id"`           // 唯一 ID
	ServerLabel string             `json:"server_label"` // MCP 服务器标签
	Tools       []MCPListToolsTool `json:"tools"`        // 工具列表
	Error       *string            `json:"error"`        // 错误信息（可选）
}

// MCPApprovalRequest 表示 MCP 审批请求（type == "mcp_approval_request"）。
// 参考：https://platform.openai.com/docs/api-reference/responses/object#responses/object-output
type MCPApprovalRequest struct {
	Type        string `json:"type"`         // 类型：mcp_approval_request
	ID          string `json:"id"`           // 唯一 ID
	ServerLabel string `json:"server_label"` // MCP 服务器标签
	Name        string `json:"name"`         // 工具名称
	Arguments   string `json:"arguments"`    // 参数（JSON 格式）
}

// CompactionBody 表示压缩项（type == "compaction"）。
// 参考：https://platform.openai.com/docs/api-reference/responses/object#responses/object-output
type CompactionBody struct {
	ID               string `json:"id"`                // 唯一 ID
	Type             string `json:"type"`              // 类型：compaction
	EncryptedContent string `json:"encrypted_content"` // 加密内容
	CreatedBy        string `json:"created_by"`        // 创建者 ID
}
