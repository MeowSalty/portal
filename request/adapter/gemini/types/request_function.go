package types

// FunctionDeclaration 表示函数声明
type FunctionDeclaration struct {
	Name                 string      `json:"name"`                  // 函数名称
	Description          string      `json:"description,omitempty"` // 函数描述
	Parameters           *Schema     `json:"parameters,omitempty"`  // 参数 schema
	ParametersJSONSchema interface{} `json:"parametersJsonSchema,omitempty"`
	Response             *Schema     `json:"response,omitempty"`
	ResponseJSONSchema   interface{} `json:"responseJsonSchema,omitempty"`
	Behavior             *string     `json:"behavior,omitempty"`
}

// FunctionResponse 表示函数调用响应
type FunctionResponse struct {
	ID           *string                `json:"id,omitempty"` // 响应 ID
	Name         string                 `json:"name"`         // 函数名称
	Response     map[string]interface{} `json:"response"`     // 函数响应
	Parts        []FunctionResponsePart `json:"parts,omitempty"`
	WillContinue *bool                  `json:"willContinue,omitempty"`
	Scheduling   *string                `json:"scheduling,omitempty"`
}

// FunctionResponsePart 表示函数响应的 Part
type FunctionResponsePart struct {
	InlineData *FunctionResponseBlob `json:"inlineData,omitempty"`
}

// FunctionResponseBlob 表示函数响应的内联数据
type FunctionResponseBlob struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

// FunctionCall 表示函数调用
type FunctionCall struct {
	ID   *string                `json:"id,omitempty"`
	Name string                 `json:"name"` // 函数名称
	Args map[string]interface{} `json:"args"` // 函数参数
}
