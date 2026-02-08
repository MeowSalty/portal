package types

// MessageDeltaUsage message_delta 中的用量统计（累积值）
type MessageDeltaUsage struct {
	CacheCreationInputTokens *int             `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     *int             `json:"cache_read_input_tokens,omitempty"`
	InputTokens              *int             `json:"input_tokens,omitempty"`
	OutputTokens             *int             `json:"output_tokens,omitempty"`
	ServerToolUse            *ServerToolUsage `json:"server_tool_use,omitempty"`
}
