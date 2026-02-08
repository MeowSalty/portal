package chat

// FunctionCall 表示函数调用
type FunctionCall struct {
	Arguments string `json:"arguments"` // 参数
	Name      string `json:"name"`      // 名称
}
