package types

// Usage 使用统计
// 注意：数值字段使用指针以区分缺失与 0。
type Usage struct {
	CacheCreationInputTokens *int                 `json:"cache_creation_input_tokens,omitempty"` // cache 创建输入 token
	CacheReadInputTokens     *int                 `json:"cache_read_input_tokens,omitempty"`     // cache 读取输入 token
	InputTokens              *int                 `json:"input_tokens,omitempty"`                // 输入 token
	OutputTokens             *int                 `json:"output_tokens,omitempty"`               // 输出 token
	CacheCreation            *CacheCreation       `json:"cache_creation,omitempty"`              // cache 细分
	ServerToolUse            *ServerToolUsage     `json:"server_tool_use,omitempty"`             // 服务器工具使用统计
	ServiceTier              *ResponseServiceTier `json:"service_tier,omitempty"`                // 服务层级
}

// CacheCreation cache 使用细分
type CacheCreation struct {
	Ephemeral1hInputTokens *int `json:"ephemeral_1h_input_tokens,omitempty"`
	Ephemeral5mInputTokens *int `json:"ephemeral_5m_input_tokens,omitempty"`
}

// ServerToolUsage 服务器工具使用统计
type ServerToolUsage struct {
	WebSearchRequests *int `json:"web_search_requests,omitempty"`
}
