package chat

import (
	"encoding/json"
)

// ChatStreamMessageRole 表示流式响应中的消息角色
// 参考 ChatCompletionStreamResponseDelta.role。
type ChatStreamMessageRole string

const (
	ChatStreamMessageRoleDeveloper ChatStreamMessageRole = "developer"
	ChatStreamMessageRoleSystem    ChatStreamMessageRole = "system"
	ChatStreamMessageRoleUser      ChatStreamMessageRole = "user"
	ChatStreamMessageRoleAssistant ChatStreamMessageRole = "assistant"
	ChatStreamMessageRoleTool      ChatStreamMessageRole = "tool"
)

// Delta 表示聊天完成增量消息（流式响应）
type Delta struct {
	Content      *string                `json:"content,omitempty"`       // 内容
	Refusal      *string                `json:"refusal,omitempty"`       // 拒绝消息
	Role         *ChatStreamMessageRole `json:"role,omitempty"`          // 角色
	FunctionCall *FunctionCall          `json:"function_call,omitempty"` // 函数调用
	ToolCalls    []ToolCallChunk        `json:"tool_calls,omitempty"`    // 工具调用

	// ExtraFields 存储未知字段
	ExtraFields map[string]interface{} `json:"-"`
}

// UnmarshalJSON 实现 Delta 的自定义 JSON 反序列化
// 捕获所有未知字段并存储到 ExtraFields
func (d *Delta) UnmarshalJSON(data []byte) error {
	// 1. 解析到通用 map 以获取所有字段
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 2. 使用类型别名避免递归调用
	type Alias Delta
	aux := &struct{ *Alias }{Alias: (*Alias)(d)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 3. 提取未知字段
	d.ExtraFields = make(map[string]interface{})
	for key, value := range raw {
		if !deltaKnownFields[key] {
			d.ExtraFields[key] = value
		}
	}

	return nil
}

// MarshalJSON 实现 Delta 的自定义 JSON 序列化
// 合并已知字段和 ExtraFields 中的未知字段
func (d Delta) MarshalJSON() ([]byte, error) {
	// 1. 如果没有未知字段，使用默认序列化
	if len(d.ExtraFields) == 0 {
		// 使用类型别名避免递归调用
		type Alias Delta
		aux := Alias(d)
		return json.Marshal(aux)
	}

	// 2. 序列化已知字段到 map
	type Alias Delta
	aux := Alias(d)
	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	// 3. 解析到 map
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// 4. 合并未知字段
	for key, value := range d.ExtraFields {
		result[key] = value
	}

	// 5. 序列化最终结果
	return json.Marshal(result)
}

// deltaKnownFields 定义 Delta 结构体的所有已知字段名称
// 用于在反序列化时识别未知字段
var deltaKnownFields = map[string]bool{
	"content":       true,
	"refusal":       true,
	"role":          true,
	"function_call": true,
	"tool_calls":    true,
}
