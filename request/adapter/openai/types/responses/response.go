package responses

// Response 表示 OpenAI Responses API 响应
// 仅保留核心字段用于转换到统一响应。
type Response struct {
	ID        string       `json:"id"`
	Object    string       `json:"object"`
	CreatedAt int64        `json:"created_at"`
	Model     string       `json:"model"`
	Output    []OutputItem `json:"output"`
	Usage     *Usage       `json:"usage,omitempty"`
}

// OutputItem 表示 Responses 输出项
// 当前主要处理 type == "message" 的输出。
type OutputItem struct {
	ID      string       `json:"id,omitempty"`
	Type    string       `json:"type"`
	Role    string       `json:"role,omitempty"`
	Content []OutputPart `json:"content,omitempty"`
}

// OutputPart 表示 Responses 输出内容片段
// 主要关注 output_text。
type OutputPart struct {
	Type        string        `json:"type"`
	Text        string        `json:"text,omitempty"`
	Annotations []interface{} `json:"annotations"`
}

// Usage 表示 Responses 使用情况
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
