package chat

import (
	"encoding/json"
)

// ChatResponseMessageRole 表示响应消息角色
// Chat Completions 响应中固定为 assistant。
type ChatResponseMessageRole string

const (
	ChatResponseMessageRoleAssistant ChatResponseMessageRole = "assistant"
)

// Message 表示聊天完成消息（非流式响应）
type Message struct {
	Content      *string                 `json:"content"`                 // 内容
	Refusal      *string                 `json:"refusal"`                 // 拒绝消息
	Role         ChatResponseMessageRole `json:"role"`                    // 角色
	FunctionCall *FunctionCall           `json:"function_call,omitempty"` // 函数调用
	ToolCalls    []MessageToolCall       `json:"tool_calls,omitempty"`    // 工具调用
	Annotations  []MessageAnnotation     `json:"annotations,omitempty"`   // 注释
	Audio        *ResponseAudio          `json:"audio,omitempty"`         // 音频

	// ExtraFields 存储未知字段
	ExtraFields map[string]interface{} `json:"-"`
}

// UnmarshalJSON 实现 Message 的自定义 JSON 反序列化
// 捕获所有未知字段并存储到 ExtraFields
func (m *Message) UnmarshalJSON(data []byte) error {
	// 1. 解析到通用 map 以获取所有字段
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 2. 使用类型别名避免递归调用
	type Alias Message
	aux := &struct{ *Alias }{Alias: (*Alias)(m)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 3. 提取未知字段
	m.ExtraFields = make(map[string]interface{})
	for key, value := range raw {
		if !messageKnownFields[key] {
			m.ExtraFields[key] = value
		}
	}

	return nil
}

// MarshalJSON 实现 Message 的自定义 JSON 序列化
// 合并已知字段和 ExtraFields 中的未知字段
func (m Message) MarshalJSON() ([]byte, error) {
	// 1. 如果没有未知字段，使用默认序列化
	if len(m.ExtraFields) == 0 {
		// 使用类型别名避免递归调用
		type Alias Message
		aux := Alias(m)
		return json.Marshal(aux)
	}

	// 2. 序列化已知字段到 map
	type Alias Message
	aux := Alias(m)
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
	for key, value := range m.ExtraFields {
		result[key] = value
	}

	// 5. 序列化最终结果
	return json.Marshal(result)
}

// messageKnownFields 定义 Message 结构体的所有已知字段名称
// 用于在反序列化时识别未知字段
var messageKnownFields = map[string]bool{
	"content":       true,
	"refusal":       true,
	"role":          true,
	"function_call": true,
	"tool_calls":    true,
	"annotations":   true,
	"audio":         true,
}
