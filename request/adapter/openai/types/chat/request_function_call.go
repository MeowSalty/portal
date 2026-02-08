package chat

import (
	"encoding/json"
)

// ToolChoiceMode 表示 tool_choice 的模式值
// 可选值：none、auto。
type ToolChoiceMode = string

const (
	ToolChoiceModeNone ToolChoiceMode = "none"
	ToolChoiceModeAuto ToolChoiceMode = "auto"
)

// FunctionCallUnion 表示函数调用联合类型
type FunctionCallUnion struct {
	Mode     *ToolChoiceMode
	Function *FunctionCallOption
}

// FunctionCallOption 表示函数调用选项
type FunctionCallOption struct {
	Name string `json:"name"` // 函数名称
}

// MarshalJSON 实现 FunctionCallUnion 的自定义 JSON 序列化
func (f FunctionCallUnion) MarshalJSON() ([]byte, error) {
	if f.Mode != nil {
		return json.Marshal(f.Mode)
	}
	if f.Function != nil {
		return json.Marshal(f.Function)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 FunctionCallUnion 的自定义 JSON 反序列化
func (f *FunctionCallUnion) UnmarshalJSON(data []byte) error {
	// 首先尝试反序列化为字符串（如 "none", "auto"）
	var mode string
	if err := json.Unmarshal(data, &mode); err == nil {
		converted := ToolChoiceMode(mode)
		f.Mode = &converted
		return nil
	}

	// 尝试反序列化为 FunctionCallOption 对象
	var funcOpt FunctionCallOption
	if err := json.Unmarshal(data, &funcOpt); err == nil && funcOpt.Name != "" {
		f.Function = &funcOpt
		return nil
	}

	return nil
}
