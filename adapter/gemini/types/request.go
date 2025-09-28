package types

import (
	"encoding/json"
)

// Request 表示 Gemini API 请求结构
type Request struct {
	Contents []Content `json:"contents"`
	// 开发者设置的系统指令
	SystemInstruction *Content `json:"systemInstruction,omitempty"`
	// 生成配置
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
	// 安全设置
	SafetySettings []SafetySetting `json:"safetySettings,omitempty"`
	// 工具定义
	Tools []Tool `json:"tools,omitempty"`
	// 工具配置
	ToolConfig *ToolConfig `json:"toolConfig,omitempty"`
}

// Content 表示内容部分
type Content struct {
	Role  string `json:"role,omitempty"` // 角色：user 或 model
	Parts []Part `json:"parts"`          // 内容部分
}

// Part 表示内容部分，可以是文本或内联数据
type Part struct {
	Text       *string     `json:"text,omitempty"`       // 文本内容
	InlineData *InlineData `json:"inlineData,omitempty"` // 内联数据（如图像）
	// 函数调用响应（用于工具调用）
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
	// 函数调用（模型生成的工具调用）
	FunctionCall *FunctionCall `json:"functionCall,omitempty"`
}

// InlineData 表示内联数据
type InlineData struct {
	MimeType string `json:"mimeType"` // MIME 类型
	Data     string `json:"data"`     // Base64 编码的数据
}

// GenerationConfig 表示生成配置
type GenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`     // 温度
	TopP            *float64 `json:"topP,omitempty"`            // Top-p 采样
	TopK            *int     `json:"topK,omitempty"`            // Top-k 采样
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"` // 最大输出 token 数
	StopSequences   []string `json:"stopSequences,omitempty"`   // 停止序列
	CandidateCount  *int     `json:"candidateCount,omitempty"`  // 候选数量
	// 响应 MIME 类型
	ResponseMIMEType *string `json:"responseMIMEType,omitempty"`
	// 响应 schema
	ResponseSchema *Schema `json:"responseSchema,omitempty"`
}

// SafetySetting 表示安全设置
type SafetySetting struct {
	Category  string `json:"category"`  // 安全类别
	Threshold string `json:"threshold"` // 阈值
}

// Tool 表示工具定义
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations,omitempty"` // 函数声明
}

// FunctionDeclaration 表示函数声明
type FunctionDeclaration struct {
	Name        string      `json:"name"`                  // 函数名称
	Description string      `json:"description,omitempty"` // 函数描述
	Parameters  interface{} `json:"parameters"`            // 参数 schema
}

// ToolConfig 表示工具配置
type ToolConfig struct {
	FunctionCallingConfig *FunctionCallingConfig `json:"functionCallingConfig,omitempty"` // 函数调用配置
}

// FunctionCallingConfig 表示函数调用配置
type FunctionCallingConfig struct {
	Mode string `json:"mode,omitempty"` // 模式：AUTO 或 ANY
}

// FunctionResponse 表示函数调用响应
type FunctionResponse struct {
	Name     string `json:"name"`            // 函数名称
	Response string `json:"response"`        // 函数响应
	ID       string `json:"id"`              // 响应 ID
	Error    string `json:"error,omitempty"` // 错误信息
}

// FunctionCall 表示函数调用
type FunctionCall struct {
	Name string                 `json:"name"` // 函数名称
	Args map[string]interface{} `json:"args"` // 函数参数
}

// Schema 表示 JSON schema
type Schema struct {
	Type        string            `json:"type"`                  // 类型
	Format      string            `json:"format,omitempty"`      // 格式
	Description string            `json:"description,omitempty"` // 描述
	Nullable    bool              `json:"nullable,omitempty"`    // 是否可为空
	Enum        []string          `json:"enum,omitempty"`        // 枚举值
	Properties  map[string]Schema `json:"properties,omitempty"`  // 属性
	Required    []string          `json:"required,omitempty"`    // 必需属性
	Items       *Schema           `json:"items,omitempty"`       // 数组项 schema
}

// MarshalJSON 实现 Part 的自定义 JSON 序列化
func (p Part) MarshalJSON() ([]byte, error) {
	type Alias Part
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(&p),
	}

	// 根据内容类型确定序列化方式
	if p.Text != nil {
		return json.Marshal(struct {
			Text string `json:"text"`
		}{
			Text: *p.Text,
		})
	} else if p.InlineData != nil {
		return json.Marshal(struct {
			InlineData *InlineData `json:"inlineData"`
		}{
			InlineData: p.InlineData,
		})
	} else if p.FunctionResponse != nil {
		return json.Marshal(struct {
			FunctionResponse *FunctionResponse `json:"functionResponse"`
		}{
			FunctionResponse: p.FunctionResponse,
		})
	} else if p.FunctionCall != nil {
		return json.Marshal(struct {
			FunctionCall *FunctionCall `json:"functionCall"`
		}{
			FunctionCall: p.FunctionCall,
		})
	}

	return json.Marshal(aux)
}
