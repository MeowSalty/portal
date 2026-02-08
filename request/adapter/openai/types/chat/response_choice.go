package chat

// Choice 表示聊天完成选择（非流式）
type Choice struct {
	FinishReason FinishReason `json:"finish_reason"` // 完成原因
	Index        int          `json:"index"`         // 索引
	Logprobs     *Logprobs    `json:"logprobs"`      // 对数概率
	Message      Message      `json:"message"`       // 消息
}

// Logprobs 表示对数概率信息
type Logprobs struct {
	Content *[]TokenLogprob `json:"content"` // 内容对数概率
	Refusal *[]TokenLogprob `json:"refusal"` // 拒绝对数概率
}

// TokenLogprob 表示 token 对数概率
type TokenLogprob struct {
	Token       string                   `json:"token"`        // token
	Bytes       *[]int                   `json:"bytes"`        // 字节表示
	Logprob     float64                  `json:"logprob"`      // 对数概率
	TopLogprobs []TokenLogprobTopLogprob `json:"top_logprobs"` // 顶部对数概率
}

// TokenLogprobTopLogprob 表示顶部 token 对数概率
// bytes 可为 null。
type TokenLogprobTopLogprob struct {
	Token   string  `json:"token"`   // token
	Bytes   *[]int  `json:"bytes"`   // 字节表示
	Logprob float64 `json:"logprob"` // 对数概率
}
