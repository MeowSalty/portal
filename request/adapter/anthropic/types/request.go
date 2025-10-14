package types

// AnthropicRequest Anthropic API 请求结构
type AnthropicRequest struct {
	Model         string         `json:"model"`                    // 模型名称
	Messages      []InputMessage `json:"messages"`                 // 消息列表
	MaxTokens     int            `json:"max_tokens"`               // 最大生成 token 数
	System        interface{}    `json:"system,omitempty"`         // 系统提示，可以是 string 或 []RequestTextBlock
	Stream        *bool          `json:"stream,omitempty"`         // 是否流式传输
	Temperature   *float64       `json:"temperature,omitempty"`    // 温度参数
	TopP          *float64       `json:"top_p,omitempty"`          // Top-p 采样
	TopK          *int           `json:"top_k,omitempty"`          // Top-k 采样
	StopSequences []string       `json:"stop_sequences,omitempty"` // 停止序列
	Metadata      *Metadata      `json:"metadata,omitempty"`       // 元数据
	Tools         []Tool         `json:"tools,omitempty"`          // 工具列表
	ToolChoice    interface{}    `json:"tool_choice,omitempty"`    // 工具选择
}

// InputMessage 输入消息结构
type InputMessage struct {
	Role    string      `json:"role"`    // 角色："user" 或 "assistant"
	Content interface{} `json:"content"` // 内容，可以是 string 或 []ContentBlock
}

// ContentBlock 内容块接口，可以是文本、图像等
type ContentBlock struct {
	Type      string       `json:"type"`                  // 类型："text", "image", "tool_use", "tool_result"
	Text      *string      `json:"text,omitempty"`        // 文本内容
	Source    *ImageSource `json:"source,omitempty"`      // 图像源
	ID        *string      `json:"id,omitempty"`          // 工具使用 ID
	Name      *string      `json:"name,omitempty"`        // 工具名称
	Input     interface{}  `json:"input,omitempty"`       // 工具输入
	ToolUseID *string      `json:"tool_use_id,omitempty"` // 工具结果对应的工具使用 ID
	Content   interface{}  `json:"content,omitempty"`     // 工具结果内容
	IsError   *bool        `json:"is_error,omitempty"`    // 是否为错误
}

// ImageSource 图像源
type ImageSource struct {
	Type      string `json:"type"`       // "base64"
	MediaType string `json:"media_type"` // "image/jpeg", "image/png", "image/gif", "image/webp"
	Data      string `json:"data"`       // base64 编码的图像数据
}

// Metadata 元数据
type Metadata struct {
	UserID *string `json:"user_id,omitempty"` // 用户 ID
}

// Tool 工具定义
type Tool struct {
	Name        string      `json:"name"`                  // 工具名称
	Description *string     `json:"description,omitempty"` // 工具描述
	InputSchema InputSchema `json:"input_schema"`          // 输入 schema
}

// InputSchema 工具输入 schema
type InputSchema struct {
	Type       string                 `json:"type"`                 // "object"
	Properties map[string]interface{} `json:"properties,omitempty"` // 属性定义
	Required   []string               `json:"required,omitempty"`   // 必需字段
}

// ToolChoiceAuto 自动工具选择
type ToolChoiceAuto struct {
	Type                   string `json:"type"` // "auto"
	DisableParallelToolUse *bool  `json:"disable_parallel_tool_use,omitempty"`
}

// ToolChoiceAny 任意工具选择
type ToolChoiceAny struct {
	Type                   string `json:"type"` // "any"
	DisableParallelToolUse *bool  `json:"disable_parallel_tool_use,omitempty"`
}

// ToolChoiceTool 指定工具选择
type ToolChoiceTool struct {
	Type                   string `json:"type"` // "tool"
	Name                   string `json:"name"` // 工具名称
	DisableParallelToolUse *bool  `json:"disable_parallel_tool_use,omitempty"`
}
