package errors

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// 上下文值的最大长度限制
const (
	// 单个上下文值的最大字符数
	maxContextValueLength = 400
	// 截断后的后缀提示
	truncateSuffix = "[被截断]"
)

// ErrorCode 定义错误码类型
type ErrorCode string

// 预定义的错误码
const (
	// 通用错误码

	// 未知错误
	ErrCodeUnknown ErrorCode = "UNKNOWN"
	// 内部错误
	ErrCodeInternal ErrorCode = "INTERNAL"
	// 无效参数
	ErrCodeInvalidArgument ErrorCode = "INVALID_ARGUMENT"
	// 资源未找到
	ErrCodeNotFound ErrorCode = "NOT_FOUND"
	// 资源已存在
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	// 权限拒绝
	ErrCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	// 认证失败
	ErrCodeAuthenticationFailed ErrorCode = "AUTHENTICATION_FAILED"
	// 资源耗尽
	ErrCodeResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"
	// 前置条件失败
	ErrCodeFailedPrecondition ErrorCode = "FAILED_PRECONDITION"
	// 操作中止
	ErrCodeAborted ErrorCode = "ABORTED"
	// 超出范围
	ErrCodeOutOfRange ErrorCode = "OUT_OF_RANGE"
	// 未实现
	ErrCodeUnimplemented ErrorCode = "UNIMPLEMENTED"
	// 服务不可用
	ErrCodeUnavailable ErrorCode = "UNAVAILABLE"
	// 数据丢失
	ErrCodeDataLoss ErrorCode = "DATA_LOSS"
	// 超时
	ErrCodeDeadlineExceeded ErrorCode = "DEADLINE_EXCEEDED"

	// 业务相关错误码

	// 无健康通道
	ErrCodeNoHealthyChannel ErrorCode = "NO_HEALTHY_CHANNEL"
	// 适配器未找到
	ErrCodeAdapterNotFound ErrorCode = "ADAPTER_NOT_FOUND"
	// 模型未找到
	ErrCodeModelNotFound ErrorCode = "MODEL_NOT_FOUND"
	// 平台未找到
	ErrCodePlatformNotFound ErrorCode = "PLATFORM_NOT_FOUND"
	// API 密钥未找到
	ErrCodeAPIKeyNotFound ErrorCode = "API_KEY_NOT_FOUND"
	// 请求失败
	ErrCodeRequestFailed ErrorCode = "REQUEST_FAILED"
	// 流处理错误
	ErrCodeStreamError ErrorCode = "STREAM_ERROR"
	// 配置无效
	ErrCodeConfigInvalid ErrorCode = "CONFIG_INVALID"
	// 超出速率限制
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	// 熔断器开启
	ErrCodeCircuitBreakerOpen ErrorCode = "CIRCUIT_BREAKER_OPEN"
)

// Error 结构化错误类型
type Error struct {
	Code       ErrorCode              // 错误码
	HTTPStatus *int                   // HTTP 状态码
	Message    string                 // 错误消息
	Cause      error                  // 原始错误
	Context    map[string]interface{} // 上下文信息
}

// Error 实现 error 接口
func (e *Error) Error() string {
	// 构建基础错误信息
	var msg string
	if e.Cause != nil {
		msg = fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	} else {
		msg = fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}

	// 如果有 HTTP 状态码，则添加到错误信息中
	if e.HTTPStatus != nil {
		msg = fmt.Sprintf("%s (HTTP Status: %d)", msg, *e.HTTPStatus)
	}

	// 如果有上下文信息，则添加到错误信息中
	if len(e.Context) > 0 {
		// 将上下文信息转换为字符串
		var builder strings.Builder
		builder.WriteString("{")
		first := true
		for k, v := range e.Context {
			if !first {
				builder.WriteString(", ")
			}
			builder.WriteString(k)
			builder.WriteString("=")
			valueStr := fmt.Sprintf("%v", v)
			if len(valueStr) > maxContextValueLength {
				// 计算截断后缀的字节长度
				suffixLen := len(truncateSuffix)
				// 计算实际可用的内容长度（需要预留后缀空间）
				maxContentLen := maxContextValueLength - suffixLen

				// 如果后缀长度大于等于最大长度，则只显示后缀
				if maxContentLen <= 0 {
					valueStr = truncateSuffix
				} else {
					// 从 maxContentLen 位置向前查找有效的 UTF-8 字符边界
					truncatePos := maxContentLen
					for truncatePos > 0 && !utf8.RuneStart(valueStr[truncatePos]) {
						truncatePos--
					}

					// 如果找到了有效的截断位置，则进行截断
					if truncatePos > 0 {
						valueStr = valueStr[:truncatePos] + truncateSuffix
					} else {
						// 如果没有找到有效的截断位置，只显示后缀
						valueStr = truncateSuffix
					}
				}
			}
			builder.WriteString(valueStr)
			first = false
		}
		builder.WriteString("}")

		msg = fmt.Sprintf("%s [context: %s]", msg, builder.String())
	}

	return msg
}

