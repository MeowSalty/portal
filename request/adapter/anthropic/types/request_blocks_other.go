package types

// SearchResultBlockParam 搜索结果内容块。
type SearchResultBlockParam struct {
	Type         ContentBlockType       `json:"type"`                    // "search_result"
	Content      []TextBlockParam       `json:"content"`                 // 内容
	Source       string                 `json:"source"`                  // 来源
	Title        string                 `json:"title"`                   // 标题
	CacheControl *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
	Citations    *CitationsConfigParam  `json:"citations,omitempty"`     // 引用配置
}

// ThinkingBlockParam 思考内容块。
type ThinkingBlockParam struct {
	Type      ContentBlockType `json:"type"`      // "thinking"
	Signature string           `json:"signature"` // 签名
	Thinking  string           `json:"thinking"`  // 思考内容
}

// RedactedThinkingBlockParam 脱敏思考内容块。
type RedactedThinkingBlockParam struct {
	Type ContentBlockType `json:"type"` // "redacted_thinking"
	Data string           `json:"data"` // 脱敏数据
}

// ToolUseBlockParam 工具使用内容块。
type ToolUseBlockParam struct {
	Type         ContentBlockType       `json:"type"`                    // "tool_use"
	ID           string                 `json:"id"`                      // 工具使用 ID
	Name         string                 `json:"name"`                    // 工具名称
	Input        map[string]interface{} `json:"input"`                   // 工具输入
	CacheControl *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
}

// ToolResultBlockParam 工具结果内容块。
type ToolResultBlockParam struct {
	Type         ContentBlockType       `json:"type"`                    // "tool_result"
	ToolUseID    string                 `json:"tool_use_id"`             // 关联工具使用 ID
	CacheControl *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
	Content      ToolResultContentParam `json:"content,omitempty"`       // 结果内容
	IsError      *bool                  `json:"is_error,omitempty"`      // 是否错误
}

// ServerToolUseBlockParam 服务器工具使用内容块。
type ServerToolUseBlockParam struct {
	Type         ContentBlockType       `json:"type"`                    // "server_tool_use"
	ID           string                 `json:"id"`                      // 工具使用 ID
	Name         string                 `json:"name"`                    // "web_search"
	Input        map[string]interface{} `json:"input"`                   // 工具输入
	CacheControl *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
}

// WebSearchToolResultBlockParam Web 搜索工具结果内容块。
type WebSearchToolResultBlockParam struct {
	Type         ContentBlockType                     `json:"type"`                    // "web_search_tool_result"
	ToolUseID    string                               `json:"tool_use_id"`             // 关联工具使用 ID
	Content      WebSearchToolResultBlockParamContent `json:"content"`                 // 结果内容
	CacheControl *CacheControlEphemeral               `json:"cache_control,omitempty"` // 缓存控制
}

// WebSearchToolRequestError Web 搜索工具请求错误。
type WebSearchToolRequestError struct {
	Type      string                       `json:"type"`       // "web_search_tool_result_error"
	ErrorCode WebSearchToolResultErrorCode `json:"error_code"` // 错误码
}

// WebSearchResultBlockParam Web 搜索结果块。
type WebSearchResultBlockParam struct {
	Type             string  `json:"type"`               // "web_search_result"
	EncryptedContent string  `json:"encrypted_content"`  // 加密内容
	Title            string  `json:"title"`              // 标题
	URL              string  `json:"url"`                // URL
	PageAge          *string `json:"page_age,omitempty"` // 页面年龄
}
