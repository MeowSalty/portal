package shared

// FunctionDefinition 表示函数定义
type FunctionDefinition struct {
	Name        string      `json:"name"`                  // 函数名称
	Description *string     `json:"description,omitempty"` // 描述
	Parameters  interface{} `json:"parameters,omitempty"`  // 参数
	Strict      *bool       `json:"strict,omitempty"`      // 是否严格校验
}
