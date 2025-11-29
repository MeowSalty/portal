package types

import (
	"encoding/json"
)

// requestKnownFields 定义 Request 结构体的所有已知字段名称
// 用于在反序列化时识别未知字段
var requestKnownFields = map[string]bool{
	"model":                 true,
	"messages":              true,
	"stream":                true,
	"frequency_penalty":     true,
	"logprobs":              true,
	"max_completion_tokens": true,
	"max_tokens":            true,
	"n":                     true,
	"presence_penalty":      true,
	"seed":                  true,
	"store":                 true,
	"temperature":           true,
	"top_logprobs":          true,
	"top_p":                 true,
	"parallel_tool_calls":   true,
	"prompt_cache_key":      true,
	"safety_identifier":     true,
	"user":                  true,
	"audio":                 true,
	"logit_bias":            true,
	"metadata":              true,
	"modalities":            true,
	"reasoning_effort":      true,
	"service_tier":          true,
	"stop":                  true,
	"stream_options":        true,
	"verbosity":             true,
	"function_call":         true,
	"functions":             true,
	"prediction":            true,
	"response_format":       true,
	"tool_choice":           true,
	"tools":                 true,
	"web_search_options":    true,
}

// requestMessageKnownFields 定义 RequestMessage 结构体的所有已知字段名称
// 用于在反序列化时识别未知字段
var requestMessageKnownFields = map[string]bool{
	"role":    true,
	"content": true,
	"name":    true,
}

// Request 表示 OpenAI 聊天完成请求参数
type Request struct {
	Model    string           `json:"model"`            // 模型名称
	Messages []RequestMessage `json:"messages"`         // 消息列表
	Stream   *bool            `json:"stream,omitempty"` // 是否流式传输

	// 可选参数
	FrequencyPenalty    *float64             `json:"frequency_penalty,omitempty"`     // 频率惩罚
	Logprobs            *bool                `json:"logprobs,omitempty"`              // 是否返回对数概率
	MaxCompletionTokens *int                 `json:"max_completion_tokens,omitempty"` // 最大完成 token 数
	MaxTokens           *int                 `json:"max_tokens,omitempty"`            // 最大 token 数
	N                   *int                 `json:"n,omitempty"`                     // 生成数量
	PresencePenalty     *float64             `json:"presence_penalty,omitempty"`      // 存在惩罚
	Seed                *int                 `json:"seed,omitempty"`                  // 随机种子
	Store               *bool                `json:"store,omitempty"`                 // 是否存储
	Temperature         *float64             `json:"temperature,omitempty"`           // 温度
	TopLogprobs         *int                 `json:"top_logprobs,omitempty"`          // 顶部对数概率数
	TopP                *float64             `json:"top_p,omitempty"`                 // Top-p 采样
	ParallelToolCalls   *bool                `json:"parallel_tool_calls,omitempty"`   // 并行工具调用
	PromptCacheKey      *string              `json:"prompt_cache_key,omitempty"`      // 提示缓存键
	SafetyIdentifier    *string              `json:"safety_identifier,omitempty"`     // 安全标识符
	User                *string              `json:"user,omitempty"`                  // 用户标识符
	Audio               *RequestAudio        `json:"audio,omitempty"`                 // 音频参数
	LogitBias           map[string]int       `json:"logit_bias,omitempty"`            // 对数偏置
	Metadata            map[string]string    `json:"metadata,omitempty"`              // 元数据
	Modalities          []string             `json:"modalities,omitempty"`            // 模态
	ReasoningEffort     *string              `json:"reasoning_effort,omitempty"`      // 推理努力
	ServiceTier         *string              `json:"service_tier,omitempty"`          // 服务层级
	Stop                *StopUnion           `json:"stop,omitempty"`                  // 停止条件
	StreamOptions       *StreamOptions       `json:"stream_options,omitempty"`        // 流选项
	Verbosity           *string              `json:"verbosity,omitempty"`             // 详细程度
	FunctionCall        *FunctionCallUnion   `json:"function_call,omitempty"`         // 函数调用
	Functions           []FunctionDefinition `json:"functions,omitempty"`             // 函数定义
	Prediction          *PredictionContent   `json:"prediction,omitempty"`            // 预测内容
	ResponseFormat      *FormatUnion         `json:"response_format,omitempty"`       // 响应格式
	ToolChoice          *ToolChoiceUnion     `json:"tool_choice,omitempty"`           // 工具选择
	Tools               []ToolUnion          `json:"tools,omitempty"`                 // 工具列表
	WebSearchOptions    *WebSearchOptions    `json:"web_search_options,omitempty"`    // 网络搜索选项

	// ExtraFields 存储未知字段
	ExtraFields map[string]interface{} `json:"-"`
}

