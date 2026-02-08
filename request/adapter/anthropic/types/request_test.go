package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRequestUnmarshalWithExtraFields 测试反序列化时收集未知字段到 ExtraFields。
func TestRequestUnmarshalWithExtraFields(t *testing.T) {
	jsonStr := `{
		"model": "claude-3-5-sonnet-20241022",
		"messages": [{"role": "user", "content": "hello"}],
		"max_tokens": 1024,
		"unknown_field_1": "value1",
		"unknown_field_2": 123,
		"unknown_field_3": {"nested": "value"}
	}`

	var req Request
	err := json.Unmarshal([]byte(jsonStr), &req)
	assert.NoError(t, err)

	// 验证显式字段
	assert.Equal(t, "claude-3-5-sonnet-20241022", req.Model)
	assert.Equal(t, 1024, req.MaxTokens)
	assert.Len(t, req.Messages, 1)

	// 验证 ExtraFields 收集了未知字段
	assert.NotNil(t, req.ExtraFields)
	assert.Len(t, req.ExtraFields, 3)

	// 验证未知字段值
	assert.Contains(t, req.ExtraFields, "unknown_field_1")
	assert.Contains(t, req.ExtraFields, "unknown_field_2")
	assert.Contains(t, req.ExtraFields, "unknown_field_3")

	// 验证原始 JSON 值
	var val1 string
	assert.NoError(t, json.Unmarshal(req.ExtraFields["unknown_field_1"], &val1))
	assert.Equal(t, "value1", val1)

	var val2 int
	assert.NoError(t, json.Unmarshal(req.ExtraFields["unknown_field_2"], &val2))
	assert.Equal(t, 123, val2)
}

// TestRequestMarshalWithExtraFields 测试序列化时合并 ExtraFields。
func TestRequestMarshalWithExtraFields(t *testing.T) {
	req := Request{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		Messages: []Message{
			{Role: RoleUser, Content: MessageContentParam{StringValue: strPtr("hello")}},
		},
		ExtraFields: map[string]json.RawMessage{
			"unknown_field_1": json.RawMessage(`"value1"`),
			"unknown_field_2": json.RawMessage(`123`),
		},
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)

	// 验证输出包含显式字段
	assert.Contains(t, string(data), `"model"`)
	assert.Contains(t, string(data), `"claude-3-5-sonnet-20241022"`)
	assert.Contains(t, string(data), `"max_tokens"`)
	assert.Contains(t, string(data), `1024`)

	// 验证输出包含 ExtraFields
	assert.Contains(t, string(data), `"unknown_field_1"`)
	assert.Contains(t, string(data), `"value1"`)
	assert.Contains(t, string(data), `"unknown_field_2"`)
	assert.Contains(t, string(data), `123`)

	// 验证可以完整反序列化回来
	var req2 Request
	err = json.Unmarshal(data, &req2)
	assert.NoError(t, err)
	assert.Equal(t, req.Model, req2.Model)
	assert.Equal(t, req.MaxTokens, req2.MaxTokens)
	assert.Len(t, req2.ExtraFields, 2)
}

// TestRequestExplicitFieldPriority 测试显式字段优先于 ExtraFields。
func TestRequestExplicitFieldPriority(t *testing.T) {
	// 输入 JSON 中 model 字段与 ExtraFields 中的 model 冲突
	jsonStr := `{
		"model": "claude-3-5-sonnet-20241022",
		"messages": [{"role": "user", "content": "hello"}],
		"max_tokens": 1024
	}`

	var req Request
	err := json.Unmarshal([]byte(jsonStr), &req)
	assert.NoError(t, err)

	// 显式字段应该被正确解析
	assert.Equal(t, "claude-3-5-sonnet-20241022", req.Model)

	// ExtraFields 不应包含显式字段名
	assert.NotContains(t, req.ExtraFields, "model")
	assert.NotContains(t, req.ExtraFields, "messages")
	assert.NotContains(t, req.ExtraFields, "max_tokens")
}

// TestRequestMarshalConflictPriority 测试序列化时显式字段优先。
func TestRequestMarshalConflictPriority(t *testing.T) {
	req := Request{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		ExtraFields: map[string]json.RawMessage{
			"model":      json.RawMessage(`"should_be_ignored"`),
			"max_tokens": json.RawMessage(`999`),
			"unknown":    json.RawMessage(`"kept"`),
		},
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)

	// 验证显式字段值被使用
	assert.Contains(t, string(data), `"claude-3-5-sonnet-20241022"`)
	assert.Contains(t, string(data), `1024`)

	// 验证 ExtraFields 中的冲突字段被忽略
	assert.NotContains(t, string(data), `"should_be_ignored"`)
	assert.NotContains(t, string(data), `999`)

	// 验证非冲突字段被保留
	assert.Contains(t, string(data), `"unknown"`)
	assert.Contains(t, string(data), `"kept"`)
}

// TestRequestEmptyExtraFields 测试空 ExtraFields 不影响序列化。
func TestRequestEmptyExtraFields(t *testing.T) {
	req := Request{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		Messages: []Message{
			{Role: RoleUser, Content: MessageContentParam{StringValue: strPtr("hello")}},
		},
		ExtraFields: map[string]json.RawMessage{},
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)

	// 验证输出不包含 ExtraFields 相关内容
	assert.NotContains(t, string(data), `"ExtraFields"`)

	// 验证可以正常反序列化
	var req2 Request
	err = json.Unmarshal(data, &req2)
	assert.NoError(t, err)
	assert.Equal(t, req.Model, req2.Model)
	assert.Nil(t, req2.ExtraFields)
}

// TestRequestNilExtraFields 测试 nil ExtraFields 不影响序列化。
func TestRequestNilExtraFields(t *testing.T) {
	req := Request{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		Messages: []Message{
			{Role: RoleUser, Content: MessageContentParam{StringValue: strPtr("hello")}},
		},
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)

	// 验证可以正常反序列化
	var req2 Request
	err = json.Unmarshal(data, &req2)
	assert.NoError(t, err)
	assert.Equal(t, req.Model, req2.Model)
	assert.Nil(t, req2.ExtraFields)
}

// TestRequestRoundTrip 测试完整的序列化/反序列化往返。
func TestRequestRoundTrip(t *testing.T) {
	originalJSON := `{
		"model": "claude-3-5-sonnet-20241022",
		"messages": [{"role": "user", "content": "hello"}],
		"max_tokens": 1024,
		"temperature": 0.7,
		"unknown_field": "unknown_value",
		"another_unknown": 42
	}`

	// 第一次反序列化
	var req1 Request
	err := json.Unmarshal([]byte(originalJSON), &req1)
	assert.NoError(t, err)

	// 序列化
	data, err := json.Marshal(req1)
	assert.NoError(t, err)

	// 第二次反序列化
	var req2 Request
	err = json.Unmarshal(data, &req2)
	assert.NoError(t, err)

	// 验证显式字段一致
	assert.Equal(t, req1.Model, req2.Model)
	assert.Equal(t, req1.MaxTokens, req2.MaxTokens)
	assert.Equal(t, req1.Temperature, req2.Temperature)

	// 验证 ExtraFields 一致
	assert.Len(t, req2.ExtraFields, 2)
	assert.Contains(t, req2.ExtraFields, "unknown_field")
	assert.Contains(t, req2.ExtraFields, "another_unknown")
}

// 辅助函数：创建字符串指针
func strPtr(s string) *string {
	return &s
}
