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

// handleHTTPError 处理 HTTP 错误
func (a *Adapter) handleHTTPError(message string, statusCode int, body []byte) error {
	// 去除 body 中的 HTML 内容
	bodyStr := stripHTML(string(body))

	// 尝试解析 JSON 响应，检查是否包含 error.type 字段
	errorType := a.classifyHTTPErrorType(body)

	// 根据错误类型和状态码处理错误
	return a.createHTTPError(message, statusCode, bodyStr, errorType)
}

// classifyHTTPErrorType 根据响应体内容分类 HTTP 错误类型
func (a *Adapter) classifyHTTPErrorType(body []byte) HTTPErrorType {
	if len(body) == 0 {
		return HTTPErrorTypeServiceUnavailable
	}

	var jsonResp struct {
		Error struct {
			Type string `json:"type"`
		} `json:"error"`
	}

	// 尝试解析 JSON，如果解析失败或没有 error.type 字段，则认为是服务不可用
	if err := json.Unmarshal(body, &jsonResp); err != nil || jsonResp.Error.Type == "" {
		return HTTPErrorTypeServiceUnavailable
	}

	return HTTPErrorTypeAPIError
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

// stripHTML 移除字符串中的完整 HTML 页面，但保留其他文本内容
// 对于混合内容（即包含完整 HTML 页面和普通文本），将只移除完整的 HTML 页面部分
func stripHTML(content string) string {
	if content == "" {
		return content
	}

	// 使用正则表达式匹配完整的 HTML 页面
	// 匹配以<!DOCTYPE html>或<html>开头，以</html>结尾的内容
	htmlPageRegex := regexp.MustCompile(`(?i)(?:<!DOCTYPE\s+html[^>]*>\s*)?<html\b[^>]*>.*?</html>`)

	// 替换所有完整的 HTML 页面为指定提示文本
	result := htmlPageRegex.ReplaceAllString(content, "[HTML content filtered]")

	// 同时移除可能存在的 HTML 标签
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	result = tagRegex.ReplaceAllString(result, "")

	// 清理多余的空白字符，包括连续的空格、制表符、换行符等
	spaceRegex := regexp.MustCompile(`\s+`)
	result = spaceRegex.ReplaceAllString(result, " ")

	// 去除首尾空格
	result = strings.TrimSpace(result)

	return result
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
