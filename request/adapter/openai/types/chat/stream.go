package chat

// ResponseChunk 表示 OpenAI 聊天完成流式响应
// 与非流式响应相比，choice 的内容为 delta。
type ResponseChunk struct {
	ID                string         `json:"id"`                           // 聊天完成 ID
	Choices           []StreamChoice `json:"choices"`                      // 选择列表
	Created           int64          `json:"created"`                      // 创建时间戳
	Model             string         `json:"model"`                        // 模型名称
	Object            string         `json:"object"`                       // 对象类型
	ServiceTier       *string        `json:"service_tier,omitempty"`       // 服务层级
	SystemFingerprint *string        `json:"system_fingerprint,omitempty"` // 系统指纹
	Usage             *Usage         `json:"usage"`                        // 使用情况
}

// StreamChoice 表示聊天完成选择（流式）
type StreamChoice struct {
	FinishReason *string   `json:"finish_reason"`   // 完成原因
	Index        int       `json:"index"`           // 索引
	Logprobs     *Logprobs `json:"logprobs"`        // 对数概率
	Delta        *Delta    `json:"delta,omitempty"` // 增量消息
}
