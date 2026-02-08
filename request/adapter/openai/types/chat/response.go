package chat

import (
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

// Response 表示 OpenAI 聊天完成响应（非流式）
type Response struct {
	ID                string              `json:"id"`                           // 聊天完成 ID
	Choices           []Choice            `json:"choices"`                      // 选择列表
	Created           int64               `json:"created"`                      // 创建时间戳
	Model             string              `json:"model"`                        // 模型名称
	Object            string              `json:"object"`                       // 对象类型
	ServiceTier       *shared.ServiceTier `json:"service_tier,omitempty"`       // 服务层级
	SystemFingerprint *string             `json:"system_fingerprint,omitempty"` // 系统指纹
	Usage             *Usage              `json:"usage,omitempty"`              // 使用情况
}
