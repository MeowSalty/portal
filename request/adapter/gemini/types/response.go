package types

// Response 表示 Gemini API 的 GenerateContentResponse 响应结构。
type Response struct {
	Candidates     []Candidate     `json:"candidates"`               // 候选响应
	PromptFeedback *PromptFeedback `json:"promptFeedback,omitempty"` // 提示反馈
	UsageMetadata  *UsageMetadata  `json:"usageMetadata,omitempty"`  // 使用情况元数据
	ModelVersion   string          `json:"modelVersion,omitempty"`   // 模型版本
	ResponseID     string          `json:"responseId,omitempty"`     // 响应 ID
	ModelStatus    *ModelStatus    `json:"modelStatus,omitempty"`    // 模型状态
}

// Candidate 表示候选响应。
type Candidate struct {
	Content               Content                `json:"content"`                         // 内容
	FinishReason          string                 `json:"finishReason,omitempty"`          // 完成原因
	FinishMessage         string                 `json:"finishMessage,omitempty"`         // 完成原因详细说明
	Index                 int                    `json:"index,omitempty"`                 // 索引
	SafetyRatings         []SafetyRating         `json:"safetyRatings,omitempty"`         // 安全评级
	CitationMetadata      *CitationMetadata      `json:"citationMetadata,omitempty"`      // 引用信息
	TokenCount            int                    `json:"tokenCount,omitempty"`            // 候选 token 数
	GroundingAttributions []GroundingAttribution `json:"groundingAttributions,omitempty"` // 归因信息
	GroundingMetadata     *GroundingMetadata     `json:"groundingMetadata,omitempty"`     // 归因元数据
	AvgLogprobs           *float64               `json:"avgLogprobs,omitempty"`           // 平均对数概率
	LogprobsResult        *LogprobsResult        `json:"logprobsResult,omitempty"`        // 对数概率详情
	URLContextMetadata    *URLContextMetadata    `json:"urlContextMetadata,omitempty"`    // URL 上下文元数据
}

// PromptFeedback 表示提示反馈。
type PromptFeedback struct {
	BlockReason   string         `json:"blockReason,omitempty"`   // 阻止原因
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"` // 安全评级
}

// UsageMetadata 表示使用情况元数据。
type UsageMetadata struct {
	PromptTokenCount           int                  `json:"promptTokenCount"`                     // 提示 token 计数
	CachedContentTokenCount    int                  `json:"cachedContentTokenCount,omitempty"`    // 缓存 token 计数
	CandidatesTokenCount       int                  `json:"candidatesTokenCount"`                 // 候选 token 计数
	ToolUsePromptTokenCount    int                  `json:"toolUsePromptTokenCount,omitempty"`    // 工具调用提示 token 计数
	ThoughtsTokenCount         int                  `json:"thoughtsTokenCount,omitempty"`         // 思考 token 计数
	TotalTokenCount            int                  `json:"totalTokenCount"`                      // 总 token 计数
	PromptTokensDetails        []ModalityTokenCount `json:"promptTokensDetails,omitempty"`        // 提示 token 详情
	CacheTokensDetails         []ModalityTokenCount `json:"cacheTokensDetails,omitempty"`         // 缓存 token 详情
	CandidatesTokensDetails    []ModalityTokenCount `json:"candidatesTokensDetails,omitempty"`    // 候选 token 详情
	ToolUsePromptTokensDetails []ModalityTokenCount `json:"toolUsePromptTokensDetails,omitempty"` // 工具调用提示 token 详情
}

// ModalityTokenCount 表示单一模态的 token 计数。
type ModalityTokenCount struct {
	Modality   string `json:"modality"`   // 模态
	TokenCount int    `json:"tokenCount"` // token 数
}

// SafetyRating 表示安全评级。
type SafetyRating struct {
	Category    string `json:"category"`    // 安全类别
	Probability string `json:"probability"` // 概率
	Blocked     bool   `json:"blocked"`     // 是否被阻止
}

// CitationMetadata 表示引用信息。
type CitationMetadata struct {
	CitationSources []CitationSource `json:"citationSources,omitempty"` // 引用来源
}

// CitationSource 表示单条引用来源。
type CitationSource struct {
	StartIndex int    `json:"startIndex,omitempty"` // 起始字节索引
	EndIndex   int    `json:"endIndex,omitempty"`   // 结束字节索引
	URI        string `json:"uri,omitempty"`        // 引用 URI
	License    string `json:"license,omitempty"`    // 许可
}

// GroundingAttribution 表示归因信息。
type GroundingAttribution struct {
	SourceID *AttributionSourceID `json:"sourceId,omitempty"` // 来源 ID
	Content  *Content             `json:"content,omitempty"`  // 来源内容
}

// AttributionSourceID 表示归因来源标识。
type AttributionSourceID struct {
	GroundingPassage       *GroundingPassageID     `json:"groundingPassage,omitempty"`       // 段落来源
	SemanticRetrieverChunk *SemanticRetrieverChunk `json:"semanticRetrieverChunk,omitempty"` // 检索块来源
}

// GroundingPassageID 表示归因段落 ID。
type GroundingPassageID struct {
	PassageID string `json:"passageId,omitempty"` // 段落 ID
	PartIndex int    `json:"partIndex,omitempty"` // Part 索引
}

// SemanticRetrieverChunk 表示检索块来源。
type SemanticRetrieverChunk struct {
	Source string `json:"source,omitempty"` // 来源
	Chunk  string `json:"chunk,omitempty"`  // 块标识
}

