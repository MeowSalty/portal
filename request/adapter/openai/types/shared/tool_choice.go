package shared

import "encoding/json"

// ToolChoiceUnion 表示工具选择联合类型
type ToolChoiceUnion struct {
	Auto        *string
	Allowed     *ToolChoiceAllowed
	Named       *ToolChoiceNamed
	NamedCustom *ToolChoiceNamedCustom
	NamedMCP    *ToolChoiceNamedMCP
	Hosted      *ToolChoiceHosted
	ApplyPatch  *ToolChoiceApplyPatch
	Shell       *ToolChoiceShell
}

// ToolChoiceAllowed 表示允许的工具选择
type ToolChoiceAllowed struct {
	Type  string                   `json:"type"`  // 类型
	Mode  string                   `json:"mode"`  // 模式
	Tools []map[string]interface{} `json:"tools"` // 工具列表
}

// ToolChoiceNamed 表示命名的工具选择（函数）
type ToolChoiceNamed struct {
	Type     string `json:"type"` // 类型
	Function struct {
		Name string `json:"name"` // 函数名称
	} `json:"function,omitempty"` // 函数
	Name *string `json:"name,omitempty"`
}

// ToolChoiceNamedCustom 表示命名的自定义工具选择
type ToolChoiceNamedCustom struct {
	Type   string `json:"type"` // 类型
	Custom struct {
		Name string `json:"name"` // 名称
	} `json:"custom,omitempty"` // 自定义
	Name *string `json:"name,omitempty"`
}

// ToolChoiceNamedMCP 表示命名的 MCP 工具选择
type ToolChoiceNamedMCP struct {
	Type        string  `json:"type"`
	ServerLabel string  `json:"server_label"`
	Name        *string `json:"name,omitempty"`
}

// ToolChoiceHosted 表示内置工具选择
type ToolChoiceHosted struct {
	Type string `json:"type"`
}

// ToolChoiceApplyPatch 表示 apply_patch 工具选择
type ToolChoiceApplyPatch struct {
	Type string `json:"type"`
}

// ToolChoiceShell 表示 shell 工具选择
type ToolChoiceShell struct {
	Type string `json:"type"`
}

// MarshalJSON 实现 ToolChoiceUnion 的自定义 JSON 序列化
func (t ToolChoiceUnion) MarshalJSON() ([]byte, error) {
	if t.Auto != nil {
		return json.Marshal(t.Auto)
	}
	if t.Allowed != nil {
		return json.Marshal(t.Allowed)
	}
	if t.Named != nil {
		return json.Marshal(t.Named)
	}
	if t.NamedCustom != nil {
		return json.Marshal(t.NamedCustom)
	}
	if t.NamedMCP != nil {
		return json.Marshal(t.NamedMCP)
	}
	if t.Hosted != nil {
		return json.Marshal(t.Hosted)
	}
	if t.ApplyPatch != nil {
		return json.Marshal(t.ApplyPatch)
	}
	if t.Shell != nil {
		return json.Marshal(t.Shell)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 ToolChoiceUnion 的自定义 JSON 反序列化
func (t *ToolChoiceUnion) UnmarshalJSON(data []byte) error {
	// 首先尝试反序列化为字符串（如 "auto", "none", "required"）
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		t.Auto = &str
		return nil
	}

	// 尝试反序列化为对象，先解析 type 字段判断类型
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	typeVal, ok := raw["type"].(string)
	if !ok {
		return nil
	}

	switch typeVal {
	case "function":
		var named ToolChoiceNamed
		if err := json.Unmarshal(data, &named); err == nil {
			t.Named = &named
			return nil
		}
	case "custom":
		var custom ToolChoiceNamedCustom
		if err := json.Unmarshal(data, &custom); err == nil {
			t.NamedCustom = &custom
			return nil
		}
	case "mcp":
		var mcp ToolChoiceNamedMCP
		if err := json.Unmarshal(data, &mcp); err == nil {
			t.NamedMCP = &mcp
			return nil
		}
	case "apply_patch":
		var patch ToolChoiceApplyPatch
		if err := json.Unmarshal(data, &patch); err == nil {
			t.ApplyPatch = &patch
			return nil
		}
	case "shell":
		var shell ToolChoiceShell
		if err := json.Unmarshal(data, &shell); err == nil {
			t.Shell = &shell
			return nil
		}
	case "file_search", "web_search_preview", "computer_use_preview", "web_search_preview_2025_03_11", "image_generation", "code_interpreter":
		var hosted ToolChoiceHosted
		if err := json.Unmarshal(data, &hosted); err == nil {
			t.Hosted = &hosted
			return nil
		}
	case "allowed_tools":
		var allowed ToolChoiceAllowed
		if err := json.Unmarshal(data, &allowed); err == nil {
			t.Allowed = &allowed
			return nil
		}
	default:
		var allowed ToolChoiceAllowed
		if err := json.Unmarshal(data, &allowed); err == nil {
			t.Allowed = &allowed
			return nil
		}
	}

	return nil
}
