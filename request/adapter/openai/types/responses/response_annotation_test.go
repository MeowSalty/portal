package responses

import (
	"encoding/json"
	"testing"
)

// TestAnnotation_JSONRoundTrip 测试 Annotation 的 JSON 序列化和反序列化
func TestAnnotation_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		expected Annotation
	}{
		{
			name: "FileCitationAnnotation",
			jsonStr: `{
				"type": "file_citation",
				"file_id": "file_123",
				"index": 0,
				"filename": "document.pdf"
			}`,
			expected: Annotation{
				FileCitation: &FileCitationAnnotation{
					FileID:   "file_123",
					Index:    0,
					Filename: "document.pdf",
				},
			},
		},
		{
			name: "URLCitationAnnotation",
			jsonStr: `{
				"type": "url_citation",
				"url": "https://example.com",
				"start_index": 10,
				"end_index": 20,
				"title": "Example Title"
			}`,
			expected: Annotation{
				URLCitation: &URLCitationAnnotation{
					URL:        "https://example.com",
					StartIndex: 10,
					EndIndex:   20,
					Title:      "Example Title",
				},
			},
		},
		{
			name: "ContainerFileCitationAnnotation",
			jsonStr: `{
				"type": "container_file_citation",
				"container_id": "container_123",
				"file_id": "file_456",
				"start_index": 5,
				"end_index": 15,
				"filename": "container_file.txt"
			}`,
			expected: Annotation{
				ContainerFileCitation: &ContainerFileCitationAnnotation{
					ContainerID: "container_123",
					FileID:      "file_456",
					StartIndex:  5,
					EndIndex:    15,
					Filename:    "container_file.txt",
				},
			},
		},
		{
			name: "FilePathAnnotation",
			jsonStr: `{
				"type": "file_path",
				"file_id": "file_789",
				"index": 1
			}`,
			expected: Annotation{
				FilePath: &FilePathAnnotation{
					FileID: "file_789",
					Index:  1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 反序列化
			var ann Annotation
			if err := json.Unmarshal([]byte(tt.jsonStr), &ann); err != nil {
				t.Fatalf("反序列化失败：%v", err)
			}

			// 验证类型
			if ann.GetType() == "" {
				t.Fatal("GetType 返回空字符串")
			}

			// 根据类型验证具体字段
			switch {
			case ann.FileCitation != nil:
				expected := tt.expected.FileCitation
				if expected == nil {
					t.Fatalf("类型不匹配：期望 FileCitationAnnotation")
				}
				if ann.FileCitation.FileID != expected.FileID {
					t.Errorf("FileID 不匹配：得到 %s, 期望 %s", ann.FileCitation.FileID, expected.FileID)
				}
				if ann.FileCitation.Index != expected.Index {
					t.Errorf("Index 不匹配：得到 %d, 期望 %d", ann.FileCitation.Index, expected.Index)
				}
				if ann.FileCitation.Filename != expected.Filename {
					t.Errorf("Filename 不匹配：得到 %s, 期望 %s", ann.FileCitation.Filename, expected.Filename)
				}

			case ann.URLCitation != nil:
				expected := tt.expected.URLCitation
				if expected == nil {
					t.Fatalf("类型不匹配：期望 URLCitationAnnotation")
				}
				if ann.URLCitation.URL != expected.URL {
					t.Errorf("URL 不匹配：得到 %s, 期望 %s", ann.URLCitation.URL, expected.URL)
				}
				if ann.URLCitation.StartIndex != expected.StartIndex {
					t.Errorf("StartIndex 不匹配：得到 %d, 期望 %d", ann.URLCitation.StartIndex, expected.StartIndex)
				}
				if ann.URLCitation.EndIndex != expected.EndIndex {
					t.Errorf("EndIndex 不匹配：得到 %d, 期望 %d", ann.URLCitation.EndIndex, expected.EndIndex)
				}
				if ann.URLCitation.Title != expected.Title {
					t.Errorf("Title 不匹配：得到 %s, 期望 %s", ann.URLCitation.Title, expected.Title)
				}

			case ann.ContainerFileCitation != nil:
				expected := tt.expected.ContainerFileCitation
				if expected == nil {
					t.Fatalf("类型不匹配：期望 ContainerFileCitationAnnotation")
				}
				if ann.ContainerFileCitation.ContainerID != expected.ContainerID {
					t.Errorf("ContainerID 不匹配：得到 %s, 期望 %s", ann.ContainerFileCitation.ContainerID, expected.ContainerID)
				}
				if ann.ContainerFileCitation.FileID != expected.FileID {
					t.Errorf("FileID 不匹配：得到 %s, 期望 %s", ann.ContainerFileCitation.FileID, expected.FileID)
				}
				if ann.ContainerFileCitation.StartIndex != expected.StartIndex {
					t.Errorf("StartIndex 不匹配：得到 %d, 期望 %d", ann.ContainerFileCitation.StartIndex, expected.StartIndex)
				}
				if ann.ContainerFileCitation.EndIndex != expected.EndIndex {
					t.Errorf("EndIndex 不匹配：得到 %d, 期望 %d", ann.ContainerFileCitation.EndIndex, expected.EndIndex)
				}
				if ann.ContainerFileCitation.Filename != expected.Filename {
					t.Errorf("Filename 不匹配：得到 %s, 期望 %s", ann.ContainerFileCitation.Filename, expected.Filename)
				}

			case ann.FilePath != nil:
				expected := tt.expected.FilePath
				if expected == nil {
					t.Fatalf("类型不匹配：期望 FilePathAnnotation")
				}
				if ann.FilePath.FileID != expected.FileID {
					t.Errorf("FileID 不匹配：得到 %s, 期望 %s", ann.FilePath.FileID, expected.FileID)
				}
				if ann.FilePath.Index != expected.Index {
					t.Errorf("Index 不匹配：得到 %d, 期望 %d", ann.FilePath.Index, expected.Index)
				}
			}

			// 序列化回 JSON
			jsonData, err := json.Marshal(ann)
			if err != nil {
				t.Fatalf("序列化失败：%v", err)
			}

			// 再次反序列化验证往返一致性
			var ann2 Annotation
			if err := json.Unmarshal(jsonData, &ann2); err != nil {
				t.Fatalf("往返反序列化失败：%v", err)
			}

			// 验证类型保持一致
			if ann2.GetType() == "" {
				t.Fatal("往返后 GetType 返回空字符串")
			}
			if ann2.GetType() != ann.GetType() {
				t.Errorf("往返后类型不匹配：得到 %s, 期望 %s", ann2.GetType(), ann.GetType())
			}
		})
	}
}

