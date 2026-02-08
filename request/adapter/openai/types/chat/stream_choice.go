package chat

// StreamChoice 表示聊天完成选择（流式）
type StreamChoice struct {
	FinishReason *FinishReason `json:"finish_reason"`      // 完成原因
	Index        int           `json:"index"`              // 索引
	Logprobs     *Logprobs     `json:"logprobs,omitempty"` // 对数概率
	Delta        Delta         `json:"delta"`              // 增量消息
}
