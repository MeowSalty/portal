package responses

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

var (
	requestKnownFieldsOnce sync.Once
	requestKnownFields     map[string]struct{}
)

// requestKnownFieldsSet 定义 Request 结构体的所有已知字段名称
// 用于在反序列化时识别未知字段
func requestKnownFieldsSet() map[string]struct{} {
	requestKnownFieldsOnce.Do(func() {
		requestKnownFields = buildJSONFieldSet(reflect.TypeOf(Request{}))
	})

	return requestKnownFields
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

// Request 表示 OpenAI Responses API 请求
// 仅保留当前需要的字段，其他字段通过 ExtraFields 透传。
type Request struct {
	Model                *string                    `json:"model,omitempty"`
	Input                *InputUnion                `json:"input,omitempty"`
	Stream               *bool                      `json:"stream,omitempty"`
	StreamOptions        *StreamOptions             `json:"stream_options,omitempty"`
	MaxOutputTokens      *int                       `json:"max_output_tokens,omitempty"`
	Temperature          *float64                   `json:"temperature,omitempty"`
	TopP                 *float64                   `json:"top_p,omitempty"`
	TopLogprobs          *int                       `json:"top_logprobs,omitempty"`
	Tools                []shared.ToolUnion         `json:"tools,omitempty"`
	ToolChoice           *shared.ToolChoiceUnion    `json:"tool_choice,omitempty"`
	ParallelToolCalls    *bool                      `json:"parallel_tool_calls,omitempty"`
	MaxToolCalls         *int                       `json:"max_tool_calls,omitempty"`
	Truncation           *TruncationStrategy        `json:"truncation,omitempty"`
	Text                 *TextConfig                `json:"text,omitempty"`
	Store                *bool                      `json:"store,omitempty"`
	Include              IncludeList                `json:"include,omitempty"`
	Metadata             map[string]string          `json:"metadata,omitempty"`
	Instructions         *string                    `json:"instructions,omitempty"`
	Reasoning            *Reasoning                 `json:"reasoning,omitempty"`
	Prompt               *PromptTemplate            `json:"prompt,omitempty"`
	Conversation         *ConversationUnion         `json:"conversation,omitempty"`
	PreviousResponseID   *string                    `json:"previous_response_id,omitempty"`
	SafetyIdentifier     *string                    `json:"safety_identifier,omitempty"`
	User                 *string                    `json:"user,omitempty"`
	PromptCacheKey       *string                    `json:"prompt_cache_key,omitempty"`
	PromptCacheRetention *PromptCacheRetention      `json:"prompt_cache_retention,omitempty"`
	ServiceTier          *shared.ServiceTier        `json:"service_tier,omitempty"`
	Background           *bool                      `json:"background,omitempty"`
	ExtraFields          map[string]json.RawMessage `json:"-"`
}

// ConversationReference 表示会话引用对象
// 对应 {"id": "conv_xxx"} 形态。
type ConversationReference struct {
	ID string `json:"id"`
}

// InputUnion 表示 input 字段的联合类型
// 支持 string 或 input items 数组。
type InputUnion struct {
	StringValue *string
	Items       []InputItem
}

// MarshalJSON 实现 InputUnion 的自定义 JSON 序列化
func (i InputUnion) MarshalJSON() ([]byte, error) {
	if i.StringValue != nil {
		return json.Marshal(i.StringValue)
	}
	if i.Items != nil {
		return json.Marshal(i.Items)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 InputUnion 的自定义 JSON 反序列化
func (i *InputUnion) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		i.StringValue = &str
		return nil
	}

	var items []InputItem
	if err := json.Unmarshal(data, &items); err == nil {
		i.Items = items
		return nil
	}

	return nil
}

// ConversationUnion 表示会话字段，支持字符串或对象。
type ConversationUnion struct {
	StringValue *string
	ObjectValue *ConversationReference
}

// MarshalJSON 实现 ConversationUnion 的自定义 JSON 序列化
func (c ConversationUnion) MarshalJSON() ([]byte, error) {
	if c.StringValue != nil {
		return json.Marshal(c.StringValue)
	}
	if c.ObjectValue != nil {
		return json.Marshal(c.ObjectValue)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON 实现 ConversationUnion 的自定义 JSON 反序列化
func (c *ConversationUnion) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		c.StringValue = &str
		return nil
	}

	var obj ConversationReference
	if err := json.Unmarshal(data, &obj); err == nil && obj.ID != "" {
		c.ObjectValue = &obj
		return nil
	}

	return nil
}

// Reasoning 表示推理模型参数
// 仅适用于 gpt-5 和 o-series。
type Reasoning struct {
	Effort          *shared.ReasoningEffort `json:"effort,omitempty"`
	Summary         *ReasoningSummary       `json:"summary,omitempty"`
	GenerateSummary *ReasoningSummary       `json:"generate_summary,omitempty"`
}

// MarshalJSON 实现 Reasoning 的自定义 JSON 序列化
// 处理逻辑：
// 1. Summary 为 nil，GenerateSummary 为 nil -> summary: null
// 2. Summary 不为 nil，GenerateSummary 为 nil -> summary: [对应的内容]
// 3. Summary 为 nil，GenerateSummary 不为 nil -> generate_summary: [对应的内容]（不输出 summary 字段）
// 4. Summary 不为 nil，GenerateSummary 不为 nil -> 错误输入，报错
func (r Reasoning) MarshalJSON() ([]byte, error) {
	// 检查是否同时设置了 Summary 和 GenerateSummary
	if r.Summary != nil && r.GenerateSummary != nil {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "Reasoning 配置错误：不能同时设置 summary 和 generate_summary 字段").
			WithCause(nil).
			WithContext("summary", *r.Summary).
			WithContext("generate_summary", *r.GenerateSummary)
	}

	// 使用 map 来灵活控制输出的字段
	result := make(map[string]interface{})
	if r.Effort != nil {
		result["effort"] = r.Effort
	}

	// 如果 Summary 有值，使用 summary 字段
	if r.Summary != nil {
		result["summary"] = *r.Summary
	} else if r.GenerateSummary != nil {
		// 如果 GenerateSummary 有值，使用 generate_summary 字段，不输出 summary 字段
		result["generate_summary"] = *r.GenerateSummary
	} else {
		// 两者都为 nil，summary 输出 null
		result["summary"] = nil
	}

	return json.Marshal(result)
}

