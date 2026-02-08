package responses

import (
	"encoding/json"
	"testing"

	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

// TestRequestExtraFields 测试 Request 的 ExtraFields 透传功能
func TestRequestExtraFields(t *testing.T) {
	tests := []struct {
		name         string
		jsonInput    string
		wantExtra    map[string]json.RawMessage
		wantPreserve map[string]interface{}
	}{
		{
			name: "未知字段被捕获到 ExtraFields",
			jsonInput: `{
				"model": "gpt-4o",
				"unknown_field": "value",
				"nested_unknown": {"key": "value"}
			}`,
			wantExtra: map[string]json.RawMessage{
				"unknown_field":  json.RawMessage(`"value"`),
				"nested_unknown": json.RawMessage(`{"key":"value"}`),
			},
		},
		{
			name: "ExtraFields 不覆盖已知字段",
			jsonInput: `{
				"model": "gpt-4o",
				"max_output_tokens": 1000,
				"temperature": 0.7,
				"custom_field": "custom_value"
			}`,
			wantExtra: map[string]json.RawMessage{
				"custom_field": json.RawMessage(`"custom_value"`),
			},
			wantPreserve: map[string]interface{}{
				"model":             "gpt-4o",
				"max_output_tokens": float64(1000),
				"temperature":       0.7,
			},
		},
		{
			name: "ExtraFields round-trip 序列化",
			jsonInput: `{
				"model": "gpt-4o",
				"extra1": "value1",
				"extra2": 123,
				"extra3": true
			}`,
			wantExtra: map[string]json.RawMessage{
				"extra1": json.RawMessage(`"value1"`),
				"extra2": json.RawMessage(`123`),
				"extra3": json.RawMessage(`true`),
			},
		},
		{
			name: "空 ExtraFields 不影响序列化",
			jsonInput: `{
				"model": "gpt-4o",
				"max_output_tokens": 1000
			}`,
			wantExtra: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req Request
			if err := json.Unmarshal([]byte(tt.jsonInput), &req); err != nil {
				t.Fatalf("Unmarshal 失败: %v", err)
			}

			// 验证 ExtraFields
			if len(req.ExtraFields) != len(tt.wantExtra) {
				t.Errorf("ExtraFields 长度不匹配：got %d, want %d", len(req.ExtraFields), len(tt.wantExtra))
			}
			for k, v := range tt.wantExtra {
				if got, ok := req.ExtraFields[k]; !ok {
					t.Errorf("ExtraFields 缺少键 %q", k)
				} else if !compareRawMessage(got, v) {
					t.Errorf("ExtraFields[%q] = %s, want %s", k, got, v)
				}
			}

			// 验证已知字段未被覆盖
			if tt.wantPreserve != nil {
				if req.Model == nil || *req.Model != tt.wantPreserve["model"].(string) {
					t.Errorf("Model 字段被覆盖: got %v, want %v", req.Model, tt.wantPreserve["model"])
				}
				if req.MaxOutputTokens == nil || *req.MaxOutputTokens != int(tt.wantPreserve["max_output_tokens"].(float64)) {
					t.Errorf("MaxOutputTokens 字段被覆盖: got %v, want %v", req.MaxOutputTokens, tt.wantPreserve["max_output_tokens"])
				}
			}

			// Round-trip 测试
			data, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("Marshal 失败: %v", err)
			}

			var req2 Request
			if err := json.Unmarshal(data, &req2); err != nil {
				t.Fatalf("Round-trip Unmarshal 失败: %v", err)
			}

			// 验证 ExtraFields 在 round-trip 后保持一致
			if len(req2.ExtraFields) != len(req.ExtraFields) {
				t.Errorf("Round-trip 后 ExtraFields 长度不匹配：got %d, want %d", len(req2.ExtraFields), len(req.ExtraFields))
			}
			for k, v := range req.ExtraFields {
				if got, ok := req2.ExtraFields[k]; !ok {
					t.Errorf("Round-trip 后 ExtraFields 缺少键 %q", k)
				} else if !compareRawMessage(got, v) {
					t.Errorf("Round-trip 后 ExtraFields[%q] = %s, want %s", k, got, v)
				}
			}
		})
	}
}

