package responses

import (
	"encoding/json"
	"testing"

	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

// TestInputContentMarshalJSON 测试 InputContent 的 JSON 序列化
func TestInputContentMarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     InputContent
		wantJSON  string
		wantError bool
	}{
		{
			name:      "空值序列化为 null",
			input:     InputContent{},
			wantJSON:  "null",
			wantError: false,
		},
		{
			name: "仅 Text 类型",
			input: InputContent{
				Text: &InputTextContent{
					Type: InputContentTypeText,
					Text: "hello world",
				},
			},
			wantJSON:  `{"type":"input_text","text":"hello world"}`,
			wantError: false,
		},
		{
			name: "仅 Text 类型 - 空文本",
			input: InputContent{
				Text: &InputTextContent{
					Type: InputContentTypeText,
					Text: "",
				},
			},
			wantJSON:  `{"type":"input_text","text":""}`,
			wantError: false,
		},
		{
			name: "仅 Image 类型 - image_url",
			input: InputContent{
				Image: &InputImageContent{
					Type:     InputContentTypeImage,
					ImageURL: ptrString("https://example.com/image.png"),
				},
			},
			wantJSON:  `{"type":"input_image","image_url":"https://example.com/image.png","detail":"auto"}`,
			wantError: false,
		},
		{
			name: "仅 Image 类型 - file_id",
			input: InputContent{
				Image: &InputImageContent{
					Type:   InputContentTypeImage,
					FileID: ptrString("file_12345"),
				},
			},
			wantJSON:  `{"type":"input_image","file_id":"file_12345","detail":"auto"}`,
			wantError: false,
		},
		{
			name: "仅 Image 类型 - image_url + detail",
			input: InputContent{
				Image: &InputImageContent{
					Type:     InputContentTypeImage,
					ImageURL: ptrString("https://example.com/image.png"),
					Detail:   ptrImageDetail(shared.ImageDetailHigh),
				},
			},
			wantJSON:  `{"type":"input_image","image_url":"https://example.com/image.png","detail":"high"}`,
			wantError: false,
		},
		{
			name: "仅 Image 类型 - file_id + detail",
			input: InputContent{
				Image: &InputImageContent{
					Type:   InputContentTypeImage,
					FileID: ptrString("file_12345"),
					Detail: ptrImageDetail(shared.ImageDetailLow),
				},
			},
			wantJSON:  `{"type":"input_image","file_id":"file_12345","detail":"low"}`,
			wantError: false,
		},
		{
			name: "仅 Image 类型 - 仅 detail",
			input: InputContent{
				Image: &InputImageContent{
					Type:   InputContentTypeImage,
					Detail: ptrImageDetail(shared.ImageDetailAuto),
				},
			},
			wantJSON:  `{"type":"input_image","detail":"auto"}`,
			wantError: false,
		},
		{
			name: "仅 File 类型 - file_id",
			input: InputContent{
				File: &InputFileContent{
					Type:   InputContentTypeFile,
					FileID: ptrString("file_12345"),
				},
			},
			wantJSON:  `{"type":"input_file","file_id":"file_12345"}`,
			wantError: false,
		},
		{
			name: "仅 File 类型 - filename",
			input: InputContent{
				File: &InputFileContent{
					Type:     InputContentTypeFile,
					Filename: ptrString("document.pdf"),
				},
			},
			wantJSON:  `{"type":"input_file","filename":"document.pdf"}`,
			wantError: false,
		},
		{
			name: "仅 File 类型 - file_data",
			input: InputContent{
				File: &InputFileContent{
					Type:     InputContentTypeFile,
					FileData: ptrString("base64encodeddata"),
				},
			},
			wantJSON:  `{"type":"input_file","file_data":"base64encodeddata"}`,
			wantError: false,
		},
		{
			name: "仅 File 类型 - file_url",
			input: InputContent{
				File: &InputFileContent{
					Type:    InputContentTypeFile,
					FileURL: ptrString("https://example.com/file.pdf"),
				},
			},
			wantJSON:  `{"type":"input_file","file_url":"https://example.com/file.pdf"}`,
			wantError: false,
		},
		{
			name: "仅 File 类型 - 多字段组合",
			input: InputContent{
				File: &InputFileContent{
					Type:     InputContentTypeFile,
					FileID:   ptrString("file_12345"),
					Filename: ptrString("document.pdf"),
				},
			},
			wantJSON:  `{"type":"input_file","file_id":"file_12345","filename":"document.pdf"}`,
			wantError: false,
		},
		{
			name: "oneof 冲突 - Text + Image",
			input: InputContent{
				Text: &InputTextContent{
					Type: InputContentTypeText,
					Text: "hello",
				},
				Image: &InputImageContent{
					Type:   InputContentTypeImage,
					FileID: ptrString("file_12345"),
				},
			},
			wantJSON:  "",
			wantError: true,
		},
		{
			name: "oneof 冲突 - Text + File",
			input: InputContent{
				Text: &InputTextContent{
					Type: InputContentTypeText,
					Text: "hello",
				},
				File: &InputFileContent{
					Type:   InputContentTypeFile,
					FileID: ptrString("file_12345"),
				},
			},
			wantJSON:  "",
			wantError: true,
		},
		{
			name: "oneof 冲突 - Image + File",
			input: InputContent{
				Image: &InputImageContent{
					Type:   InputContentTypeImage,
					FileID: ptrString("file_12345"),
				},
				File: &InputFileContent{
					Type:   InputContentTypeFile,
					FileID: ptrString("file_67890"),
				},
			},
			wantJSON:  "",
			wantError: true,
		},
		{
			name: "oneof 冲突 - Text + Image + File",
			input: InputContent{
				Text: &InputTextContent{
					Type: InputContentTypeText,
					Text: "hello",
				},
				Image: &InputImageContent{
					Type:   InputContentTypeImage,
					FileID: ptrString("file_12345"),
				},
				File: &InputFileContent{
					Type:   InputContentTypeFile,
					FileID: ptrString("file_67890"),
				},
			},
			wantJSON:  "",
			wantError: true,
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

// TestInputContentUnmarshalJSON 测试 InputContent 的 JSON 反序列化
func TestInputContentUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantType  string // "text", "image", "file", "none"
		wantError bool
	}{
		{
			name:      "null 值",
			jsonInput: `null`,
			wantType:  "none",
			wantError: false,
		},
		{
			name:      "缺少 type 字段",
			jsonInput: `{"text":"hello"}`,
			wantType:  "",
			wantError: true,
		},
		{
			name:      "空对象",
			jsonInput: `{}`,
			wantType:  "",
			wantError: true,
		},
		{
			name:      "type 为空字符串",
			jsonInput: `{"type":""}`,
			wantType:  "",
			wantError: true,
		},
		{
			name:      "type = input_text",
			jsonInput: `{"type":"input_text","text":"hello world"}`,
			wantType:  "text",
			wantError: false,
		},
		{
			name:      "type = input_text - 空文本",
			jsonInput: `{"type":"input_text","text":""}`,
			wantType:  "text",
			wantError: false,
		},
		{
			name:      "type = input_image - image_url",
			jsonInput: `{"type":"input_image","image_url":"https://example.com/image.png"}`,
			wantType:  "image",
			wantError: false,
		},
		{
			name:      "type = input_image - file_id",
			jsonInput: `{"type":"input_image","file_id":"file_12345"}`,
			wantType:  "image",
			wantError: false,
		},
		{
			name:      "type = input_image - image_url + detail",
			jsonInput: `{"type":"input_image","image_url":"https://example.com/image.png","detail":"high"}`,
			wantType:  "image",
			wantError: false,
		},
		{
			name:      "type = input_image - file_id + detail",
			jsonInput: `{"type":"input_image","file_id":"file_12345","detail":"low"}`,
			wantType:  "image",
			wantError: false,
		},
		{
			name:      "type = input_image - 仅 detail",
			jsonInput: `{"type":"input_image","detail":"auto"}`,
			wantType:  "image",
			wantError: false,
		},
		{
			name:      "type = input_file - file_id",
			jsonInput: `{"type":"input_file","file_id":"file_12345"}`,
			wantType:  "file",
			wantError: false,
		},
		{
			name:      "type = input_file - filename",
			jsonInput: `{"type":"input_file","filename":"document.pdf"}`,
			wantType:  "file",
			wantError: false,
		},
		{
			name:      "type = input_file - file_data",
			jsonInput: `{"type":"input_file","file_data":"base64encodeddata"}`,
			wantType:  "file",
			wantError: false,
		},
		{
			name:      "type = input_file - file_url",
			jsonInput: `{"type":"input_file","file_url":"https://example.com/file.pdf"}`,
			wantType:  "file",
			wantError: false,
		},
		{
			name:      "type = input_file - 多字段组合",
			jsonInput: `{"type":"input_file","file_id":"file_12345","filename":"document.pdf"}`,
			wantType:  "file",
			wantError: false,
		},
		{
			name:      "不支持的 type - input_video",
			jsonInput: `{"type":"input_video","url":"https://example.com/video.mp4"}`,
			wantType:  "",
			wantError: true,
		},
		{
			name:      "不支持的 type - unknown_type",
			jsonInput: `{"type":"unknown_type","data":"value"}`,
			wantType:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c InputContent
			err := json.Unmarshal([]byte(tt.jsonInput), &c)

			if tt.wantError {
				if err == nil {
					t.Error("期望反序列化失败，但没有返回错误")
				}
				return
			}

			if err != nil {
				t.Fatalf("反序列化失败: %v", err)
			}

			// 验证类型分支
			switch tt.wantType {
			case "text":
				if c.Text == nil {
					t.Error("期望 Text 不为 nil")
				}
				if c.Image != nil {
					t.Error("期望 Image 为 nil")
				}
				if c.File != nil {
					t.Error("期望 File 为 nil")
				}
			case "image":
				if c.Image == nil {
					t.Error("期望 Image 不为 nil")
				}
				if c.Text != nil {
					t.Error("期望 Text 为 nil")
				}
				if c.File != nil {
					t.Error("期望 File 为 nil")
				}
			case "file":
				if c.File == nil {
					t.Error("期望 File 不为 nil")
				}
				if c.Text != nil {
					t.Error("期望 Text 为 nil")
				}
				if c.Image != nil {
					t.Error("期望值 Image 为 nil")
				}
			case "none":
				if c.Text != nil {
					t.Error("期望 Text 为 nil")
				}
				if c.Image != nil {
					t.Error("期望 Image 为 nil")
				}
				if c.File != nil {
					t.Error("期望 File 为 nil")
				}
			}
		})
	}
}

