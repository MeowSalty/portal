package types

// CitationMetadata 表示引用信息。
type CitationMetadata struct {
	CitationSources []CitationSource `json:"citationSources,omitempty"` // 引用来源
}

// CitationSource 表示单条引用来源。
type CitationSource struct {
	StartIndex int32  `json:"startIndex,omitempty"` // 起始字节索引
	EndIndex   int32  `json:"endIndex,omitempty"`   // 结束字节索引
	URI        string `json:"uri,omitempty"`        // 引用 URI
	License    string `json:"license,omitempty"`    // 许可
}