// compareRawMessage 比较两个 json.RawMessage 是否相等
func compareRawMessage(a, b json.RawMessage) bool {
	return string(a) == string(b)
}

// TestInputUnion 测试 InputUnion 的双态（字符串/数组）
func TestInputUnion(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantType  string // "string" or "items"
		wantValue interface{}
	}{
		{
			name:      "字符串输入",
			jsonInput: `"Hello, world!"`,
			wantType:  "string",
			wantValue: "Hello, world!",
		},
		{
			name: "数组输入 - 单个消息",
			jsonInput: `[{
				"type": "message",
				"role": "user",
				"content": [{"type": "input_text", "text": "Hello"}]
			}]`,
			wantType: "items",
		},
		{
			name: "数组输入 - 多个消息",
			jsonInput: `[
				{
					"type": "message",
					"role": "system",
					"content": [{"type": "input_text", "text": "You are helpful"}]
				},
				{
					"type": "message",
					"role": "user",
					"content": [{"type": "input_text", "text": "Hello"}]
				}
			]`,
			wantType: "items",
		},
		{
			name:      "空值输入",
			jsonInput: `null`,
			wantType:  "none",
		},
		{
			name:      "空数组输入",
			jsonInput: `[]`,
			wantType:  "items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input InputUnion
			if err := json.Unmarshal([]byte(tt.jsonInput), &input); err != nil {
				t.Fatalf("Unmarshal 失败: %v", err)
			}

			// 验证类型分支
			switch tt.wantType {
			case "string":
				if input.StringValue == nil {
					t.Error("期望 StringValue 不为 nil")
				} else if *input.StringValue != tt.wantValue.(string) {
					t.Errorf("StringValue = %q, want %q", *input.StringValue, tt.wantValue.(string))
				}
				if input.Items != nil {
					t.Error("期望 Items 为 nil")
				}
			case "items":
				if input.Items == nil {
					t.Error("期望 Items 不为 nil")
				}
				if input.StringValue != nil {
					t.Error("期望 StringValue 为 nil")
				}
			case "none":
				if input.StringValue != nil || input.Items != nil {
					t.Error("期望所有字段都为 nil")
				}
			}

			// Round-trip 测试
			data, err := json.Marshal(input)
			if err != nil {
				t.Fatalf("Marshal 失败: %v", err)
			}

			var input2 InputUnion
			if err := json.Unmarshal(data, &input2); err != nil {
				t.Fatalf("Round-trip Unmarshal 失败: %v", err)
			}

			// 验证 round-trip 后类型一致
			if (input.StringValue == nil) != (input2.StringValue == nil) {
				t.Error("Round-trip 后 StringValue 状态不一致")
			}
			if (input.Items == nil) != (input2.Items == nil) {
				t.Error("Round-trip 后 Items 状态不一致")
			}
		})
	}
}

