package adapter

import (
	"testing"
)

func TestExtractHTMLError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Cloudflare 502 错误页面",
			input: `<!DOCTYPE html>
<html>
<head>
    <title>502 Bad Gateway</title>
</head>
<body>
    <h1>502 Bad Gateway</h1>
    <p>The web server reported a bad gateway error.</p>
    <p class="cf-error-details">Cloudflare Ray ID: abc123</p>
</body>
</html>`,
			expected: "502 Bad Gateway: The web server reported a bad gateway error.",
		},
		{
			name: "Nginx 默认错误页面",
			input: `<html>
<head><title>502 Bad Gateway</title></head>
<body>
<center><h1>502 Bad Gateway</h1></center>
<hr><center>nginx/1.18.0</center>
</body>
</html>`,
			expected: "502 Bad Gateway",
		},
		{
			name:     "只有标题的简单页面",
			input:    `<html><head><title>Service Unavailable</title></head></html>`,
			expected: "Service Unavailable",
		},
		{
			name: "混合内容",
			input: `{"error": "some error"}<!DOCTYPE html>
<html>
<head>
    <title>503 Service Temporarily Unavailable</title>
</head>
<body>
    <h1>503 Service Temporarily Unavailable</h1>
    <p>The server is temporarily unable to service your request.</p>
</body>
</html>`,
			expected: "503 Service Temporarily Unavailable: The server is temporarily unable to service your request.",
		},
		{
			name:     "空内容",
			input:    "",
			expected: "",
		},
		{
			name:     "纯文本内容",
			input:    "This is a plain text error message",
			expected: "This is a plain text error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHTMLError(tt.input)
			if result != tt.expected {
				t.Errorf("extractHTMLError(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}

}

func TestExtractTagContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		tagName  string
		expected string
	}{
		{
			name:     "提取 title 标签",
			content:  `<html><head><title>Error Page</title></head></html>`,
			tagName:  "title",
			expected: "Error Page",
		},
		{
			name:     "提取 h1 标签",
			content:  `<h1>Bad Gateway</h1>`,
			tagName:  "h1",
			expected: "Bad Gateway",
		},
		{
			name:     "提取 p 标签",
			content:  `<p>Service is temporarily unavailable.</p>`,
			tagName:  "p",
			expected: "Service is temporarily unavailable.",
		},
		{
			name:     "标签不存在",
			content:  `<div>Some content</div>`,
			tagName:  "h2",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTagContent(tt.content, tt.tagName)
			if result != tt.expected {
				t.Errorf("extractTagContent(%q, %q) = %q; want %q", tt.content, tt.tagName, result, tt.expected)
			}
		})
	}
}

func TestIsHTMLContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "包含 HTML 标签",
			input:    `<html><body><h1>Error</h1></body></html>`,
			expected: true,
		},
		{
			name:     "纯文本内容",
			input:    "This is plain text",
			expected: false,
		},
		{
			name:     "包含多个 HTML 标签",
			input:    `<div><p>Error message</p></div>`,
			expected: true,
		},
		{
			name:     "JSON 内容",
			input:    `{"error": "message"}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHTMLContent(tt.input)
			if result != tt.expected {
				t.Errorf("isHTMLContent(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "清理多余空白",
			input:    "This   has   extra   spaces",
			expected: "This has extra spaces",
		},
		{
			name:     "清理换行符",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1 Line 2 Line 3",
		},
		{
			name:     "去除首尾空格",
			input:    "   text with spaces   ",
			expected: "text with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanWhitespace(tt.input)
			if result != tt.expected {
				t.Errorf("cleanWhitespace(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
