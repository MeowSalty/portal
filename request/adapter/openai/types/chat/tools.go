package chat

import (
	"encoding/json"
	"reflect"

	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

// ChatToolUnion 表示 Chat Completions 专用的工具联合类型
// 仅支持 function 和 custom 两种类型，符合 OpenAI 官方 Chat Completions 规范
type ChatToolUnion struct {
	Function *ChatToolFunction
	Custom   *ChatToolCustom
}

// ChatToolFunction 表示 Chat Completions 的函数工具
type ChatToolFunction struct {
	Type        string                    `json:"type"`
	Function    shared.FunctionDefinition `json:"function,omitempty"`
	Name        *string                   `json:"name,omitempty"`
	Description *string                   `json:"description,omitempty"`
	Parameters  interface{}               `json:"parameters,omitempty"`
	Strict      *bool                     `json:"strict,omitempty"`
}

// ChatToolCustom 表示 Chat Completions 的自定义工具
type ChatToolCustom struct {
	Type        string      `json:"type"`
	Custom      interface{} `json:"custom,omitempty"`
	Name        *string     `json:"name,omitempty"`
	Description *string     `json:"description,omitempty"`
	Format      interface{} `json:"format,omitempty"`
}

// MarshalJSON 实现 ChatToolUnion 的自定义 JSON 序列化
func (t ChatToolUnion) MarshalJSON() ([]byte, error) {
	switch {
	case t.Function != nil:
		return json.Marshal(t.Function)
	case t.Custom != nil:
		return json.Marshal(t.Custom)
	default:
		return json.Marshal(nil)
	}
}

// UnmarshalJSON 实现 ChatToolUnion 的自定义 JSON 反序列化
func (t *ChatToolUnion) UnmarshalJSON(data []byte) error {
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
		var function ChatToolFunction
		if err := json.Unmarshal(data, &function); err == nil {
			t.Function = &function
			return nil
		}
	case "custom":
		var custom ChatToolCustom
		if err := json.Unmarshal(data, &custom); err == nil {
			t.Custom = &custom
			return nil
		}
	default:
		// 不支持的类型，返回错误
		return &json.UnmarshalTypeError{
			Value:  "tool type",
			Type:   reflect.TypeOf(ChatToolUnion{}),
			Offset: 0,
			Struct: "ChatToolUnion",
			Field:  "type",
		}
	}

	return nil
}
