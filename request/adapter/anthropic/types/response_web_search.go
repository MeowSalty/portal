package types

import (
	"encoding/json"
)

// WebSearchToolResultBlockContent web search 结果内容
// 可能是错误对象或 WebSearchResultBlock 数组。
type WebSearchToolResultBlockContent struct {
	Results []WebSearchResultBlock
	Error   *WebSearchToolResultError
}

// MarshalJSON 支持 WebSearchToolResultBlockContent 的联合类型序列化
func (c WebSearchToolResultBlockContent) MarshalJSON() ([]byte, error) {
	if c.Error != nil {
		return json.Marshal(c.Error)
	}
	if c.Results != nil {
		return json.Marshal(c.Results)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 支持 WebSearchToolResultBlockContent 的联合类型解析
func (c *WebSearchToolResultBlockContent) UnmarshalJSON(data []byte) error {
	var results []WebSearchResultBlock
	if err := json.Unmarshal(data, &results); err == nil {
		c.Results = results
		c.Error = nil
		return nil
	}

	var errObj WebSearchToolResultError
	if err := json.Unmarshal(data, &errObj); err == nil {
		if errObj.Type != "" || errObj.ErrorCode != "" {
			c.Error = &errObj
			c.Results = nil
			return nil
		}
	}

	return nil
}

// WebSearchToolResultError web search 工具结果错误
// 对应 WebSearchToolResultError。
type WebSearchToolResultError struct {
	Type      string                       `json:"type"`       // "web_search_tool_result_error"
	ErrorCode WebSearchToolResultErrorCode `json:"error_code"` // 错误码
}

// WebSearchResultBlock web search 单条结果
// 对应 WebSearchResultBlock。
type WebSearchResultBlock struct {
	Type             string `json:"type"`              // "web_search_result"
	Title            string `json:"title"`             // 结果标题
	URL              string `json:"url"`               // 结果链接
	EncryptedContent string `json:"encrypted_content"` // 加密内容
	PageAge          string `json:"page_age"`          // 页面年龄
}
