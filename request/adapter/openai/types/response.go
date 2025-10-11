package types

// Response 表示 OpenAI 聊天完成响应
type Response struct {
	ID                string   `json:"id"`                           // 聊天完成 ID
	Choices           []Choice `json:"choices"`                      // 选择列表
	Created           int64    `json:"created"`                      // 创建时间戳
	Model             string   `json:"model"`                        // 模型名称
	Object            string   `json:"object"`                       // 对象类型
	ServiceTier       *string  `json:"service_tier,omitempty"`       // 服务层级
	SystemFingerprint *string  `json:"system_fingerprint,omitempty"` // 系统指纹
	Usage             *Usage   `json:"usage"`                        // 使用情况
}

// Choice 表示聊天完成选择
type Choice struct {
	FinishReason *string   `json:"finish_reason"`   // 完成原因
	Index        int       `json:"index"`           // 索引
	Logprobs     *Logprobs `json:"logprobs"`        // 对数概率
	Delta        *Delta    `json:"delta,omitempty"` // 消息
}

// Logprobs 表示对数概率信息
type Logprobs struct {
	Content []TokenLogprob `json:"content,omitempty"` // 内容对数概率
	Refusal []TokenLogprob `json:"refusal,omitempty"` // 拒绝对数概率
}

// TokenLogprob 表示 token 对数概率
type TokenLogprob struct {
	Token       string                   `json:"token"`                  // token
	Bytes       []int                    `json:"bytes"`                  // 字节表示
	Logprob     float64                  `json:"logprob"`                // 对数概率
	TopLogprobs []TokenLogprobTopLogprob `json:"top_logprobs,omitempty"` // 顶部对数概率
}

// TokenLogprobTopLogprob 表示顶部 token 对数概率
type TokenLogprobTopLogprob struct {
	Token   string  `json:"token"`   // token
	Bytes   []int   `json:"bytes"`   // 字节表示
	Logprob float64 `json:"logprob"` // 对数概率
}

// Delta 表示聊天完成消息
type Delta struct {
	Content      *string       `json:"content,omitempty"`       // 内容
	Refusal      *string       `json:"refusal,omitempty"`       // 拒绝消息
	Role         string        `json:"role"`                    // 角色
	FunctionCall *FunctionCall `json:"function_call,omitempty"` // 函数调用
	ToolCalls    []ToolCall    `json:"tool_calls,omitempty"`    // 工具调用
}

// ResponseAudio 表示音频响应
type ResponseAudio struct {
	ID         string `json:"id"`         // 音频 ID
	Data       string `json:"data"`       // 音频数据
	ExpiresAt  int    `json:"expires_at"` // 过期时间
	Transcript string `json:"transcript"` // 转录文本
}

// FunctionCall 表示函数调用
type FunctionCall struct {
	Arguments string `json:"arguments"` // 参数
	Name      string `json:"name"`      // 名称
}

// ToolCall 表示工具调用
type ToolCall struct {
	Index    int               `json:"index"`
	ID       string            `json:"id"`                 // 工具调用 ID
	Type     string            `json:"type"`               // 类型
	Function *ToolCallFunction `json:"function,omitempty"` // 函数
}

// ToolCallFunction 表示工具调用函数
type ToolCallFunction struct {
	Arguments string `json:"arguments"` // 参数
	Name      string `json:"name"`      // 名称
}

// Usage 表示使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`     // 提示 token 数
	CompletionTokens int `json:"completion_tokens"` // 完成 token 数
	TotalTokens      int `json:"total_tokens"`      // 总 token 数
}
