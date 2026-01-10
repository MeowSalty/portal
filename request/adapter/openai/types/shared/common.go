package shared

import "encoding/json"

// FunctionDefinition 表示函数定义
type FunctionDefinition struct {
	Name        string      `json:"name"`                  // 函数名称
	Description *string     `json:"description,omitempty"` // 描述
	Parameters  interface{} `json:"parameters"`            // 参数
}

// ToolChoiceUnion 表示工具选择联合类型
type ToolChoiceUnion struct {
	Auto        *string
	Allowed     *ToolChoiceAllowed
	Named       *ToolChoiceNamed
	NamedCustom *ToolChoiceNamedCustom
}

// ToolChoiceAllowed 表示允许的工具选择
type ToolChoiceAllowed struct {
	Type  string                   `json:"type"`  // 类型
	Mode  string                   `json:"mode"`  // 模式
	Tools []map[string]interface{} `json:"tools"` // 工具列表
}

// ToolChoiceNamed 表示命名的工具选择
type ToolChoiceNamed struct {
	Type     string `json:"type"` // 类型
	Function struct {
		Name string `json:"name"` // 函数名称
	} `json:"function"` // 函数
}

// ToolChoiceNamedCustom 表示命名的自定义工具选择
type ToolChoiceNamedCustom struct {
	Type   string `json:"type"` // 类型
	Custom struct {
		Name string `json:"name"` // 名称
	} `json:"custom"` // 自定义
}

// ToolUnion 表示工具联合类型
type ToolUnion struct {
	Function *ToolFunction
	Custom   *ToolCustom
}

// ToolFunction 表示函数工具
type ToolFunction struct {
	Type     string             `json:"type"`     // 类型
	Function FunctionDefinition `json:"function"` // 函数定义
}

// ToolCustom 表示自定义工具
type ToolCustom struct {
	Type   string      `json:"type"`   // 类型
	Custom interface{} `json:"custom"` // 自定义内容
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
	default:
		// 尝试 ToolChoiceAllowed
		var allowed ToolChoiceAllowed
		if err := json.Unmarshal(data, &allowed); err == nil {
			t.Allowed = &allowed
			return nil
		}
	}

	return nil
}

// MarshalJSON 实现 ToolUnion 的自定义 JSON 序列化
func (t ToolUnion) MarshalJSON() ([]byte, error) {
	if t.Function != nil {
		return json.Marshal(t.Function)
	}
	if t.Custom != nil {
		return json.Marshal(t.Custom)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 ToolUnion 的自定义 JSON 反序列化
func (t *ToolUnion) UnmarshalJSON(data []byte) error {
	// 解析到通用 map 以检查 type 字段
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
		var function ToolFunction
		if err := json.Unmarshal(data, &function); err == nil {
			t.Function = &function
			return nil
		}
	case "custom":
		var custom ToolCustom
		if err := json.Unmarshal(data, &custom); err == nil {
			t.Custom = &custom
			return nil
		}
	}

	return nil
}
