package types

// UsageMetadata 表示使用情况元数据。
type UsageMetadata struct {
	PromptTokenCount           int32                `json:"promptTokenCount"`                     // 提示 token 计数
	CachedContentTokenCount    int32                `json:"cachedContentTokenCount,omitempty"`    // 缓存 token 计数
	CandidatesTokenCount       int32                `json:"candidatesTokenCount"`                 // 候选 token 计数
	ToolUsePromptTokenCount    int32                `json:"toolUsePromptTokenCount,omitempty"`    // 工具调用提示 token 计数
	ThoughtsTokenCount         int32                `json:"thoughtsTokenCount,omitempty"`         // 思考 token 计数
	TotalTokenCount            int32                `json:"totalTokenCount"`                      // 总 token 计数
	PromptTokensDetails        []ModalityTokenCount `json:"promptTokensDetails,omitempty"`        // 提示 token 详情
	CacheTokensDetails         []ModalityTokenCount `json:"cacheTokensDetails,omitempty"`         // 缓存 token 详情
	CandidatesTokensDetails    []ModalityTokenCount `json:"candidatesTokensDetails,omitempty"`    // 候选 token 详情
	ToolUsePromptTokensDetails []ModalityTokenCount `json:"toolUsePromptTokensDetails,omitempty"` // 工具调用提示 token 详情
}

// ModalityTokenCount 表示单一模态的 token 计数。
type ModalityTokenCount struct {
	Modality   Modality `json:"modality"`   // 模态
	TokenCount int32    `json:"tokenCount"` // token 数
}
