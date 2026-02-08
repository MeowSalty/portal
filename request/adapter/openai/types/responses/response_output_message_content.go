package responses

import (
	"encoding/json"

	portalErrors "github.com/MeowSalty/portal/errors"
)

// OutputMessageContentType 表示输出消息内容类型。
type OutputMessageContentType string

const (
	OutputMessageContentTypeOutputText OutputMessageContentType = "output_text"
	OutputMessageContentTypeRefusal    OutputMessageContentType = "refusal"
)

// OutputMessageContent 表示输出消息内容的联合类型。
// 使用多指针 oneof 结构，仅子结构体携带 `type` 字段。
//
// 强类型约束：OutputText 和 Refusal 互斥，仅能有一个非空。
type OutputMessageContent struct {
	OutputText *OutputTextContent `json:"-"` // 文本输出内容
	Refusal    *RefusalContent    `json:"-"` // 拒绝内容
}

// UnmarshalJSON 实现 OutputMessageContent 的自定义反序列化。
// 先读取 type 字段，再按类型反序列化到对应指针。
func (o *OutputMessageContent) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var base struct {
		Type OutputMessageContentType `json:"type"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	if base.Type == "" {
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出消息内容类型为空")
	}

	*o = OutputMessageContent{}

	switch base.Type {
	case OutputMessageContentTypeOutputText:
		var content OutputTextContent
		if err := json.Unmarshal(data, &content); err != nil {
			return err
		}
		content.Type = OutputMessageContentTypeOutputText
		o.OutputText = &content
	case OutputMessageContentTypeRefusal:
		var content RefusalContent
		if err := json.Unmarshal(data, &content); err != nil {
			return err
		}
		content.Type = OutputMessageContentTypeRefusal
		o.Refusal = &content
	default:
		return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出消息内容类型不支持")
	}

	return nil
}

// MarshalJSON 实现 OutputMessageContent 的自定义序列化。
// 要求仅一个指针非空；两者皆空或同时非空报错。
func (o OutputMessageContent) MarshalJSON() ([]byte, error) {
	// 互斥校验：仅能有一个非空
	if o.OutputText == nil && o.Refusal == nil {
		return nil, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出消息内容不能为空")
	}
	if o.OutputText != nil && o.Refusal != nil {
		return nil, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出消息内容类型冲突，不能同时存在 output_text 和 refusal")
	}

	// 序列化非空指针
	if o.OutputText != nil {
		// 确保 annotations 为空数组而非 null
		if o.OutputText.Annotations == nil {
			o.OutputText.Annotations = []Annotation{}
		}
		return json.Marshal(o.OutputText)
	}
	if o.Refusal != nil {
		return json.Marshal(o.Refusal)
	}

	return nil, portalErrors.New(portalErrors.ErrCodeInvalidArgument, "输出消息内容类型不支持")
}

// OutputTextContent 表示文本输出内容。
type OutputTextContent struct {
	Type        OutputMessageContentType `json:"type"`               // 内容类型
	Text        string                   `json:"text"`               // 文本内容
	Annotations []Annotation             `json:"annotations"`        // 注释
	Logprobs    []LogProb                `json:"logprobs,omitempty"` // 对数概率
}

// RefusalContent 表示拒绝内容。
type RefusalContent struct {
	Type    OutputMessageContentType `json:"type"`    // 内容类型
	Refusal string                   `json:"refusal"` // 拒绝说明
}
