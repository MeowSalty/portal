package types_test

import (
	"encoding/json"
	"testing"

	"github.com/MeowSalty/portal/types"
)

// TestResponseMessage_UnmarshalJSON_ExtraFields 测试 ResponseMessage 反序列化时捕获未知字段
func TestResponseMessage_UnmarshalJSON_ExtraFields(t *testing.T) {
	// 构造包含未知字段的 JSON
	jsonData := `{
		"role": "assistant",
		"content": "Core message",
		"core_unknown1": "core_value1",
		"core_unknown2": 111
	}`

	var msg types.ResponseMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证已知字段
	if msg.Role != "assistant" {
		t.Errorf("期望角色为 'assistant'，实际为 '%s'", msg.Role)
	}

	if msg.Content == nil || *msg.Content != "Core message" {
		t.Errorf("期望内容为 'Core message'，实际为 '%v'", msg.Content)
	}

	// 验证未知字段被正确捕获
	if len(msg.ExtraFields) != 2 {
		t.Fatalf("期望捕获 2 个未知字段，实际为 %d", len(msg.ExtraFields))
	}

	// 验证未知字段的值
	if msg.ExtraFields["core_unknown1"] != "core_value1" {
		t.Errorf("期望 core_unknown1 为 'core_value1'，实际为 '%v'", msg.ExtraFields["core_unknown1"])
	}

	if msg.ExtraFields["core_unknown2"] != float64(111) {
		t.Errorf("期望 core_unknown2 为 111，实际为 '%v'", msg.ExtraFields["core_unknown2"])
	}

	// 验证来源格式
	if msg.ExtraFieldsFormat != "" {
		t.Errorf("期望来源格式为空，实际为 '%s'", msg.ExtraFieldsFormat)
	}
}

// TestResponseMessage_MarshalJSON_ExtraFields 测试 ResponseMessage 序列化时包含未知字段
func TestResponseMessage_MarshalJSON_ExtraFields(t *testing.T) {
	// 构造包含未知字段的 ResponseMessage
	msg := types.ResponseMessage{
		Role:    "assistant",
		Content: stringPtr("Core test message"),
		ExtraFields: map[string]interface{}{
			"core_custom1": "core_custom_value1",
			"core_custom2": 222,
		},
	}
	msg.ExtraFields["core_custom3"] = []string{"a", "b"}

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
	if result["core_custom1"] != "core_custom_value1" {
		t.Errorf("期望 core_custom1 为 'core_custom_value1'，实际为 '%v'", result["core_custom1"])
	}

	if result["core_custom2"] != float64(222) {
		t.Errorf("期望 core_custom2 为 222，实际为 '%v'", result["core_custom2"])
	}

	// 验证嵌套数组字段
	nestedArray, ok := result["core_custom3"].([]interface{})
	if !ok {
		t.Error("期望 core_custom3 为数组")
	} else if len(nestedArray) != 2 {
		t.Errorf("期望嵌套数组长度为 2，实际为 %d", len(nestedArray))
	} else if nestedArray[0] != "a" || nestedArray[1] != "b" {
		t.Errorf("期望嵌套数组内容为 ['a', 'b']，实际为 '%v'", nestedArray)
	}
}

// TestDelta_UnmarshalJSON_ExtraFields 测试 Delta 反序列化时捕获未知字段
func TestDelta_UnmarshalJSON_ExtraFields(t *testing.T) {
	// 构造包含未知字段的 JSON
	jsonData := `{
		"role": "assistant",
		"content": "Core streaming",
		"core_delta_unknown1": "delta_value1",
		"core_delta_unknown2": 333
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
	if delta.ExtraFields["core_delta_unknown1"] != "delta_value1" {
		t.Errorf("期望 core_delta_unknown1 为 'delta_value1'，实际为 '%v'", delta.ExtraFields["core_delta_unknown1"])
	}

	if delta.ExtraFields["core_delta_unknown2"] != float64(333) {
		t.Errorf("期望 core_delta_unknown2 为 333，实际为 '%v'", delta.ExtraFields["core_delta_unknown2"])
	}

	// 验证来源格式
	if delta.ExtraFieldsFormat != "" {
		t.Errorf("期望来源格式为空，实际为 '%s'", delta.ExtraFieldsFormat)
	}
}

// TestDelta_MarshalJSON_ExtraFields 测试 Delta 序列化时包含未知字段
func TestDelta_MarshalJSON_ExtraFields(t *testing.T) {
	// 构造包含未知字段的 Delta
	delta := types.Delta{
		Role:    stringPtr("assistant"),
		Content: stringPtr("Core streaming content"),
		ExtraFields: map[string]interface{}{
			"core_delta_field1": "core_delta_value1",
			"core_delta_field2": 444,
		},
	}
	delta.ExtraFields["core_delta_field3"] = map[string]string{"key": "value"}

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
	if result["core_delta_field1"] != "core_delta_value1" {
		t.Errorf("期望 core_delta_field1 为 'core_delta_value1'，实际为 '%v'", result["core_delta_field1"])
	}

	if result["core_delta_field2"] != float64(444) {
		t.Errorf("期望 core_delta_field2 为 444，实际为 '%v'", result["core_delta_field2"])
	}

	// 验证嵌套字段
	nested, ok := result["core_delta_field3"].(map[string]interface{})
	if !ok {
		t.Error("期望 core_delta_field3 为嵌套对象")
	} else if nested["key"] != "value" {
		t.Errorf("期望嵌套字段 key 为 'value'，实际为 '%v'", nested["key"])
	}
}

// TestResponseMessage_MarshalJSON_NoExtraFields 测试没有未知字段时的序列化
func TestResponseMessage_MarshalJSON_NoExtraFields(t *testing.T) {
	// 构造没有未知字段的 ResponseMessage
	msg := types.ResponseMessage{
		Role:    "assistant",
		Content: stringPtr("Core test without extras"),
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
	if _, exists := result["extra_field"]; exists {
		t.Error("期望没有未知字段被添加")
	}
}

// TestDelta_MarshalJSON_NoExtraFields 测试没有未知字段时的序列化
func TestDelta_MarshalJSON_NoExtraFields(t *testing.T) {
	// 构造没有未知字段的 Delta
	delta := types.Delta{
		Role:    stringPtr("assistant"),
		Content: stringPtr("Core streaming without extras"),
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
