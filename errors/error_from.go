package errors

// ErrorFromValue 定义 error_from 的枚举值类型。
type ErrorFromValue string

const (
	// ErrorFromClient 表示客户端导致的错误。
	ErrorFromClient ErrorFromValue = "client"
	// ErrorFromServer 表示 portal 自身导致的错误。
	ErrorFromServer ErrorFromValue = "server"
	// ErrorFromUpstream 表示目标服务器自身错误。
	ErrorFromUpstream ErrorFromValue = "upstream"
	// ErrorFromUpstreamDependency 表示目标服务器访问其上游失败。
	ErrorFromUpstreamDependency ErrorFromValue = "upstream_dependency"
)

// GetErrorFrom 从错误上下文中提取 error_from 枚举值。
func GetErrorFrom(err error) ErrorFromValue {
	context := GetContext(err)
	if context == nil {
		return ""
	}

	raw, ok := context["error_from"]
	if !ok {
		return ""
	}

	var value string
	switch v := raw.(type) {
	case string:
		value = v
	case ErrorFromValue:
		value = string(v)
	default:
		return ""
	}

	errorFrom := ErrorFromValue(value)
	switch errorFrom {
	case ErrorFromClient, ErrorFromServer, ErrorFromUpstream, ErrorFromUpstreamDependency:
		return errorFrom
	default:
		return ""
	}
}

// IsUpstreamError 判断错误是否来自 upstream 或 upstream_dependency。
func IsUpstreamError(err error) bool {
	switch GetErrorFrom(err) {
	case ErrorFromUpstream, ErrorFromUpstreamDependency:
		return true
	default:
		return false
	}
}
