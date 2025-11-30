package adapter

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/MeowSalty/portal/errors"
)

// HTTPErrorType 表示 HTTP 错误类型
type HTTPErrorType int

const (
	// HTTPErrorTypeAPIError API 服务器正常接收请求但返回错误
	HTTPErrorTypeAPIError HTTPErrorType = iota
	// HTTPErrorTypeServiceUnavailable API 服务器无法正常接收请求
	HTTPErrorTypeServiceUnavailable
)

// isJSONContentType 检查 Content-Type 是否为 JSON 类型
func isJSONContentType(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/json")
}

// handleHTTPError 处理 HTTP 错误
func (a *Adapter) handleHTTPError(message string, statusCode int, contentType string, body []byte) error {
	if len(body) == 0 {
		return a.createHTTPError(message, statusCode, "", HTTPErrorTypeServiceUnavailable)
	}

	var bodyStr string
	var errorType HTTPErrorType

	// 根据 Content-Type 决定处理方式
	if isJSONContentType(contentType) {
		// Content-Type 表明是 JSON，尝试解析
		var jsonData map[string]interface{}
		err := json.Unmarshal(body, &jsonData)

		if err != nil {
			// JSON 解析失败，回退到 HTML/文本处理
			bodyStr = stripHTML(string(body))
			errorType = HTTPErrorTypeServiceUnavailable
		} else {
			// JSON 解析成功
			bodyStr = a.processBodyHTML(jsonData)
			errorType = a.classifyHTTPErrorType(jsonData)
		}
	} else {
		// 非 JSON 类型，直接处理为 HTML/文本
		bodyStr = stripHTML(string(body))
		errorType = HTTPErrorTypeServiceUnavailable
	}

	// 根据错误类型和状态码处理错误
	return a.createHTTPError(message, statusCode, bodyStr, errorType)
}

// processBodyHTML 处理已解析的 JSON 数据中的 HTML 内容
// 如果包含 error.message 字段，则清理该字段中的 HTML 并重新序列化
func (a *Adapter) processBodyHTML(jsonData map[string]interface{}) string {
	// 检查是否存在 error.message 字段
	if errorObj, ok := jsonData["error"].(map[string]interface{}); ok {
		if message, ok := errorObj["message"].(string); ok {
			// 清理 message 中的 HTML 内容
			errorObj["message"] = stripHTML(message)
		}
	}

	// 重新序列化 JSON
	cleanedBody, err := json.Marshal(jsonData)
	if err != nil {
		// 序列化失败，返回空字符串
		return ""
	}

	return string(cleanedBody)
}

// classifyHTTPErrorType 根据已解析的 JSON 数据分类 HTTP 错误类型
func (a *Adapter) classifyHTTPErrorType(jsonData map[string]interface{}) HTTPErrorType {
	// 检查是否存在 error.type 字段
	if errorObj, ok := jsonData["error"].(map[string]interface{}); ok {
		if errorType, ok := errorObj["type"].(string); ok && errorType != "" {
			return HTTPErrorTypeAPIError
		}
	}

	return HTTPErrorTypeServiceUnavailable
}

// createHTTPError 根据错误类型和状态码创建适当的错误
func (a *Adapter) createHTTPError(message string, statusCode int, bodyStr string, errorType HTTPErrorType) error {
	// 定义错误码
	var errCode errors.ErrorCode

	switch {
	case statusCode == 401:
		// 401 状态码表示认证失败 (密钥问题)
		errCode = errors.ErrCodeAuthenticationFailed
	case errorType == HTTPErrorTypeAPIError:
		// API 服务器正常接收请求但返回错误
		errCode = errors.ErrCodeRequestFailed
	default:
		// API 服务器无法正常接收请求
		errCode = errors.ErrCodeUnavailable
	}

	return errors.NewWithHTTPStatus(errCode, message, statusCode).
		WithContext("response_body", bodyStr)
}

// handleParseError 处理解析错误
func (a *Adapter) handleParseError(operation string, err error, body []byte) error {
	return errors.Wrap(errors.ErrCodeInternal, "解析响应失败", err).
		WithContext("operation", operation).
		WithContext("response_body", string(body))
}