// TestConversationUnion 测试 ConversationUnion 的双态（字符串/对象）
func TestConversationUnion(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantType  string // "string" or "object"
		wantValue interface{}
	}{
		{
			name:      "字符串会话 ID",
			jsonInput: `"conv_12345"`,
			wantType:  "string",
			wantValue: "conv_12345",
		},
		{
			name:      "对象会话引用",
			jsonInput: `{"id": "conv_67890"}`,
			wantType:  "object",
			wantValue: ConversationReference{ID: "conv_67890"},
		},
		{
			name:      "空值输入",
			jsonInput: `null`,
			wantType:  "none",
		},
		{
			name:      "空对象（无效，应被忽略）",
			jsonInput: `{}`,
			wantType:  "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var conv ConversationUnion
			if err := json.Unmarshal([]byte(tt.jsonInput), &conv); err != nil {
				t.Fatalf("Unmarshal 失败: %v", err)
			}

			// 验证类型分支
			switch tt.wantType {
			case "string":
				if conv.StringValue == nil {
					t.Error("期望 StringValue 不为 nil")
				} else if *conv.StringValue != tt.wantValue.(string) {
					t.Errorf("StringValue = %q, want %q", *conv.StringValue, tt.wantValue.(string))
				}
				if conv.ObjectValue != nil {
					t.Error("期望 ObjectValue 为 nil")
				}
			case "object":
				if conv.ObjectValue == nil {
					t.Error("期望 ObjectValue 不为 nil")
				} else if conv.ObjectValue.ID != tt.wantValue.(ConversationReference).ID {
					t.Errorf("ObjectValue.ID = %q, want %q", conv.ObjectValue.ID, tt.wantValue.(ConversationReference).ID)
				}
				if conv.StringValue != nil {
					t.Error("期望 StringValue 为 nil")
				}
			case "none":
				if conv.StringValue != nil || conv.ObjectValue != nil {
					t.Error("期望所有字段都为 nil")
				}
			}

			// Round-trip 测试
			data, err := json.Marshal(conv)
			if err != nil {
				t.Fatalf("Marshal 失败: %v", err)
			}

			var conv2 ConversationUnion
			if err := json.Unmarshal(data, &conv2); err != nil {
				t.Fatalf("Round-trip Unmarshal 失败: %v", err)
			}

			// 验证 round-trip 后类型一致
			if (conv.StringValue == nil) != (conv2.StringValue == nil) {
				t.Error("Round-trip 后 StringValue 状态不一致")
			}
			if (conv.ObjectValue == nil) != (conv2.ObjectValue == nil) {
				t.Error("Round-trip 后 ObjectValue 状态不一致")
			}
		})
	}
}