// GroundingMetadata 表示归因元数据。
type GroundingMetadata struct {
	SearchEntryPoint             *SearchEntryPoint  `json:"searchEntryPoint,omitempty"`             // 搜索入口
	GroundingChunks              []GroundingChunk   `json:"groundingChunks,omitempty"`              // 归因块
	GroundingSupports            []GroundingSupport `json:"groundingSupports,omitempty"`            // 归因支持
	RetrievalMetadata            *RetrievalMetadata `json:"retrievalMetadata,omitempty"`            // 检索元数据
	WebSearchQueries             []string           `json:"webSearchQueries,omitempty"`             // 搜索查询
	GoogleMapsWidgetContextToken string             `json:"googleMapsWidgetContextToken,omitempty"` // 地图上下文 token
}

// SearchEntryPoint 表示搜索入口。
type SearchEntryPoint struct {
	RenderedContent string `json:"renderedContent,omitempty"` // 渲染内容
	SDKBlob         string `json:"sdkBlob,omitempty"`         // SDK 数据
}

// GroundingChunk 表示归因块。
type GroundingChunk struct {
	Web              *Web              `json:"web,omitempty"`              // Web 来源
	RetrievedContext *RetrievedContext `json:"retrievedContext,omitempty"` // 检索上下文
	Maps             *Maps             `json:"maps,omitempty"`             // 地图来源
}

// Web 表示 Web 归因块。
type Web struct {
	URI   string `json:"uri,omitempty"`   // URI
	Title string `json:"title,omitempty"` // 标题
}

// RetrievedContext 表示检索上下文。
type RetrievedContext struct {
	URI             string `json:"uri,omitempty"`             // URI
	Title           string `json:"title,omitempty"`           // 标题
	Text            string `json:"text,omitempty"`            // 文本
	FileSearchStore string `json:"fileSearchStore,omitempty"` // 文件检索存储
}

// Maps 表示地图归因块。
type Maps struct {
	URI                string              `json:"uri,omitempty"`                // URI
	Title              string              `json:"title,omitempty"`              // 标题
	Text               string              `json:"text,omitempty"`               // 文本
	PlaceID            string              `json:"placeId,omitempty"`            // 地点 ID
	PlaceAnswerSources *PlaceAnswerSources `json:"placeAnswerSources,omitempty"` // 地点来源
}

// PlaceAnswerSources 表示地点来源集合。
type PlaceAnswerSources struct {
	ReviewSnippets []ReviewSnippet `json:"reviewSnippets,omitempty"` // 评价片段
}

// ReviewSnippet 表示评价片段。
type ReviewSnippet struct {
	ReviewID      string `json:"reviewId,omitempty"`      // 评价 ID
	GoogleMapsURI string `json:"googleMapsUri,omitempty"` // 地图链接
	Title         string `json:"title,omitempty"`         // 标题
}

// GroundingSupport 表示归因支持。
type GroundingSupport struct {
	Segment               *Segment  `json:"segment,omitempty"`               // 内容片段
	GroundingChunkIndices []int     `json:"groundingChunkIndices,omitempty"` // 归因块索引
	ConfidenceScores      []float32 `json:"confidenceScores,omitempty"`      // 置信分数
}

// Segment 表示内容片段。
type Segment struct {
	PartIndex  int    `json:"partIndex,omitempty"`  // Part 索引
	StartIndex int    `json:"startIndex,omitempty"` // 起始字节索引
	EndIndex   int    `json:"endIndex,omitempty"`   // 结束字节索引
	Text       string `json:"text,omitempty"`       // 文本
}

// RetrievalMetadata 表示检索元数据。
type RetrievalMetadata struct {
	GoogleSearchDynamicRetrievalScore float32 `json:"googleSearchDynamicRetrievalScore,omitempty"` // 动态检索分数
}

// LogprobsResult 表示对数概率结果。
type LogprobsResult struct {
	LogProbabilitySum float32                   `json:"logProbabilitySum,omitempty"` // 对数概率和
	TopCandidates     []TopCandidates           `json:"topCandidates,omitempty"`     // Top 候选
	ChosenCandidates  []LogprobsResultCandidate `json:"chosenCandidates,omitempty"`  // 选中候选
}

// TopCandidates 表示解码步的候选集合。
type TopCandidates struct {
	Candidates []LogprobsResultCandidate `json:"candidates,omitempty"` // 候选
}

// LogprobsResultCandidate 表示对数概率候选。
type LogprobsResultCandidate struct {
	Token          string  `json:"token,omitempty"`          // token
	TokenID        int     `json:"tokenId,omitempty"`        // token ID
	LogProbability float32 `json:"logProbability,omitempty"` // 对数概率
}

// URLContextMetadata 表示 URL 上下文元数据。
type URLContextMetadata struct {
	URLMetadata []URLMetadata `json:"urlMetadata,omitempty"` // URL 元数据
}

// URLMetadata 表示 URL 元数据。
type URLMetadata struct {
	RetrievedURL       string `json:"retrievedUrl,omitempty"`       // URL
	URLRetrievalStatus string `json:"urlRetrievalStatus,omitempty"` // 检索状态
}

// ModelStatus 表示模型状态。
type ModelStatus struct {
	ModelStage     string `json:"modelStage,omitempty"`     // 模型阶段
	RetirementTime string `json:"retirementTime,omitempty"` // 退役时间
	Message        string `json:"message,omitempty"`        // 状态信息
}

// ErrorResponse 表示错误响应。
type ErrorResponse struct {
	Error ErrorDetail `json:"error"` // 错误详情
}

// ErrorDetail 表示错误详情。
type ErrorDetail struct {
	Code    int                      `json:"code"`              // 错误代码
	Message string                   `json:"message"`           // 错误消息
	Status  string                   `json:"status"`            // 状态
	Details []map[string]interface{} `json:"details,omitempty"` // 详细信息
}
