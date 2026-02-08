package types

import (
	"encoding/json"
	"fmt"
)

// TextBlockParam 文本内容块。
type TextBlockParam struct {
	Type         ContentBlockType       `json:"type"`                    // "text"
	Text         string                 `json:"text"`                    // 文本内容
	CacheControl *CacheControlEphemeral `json:"cache_control,omitempty"` // 缓存控制
	Citations    []TextCitationParam    `json:"citations,omitempty"`     // 引用
}

// CitationsConfigParam 引用配置。
type CitationsConfigParam struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// TextCitationType 引用类型。
type TextCitationType string

const (
	TextCitationTypeCharLocation         TextCitationType = "char_location"
	TextCitationTypePageLocation         TextCitationType = "page_location"
	TextCitationTypeContentBlockLocation TextCitationType = "content_block_location"
	TextCitationTypeWebSearchResult      TextCitationType = "web_search_result_location"
	TextCitationTypeSearchResult         TextCitationType = "search_result_location"
)

// CitationCharLocationParam 字符位置引用。
type CitationCharLocationParam struct {
	Type           TextCitationType `json:"type"`
	CitedText      string           `json:"cited_text"`
	DocumentIndex  int              `json:"document_index"`
	DocumentTitle  string           `json:"document_title"`
	StartCharIndex int              `json:"start_char_index"`
	EndCharIndex   int              `json:"end_char_index"`
}

// CitationPageLocationParam 页码位置引用。
type CitationPageLocationParam struct {
	Type            TextCitationType `json:"type"`
	CitedText       string           `json:"cited_text"`
	DocumentIndex   int              `json:"document_index"`
	DocumentTitle   string           `json:"document_title"`
	StartPageNumber int              `json:"start_page_number"`
	EndPageNumber   int              `json:"end_page_number"`
}

// CitationContentBlockLocationParam 内容块位置引用。
type CitationContentBlockLocationParam struct {
	Type            TextCitationType `json:"type"`
	CitedText       string           `json:"cited_text"`
	DocumentIndex   int              `json:"document_index"`
	DocumentTitle   string           `json:"document_title"`
	StartBlockIndex int              `json:"start_block_index"`
	EndBlockIndex   int              `json:"end_block_index"`
}

// CitationWebSearchResultLocationParam Web 搜索结果引用。
type CitationWebSearchResultLocationParam struct {
	Type           TextCitationType `json:"type"`
	CitedText      string           `json:"cited_text"`
	EncryptedIndex string           `json:"encrypted_index"`
	Title          string           `json:"title"`
	URL            string           `json:"url"`
}

// CitationSearchResultLocationParam 搜索结果引用。
type CitationSearchResultLocationParam struct {
	Type              TextCitationType `json:"type"`
	CitedText         string           `json:"cited_text"`
	StartBlockIndex   int              `json:"start_block_index"`
	EndBlockIndex     int              `json:"end_block_index"`
	SearchResultIndex int              `json:"search_result_index"`
	Source            string           `json:"source"`
	Title             string           `json:"title"`
}

// TextCitationParam 引用参数联合类型。
type TextCitationParam struct {
	CharLocation         *CitationCharLocationParam
	PageLocation         *CitationPageLocationParam
	ContentBlockLocation *CitationContentBlockLocationParam
	WebSearchResult      *CitationWebSearchResultLocationParam
	SearchResult         *CitationSearchResultLocationParam
}

// MarshalJSON 实现 TextCitationParam 的序列化。
func (c TextCitationParam) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}
	if c.CharLocation != nil {
		set(c.CharLocation)
	}
	if c.PageLocation != nil {
		set(c.PageLocation)
	}
	if c.ContentBlockLocation != nil {
		set(c.ContentBlockLocation)
	}
	if c.WebSearchResult != nil {
		set(c.WebSearchResult)
	}
	if c.SearchResult != nil {
		set(c.SearchResult)
	}
	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("引用只能设置一种类型")
	}
	return json.Marshal(payload)
}

// UnmarshalJSON 实现 TextCitationParam 的反序列化。
func (c *TextCitationParam) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type TextCitationType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("引用解析失败：%w", err)
	}

	switch t.Type {
	case TextCitationTypeCharLocation:
		var v CitationCharLocationParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("字符位置引用解析失败：%w", err)
		}
		c.CharLocation = &v
	case TextCitationTypePageLocation:
		var v CitationPageLocationParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("页码引用解析失败：%w", err)
		}
		c.PageLocation = &v
	case TextCitationTypeContentBlockLocation:
		var v CitationContentBlockLocationParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("内容块引用解析失败：%w", err)
		}
		c.ContentBlockLocation = &v
	case TextCitationTypeWebSearchResult:
		var v CitationWebSearchResultLocationParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("Web 搜索引用解析失败：%w", err)
		}
		c.WebSearchResult = &v
	case TextCitationTypeSearchResult:
		var v CitationSearchResultLocationParam
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("搜索结果引用解析失败：%w", err)
		}
		c.SearchResult = &v
	default:
		return fmt.Errorf("不支持的引用类型: %s", t.Type)
	}

	return nil
}
