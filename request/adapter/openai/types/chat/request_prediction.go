package chat

import (
	"encoding/json"
)

// PredictionType 表示 prediction 的类型
// 固定为 content。
type PredictionType = string

const (
	PredictionTypeContent PredictionType = "content"
)

// PredictionContent 表示预测内容
type PredictionContent struct {
	Type    PredictionType         `json:"type"`    // 类型
	Content PredictionContentUnion `json:"content"` // 内容
}

// PredictionContentUnion 表示预测内容联合类型
// 支持字符串或文本内容片段数组。
type PredictionContentUnion struct {
	StringValue  *string
	ContentParts []PredictionContentPart
}

// PredictionContentPartType 表示预测内容片段类型
type PredictionContentPartType = string

const (
	PredictionContentPartTypeText PredictionContentPartType = "text"
)

// PredictionContentPart 表示预测内容片段
// 仅支持文本内容。
type PredictionContentPart struct {
	Type PredictionContentPartType `json:"type"`
	Text string                    `json:"text"`
}

// MarshalJSON 实现 PredictionContentUnion 的自定义 JSON 序列化
func (pc PredictionContentUnion) MarshalJSON() ([]byte, error) {
	if pc.StringValue != nil {
		return json.Marshal(pc.StringValue)
	}
	if pc.ContentParts != nil {
		return json.Marshal(pc.ContentParts)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 PredictionContentUnion 的自定义 JSON 反序列化
func (pc *PredictionContentUnion) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		pc.StringValue = &str
		return nil
	}

	var parts []PredictionContentPart
	if err := json.Unmarshal(data, &parts); err == nil {
		pc.ContentParts = parts
		return nil
	}

	return nil
}
