package chat

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

var (
	requestKnownFieldsOnce        sync.Once
	requestKnownFields            map[string]struct{}
	requestMessageKnownFieldsOnce sync.Once
	requestMessageKnownFields     map[string]struct{}
)

// requestKnownFields 定义 Request 结构体的所有已知字段名称
// 用于在反序列化时识别未知字段
func requestKnownFieldsSet() map[string]struct{} {
	requestKnownFieldsOnce.Do(func() {
		requestKnownFields = buildJSONFieldSet(reflect.TypeOf(Request{}))
	})

	return requestKnownFields
}

// requestMessageKnownFields 定义 RequestMessage 结构体的所有已知字段名称
// 用于在反序列化时识别未知字段
func requestMessageKnownFieldsSet() map[string]struct{} {
	requestMessageKnownFieldsOnce.Do(func() {
		requestMessageKnownFields = buildJSONFieldSet(reflect.TypeOf(RequestMessage{}))
	})

	return requestMessageKnownFields
}

// buildJSONFieldSet 基于结构体 json tag 生成字段集合
func buildJSONFieldSet(structType reflect.Type) map[string]struct{} {
	fieldSet := make(map[string]struct{})
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		tag := field.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" {
			name = field.Name
		}
		if name == "-" {
			continue
		}

		fieldSet[name] = struct{}{}
	}

	return fieldSet
}

// ChatModalities 表示 Chat Completions 的模态
// 可选值：text、audio。
type ChatModalities = string

const (
	ChatModalitiesText  ChatModalities = "text"
	ChatModalitiesAudio ChatModalities = "audio"
)

// Request 表示 OpenAI 聊天完成请求参数
type Request struct {
	Model    string           `json:"model"`            // 模型名称
	Messages []RequestMessage `json:"messages"`         // 消息列表
	Stream   *bool            `json:"stream,omitempty"` // 是否流式传输

	// 可选参数
	FrequencyPenalty     *float64                    `json:"frequency_penalty,omitempty"`      // 频率惩罚
	Logprobs             *bool                       `json:"logprobs,omitempty"`               // 是否返回对数概率
	MaxCompletionTokens  *int                        `json:"max_completion_tokens,omitempty"`  // 最大完成 token 数
	MaxTokens            *int                        `json:"max_tokens,omitempty"`             // 最大 token 数
	N                    *int                        `json:"n,omitempty"`                      // 生成数量
	PresencePenalty      *float64                    `json:"presence_penalty,omitempty"`       // 存在惩罚
	Seed                 *int                        `json:"seed,omitempty"`                   // 随机种子
	Store                *bool                       `json:"store,omitempty"`                  // 是否存储
	Temperature          *float64                    `json:"temperature,omitempty"`            // 温度
	TopLogprobs          *int                        `json:"top_logprobs,omitempty"`           // 顶部对数概率数
	TopP                 *float64                    `json:"top_p,omitempty"`                  // Top-p 采样
	ParallelToolCalls    *bool                       `json:"parallel_tool_calls,omitempty"`    // 并行工具调用
	PromptCacheKey       *string                     `json:"prompt_cache_key,omitempty"`       // 提示缓存键
	PromptCacheRetention *string                     `json:"prompt_cache_retention,omitempty"` // 提示缓存保留策略
	SafetyIdentifier     *string                     `json:"safety_identifier,omitempty"`      // 安全标识符
	User                 *string                     `json:"user,omitempty"`                   // 用户标识符
	Audio                *RequestAudio               `json:"audio,omitempty"`                  // 音频参数
	LogitBias            map[string]int              `json:"logit_bias,omitempty"`             // 对数偏置
	Metadata             map[string]string           `json:"metadata,omitempty"`               // 元数据
	Modalities           []ChatModalities            `json:"modalities,omitempty"`             // 模态
	ReasoningEffort      *shared.ReasoningEffort     `json:"reasoning_effort,omitempty"`       // 推理努力
	ServiceTier          *shared.ServiceTier         `json:"service_tier,omitempty"`           // 服务层级
	Stop                 *StopConfiguration          `json:"stop,omitempty"`                   // 停止条件
	StreamOptions        *StreamOptions              `json:"stream_options,omitempty"`         // 流选项
	Verbosity            *shared.VerbosityLevel      `json:"verbosity,omitempty"`              // 详细程度
	FunctionCall         *FunctionCallUnion          `json:"function_call,omitempty"`          // 函数调用
	Functions            []shared.FunctionDefinition `json:"functions,omitempty"`              // 函数定义
	Prediction           *PredictionContent          `json:"prediction,omitempty"`             // 预测内容
	ResponseFormat       *FormatUnion                `json:"response_format,omitempty"`        // 响应格式
	ToolChoice           *shared.ToolChoiceUnion     `json:"tool_choice,omitempty"`            // 工具选择
	Tools                []ChatToolUnion             `json:"tools,omitempty"`                  // 工具列表
	WebSearchOptions     *WebSearchOptions           `json:"web_search_options,omitempty"`     // 网络搜索选项

	// ExtraFields 存储未知字段
	ExtraFields map[string]interface{} `json:"-"`

	// 自定义 HTTP 头部（不会被序列化到请求体中）
	// 用于透传 User-Agent、Referer 等 HTTP 头部信息
	Headers map[string]string `json:"-"`
}

// StreamOptions 表示流选项
type StreamOptions struct {
	IncludeUsage       *bool `json:"include_usage,omitempty"`       // 是否包含使用情况
	IncludeObfuscation *bool `json:"include_obfuscation,omitempty"` // 是否启用流混淆
}

// MessageRole 表示聊天消息角色
// 参考 OpenAI Chat Completions 文档。
type MessageRole = string

const (
	MessageRoleDeveloper MessageRole = "developer"
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
	MessageRoleFunction  MessageRole = "function"
)

// RequestMessage 表示消息参数
type RequestMessage struct {
	Role         MessageRole          `json:"role"`                    // 角色
	Content      MessageContent       `json:"content"`                 // 内容
	Name         *string              `json:"name,omitempty"`          // 名称
	ToolCallID   *string              `json:"tool_call_id,omitempty"`  // 工具调用 ID
	ToolCalls    []RequestToolCall    `json:"tool_calls,omitempty"`    // 工具调用
	FunctionCall *RequestFunctionCall `json:"function_call,omitempty"` // 函数调用
	Refusal      *string              `json:"refusal,omitempty"`       // 拒绝内容
	Audio        *AssistantAudio      `json:"audio,omitempty"`         // 音频引用

	// ExtraFields 存储未知字段
	ExtraFields map[string]interface{} `json:"-"`
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
	knownFields := requestKnownFieldsSet()
	for key, value := range raw {
		if _, ok := knownFields[key]; !ok {
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
	knownFields := requestMessageKnownFieldsSet()
	for key, value := range raw {
		if _, ok := knownFields[key]; !ok {
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
