package types

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
	PartIndex int32  `json:"partIndex,omitempty"` // Part 索引
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
	GroundingChunkIndices []int32   `json:"groundingChunkIndices,omitempty"` // 归因块索引
	ConfidenceScores      []float32 `json:"confidenceScores,omitempty"`      // 置信分数
}

// Segment 表示内容片段。
type Segment struct {
	PartIndex  int32  `json:"partIndex,omitempty"`  // Part 索引
	StartIndex int32  `json:"startIndex,omitempty"` // 起始字节索引
	EndIndex   int32  `json:"endIndex,omitempty"`   // 结束字节索引
	Text       string `json:"text,omitempty"`       // 文本
}

// RetrievalMetadata 表示检索元数据。
type RetrievalMetadata struct {
	GoogleSearchDynamicRetrievalScore float32 `json:"googleSearchDynamicRetrievalScore,omitempty"` // 动态检索分数
}
