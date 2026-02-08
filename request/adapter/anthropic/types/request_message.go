package types

import (
	"encoding/json"
	"fmt"
)

// ServiceTier 服务层级。
type ServiceTier string

const (
	ServiceTierAuto         ServiceTier = "auto"
	ServiceTierStandardOnly ServiceTier = "standard_only"
)

// Role 消息角色。
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message 输入消息。
type Message struct {
	Role    Role                `json:"role"`    // "user" 或 "assistant"
	Content MessageContentParam `json:"content"` // string 或 []ContentBlockParam
}

// MessageContentParam 表示 string 或 []ContentBlockParam。
type MessageContentParam struct {
	StringValue *string
	Blocks      []ContentBlockParam
}

// MarshalJSON 实现 MessageContentParam 的序列化。
func (c MessageContentParam) MarshalJSON() ([]byte, error) {
	if c.StringValue != nil {
		return json.Marshal(c.StringValue)
	}
	if c.Blocks != nil {
		return json.Marshal(c.Blocks)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 MessageContentParam 的反序列化。
func (c *MessageContentParam) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		c.StringValue = &str
		return nil
	}

	var blocks []ContentBlockParam
	if err := json.Unmarshal(data, &blocks); err == nil {
		c.Blocks = blocks
		return nil
	}

	return fmt.Errorf("content 只允许 string 或内容块数组")
}

// SystemParam 表示 string 或 []TextBlockParam。
type SystemParam struct {
	StringValue *string
	Blocks      []TextBlockParam
}

// MarshalJSON 实现 SystemParam 的序列化。
func (c SystemParam) MarshalJSON() ([]byte, error) {
	if c.StringValue != nil {
		return json.Marshal(c.StringValue)
	}
	if c.Blocks != nil {
		return json.Marshal(c.Blocks)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 SystemParam 的反序列化。
func (c *SystemParam) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		c.StringValue = &str
		return nil
	}

	var blocks []TextBlockParam
	if err := json.Unmarshal(data, &blocks); err == nil {
		c.Blocks = blocks
		return nil
	}

	return fmt.Errorf("system 只允许 string 或文本块数组")
}

// Metadata 请求元数据。
type Metadata struct {
	UserID *string `json:"user_id,omitempty"` // 用户标识
}
