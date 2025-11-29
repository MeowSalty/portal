package types

import (
	"encoding/json"
	"testing"
)

// TestRequest_UnmarshalJSON_WithExtraFields 测试解析包含未知字段的 JSON
func TestRequest_UnmarshalJSON_WithExtraFields(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"messages": [
			{"role": "user", "content": "Hello"}
		],
		"temperature": 0.7,
		"custom_field": "custom_value",
		"another_field": 123,
		"nested_field": {
			"key": "value"
		}
	}`

	var req Request
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证已知字段
	if req.Model != "gpt-4" {
		t.Errorf("期望 Model 为 'gpt-4', 实际为 '%s'", req.Model)
	}
	if len(req.Messages) != 1 {
		t.Errorf("期望 Messages 长度为 1, 实际为 %d", len(req.Messages))
	}
	if req.Temperature == nil || *req.Temperature != 0.7 {
		t.Errorf("期望 Temperature 为 0.7, 实际为 %v", req.Temperature)
	}

	// 验证未知字段
	if len(req.ExtraFields) != 3 {
		t.Errorf("期望 ExtraFields 长度为 3, 实际为 %d", len(req.ExtraFields))
	}

	if val, ok := req.ExtraFields["custom_field"]; !ok || val != "custom_value" {
		t.Errorf("期望 custom_field 为 'custom_value', 实际为 %v", val)
	}

	if val, ok := req.ExtraFields["another_field"]; !ok {
		t.Error("期望 another_field 存在")
	} else if floatVal, ok := val.(float64); !ok || floatVal != 123 {
		t.Errorf("期望 another_field 为 123, 实际为 %v (类型: %T)", val, val)
	}

	if val, ok := req.ExtraFields["nested_field"]; !ok {
		t.Error("期望 nested_field 存在")
	} else if nested, ok := val.(map[string]interface{}); !ok {
		t.Errorf("期望 nested_field 为 map, 实际类型为 %T", val)
	} else if nested["key"] != "value" {
		t.Errorf("期望 nested_field.key 为 'value', 实际为 %v", nested["key"])
	}
}

// TestRequest_UnmarshalJSON_WithoutExtraFields 测试解析不包含未知字段的 JSON
func TestRequest_UnmarshalJSON_WithoutExtraFields(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"messages": [
			{"role": "user", "content": "Hello"}
		],
		"temperature": 0.7
	}`

	var req Request
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证已知字段
	if req.Model != "gpt-4" {
		t.Errorf("期望 Model 为 'gpt-4', 实际为 '%s'", req.Model)
	}

	// 验证 ExtraFields 为空
	if len(req.ExtraFields) != 0 {
		t.Errorf("期望 ExtraFields 为空，实际长度为 %d", len(req.ExtraFields))
	}
}