// TestInputContentRoundTrip 测试 InputContent 的序列化/反序列化闭环一致性
func TestInputContentRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input InputContent
	}{
		{
			name: "Text 类型 round-trip",
			input: InputContent{
				Text: &InputTextContent{
					Type: InputContentTypeText,
					Text: "hello world",
				},
			},
		},
		{
			name: "Text 类型 round-trip - 空文本",
			input: InputContent{
				Text: &InputTextContent{
					Type: InputContentTypeText,
					Text: "",
				},
			},
		},
		{
			name: "Image 类型 round-trip - image_url",
			input: InputContent{
				Image: &InputImageContent{
					Type:     InputContentTypeImage,
					ImageURL: ptrString("https://example.com/image.png"),
					Detail:   ptrImageDetail(shared.ImageDetailAuto),
				},
			},
		},
		{
			name: "Image 类型 round-trip - file_id",
			input: InputContent{
				Image: &InputImageContent{
					Type:   InputContentTypeImage,
					FileID: ptrString("file_12345"),
					Detail: ptrImageDetail(shared.ImageDetailAuto),
				},
			},
		},
		{
			name: "Image 类型 round-trip - image_url + detail",
			input: InputContent{
				Image: &InputImageContent{
					Type:     InputContentTypeImage,
					ImageURL: ptrString("https://example.com/image.png"),
					Detail:   ptrImageDetail(shared.ImageDetailHigh),
				},
			},
		},
		{
			name: "Image 类型 round-trip - file_id + detail",
			input: InputContent{
				Image: &InputImageContent{
					Type:   InputContentTypeImage,
					FileID: ptrString("file_12345"),
					Detail: ptrImageDetail(shared.ImageDetailLow),
				},
			},
		},
		{
			name: "File 类型 round-trip - file_id",
			input: InputContent{
				File: &InputFileContent{
					Type:   InputContentTypeFile,
					FileID: ptrString("file_12345"),
				},
			},
		},
		{
			name: "File 类型 round-trip - filename",
			input: InputContent{
				File: &InputFileContent{
					Type:     InputContentTypeFile,
					Filename: ptrString("document.pdf"),
				},
			},
		},
		{
			name: "File 类型 round-trip - file_data",
			input: InputContent{
				File: &InputFileContent{
					Type:     InputContentTypeFile,
					FileData: ptrString("base64encodeddata"),
				},
			},
		},
		{
			name: "File 类型 round-trip - file_url",
			input: InputContent{
				File: &InputFileContent{
					Type:    InputContentTypeFile,
					FileURL: ptrString("https://example.com/file.pdf"),
				},
			},
		},
		{
			name: "File 类型 round-trip - 多字段组合",
			input: InputContent{
				File: &InputFileContent{
					Type:     InputContentTypeFile,
					FileID:   ptrString("file_12345"),
					Filename: ptrString("document.pdf"),
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
			var got InputContent
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("反序列化失败: %v", err)
			}

			// 验证 Text 字段
			if (tt.input.Text == nil) != (got.Text == nil) {
				t.Error("Round-trip 后 Text 状态不一致")
			} else if tt.input.Text != nil {
				if got.Text == nil {
					t.Error("Round-trip 后 Text 变为 nil")
				} else {
					if tt.input.Text.Type != got.Text.Type {
						t.Errorf("Round-trip 后 Text.Type 不一致: got %q, want %q", got.Text.Type, tt.input.Text.Type)
					}
					if tt.input.Text.Text != got.Text.Text {
						t.Errorf("Round-trip 后 Text.Text 不一致: got %q, want %q", got.Text.Text, tt.input.Text.Text)
					}
				}
			}

			// 验证 Image 字段
			if (tt.input.Image == nil) != (got.Image == nil) {
				t.Error("Round-trip 后 Image 状态不一致")
			} else if tt.input.Image != nil {
				if got.Image == nil {
					t.Error("Round-trip 后 Image 变为 nil")
				} else {
					if tt.input.Image.Type != got.Image.Type {
						t.Errorf("Round-trip 后 Image.Type 不一致: got %q, want %q", got.Image.Type, tt.input.Image.Type)
					}
					if (tt.input.Image.ImageURL == nil) != (got.Image.ImageURL == nil) {
						t.Error("Round-trip 后 Image.ImageURL 状态不一致")
					} else if tt.input.Image.ImageURL != nil && got.Image.ImageURL != nil {
						if *tt.input.Image.ImageURL != *got.Image.ImageURL {
							t.Errorf("Round-trip 后 Image.ImageURL 不一致: got %q, want %q", *got.Image.ImageURL, *tt.input.Image.ImageURL)
						}
					}
					if (tt.input.Image.FileID == nil) != (got.Image.FileID == nil) {
						t.Error("Round-trip 后 Image.FileID 状态不一致")
					} else if tt.input.Image.FileID != nil && got.Image.FileID != nil {
						if *tt.input.Image.FileID != *got.Image.FileID {
							t.Errorf("Round-trip 后 Image.FileID 不一致: got %q, want %q", *got.Image.FileID, *tt.input.Image.FileID)
						}
					}
					// Detail 字段在 nil 时会被自动设置为 auto，所以需要特殊处理
					if tt.input.Image.Detail == nil {
						// 输入为 nil 时，期望输出为 auto
						if got.Image.Detail == nil || *got.Image.Detail != shared.ImageDetailAuto {
							t.Errorf("Round-trip 后 Image.Detail 应为 auto，got %v", got.Image.Detail)
						}
					} else {
						// 输入不为 nil 时，期望输出与输入一致
						if got.Image.Detail == nil {
							t.Error("Round-trip 后 Image.Detail 变为 nil")
						} else if *tt.input.Image.Detail != *got.Image.Detail {
							t.Errorf("Round-trip 后 Image.Detail 不一致: got %q, want %q", *got.Image.Detail, *tt.input.Image.Detail)
						}
					}
				}
			}

			// 验证 File 字段
			if (tt.input.File == nil) != (got.File == nil) {
				t.Error("Round-trip 后 File 状态不一致")
			} else if tt.input.File != nil {
				if got.File == nil {
					t.Error("Round-trip 后 File 变为 nil")
				} else {
					if tt.input.File.Type != got.File.Type {
						t.Errorf("Round-trip 后 File.Type 不一致: got %q, want %q", got.File.Type, tt.input.File.Type)
					}
					if (tt.input.File.FileID == nil) != (got.File.FileID == nil) {
						t.Error("Round-trip 后 File.FileID 状态不一致")
					} else if tt.input.File.FileID != nil && got.File.FileID != nil {
						if *tt.input.File.FileID != *got.File.FileID {
							t.Errorf("Round-trip 后 File.FileID 不一致: got %q, want %q", *got.File.FileID, *tt.input.File.FileID)
						}
					}
					if (tt.input.File.Filename == nil) != (got.File.Filename == nil) {
						t.Error("Round-trip 后 File.Filename 状态不一致")
					} else if tt.input.File.Filename != nil && got.File.Filename != nil {
						if *tt.input.File.Filename != *got.File.Filename {
							t.Errorf("Round-trip 后 File.Filename 不一致: got %q, want %q", *got.File.Filename, *tt.input.File.Filename)
						}
					}
					if (tt.input.File.FileData == nil) != (got.File.FileData == nil) {
						t.Error("Round-trip 后 File.FileData 状态不一致")
					} else if tt.input.File.FileData != nil && got.File.FileData != nil {
						if *tt.input.File.FileData != *got.File.FileData {
							t.Errorf("Round-trip 后 File.FileData 不一致: got %q, want %q", *got.File.FileData, *tt.input.File.FileData)
						}
					}
					if (tt.input.File.FileURL == nil) != (got.File.FileURL == nil) {
						t.Error("Round-trip 后 File.FileURL 状态不一致")
					} else if tt.input.File.FileURL != nil && got.File.FileURL != nil {
						if *tt.input.File.FileURL != *got.File.FileURL {
							t.Errorf("Round-trip 后 File.FileURL 不一致: got %q, want %q", *got.File.FileURL, *tt.input.File.FileURL)
						}
					}
				}
			}
		})
	}
}

