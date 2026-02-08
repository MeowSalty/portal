package chat

import (
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

// StreamObject 表示流式响应对象类型
// 该值固定为 chat.completion.chunk。
type StreamObject string

const (
	StreamObjectChatCompletionChunk StreamObject = "chat.completion.chunk"
)

// StreamEvent 表示 OpenAI 聊天完成流式响应
// 与非流式响应相比，choice 的内容为 delta。
type StreamEvent struct {
	ID                string              `json:"id"`                           // 聊天完成 ID
	Choices           []StreamChoice      `json:"choices"`                      // 选择列表
	Created           int64               `json:"created"`                      // 创建时间戳
	Model             string              `json:"model"`                        // 模型名称
	Object            StreamObject        `json:"object"`                       // 对象类型
	ServiceTier       *shared.ServiceTier `json:"service_tier,omitempty"`       // 服务层级
	SystemFingerprint *string             `json:"system_fingerprint,omitempty"` // 系统指纹
	Usage             *Usage              `json:"usage,omitempty"`              // 使用情况
}
