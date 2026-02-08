package types

import (
	"encoding/json"
	"fmt"
)

// ContentBlockType 内容块类型。
type ContentBlockType string

const (
	ContentBlockTypeText                ContentBlockType = "text"
	ContentBlockTypeImage               ContentBlockType = "image"
	ContentBlockTypeDocument            ContentBlockType = "document"
	ContentBlockTypeSearchResult        ContentBlockType = "search_result"
	ContentBlockTypeThinking            ContentBlockType = "thinking"
	ContentBlockTypeRedactedThinking    ContentBlockType = "redacted_thinking"
	ContentBlockTypeToolUse             ContentBlockType = "tool_use"
	ContentBlockTypeToolResult          ContentBlockType = "tool_result"
	ContentBlockTypeServerToolUse       ContentBlockType = "server_tool_use"
	ContentBlockTypeWebSearchToolResult ContentBlockType = "web_search_tool_result"
)

// ContentBlockParam 请求内容块联合类型。
type ContentBlockParam struct {
	Text                *TextBlockParam
	Image               *ImageBlockParam
	Document            *DocumentBlockParam
	SearchResult        *SearchResultBlockParam
	Thinking            *ThinkingBlockParam
	RedactedThinking    *RedactedThinkingBlockParam
	ToolUse             *ToolUseBlockParam
	ToolResult          *ToolResultBlockParam
	ServerToolUse       *ServerToolUseBlockParam
	WebSearchToolResult *WebSearchToolResultBlockParam
}

// MarshalJSON 实现 ContentBlockParam 的序列化。
func (c ContentBlockParam) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}
	if c.Text != nil {
		set(c.Text)
	}
	if c.Image != nil {
		set(c.Image)
	}
	if c.Document != nil {
		set(c.Document)
	}
	if c.SearchResult != nil {
		set(c.SearchResult)
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
	if c.ToolResult != nil {
		set(c.ToolResult)
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
		return nil, fmt.Errorf("内容块只能设置一种具体类型")
	}
	return json.Marshal(payload)
}

// UnmarshalJSON 实现 ContentBlockParam 的反序列化。
func (c *ContentBlockParam) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type ContentBlockType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("内容块解析失败：%w", err)
	}

	switch t.Type {
	case ContentBlockTypeText:
		var v TextBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("文本块解析失败：%w", err)
		}
		c.Text = &v
	case ContentBlockTypeImage:
		var v ImageBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("图片块解析失败：%w", err)
		}
		c.Image = &v
	case ContentBlockTypeDocument:
		var v DocumentBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("文档块解析失败：%w", err)
		}
		c.Document = &v
	case ContentBlockTypeSearchResult:
		var v SearchResultBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("搜索结果块解析失败：%w", err)
		}
		c.SearchResult = &v
	case ContentBlockTypeThinking:
		var v ThinkingBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("思考块解析失败：%w", err)
		}
		c.Thinking = &v
	case ContentBlockTypeRedactedThinking:
		var v RedactedThinkingBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("脱敏思考块解析失败：%w", err)
		}
		c.RedactedThinking = &v
	case ContentBlockTypeToolUse:
		var v ToolUseBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("工具使用块解析失败：%w", err)
		}
		c.ToolUse = &v
	case ContentBlockTypeToolResult:
		var v ToolResultBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("工具结果块解析失败：%w", err)
		}
		c.ToolResult = &v
	case ContentBlockTypeServerToolUse:
		var v ServerToolUseBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("服务器工具使用块解析失败：%w", err)
		}
		c.ServerToolUse = &v
	case ContentBlockTypeWebSearchToolResult:
		var v WebSearchToolResultBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("Web 搜索工具结果块解析失败：%w", err)
		}
		c.WebSearchToolResult = &v
	default:
		return fmt.Errorf("不支持的内容块类型: %s", t.Type)
	}

	return nil
}

// ContentBlock 请求内容块（兼容名称）。
type ContentBlock = ContentBlockParam
