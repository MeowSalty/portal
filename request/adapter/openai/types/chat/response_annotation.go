package chat

// MessageAnnotation 表示消息注释
// 当前仅支持 URL 引用。
type MessageAnnotation struct {
	Type        string       `json:"type"`
	URLCitation *URLCitation `json:"url_citation,omitempty"`
}

// URLCitation 表示 URL 引用内容
// 用于 web search 注释。
type URLCitation struct {
	EndIndex   int    `json:"end_index"`
	StartIndex int    `json:"start_index"`
	URL        string `json:"url"`
	Title      string `json:"title"`
}
