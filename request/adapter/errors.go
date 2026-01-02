package adapter

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/MeowSalty/portal/errors"
)

// ErrorFrom 表示错误来源
// 网关 (A) → 服务器 (B) → 外部服务 (C)
type ErrorFrom int

const (
	// ErrorFromServer B 产生的错误
	ErrorFromServer ErrorFrom = iota
	// ErrorFromUpstream C 产生的错误，经 B 转发
	ErrorFromUpstream
)

// isJSONContentType 检查 Content-Type 是否为 JSON 类型
func isJSONContentType(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/json")
}

// handleHTTPError 处理 HTTP 错误
func (a *Adapter) handleHTTPError(message string, statusCode int, contentType string, body []byte) error {
	if len(body) == 0 {
		return a.createHTTPError(message, statusCode, "", ErrorFromServer)
	}

	var bodyStr string
	var errorFrom ErrorFrom

	// 根据 Content-Type 决定处理方式
	if isJSONContentType(contentType) {
		// Content-Type 表明是 JSON，尝试解析
		var jsonData map[string]interface{}
		err := json.Unmarshal(body, &jsonData)

		if err != nil {
			// JSON 解析失败，回退到 HTML/文本处理
			bodyStr = stripHTML(string(body))
			errorFrom = ErrorFromServer
		} else {
			// JSON 解析成功
			bodyStr = a.processBodyHTML(jsonData)
			errorFrom = a.classifyErrorFrom(jsonData, statusCode)
		}
	} else {
		// 非 JSON 类型，直接处理为 HTML/文本
		bodyStr = stripHTML(string(body))
		errorFrom = ErrorFromServer
	}

	// 根据错误来源和状态码处理错误
	return a.createHTTPError(message, statusCode, bodyStr, errorFrom)
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

// classifyErrorFrom 根据已解析的 JSON 数据分类错误来源
func (a *Adapter) classifyErrorFrom(jsonData map[string]interface{}, statusCode int) ErrorFrom {
	if errorObj, ok := jsonData["error"].(map[string]interface{}); ok {
		if errorType, ok := errorObj["type"].(string); ok {
			if errorType == "upstream_error" || errorType == "openai_error" {
				return ErrorFromUpstream
			}
			// 当状态码为 503 且错误类型为 one_hub_error 或 new_api_error 时也视为上游错误
			if statusCode == 503 && (errorType == "one_hub_error" || errorType == "new_api_error") {
				return ErrorFromUpstream
			}
		}
	}
	return ErrorFromServer
}

// createHTTPError 根据错误来源和状态码创建适当的错误
func (a *Adapter) createHTTPError(message string, statusCode int, bodyStr string, errorFrom ErrorFrom) error {
	// 定义错误码
	var errCode errors.ErrorCode

	switch errorFrom {
	case ErrorFromServer:
		// B 产生的错误，根据 HTTP 状态码映射
		errCode = mapHTTPStatusToErrorCode(statusCode)
	case ErrorFromUpstream:
		// C 产生的错误，经 B 转发，视为请求错误
		errCode = errors.ErrCodeRequestFailed
	default:
		// 未知来源
		errCode = errors.ErrCodeUnknown
	}

	return errors.NewWithHTTPStatus(errCode, message, statusCode).
		WithContext("response_body", bodyStr).
		WithContext("error_from", errorFromToString(errorFrom))
}

// mapHTTPStatusToErrorCode 将 HTTP 状态码映射到错误码
func mapHTTPStatusToErrorCode(statusCode int) errors.ErrorCode {
	switch statusCode {
	// 4xx 客户端错误
	case 400:
		return errors.ErrCodeInvalidArgument
	case 401:
		return errors.ErrCodeAuthenticationFailed
	case 403:
		return errors.ErrCodePermissionDenied
	case 404:
		return errors.ErrCodeNotFound
	case 408:
		return errors.ErrCodeDeadlineExceeded
	case 422:
		return errors.ErrCodeInvalidArgument
	case 429:
		return errors.ErrCodeRateLimitExceeded
	// 5xx 服务端错误
	case 500:
		return errors.ErrCodeInternal
	case 502, 503:
		return errors.ErrCodeUnavailable
	case 504:
		return errors.ErrCodeDeadlineExceeded
	default:
		// 根据状态码范围判断
		if statusCode >= 400 && statusCode < 500 {
			return errors.ErrCodeInvalidArgument
		} else if statusCode >= 500 && statusCode < 600 {
			return errors.ErrCodeInternal
		}
		return errors.ErrCodeUnknown
	}
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

// errorFromToString 将错误来源转换为字符串描述
func errorFromToString(errorFrom ErrorFrom) string {
	switch errorFrom {
	case ErrorFromServer:
		return "server"
	case ErrorFromUpstream:
		return "upstream"
	default:
		return "unknown"
	}
}
