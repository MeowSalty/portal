package types

import (
	"encoding/json"
	"testing"
)

// stringPtr 是一个辅助函数，用于创建字符串指针
func stringPtr(s string) *string {
	return &s
}

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
	}
}

// TestToolChoiceUnion_MarshalJSON_Auto 测试 ToolChoiceUnion 的 Auto 字符串模式序列化
func TestToolChoiceUnion_MarshalJSON_Auto(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "auto 模式",
			value:    "auto",
			expected: `"auto"`,
		},
		{
			name:     "none 模式",
			value:    "none",
			expected: `"none"`,
		},
		{
			name:     "required 模式",
			value:    "required",
			expected: `"required"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolChoice := ToolChoiceUnion{
				Auto: &tt.value,
			}

			data, err := json.Marshal(toolChoice)
			if err != nil {
				t.Fatalf("序列化失败: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("期望序列化结果为 %s, 实际为 %s", tt.expected, string(data))
			}
		})
	}
}

// TestToolChoiceUnion_MarshalJSON_Named 测试 ToolChoiceUnion 的 Named 函数模式序列化
func TestToolChoiceUnion_MarshalJSON_Named(t *testing.T) {
	toolChoice := ToolChoiceUnion{
		Named: &ToolChoiceNamed{
			Type: "function",
			Function: struct {
				Name string `json:"name"`
			}{
				Name: "get_weather",
			},
		},
	}

	data, err := json.Marshal(toolChoice)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化以验证结构
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("验证序列化结果失败: %v", err)
	}

	// 验证 type 字段
	if result["type"] != "function" {
		t.Errorf("期望 type 为 'function', 实际为 %v", result["type"])
	}

	// 验证 function 字段
	function, ok := result["function"].(map[string]interface{})
	if !ok {
		t.Fatalf("期望 function 为 map, 实际类型为 %T", result["function"])
	}

	if function["name"] != "get_weather" {
		t.Errorf("期望 function.name 为 'get_weather', 实际为 %v", function["name"])
	}
}

// TestToolChoiceUnion_MarshalJSON_NamedCustom 测试 ToolChoiceUnion 的 NamedCustom 自定义模式序列化
func TestToolChoiceUnion_MarshalJSON_NamedCustom(t *testing.T) {
	toolChoice := ToolChoiceUnion{
		NamedCustom: &ToolChoiceNamedCustom{
			Type: "custom",
			Custom: struct {
				Name string `json:"name"`
			}{
				Name: "custom_tool",
			},
		},
	}

	data, err := json.Marshal(toolChoice)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化以验证结构
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("验证序列化结果失败: %v", err)
	}

	// 验证 type 字段
	if result["type"] != "custom" {
		t.Errorf("期望 type 为 'custom', 实际为 %v", result["type"])
	}

	// 验证 custom 字段
	custom, ok := result["custom"].(map[string]interface{})
	if !ok {
		t.Fatalf("期望 custom 为 map, 实际类型为 %T", result["custom"])
	}

	if custom["name"] != "custom_tool" {
		t.Errorf("期望 custom.name 为 'custom_tool', 实际为 %v", custom["name"])
	}
}

// TestToolChoiceUnion_MarshalJSON_Allowed 测试 ToolChoiceUnion 的 Allowed 允许模式序列化
func TestToolChoiceUnion_MarshalJSON_Allowed(t *testing.T) {
	toolChoice := ToolChoiceUnion{
		Allowed: &ToolChoiceAllowed{
			Type: "allowed",
			Mode: "any",
			Tools: []map[string]interface{}{
				{
					"type": "function",
					"name": "tool1",
				},
				{
					"type": "function",
					"name": "tool2",
				},
			},
		},
	}

	data, err := json.Marshal(toolChoice)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化以验证结构
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("验证序列化结果失败: %v", err)
	}

	// 验证 type 字段
	if result["type"] != "allowed" {
		t.Errorf("期望 type 为 'allowed', 实际为 %v", result["type"])
	}

	// 验证 mode 字段
	if result["mode"] != "any" {
		t.Errorf("期望 mode 为 'any', 实际为 %v", result["mode"])
	}

	// 验证 tools 字段
	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatalf("期望 tools 为 array, 实际类型为 %T", result["tools"])
	}

	if len(tools) != 2 {
		t.Errorf("期望 tools 长度为 2, 实际为 %d", len(tools))
	}
}

// TestToolChoiceUnion_MarshalJSON_Nil 测试 ToolChoiceUnion 的空值序列化
func TestToolChoiceUnion_MarshalJSON_Nil(t *testing.T) {
	toolChoice := ToolChoiceUnion{}

	data, err := json.Marshal(toolChoice)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	expected := "null"
	if string(data) != expected {
		t.Errorf("期望序列化结果为 %s, 实际为 %s", expected, string(data))
	}
}

// TestRequest_WithToolChoice 测试包含 tool_choice 字段的 Request 序列化
func TestRequest_WithToolChoice(t *testing.T) {
	tests := []struct {
		name       string
		toolChoice *ToolChoiceUnion
		checkFunc  func(*testing.T, map[string]interface{})
	}{
		{
			name: "auto 模式",
			toolChoice: &ToolChoiceUnion{
				Auto: stringPtr("auto"),
			},
			checkFunc: func(t *testing.T, result map[string]interface{}) {
				if result["tool_choice"] != "auto" {
					t.Errorf("期望 tool_choice 为 'auto', 实际为 %v", result["tool_choice"])
				}
			},
		},
		{
			name: "named 函数模式",
			toolChoice: &ToolChoiceUnion{
				Named: &ToolChoiceNamed{
					Type: "function",
					Function: struct {
						Name string `json:"name"`
					}{
						Name: "get_current_weather",
					},
				},
			},
			checkFunc: func(t *testing.T, result map[string]interface{}) {
				toolChoice, ok := result["tool_choice"].(map[string]interface{})
				if !ok {
					t.Fatalf("期望 tool_choice 为 map, 实际类型为 %T", result["tool_choice"])
				}

				if toolChoice["type"] != "function" {
					t.Errorf("期望 tool_choice.type 为 'function', 实际为 %v", toolChoice["type"])
				}

				function, ok := toolChoice["function"].(map[string]interface{})
				if !ok {
					t.Fatalf("期望 tool_choice.function 为 map, 实际类型为 %T", toolChoice["function"])
				}

				if function["name"] != "get_current_weather" {
					t.Errorf("期望 tool_choice.function.name 为 'get_current_weather', 实际为 %v", function["name"])
				}
			},
		},
		{
			name:       "nil 值",
			toolChoice: nil,
			checkFunc: func(t *testing.T, result map[string]interface{}) {
				if _, exists := result["tool_choice"]; exists {
					t.Error("期望 tool_choice 不存在，但实际存在")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := Request{
				Model: "gpt-4",
				Messages: []RequestMessage{
					{Role: "user", Content: "What's the weather?"},
				},
				ToolChoice: tt.toolChoice,
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

			// 执行检查函数
			tt.checkFunc(t, result)
		})
	}
}

// TestToolChoiceUnionUnmarshal 测试 ToolChoiceUnion 的 JSON 反序列化
func TestToolChoiceUnionUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "字符串值 - auto",
			json:    `{"tool_choice": "auto"}`,
			wantErr: false,
		},
		{
			name:    "字符串值 - none",
			json:    `{"tool_choice": "none"}`,
			wantErr: false,
		},
		{
			name:    "字符串值 - required",
			json:    `{"tool_choice": "required"}`,
			wantErr: false,
		},
		{
			name:    "函数类型",
			json:    `{"tool_choice": {"type": "function", "function": {"name": "测试函数"}}}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req struct {
				ToolChoice *ToolChoiceUnion `json:"tool_choice,omitempty"`
			}
			err := json.Unmarshal([]byte(tt.json), &req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && req.ToolChoice == nil {
				t.Error("ToolChoice 不应为 nil")
			}
		})
	}
}