// Unwrap 支持错误链
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is 支持错误比较
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithContext 添加上下文信息
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithHTTPStatus 设置 HTTP 状态码
func (e *Error) WithHTTPStatus(status int) *Error {
	e.HTTPStatus = &status
	return e
}

// New 创建新的错误
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// NewWithHTTPStatus 创建新的带 HTTP 状态码的错误
func NewWithHTTPStatus(code ErrorCode, message string, status int) *Error {
	return &Error{
		Code:       code,
		HTTPStatus: &status,
		Message:    message,
	}
}

// Wrap 包装已有错误
func Wrap(code ErrorCode, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// WrapWithHTTPStatus 包装已有错误并设置 HTTP 状态码
func WrapWithHTTPStatus(code ErrorCode, message string, cause error, status int) *Error {
	return &Error{
		Code:       code,
		HTTPStatus: &status,
		Message:    message,
		Cause:      cause,
	}
}

// WrapWithContext 包装错误并添加上下文
func WrapWithContext(code ErrorCode, message string, cause error, context map[string]interface{}) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: context,
	}
}

// WrapWithHTTPStatusAndContext 包装错误并添加 HTTP 状态码和上下文
func WrapWithHTTPStatusAndContext(code ErrorCode, message string, cause error, status int, context map[string]interface{}) *Error {
	return &Error{
		Code:       code,
		HTTPStatus: &status,
		Message:    message,
		Cause:      cause,
		Context:    context,
	}
}

// IsCode 检查错误是否为特定错误码
func IsCode(err error, code ErrorCode) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}

// GetCode 获取错误码
func GetCode(err error) ErrorCode {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ErrCodeUnknown
}

// GetContext 获取错误上下文
func GetContext(err error) map[string]interface{} {
	var e *Error
	if errors.As(err, &e) {
		return e.Context
	}
	return nil
}

// GetHTTPStatus 获取 HTTP 状态码
func GetHTTPStatus(err error) int {
	var e *Error
	if errors.As(err, &e) && e.HTTPStatus != nil {
		return *e.HTTPStatus
	}
	return 0
}

// HasHTTPStatus 检查是否设置了 HTTP 状态码
func HasHTTPStatus(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.HTTPStatus != nil
	}
	return false
}

// GetMessage 获取错误消息
func GetMessage(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Message
	}
	return ""
}

// 预定义的常用错误
var (
	// 通用错误
	ErrUnknown              = New(ErrCodeUnknown, "未知错误")
	ErrInternal             = New(ErrCodeInternal, "内部错误")
	ErrInvalidArgument      = New(ErrCodeInvalidArgument, "无效参数")
	ErrNotFound             = New(ErrCodeNotFound, "资源未找到")
	ErrAlreadyExists        = New(ErrCodeAlreadyExists, "资源已存在")
	ErrPermissionDenied     = New(ErrCodePermissionDenied, "权限被拒绝")
	ErrAuthenticationFailed = New(ErrCodeAuthenticationFailed, "认证失败")
	ErrResourceExhausted    = New(ErrCodeResourceExhausted, "资源耗尽")
	ErrFailedPrecondition   = New(ErrCodeFailedPrecondition, "前置条件失败")
	ErrAborted              = New(ErrCodeAborted, "操作中止")
	ErrOutOfRange           = New(ErrCodeOutOfRange, "超出范围")
	ErrUnimplemented        = New(ErrCodeUnimplemented, "功能未实现")
	ErrUnavailable          = New(ErrCodeUnavailable, "服务不可用")
	ErrDataLoss             = New(ErrCodeDataLoss, "数据丢失")
	ErrDeadlineExceeded     = New(ErrCodeDeadlineExceeded, "超过截止时间")

	// 业务错误
	ErrNoHealthyChannel   = New(ErrCodeNoHealthyChannel, "没有可用的健康通道")
	ErrAdapterNotFound    = New(ErrCodeAdapterNotFound, "未找到适配器")
	ErrModelNotFound      = New(ErrCodeModelNotFound, "未找到模型")
	ErrPlatformNotFound   = New(ErrCodePlatformNotFound, "未找到平台")
	ErrAPIKeyNotFound     = New(ErrCodeAPIKeyNotFound, "未找到 API 密钥")
	ErrRequestFailed      = New(ErrCodeRequestFailed, "请求失败")
	ErrStreamError        = New(ErrCodeStreamError, "流处理错误")
	ErrConfigInvalid      = New(ErrCodeConfigInvalid, "配置无效")
	ErrRateLimitExceeded  = New(ErrCodeRateLimitExceeded, "超出速率限制")
	ErrCircuitBreakerOpen = New(ErrCodeCircuitBreakerOpen, "熔断器已开启")
)
