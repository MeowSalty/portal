package types

// URLContextMetadata 表示 URL 上下文元数据。
type URLContextMetadata struct {
	URLMetadata []URLMetadata `json:"urlMetadata,omitempty"` // URL 元数据
}

// URLMetadata 表示 URL 元数据。
type URLMetadata struct {
	RetrievedURL       string             `json:"retrievedUrl,omitempty"`       // URL
	URLRetrievalStatus URLRetrievalStatus `json:"urlRetrievalStatus,omitempty"` // 检索状态
}
