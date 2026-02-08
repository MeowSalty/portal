package responses

import "encoding/json"

// InputFunctionToolCallOutput 函数工具调用输出
// 用于输入侧 function_call_output 类型
type InputFunctionToolCallOutput struct {
	Type   InputItemType   `json:"type"`
	ID     *string         `json:"id,omitempty"`
	CallID string          `json:"call_id"` // 调用 ID
	Output json.RawMessage `json:"output"`  // 输出内容：string | []InputContent
}

// InputComputerToolCallOutput 计算机工具调用输出
// 用于输入侧 computer_call_output 类型
type InputComputerToolCallOutput struct {
	Type   InputItemType   `json:"type"`
	ID     *string         `json:"id,omitempty"`
	CallID string          `json:"call_id"` // 调用 ID
	Output json.RawMessage `json:"output"`  // 输出内容：截图对象
}

// InputLocalShellToolCallOutput 本地 Shell 工具调用输出
// 用于输入侧 local_shell_call_output 类型
type InputLocalShellToolCallOutput struct {
	Type   InputItemType   `json:"type"`
	ID     *string         `json:"id,omitempty"`
	CallID string          `json:"call_id"` // 调用 ID
	Output json.RawMessage `json:"output"`  // 输出内容：string | []InputContent
	Status string          `json:"status"`  // 状态
}

// InputFunctionShellToolCallOutput 函数 Shell 工具调用输出
// 用于输入侧 shell_call_output 类型
type InputFunctionShellToolCallOutput struct {
	Type            InputItemType   `json:"type"`
	ID              string          `json:"id"`
	CallID          string          `json:"call_id"`                     // 调用 ID
	Output          json.RawMessage `json:"output"`                      // 输出内容：string | []InputContent
	MaxOutputLength *int            `json:"max_output_length,omitempty"` // 最大输出长度
	CreatedBy       string          `json:"created_by"`                  // 创建者
}

// InputApplyPatchToolCallOutput 应用补丁工具调用输出
// 用于输入侧 apply_patch_call_output 类型
type InputApplyPatchToolCallOutput struct {
	Type   InputItemType `json:"type"`
	ID     *string       `json:"id,omitempty"`
	CallID string        `json:"call_id"` // 调用 ID
	Status string        `json:"status"`  // 状态
}

// InputCustomToolCallOutput 自定义工具调用输出
// 用于输入侧 custom_tool_call_output 类型
type InputCustomToolCallOutput struct {
	Type   InputItemType   `json:"type"`
	ID     *string         `json:"id,omitempty"`
	CallID string          `json:"call_id"` // 调用 ID
	Output json.RawMessage `json:"output"`  // 输出内容：string | []InputContent
}
