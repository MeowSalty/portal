package responses

import (
	"encoding/json"
	"testing"

	portalErrors "github.com/MeowSalty/portal/errors"
)

// TestInputMessageContentMarshalJSON 测试 InputMessageContent 的 JSON 序列化
func TestInputMessageContentMarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     InputMessageContent
		wantJSON  string
		wantError bool
	}{
		{
			name: "MarshalJSON_String",
			input: InputMessageContent{
				String: ptrString("hi"),
				List:   nil,
			},
			wantJSON:  `"hi"`,
			wantError: false,
		},
		{
			name: "MarshalJSON_List",
			input: InputMessageContent{
				String: nil,
				List: &[]InputContent{
					{
						Text: &InputTextContent{
							Type: InputContentTypeText,
							Text: "hi",
						},
					},
				},
			},
			wantJSON:  `[{"type":"input_text","text":"hi"}]`,
			wantError: false,
		},
		{
			name: "MarshalJSON_StringOverridesList",
			input: InputMessageContent{
				String: ptrString("hi"),
				List: &[]InputContent{
					{
						Text: &InputTextContent{
							Type: InputContentTypeText,
							Text: "hello",
						},
					},
				},
			},
			wantJSON:  `"hi"`,
			wantError: false,
		},
		{
			name: "MarshalJSON_Null",
			input: InputMessageContent{
				String: nil,
				List:   nil,
			},
			wantJSON:  `null`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("期望序列化失败，但没有返回错误")
				}
				return
			}

			if err != nil {
				t.Fatalf("序列化失败: %v", err)
			}

			gotJSON := string(data)
			if gotJSON != tt.wantJSON {
				t.Errorf("序列化结果不匹配:\ngot  %s\nwant %s", gotJSON, tt.wantJSON)
			}
		})
	}
}

// TestInputMessageContentUnmarshalJSON 测试 InputMessageContent 的 JSON 反序列化
func TestInputMessageContentUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantType  string // "string", "list", "none"
		wantError bool
	}{
		{
			name:      "UnmarshalJSON_String",
			jsonInput: `"hi"`,
			wantType:  "string",
			wantError: false,
		},
		{
			name:      "UnmarshalJSON_List",
			jsonInput: `[{"type":"input_text","text":"hi"}]`,
			wantType:  "list",
			wantError: false,
		},
		{
			name:      "UnmarshalJSON_Null",
			jsonInput: `null`,
			wantType:  "none",
			wantError: false,
		},
		{
			name:      "UnmarshalJSON_Invalid - 数字",
			jsonInput: `123`,
			wantType:  "",
			wantError: true,
		},
		{
			name:      "UnmarshalJSON_Invalid - 对象",
			jsonInput: `{}`,
			wantType:  "",
			wantError: true,
		},
		{
			name:      "UnmarshalJSON_Invalid - 布尔值",
			jsonInput: `true`,
			wantType:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got InputMessageContent
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

			switch tt.wantType {
			case "string":
				if got.String == nil {
					t.Error("期望 String 非空，但为 nil")
				}
				if got.List != nil {
					t.Error("期望 List 为空，但非 nil")
				}
			case "list":
				if got.List == nil {
					t.Error("期望 List 非空，但为 nil")
				}
				if got.String != nil {
					t.Error("期望 String 为空，但非 nil")
				}
			case "none":
				if got.String != nil {
					t.Error("期望 String 为空，但非 nil")
				}
				if got.List != nil {
					t.Error("期望 List 为空，但非 nil")
				}
			}
		})
	}
}

// TestInputMessageContentRoundTrip 测试 InputMessageContent 的序列化/反序列化往返
func TestInputMessageContentRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input InputMessageContent
	}{
		{
			name: "String 类型 round-trip",
			input: InputMessageContent{
				String: ptrString("hello world"),
			},
		},
		{
			name: "List 类型 round-trip - 单个元素",
			input: InputMessageContent{
				List: &[]InputContent{
					{
						Text: &InputTextContent{
							Type: InputContentTypeText,
							Text: "hi",
						},
					},
				},
			},
		},
		{
			name: "List 类型 round-trip - 多个元素",
			input: InputMessageContent{
				List: &[]InputContent{
					{
						Text: &InputTextContent{
							Type: InputContentTypeText,
							Text: "hello",
						},
					},
					{
						Image: &InputImageContent{
							Type:     InputContentTypeImage,
							ImageURL: ptrString("https://example.com/image.png"),
						},
					},
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
			var got InputMessageContent
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("反序列化失败: %v", err)
			}

			// 验证 String 字段
			if (tt.input.String == nil) != (got.String == nil) {
				t.Error("Round-trip 后 String 状态不一致")
			} else if tt.input.String != nil && got.String != nil {
				if *tt.input.String != *got.String {
					t.Errorf("Round-trip 后 String 值不一致: got %q, want %q", *got.String, *tt.input.String)
				}
			}

			// 验证 List 字段
			if (tt.input.List == nil) != (got.List == nil) {
				t.Error("Round-trip 后 List 状态不一致")
			} else if tt.input.List != nil && got.List != nil {
				if len(*tt.input.List) != len(*got.List) {
					t.Errorf("Round-trip 后 List 长度不一致：got %d, want %d", len(*got.List), len(*tt.input.List))
				}
			}
		})
	}
}
