package types

import (
	"encoding/json"
	"fmt"
)

// ToolType 工具类型。
type ToolType string

const (
	ToolTypeCustom             ToolType = "custom"
	ToolTypeBash20250124       ToolType = "bash_20250124"
	ToolTypeTextEditor20250124 ToolType = "text_editor_20250124"
	ToolTypeTextEditor20250429 ToolType = "text_editor_20250429"
	ToolTypeTextEditor20250728 ToolType = "text_editor_20250728"
	ToolTypeWebSearch20250305  ToolType = "web_search_20250305"
)

// ToolName 工具名称。
type ToolName string

const (
	ToolNameBash                    ToolName = "bash"
	ToolNameStrReplaceEditor        ToolName = "str_replace_editor"
	ToolNameStrReplaceBasedEditTool ToolName = "str_replace_based_edit_tool"
	ToolNameWebSearch               ToolName = "web_search"
)

// ToolUnion 工具联合类型。
type ToolUnion struct {
	Custom             *Tool
	Bash20250124       *ToolBash20250124
	TextEditor20250124 *ToolTextEditor20250124
	TextEditor20250429 *ToolTextEditor20250429
	TextEditor20250728 *ToolTextEditor20250728
	WebSearch20250305  *WebSearchTool20250305
}

// MarshalJSON 实现 ToolUnion 的序列化。
func (t ToolUnion) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}
	if t.Custom != nil {
		set(t.Custom)
	}
	if t.Bash20250124 != nil {
		set(t.Bash20250124)
	}
	if t.TextEditor20250124 != nil {
		set(t.TextEditor20250124)
	}
	if t.TextEditor20250429 != nil {
		set(t.TextEditor20250429)
	}
	if t.TextEditor20250728 != nil {
		set(t.TextEditor20250728)
	}
	if t.WebSearch20250305 != nil {
		set(t.WebSearch20250305)
	}
	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("tools 只能设置一种具体工具类型")
	}
	return json.Marshal(payload)
}

// UnmarshalJSON 实现 ToolUnion 的反序列化。
func (t *ToolUnion) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type ToolType `json:"type"`
	}
	var h typeHolder
	if err := json.Unmarshal(data, &h); err != nil {
		return fmt.Errorf("工具解析失败：%w", err)
	}

	switch h.Type {
	case ToolTypeBash20250124:
		var v ToolBash20250124
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("bash 工具解析失败：%w", err)
		}
		t.Bash20250124 = &v
	case ToolTypeTextEditor20250124:
		var v ToolTextEditor20250124
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("text_editor_20250124 工具解析失败：%w", err)
		}
		t.TextEditor20250124 = &v
	case ToolTypeTextEditor20250429:
		var v ToolTextEditor20250429
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("text_editor_20250429 工具解析失败：%w", err)
		}
		t.TextEditor20250429 = &v
	case ToolTypeTextEditor20250728:
		var v ToolTextEditor20250728
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("text_editor_20250728 工具解析失败：%w", err)
		}
		t.TextEditor20250728 = &v
	case ToolTypeWebSearch20250305:
		var v WebSearchTool20250305
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("web_search 工具解析失败：%w", err)
		}
		t.WebSearch20250305 = &v
	case ToolTypeCustom, "":
		var v Tool
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("自定义工具解析失败：%w", err)
		}
		t.Custom = &v
	default:
		return fmt.Errorf("不支持的工具类型: %s", h.Type)
	}

	return nil
}

// Tool 工具定义。
type Tool struct {
	InputSchema  InputSchema            `json:"input_schema"`            // 输入 schema
	Name         string                 `json:"name"`                    // 工具名称
	CacheControl *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
	Description  *string                `json:"description,omitempty"`   // 描述
	Type         *ToolType              `json:"type,omitempty"`          // "custom"
}

// ToolBash20250124 bash 工具。
type ToolBash20250124 struct {
	Name         ToolName               `json:"name"`                    // "bash"
	Type         ToolType               `json:"type"`                    // "bash_20250124"
	CacheControl *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
}

// ToolTextEditor20250124 文本编辑工具。
type ToolTextEditor20250124 struct {
	Name         ToolName               `json:"name"`                    // "str_replace_editor"
	Type         ToolType               `json:"type"`                    // "text_editor_20250124"
	CacheControl *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
}

// ToolTextEditor20250429 文本编辑工具。
type ToolTextEditor20250429 struct {
	Name         ToolName               `json:"name"`                    // "str_replace_based_edit_tool"
	Type         ToolType               `json:"type"`                    // "text_editor_20250429"
	CacheControl *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
}

// ToolTextEditor20250728 文本编辑工具。
type ToolTextEditor20250728 struct {
	Name          ToolName               `json:"name"`                     // "str_replace_based_edit_tool"
	Type          ToolType               `json:"type"`                     // "text_editor_20250728"
	CacheControl  *CacheControlEphemeral `json:"cache_control,omitempty"`  // 缓存控制
	MaxCharacters *int                   `json:"max_characters,omitempty"` // 最大字符数
}

// WebSearchTool20250305 Web 搜索工具。
type WebSearchTool20250305 struct {
	Name           ToolName               `json:"name"` // "web_search"
	Type           ToolType               `json:"type"` // "web_search_20250305"
	AllowedDomains []string               `json:"allowed_domains,omitempty"`
	BlockedDomains []string               `json:"blocked_domains,omitempty"`
	CacheControl   *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
	MaxUses        *int                   `json:"max_uses,omitempty"`
	UserLocation   *WebSearchUserLocation `json:"user_location,omitempty"`
}

// WebSearchUserLocation 用户位置。
type WebSearchUserLocation struct {
	Type     string  `json:"type"` // "approximate"
	City     *string `json:"city,omitempty"`
	Country  *string `json:"country,omitempty"`
	Region   *string `json:"region,omitempty"`
	Timezone *string `json:"timezone,omitempty"`
}

// InputSchemaType 输入 schema 类型。
type InputSchemaType string

const (
	InputSchemaTypeObject InputSchemaType = "object"
)

// InputSchema 工具输入 schema。
type InputSchema struct {
	Type       InputSchemaType        `json:"type"`                 // "object"
	Properties map[string]interface{} `json:"properties,omitempty"` // 属性
	Required   []string               `json:"required,omitempty"`   // 必填字段
}
