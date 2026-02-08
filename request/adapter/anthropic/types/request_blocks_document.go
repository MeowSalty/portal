package types

import (
	"encoding/json"
	"fmt"
)

// DocumentMediaType 文档媒体类型。
type DocumentMediaType string

const (
	DocumentMediaTypePDF       DocumentMediaType = "application/pdf"
	DocumentMediaTypeTextPlain DocumentMediaType = "text/plain"
)

// DocumentSourceType 文档来源类型。
type DocumentSourceType string

const (
	DocumentSourceTypeBase64  DocumentSourceType = "base64"
	DocumentSourceTypeText    DocumentSourceType = "text"
	DocumentSourceTypeContent DocumentSourceType = "content"
	DocumentSourceTypeURL     DocumentSourceType = "url"
)

// Base64PDFSource base64 PDF 来源。
type Base64PDFSource struct {
	Type      DocumentSourceType `json:"type"`       // "base64"
	MediaType DocumentMediaType  `json:"media_type"` // "application/pdf"
	Data      string             `json:"data"`       // base64 数据
}

// PlainTextSource 纯文本来源。
type PlainTextSource struct {
	Type      DocumentSourceType `json:"type"`       // "text"
	MediaType DocumentMediaType  `json:"media_type"` // "text/plain"
	Data      string             `json:"data"`       // 文本内容
}

// ContentBlockSource content 类型来源。
type ContentBlockSource struct {
	Type    DocumentSourceType        `json:"type"`    // "content"
	Content ContentBlockSourceContent `json:"content"` // string 或 []ContentBlockSourceItem
}

// URLPDFSource URL PDF 来源。
type URLPDFSource struct {
	Type DocumentSourceType `json:"type"` // "url"
	URL  string             `json:"url"`  // 文档 URL
}

// DocumentSource 文档来源联合类型。
type DocumentSource struct {
	Base64  *Base64PDFSource
	Text    *PlainTextSource
	Content *ContentBlockSource
	URL     *URLPDFSource
}

// MarshalJSON 实现 DocumentSource 的序列化。
func (s DocumentSource) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}
	if s.Base64 != nil {
		set(s.Base64)
	}
	if s.Text != nil {
		set(s.Text)
	}
	if s.Content != nil {
		set(s.Content)
	}
	if s.URL != nil {
		set(s.URL)
	}
	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("文档来源只能设置一种类型")
	}
	return json.Marshal(payload)
}

// UnmarshalJSON 实现 DocumentSource 的反序列化。
func (s *DocumentSource) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type DocumentSourceType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("文档来源解析失败：%w", err)
	}

	switch t.Type {
	case DocumentSourceTypeBase64:
		var v Base64PDFSource
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("base64 文档来源解析失败：%w", err)
		}
		s.Base64 = &v
	case DocumentSourceTypeText:
		var v PlainTextSource
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("文本文档来源解析失败：%w", err)
		}
		s.Text = &v
	case DocumentSourceTypeContent:
		var v ContentBlockSource
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("content 文档来源解析失败：%w", err)
		}
		s.Content = &v
	case DocumentSourceTypeURL:
		var v URLPDFSource
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("URL 文档来源解析失败：%w", err)
		}
		s.URL = &v
	default:
		return fmt.Errorf("不支持的文档来源类型: %s", t.Type)
	}

	return nil
}

// DocumentBlockParam 文档内容块。
type DocumentBlockParam struct {
	Type         ContentBlockType       `json:"type"`                    // "document"
	Source       DocumentSource         `json:"source"`                  // 文档来源
	CacheControl *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
	Citations    *CitationsConfigParam  `json:"citations,omitempty"`     // 引用配置
	Context      *string                `json:"context,omitempty"`       // 上下文
	Title        *string                `json:"title,omitempty"`         // 标题
}

// ContentBlockSourceItem content 类型允许的块（文本或图片）。
type ContentBlockSourceItem struct {
	Text  *TextBlockParam
	Image *ImageBlockParam
}

// MarshalJSON 实现 ContentBlockSourceItem 的序列化。
func (c ContentBlockSourceItem) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}
	if c.Text != nil {
		set(c.Text)
	}
	if c.Image != nil {
		set(c.Image)
	}
	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("content 块只能是文本或图片")
	}
	return json.Marshal(payload)
}

// UnmarshalJSON 实现 ContentBlockSourceItem 的反序列化。
func (c *ContentBlockSourceItem) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type ContentBlockType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("content 块解析失败：%w", err)
	}

	switch t.Type {
	case ContentBlockTypeText:
		var v TextBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("content 文本块解析失败：%w", err)
		}
		c.Text = &v
	case ContentBlockTypeImage:
		var v ImageBlockParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("content 图片块解析失败：%w", err)
		}
		c.Image = &v
	default:
		return fmt.Errorf("content 块类型只允许 text 或 image")
	}

	return nil
}

// ContentBlockSourceContent 表示 string 或 []ContentBlockSourceItem。
type ContentBlockSourceContent struct {
	StringValue *string
	Blocks      []ContentBlockSourceItem
}

// MarshalJSON 实现 ContentBlockSourceContent 的序列化。
func (c ContentBlockSourceContent) MarshalJSON() ([]byte, error) {
	if c.StringValue != nil {
		return json.Marshal(c.StringValue)
	}
	if c.Blocks != nil {
		return json.Marshal(c.Blocks)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 ContentBlockSourceContent 的反序列化。
func (c *ContentBlockSourceContent) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		c.StringValue = &str
		return nil
	}

	var blocks []ContentBlockSourceItem
	if err := json.Unmarshal(data, &blocks); err == nil {
		c.Blocks = blocks
		return nil
	}

	return fmt.Errorf("content 只允许 string 或内容块数组")
}
