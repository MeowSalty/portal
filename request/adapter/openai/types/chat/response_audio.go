package chat

// ResponseAudio 表示音频响应
type ResponseAudio struct {
	ID         string `json:"id"`         // 音频 ID
	Data       string `json:"data"`       // 音频数据
	ExpiresAt  int    `json:"expires_at"` // 过期时间
	Transcript string `json:"transcript"` // 转录文本
}