// UnmarshalJSON 实现 Reasoning 的自定义 JSON 反序列化
// 处理逻辑：
// 1. 解析 effort 字段
// 2. 如果存在 summary 字段，设置 Summary
// 3. 如果存在 generate_summary 字段，设置 GenerateSummary
// 4. 如果同时存在 summary 和 generate_summary 字段，返回错误
func (r *Reasoning) UnmarshalJSON(data []byte) error {
	// 定义临时结构体用于解析
	var raw struct {
		Effort          *shared.ReasoningEffort `json:"effort,omitempty"`
		Summary         *ReasoningSummary       `json:"summary,omitempty"`
		GenerateSummary *ReasoningSummary       `json:"generate_summary,omitempty"`
	}

	// 解析 JSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 检查是否同时设置了 Summary 和 GenerateSummary
	if raw.Summary != nil && raw.GenerateSummary != nil {
		return errors.New(errors.ErrCodeInvalidArgument, "Reasoning 配置错误：不能同时设置 summary 和 generate_summary 字段").
			WithCause(nil).
			WithContext("summary", *raw.Summary).
			WithContext("generate_summary", *raw.GenerateSummary)
	}

	// 设置解析后的值
	r.Effort = raw.Effort
	r.Summary = raw.Summary
	r.GenerateSummary = raw.GenerateSummary

	return nil
}

// PromptTemplate 表示可复用的提示模板
// 详情参考 OpenAI 文档的 reusable prompts。
type PromptTemplate struct {
	ID        string             `json:"id"`
	Version   *string            `json:"version,omitempty"`
	Variables *PromptVariableMap `json:"variables,omitempty"`
}

// PromptVariableMap 表示提示变量的联合类型
// 支持字符串、输入文本、图片或文件内容。
type PromptVariableMap map[string]PromptVariableValue

// MarshalJSON 实现 PromptVariableMap 的自定义 JSON 序列化
func (pm PromptVariableMap) MarshalJSON() ([]byte, error) {
	if pm == nil {
		return json.Marshal(nil)
	}
	result := make(map[string]PromptVariableValue, len(pm))
	for key, value := range pm {
		result[key] = value
	}
	return json.Marshal(result)
}

// UnmarshalJSON 实现 PromptVariableMap 的自定义 JSON 反序列化
func (pm *PromptVariableMap) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var raw map[string]PromptVariableValue
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if pm == nil {
		return nil
	}
	*pm = raw
	return nil
}

// PromptVariableValue 表示提示变量的值联合类型
type PromptVariableValue struct {
	StringValue *string
	Text        *InputTextContent
	Image       *InputImageContent
	File        *InputFileContent
}

// MarshalJSON 实现 PromptVariableValue 的自定义 JSON 序列化
func (pv PromptVariableValue) MarshalJSON() ([]byte, error) {
	switch {
	case pv.StringValue != nil:
		return json.Marshal(pv.StringValue)
	case pv.Text != nil:
		return json.Marshal(pv.Text)
	case pv.Image != nil:
		return json.Marshal(pv.Image)
	case pv.File != nil:
		return json.Marshal(pv.File)
	default:
		return json.Marshal(nil)
	}
}