// TestAnnotation_ValidateOneOf 测试 validateOneOf 互斥校验
func TestAnnotation_ValidateOneOf(t *testing.T) {
	tests := []struct {
		name        string
		annotation  Annotation
		expectError bool
	}{
		{
			name:        "全空应该报错",
			annotation:  Annotation{},
			expectError: true,
		},
		{
			name: "多个指针非空应该报错",
			annotation: Annotation{
				FileCitation: &FileCitationAnnotation{},
				URLCitation:  &URLCitationAnnotation{},
			},
			expectError: true,
		},
		{
			name: "仅 FileCitation 非空应该通过",
			annotation: Annotation{
				FileCitation: &FileCitationAnnotation{FileID: "file_123"},
			},
			expectError: false,
		},
		{
			name: "仅 URLCitation 非空应该通过",
			annotation: Annotation{
				URLCitation: &URLCitationAnnotation{URL: "https://example.com"},
			},
			expectError: false,
		},
		{
			name: "仅 ContainerFileCitation 非空应该通过",
			annotation: Annotation{
				ContainerFileCitation: &ContainerFileCitationAnnotation{FileID: "file_456"},
			},
			expectError: false,
		},
		{
			name: "仅 FilePath 非空应该通过",
			annotation: Annotation{
				FilePath: &FilePathAnnotation{FileID: "file_789"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.annotation.validateOneOf()
			if tt.expectError && err == nil {
				t.Error("期望报错但没有报错")
			}
			if !tt.expectError && err != nil {
				t.Errorf("不期望报错但报错了：%v", err)
			}
		})
	}
}

// TestAnnotation_MarshalJSON 测试序列化错误情况
func TestAnnotation_MarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		annotation  Annotation
		expectError bool
	}{
		{
			name:        "全空应该报错",
			annotation:  Annotation{},
			expectError: true,
		},
		{
			name: "多个指针非空应该报错",
			annotation: Annotation{
				FileCitation: &FileCitationAnnotation{},
				URLCitation:  &URLCitationAnnotation{},
			},
			expectError: true,
		},
		{
			name: "仅 FileCitation 非空应该通过",
			annotation: Annotation{
				FileCitation: &FileCitationAnnotation{FileID: "file_123"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := json.Marshal(tt.annotation)
			if tt.expectError && err == nil {
				t.Error("期望报错但没有报错")
			}
			if !tt.expectError && err != nil {
				t.Errorf("不期望报错但报错了：%v", err)
			}
		})
	}
}