// TestRequest_MarshalJSON_WithExtraFields 测试序列化包含未知字段的请求
func TestRequest_MarshalJSON_WithExtraFields(t *testing.T) {
	req := Request{
		Model: "gpt-4",
		Messages: []RequestMessage{
			{Role: "user", Content: "Hello"},
		},
		ExtraFields: map[string]interface{}{
			"custom_field":  "custom_value",
			"another_field": 123,
			"nested_field": map[string]interface{}{
				"key": "value",
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化以验证
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("验证序列化结果失败: %v", err)
	}

	// 验证已知字段
	if result["model"] != "gpt-4" {
		t.Errorf("期望 model 为 'gpt-4', 实际为 %v", result["model"])
	}

	// 验证未知字段被包含
	if result["custom_field"] != "custom_value" {
		t.Errorf("期望 custom_field 为 'custom_value', 实际为 %v", result["custom_field"])
	}

	if val, ok := result["another_field"].(float64); !ok || val != 123 {
		t.Errorf("期望 another_field 为 123, 实际为 %v", result["another_field"])
	}

	if nested, ok := result["nested_field"].(map[string]interface{}); !ok {
		t.Errorf("期望 nested_field 为 map, 实际类型为 %T", result["nested_field"])
	} else if nested["key"] != "value" {
		t.Errorf("期望 nested_field.key 为 'value', 实际为 %v", nested["key"])
	}
}

// TestRequest_MarshalJSON_WithoutExtraFields 测试序列化不包含未知字段的请求
func TestRequest_MarshalJSON_WithoutExtraFields(t *testing.T) {
	temp := 0.7
	req := Request{
		Model: "gpt-4",
		Messages: []RequestMessage{
			{Role: "user", Content: "Hello"},
		},
		Temperature: &temp,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化以验证
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("验证序列化结果失败: %v", err)
	}

	// 验证已知字段
	if result["model"] != "gpt-4" {
		t.Errorf("期望 model 为 'gpt-4', 实际为 %v", result["model"])
	}

	if result["temperature"] != 0.7 {
		t.Errorf("期望 temperature 为 0.7, 实际为 %v", result["temperature"])
	}

	// 验证不包含 ExtraFields 相关的内容
	// (ExtraFields 本身有 json:"-" 标签，不应出现)
	if _, exists := result["ExtraFields"]; exists {
		t.Error("不应包含 ExtraFields 字段")
	}
}

// TestRequest_RoundTrip 测试序列化和反序列化的往返转换
func TestRequest_RoundTrip(t *testing.T) {
	originalJSON := `{
		"model": "gpt-4",
		"messages": [
			{"role": "user", "content": "Hello"}
		],
		"temperature": 0.7,
		"max_tokens": 100,
		"custom_field": "custom_value",
		"another_field": 123,
		"array_field": [1, 2, 3]
	}`

	// 第一次反序列化
	var req1 Request
	err := json.Unmarshal([]byte(originalJSON), &req1)
	if err != nil {
		t.Fatalf("第一次反序列化失败: %v", err)
	}

	// 序列化
	data, err := json.Marshal(req1)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 第二次反序列化
	var req2 Request
	err = json.Unmarshal(data, &req2)
	if err != nil {
		t.Fatalf("第二次反序列化失败: %v", err)
	}

	// 验证已知字段保持一致
	if req1.Model != req2.Model {
		t.Errorf("Model 不一致: %s vs %s", req1.Model, req2.Model)
	}
	if *req1.Temperature != *req2.Temperature {
		t.Errorf("Temperature 不一致：%f vs %f", *req1.Temperature, *req2.Temperature)
	}
	if *req1.MaxTokens != *req2.MaxTokens {
		t.Errorf("MaxTokens 不一致：%d vs %d", *req1.MaxTokens, *req2.MaxTokens)
	}

	// 验证未知字段保持一致
	if len(req1.ExtraFields) != len(req2.ExtraFields) {
		t.Errorf("ExtraFields 长度不一致：%d vs %d", len(req1.ExtraFields), len(req2.ExtraFields))
	}

	if req1.ExtraFields["custom_field"] != req2.ExtraFields["custom_field"] {
		t.Errorf("custom_field 不一致: %v vs %v", req1.ExtraFields["custom_field"], req2.ExtraFields["custom_field"])
	}
}

// TestRequest_ExtraFields_VariousTypes 测试各种类型的未知字段
func TestRequest_ExtraFields_VariousTypes(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"messages": [{"role": "user", "content": "test"}],
		"string_field": "text",
		"number_field": 42.5,
		"bool_field": true,
		"null_field": null,
		"array_field": ["a", "b", "c"],
		"object_field": {"nested": "value"}
	}`

	var req Request
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	tests := []struct {
		name      string
		fieldName string
		wantType  string
		check     func(interface{}) bool
	}{
		{
			name:      "字符串类型",
			fieldName: "string_field",
			wantType:  "string",
			check:     func(v interface{}) bool { _, ok := v.(string); return ok },
		},
		{
			name:      "数字类型",
			fieldName: "number_field",
			wantType:  "float64",
			check:     func(v interface{}) bool { _, ok := v.(float64); return ok },
		},
		{
			name:      "布尔类型",
			fieldName: "bool_field",
			wantType:  "bool",
			check:     func(v interface{}) bool { _, ok := v.(bool); return ok },
		},
		{
			name:      "null 类型",
			fieldName: "null_field",
			wantType:  "nil",
			check:     func(v interface{}) bool { return v == nil },
		},
		{
			name:      "数组类型",
			fieldName: "array_field",
			wantType:  "[]interface{}",
			check:     func(v interface{}) bool { _, ok := v.([]interface{}); return ok },
		},
		{
			name:      "对象类型",
			fieldName: "object_field",
			wantType:  "map[string]interface{}",
			check:     func(v interface{}) bool { _, ok := v.(map[string]interface{}); return ok },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, exists := req.ExtraFields[tt.fieldName]
			if !exists {
				t.Errorf("字段 %s 不存在", tt.fieldName)
				return
			}
			if !tt.check(val) {
				t.Errorf("字段 %s 类型错误, 期望 %s, 实际为 %T", tt.fieldName, tt.wantType, val)
			}
		})
	}
}

// TestRequestMessage_UnmarshalJSON_WithExtraFields 测试 RequestMessage 解析包含未知字段的 JSON
func TestRequestMessage_UnmarshalJSON_WithExtraFields(t *testing.T) {
	jsonData := `{
		"role": "user",
		"content": "Hello",
		"custom_field": "custom_value",
		"another_field": 123,
		"nested_field": {
			"key": "value"
		}
	}`

	var msg RequestMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证已知字段
	if msg.Role != "user" {
		t.Errorf("期望 Role 为 'user', 实际为 '%s'", msg.Role)
	}
	if msg.Content != "Hello" {
		t.Errorf("期望 Content 为 'Hello', 实际为 '%v'", msg.Content)
	}

	// 验证未知字段
	if len(msg.ExtraFields) != 3 {
		t.Errorf("期望 ExtraFields 长度为 3, 实际为 %d", len(msg.ExtraFields))
	}

	if val, ok := msg.ExtraFields["custom_field"]; !ok || val != "custom_value" {
		t.Errorf("期望 custom_field 为 'custom_value', 实际为 %v", val)
	}

	if val, ok := msg.ExtraFields["another_field"]; !ok {
		t.Error("期望 another_field 存在")
	} else if floatVal, ok := val.(float64); !ok || floatVal != 123 {
		t.Errorf("期望 another_field 为 123, 实际为 %v", val)
	}

	if val, ok := msg.ExtraFields["nested_field"]; !ok {
		t.Error("期望 nested_field 存在")
	} else if nested, ok := val.(map[string]interface{}); !ok {
		t.Errorf("期望 nested_field 为 map, 实际类型为 %T", val)
	} else if nested["key"] != "value" {
		t.Errorf("期望 nested_field.key 为 'value', 实际为 %v", nested["key"])
	}
}

// TestRequestMessage_UnmarshalJSON_WithoutExtraFields 测试 RequestMessage 解析不包含未知字段的 JSON
func TestRequestMessage_UnmarshalJSON_WithoutExtraFields(t *testing.T) {
	jsonData := `{
		"role": "user",
		"content": "Hello"
	}`

	var msg RequestMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证已知字段
	if msg.Role != "user" {
		t.Errorf("期望 Role 为 'user', 实际为 '%s'", msg.Role)
	}
	if msg.Content != "Hello" {
		t.Errorf("期望 Content 为 'Hello', 实际为 '%v'", msg.Content)
	}

	// 验证 ExtraFields 为空
	if len(msg.ExtraFields) != 0 {
		t.Errorf("期望 ExtraFields 为空，实际长度为 %d", len(msg.ExtraFields))
	}
}

// TestRequestMessage_MarshalJSON_WithExtraFields 测试 RequestMessage 序列化包含未知字段的请求
func TestRequestMessage_MarshalJSON_WithExtraFields(t *testing.T) {
	msg := RequestMessage{
		Role:    "user",
		Content: "Hello",
		ExtraFields: map[string]interface{}{
			"custom_field":  "custom_value",
			"another_field": 123,
			"nested_field": map[string]interface{}{
				"key": "value",
			},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化以验证
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("验证序列化结果失败: %v", err)
	}

	// 验证已知字段
	if result["role"] != "user" {
		t.Errorf("期望 role 为 'user', 实际为 %v", result["role"])
	}

	// 验证未知字段被包含
	if result["custom_field"] != "custom_value" {
		t.Errorf("期望 custom_field 为 'custom_value', 实际为 %v", result["custom_field"])
	}

	if val, ok := result["another_field"].(float64); !ok || val != 123 {
		t.Errorf("期望 another_field 为 123, 实际为 %v", result["another_field"])
	}

	if nested, ok := result["nested_field"].(map[string]interface{}); !ok {
		t.Errorf("期望 nested_field 为 map, 实际类型为 %T", result["nested_field"])
	} else if nested["key"] != "value" {
		t.Errorf("期望 nested_field.key 为 'value', 实际为 %v", nested["key"])
	}
}

// TestRequestMessage_MarshalJSON_WithoutExtraFields 测试 RequestMessage 序列化不包含未知字段的请求
func TestRequestMessage_MarshalJSON_WithoutExtraFields(t *testing.T) {
	msg := RequestMessage{
		Role:    "user",
		Content: "Hello",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化以验证
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("验证序列化结果失败: %v", err)
	}

	// 验证已知字段
	if result["role"] != "user" {
		t.Errorf("期望 role 为 'user', 实际为 %v", result["role"])
	}

	// 验证不包含 ExtraFields 相关的内容
	// (ExtraFields 本身有 json:"-" 标签，不应出现)
	if _, exists := result["ExtraFields"]; exists {
		t.Error("不应包含 ExtraFields 字段")
	}
}

// TestRequestMessage_RoundTrip 测试 RequestMessage 序列化和反序列化的往返转换
func TestRequestMessage_RoundTrip(t *testing.T) {
	originalJSON := `{
		"role": "user",
		"content": "Hello",
		"custom_field": "custom_value",
		"another_field": 123,
		"array_field": [1, 2, 3]
	}`

	// 第一次反序列化
	var msg1 RequestMessage
	err := json.Unmarshal([]byte(originalJSON), &msg1)
	if err != nil {
		t.Fatalf("第一次反序列化失败: %v", err)
	}

	// 序列化
	data, err := json.Marshal(msg1)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 第二次反序列化
	var msg2 RequestMessage
	err = json.Unmarshal(data, &msg2)
	if err != nil {
		t.Fatalf("第二次反序列化失败: %v", err)
	}

	// 验证已知字段保持一致
	if msg1.Role != msg2.Role {
		t.Errorf("Role 不一致: %s vs %s", msg1.Role, msg2.Role)
	}
	if msg1.Content != msg2.Content {
		t.Errorf("Content 不一致: %v vs %v", msg1.Content, msg2.Content)
	}

	// 验证未知字段保持一致
	if len(msg1.ExtraFields) != len(msg2.ExtraFields) {
		t.Errorf("ExtraFields 长度不一致：%d vs %d", len(msg1.ExtraFields), len(msg2.ExtraFields))
	}

	if msg1.ExtraFields["custom_field"] != msg2.ExtraFields["custom_field"] {
		t.Errorf("custom_field 不一致: %v vs %v", msg1.ExtraFields["custom_field"], msg2.ExtraFields["custom_field"])
		t.Errorf("custom_field 不一致: %v vs %v", msg1.ExtraFields["custom_field"], msg2.ExtraFields["custom_field"])
	}
}
