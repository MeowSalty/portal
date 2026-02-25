package responses

import (
	"encoding/json"
	"testing"

	portalErrors "github.com/MeowSalty/portal/errors"
)

// TestInputItemMarshalJSON 测试 InputItem 的 JSON 序列化
func TestInputItemMarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     InputItem
		wantError bool
		validator func(t *testing.T, data []byte)
	}{
		{
			name: "MarshalJSON_OneofMessage",
			input: InputItem{
				Message: &InputMessage{
					Type: InputItemTypeMessage,
					Role: ResponseMessageRoleUser,
					Content: InputMessageContent{
						String: ptrString("hi"),
					},
				},
			},
			wantError: false,
			validator: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("解析 JSON 失败: %v", err)
				}
				if result["type"] != "message" {
					t.Errorf("期望 type=message，got %v", result["type"])
				}
				if result["role"] != "user" {
					t.Errorf("期望 role=user，got %v", result["role"])
				}
			},
		},
		{
			name: "MarshalJSON_OneofOutputMessage",
			input: InputItem{
				OutputMessage: &OutputMessage{
					Type:   "message",
					ID:     "msg_123",
					Role:   "assistant",
					Status: "completed",
					Content: []OutputMessageContent{
						{
							OutputText: &OutputTextContent{
								Type: "output_text",
								Text: "ok",
							},
						},
					},
				},
			},
			wantError: false,
			validator: func(t *testing.T, data []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("解析 JSON 失败: %v", err)
				}
				if result["type"] != "message" {
					t.Errorf("期望 type=message，got %v", result["type"])
				}
				if result["status"] != "completed" {
					t.Errorf("期望 status 字段存在且为 completed，got %v", result["status"])
				}
			},
		},
		{
			name:      "MarshalJSON_Null",
			input:     InputItem{},
			wantError: false,
			validator: func(t *testing.T, data []byte) {
				if string(data) != "null" {
					t.Errorf("期望 null，got %s", string(data))
				}
			},
		},
		{
			name: "MarshalJSON_MultipleTypesError",
			input: InputItem{
				Message: &InputMessage{
					Type: InputItemTypeMessage,
					Role: ResponseMessageRoleUser,
					Content: InputMessageContent{
						String: ptrString("hi"),
					},
				},
				ItemReference: &ItemReferenceParam{
					ID: "item_123",
				},
			},
			wantError: true,
			validator: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("期望序列化失败，但没有返回错误")
				} else {
					// 验证错误码是否为 ErrCodeInvalidArgument
					var portalErr *portalErrors.Error
					if portalErrors.As(err, &portalErr) {
						if portalErr.Code != portalErrors.ErrCodeInvalidArgument {
							t.Errorf("错误码不匹配: got %s, want %s", portalErr.Code, portalErrors.ErrCodeInvalidArgument)
						}
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("序列化失败: %v", err)
			}

			if tt.validator != nil {
				tt.validator(t, data)
			}
		})
	}
}

