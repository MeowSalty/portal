package types_test

import (
	"encoding/json"
	"testing"

	"github.com/MeowSalty/portal/request/adapter/openai/types"
)

// TestMessage_UnmarshalJSON_ExtraFields 测试 Message 反序列化时捕获未知字段
func TestMessage_UnmarshalJSON_ExtraFields(t *testing.T) {
	// 构造包含未知字段的 JSON
	jsonData := `{
		"role": "assistant",
		"content": "Hello!",
		"unknown_field1": "value1",
		"unknown_field2": 123,
		"unknown_field3": true
	}`

	var msg types.Message
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证已知字段
	if msg.Role != "assistant" {
		t.Errorf("期望角色为 'assistant'，实际为 '%s'", msg.Role)
	}

	if msg.Content == nil || *msg.Content != "Hello!" {
		t.Errorf("期望内容为 'Hello!'，实际为 '%v'", msg.Content)
	}

	// 验证未知字段被正确捕获
	if len(msg.ExtraFields) != 3 {
		t.Fatalf("期望捕获 3 个未知字段，实际为 %d", len(msg.ExtraFields))
	}

	// 验证未知字段的值
	if msg.ExtraFields["unknown_field1"] != "value1" {
		t.Errorf("期望 unknown_field1 为 'value1'，实际为 '%v'", msg.ExtraFields["unknown_field1"])
	}

	if msg.ExtraFields["unknown_field2"] != float64(123) {
		t.Errorf("期望 unknown_field2 为 123，实际为 '%v'", msg.ExtraFields["unknown_field2"])
	}

	if msg.ExtraFields["unknown_field3"] != true {
		t.Errorf("期望 unknown_field3 为 true，实际为 '%v'", msg.ExtraFields["unknown_field3"])
	}
}

// TestMessage_MarshalJSON_ExtraFields 测试 Message 序列化时包含未知字段
func TestMessage_MarshalJSON_ExtraFields(t *testing.T) {
	// 构造包含未知字段的 Message
	msg := types.Message{
		Role:    "assistant",
		Content: stringPtr("Hello!"),
		ExtraFields: map[string]interface{}{
			"custom_field1": "custom_value1",
			"custom_field2": 456,
		},
	}
	msg.ExtraFields["custom_field3"] = false

	// 序列化
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 解析序列化结果
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("解析序列化结果失败: %v", err)
	}

	// 验证已知字段
	if result["role"] != "assistant" {
		t.Errorf("期望角色为 'assistant'，实际为 '%v'", result["role"])
	}

	// 验证未知字段被正确序列化
	if result["custom_field1"] != "custom_value1" {
		t.Errorf("期望 custom_field1 为 'custom_value1'，实际为 '%v'", result["custom_field1"])
	}

	if result["custom_field2"] != float64(456) {
		t.Errorf("期望 custom_field2 为 456，实际为 '%v'", result["custom_field2"])
	}

	if result["custom_field3"] != false {
		t.Errorf("期望 custom_field3 为 false，实际为 '%v'", result["custom_field3"])
	}
}

// TestDelta_UnmarshalJSON_ExtraFields 测试 Delta 反序列化时捕获未知字段
func TestDelta_UnmarshalJSON_ExtraFields(t *testing.T) {
	// 构造包含未知字段的 JSON
	jsonData := `{
		"role": "assistant",
		"content": "Streaming response",
		"stream_field1": "stream_value1",
		"stream_field2": 789
	}`

	var delta types.Delta
	err := json.Unmarshal([]byte(jsonData), &delta)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证已知字段
	if delta.Role == nil || *delta.Role != "assistant" {
		t.Errorf("期望角色为 'assistant'，实际为 '%v'", delta.Role)
	}

	// 验证未知字段被正确捕获
	if len(delta.ExtraFields) != 2 {
		t.Fatalf("期望捕获 2 个未知字段，实际为 %d", len(delta.ExtraFields))
	}

	// 验证未知字段的值
	if delta.ExtraFields["stream_field1"] != "stream_value1" {
		t.Errorf("期望 stream_field1 为 'stream_value1'，实际为 '%v'", delta.ExtraFields["stream_field1"])
	}

	if delta.ExtraFields["stream_field2"] != float64(789) {
		t.Errorf("期望 stream_field2 为 789，实际为 '%v'", delta.ExtraFields["stream_field2"])
	}
}

// TestDelta_MarshalJSON_ExtraFields 测试 Delta 序列化时包含未知字段
func TestDelta_MarshalJSON_ExtraFields(t *testing.T) {
	// 构造包含未知字段的 Delta
	delta := types.Delta{
		Role:    stringPtr("assistant"),
		Content: stringPtr("Streaming content"),
		ExtraFields: map[string]interface{}{
			"delta_field1": "delta_value1",
			"delta_field2": 999,
		},
	}
	delta.ExtraFields["delta_field3"] = map[string]interface{}{"nested": "value"}

	// 序列化
	data, err := json.Marshal(delta)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 解析序列化结果
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("解析序列化结果失败: %v", err)
	}

	// 验证未知字段被正确序列化
	if result["delta_field1"] != "delta_value1" {
		t.Errorf("期望 delta_field1 为 'delta_value1'，实际为 '%v'", result["delta_field1"])
	}

	if result["delta_field2"] != float64(999) {
		t.Errorf("期望 delta_field2 为 999，实际为 '%v'", result["delta_field2"])
	}

	// 验证嵌套字段
	nested, ok := result["delta_field3"].(map[string]interface{})
	if !ok {
		t.Error("期望 delta_field3 为嵌套对象")
	} else if nested["nested"] != "value" {
		t.Errorf("期望嵌套字段 nested 为 'value'，实际为 '%v'", nested["nested"])
	}
}

// TestMessage_MarshalJSON_NoExtraFields 测试没有未知字段时的序列化
func TestMessage_MarshalJSON_NoExtraFields(t *testing.T) {
	// 构造没有未知字段的 Message
	msg := types.Message{
		Role:    "assistant",
		Content: stringPtr("Test message"),
	}

	// 序列化
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 解析序列化结果
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("解析序列化结果失败: %v", err)
	}

	// 验证没有未知字段被添加
	if _, exists := result["unknown_field"]; exists {
		t.Error("期望没有未知字段被添加")
	}
}

// TestDelta_MarshalJSON_NoExtraFields 测试没有未知字段时的序列化
func TestDelta_MarshalJSON_NoExtraFields(t *testing.T) {
	// 构造没有未知字段的 Delta
	delta := types.Delta{
		Role:    stringPtr("assistant"),
		Content: stringPtr("Streaming test"),
	}

	// 序列化
	data, err := json.Marshal(delta)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 解析序列化结果
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("解析序列化结果失败: %v", err)
	}

	// 验证没有未知字段被添加
	if _, exists := result["extra_field"]; exists {
		t.Error("期望没有未知字段被添加")
	}
}

// 辅助函数：创建字符串指针
func stringPtr(s string) *string {
	return &s
}
