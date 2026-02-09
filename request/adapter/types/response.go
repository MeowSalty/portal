package types

// ResponseFinishReason 表示统一完成原因。
type ResponseFinishReason string

const (
	ResponseFinishReasonStop                    ResponseFinishReason = "stop"
	ResponseFinishReasonLength                  ResponseFinishReason = "length"
	ResponseFinishReasonToolCalls               ResponseFinishReason = "tool_calls"
	ResponseFinishReasonContentFilter           ResponseFinishReason = "content_filter"
	ResponseFinishReasonRecitation              ResponseFinishReason = "recitation"
	ResponseFinishReasonLanguage                ResponseFinishReason = "language"
	ResponseFinishReasonToolCallMalformed       ResponseFinishReason = "tool_call_malformed"
	ResponseFinishReasonToolCallUnexpected      ResponseFinishReason = "tool_call_unexpected"
	ResponseFinishReasonToolCallLimit           ResponseFinishReason = "tool_call_limit"
	ResponseFinishReasonThoughtSignatureMissing ResponseFinishReason = "thought_signature_missing"
	ResponseFinishReasonFailed                  ResponseFinishReason = "failed"
	ResponseFinishReasonUnknown                 ResponseFinishReason = "unknown"
)

// ResponseContract 表示统一非流式响应。
type ResponseContract struct {
	Source VendorSource `json:"source"`

	ID        string  `json:"id,omitempty"`
	Object    *string `json:"object,omitempty"`     // 统一响应对象类型
	CreatedAt *int64  `json:"created_at,omitempty"` // 秒级时间戳
	Model     *string `json:"model,omitempty"`

	Status      *string `json:"status,omitempty"`       // 统一响应状态
	CompletedAt *int64  `json:"completed_at,omitempty"` // 秒级时间戳

	Choices []ResponseChoice `json:"choices,omitempty"`
	Usage   *ResponseUsage   `json:"usage,omitempty"`
	Error   *ResponseError   `json:"error,omitempty"`

	Extras map[string]interface{} `json:"-"`
}

// ResponseChoice 表示候选响应。
type ResponseChoice struct {
	Index        *int                  `json:"index,omitempty"`
	FinishReason *ResponseFinishReason `json:"finish_reason,omitempty"`
	// NativeFinishReason 保留原始厂商的完成原因值
	NativeFinishReason *string           `json:"native_finish_reason,omitempty"`
	Message            *ResponseMessage  `json:"message,omitempty"`
	Logprobs           *ResponseLogprobs `json:"logprobs,omitempty"`

	Extras map[string]interface{} `json:"-"`
}

// ResponseMessage 表示统一响应消息。
type ResponseMessage struct {
	ID      *string `json:"id,omitempty"`
	Role    *string `json:"role,omitempty"`
	Content *string `json:"content,omitempty"` // 纯文本
	Refusal *string `json:"refusal,omitempty"` // 拒绝说明

	Parts       []ResponseContentPart `json:"parts,omitempty"`
	ToolCalls   []ResponseToolCall    `json:"tool_calls,omitempty"`
	ToolResults []ResponseToolResult  `json:"tool_results,omitempty"`

	Audio *ResponseAudio `json:"audio,omitempty"`

	Extras map[string]interface{} `json:"-"`
}

// ResponseContentPart 表示结构化内容片段。
// Type 建议值：text, refusal, thinking, tool_call, tool_result, other。
type ResponseContentPart struct {
	Type string `json:"type"`

	Text        *string              `json:"text,omitempty"`
	Annotations []ResponseAnnotation `json:"annotations,omitempty"`

	ToolCall   *ResponseToolCall   `json:"tool_call,omitempty"`
	ToolResult *ResponseToolResult `json:"tool_result,omitempty"`

	Extras map[string]interface{} `json:"-"`
}

// ResponseAnnotation 表示引用或注释。
// Type 建议值：citation, url_citation, file_citation, other。
type ResponseAnnotation struct {
	Type string `json:"type"`

	StartIndex *int    `json:"start_index,omitempty"`
	EndIndex   *int    `json:"end_index,omitempty"`
	URL        *string `json:"url,omitempty"`
	Title      *string `json:"title,omitempty"`
	FileID     *string `json:"file_id,omitempty"`

	Extras map[string]interface{} `json:"-"`
}

// ResponseToolCall 表示工具调用。
type ResponseToolCall struct {
	ID        *string                `json:"id,omitempty"`
	Type      *string                `json:"type,omitempty"`
	Name      *string                `json:"name,omitempty"`
	Arguments *string                `json:"arguments,omitempty"` // JSON 字符串
	Payload   map[string]interface{} `json:"payload,omitempty"`   // 结构化参数

	Extras map[string]interface{} `json:"-"`
}

// ResponseToolResult 表示工具调用结果。
type ResponseToolResult struct {
	ID      *string                `json:"id,omitempty"`
	Name    *string                `json:"name,omitempty"`
	Content *string                `json:"content,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`

	Extras map[string]interface{} `json:"-"`
}

// ResponseAudio 表示音频响应。
type ResponseAudio struct {
	ID         *string `json:"id,omitempty"`
	Data       *string `json:"data,omitempty"`
	ExpiresAt  *int    `json:"expires_at,omitempty"`
	Transcript *string `json:"transcript,omitempty"`

	Extras map[string]interface{} `json:"-"`
}

// ResponseLogprobs 表示对数概率。
type ResponseLogprobs struct {
	Content []ResponseTokenLogprob `json:"content,omitempty"`
	Refusal []ResponseTokenLogprob `json:"refusal,omitempty"`

	Extras map[string]interface{} `json:"-"`
}

// ResponseTokenLogprob 表示 token 对数概率。
type ResponseTokenLogprob struct {
	Token       string                    `json:"token"`
	Bytes       []int                     `json:"bytes,omitempty"`
	Logprob     float64                   `json:"logprob"`
	TopLogprobs []ResponseTokenLogprobTop `json:"top_logprobs,omitempty"`
}

// ResponseTokenLogprobTop 表示顶部 token 对数概率。
type ResponseTokenLogprobTop struct {
	Token   string  `json:"token"`
	Bytes   []int   `json:"bytes,omitempty"`
	Logprob float64 `json:"logprob"`
}

// ResponseUsage 表示使用统计。
type ResponseUsage struct {
	InputTokens  *int `json:"input_tokens,omitempty"`
	OutputTokens *int `json:"output_tokens,omitempty"`
	TotalTokens  *int `json:"total_tokens,omitempty"`

	Extras map[string]interface{} `json:"-"`
}

// ResponseError 表示错误信息。
type ResponseError struct {
	Code    *string `json:"code,omitempty"`
	Message *string `json:"message,omitempty"`
	Type    *string `json:"type,omitempty"`
	Param   *string `json:"param,omitempty"`

	Extras map[string]interface{} `json:"-"`
}
