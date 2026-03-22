package request

import (
	"testing"

	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
)

func TestExtractNativeHeaders_OpenAIChat(t *testing.T) {
	req := &openaiChat.Request{
		Headers: map[string]string{
			"User-Agent": "portal-test",
			"Referer":    "https://example.com",
		},
	}

	headers := extractNativeHeaders(req)
	if len(headers) != 2 {
		t.Fatalf("提取头部数量不正确，actual=%d", len(headers))
	}
	if got := headers["User-Agent"]; got != "portal-test" {
		t.Fatalf("User-Agent 不正确，actual=%q", got)
	}
	if got := headers["Referer"]; got != "https://example.com" {
		t.Fatalf("Referer 不正确，actual=%q", got)
	}
}

func TestExtractNativeHeaders_OpenAIResponses(t *testing.T) {
	req := &openaiResponses.Request{
		Headers: map[string]string{
			"OpenAI-Organization": "org_test",
		},
	}

	headers := extractNativeHeaders(req)
	if len(headers) != 1 {
		t.Fatalf("提取头部数量不正确，actual=%d", len(headers))
	}
	if got := headers["OpenAI-Organization"]; got != "org_test" {
		t.Fatalf("OpenAI-Organization 不正确，actual=%q", got)
	}
}

func TestExtractNativeHeaders_Anthropic(t *testing.T) {
	req := &anthropicTypes.Request{
		Headers: map[string]string{
			"X-Trace-ID": "trace-anthropic",
		},
	}

	headers := extractNativeHeaders(req)
	if len(headers) != 1 {
		t.Fatalf("提取头部数量不正确，actual=%d", len(headers))
	}
	if got := headers["X-Trace-ID"]; got != "trace-anthropic" {
		t.Fatalf("X-Trace-ID 不正确，actual=%q", got)
	}
}

func TestExtractNativeHeaders_Gemini(t *testing.T) {
	req := &geminiTypes.Request{
		Headers: map[string]string{
			"X-Client": "gemini-client",
		},
	}

	headers := extractNativeHeaders(req)
	if len(headers) != 1 {
		t.Fatalf("提取头部数量不正确，actual=%d", len(headers))
	}
	if got := headers["X-Client"]; got != "gemini-client" {
		t.Fatalf("X-Client 不正确，actual=%q", got)
	}
}

func TestExtractNativeHeaders_CloneAndNilCases(t *testing.T) {
	req := &openaiChat.Request{
		Headers: map[string]string{
			"User-Agent": "ua-1",
		},
	}

	headers := extractNativeHeaders(req)
	if headers == nil {
		t.Fatalf("提取结果不应为 nil")
	}

	headers["User-Agent"] = "ua-2"
	if got := req.Headers["User-Agent"]; got != "ua-1" {
		t.Fatalf("应返回拷贝，原始请求头不应被修改，actual=%q", got)
	}

	if got := extractNativeHeaders(nil); got != nil {
		t.Fatalf("nil payload 应返回 nil")
	}

	if got := extractNativeHeaders("not-struct"); got != nil {
		t.Fatalf("非结构体 payload 应返回 nil")
	}

	type noHeaderPayload struct {
		Name string
	}
	if got := extractNativeHeaders(noHeaderPayload{Name: "x"}); got != nil {
		t.Fatalf("不含 Headers 字段的结构体应返回 nil")
	}
}
