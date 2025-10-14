package types

// AnthropicResponse Anthropic API 响应结构
type AnthropicResponse struct {
	ID           string            `json:"id"`                      // 消息 ID
	Type         string            `json:"type"`                    // "message"
	Role         string            `json:"role"`                    // "assistant"
	Content      []ResponseContent `json:"content"`                 // 内容块数组
	Model        string            `json:"model"`                   // 使用的模型
	StopReason   *string           `json:"stop_reason"`             // 停止原因
	StopSequence *string           `json:"stop_sequence,omitempty"` // 停止序列
	Usage        Usage             `json:"usage"`                   // 使用统计
}

// ResponseContent 响应内容块
type ResponseContent struct {
	Type  string      `json:"type"`            // "text", "tool_use"
	Text  *string     `json:"text,omitempty"`  // 文本内容
	ID    *string     `json:"id,omitempty"`    // 工具使用 ID
	Name  *string     `json:"name,omitempty"`  // 工具名称
	Input interface{} `json:"input,omitempty"` // 工具输入
}

// Usage 使用统计
type Usage struct {
	InputTokens  int `json:"input_tokens"`  // 输入 token 数
	OutputTokens int `json:"output_tokens"` // 输出 token 数
}

// StreamEvent 流式事件
type StreamEvent struct {
	Type         string             `json:"type"` // 事件类型
	Message      *AnthropicResponse `json:"message,omitempty"`
	Index        *int               `json:"index,omitempty"`
	ContentBlock *ResponseContent   `json:"content_block,omitempty"`
	Delta        *Delta             `json:"delta,omitempty"`
	Usage        *Usage             `json:"usage,omitempty"`
}

// Delta 增量更新
type Delta struct {
	Type         string  `json:"type,omitempty"`          // "text_delta", "input_json_delta"
	Text         *string `json:"text,omitempty"`          // 文本增量
	PartialJSON  *string `json:"partial_json,omitempty"`  // 部分 JSON
	StopReason   *string `json:"stop_reason,omitempty"`   // 停止原因
	StopSequence *string `json:"stop_sequence,omitempty"` // 停止序列
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Type  string `json:"type"`  // "error"
	Error Error  `json:"error"` // 错误详情
}

// Error 错误详情
type Error struct {
	Type    string `json:"type"`    // 错误类型
	Message string `json:"message"` // 错误消息
}