// extractHTMLError 从 HTML 页面中提取有效的错误信息
//
// 该函数会尝试从 HTML 内容中提取以下信息：
// - 页面标题（<title> 标签）
// - 主标题（<h1> 标签）
// - 错误描述（<p> 标签中的文本）
//
// 如果内容不是 HTML 或无法提取有效信息，则返回原始内容（经过基本清理）
//
// 参数：
//   - content: 可能包含 HTML 的字符串内容
//
// 返回：
//   - string: 提取的错误信息，格式为 "标题：描述" 或清理后的原始内容
func extractHTMLError(content string) string {
	if content == "" {
		return content
	}

	// 提取标题信息（按优先级）
	title := extractTagContent(content, "title")
	h1 := extractTagContent(content, "h1")
	h2 := extractTagContent(content, "h2")

	// 提取描述信息（按优先级）
	pContent := extractTagContent(content, "p")
	centerContent := extractTagContent(content, "center")

	// 构建标题
	var extractedTitle string
	if h1 != "" {
		extractedTitle = h1
	} else if h2 != "" {
		extractedTitle = h2
	} else if title != "" {
		extractedTitle = title
	}

	// 构建描述
	var extractedDesc string
	if pContent != "" {
		extractedDesc = pContent
	} else if centerContent != "" {
		extractedDesc = centerContent
	}

	// 如果提取到了标题和描述，组合成格式化的错误信息
	if extractedTitle != "" && extractedDesc != "" {
		// 如果标题和描述相同，只返回标题
		if extractedTitle == extractedDesc {
			return extractedTitle
		}
		return extractedTitle + ": " + extractedDesc
	} else if extractedTitle != "" {
		return extractedTitle
	} else if extractedDesc != "" {
		return extractedDesc
	}

	// 如果没有提取到有效信息，返回原始内容（经过基本清理）
	return cleanWhitespace(content)
}

// extractTagContent 提取指定标签的内容
func extractTagContent(content, tagName string) string {
	// 构建正则表达式匹配指定标签的内容
	pattern := `(?i)<` + tagName + `(?:\s+[^>]*)?>(.*?)</` + tagName + `>`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(content)
	if len(matches) > 1 {
		// 提取内容后，移除其中可能存在的 HTML 标签
		extracted := matches[1]
		// 移除所有 HTML 标签
		tagRegex := regexp.MustCompile(`<[^>]*>`)
		cleaned := tagRegex.ReplaceAllString(extracted, "")
		return strings.TrimSpace(cleaned)
	}
	return ""
}

// isHTMLContent 检查内容是否为 HTML
func isHTMLContent(content string) bool {
	// 检查是否包含 HTML 标签
	htmlTagRegex := regexp.MustCompile(`(?i)<[^>]+>`)
	return htmlTagRegex.MatchString(content)
}

// cleanWhitespace 清理多余的空白字符
func cleanWhitespace(content string) string {
	// 清理多余的空白字符，包括连续的空格、制表符、换行符等
	spaceRegex := regexp.MustCompile(`\s+`)
	result := spaceRegex.ReplaceAllString(content, " ")
	return strings.TrimSpace(result)
}

// stripHTML 从字符串中提取有效的错误信息
//
// 对于 HTML 内容，会尝试提取标题和描述信息
// 对于非 HTML 内容，会清理多余的空白字符
func stripHTML(content string) string {
	if content == "" {
		return content
	}

	// 检查是否包含 HTML 内容
	if isHTMLContent(content) {
		return extractHTMLError(content)
	}

	// 非 HTML 内容，只清理空白字符
	return cleanWhitespace(content)
}

// stripErrorHTML 从错误消息中移除 HTML 内容
func stripErrorHTML(err error) error {
	if err == nil {
		return nil
	}

	// 获取错误消息并移除其中的 HTML 内容
	cleanMsg := stripHTML(err.Error())

	// 如果清理后的消息与原消息相同，则直接返回原错误
	if cleanMsg == err.Error() {
		return err
	}

	// 否则创建一个新的错误，包含清理后的消息
	return errors.New(errors.ErrCodeRequestFailed, cleanMsg)
}
