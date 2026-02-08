package types

import (
	"encoding/json"
)

// Request 表示 Gemini API 请求结构
type Request struct {
	// Model 通过 URL 传递，不参与 JSON 序列化
	Model string `json:"-"`
	// 对话内容
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
	// 缓存内容引用
	CachedContent *string `json:"cachedContent,omitempty"`
	// ExtraFields 存储未知字段（仅顶层）
	ExtraFields map[string]json.RawMessage `json:"-"`
}

// requestExplicitFields 定义 Request 结构体的所有显式字段名称
// 用于在序列化/反序列化时识别已知字段
var requestExplicitFields = map[string]bool{
	"contents":          true,
	"systemInstruction": true,
	"generationConfig":  true,
	"safetySettings":    true,
	"tools":             true,
	"toolConfig":        true,
	"cachedContent":     true,
}

// MarshalJSON 实现 Request 的序列化，合并显式字段与 ExtraFields
func (r Request) MarshalJSON() ([]byte, error) {
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
		if !requestExplicitFields[key] {
			result[key] = value
		}
	}

	return json.Marshal(result)
}

// UnmarshalJSON 实现 Request 的反序列化，收集未知字段到 ExtraFields
func (r *Request) UnmarshalJSON(data []byte) error {
	// 解析为 map 以便收集未知字段
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 收集未知字段，只有存在未知字段时才初始化 ExtraFields
	var unknownFields map[string]json.RawMessage
	for key, value := range raw {
		if !requestExplicitFields[key] {
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
