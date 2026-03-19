package adapter

import (
	"testing"

	portalErrors "github.com/MeowSalty/portal/errors"
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

func TestClassifyErrorFrom_ByTypeCodeMessageRules(t *testing.T) {
	a := &Adapter{}

	tests := []struct {
		name string
		data map[string]interface{}
		want portalErrors.ErrorFromValue
	}{
		{
			name: "按 type 命中 upstream",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"type": "openai_error",
				},
			},
			want: portalErrors.ErrorFromUpstream,
		},
		{
			name: "按 code 命中 upstream",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"code": "do_request_failed",
				},
			},
			want: portalErrors.ErrorFromUpstream,
		},
		{
			name: "按 message 命中 upstream",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"message": "failed to retrieve proxy group",
				},
			},
			want: portalErrors.ErrorFromUpstream,
		},
		{
			name: "按 type 命中 server",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"type": "one_hub_error",
				},
			},
			want: portalErrors.ErrorFromServer,
		},
		{
			name: "按 code 命中 server",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"code": "model_not_found",
				},
			},
			want: portalErrors.ErrorFromServer,
		},
		{
			name: "按 message 命中 server",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"message": "用户额度不足",
				},
			},
			want: portalErrors.ErrorFromServer,
		},
		{
			name: "优先级 upstream 高于 server",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"type":    "openai_error",
					"code":    "model_not_found",
					"message": "用户额度不足",
				},
			},
			want: portalErrors.ErrorFromUpstream,
		},
		{
			name: "未命中规则但有 type/code → 智能兜底 server",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"type":    "unknown",
					"code":    "unknown",
					"message": "unknown",
				},
			},
			want: portalErrors.ErrorFromServer,
		},
		{
			name: "智能兜底：仅有 type → server",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"type": "totally_new_server_error",
				},
			},
			want: portalErrors.ErrorFromServer,
		},
		{
			name: "智能兜底：仅有 code → server",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"code": "brand_new_server_code",
				},
			},
			want: portalErrors.ErrorFromServer,
		},
		{
			name: "纯 message 无 type/code 且未命中规则 → server（hasBody=true 兜底）",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"message": "a plain unknown message",
				},
			},
			want: portalErrors.ErrorFromServer,
		},
		{
			name: "空 error 对象 → server（hasBody=true 兜底）",
			data: map[string]interface{}{
				"error": map[string]interface{}{},
			},
			want: portalErrors.ErrorFromServer,
		},
		{
			name: "无 error 字段 → server（hasBody=true 兜底）",
			data: map[string]interface{}{
				"message": "no structured error object",
			},
			want: portalErrors.ErrorFromServer,
		},
		{
			name: "用户示例：new_api_error + 额度用尽消息 → server",
			data: map[string]interface{}{
				"error": map[string]interface{}{
					"type":    "new_api_error",
					"message": "用户额度不足，请充值",
				},
			},
			want: portalErrors.ErrorFromServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := a.classifyErrorFrom(tt.data)
			if got != tt.want {
				t.Fatalf("classifyErrorFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleParseError_SetsErrorFromGateway(t *testing.T) {
	a := &Adapter{}
	innerErr := portalErrors.New(portalErrors.ErrCodeInvalidArgument, "无效参数")

	err := a.handleParseError("响应解析错误", innerErr, []byte("bad body"))

	if got := portalErrors.GetCode(err); got != portalErrors.ErrCodeInternal {
		t.Fatalf("GetCode() = %s, want %s", got, portalErrors.ErrCodeInternal)
	}

	if from := portalErrors.GetErrorFrom(err); from != portalErrors.ErrorFromGateway {
		t.Fatalf("GetErrorFrom() = %q, want %q", from, portalErrors.ErrorFromGateway)
	}

	ctx := portalErrors.GetContext(err)
	if ctx == nil {
		t.Fatalf("GetContext() 返回 nil")
	}

	if got, ok := ctx["operation"].(string); !ok || got != "响应解析错误" {
		t.Fatalf("operation 上下文不符合预期：%+v", ctx["operation"])
	}

	if got, ok := ctx["response_body"].(string); !ok || got != "bad body" {
		t.Fatalf("response_body 上下文不符合预期：%+v", ctx["response_body"])
	}
}

func TestHandleHTTPError_JSONBodyWithNonJSONContentType_ClassifiesAsServer(t *testing.T) {
	a := &Adapter{}

	err := a.handleHTTPError("API 返回错误状态码", 500, []byte(`{"error":{"message":"auth_unavailable: no auth available","type":"server_error","code":"internal_server_error"}}`))

	if from := portalErrors.GetErrorFrom(err); from != portalErrors.ErrorFromServer {
		t.Fatalf("GetErrorFrom() = %q, want %q", from, portalErrors.ErrorFromServer)
	}

	if got := portalErrors.GetCode(err); got != portalErrors.ErrCodeRequestFailed {
		t.Fatalf("GetCode() = %s, want %s", got, portalErrors.ErrCodeRequestFailed)
	}
}

func TestHandleHTTPError_JSONBodyWithNonJSONContentType_ClassifiesAsUpstream(t *testing.T) {
	a := &Adapter{}

	err := a.handleHTTPError("API 返回错误状态码", 403, []byte(`{"error":{"message":"request_error","type":"bad_response_status_code","code":"bad_response_status_code"}}`))

	if from := portalErrors.GetErrorFrom(err); from != portalErrors.ErrorFromUpstream {
		t.Fatalf("GetErrorFrom() = %q, want %q", from, portalErrors.ErrorFromUpstream)
	}

	if got := portalErrors.GetCode(err); got != portalErrors.ErrCodeRequestFailed {
		t.Fatalf("GetCode() = %s, want %s", got, portalErrors.ErrCodeRequestFailed)
	}
}

func TestHandleHTTPError_PlainTextBody_ClassifiesAsServer(t *testing.T) {
	a := &Adapter{}

	err := a.handleHTTPError("API 返回错误状态码", 502, []byte("simple backend failure"))

	if from := portalErrors.GetErrorFrom(err); from != portalErrors.ErrorFromServer {
		t.Fatalf("GetErrorFrom() = %q, want %q", from, portalErrors.ErrorFromServer)
	}

	ctx := portalErrors.GetContext(err)
	if got, ok := ctx["response_body"].(string); !ok || got != "simple backend failure" {
		t.Fatalf("response_body 上下文不符合预期：%+v", ctx["response_body"])
	}
}

func TestHandleHTTPError_HTMLBody_ClassifiesAsServerAndExtractReadableText(t *testing.T) {
	a := &Adapter{}
	body := []byte(`<html><head><title>502 Bad Gateway</title></head><body><h1>502 Bad Gateway</h1><p>The web server reported a bad gateway error.</p></body></html>`)

	err := a.handleHTTPError("API 返回错误状态码", 502, body)

	if from := portalErrors.GetErrorFrom(err); from != portalErrors.ErrorFromServer {
		t.Fatalf("GetErrorFrom() = %q, want %q", from, portalErrors.ErrorFromServer)
	}

	ctx := portalErrors.GetContext(err)
	if got, ok := ctx["response_body"].(string); !ok || got != "502 Bad Gateway: The web server reported a bad gateway error." {
		t.Fatalf("response_body 上下文不符合预期：%+v", ctx["response_body"])
	}
}

func TestHandleHTTPError_NonJSONTextWithUpstreamKeyword_ClassifiesAsUpstream(t *testing.T) {
	a := &Adapter{}

	err := a.handleHTTPError("API 返回错误状态码", 502, []byte("upstream timeout while contacting model provider"))

	if from := portalErrors.GetErrorFrom(err); from != portalErrors.ErrorFromUpstream {
		t.Fatalf("GetErrorFrom() = %q, want %q", from, portalErrors.ErrorFromUpstream)
	}
}

func TestHandleHTTPError_InvalidJSONButNonEmptyBody_ClassifiesAsServer(t *testing.T) {
	a := &Adapter{}

	err := a.handleHTTPError("API 返回错误状态码", 500, []byte(`{"error":`))

	if from := portalErrors.GetErrorFrom(err); from != portalErrors.ErrorFromServer {
		t.Fatalf("GetErrorFrom() = %q, want %q", from, portalErrors.ErrorFromServer)
	}

	if got := portalErrors.GetCode(err); got != portalErrors.ErrCodeRequestFailed {
		t.Fatalf("GetCode() = %s, want %s", got, portalErrors.ErrCodeRequestFailed)
	}
}

func TestHandleHTTPError_EmptyBody_ClassifiesAsGateway(t *testing.T) {
	a := &Adapter{}

	err := a.handleHTTPError("API 返回错误状态码", 504, nil)

	if from := portalErrors.GetErrorFrom(err); from != portalErrors.ErrorFromGateway {
		t.Fatalf("GetErrorFrom() = %q, want %q", from, portalErrors.ErrorFromGateway)
	}

	if got := portalErrors.GetCode(err); got != portalErrors.ErrCodeDeadlineExceeded {
		t.Fatalf("GetCode() = %s, want %s", got, portalErrors.ErrCodeDeadlineExceeded)
	}
}
