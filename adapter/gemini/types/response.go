package types

// Response 表示 Gemini API 响应结构
type Response struct {
	Candidates     []Candidate     `json:"candidates"`               // 候选响应
	PromptFeedback *PromptFeedback `json:"promptFeedback,omitempty"` // 提示反馈
	UsageMetadata  *UsageMetadata  `json:"usageMetadata,omitempty"`  // 使用情况元数据
	ModelVersion   string          `json:"modelVersion,omitempty"`   // 模型版本
	ResponseID     string          `json:"responseId,omitempty"`     // 响应 ID
}

// Candidate 表示候选响应
type Candidate struct {
	Content       Content        `json:"content"`                 // 内容
	FinishReason  string         `json:"finishReason,omitempty"`  // 完成原因
	Index         int            `json:"index,omitempty"`         // 索引
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"` // 安全评级
}

// PromptFeedback 表示提示反馈
type PromptFeedback struct {
	BlockReason   string         `json:"blockReason,omitempty"`   // 阻止原因
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"` // 安全评级
}

// UsageMetadata 表示使用情况元数据
type UsageMetadata struct {
	PromptTokenCount        int            `json:"promptTokenCount"`                  // 提示 token 计数
	CandidatesTokenCount    int            `json:"candidatesTokenCount"`              // 候选 token 计数
	TotalTokenCount         int            `json:"totalTokenCount"`                   // 总 token 计数
	PromptTokensDetails     []TokenDetails `json:"promptTokensDetails,omitempty"`     // 提示 token 详情
	CandidatesTokensDetails []TokenDetails `json:"candidatesTokensDetails,omitempty"` // 候选 token 详情
}

// TokenDetails 表示 token 详情
type TokenDetails struct {
	TotalTokens int `json:"totalTokens"` // 总 token 数
}

// SafetyRating 表示安全评级
type SafetyRating struct {
	Category    string `json:"category"`    // 安全类别
	Probability string `json:"probability"` // 概率
	Blocked     bool   `json:"blocked"`     // 是否被阻止
}

// ErrorResponse 表示错误响应
type ErrorResponse struct {
	Error ErrorDetail `json:"error"` // 错误详情
}

// ErrorDetail 表示错误详情
type ErrorDetail struct {
	Code    int                      `json:"code"`              // 错误代码
	Message string                   `json:"message"`           // 错误消息
	Status  string                   `json:"status"`            // 状态
	Details []map[string]interface{} `json:"details,omitempty"` // 详细信息
}
