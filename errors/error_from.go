package errors

// ErrorFromValue 定义 error_from 的枚举值类型。
type ErrorFromValue string

const (
	// ErrorFromClient 表示客户端导致的错误。
	ErrorFromClient ErrorFromValue = "client"
	// ErrorFromGateway 表示网关（portal）自身导致的错误。
	ErrorFromGateway ErrorFromValue = "gateway"
	// ErrorFromServer 表示目标服务器导致的错误。
	ErrorFromServer ErrorFromValue = "server"
	// ErrorFromUpstream 表示目标服务器上游导致的错误。
	ErrorFromUpstream ErrorFromValue = "upstream"
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
	case ErrorFromClient, ErrorFromGateway, ErrorFromServer, ErrorFromUpstream:
		return errorFrom
	default:
		return ""
	}
}

// IsUpstreamError 判断错误是否来自 server 或 upstream。
func IsUpstreamError(err error) bool {
	switch GetErrorFrom(err) {
	case ErrorFromServer, ErrorFromUpstream:
		return true
	default:
		return false
	}
}