// TestReasoning 测试 Reasoning 的字段映射
func TestReasoning(t *testing.T) {
	tests := []struct {
		name                    string
		jsonInput               string
		wantEffort              *shared.ReasoningEffort
		wantSummary             *ReasoningSummary
		wantGenerateSummary     *ReasoningSummary
		wantMarshalSummaryField string // "summary", "generate_summary", 或 "null"
	}{
		{
			name: "使用 summary 字段",
			jsonInput: `{
				"effort": "medium",
				"summary": "detailed"
			}`,
			wantEffort:              ptrReasoningEffort(shared.ReasoningEffortMedium),
			wantSummary:             ptrReasoningSummary(ReasoningSummaryDetailed),
			wantGenerateSummary:     nil,
			wantMarshalSummaryField: "summary",
		},
		{
			name: "使用 generate_summary 字段",
			jsonInput: `{
				"effort": "high",
				"generate_summary": "concise"
			}`,
			wantEffort:              ptrReasoningEffort(shared.ReasoningEffortHigh),
			wantSummary:             nil,
			wantGenerateSummary:     ptrReasoningSummary(ReasoningSummaryConcise),
			wantMarshalSummaryField: "generate_summary",
		},
		{
			name: "同时设置 summary 和 generate_summary 应报错",
			jsonInput: `{
				"effort": "low",
				"summary": "auto",
				"generate_summary": "detailed"
			}`,
			wantEffort:              nil,
			wantSummary:             nil,
			wantGenerateSummary:     nil,
			wantMarshalSummaryField: "",
		},
		{
			name: "只有 effort 字段",
			jsonInput: `{
				"effort": "medium"
			}`,
			wantEffort:              ptrReasoningEffort(shared.ReasoningEffortMedium),
			wantSummary:             nil,
			wantGenerateSummary:     nil,
			wantMarshalSummaryField: "null",
		},
		{
			name:                    "空对象",
			jsonInput:               `{}`,
			wantEffort:              nil,
			wantSummary:             nil,
			wantGenerateSummary:     nil,
			wantMarshalSummaryField: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r Reasoning
			err := json.Unmarshal([]byte(tt.jsonInput), &r)

			// 对于同时设置 summary 和 generate_summary 的情况，期望反序列化失败
			if tt.name == "同时设置 summary 和 generate_summary 应报错" {
				if err == nil {
					t.Error("期望反序列化失败，但没有返回错误")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unmarshal 失败: %v", err)
			}

			// 验证 effort 字段
			if (r.Effort == nil) != (tt.wantEffort == nil) {
				t.Errorf("Effort 状态不匹配: got %v, want %v", r.Effort, tt.wantEffort)
			} else if r.Effort != nil && *r.Effort != *tt.wantEffort {
				t.Errorf("Effort = %v, want %v", *r.Effort, *tt.wantEffort)
			}

			// 验证 Summary 字段
			if (r.Summary == nil) != (tt.wantSummary == nil) {
				t.Errorf("Summary 状态不匹配: got %v, want %v", r.Summary, tt.wantSummary)
			} else if r.Summary != nil && *r.Summary != *tt.wantSummary {
				t.Errorf("Summary = %v, want %v", *r.Summary, *tt.wantSummary)
			}

			// 验证 GenerateSummary 字段
			if (r.GenerateSummary == nil) != (tt.wantGenerateSummary == nil) {
				t.Errorf("GenerateSummary 状态不匹配: got %v, want %v", r.GenerateSummary, tt.wantGenerateSummary)
			} else if r.GenerateSummary != nil && *r.GenerateSummary != *tt.wantGenerateSummary {
				t.Errorf("GenerateSummary = %v, want %v", *r.GenerateSummary, *tt.wantGenerateSummary)
			}

			// 序列化测试：验证输出字段
			data, err := json.Marshal(r)
			if err != nil {
				t.Fatalf("Marshal 失败: %v", err)
			}

			var raw map[string]interface{}
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("解析序列化结果失败: %v", err)
			}

			// 验证序列化输出的字段
			switch tt.wantMarshalSummaryField {
			case "summary":
				if _, ok := raw["summary"]; !ok {
					t.Error("序列化结果中应包含 summary 字段")
				}
				if _, ok := raw["generate_summary"]; ok {
					t.Error("序列化结果中不应包含 generate_summary 字段")
				}
			case "generate_summary":
				if _, ok := raw["generate_summary"]; !ok {
					t.Error("序列化结果中应包含 generate_summary 字段")
				}
				if _, ok := raw["summary"]; ok {
					t.Error("序列化结果中不应包含 summary 字段")
				}
			case "null":
				if val, ok := raw["summary"]; !ok {
					t.Error("序列化结果中应包含 summary 字段（值为 null）")
				} else if val != nil {
					t.Errorf("序列化结果中 summary 应为 null，got %v", val)
				}
				if _, ok := raw["generate_summary"]; ok {
					t.Error("序列化结果中不应包含 generate_summary 字段")
				}
			}

			// Round-trip 测试
			var r2 Reasoning
			if err := json.Unmarshal(data, &r2); err != nil {
				t.Fatalf("Round-trip Unmarshal 失败: %v", err)
			}

			// 验证 round-trip 后字段一致
			if (r.Effort == nil) != (r2.Effort == nil) {
				t.Error("Round-trip 后 Effort 状态不一致")
			} else if r.Effort != nil && *r.Effort != *r2.Effort {
				t.Errorf("Round-trip 后 Effort 不一致: got %v, want %v", *r2.Effort, *r.Effort)
			}

			if (r.Summary == nil) != (r2.Summary == nil) {
				t.Error("Round-trip 后 Summary 状态不一致")
			} else if r.Summary != nil && *r.Summary != *r2.Summary {
				t.Errorf("Round-trip 后 Summary 不一致: got %v, want %v", *r2.Summary, *r.Summary)
			}

			if (r.GenerateSummary == nil) != (r2.GenerateSummary == nil) {
				t.Error("Round-trip 后 GenerateSummary 状态不一致")
			} else if r.GenerateSummary != nil && *r.GenerateSummary != *r2.GenerateSummary {
				t.Errorf("Round-trip 后 GenerateSummary 不一致: got %v, want %v", *r2.GenerateSummary, *r.GenerateSummary)
			}
		})
	}
}

