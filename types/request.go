package types

import (
	"encoding/json"
)

// Request 表示主要的 API 请求结构
type Request struct {
	// Messages 或 Prompt 必须有一个存在
	Messages []Message `json:"messages,omitempty"`
	// Messages 或 Prompt 必须有一个存在
	Prompt *string `json:"prompt,omitempty"`

	Model string `json:"model"` // 参见 "Supported Models" 部分

	// 强制模型产生特定的输出格式
	// 查看模型页面和文档页面了解哪些模型支持此功能
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// 停止生成的条件，可以是字符串或字符串数组
	Stop Stop `json:"stop,omitempty"`
	// 是否启用流式传输
	Stream *bool `json:"stream,omitempty"`

	// 参见 LLM 参数 (openrouter.ai/docs/api-reference/parameters)
	// 范围：[1, context_length)
	MaxTokens *int `json:"max_tokens,omitempty"`
	// 范围：[0, 2]
	Temperature *float64 `json:"temperature,omitempty"`

	// 工具调用
	// 对于实现 OpenAI 接口的提供商，将按原样传递
	// 对于具有自定义接口的提供商，我们会转换和映射属性
	// 否则，我们将工具转换为 YAML 模板。模型使用助手消息进行响应
	// 查看支持工具调用的模型：openrouter.ai/models?supported_parameters=tools
	Tools []Tool `json:"tools,omitempty"`
	// 工具选择偏好
	ToolChoice ToolChoice `json:"tool_choice,omitempty"`

	// 高级可选参数
	Seed *int `json:"seed,omitempty"` // 仅限整数
	// 范围：(0, 1]
	TopP *float64 `json:"top_p,omitempty"`
	// 范围：[1, Infinity)，OpenAI 模型不可用
	TopK *int `json:"top_k,omitempty"`
	// 范围：[-2, 2]
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	// 范围：[-2, 2]
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`
	// 范围：(0, 2]
	RepetitionPenalty *float64 `json:"repetition_penalty,omitempty"`
	// 对数偏置映射
	LogitBias map[int]float64 `json:"logit_bias,omitempty"`
	// 仅限整数
	TopLogprobs *int `json:"top_logprobs,omitempty"`
	// 范围：[0, 1]
	MinP *float64 `json:"min_p,omitempty"`
	// 范围：[0, 1]
	TopA *float64 `json:"top_a,omitempty"`

	// 通过为模型提供预测输出来减少延迟
	// https://platform.openai.com/docs/guides/latency-optimization#use-predicted-outputs
	Prediction *Prediction `json:"prediction,omitempty"`

	// OpenRouter 独有参数
	// 参见 "Prompt Transforms" 部分：openrouter.ai/docs/transforms
	Transforms []string `json:"transforms,omitempty"`
	// 参见 "Model Routing" 部分：openrouter.ai/docs/model-routing
	Models []string `json:"models,omitempty"`
	// 回退路由
	Route *string `json:"route,omitempty"`
	// 参见 "Provider Routing" 部分：openrouter.ai/docs/provider-routing
	Provider *ProviderPreferences `json:"provider,omitempty"`
	// 用于标识最终用户的稳定标识符。用于帮助检测和防止滥用
	User *string `json:"user,omitempty"`

	// 自定义 HTTP 头部（不会被序列化到请求体中）
	// 用于透传 User-Agent、Referer 等 HTTP 头部信息
	Headers map[string]string `json:"-"`
}

// ResponseFormat 定义输出格式约束
type ResponseFormat struct {
	Type string `json:"type"` // 输出格式类型，如 "json_object"
}

// Stop 可以是字符串或字符串数组
type Stop struct {
	StringValue *string
	StringArray []string
}

// MarshalJSON 实现 Stop 的自定义 JSON 序列化
func (s Stop) MarshalJSON() ([]byte, error) {
	if s.StringValue != nil {
		return json.Marshal(s.StringValue)
	}
	if s.StringArray != nil {
		return json.Marshal(s.StringArray)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 Stop 的自定义 JSON 反序列化
func (s *Stop) UnmarshalJSON(data []byte) error {
	// 首先尝试反序列化为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.StringValue = &str
		return nil
	}

	// 尝试反序列化为字符串数组
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		s.StringArray = arr
		return nil
	}

	return nil
}

// TextContent 表示文本内容部分
type TextContent struct {
	Type string `json:"type"` // 内容类型，固定为 "text"
	Text string `json:"text"` // 文本内容
}

// ImageContentPart 表示图像内容部分
type ImageContentPart struct {
	Type     string   `json:"type"`      // 内容类型，固定为 "image_url"
	ImageURL ImageURL `json:"image_url"` // 图像 URL 信息
}

// ImageURL 表示图像 URL 详细信息
type ImageURL struct {
	URL    string  `json:"url"`              // URL 或 base64 编码的图像数据
	Detail *string `json:"detail,omitempty"` // 可选，默认为 "auto"
}

// ContentPart 表示内容部分，可以是文本或图像
type ContentPart struct {
	Type     string    `json:"type"`
	Text     *string   `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// Message 表示聊天消息，具有不同的角色
type Message struct {
	// 消息角色：'user'、'assistant'、'system' 或 'tool'
	Role string `json:"role"`
	// 消息内容：对于 "user" 角色可以是字符串或 ContentPart 数组
	// 对于其他角色通常是字符串
	Content MessageContent `json:"content"`
	// 如果包含 "name"，对于非 OpenAI 模型会这样预置：`{name}: {content}`
	Name *string `json:"name,omitempty"`
	// 仅适用于 'tool' 角色的工具调用 ID
	ToolCallID *string `json:"tool_call_id,omitempty"`
}

// MessageContent 处理字符串或 ContentPart 数组
type MessageContent struct {
	StringValue  *string
	ContentParts []ContentPart
}

// MarshalJSON 实现 MessageContent 的自定义 JSON 序列化
func (mc MessageContent) MarshalJSON() ([]byte, error) {
	if mc.StringValue != nil {
		return json.Marshal(mc.StringValue)
	}
	if mc.ContentParts != nil {
		return json.Marshal(mc.ContentParts)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 MessageContent 的自定义 JSON 反序列化
func (mc *MessageContent) UnmarshalJSON(data []byte) error {
	// 首先尝试反序列化为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		mc.StringValue = &str
		return nil
	}

	// 尝试反序列化为 ContentPart 数组
	var parts []ContentPart
	if err := json.Unmarshal(data, &parts); err == nil {
		mc.ContentParts = parts
		return nil
	}

	return nil
}

// FunctionDescription 描述函数工具
type FunctionDescription struct {
	Description *string     `json:"description,omitempty"` // 函数描述
	Name        string      `json:"name"`                  // 函数名称
	Parameters  interface{} `json:"parameters"`            // JSON Schema 对象
}

// Tool 表示函数工具
type Tool struct {
	Type     string              `json:"type"`     // 工具类型，固定为 "function"
	Function FunctionDescription `json:"function"` // 函数描述
}

// ToolChoice 表示工具选择偏好
type ToolChoice struct {
	Mode     *string             // 模式："none" 或 "auto"
	Function *FunctionToolChoice // 函数特定的工具选择
}

// FunctionToolChoice 表示函数特定的工具选择
type FunctionToolChoice struct {
	Type     string             `json:"type"`     // 类型，固定为 "function"
	Function FunctionNameChoice `json:"function"` // 函数名称选择
}

// FunctionNameChoice 表示函数名称选择
type FunctionNameChoice struct {
	Name string `json:"name"` // 函数名称
}

// MarshalJSON 实现 ToolChoice 的自定义 JSON 序列化
func (tc ToolChoice) MarshalJSON() ([]byte, error) {
	if tc.Mode != nil {
		switch *tc.Mode {
		case "none", "auto":
			return json.Marshal(*tc.Mode)
		}
	}
	if tc.Function != nil {
		return json.Marshal(tc.Function)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 ToolChoice 的自定义 JSON 反序列化
func (tc *ToolChoice) UnmarshalJSON(data []byte) error {
	// 首先尝试反序列化为字符串
	var mode string
	if err := json.Unmarshal(data, &mode); err == nil {
		if mode == "none" || mode == "auto" {
			tc.Mode = &mode
			return nil
		}
	}

	// 尝试反序列化为函数工具选择
	var funcChoice FunctionToolChoice
	if err := json.Unmarshal(data, &funcChoice); err == nil {
		tc.Function = &funcChoice
		return nil
	}

	return nil
}

// Prediction 表示延迟优化预测
type Prediction struct {
	Type    string `json:"type"`    // 预测类型，固定为 "content"
	Content string `json:"content"` // 预测内容
}

// ProviderPreferences 表示提供商路由偏好
type ProviderPreferences struct {
	// 根据实际的提供商偏好结构定义
	// 这是一个占位符 - 根据实际的 API 文档进行调整
	Preference map[string]interface{} `json:"preference,omitempty"`
}
