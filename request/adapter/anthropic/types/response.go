package types

// Response Anthropic API 非流式响应结构
// 对应文档中的 Message。
type Response struct {
	ID           string                 `json:"id"`              // 消息 ID
	Type         ResponseType           `json:"type"`            // "message"
	Role         Role                   `json:"role"`            // "assistant"
	Content      []ResponseContentBlock `json:"content"`         // 内容块数组
	Model        string                 `json:"model"`           // 使用的模型
	StopReason   *StopReason            `json:"stop_reason"`     // 停止原因（流式 message_start 可能为 null）
	StopSequence *string                `json:"stop_sequence"`   // 停止序列
	Usage        *Usage                 `json:"usage,omitempty"` // 使用统计
}
