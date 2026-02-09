package converter

import (
	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// convertCitationsToAnnotations 转换引用为注释。
func convertCitationsToAnnotations(citations []anthropicTypes.TextCitation) ([]types.ResponseAnnotation, error) {
	var annotations []types.ResponseAnnotation

	for _, citation := range citations {
		annotation := types.ResponseAnnotation{
			Extras: make(map[string]interface{}),
		}

		if citation.CharLocation != nil {
			annotation.Type = "file_citation"
			annotation.StartIndex = &citation.CharLocation.StartCharIndex
			annotation.EndIndex = &citation.CharLocation.EndCharIndex
			annotation.FileID = &citation.CharLocation.FileID

			// 保存完整的原始结构到 Extras
			if err := SaveVendorExtra("anthropic.citation_type", "char_location", annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 类型失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.cited_text", citation.CharLocation.CitedText, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 文本失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.document_index", citation.CharLocation.DocumentIndex, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 文档索引失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.document_title", citation.CharLocation.DocumentTitle, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 文档标题失败", "error", err)
			}
		} else if citation.PageLocation != nil {
			annotation.Type = "file_citation"
			annotation.FileID = &citation.PageLocation.FileID

			if err := SaveVendorExtra("anthropic.citation_type", "page_location", annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 类型失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.cited_text", citation.PageLocation.CitedText, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 文本失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.document_index", citation.PageLocation.DocumentIndex, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 文档索引失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.document_title", citation.PageLocation.DocumentTitle, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 文档标题失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.start_page_number", citation.PageLocation.StartPageNumber, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 页码起始失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.end_page_number", citation.PageLocation.EndPageNumber, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 页码结束失败", "error", err)
			}
		} else if citation.ContentBlockLocation != nil {
			annotation.Type = "file_citation"
			annotation.FileID = &citation.ContentBlockLocation.FileID

			if err := SaveVendorExtra("anthropic.citation_type", "content_block_location", annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 类型失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.cited_text", citation.ContentBlockLocation.CitedText, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 文本失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.document_index", citation.ContentBlockLocation.DocumentIndex, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 文档索引失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.document_title", citation.ContentBlockLocation.DocumentTitle, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 文档标题失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.start_block_index", citation.ContentBlockLocation.StartBlockIndex, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 块起始失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.end_block_index", citation.ContentBlockLocation.EndBlockIndex, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 块结束失败", "error", err)
			}
		} else if citation.WebSearchResult != nil {
			annotation.Type = "url_citation"
			annotation.URL = &citation.WebSearchResult.URL
			annotation.Title = &citation.WebSearchResult.Title

			if err := SaveVendorExtra("anthropic.citation_type", "web_search_result", annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 类型失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.cited_text", citation.WebSearchResult.CitedText, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 文本失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.encrypted_index", citation.WebSearchResult.EncryptedIndex, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 加密索引失败", "error", err)
			}
		} else if citation.SearchResult != nil {
			annotation.Type = "other"

			if err := SaveVendorExtra("anthropic.citation_type", "search_result", annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 类型失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.cited_text", citation.SearchResult.CitedText, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 文本失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.start_block_index", citation.SearchResult.StartBlockIndex, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 块起始失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.end_block_index", citation.SearchResult.EndBlockIndex, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 块结束失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.search_result_index", citation.SearchResult.SearchResultIndex, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 搜索索引失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.source", citation.SearchResult.Source, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 来源失败", "error", err)
			}
			if err := SaveVendorExtra("anthropic.title", citation.SearchResult.Title, annotation.Extras); err != nil {
				logger.Default().Warn("保存 Citation 标题失败", "error", err)
			}
		}

		annotations = append(annotations, annotation)
	}

	return annotations, nil
}

