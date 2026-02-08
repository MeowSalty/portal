package types

import (
	"encoding/json"
	"fmt"
)

// CitationCharLocation 字符位置引用。
type CitationCharLocation struct {
	Type           TextCitationType `json:"type"`
	CitedText      string           `json:"cited_text"`
	DocumentIndex  int              `json:"document_index"`
	DocumentTitle  string           `json:"document_title"`
	FileID         string           `json:"file_id"`
	StartCharIndex int              `json:"start_char_index"`
	EndCharIndex   int              `json:"end_char_index"`
}

// CitationPageLocation 页码位置引用。
type CitationPageLocation struct {
	Type            TextCitationType `json:"type"`
	CitedText       string           `json:"cited_text"`
	DocumentIndex   int              `json:"document_index"`
	DocumentTitle   string           `json:"document_title"`
	FileID          string           `json:"file_id"`
	StartPageNumber int              `json:"start_page_number"`
	EndPageNumber   int              `json:"end_page_number"`
}

// CitationContentBlockLocation 内容块位置引用。
type CitationContentBlockLocation struct {
	Type            TextCitationType `json:"type"`
	CitedText       string           `json:"cited_text"`
	DocumentIndex   int              `json:"document_index"`
	DocumentTitle   string           `json:"document_title"`
	FileID          string           `json:"file_id"`
	StartBlockIndex int              `json:"start_block_index"`
	EndBlockIndex   int              `json:"end_block_index"`
}

// CitationWebSearchResultLocation Web 搜索结果引用。
type CitationWebSearchResultLocation struct {
	Type           TextCitationType `json:"type"`
	CitedText      string           `json:"cited_text"`
	EncryptedIndex string           `json:"encrypted_index"`
	Title          string           `json:"title"`
	URL            string           `json:"url"`
}

// CitationSearchResultLocation 搜索结果引用。
type CitationSearchResultLocation struct {
	Type              TextCitationType `json:"type"`
	CitedText         string           `json:"cited_text"`
	StartBlockIndex   int              `json:"start_block_index"`
	EndBlockIndex     int              `json:"end_block_index"`
	SearchResultIndex int              `json:"search_result_index"`
	Source            string           `json:"source"`
	Title             string           `json:"title"`
}

// TextCitation 文本引用联合类型。
type TextCitation struct {
	CharLocation         *CitationCharLocation
	PageLocation         *CitationPageLocation
	ContentBlockLocation *CitationContentBlockLocation
	WebSearchResult      *CitationWebSearchResultLocation
	SearchResult         *CitationSearchResultLocation
}

// MarshalJSON 实现 TextCitation 的序列化。
func (c TextCitation) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON 实现 TextCitation 的反序列化。
func (c *TextCitation) UnmarshalJSON(data []byte) error {
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
		var v CitationCharLocation
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("字符位置引用解析失败：%w", err)
		}
		c.CharLocation = &v
	case TextCitationTypePageLocation:
		var v CitationPageLocation
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("页码引用解析失败：%w", err)
		}
		c.PageLocation = &v
	case TextCitationTypeContentBlockLocation:
		var v CitationContentBlockLocation
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("内容块引用解析失败：%w", err)
		}
		c.ContentBlockLocation = &v
	case TextCitationTypeWebSearchResult:
		var v CitationWebSearchResultLocation
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("Web 搜索引用解析失败：%w", err)
		}
		c.WebSearchResult = &v
	case TextCitationTypeSearchResult:
		var v CitationSearchResultLocation
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("搜索结果引用解析失败：%w", err)
		}
		c.SearchResult = &v
	default:
		return fmt.Errorf("不支持的引用类型: %s", t.Type)
	}

	return nil
}