// RequestMessage 表示消息参数
type RequestMessage struct {
	Role    string      `json:"role"`           // 角色
	Content interface{} `json:"content"`        // 内容
	Name    *string     `json:"name,omitempty"` // 名称

	// ExtraFields 存储未知字段
	ExtraFields map[string]interface{} `json:"-"`
}

// RequestAudio 表示音频参数
type RequestAudio struct {
	Format string `json:"format"` // 格式
	Voice  string `json:"voice"`  // 声音
}

// StreamOptions 表示流选项
type StreamOptions struct {
	IncludeObfuscation *bool `json:"include_obfuscation,omitempty"` // 是否包含混淆
	IncludeUsage       *bool `json:"include_usage,omitempty"`       // 是否包含使用情况
}

// StopUnion 表示停止条件联合类型
type StopUnion struct {
	StringValue *string
	StringArray []string
}

// FunctionCallUnion 表示函数调用联合类型
type FunctionCallUnion struct {
	Mode     *string
	Function *FunctionCallOption
}

// FunctionCallOption 表示函数调用选项
type FunctionCallOption struct {
	Name string `json:"name"` // 函数名称
}

// FunctionDefinition 表示函数定义
type FunctionDefinition struct {
	Name        string      `json:"name"`                  // 函数名称
	Description *string     `json:"description,omitempty"` // 描述
	Parameters  interface{} `json:"parameters"`            // 参数
}

// PredictionContent 表示预测内容
type PredictionContent struct {
	Type    string      `json:"type"`    // 类型
	Content interface{} `json:"content"` // 内容
}

// FormatUnion 表示响应格式联合类型
type FormatUnion struct {
	Text       *FormatText
	JSONSchema *FormatJSONSchema
	JSONObject *FormatJSONObject
}

// FormatText 表示文本响应格式
type FormatText struct {
	Type string `json:"type"` // 类型
}

// FormatJSONSchema 表示 JSON Schema 响应格式
type FormatJSONSchema struct {
	Type       string      `json:"type"`        // 类型
	JSONSchema interface{} `json:"json_schema"` // JSON Schema
}

// FormatJSONObject 表示 JSON 对象响应格式
type FormatJSONObject struct {
	Type string `json:"type"` // 类型
}

// ToolChoiceUnion 表示工具选择联合类型
type ToolChoiceUnion struct {
	Auto        *string
	Allowed     *ToolChoiceAllowed
	Named       *ToolChoiceNamed
	NamedCustom *ToolChoiceNamedCustom
}

// ToolChoiceAllowed 表示允许的工具选择
type ToolChoiceAllowed struct {
	Type  string                   `json:"type"`  // 类型
	Mode  string                   `json:"mode"`  // 模式
	Tools []map[string]interface{} `json:"tools"` // 工具列表
}

// ToolChoiceNamed 表示命名的工具选择
type ToolChoiceNamed struct {
	Type     string `json:"type"` // 类型
	Function struct {
		Name string `json:"name"` // 函数名称
	} `json:"function"` // 函数
}

// ToolChoiceNamedCustom 表示命名的自定义工具选择
type ToolChoiceNamedCustom struct {
	Type   string `json:"type"` // 类型
	Custom struct {
		Name string `json:"name"` // 名称
	} `json:"custom"` // 自定义
}

// ToolUnion 表示工具联合类型
type ToolUnion struct {
	Function *ToolFunction
	Custom   *ToolCustom
}

// ToolFunction 表示函数工具
type ToolFunction struct {
	Type     string             `json:"type"`     // 类型
	Function FunctionDefinition `json:"function"` // 函数定义
}

// ToolCustom 表示自定义工具
type ToolCustom struct {
	Type   string      `json:"type"`   // 类型
	Custom interface{} `json:"custom"` // 自定义内容
}

// WebSearchOptions 表示网络搜索选项
type WebSearchOptions struct {
	UserLocation      *UserLocation `json:"user_location,omitempty"`       // 用户位置
	SearchContextSize *string       `json:"search_context_size,omitempty"` // 搜索上下文大小
}

// UserLocation 表示用户位置
type UserLocation struct {
	Type        string               `json:"type"`                  // 类型
	Approximate *ApproximateLocation `json:"approximate,omitempty"` // 近似位置
}

// ApproximateLocation 表示近似位置
type ApproximateLocation struct {
	City     *string `json:"city,omitempty"`     // 城市
	Country  *string `json:"country,omitempty"`  // 国家
	Region   *string `json:"region,omitempty"`   // 地区
	Timezone *string `json:"timezone,omitempty"` // 时区
}

