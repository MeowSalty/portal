package types

import (
	"encoding/json"
)

// Request Anthropic Messages Create 请求参数。
type Request struct {
	Model         string                     `json:"model"`                    // 模型名称
	Messages      []Message                  `json:"messages"`                 // 输入消息
	MaxTokens     int                        `json:"max_tokens"`               // 最大生成 token
	Metadata      *Metadata                  `json:"metadata,omitempty"`       // 元数据
	ServiceTier   *ServiceTier               `json:"service_tier,omitempty"`   // 服务层级
	StopSequences []string                   `json:"stop_sequences,omitempty"` // 停止序列
	Stream        *bool                      `json:"stream,omitempty"`         // 是否流式
	System        *SystemParam               `json:"system,omitempty"`         // system prompt：string 或 []TextBlockParam
	Temperature   *float64                   `json:"temperature,omitempty"`    // 温度
	Thinking      *ThinkingConfigParam       `json:"thinking,omitempty"`       // 思考配置
	ToolChoice    *ToolChoiceParam           `json:"tool_choice,omitempty"`    // 工具选择
	Tools         []ToolUnion                `json:"tools,omitempty"`          // 工具定义
	TopK          *int                       `json:"top_k,omitempty"`          // top_k
	TopP          *float64                   `json:"top_p,omitempty"`          // top_p
	ExtraFields   map[string]json.RawMessage `json:"-"`                        // 未知字段透传（仅顶层）

	// 自定义 HTTP 头部（不会被序列化到请求体中）
	// 用于透传 User-Agent、Referer 等 HTTP 头部信息
	Headers map[string]string `json:"-"`
}

// MarshalJSON 实现 Request 的序列化，合并显式字段与 ExtraFields。
func (r Request) MarshalJSON() ([]byte, error) {
	// 定义显式字段名集合
	explicitFields := map[string]bool{
		"model":          true,
		"messages":       true,
		"max_tokens":     true,
		"metadata":       true,
		"service_tier":   true,
		"stop_sequences": true,
		"stream":         true,
		"system":         true,
		"temperature":    true,
		"thinking":       true,
		"tool_choice":    true,
		"tools":          true,
		"top_k":          true,
		"top_p":          true,
	}

	// 先序列化显式字段
	type requestAlias Request
	alias := requestAlias(r)
	data, err := json.Marshal(alias)
	if err != nil {
		return nil, err
	}

	// 如果没有 ExtraFields，直接返回
	if len(r.ExtraFields) == 0 {
		return data, nil
	}

	// 解析为 map 以便合并
	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// 合并 ExtraFields，显式字段优先
	for key, value := range r.ExtraFields {
		if !explicitFields[key] {
			result[key] = value
		}
	}

	return json.Marshal(result)
}

// UnmarshalJSON 实现 Request 的反序列化，收集未知字段到 ExtraFields。
func (r *Request) UnmarshalJSON(data []byte) error {
	// 定义显式字段名集合
	explicitFields := map[string]bool{
		"model":          true,
		"messages":       true,
		"max_tokens":     true,
		"metadata":       true,
		"service_tier":   true,
		"stop_sequences": true,
		"stream":         true,
		"system":         true,
		"temperature":    true,
		"thinking":       true,
		"tool_choice":    true,
		"tools":          true,
		"top_k":          true,
		"top_p":          true,
	}

	// 解析为 map 以便收集未知字段
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 收集未知字段，只有存在未知字段时才初始化 ExtraFields
	var unknownFields map[string]json.RawMessage
	for key, value := range raw {
		if !explicitFields[key] {
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

	// 使用别名类型避免递归调用 UnmarshalJSON
	type requestAlias Request
	alias := (*requestAlias)(r)
	return json.Unmarshal(data, alias)
}
