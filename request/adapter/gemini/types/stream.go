package types

// StreamResponse 表示 Gemini API 的 streamGenerateContent 流式响应块。
// 文档中流式块与非流式响应结构一致，但以多段 GenerateContentResponse 形式返回。
type StreamResponse struct {
	Candidates     []Candidate     `json:"candidates"`               // 候选响应
	PromptFeedback *PromptFeedback `json:"promptFeedback,omitempty"` // 提示反馈
	UsageMetadata  *UsageMetadata  `json:"usageMetadata,omitempty"`  // 使用情况元数据
	ModelVersion   string          `json:"modelVersion,omitempty"`   // 模型版本
	ResponseID     string          `json:"responseId,omitempty"`     // 响应 ID
	ModelStatus    *ModelStatus    `json:"modelStatus,omitempty"`    // 模型状态
}
