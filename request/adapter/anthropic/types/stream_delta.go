package types

import (
	"encoding/json"
	"fmt"
)

// ContentBlockDelta 内容块增量更新。
// 使用联合类型以严格匹配流式响应格式。
type ContentBlockDelta struct {
	Text      *TextDelta
	InputJSON *InputJSONDelta
	Thinking  *ThinkingDelta
	Signature *SignatureDelta
	Citations *CitationsDelta
}

// DeltaType 内容块增量类型。
type DeltaType string

const (
	DeltaTypeText      DeltaType = "text_delta"
	DeltaTypeInputJSON DeltaType = "input_json_delta"
	DeltaTypeThinking  DeltaType = "thinking_delta"
	DeltaTypeSignature DeltaType = "signature_delta"
	DeltaTypeCitations DeltaType = "citations_delta"
)

// TextDelta 文本增量。
type TextDelta struct {
	Type DeltaType `json:"type"`
	Text string    `json:"text"`
}

// InputJSONDelta 工具输入 JSON 增量。
type InputJSONDelta struct {
	Type        DeltaType `json:"type"`
	PartialJSON string    `json:"partial_json"`
}

// ThinkingDelta 思考内容增量。
type ThinkingDelta struct {
	Type     DeltaType `json:"type"`
	Thinking string    `json:"thinking"`
}

// SignatureDelta 思考签名增量。
type SignatureDelta struct {
	Type      DeltaType `json:"type"`
	Signature string    `json:"signature"`
}

// CitationsDelta 引用增量。
type CitationsDelta struct {
	Type     DeltaType    `json:"type"`
	Citation TextCitation `json:"citation"`
}

// MarshalJSON 实现 ContentBlockDelta 的序列化。
func (c ContentBlockDelta) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}
	if c.Text != nil {
		set(c.Text)
	}
	if c.InputJSON != nil {
		set(c.InputJSON)
	}
	if c.Thinking != nil {
		set(c.Thinking)
	}
	if c.Signature != nil {
		set(c.Signature)
	}
	if c.Citations != nil {
		set(c.Citations)
	}
	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("内容块增量只能设置一种类型")
	}
	return json.Marshal(payload)
}

// UnmarshalJSON 实现 ContentBlockDelta 的反序列化。
func (c *ContentBlockDelta) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type DeltaType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("内容块增量解析失败：%w", err)
	}

	switch t.Type {
	case DeltaTypeText:
		var v TextDelta
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("文本增量解析失败：%w", err)
		}
		c.Text = &v
	case DeltaTypeInputJSON:
		var v InputJSONDelta
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("工具输入 JSON 增量解析失败：%w", err)
		}
		c.InputJSON = &v
	case DeltaTypeThinking:
		var v ThinkingDelta
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("思考内容增量解析失败：%w", err)
		}
		c.Thinking = &v
	case DeltaTypeSignature:
		var v SignatureDelta
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("思考签名增量解析失败：%w", err)
		}
		c.Signature = &v
	case DeltaTypeCitations:
		var v CitationsDelta
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("引用增量解析失败：%w", err)
		}
		c.Citations = &v
	default:
		return fmt.Errorf("不支持的内容块增量类型: %s", t.Type)
	}

	return nil
}
