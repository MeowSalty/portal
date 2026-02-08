package types

// TextBlock 文本内容块。
type TextBlock struct {
	Type      ResponseContentBlockType `json:"type"`      // "text"
	Text      string                   `json:"text"`      // 文本内容
	Citations []TextCitation           `json:"citations"` // 文本引用
}

// ThinkingBlock 思考内容块。
type ThinkingBlock struct {
	Type      ResponseContentBlockType `json:"type"`      // "thinking"
	Thinking  string                   `json:"thinking"`  // 思考内容
	Signature string                   `json:"signature"` // 思考签名
}

// RedactedThinkingBlock 脱敏思考内容块。
type RedactedThinkingBlock struct {
	Type ResponseContentBlockType `json:"type"` // "redacted_thinking"
	Data string                   `json:"data"` // 脱敏思考数据
}

// ToolUseBlock 工具使用内容块。
type ToolUseBlock struct {
	Type  ResponseContentBlockType `json:"type"`  // "tool_use"
	ID    string                   `json:"id"`    // 工具使用 ID
	Name  string                   `json:"name"`  // 工具名称
	Input map[string]interface{}   `json:"input"` // 工具输入
}

// ServerToolUseBlock 服务器工具使用内容块。
type ServerToolUseBlock struct {
	Type  ResponseContentBlockType `json:"type"`  // "server_tool_use"
	ID    string                   `json:"id"`    // 工具使用 ID
	Name  string                   `json:"name"`  // 工具名称
	Input map[string]interface{}   `json:"input"` // 工具输入
}

// WebSearchToolResultBlock Web 搜索工具结果内容块。
type WebSearchToolResultBlock struct {
	Type      ResponseContentBlockType        `json:"type"`        // "web_search_tool_result"
	ToolUseID string                          `json:"tool_use_id"` // 关联的工具使用 ID
	Content   WebSearchToolResultBlockContent `json:"content"`     // Web 搜索结果内容
}
