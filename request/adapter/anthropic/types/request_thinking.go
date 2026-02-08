package types

import (
	"encoding/json"
	"fmt"
)

// ThinkingConfigType 思考配置类型。
type ThinkingConfigType string

const (
	ThinkingConfigTypeEnabled  ThinkingConfigType = "enabled"
	ThinkingConfigTypeDisabled ThinkingConfigType = "disabled"
)

// ThinkingConfigEnabled 启用思考配置。
type ThinkingConfigEnabled struct {
	Type         ThinkingConfigType `json:"type"`          // "enabled"
	BudgetTokens int                `json:"budget_tokens"` // 预算 token
}

// ThinkingConfigDisabled 禁用思考配置。
type ThinkingConfigDisabled struct {
	Type ThinkingConfigType `json:"type"` // "disabled"
}

// ThinkingConfigParam 思考配置联合类型。
type ThinkingConfigParam struct {
	Enabled  *ThinkingConfigEnabled
	Disabled *ThinkingConfigDisabled
}

// MarshalJSON 实现 ThinkingConfigParam 的序列化。
func (c ThinkingConfigParam) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}
	if c.Enabled != nil {
		set(c.Enabled)
	}
	if c.Disabled != nil {
		set(c.Disabled)
	}
	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("thinking 只能设置一种类型")
	}
	return json.Marshal(payload)
}

// UnmarshalJSON 实现 ThinkingConfigParam 的反序列化。
func (c *ThinkingConfigParam) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type ThinkingConfigType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("thinking 解析失败：%w", err)
	}

	switch t.Type {
	case ThinkingConfigTypeEnabled:
		var v ThinkingConfigEnabled
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("thinking enabled 解析失败：%w", err)
		}
		c.Enabled = &v
	case ThinkingConfigTypeDisabled:
		var v ThinkingConfigDisabled
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("thinking disabled 解析失败：%w", err)
		}
		c.Disabled = &v
	default:
		return fmt.Errorf("不支持的 thinking 类型: %s", t.Type)
	}

	return nil
}
