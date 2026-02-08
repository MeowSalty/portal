package types

import (
	"encoding/json"
	"fmt"
)

// ToolChoiceType 工具选择类型。
type ToolChoiceType string

const (
	ToolChoiceTypeAuto ToolChoiceType = "auto"
	ToolChoiceTypeAny  ToolChoiceType = "any"
	ToolChoiceTypeTool ToolChoiceType = "tool"
	ToolChoiceTypeNone ToolChoiceType = "none"
)

// ToolChoiceParam 工具选择联合类型。
type ToolChoiceParam struct {
	Auto *ToolChoiceAuto
	Any  *ToolChoiceAny
	Tool *ToolChoiceTool
	None *ToolChoiceNone
}

// MarshalJSON 实现 ToolChoiceParam 的序列化。
func (c ToolChoiceParam) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}
	if c.Auto != nil {
		set(c.Auto)
	}
	if c.Any != nil {
		set(c.Any)
	}
	if c.Tool != nil {
		set(c.Tool)
	}
	if c.None != nil {
		set(c.None)
	}
	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("tool_choice 只能设置一种类型")
	}
	return json.Marshal(payload)
}

// UnmarshalJSON 实现 ToolChoiceParam 的反序列化。
func (c *ToolChoiceParam) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type ToolChoiceType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("tool_choice 解析失败：%w", err)
	}

	switch t.Type {
	case ToolChoiceTypeAuto:
		var v ToolChoiceAuto
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("tool_choice auto 解析失败：%w", err)
		}
		c.Auto = &v
	case ToolChoiceTypeAny:
		var v ToolChoiceAny
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("tool_choice any 解析失败：%w", err)
		}
		c.Any = &v
	case ToolChoiceTypeTool:
		var v ToolChoiceTool
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("tool_choice tool 解析失败：%w", err)
		}
		c.Tool = &v
	case ToolChoiceTypeNone:
		var v ToolChoiceNone
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("tool_choice none 解析失败：%w", err)
		}
		c.None = &v
	default:
		return fmt.Errorf("不支持的 tool_choice 类型: %s", t.Type)
	}

	return nil
}

// ToolChoiceAuto 工具选择：自动。
type ToolChoiceAuto struct {
	Type                   ToolChoiceType `json:"type"`                                // "auto"
	DisableParallelToolUse *bool          `json:"disable_parallel_tool_use,omitempty"` // 是否禁用并行
}

// ToolChoiceAny 工具选择：任意。
type ToolChoiceAny struct {
	Type                   ToolChoiceType `json:"type"`                                // "any"
	DisableParallelToolUse *bool          `json:"disable_parallel_tool_use,omitempty"` // 是否禁用并行
}

// ToolChoiceTool 工具选择：指定工具。
type ToolChoiceTool struct {
	Type                   ToolChoiceType `json:"type"`                                // "tool"
	Name                   string         `json:"name"`                                // 工具名称
	DisableParallelToolUse *bool          `json:"disable_parallel_tool_use,omitempty"` // 是否禁用并行
}

// ToolChoiceNone 工具选择：不使用工具。
type ToolChoiceNone struct {
	Type ToolChoiceType `json:"type"` // "none"
}