// TestReasoningMarshalError 测试 Reasoning 的 MarshalJSON 错误情况
func TestReasoningMarshalError(t *testing.T) {
	tests := []struct {
		name      string
		reasoning Reasoning
	}{
		{
			name: "同时设置 Summary 和 GenerateSummary 应报错",
			reasoning: Reasoning{
				Effort:          ptrReasoningEffort(shared.ReasoningEffortMedium),
				Summary:         ptrReasoningSummary(ReasoningSummaryDetailed),
				GenerateSummary: ptrReasoningSummary(ReasoningSummaryConcise),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := json.Marshal(tt.reasoning)
			if err == nil {
				t.Error("期望序列化失败，但没有返回错误")
			}
		})
	}
}

// TestPromptVariableMap 测试 PromptVariableMap 的联合类型
func TestPromptVariableMap(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantCount int
	}{
		{
			name: "字符串变量",
			jsonInput: `{
				"var1": "string value",
				"var2": "another string"
			}`,
			wantCount: 2,
		},
		{
			name: "混合类型变量",
			jsonInput: `{
				"text_var": "simple text",
				"input_text_var": {"type": "input_text", "text": "structured text"},
				"image_var": {"type": "input_image", "image_url": "https://example.com/image.png"},
				"file_var": {"type": "input_file", "file_id": "file_123"}
			}`,
			wantCount: 4,
		},
		{
			name:      "空对象",
			jsonInput: `{}`,
			wantCount: 0,
		},
		{
			name:      "null 值",
			jsonInput: `null`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pm PromptVariableMap
			if err := json.Unmarshal([]byte(tt.jsonInput), &pm); err != nil {
				t.Fatalf("Unmarshal 失败: %v", err)
			}

			if len(pm) != tt.wantCount {
				t.Errorf("PromptVariableMap 长度 = %d, want %d", len(pm), tt.wantCount)
			}

			// 验证变量类型
			for key, value := range pm {
				switch {
				case value.StringValue != nil:
					t.Logf("变量 %q 是字符串类型: %q", key, *value.StringValue)
				case value.Text != nil:
					t.Logf("变量 %q 是 InputTextContent 类型", key)
				case value.Image != nil:
					t.Logf("变量 %q 是 InputImageContent 类型", key)
				case value.File != nil:
					t.Logf("变量 %q 是 InputFileContent 类型", key)
				default:
					t.Errorf("变量 %q 没有设置任何类型", key)
				}
			}

			// Round-trip 测试
			data, err := json.Marshal(pm)
			if err != nil {
				t.Fatalf("Marshal 失败: %v", err)
			}

			var pm2 PromptVariableMap
			if err := json.Unmarshal(data, &pm2); err != nil {
				t.Fatalf("Round-trip Unmarshal 失败: %v", err)
			}

			// 验证 round-trip 后长度一致
			if len(pm2) != len(pm) {
				t.Errorf("Round-trip 后 PromptVariableMap 长度不一致：got %d, want %d", len(pm2), len(pm))
			}

			// 验证 round-trip 后每个变量的类型一致
			for key, value := range pm {
				value2, ok := pm2[key]
				if !ok {
					t.Errorf("Round-trip 后缺少变量 %q", key)
					continue
				}

				if (value.StringValue == nil) != (value2.StringValue == nil) {
					t.Errorf("Round-trip 后变量 %q 的 StringValue 状态不一致", key)
				}
				if (value.Text == nil) != (value2.Text == nil) {
					t.Errorf("Round-trip 后变量 %q 的 Text 状态不一致", key)
				}
				if (value.Image == nil) != (value2.Image == nil) {
					t.Errorf("Round-trip 后变量 %q 的 Image 状态不一致", key)
				}
				if (value.File == nil) != (value2.File == nil) {
					t.Errorf("Round-trip 后变量 %q 的 File 状态不一致", key)
				}
			}
		})
	}
}