// TestAnnotation_UnmarshalJSON 测试反序列化错误情况
func TestAnnotation_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonStr     string
		expectError bool
	}{
		{
			name:        "null 应该通过",
			jsonStr:     "null",
			expectError: false,
		},
		{
			name:        "未知类型应该报错",
			jsonStr:     `{"type": "unknown_type"}`,
			expectError: true,
		},
		{
			name:        "缺少 type 字段应该报错",
			jsonStr:     `{"file_id": "file_123"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ann Annotation
			err := json.Unmarshal([]byte(tt.jsonStr), &ann)
			if tt.expectError && err == nil {
				t.Error("期望报错但没有报错")
			}
			if !tt.expectError && err != nil {
				t.Errorf("不期望报错但报错了：%v", err)
			}
		})
	}
}

// TestContainerFileCitationAnnotation_NewFields 测试 ContainerFileCitationAnnotation 新增字段
func TestContainerFileCitationAnnotation_NewFields(t *testing.T) {
	jsonStr := `{
		"type": "container_file_citation",
		"container_id": "container_123",
		"file_id": "file_456",
		"start_index": 5,
		"end_index": 15,
		"filename": "container_file.txt"
	}`

	var ann Annotation
	if err := json.Unmarshal([]byte(jsonStr), &ann); err != nil {
		t.Fatalf("反序列化失败：%v", err)
	}

	if ann.ContainerFileCitation == nil {
		t.Fatalf("ContainerFileCitation 为空")
	}

	cfc := ann.ContainerFileCitation

	// 验证新增的必填字段
	if cfc.StartIndex != 5 {
		t.Errorf("StartIndex 不匹配：得到 %d, 期望 5", cfc.StartIndex)
	}
	if cfc.EndIndex != 15 {
		t.Errorf("EndIndex 不匹配：得到 %d, 期望 15", cfc.EndIndex)
	}
}

// TestFilePathAnnotation_NewFields 测试 FilePathAnnotation 新增字段
func TestFilePathAnnotation_NewFields(t *testing.T) {
	jsonStr := `{
		"type": "file_path",
		"file_id": "file_789",
		"index": 1
	}`

	var ann Annotation
	if err := json.Unmarshal([]byte(jsonStr), &ann); err != nil {
		t.Fatalf("反序列化失败：%v", err)
	}

	if ann.FilePath == nil {
		t.Fatalf("FilePath 为空")
	}

	fp := ann.FilePath

	// 验证新增的必填字段
	if fp.Index != 1 {
		t.Errorf("Index 不匹配：得到 %d, 期望 1", fp.Index)
	}
}

// TestAnnotation_Array 测试 Annotation 数组的序列化和反序列化
func TestAnnotation_Array(t *testing.T) {
	jsonStr := `[
		{
			"type": "file_citation",
			"file_id": "file_123",
			"index": 0,
			"filename": "document.pdf"
		},
		{
			"type": "url_citation",
			"url": "https://example.com",
			"start_index": 10,
			"end_index": 20,
			"title": "Example Title"
		},
		{
			"type": "container_file_citation",
			"container_id": "container_123",
			"file_id": "file_456",
			"start_index": 5,
			"end_index": 15,
			"filename": "container_file.txt"
		},
		{
			"type": "file_path",
			"file_id": "file_789",
			"index": 1
		}
	]`

	var annotations []Annotation
	if err := json.Unmarshal([]byte(jsonStr), &annotations); err != nil {
		t.Fatalf("反序列化失败：%v", err)
	}

	if len(annotations) != 4 {
		t.Fatalf("注释数量不匹配：得到 %d, 期望 4", len(annotations))
	}

	// 验证每个注释的类型
	if annotations[0].GetType() != AnnotationTypeFileCitation {
		t.Errorf("第一个注释类型不匹配")
	}
	if annotations[1].GetType() != AnnotationTypeURLCitation {
		t.Errorf("第二个注释类型不匹配")
	}
	if annotations[2].GetType() != AnnotationTypeContainerFileCitation {
		t.Errorf("第三个注释类型不匹配")
	}
	if annotations[3].GetType() != AnnotationTypeFilePath {
		t.Errorf("第四个注释类型不匹配")
	}

	// 序列化回 JSON
	jsonData, err := json.Marshal(annotations)
	if err != nil {
		t.Fatalf("序列化失败：%v", err)
	}

	// 再次反序列化验证往返一致性
	var annotations2 []Annotation
	if err := json.Unmarshal(jsonData, &annotations2); err != nil {
		t.Fatalf("往返反序列化失败：%v", err)
	}

	if len(annotations2) != 4 {
		t.Fatalf("往返后注释数量不匹配：得到 %d, 期望 4", len(annotations2))
	}
}
