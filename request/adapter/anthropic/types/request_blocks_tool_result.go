package types

import (
	"encoding/json"
	"fmt"
)

// ToolResultContentBlockParam 工具结果允许的块类型。
type ToolResultContentBlockParam struct {
	Text         *TextBlockParam
	Image        *ImageBlockParam
	SearchResult *SearchResultBlockParam
	Document     *DocumentBlockParam
}

// MarshalJSON 实现 ToolResultContentBlockParam 的序列化。
func (c ToolResultContentBlockParam) MarshalJSON() ([]byte, error) {
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
	if c.SearchResult != nil {
		set(c.SearchResult)
	}
	if c.Document != nil {
		set(c.Document)
	}
	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("tool_result 内容块只能设置一种类型")
	}
	return json.Marshal(payload)
}

// UnmarshalJSON 实现 ToolResultContentBlockParam 的反序列化。
func (c *ToolResultContentBlockParam) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type ContentBlockType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("tool_result 内容块解析失败：%w", err)
	}

	switch t.Type {
	case ContentBlockTypeText:
		var v TextBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("tool_result 文本块解析失败：%w", err)
		}
		c.Text = &v
	case ContentBlockTypeImage:
		var v ImageBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("tool_result 图片块解析失败：%w", err)
		}
		c.Image = &v
	case ContentBlockTypeSearchResult:
		var v SearchResultBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("tool_result 搜索结果块解析失败：%w", err)
		}
		c.SearchResult = &v
	case ContentBlockTypeDocument:
		var v DocumentBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("tool_result 文档块解析失败：%w", err)
		}
		c.Document = &v
	default:
		return fmt.Errorf("tool_result 内容块类型不支持: %s", t.Type)
	}

	return nil
}

// ToolResultContentParam 表示 string 或 []ToolResultContentBlockParam。
type ToolResultContentParam struct {
	StringValue *string
	Blocks      []ToolResultContentBlockParam
}

// MarshalJSON 实现 ToolResultContentParam 的序列化。
func (c ToolResultContentParam) MarshalJSON() ([]byte, error) {
	if c.StringValue != nil {
		return json.Marshal(c.StringValue)
	}
	if c.Blocks != nil {
		return json.Marshal(c.Blocks)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 ToolResultContentParam 的反序列化。
func (c *ToolResultContentParam) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		c.StringValue = &str
		return nil
	}

	var blocks []ToolResultContentBlockParam
	if err := json.Unmarshal(data, &blocks); err == nil {
		c.Blocks = blocks
		return nil
	}

	return fmt.Errorf("tool_result.content 只允许 string 或指定内容块数组")
}

// WebSearchToolResultErrorCode Web 搜索错误码。
type WebSearchToolResultErrorCode string

const (
	WebSearchToolResultErrorInvalidToolInput WebSearchToolResultErrorCode = "invalid_tool_input"
	WebSearchToolResultErrorUnavailable      WebSearchToolResultErrorCode = "unavailable"
	WebSearchToolResultErrorMaxUsesExceeded  WebSearchToolResultErrorCode = "max_uses_exceeded"
	WebSearchToolResultErrorTooManyRequests  WebSearchToolResultErrorCode = "too_many_requests"
	WebSearchToolResultErrorQueryTooLong     WebSearchToolResultErrorCode = "query_too_long"
)

// WebSearchToolResultBlockParamContent 表示 []WebSearchResultBlockParam 或 WebSearchToolRequestError。
type WebSearchToolResultBlockParamContent struct {
	Results []WebSearchResultBlockParam
	Error   *WebSearchToolRequestError
}

// MarshalJSON 实现 WebSearchToolResultBlockParamContent 的序列化。
func (c WebSearchToolResultBlockParamContent) MarshalJSON() ([]byte, error) {
	if c.Error != nil {
		return json.Marshal(c.Error)
	}
	if c.Results != nil {
		return json.Marshal(c.Results)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 WebSearchToolResultBlockParamContent 的反序列化。
func (c *WebSearchToolResultBlockParamContent) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var results []WebSearchResultBlockParam
	if err := json.Unmarshal(data, &results); err == nil {
		c.Results = results
		return nil
	}

	var errObj WebSearchToolRequestError
	if err := json.Unmarshal(data, &errObj); err == nil {
		if errObj.Type != "" || errObj.ErrorCode != "" {
			c.Error = &errObj
			return nil
		}
	}

	return fmt.Errorf("Web 搜索结果内容解析失败")
}