// TestInputImageContentDefaultDetail 测试 InputImageContent 的 detail 默认值行为
func TestInputImageContentDefaultDetail(t *testing.T) {
	t.Run("MarshalJSON_Image_DefaultDetail", func(t *testing.T) {
		// 测试：当 detail 为 nil 时，序列化应输出 "auto"
		input := InputContent{
			Image: &InputImageContent{
				Type:     InputContentTypeImage,
				ImageURL: ptrString("https://example.com/image.png"),
				Detail:   nil, // detail 为 nil
			},
		}

		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("序列化失败: %v", err)
		}

		// 验证 JSON 中包含 detail="auto"
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("解析 JSON 失败: %v", err)
		}

		if result["detail"] != "auto" {
			t.Errorf("期望 detail=auto，got %v", result["detail"])
		}
	})

	t.Run("UnmarshalJSON_Image_DefaultDetail", func(t *testing.T) {
		// 测试：当 JSON 中没有 detail 字段时，应设置为 auto
		jsonInput := `{"type":"input_image","image_url":"https://example.com/image.png"}`

		var got InputContent
		if err := json.Unmarshal([]byte(jsonInput), &got); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}

		if got.Image == nil {
			t.Fatal("期望 Image 非空，但为 nil")
		}

		if got.Image.Detail == nil {
			t.Error("期望 Detail 非空，但为 nil")
		} else if *got.Image.Detail != "auto" {
			t.Errorf("期望 Detail=auto，got %s", *got.Image.Detail)
		}
	})
}

// 辅助函数

// ptrString 返回 string 指针
func ptrString(s string) *string {
	return &s
}

// ptrImageDetail 返回 ImageDetail 指针
func ptrImageDetail(d shared.ImageDetail) *shared.ImageDetail {
	return &d
}
