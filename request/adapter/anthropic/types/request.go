package types

import (
	"encoding/json"

	coreTypes "github.com/MeowSalty/portal/types"
)

// Request Anthropic API 请求结构
type Request struct {
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

// UnmarshalJSON 实现 Request 的自定义 JSON 反序列化
func (r *Request) UnmarshalJSON(data []byte) error {
	// 使用临时结构体避免递归调用
	type Alias Request
	aux := &struct {
		System     json.RawMessage `json:"system,omitempty"`
		ToolChoice json.RawMessage `json:"tool_choice,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 解析 System 字段
	if len(aux.System) > 0 {
		// 尝试解析为字符串
		var str string
		if err := json.Unmarshal(aux.System, &str); err == nil {
			r.System = str
		} else {
			// 尝试解析为 ContentBlock 数组
			var blocks []ContentBlock
			if err := json.Unmarshal(aux.System, &blocks); err == nil {
				r.System = blocks
			}
		}
	}

	// 解析 ToolChoice 字段
	if len(aux.ToolChoice) > 0 {
		// 先解析为 map 以获取 type 字段
		var typeMap map[string]interface{}
		if err := json.Unmarshal(aux.ToolChoice, &typeMap); err == nil {
			if typeStr, ok := typeMap["type"].(string); ok {
				switch typeStr {
				case "auto":
					var toolChoice ToolChoiceAuto
					if err := json.Unmarshal(aux.ToolChoice, &toolChoice); err == nil {
						r.ToolChoice = toolChoice
					}
				case "any":
					var toolChoice ToolChoiceAny
					if err := json.Unmarshal(aux.ToolChoice, &toolChoice); err == nil {
						r.ToolChoice = toolChoice
					}
				case "tool":
					var toolChoice ToolChoiceTool
					if err := json.Unmarshal(aux.ToolChoice, &toolChoice); err == nil {
						r.ToolChoice = toolChoice
					}
				}
			}
		}
	}

	return nil
}

// InputMessage 输入消息结构
type InputMessage struct {
	Role    string      `json:"role"`    // 角色："user" 或 "assistant"
	Content interface{} `json:"content"` // 内容，可以是 string 或 []ContentBlock
}

// UnmarshalJSON 实现 InputMessage 的自定义 JSON 反序列化
func (m *InputMessage) UnmarshalJSON(data []byte) error {
	// 使用临时结构体避免递归调用
	type Alias InputMessage
	aux := &struct {
		Content json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 尝试解析 Content 为字符串
	var str string
	if err := json.Unmarshal(aux.Content, &str); err == nil {
		m.Content = str
		return nil
	}

	// 尝试解析为 ContentBlock 数组
	var blocks []ContentBlock
	if err := json.Unmarshal(aux.Content, &blocks); err == nil {
		m.Content = blocks
		return nil
	}

	// 如果都失败，保持为 nil
	m.Content = nil
	return nil
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

// UnmarshalJSON 实现 ContentBlock 的自定义 JSON 反序列化
func (cb *ContentBlock) UnmarshalJSON(data []byte) error {
	// 使用临时结构体避免递归调用
	type Alias ContentBlock
	aux := &struct {
		Input   json.RawMessage `json:"input,omitempty"`
		Content json.RawMessage `json:"content,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(cb),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 解析 Input 字段 - 保持为原始 map 或值
	if len(aux.Input) > 0 {
		var input map[string]interface{}
		if err := json.Unmarshal(aux.Input, &input); err == nil {
			cb.Input = input
		}
	}

	// 解析 Content 字段
	if len(aux.Content) > 0 {
		// 尝试解析为字符串
		var str string
		if err := json.Unmarshal(aux.Content, &str); err == nil {
			cb.Content = str
		} else {
			// 尝试解析为 ContentBlock 数组
			var blocks []ContentBlock
			if err := json.Unmarshal(aux.Content, &blocks); err == nil {
				cb.Content = blocks
			} else {
				// 尝试解析为通用 map (用于其他类型的内容)
				var content map[string]interface{}
				if err := json.Unmarshal(aux.Content, &content); err == nil {
					cb.Content = content
				}
			}
		}
	}

	return nil
}

// ContentBlocks 是 ContentBlock 的切片类型，用于添加相关方法
type ContentBlocks []ContentBlock

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

// ConvertCoreRequest 将 Anthropic 请求转换为核心请求
func (anthropicReq *Request) ConvertCoreRequest() *coreTypes.Request {
	coreReq := &coreTypes.Request{
		Model: anthropicReq.Model,
	}

	// 转换流参数
	if anthropicReq.Stream != nil {
		coreReq.Stream = anthropicReq.Stream
	}

	// 转换温度参数
	if anthropicReq.Temperature != nil {
		coreReq.Temperature = anthropicReq.Temperature
	}

	// 转换 TopP 参数
	if anthropicReq.TopP != nil {
		coreReq.TopP = anthropicReq.TopP
	}

	// 转换 TopK 参数
	if anthropicReq.TopK != nil {
		coreReq.TopK = anthropicReq.TopK
	}

	// 转换最大 token 数
	maxTokens := anthropicReq.MaxTokens
	coreReq.MaxTokens = &maxTokens

	// 转换停止序列
	if len(anthropicReq.StopSequences) > 0 {
		if len(anthropicReq.StopSequences) == 1 {
			coreReq.Stop.StringValue = &anthropicReq.StopSequences[0]
		} else {
			coreReq.Stop.StringArray = anthropicReq.StopSequences
		}
	}

	// 转换消息
	coreReq.Messages = make([]coreTypes.Message, 0, len(anthropicReq.Messages)+1)

	// 如果存在系统消息，添加到消息列表开头
	if anthropicReq.System != nil {
		systemMsg := coreTypes.Message{
			Role: "system",
		}

		switch sys := anthropicReq.System.(type) {
		case string:
			systemMsg.Content.StringValue = &sys
		case []ContentBlock:
			// 如果系统消息是内容块数组，转换为核心格式
			contentParts := ContentBlocks(sys).convertAnthropicContentBlocks()
			systemMsg.Content.ContentParts = contentParts
		}

		coreReq.Messages = append(coreReq.Messages, systemMsg)
	}

	// 转换普通消息
	for _, msg := range anthropicReq.Messages {
		coreMsg := coreTypes.Message{
			Role: msg.Role,
		}

		// 转换消息内容
		switch content := msg.Content.(type) {
		case string:
			// 字符串内容
			coreMsg.Content.StringValue = &content
		case []ContentBlock:
			// 内容块数组
			coreMsg.Content.ContentParts = ContentBlocks(content).convertAnthropicContentBlocks()
		}

		coreReq.Messages = append(coreReq.Messages, coreMsg)
	}

	// 转换工具（如果存在）
	if len(anthropicReq.Tools) > 0 {
		coreReq.Tools = make([]coreTypes.Tool, len(anthropicReq.Tools))
		for i, tool := range anthropicReq.Tools {
			coreReq.Tools[i] = coreTypes.Tool{
				Type: "function",
				Function: coreTypes.FunctionDescription{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema.convertAnthropicInputSchema(),
				},
			}
		}
	}

	// 转换工具选择
	if anthropicReq.ToolChoice != nil {
		coreReq.ToolChoice = convertAnthropicToolChoice(anthropicReq.ToolChoice)
	}

	// 转换用户元数据
	if anthropicReq.Metadata != nil && anthropicReq.Metadata.UserID != nil {
		coreReq.User = anthropicReq.Metadata.UserID
	}

	return coreReq
}

// convertAnthropicContentBlocks 转换 Anthropic 内容块为核心内容部分
func (blocks ContentBlocks) convertAnthropicContentBlocks() []coreTypes.ContentPart {
	contentParts := make([]coreTypes.ContentPart, 0, len(blocks))

	for _, block := range blocks {
		switch block.Type {
		case "text":
			if block.Text != nil {
				contentParts = append(contentParts, coreTypes.ContentPart{
					Type: "text",
					Text: block.Text,
				})
			}
		case "image":
			if block.Source != nil {
				// 将 Anthropic 的 base64 图像格式转换为 data URL
				url := "data:" + block.Source.MediaType + ";base64," + block.Source.Data
				contentParts = append(contentParts, coreTypes.ContentPart{
					Type: "image_url",
					ImageURL: &coreTypes.ImageURL{
						URL: url,
					},
				})
			}
		case "tool_use":
			// 工具使用内容暂不处理，因为核心 ContentPart 不支持
			// 可能需要在未来扩展核心类型
		case "tool_result":
			// 工具结果内容暂不处理
		}
	}

	return contentParts
}

// convertAnthropicInputSchema 转换 Anthropic InputSchema 为核心参数格式
func (schema InputSchema) convertAnthropicInputSchema() interface{} {
	params := map[string]interface{}{
		"type": schema.Type,
	}

	if schema.Properties != nil {
		params["properties"] = schema.Properties
	}

	if len(schema.Required) > 0 {
		params["required"] = schema.Required
	}

	return params
}

// convertAnthropicToolChoice 转换 Anthropic 工具选择为核心格式
func convertAnthropicToolChoice(toolChoice interface{}) coreTypes.ToolChoice {
	var coreToolChoice coreTypes.ToolChoice

	switch tc := toolChoice.(type) {
	case ToolChoiceAuto:
		mode := "auto"
		coreToolChoice.Mode = &mode
	case ToolChoiceAny:
		// "any" 在核心格式中映射为 "auto"
		mode := "auto"
		coreToolChoice.Mode = &mode
	case ToolChoiceTool:
		// 指定工具选择
		coreToolChoice.Function = &coreTypes.FunctionToolChoice{
			Type: "function",
			Function: coreTypes.FunctionNameChoice{
				Name: tc.Name,
			},
		}
	case map[string]interface{}:
		// 处理通过 JSON 解析得到的 map 格式
		if typeStr, ok := tc["type"].(string); ok {
			switch typeStr {
			case "auto":
				mode := "auto"
				coreToolChoice.Mode = &mode
			case "any":
				mode := "auto"
				coreToolChoice.Mode = &mode
			case "tool":
				if name, ok := tc["name"].(string); ok {
					coreToolChoice.Function = &coreTypes.FunctionToolChoice{
						Type: "function",
						Function: coreTypes.FunctionNameChoice{
							Name: name,
						},
					}
				}
			}
		}
	}

	return coreToolChoice
}