// TestPromptVariableValue 测试 PromptVariableValue 的联合类型
func TestPromptVariableValue(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantType  string
	}{
		{
			name:      "字符串值",
			jsonInput: `"simple string"`,
			wantType:  "string",
		},
		{
			name:      "InputTextContent 值",
			jsonInput: `{"type": "input_text", "text": "structured text"}`,
			wantType:  "text",
		},
		{
			name:      "InputImageContent 值",
			jsonInput: `{"type": "input_image", "image_url": "https://example.com/image.png"}`,
			wantType:  "image",
		},
		{
			name:      "InputFileContent 值",
			jsonInput: `{"type": "input_file", "file_id": "file_123"}`,
			wantType:  "file",
		},
		{
			name:      "null 值",
			jsonInput: `null`,
			wantType:  "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pv PromptVariableValue
			if err := json.Unmarshal([]byte(tt.jsonInput), &pv); err != nil {
				t.Fatalf("Unmarshal 失败: %v", err)
			}

			// 验证类型分支
			switch tt.wantType {
			case "string":
				if pv.StringValue == nil {
					t.Error("期望 StringValue 不为 nil")
				}
				if pv.Text != nil || pv.Image != nil || pv.File != nil {
					t.Error("期望其他字段为 nil")
				}
			case "text":
				if pv.Text == nil {
					t.Error("期望 Text 不为 nil")
				}
				if pv.StringValue != nil || pv.Image != nil || pv.File != nil {
					t.Error("期望其他字段为 nil")
				}
			case "image":
				if pv.Image == nil {
					t.Error("期望 Image 不为 nil")
				}
				if pv.StringValue != nil || pv.Text != nil || pv.File != nil {
					t.Error("期望其他字段为 nil")
				}
			case "file":
				if pv.File == nil {
					t.Error("期望 File 不为 nil")
				}
				if pv.StringValue != nil || pv.Text != nil || pv.Image != nil {
					t.Error("期望其他字段为 nil")
				}
			case "none":
				if pv.StringValue != nil || pv.Text != nil || pv.Image != nil || pv.File != nil {
					t.Error("期望所有字段都为 nil")
				}
			}

			// Round-trip 测试
			data, err := json.Marshal(pv)
			if err != nil {
				t.Fatalf("Marshal 失败: %v", err)
			}

			var pv2 PromptVariableValue
			if err := json.Unmarshal(data, &pv2); err != nil {
				t.Fatalf("Round-trip Unmarshal 失败: %v", err)
			}

			// 验证 round-trip 后类型一致
			if (pv.StringValue == nil) != (pv2.StringValue == nil) {
				t.Error("Round-trip 后 StringValue 状态不一致")
			}
			if (pv.Text == nil) != (pv2.Text == nil) {
				t.Error("Round-trip 后 Text 状态不一致")
			}
			if (pv.Image == nil) != (pv2.Image == nil) {
				t.Error("Round-trip 后 Image 状态不一致")
			}
			if (pv.File == nil) != (pv2.File == nil) {
				t.Error("Round-trip 后 File 状态不一致")
			}
		})
	}
}

