package responses

import "encoding/json"

// InputFunctionShellToolCall 函数 Shell 工具调用
// 用于输入侧 shell_call 类型
type InputFunctionShellToolCall struct {
	Type      InputItemType   `json:"type"`
	ID        *string         `json:"id,omitempty"`
	CallID    string          `json:"call_id"`    // 调用 ID
	Action    json.RawMessage `json:"action"`     // Shell 动作
	Status    string          `json:"status"`     // 状态
	CreatedBy string          `json:"created_by"` // 创建者
}

// InputApplyPatchToolCall 应用补丁工具调用
// 用于输入侧 apply_patch_call 类型
type InputApplyPatchToolCall struct {
	Type      InputItemType   `json:"type"`
	ID        *string         `json:"id,omitempty"`
	CallID    string          `json:"call_id"`   // 调用 ID
	Operation json.RawMessage `json:"operation"` // 补丁操作
	Status    string          `json:"status"`    // 状态
}

// InputMCPListToolsToolCall MCP 列出工具
// 用于输入侧 mcp_list_tools 类型
type InputMCPListToolsToolCall struct {
	Type        InputItemType `json:"type"`
	ID          string        `json:"id"`
	ServerLabel string        `json:"server_label"`    // MCP 服务器标签
	Tools       interface{}   `json:"tools"`           // 工具列表
	Error       *string       `json:"error,omitempty"` // 错误信息
}
