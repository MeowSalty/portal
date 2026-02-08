package chat

// Usage 表示使用情况
type Usage struct {
	PromptTokens            int                      `json:"prompt_tokens"`                       // 提示 token 数
	CompletionTokens        int                      `json:"completion_tokens"`                   // 完成 token 数
	TotalTokens             int                      `json:"total_tokens"`                        // 总 token 数
	PromptTokensDetails     *PromptTokensDetails     `json:"prompt_tokens_details,omitempty"`     // 提示 token 细节
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details,omitempty"` // 完成 token 细节
}

// PromptTokensDetails 表示提示 token 细节
// 与 CompletionUsage 中 prompt_tokens_details 对齐。
type PromptTokensDetails struct {
	AudioTokens  int `json:"audio_tokens,omitempty"`  // 音频 token 数
	CachedTokens int `json:"cached_tokens,omitempty"` // 缓存 token 数
}

// CompletionTokensDetails 表示完成 token 细节
// 与 CompletionUsage 中 completion_tokens_details 对齐。
type CompletionTokensDetails struct {
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens,omitempty"` // 采纳预测 token 数
	AudioTokens              int `json:"audio_tokens,omitempty"`               // 音频 token 数
	ReasoningTokens          int `json:"reasoning_tokens,omitempty"`           // 推理 token 数
	RejectedPredictionTokens int `json:"rejected_prediction_tokens,omitempty"` // 拒绝预测 token 数
}