// TestTextFormatUnion 测试 TextFormatUnion 的多分支
func TestTextFormatUnion(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantType  string
	}{
		{
			name:      "text 类型",
			jsonInput: `{"type": "text"}`,
			wantType:  "text",
		},
		{
			name: "json_schema 类型",
			jsonInput: `{
				"type": "json_schema",
				"name": "person",
				"description": "A person",
				"schema": {"type": "object"},
				"strict": true
			}`,
			wantType: "json_schema",
		},
		{
			name:      "json_object 类型",
			jsonInput: `{"type": "json_object"}`,
			wantType:  "json_object",
		},
		{
			name:      "空 type（默认为 text）",
			jsonInput: `{}`,
			wantType:  "text",
		},
		{
			name:      "null 值",
			jsonInput: `null`,
			wantType:  "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tf TextFormatUnion
			if err := json.Unmarshal([]byte(tt.jsonInput), &tf); err != nil {
				t.Fatalf("Unmarshal 失败: %v", err)
			}

			// 验证类型分支
			switch tt.wantType {
			case "text":
				if tf.Text == nil {
					t.Error("期望 Text 不为 nil")
				}
				if tf.JSONSchema != nil || tf.JSONObject != nil {
					t.Error("期望其他字段为 nil")
				}
			case "json_schema":
				if tf.JSONSchema == nil {
					t.Error("期望 JSONSchema 不为 nil")
				}
				if tf.Text != nil || tf.JSONObject != nil {
					t.Error("期望其他字段为 nil")
				}
			case "json_object":
				if tf.JSONObject == nil {
					t.Error("期望 JSONObject 不为 nil")
				}
				if tf.Text != nil || tf.JSONSchema != nil {
					t.Error("期望其他字段为 nil")
				}
			case "none":
				if tf.Text != nil || tf.JSONSchema != nil || tf.JSONObject != nil {
					t.Error("期望所有字段都为 nil")
				}
			}

			// Round-trip 测试
			data, err := json.Marshal(tf)
			if err != nil {
				t.Fatalf("Marshal 失败: %v", err)
			}

			var tf2 TextFormatUnion
			if err := json.Unmarshal(data, &tf2); err != nil {
				t.Fatalf("Round-trip Unmarshal 失败: %v", err)
			}

			// 验证 round-trip 后类型一致
			if (tf.Text == nil) != (tf2.Text == nil) {
				t.Error("Round-trip 后 Text 状态不一致")
			}
			if (tf.JSONSchema == nil) != (tf2.JSONSchema == nil) {
				t.Error("Round-trip 后 JSONSchema 状态不一致")
			}
			if (tf.JSONObject == nil) != (tf2.JSONObject == nil) {
				t.Error("Round-trip 后 JSONObject 状态不一致")
			}
		})
	}
}

// TestStreamOptions 测试 StreamOptions 的简单结构
func TestStreamOptions(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantValue *bool
	}{
		{
			name:      "include_obfuscation 为 true",
			jsonInput: `{"include_obfuscation": true}`,
			wantValue: ptrBool(true),
		},
		{
			name:      "include_obfuscation 为 false",
			jsonInput: `{"include_obfuscation": false}`,
			wantValue: ptrBool(false),
		},
		{
			name:      "空对象",
			jsonInput: `{}`,
			wantValue: nil,
		},
		{
			name:      "null 值",
			jsonInput: `null`,
			wantValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var so StreamOptions
			if err := json.Unmarshal([]byte(tt.jsonInput), &so); err != nil {
				t.Fatalf("Unmarshal 失败: %v", err)
			}

			// 验证字段值
			if (so.IncludeObfuscation == nil) != (tt.wantValue == nil) {
				t.Errorf("IncludeObfuscation 状态不匹配: got %v, want %v", so.IncludeObfuscation, tt.wantValue)
			} else if so.IncludeObfuscation != nil && *so.IncludeObfuscation != *tt.wantValue {
				t.Errorf("IncludeObfuscation = %v, want %v", *so.IncludeObfuscation, *tt.wantValue)
			}

			// Round-trip 测试
			data, err := json.Marshal(so)
			if err != nil {
				t.Fatalf("Marshal 失败: %v", err)
			}

			var so2 StreamOptions
			if err := json.Unmarshal(data, &so2); err != nil {
				t.Fatalf("Round-trip Unmarshal 失败: %v", err)
			}

			// 验证 round-trip 后字段一致
			if (so.IncludeObfuscation == nil) != (so2.IncludeObfuscation == nil) {
				t.Error("Round-trip 后 IncludeObfuscation 状态不一致")
			} else if so.IncludeObfuscation != nil && *so.IncludeObfuscation != *so2.IncludeObfuscation {
				t.Errorf("Round-trip 后 IncludeObfuscation 不一致: got %v, want %v", *so2.IncludeObfuscation, *so.IncludeObfuscation)
			}
		})
	}
}