// MarshalJSON 实现 StopUnion 的自定义 JSON 序列化
func (s StopUnion) MarshalJSON() ([]byte, error) {
	if s.StringValue != nil {
		return json.Marshal(s.StringValue)
	}
	if s.StringArray != nil {
		return json.Marshal(s.StringArray)
	}
	return json.Marshal(nil)
}

// MarshalJSON 实现 FunctionCallUnion 的自定义 JSON 序列化
func (f FunctionCallUnion) MarshalJSON() ([]byte, error) {
	if f.Mode != nil {
		return json.Marshal(f.Mode)
	}
	if f.Function != nil {
		return json.Marshal(f.Function)
	}
	return json.Marshal(nil)
}

// MarshalJSON 实现 ResponseFormatUnion 的自定义 JSON 序列化
func (r FormatUnion) MarshalJSON() ([]byte, error) {
	if r.Text != nil {
		return json.Marshal(r.Text)
	}
	if r.JSONSchema != nil {
		return json.Marshal(r.JSONSchema)
	}
	if r.JSONObject != nil {
		return json.Marshal(r.JSONObject)
	}
	return json.Marshal(nil)
}

// MarshalJSON 实现 ToolChoiceUnion 的自定义 JSON 序列化
func (t ToolChoiceUnion) MarshalJSON() ([]byte, error) {
	if t.Auto != nil {
		return json.Marshal(t.Auto)
	}
	if t.Allowed != nil {
		return json.Marshal(t.Allowed)
	}
	if t.Named != nil {
		return json.Marshal(t.Named)
	}
	if t.NamedCustom != nil {
		return json.Marshal(t.NamedCustom)
	}
	return json.Marshal(nil)
}

// MarshalJSON 实现 ToolUnion 的自定义 JSON 序列化
func (t ToolUnion) MarshalJSON() ([]byte, error) {
	if t.Function != nil {
		return json.Marshal(t.Function)
	}
	if t.Custom != nil {
		return json.Marshal(t.Custom)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 Request 的自定义 JSON 反序列化
// 捕获所有未知字段并存储到 ExtraFields
func (r *Request) UnmarshalJSON(data []byte) error {
	// 1. 解析到通用 map 以获取所有字段
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 2. 使用类型别名避免递归调用
	type Alias Request
	aux := &struct{ *Alias }{Alias: (*Alias)(r)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 3. 提取未知字段
	r.ExtraFields = make(map[string]interface{})
	for key, value := range raw {
		if !requestKnownFields[key] {
			r.ExtraFields[key] = value
		}
	}

	return nil
}

// MarshalJSON 实现 Request 的自定义 JSON 序列化
// 合并已知字段和 ExtraFields 中的未知字段
func (r Request) MarshalJSON() ([]byte, error) {
	// 1. 如果没有未知字段，使用默认序列化
	if len(r.ExtraFields) == 0 {
		// 使用类型别名避免递归调用
		type Alias Request
		aux := Alias(r)
		return json.Marshal(aux)
	}

	// 2. 序列化已知字段到 map
	type Alias Request
	aux := Alias(r)
	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	// 3. 解析到 map
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// 4. 合并未知字段
	for key, value := range r.ExtraFields {
		result[key] = value
	}

	// 5. 序列化最终结果
	return json.Marshal(result)
}

// UnmarshalJSON 实现 RequestMessage 的自定义 JSON 反序列化
// 捕获所有未知字段并存储到 ExtraFields
func (m *RequestMessage) UnmarshalJSON(data []byte) error {
	// 1. 解析到通用 map 以获取所有字段
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 2. 使用类型别名避免递归调用
	type Alias RequestMessage
	aux := &struct{ *Alias }{Alias: (*Alias)(m)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 3. 提取未知字段
	m.ExtraFields = make(map[string]interface{})
	for key, value := range raw {
		if !requestMessageKnownFields[key] {
			m.ExtraFields[key] = value
		}
	}

	return nil
}

// MarshalJSON 实现 RequestMessage 的自定义 JSON 序列化
// 合并已知字段和 ExtraFields 中的未知字段
func (m RequestMessage) MarshalJSON() ([]byte, error) {
	// 1. 如果没有未知字段，使用默认序列化
	if len(m.ExtraFields) == 0 {
		// 使用类型别名避免递归调用
		type Alias RequestMessage
		aux := Alias(m)
		return json.Marshal(aux)
	}

	// 2. 序列化已知字段到 map
	type Alias RequestMessage
	aux := Alias(m)
	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	// 3. 解析到 map
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// 4. 合并未知字段
	for key, value := range m.ExtraFields {
		result[key] = value
	}

	// 5. 序列化最终结果
	return json.Marshal(result)
}
