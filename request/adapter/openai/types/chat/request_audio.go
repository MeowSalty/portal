package chat

// RequestAudio 表示音频参数
type RequestAudio struct {
	Format AudioFormat `json:"format"` // 格式
	Voice  string      `json:"voice"`  // 声音
}

// AssistantAudio 表示助手音频引用
type AssistantAudio struct {
	ID string `json:"id"` // 音频 ID
}