// UnmarshalJSON 实现 PromptVariableValue 的自定义 JSON 反序列化
func (pv *PromptVariableValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		pv.StringValue = &str
		return nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	typeVal, _ := raw["type"].(string)
	switch typeVal {
	case "input_text":
		var text InputTextContent
		if err := json.Unmarshal(data, &text); err == nil {
			pv.Text = &text
		}
	case "input_image":
		var image InputImageContent
		if err := json.Unmarshal(data, &image); err == nil {
			pv.Image = &image
		}
	case "input_file":
		var file InputFileContent
		if err := json.Unmarshal(data, &file); err == nil {
			pv.File = &file
		}
	}

	return nil
}

// TextConfig 表示文本输出配置
// 用于结构化输出和文本详细程度控制。
type TextConfig struct {
	Format    *TextFormatUnion       `json:"format,omitempty"`
	Verbosity *shared.VerbosityLevel `json:"verbosity,omitempty"`
}

// TextFormatUnion 表示文本响应格式联合类型
type TextFormatUnion struct {
	Text       *TextFormat
	JSONSchema *TextFormatJSONSchema
	JSONObject *TextFormatJSONObject
}

// TextFormat 表示文本响应格式
// type 固定为 "text"。
type TextFormat struct {
	Type TextResponseFormatType `json:"type"`
}

// TextFormatJSONSchema 表示 JSON Schema 文本响应格式
// 对应 text.format 的 json_schema 结构。
type TextFormatJSONSchema struct {
	Type        TextResponseFormatType `json:"type"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Schema      map[string]interface{} `json:"schema"`
	Strict      *bool                  `json:"strict,omitempty"`
}

// TextFormatJSONObject 表示 JSON 对象文本响应格式
type TextFormatJSONObject struct {
	Type TextResponseFormatType `json:"type"`
}

// MarshalJSON 实现 TextFormatUnion 的自定义 JSON 序列化
func (t TextFormatUnion) MarshalJSON() ([]byte, error) {
	switch {
	case t.Text != nil:
		return json.Marshal(t.Text)
	case t.JSONSchema != nil:
		return json.Marshal(t.JSONSchema)
	case t.JSONObject != nil:
		return json.Marshal(t.JSONObject)
	default:
		return json.Marshal(nil)
	}
}

// UnmarshalJSON 实现 TextFormatUnion 的自定义 JSON 反序列化
func (t *TextFormatUnion) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	typeVal, _ := raw["type"].(string)
	switch typeVal {
	case "text", "":
		var text TextFormat
		if err := json.Unmarshal(data, &text); err == nil {
			t.Text = &text
		}
	case "json_schema":
		var schema TextFormatJSONSchema
		if err := json.Unmarshal(data, &schema); err == nil {
			t.JSONSchema = &schema
		}
	case "json_object":
		var obj TextFormatJSONObject
		if err := json.Unmarshal(data, &obj); err == nil {
			t.JSONObject = &obj
		}
	}

	return nil
}

// StreamOptions 表示流式响应选项
// 仅在 stream=true 时设置。
type StreamOptions struct {
	IncludeObfuscation *bool `json:"include_obfuscation,omitempty"`
}

// UnmarshalJSON 实现 Request 的自定义 JSON 反序列化
// 捕获所有未知字段并存储到 ExtraFields
func (r *Request) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	type Alias Request
	aux := &struct{ *Alias }{Alias: (*Alias)(r)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 收集未知字段，只有存在未知字段时才初始化 ExtraFields
	knownFields := requestKnownFieldsSet()
	var unknownFields map[string]json.RawMessage
	for key, value := range raw {
		if _, ok := knownFields[key]; !ok {
			if unknownFields == nil {
				unknownFields = make(map[string]json.RawMessage)
			}
			unknownFields[key] = value
		}
	}

	// 只有存在未知字段时才设置 ExtraFields
	if len(unknownFields) > 0 {
		r.ExtraFields = unknownFields
	}

	return nil
}

// MarshalJSON 实现 Request 的自定义 JSON 序列化
// 合并已知字段和 ExtraFields 中的未知字段
func (r Request) MarshalJSON() ([]byte, error) {
	if len(r.ExtraFields) == 0 {
		type Alias Request
		aux := Alias(r)
		return json.Marshal(aux)
	}

	type Alias Request
	aux := Alias(r)
	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// 合并 ExtraFields，显式字段优先
	knownFields := requestKnownFieldsSet()
	for key, value := range r.ExtraFields {
		if _, ok := knownFields[key]; !ok {
			result[key] = value
		}
	}

	return json.Marshal(result)
}
