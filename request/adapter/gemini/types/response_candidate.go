package types

// Candidate 表示候选响应。
type Candidate struct {
	Content               Content                `json:"content"`                         // 内容
	FinishReason          FinishReason           `json:"finishReason,omitempty"`          // 完成原因
	FinishMessage         string                 `json:"finishMessage,omitempty"`         // 完成原因详细说明
	Index                 int32                  `json:"index,omitempty"`                 // 索引
	SafetyRatings         []SafetyRating         `json:"safetyRatings,omitempty"`         // 安全评级
	CitationMetadata      *CitationMetadata      `json:"citationMetadata,omitempty"`      // 引用信息
	TokenCount            int32                  `json:"tokenCount,omitempty"`            // 候选 token 数
	GroundingAttributions []GroundingAttribution `json:"groundingAttributions,omitempty"` // 归因信息
	GroundingMetadata     *GroundingMetadata     `json:"groundingMetadata,omitempty"`     // 归因元数据
	AvgLogprobs           *float64               `json:"avgLogprobs,omitempty"`           // 平均对数概率
	LogprobsResult        *LogprobsResult        `json:"logprobsResult,omitempty"`        // 对数概率详情
	URLContextMetadata    *URLContextMetadata    `json:"urlContextMetadata,omitempty"`    // URL 上下文元数据
}

// PromptFeedback 表示提示反馈。
type PromptFeedback struct {
	BlockReason   BlockReason    `json:"blockReason,omitempty"`   // 阻止原因
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"` // 安全评级
}
