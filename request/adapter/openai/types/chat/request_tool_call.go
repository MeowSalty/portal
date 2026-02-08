package chat

// RequestToolCallType 表示工具调用类型
type RequestToolCallType = string

const (
	RequestToolCallTypeFunction RequestToolCallType = "function"
)

// RequestToolCall 表示工具调用
type RequestToolCall struct {
	ID       string                  `json:"id"`
	Type     RequestToolCallType     `json:"type"`
	Function RequestToolCallFunction `json:"function"`
}

// RequestToolCallFunction 表示工具调用的函数信息
type RequestToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// RequestFunctionCall 表示请求中的函数调用
// Deprecated: 使用工具调用替代。
type RequestFunctionCall struct {
	Name      string `json:"name"`      // 函数名称
	Arguments string `json:"arguments"` // 参数
}
