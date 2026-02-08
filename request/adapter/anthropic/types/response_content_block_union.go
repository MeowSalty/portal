package types

import (
	"encoding/json"
	"fmt"
)

// ResponseContentBlock 响应内容块
// 对应文档中的 ContentBlock。
type ResponseContentBlock struct {
	Text                *TextBlock
	Thinking            *ThinkingBlock
	RedactedThinking    *RedactedThinkingBlock
	ToolUse             *ToolUseBlock
	ServerToolUse       *ServerToolUseBlock
	WebSearchToolResult *WebSearchToolResultBlock
}

// MarshalJSON 实现 ResponseContentBlock 的序列化。
func (c ResponseContentBlock) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}
	if c.Text != nil {
		set(c.Text)
	}
	if c.Thinking != nil {
		set(c.Thinking)
	}
	if c.RedactedThinking != nil {
		set(c.RedactedThinking)
	}
	if c.ToolUse != nil {
		set(c.ToolUse)
	}
	if c.ServerToolUse != nil {
		set(c.ServerToolUse)
	}
	if c.WebSearchToolResult != nil {
		set(c.WebSearchToolResult)
	}
	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("响应内容块只能设置一种类型")
	}
	return json.Marshal(payload)
}

// UnmarshalJSON 实现 ResponseContentBlock 的反序列化。
func (c *ResponseContentBlock) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type ResponseContentBlockType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("响应内容块解析失败：%w", err)
	}

	switch t.Type {
	case ResponseContentBlockText:
		var v TextBlock
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("文本内容块解析失败：%w", err)
		}
		c.Text = &v
	case ResponseContentBlockThinking:
		var v ThinkingBlock
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("思考内容块解析失败：%w", err)
		}
		c.Thinking = &v
	case ResponseContentBlockRedactedThinking:
		var v RedactedThinkingBlock
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("脱敏思考内容块解析失败：%w", err)
		}
		c.RedactedThinking = &v
	case ResponseContentBlockToolUse:
		var v ToolUseBlock
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("工具使用内容块解析失败：%w", err)
		}
		c.ToolUse = &v
	case ResponseContentBlockServerToolUse:
		var v ServerToolUseBlock
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("服务器工具使用内容块解析失败：%w", err)
		}
		c.ServerToolUse = &v
	case ResponseContentBlockWebSearchToolResult:
		var v WebSearchToolResultBlock
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("Web 搜索工具结果内容块解析失败：%w", err)
		}
		c.WebSearchToolResult = &v
	default:
		return fmt.Errorf("不支持的响应内容块类型: %s", t.Type)
	}

	return nil
}
