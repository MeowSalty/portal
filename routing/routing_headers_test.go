package routing

import "testing"

func TestMergeCustomHeaders_EmptyReturnsNil(t *testing.T) {
	merged := mergeCustomHeaders(nil, nil)
	if merged != nil {
		t.Fatalf("合并空头部时应返回 nil")
	}
}

func TestMergeCustomHeaders_EndpointOverride(t *testing.T) {
	platform := map[string]string{
		"X-Common":   "platform",
		"X-Platform": "p",
	}
	endpoint := map[string]string{
		"X-Common":   "endpoint",
		"X-Endpoint": "e",
	}

	merged := mergeCustomHeaders(platform, endpoint)
	if merged == nil {
		t.Fatalf("合并结果不应为 nil")
	}

	if got := merged["X-Common"]; got != "endpoint" {
		t.Fatalf("端点同名头部应覆盖平台头部，actual=%q", got)
	}
	if got := merged["X-Platform"]; got != "p" {
		t.Fatalf("平台头部应保留，actual=%q", got)
	}
	if got := merged["X-Endpoint"]; got != "e" {
		t.Fatalf("端点头部应保留，actual=%q", got)
	}
}