// TestRequestFullRoundTrip 测试完整的 Request 结构 round-trip
func TestRequestFullRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
	}{
		{
			name: "完整请求结构",
			jsonInput: `{
				"model": "gpt-4o",
				"input": "Hello, world!",
				"stream": true,
				"stream_options": {"include_obfuscation": false},
				"max_output_tokens": 1000,
				"temperature": 0.7,
				"top_p": 0.9,
				"top_logprobs": 5,
				"tools": [{"type": "function", "function": {"name": "test", "description": "test function"}}],
				"tool_choice": {"type": "auto"},
				"parallel_tool_calls": true,
				"max_tool_calls": 10,
				"truncation": "auto",
				"text": {
					"format": {"type": "text"},
					"verbosity": "low"
				},
				"store": true,
				"include": ["file_search_call.results"],
				"metadata": {"key": "value"},
				"instructions": "You are helpful",
				"reasoning": {
					"effort": "medium",
					"summary": "detailed"
				},
				"prompt": {
					"id": "prompt_123",
					"version": "v1",
					"variables": {
						"var1": "value1",
						"var2": {"type": "input_text", "text": "structured"}
					}
				},
				"conversation": "conv_456",
				"previous_response_id": "resp_789",
				"safety_identifier": "safe_123",
				"user": "user_123",
				"prompt_cache_key": "cache_key",
				"prompt_cache_retention": "in-memory",
				"service_tier": "auto",
				"background": false,
				"custom_field": "custom_value"
			}`,
		},
		{
			name: "最小请求结构",
			jsonInput: `{
				"model": "gpt-4o",
				"input": "Hello"
			}`,
		},
		{
			name: "使用数组输入的请求",
			jsonInput: `{
				"model": "gpt-4o",
				"input": [{
					"type": "message",
					"role": "user",
					"content": [{"type": "input_text", "text": "Hello"}]
				}]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req Request
			if err := json.Unmarshal([]byte(tt.jsonInput), &req); err != nil {
				t.Fatalf("Unmarshal 失败: %v", err)
			}

			// 序列化
			data, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("Marshal 失败: %v", err)
			}

			// Round-trip 反序列化
			var req2 Request
			if err := json.Unmarshal(data, &req2); err != nil {
				t.Fatalf("Round-trip Unmarshal 失败: %v", err)
			}

			// 验证关键字段
			if (req.Model == nil) != (req2.Model == nil) {
				t.Error("Round-trip 后 Model 状态不一致")
			} else if req.Model != nil && *req.Model != *req2.Model {
				t.Errorf("Round-trip 后 Model 不一致: got %q, want %q", *req2.Model, *req.Model)
			}

			if (req.Input == nil) != (req2.Input == nil) {
				t.Error("Round-trip 后 Input 状态不一致")
			}

			if (req.Reasoning == nil) != (req2.Reasoning == nil) {
				t.Error("Round-trip 后 Reasoning 状态不一致")
			}

			if (req.StreamOptions == nil) != (req2.StreamOptions == nil) {
				t.Error("Round-trip 后 StreamOptions 状态不一致")
			}

			// 验证 ExtraFields
			if len(req.ExtraFields) != len(req2.ExtraFields) {
				t.Errorf("Round-trip 后 ExtraFields 长度不一致：got %d, want %d", len(req2.ExtraFields), len(req.ExtraFields))
			}
			for k, v := range req.ExtraFields {
				if got, ok := req2.ExtraFields[k]; !ok {
					t.Errorf("Round-trip 后 ExtraFields 缺少键 %q", k)
				} else if !compareRawMessage(got, v) {
					t.Errorf("Round-trip 后 ExtraFields[%q] 不一致: got %s, want %s", k, got, v)
				}
			}
		})
	}
}

// 辅助函数

// ptrBool 返回 bool 指针
func ptrBool(b bool) *bool {
	return &b
}

// ptrReasoningEffort 返回 ReasoningEffort 指针
func ptrReasoningEffort(e shared.ReasoningEffort) *shared.ReasoningEffort {
	return &e
}

// ptrReasoningSummary 返回 ReasoningSummary 指针
func ptrReasoningSummary(s ReasoningSummary) *ReasoningSummary {
	return &s
}
