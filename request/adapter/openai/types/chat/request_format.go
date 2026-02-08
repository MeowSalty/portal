package chat

import (
	"encoding/json"
)

// FormatUnion 表示响应格式联合类型
type FormatUnion struct {
	Text       *FormatText
	JSONSchema *FormatJSONSchema
	JSONObject *FormatJSONObject
}

// FormatText 表示文本响应格式
type FormatText struct {
	Type ResponseFormatType `json:"type"` // 类型
}

// FormatJSONSchema 表示 JSON Schema 响应格式
type FormatJSONSchema struct {
	Type       ResponseFormatType   `json:"type"`        // 类型
	JSONSchema FormatJSONSchemaSpec `json:"json_schema"` // JSON Schema
}

// FormatJSONSchemaSpec 表示 JSON Schema 细节
type FormatJSONSchemaSpec struct {
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Schema      map[string]interface{} `json:"schema"`
	Strict      *bool                  `json:"strict,omitempty"`
}

// ResponseFormatType 表示响应格式类型
// 可选值：text、json_schema、json_object。
type ResponseFormatType = string

const (
	ResponseFormatTypeText       ResponseFormatType = "text"
	ResponseFormatTypeJSONSchema ResponseFormatType = "json_schema"
	ResponseFormatTypeJSONObject ResponseFormatType = "json_object"
)

// FormatJSONObject 表示 JSON 对象响应格式
type FormatJSONObject struct {
	Type ResponseFormatType `json:"type"` // 类型
}

// MarshalJSON 实现 ResponseFormatUnion 的自定义 JSON 序列化
func (r FormatUnion) MarshalJSON() ([]byte, error) {
	if r.Text != nil {
		return json.Marshal(r.Text)
	}
	if r.JSONSchema != nil {
		return json.Marshal(r.JSONSchema)
	}
	if r.JSONObject != nil {
		return json.Marshal(r.JSONObject)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 FormatUnion 的自定义 JSON 反序列化
func (r *FormatUnion) UnmarshalJSON(data []byte) error {
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
	case "text":
		var text FormatText
		if err := json.Unmarshal(data, &text); err == nil {
			r.Text = &text
			return nil
		}
	case "json_schema":
		var jsonSchema FormatJSONSchema
		if err := json.Unmarshal(data, &jsonSchema); err == nil {
			r.JSONSchema = &jsonSchema
			return nil
		}
	case "json_object":
		var jsonObject FormatJSONObject
		if err := json.Unmarshal(data, &jsonObject); err == nil {
			r.JSONObject = &jsonObject
			return nil
		}
	}

	return nil
}
