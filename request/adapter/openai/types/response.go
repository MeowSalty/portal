package types

import "encoding/json"

// Response 表示 OpenAI 聊天完成响应
type Response struct {
	ID                string   `json:"id"`                           // 聊天完成 ID
	Choices           []Choice `json:"choices"`                      // 选择列表
	Created           int64    `json:"created"`                      // 创建时间戳
	Model             string   `json:"model"`                        // 模型名称
	Object            string   `json:"object"`                       // 对象类型
	ServiceTier       *string  `json:"service_tier,omitempty"`       // 服务层级
	SystemFingerprint *string  `json:"system_fingerprint,omitempty"` // 系统指纹
	Usage             *Usage   `json:"usage"`                        // 使用情况
}

// Choice 表示聊天完成选择
type Choice struct {
	FinishReason *string   `json:"finish_reason"`     // 完成原因
	Index        int       `json:"index"`             // 索引
	Logprobs     *Logprobs `json:"logprobs"`          // 对数概率
	Message      *Message  `json:"message,omitempty"` // 消息（非流式响应）
	Delta        *Delta    `json:"delta,omitempty"`   // 增量消息（流式响应）
}

// Logprobs 表示对数概率信息
type Logprobs struct {
	Content []TokenLogprob `json:"content,omitempty"` // 内容对数概率
	Refusal []TokenLogprob `json:"refusal,omitempty"` // 拒绝对数概率
}

// TokenLogprob 表示 token 对数概率
type TokenLogprob struct {
	Token       string                   `json:"token"`                  // token
	Bytes       []int                    `json:"bytes"`                  // 字节表示
	Logprob     float64                  `json:"logprob"`                // 对数概率
	TopLogprobs []TokenLogprobTopLogprob `json:"top_logprobs,omitempty"` // 顶部对数概率
}

// TokenLogprobTopLogprob 表示顶部 token 对数概率
type TokenLogprobTopLogprob struct {
	Token   string  `json:"token"`   // token
	Bytes   []int   `json:"bytes"`   // 字节表示
	Logprob float64 `json:"logprob"` // 对数概率
}

// Message 表示聊天完成消息（非流式响应）
type Message struct {
	Content      *string       `json:"content,omitempty"`       // 内容
	Refusal      *string       `json:"refusal,omitempty"`       // 拒绝消息
	Role         string        `json:"role"`                    // 角色
	FunctionCall *FunctionCall `json:"function_call,omitempty"` // 函数调用
	ToolCalls    []ToolCall    `json:"tool_calls,omitempty"`    // 工具调用

	// ExtraFields 存储未知字段
	ExtraFields map[string]interface{} `json:"-"`
}

// Delta 表示聊天完成增量消息（流式响应）
type Delta struct {
	Content      *string       `json:"content,omitempty"`       // 内容
	Refusal      *string       `json:"refusal,omitempty"`       // 拒绝消息
	Role         *string       `json:"role,omitempty"`          // 角色
	FunctionCall *FunctionCall `json:"function_call,omitempty"` // 函数调用
	ToolCalls    []ToolCall    `json:"tool_calls,omitempty"`    // 工具调用

	// ExtraFields 存储未知字段
	ExtraFields map[string]interface{} `json:"-"`
}

// ResponseAudio 表示音频响应
type ResponseAudio struct {
	ID         string `json:"id"`         // 音频 ID
	Data       string `json:"data"`       // 音频数据
	ExpiresAt  int    `json:"expires_at"` // 过期时间
	Transcript string `json:"transcript"` // 转录文本
}

// FunctionCall 表示函数调用
type FunctionCall struct {
	Arguments string `json:"arguments"` // 参数
	Name      string `json:"name"`      // 名称
}

// ToolCall 表示工具调用
type ToolCall struct {
	Index    int               `json:"index"`
	ID       *string           `json:"id,omitempty"`       // 工具调用 ID（流式响应中可选）
	Type     *string           `json:"type,omitempty"`     // 类型（流式响应中可选）
	Function *ToolCallFunction `json:"function,omitempty"` // 函数
}

// ToolCallFunction 表示工具调用函数
type ToolCallFunction struct {
	Arguments *string `json:"arguments,omitempty"` // 参数（流式响应中逐步累积）
	Name      *string `json:"name,omitempty"`      // 名称（流式响应中可选）
}

// Usage 表示使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`     // 提示 token 数
	CompletionTokens int `json:"completion_tokens"` // 完成 token 数
	TotalTokens      int `json:"total_tokens"`      // 总 token 数
}

// UnmarshalJSON 实现 Message 的自定义 JSON 反序列化
// 捕获所有未知字段并存储到 ExtraFields
func (m *Message) UnmarshalJSON(data []byte) error {
	// 1. 解析到通用 map 以获取所有字段
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 2. 使用类型别名避免递归调用
	type Alias Message
	aux := &struct{ *Alias }{Alias: (*Alias)(m)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 3. 提取未知字段
	m.ExtraFields = make(map[string]interface{})
	for key, value := range raw {
		if !messageKnownFields[key] {
			m.ExtraFields[key] = value
		}
	}

	return nil
}

// MarshalJSON 实现 Message 的自定义 JSON 序列化
// 合并已知字段和 ExtraFields 中的未知字段
func (m Message) MarshalJSON() ([]byte, error) {
	// 1. 如果没有未知字段，使用默认序列化
	if len(m.ExtraFields) == 0 {
		// 使用类型别名避免递归调用
		type Alias Message
		aux := Alias(m)
		return json.Marshal(aux)
	}

	// 2. 序列化已知字段到 map
	type Alias Message
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

// UnmarshalJSON 实现 Delta 的自定义 JSON 反序列化
// 捕获所有未知字段并存储到 ExtraFields
func (d *Delta) UnmarshalJSON(data []byte) error {
	// 1. 解析到通用 map 以获取所有字段
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 2. 使用类型别名避免递归调用
	type Alias Delta
	aux := &struct{ *Alias }{Alias: (*Alias)(d)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 3. 提取未知字段
	d.ExtraFields = make(map[string]interface{})
	for key, value := range raw {
		if !deltaKnownFields[key] {
			d.ExtraFields[key] = value
		}
	}

	return nil
}

// MarshalJSON 实现 Delta 的自定义 JSON 序列化
// 合并已知字段和 ExtraFields 中的未知字段
func (d Delta) MarshalJSON() ([]byte, error) {
	// 1. 如果没有未知字段，使用默认序列化
	if len(d.ExtraFields) == 0 {
		// 使用类型别名避免递归调用
		type Alias Delta
		aux := Alias(d)
		return json.Marshal(aux)
	}

	// 2. 序列化已知字段到 map
	type Alias Delta
	aux := Alias(d)
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
	for key, value := range d.ExtraFields {
		result[key] = value
	}

	// 5. 序列化最终结果
	return json.Marshal(result)
}

// messageKnownFields 定义 Message 结构体的所有已知字段名称
// 用于在反序列化时识别未知字段
var messageKnownFields = map[string]bool{
	"content":       true,
	"refusal":       true,
	"role":          true,
	"function_call": true,
	"tool_calls":    true,
}

// deltaKnownFields 定义 Delta 结构体的所有已知字段名称
// 用于在反序列化时识别未知字段
var deltaKnownFields = map[string]bool{
	"content":       true,
	"refusal":       true,
	"role":          true,
	"function_call": true,
	"tool_calls":    true,
}