// TestFullRequestUnmarshal 测试完整请求的反序列化（包含 tool_choice）
func TestFullRequestUnmarshal(t *testing.T) {
	jsonData := `{
  "model": "deepseek-v3",
  "messages": [
    {
      "role": "user",
      "content": "测试消息"
    }
  ],
  "stream": true,
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "apply_diff",
        "description": "Apply precise modifications",
        "parameters": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string"
            }
          },
          "required": ["path"]
        }
      }
    }
  ],
  "tool_choice": "auto"
}`

	var req Request
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if req.Model != "deepseek-v3" {
		t.Errorf("Model = %v, want deepseek-v3", req.Model)
	}

	if req.ToolChoice == nil {
		t.Fatal("ToolChoice 不应为 nil")
	}

	if req.ToolChoice.Auto == nil {
		t.Error("ToolChoice.Auto 不应为 nil")
	} else if *req.ToolChoice.Auto != "auto" {
		t.Errorf("ToolChoice.Auto = %v, want auto", *req.ToolChoice.Auto)
	}
}

// TestStopUnionUnmarshal 测试 StopUnion 的反序列化
func TestStopUnionUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "字符串值",
			json:    `{"stop": "END"}`,
			wantErr: false,
		},
		{
			name:    "数组值",
			json:    `{"stop": ["END", "STOP"]}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req struct {
				Stop *StopUnion `json:"stop,omitempty"`
			}
			err := json.Unmarshal([]byte(tt.json), &req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestToolUnionUnmarshal 测试 ToolUnion 的反序列化
func TestToolUnionUnmarshal(t *testing.T) {
	jsonData := `{
  "type": "function",
  "function": {
    "name": "test_function",
    "description": "A test function",
    "parameters": {
      "type": "object",
      "properties": {
        "param1": {"type": "string"}
      }
    }
  }
}`

	var tool ToolUnion
	err := json.Unmarshal([]byte(jsonData), &tool)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if tool.Function == nil {
		t.Fatal("Function 不应为 nil")
	}

	if tool.Function.Function.Name != "test_function" {
		t.Errorf("Function.Name = %v, 应该为 test_function", tool.Function.Function.Name)
	}
}
