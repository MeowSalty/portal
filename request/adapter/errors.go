package adapter

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/MeowSalty/portal/errors"
)

// 预编译正则表达式，避免每次调用时重新编译
var (
	htmlTagRegex    = regexp.MustCompile(`(?i)<[^>]+>`)
	allTagsRegex    = regexp.MustCompile(`<[^>]*>`)
	whitespaceRegex = regexp.MustCompile(`\s+`)
	tagPatterns     = map[string]*regexp.Regexp{
		"title":  regexp.MustCompile(`(?i)<title(?:\s+[^>]*)?>(.+?)</title>`),
		"h1":     regexp.MustCompile(`(?i)<h1(?:\s+[^>]*)?>(.+?)</h1>`),
		"h2":     regexp.MustCompile(`(?i)<h2(?:\s+[^>]*)?>(.+?)</h2>`),
		"p":      regexp.MustCompile(`(?i)<p(?:\s+[^>]*)?>(.+?)</p>`),
		"center": regexp.MustCompile(`(?i)<center(?:\s+[^>]*)?>(.+?)</center>`),
	}
)

// errorClassifyInput 统一错误来源分类输入。
type errorClassifyInput struct {
	errorType    string
	errorCode    string
	errorMessage string
	rawText      string
	hasBody      bool
	// hasHTTPResponse 表示是否已收到目标服务器的 HTTP 响应。
	// 语义优先级高于 hasBody：即使响应体为空，只要已收到响应，也应默认归类为 server。
	hasHTTPResponse bool
	isStructured    bool
}

// handleHTTPError 处理 HTTP 错误
func (a *Adapter) handleHTTPError(message string, statusCode int, body []byte) error {
	bodyStr, classifyInput := a.normalizeHTTPErrorBody(body)
	// handleHTTPError 仅在已拿到 HTTP 响应时调用。
	classifyInput.hasHTTPResponse = true
	errorFrom := classifyErrorFromInput(classifyInput)

	return a.createHTTPError(message, statusCode, bodyStr, errorFrom)
}

// normalizeHTTPErrorBody 规范化 HTTP 错误体，并构造统一分类输入。
func (a *Adapter) normalizeHTTPErrorBody(body []byte) (string, errorClassifyInput) {
	input := errorClassifyInput{hasBody: len(body) > 0}
	if !input.hasBody {
		return "", input
	}

	// 优先尝试解析 JSON：部分服务端返回 JSON 但 Content-Type 非 application/json。
	var jsonData map[string]interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		bodyStr := a.processBodyHTML(jsonData)
		errorType, errorCode, errorMessage := extractErrorFields(jsonData)

		input.errorType = errorType
		input.errorCode = errorCode
		input.errorMessage = errorMessage
		input.isStructured = errorType != "" || errorCode != "" || errorMessage != ""
		input.rawText = strings.ToLower(stripHTML(bodyStr))

		return bodyStr, input
	}

	bodyStr := stripHTML(string(body))
	input.rawText = strings.ToLower(bodyStr)

	return bodyStr, input
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

// extractErrorFields 从已解析 JSON 中提取错误字段。
func extractErrorFields(jsonData map[string]interface{}) (errorType, errorCode, errorMessage string) {
	errorObj, ok := jsonData["error"].(map[string]interface{})
	if !ok {
		return "", "", ""
	}

	if v, ok := errorObj["type"].(string); ok {
		errorType = strings.ToLower(v)
	}
	if v, ok := errorObj["code"].(string); ok {
		errorCode = strings.ToLower(v)
	}
	if v, ok := errorObj["message"].(string); ok {
		errorMessage = strings.ToLower(v)
	}

	return errorType, errorCode, errorMessage
}

// classifyErrorFromInput 统一分类错误来源。
//
// 优先级：
// 1. 结构化字段（type/code/message）
// 2. 非结构化文本（rawText）
// 3. 兜底：hasHTTPResponse=true -> server，hasHTTPResponse=false -> gateway
func classifyErrorFromInput(input errorClassifyInput) errors.ErrorFromValue {
	classifierInput := errors.ClassifierInput{
		ErrorType:            input.errorType,
		VendorCode:           input.errorCode,
		Message:              input.errorMessage,
		ErrorMessage:         input.errorMessage,
		RawText:              input.rawText,
		HTTPResponseReceived: input.hasHTTPResponse,
	}

	result := errors.ClassifyError(classifierInput)
	return result.Source.Value
}

// createHTTPError 根据错误来源和状态码创建适当的错误
func (a *Adapter) createHTTPError(message string, statusCode int, bodyStr string, errorFrom errors.ErrorFromValue) error {
	// 定义错误码
	var errCode errors.ErrorCode

	switch errorFrom {
	case errors.ErrorFromGateway:
		// 网关产生的错误，视为该服务不可用
		errCode = errors.ErrCodeUnavailable
	case errors.ErrorFromServer:
		// 供应商产生的错误，根据 HTTP 状态码映射
		errCode = mapHTTPStatusToErrorCode(statusCode)
	case errors.ErrorFromUpstream:
		// 供应商的上游产生的错误，经供应商转发，视为请求错误
		errCode = errors.ErrCodeRequestFailed
	default:
		// 未知来源
		errCode = errors.ErrCodeUnknown
	}

	return errors.NewWithHTTPStatus(errCode, message, statusCode).
		WithContext("response_body", bodyStr).
		WithContext("http_response_received", true).
		WithContext("error_from", string(errorFrom))
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
		WithContext("response_body", string(body)).
		WithContext("error_from", string(errors.ErrorFromGateway))
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
	regex, ok := tagPatterns[tagName]
	if !ok {
		return ""
	}
	matches := regex.FindStringSubmatch(content)
	if len(matches) > 1 {
		// 提取内容后，移除其中可能存在的 HTML 标签
		cleaned := allTagsRegex.ReplaceAllString(matches[1], "")
		return strings.TrimSpace(cleaned)
	}
	return ""
}

// isHTMLContent 检查内容是否为 HTML
func isHTMLContent(content string) bool {
	return htmlTagRegex.MatchString(content)
}

// cleanWhitespace 清理多余的空白字符
func cleanWhitespace(content string) string {
	result := whitespaceRegex.ReplaceAllString(content, " ")
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