// TestInputItemUnmarshalJSON 测试 InputItem 的 JSON 反序列化
func TestInputItemUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantError bool
		validator func(t *testing.T, item InputItem)
	}{
		{
			name:      "UnmarshalJSON_MessageAsInput",
			jsonInput: `{"type":"message","role":"user","content":"hi"}`,
			wantError: false,
			validator: func(t *testing.T, item InputItem) {
				if item.Message == nil {
					t.Error("期望 Message 非空，但为 nil")
				}
				if item.OutputMessage != nil {
					t.Error("期望 OutputMessage 为空，但非 nil")
				}
				if item.Message != nil {
					if item.Message.Role != ResponseMessageRoleUser {
						t.Errorf("期望 Role=user，got %s", item.Message.Role)
					}
					if item.Message.Content.String == nil {
						t.Error("期望 Content.String 非空，但为 nil")
					} else if *item.Message.Content.String != "hi" {
						t.Errorf("期望 Content.String=hi，got %s", *item.Message.Content.String)
					}
				}
			},
		},
		{
			name:      "UnmarshalJSON_MessageAsOutput",
			jsonInput: `{"type":"message","status":"completed","role":"assistant","content":[{"type":"output_text","text":"ok","annotations":[]}]}`,
			wantError: false,
			validator: func(t *testing.T, item InputItem) {
				if item.Message != nil {
					t.Error("期望 Message 为空，但非 nil")
				}
				if item.OutputMessage == nil {
					t.Error("期望 OutputMessage 非空，但为 nil")
				}
				if item.OutputMessage != nil {
					if item.OutputMessage.Status != "completed" {
						t.Errorf("期望 Status=completed，got %s", item.OutputMessage.Status)
					}
					if item.OutputMessage.Role != "assistant" {
						t.Errorf("期望 Role=assistant，got %s", item.OutputMessage.Role)
					}
				}
			},
		},
		{
			name:      "UnmarshalJSON_NormalizeOutputMessage",
			jsonInput: `{"role":"assistant","content":[{"type":"output_text","text":"ok","annotations":[]}]}`,
			wantError: false,
			validator: func(t *testing.T, item InputItem) {
				if item.Message != nil {
					t.Error("期望 Message 为空，但非 nil")
				}
				if item.OutputMessage == nil {
					t.Error("期望 OutputMessage 非空，但为 nil")
				}
				if item.OutputMessage != nil {
					if item.OutputMessage.Status != "completed" {
						t.Errorf("期望 Status=completed，got %s", item.OutputMessage.Status)
					}
					if item.OutputMessage.Type != OutputItemTypeMessage {
						t.Errorf("期望 Type=message，got %s", item.OutputMessage.Type)
					}
					if item.OutputMessage.ID == "" {
						t.Error("期望 ID 被补齐，但为空")
					}
					if item.OutputMessage.Role != "assistant" {
						t.Errorf("期望 Role=assistant，got %s", item.OutputMessage.Role)
					}
				}
			},
		},
		{
			name:      "UnmarshalJSON_ItemReference",
			jsonInput: `{"type":"item_reference","id":"item_123"}`,
			wantError: false,
			validator: func(t *testing.T, item InputItem) {
				if item.ItemReference == nil {
					t.Error("期望 ItemReference 非空，但为 nil")
				}
				if item.ItemReference != nil {
					if item.ItemReference.ID != "item_123" {
						t.Errorf("期望 ID=item_123，got %s", item.ItemReference.ID)
					}
				}
			},
		},
		{
			name:      "UnmarshalJSON_UnknownType",
			jsonInput: `{"type":"unknown"}`,
			wantError: true,
			validator: nil,
		},
		{
			name:      "UnmarshalJSON_Null",
			jsonInput: `null`,
			wantError: false,
			validator: func(t *testing.T, item InputItem) {
				if item.Message != nil || item.OutputMessage != nil || item.ItemReference != nil {
					t.Error("期望所有字段为空，但有非空字段")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got InputItem
			err := json.Unmarshal([]byte(tt.jsonInput), &got)

			if tt.wantError {
				if err == nil {
					t.Error("期望反序列化失败，但没有返回错误")
				} else {
					// 验证错误码是否为 ErrCodeInvalidArgument
					var portalErr *portalErrors.Error
					if portalErrors.As(err, &portalErr) {
						if portalErr.Code != portalErrors.ErrCodeInvalidArgument {
							t.Errorf("错误码不匹配: got %s, want %s", portalErr.Code, portalErrors.ErrCodeInvalidArgument)
						}
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("反序列化失败: %v", err)
			}

			if tt.validator != nil {
				tt.validator(t, got)
			}
		})
	}
}

// TestInputItemRoundTrip 测试 InputItem 的序列化/反序列化往返
func TestInputItemRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input InputItem
	}{
		{
			name: "Message 类型 round-trip",
			input: InputItem{
				Message: &InputMessage{
					Type: InputItemTypeMessage,
					Role: ResponseMessageRoleUser,
					Content: InputMessageContent{
						String: ptrString("hello world"),
					},
				},
			},
		},
		{
			name: "OutputMessage 类型 round-trip",
			input: InputItem{
				OutputMessage: &OutputMessage{
					Type:   "message",
					ID:     "msg_456",
					Role:   "assistant",
					Status: "in_progress",
					Content: []OutputMessageContent{
						{
							OutputText: &OutputTextContent{
								Type:        "output_text",
								Text:        "processing...",
								Annotations: []Annotation{},
							},
						},
					},
				},
			},
		},
		{
			name: "ItemReference 类型 round-trip",
			input: InputItem{
				ItemReference: &ItemReferenceParam{
					Type: ptrInputItemType(InputItemTypeItemReference),
					ID:   "item_789",
				},
			},
		},
		{
			name: "FunctionCall 类型 round-trip",
			input: InputItem{
				FunctionCall: &FunctionToolCall{
					Type:      InputItemTypeFunctionCall,
					ID:        ptrString("call_001"),
					CallID:    "call_abc",
					Name:      "my_function",
					Arguments: `{"param":"value"}`,
					Status:    ptrString("in_progress"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 序列化
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("序列化失败: %v", err)
			}

			// 反序列化
			var got InputItem
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("反序列化失败: %v", err)
			}

			// 验证 Message 字段
			if (tt.input.Message == nil) != (got.Message == nil) {
				t.Error("Round-trip 后 Message 状态不一致")
			}

			// 验证 OutputMessage 字段
			if (tt.input.OutputMessage == nil) != (got.OutputMessage == nil) {
				t.Error("Round-trip 后 OutputMessage 状态不一致")
			} else if tt.input.OutputMessage != nil && got.OutputMessage != nil {
				if tt.input.OutputMessage.ID != got.OutputMessage.ID {
					t.Errorf("Round-trip 后 OutputMessage.ID 不一致: got %s, want %s", got.OutputMessage.ID, tt.input.OutputMessage.ID)
				}
			}

			// 验证 ItemReference 字段
			if (tt.input.ItemReference == nil) != (got.ItemReference == nil) {
				t.Error("Round-trip 后 ItemReference 状态不一致")
			} else if tt.input.ItemReference != nil && got.ItemReference != nil {
				if tt.input.ItemReference.ID != got.ItemReference.ID {
					t.Errorf("Round-trip 后 ItemReference.ID 不一致: got %s, want %s", got.ItemReference.ID, tt.input.ItemReference.ID)
				}
			}

			// 验证 FunctionCall 字段
			if (tt.input.FunctionCall == nil) != (got.FunctionCall == nil) {
				t.Error("Round-trip 后 FunctionCall 状态不一致")
			} else if tt.input.FunctionCall != nil && got.FunctionCall != nil {
				if tt.input.FunctionCall.Name != got.FunctionCall.Name {
					t.Errorf("Round-trip 后 FunctionCall.Name 不一致: got %s, want %s", got.FunctionCall.Name, tt.input.FunctionCall.Name)
				}
			}
		})
	}
}

// 辅助函数
func ptrInputItemType(t InputItemType) *InputItemType {
	return &t
}