// convertAnnotationsToCitations 从 Annotations 转换为 Citations。
func convertAnnotationsToCitations(annotations []types.ResponseAnnotation) ([]anthropicTypes.TextCitation, error) {
	var citations []anthropicTypes.TextCitation

	for _, annotation := range annotations {
		citation := anthropicTypes.TextCitation{}

		citationType := ""
		if found, err := GetVendorExtra("anthropic.citation_type", annotation.Extras, &citationType); err != nil {
			logger.Default().Warn("读取 Citation 类型失败", "error", err)
		} else if !found {
			if fallback, ok := annotation.Extras["anthropic.citation_type"].(string); ok {
				citationType = fallback
			}
		}

		switch citationType {
		case "char_location":
			charLoc := &anthropicTypes.CitationCharLocation{
				Type: anthropicTypes.TextCitationTypeCharLocation,
			}
			if annotation.StartIndex != nil {
				charLoc.StartCharIndex = *annotation.StartIndex
			}
			if annotation.EndIndex != nil {
				charLoc.EndCharIndex = *annotation.EndIndex
			}
			if annotation.FileID != nil {
				charLoc.FileID = *annotation.FileID
			}
			if citedText, ok := getStringExtra("anthropic.cited_text", annotation.Extras); ok {
				charLoc.CitedText = citedText
			}
			if docIndex, ok := getIntExtra("anthropic.document_index", annotation.Extras); ok {
				charLoc.DocumentIndex = docIndex
			}
			if docTitle, ok := getStringExtra("anthropic.document_title", annotation.Extras); ok {
				charLoc.DocumentTitle = docTitle
			}
			citation.CharLocation = charLoc

		case "page_location":
			pageLoc := &anthropicTypes.CitationPageLocation{
				Type: anthropicTypes.TextCitationTypePageLocation,
			}
			if annotation.FileID != nil {
				pageLoc.FileID = *annotation.FileID
			}
			if citedText, ok := getStringExtra("anthropic.cited_text", annotation.Extras); ok {
				pageLoc.CitedText = citedText
			}
			if docIndex, ok := getIntExtra("anthropic.document_index", annotation.Extras); ok {
				pageLoc.DocumentIndex = docIndex
			}
			if docTitle, ok := getStringExtra("anthropic.document_title", annotation.Extras); ok {
				pageLoc.DocumentTitle = docTitle
			}
			if startPage, ok := getIntExtra("anthropic.start_page_number", annotation.Extras); ok {
				pageLoc.StartPageNumber = startPage
			}
			if endPage, ok := getIntExtra("anthropic.end_page_number", annotation.Extras); ok {
				pageLoc.EndPageNumber = endPage
			}
			citation.PageLocation = pageLoc

		case "content_block_location":
			blockLoc := &anthropicTypes.CitationContentBlockLocation{
				Type: anthropicTypes.TextCitationTypeContentBlockLocation,
			}
			if annotation.FileID != nil {
				blockLoc.FileID = *annotation.FileID
			}
			if citedText, ok := getStringExtra("anthropic.cited_text", annotation.Extras); ok {
				blockLoc.CitedText = citedText
			}
			if docIndex, ok := getIntExtra("anthropic.document_index", annotation.Extras); ok {
				blockLoc.DocumentIndex = docIndex
			}
			if docTitle, ok := getStringExtra("anthropic.document_title", annotation.Extras); ok {
				blockLoc.DocumentTitle = docTitle
			}
			if startBlock, ok := getIntExtra("anthropic.start_block_index", annotation.Extras); ok {
				blockLoc.StartBlockIndex = startBlock
			}
			if endBlock, ok := getIntExtra("anthropic.end_block_index", annotation.Extras); ok {
				blockLoc.EndBlockIndex = endBlock
			}
			citation.ContentBlockLocation = blockLoc

		case "web_search_result":
			webSearch := &anthropicTypes.CitationWebSearchResultLocation{
				Type: anthropicTypes.TextCitationTypeWebSearchResult,
			}
			if annotation.URL != nil {
				webSearch.URL = *annotation.URL
			}
			if annotation.Title != nil {
				webSearch.Title = *annotation.Title
			}
			if citedText, ok := getStringExtra("anthropic.cited_text", annotation.Extras); ok {
				webSearch.CitedText = citedText
			}
			if encryptedIndex, ok := getStringExtra("anthropic.encrypted_index", annotation.Extras); ok {
				webSearch.EncryptedIndex = encryptedIndex
			}
			citation.WebSearchResult = webSearch

		case "search_result":
			searchResult := &anthropicTypes.CitationSearchResultLocation{
				Type: anthropicTypes.TextCitationTypeSearchResult,
			}
			if citedText, ok := getStringExtra("anthropic.cited_text", annotation.Extras); ok {
				searchResult.CitedText = citedText
			}
			if startBlock, ok := getIntExtra("anthropic.start_block_index", annotation.Extras); ok {
				searchResult.StartBlockIndex = startBlock
			}
			if endBlock, ok := getIntExtra("anthropic.end_block_index", annotation.Extras); ok {
				searchResult.EndBlockIndex = endBlock
			}
			if searchIndex, ok := getIntExtra("anthropic.search_result_index", annotation.Extras); ok {
				searchResult.SearchResultIndex = searchIndex
			}
			if source, ok := getStringExtra("anthropic.source", annotation.Extras); ok {
				searchResult.Source = source
			}
			if title, ok := getStringExtra("anthropic.title", annotation.Extras); ok {
				searchResult.Title = title
			}
			citation.SearchResult = searchResult
		}

		citations = append(citations, citation)
	}

	return citations, nil
}
