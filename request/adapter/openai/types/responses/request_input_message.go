package responses

import (
	"encoding/json"

	portalErrors "github.com/MeowSalty/portal/errors"
)

// InputMessage 表示简化输入消息
// role 支持 user/system/developer/assistant。
// content 支持 string 或 []InputContent 的联合类型。
type InputMessage struct {
	Type    InputItemType       `json:"type,omitempty"`
	ID      *string             `json:"id,omitempty"`
	Role    ResponseMessageRole `json:"role"`
	Content InputMessageContent `json:"content"`
}

// InputMessageContent 表示输入消息内容的联合类型
// 支持 string（文本输入）或 []InputContent（内容列表）。
type InputMessageContent struct {
	// String 表示文本输入
	String *string `json:"-"`
	// List 表示内容列表
	List *[]InputContent `json:"-"`
}

// MarshalJSON 实现 InputMessageContent 的自定义 JSON 序列化
// 优先序列化 String，其次序列化 List
func (c InputMessageContent) MarshalJSON() ([]byte, error) {
	// 优先使用 String
	if c.String != nil {
		return json.Marshal(c.String)
	}
	// 其次使用 List
	if c.List != nil {
		return json.Marshal(c.List)
	}
	// 都为空则序列化为 null
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 InputMessageContent 的自定义 JSON 反序列化
// 支持 string 或 []InputContent
func (c *InputMessageContent) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	// 尝试解析为 string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*c = InputMessageContent{String: &str}
		return nil
	}

	// 尝试解析为 []InputContent
	var list []InputContent
	if err := json.Unmarshal(data, &list); err == nil {
		*c = InputMessageContent{List: &list}
		return nil
	}

	return portalErrors.New(portalErrors.ErrCodeInvalidArgument, "input message content 必须是 string 或 []InputContent")
}

// NewInputMessageContentFromString 创建文本输入内容
func NewInputMessageContentFromString(text string) InputMessageContent {
	return InputMessageContent{String: &text}
}

// NewInputMessageContentFromList 创建内容列表输入
func NewInputMessageContentFromList(list []InputContent) InputMessageContent {
	return InputMessageContent{List: &list}
}

// ItemReferenceParam 表示引用项
// 用于 item_reference。
type ItemReferenceParam struct {
	Type *InputItemType `json:"type,omitempty"`
	ID   string         `json:"id"`
}
