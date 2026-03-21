package adapter

import (
	"testing"
	"unicode/utf8"
)

func TestBuildRequestBodyPreview_NoTruncate(t *testing.T) {
	body := []byte(`{"model":"gpt-4o-mini","input":"hello"}`)

	preview, truncated := buildRequestBodyPreview(body, 1024)
	if truncated {
		t.Fatalf("不应发生截断")
	}
	if preview != string(body) {
		t.Fatalf("预览内容不匹配，got=%q want=%q", preview, string(body))
	}
}

func TestBuildRequestBodyPreview_TruncateWithUTF8Boundary(t *testing.T) {
	body := []byte(`{"text":"你好世界你好世界你好世界"}`)

	preview, truncated := buildRequestBodyPreview(body, 16)
	if !truncated {
		t.Fatalf("应标记为已截断")
	}
	if len(preview) > 16 {
		t.Fatalf("预览长度不应超过上限，got=%d", len(preview))
	}
	if !utf8.ValidString(preview) {
		t.Fatalf("预览应保持有效 UTF-8")
	}
}
