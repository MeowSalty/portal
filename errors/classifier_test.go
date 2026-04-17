package errors

import (
	"strings"
	"testing"
)

func TestClassifyError_上游服务端错误_504归类Platform(t *testing.T) {
	result := ClassifyError(ClassifierInput{
		Message:              "Gateway Timeout while waiting response",
		HTTPStatus:           504,
		HTTPResponseReceived: true,
	})

	if result.Source.Value != ErrorFromServer {
		t.Fatalf("source = %q, want %q", result.Source.Value, ErrorFromServer)
	}
	if result.Resource.Value != ErrorResourcePlatform {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourcePlatform)
	}
	if !containsString(result.Resource.MatchedRules, "resource-upstream-server-platform") {
		t.Fatalf("resource matched rules %v should contain resource-upstream-server-platform", result.Resource.MatchedRules)
	}
}

func TestClassifyError_上游服务端错误_502归类Platform(t *testing.T) {
	result := ClassifyError(ClassifierInput{
		Message:              "error code: 502",
		HTTPStatus:           502,
		HTTPResponseReceived: true,
	})

	if result.Resource.Value != ErrorResourcePlatform {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourcePlatform)
	}
	if !containsString(result.Resource.MatchedRules, "resource-upstream-server-platform") {
		t.Fatalf("resource matched rules %v should contain resource-upstream-server-platform", result.Resource.MatchedRules)
	}
}

func TestClassifyError_上游服务端错误_503归类Platform(t *testing.T) {
	result := ClassifyError(ClassifierInput{
		Message:              "system disk overloaded",
		HTTPStatus:           503,
		HTTPResponseReceived: true,
	})

	if result.Resource.Value != ErrorResourcePlatform {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourcePlatform)
	}
	if !containsString(result.Resource.MatchedRules, "resource-upstream-server-platform") {
		t.Fatalf("resource matched rules %v should contain resource-upstream-server-platform", result.Resource.MatchedRules)
	}
}

func TestClassifyError_上游服务端错误_522归类Platform(t *testing.T) {
	result := ClassifyError(ClassifierInput{
		Message:              "error code: 522",
		HTTPStatus:           522,
		HTTPResponseReceived: true,
	})

	if result.Resource.Value != ErrorResourcePlatform {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourcePlatform)
	}
	if !containsString(result.Resource.MatchedRules, "resource-upstream-server-platform") {
		t.Fatalf("resource matched rules %v should contain resource-upstream-server-platform", result.Resource.MatchedRules)
	}
}

func TestClassifyError_平台关键词_system归类Platform(t *testing.T) {
	result := ClassifyError(ClassifierInput{
		Message:              "system error occurred",
		HTTPStatus:           500,
		HTTPResponseReceived: true,
	})

	if result.Resource.Value != ErrorResourcePlatform {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourcePlatform)
	}
	if !containsString(result.Resource.MatchedRules, "resource-platform-strong-keywords") {
		t.Fatalf("resource matched rules %v should contain resource-platform-strong-keywords", result.Resource.MatchedRules)
	}
}

func TestClassifyError_优先级_认证证据覆盖模型词(t *testing.T) {
	result := ClassifyError(ClassifierInput{
		Code:                 ErrCodeAuthenticationFailed,
		Message:              "authentication failed for model gpt-4o",
		HTTPStatus:           401,
		HTTPResponseReceived: true,
	})

	if result.Resource.Value != ErrorResourceAPIKey {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourceAPIKey)
	}
	if result.Resource.Confidence != ConfidenceHigh {
		t.Fatalf("resource confidence = %q, want %q", result.Resource.Confidence, ConfidenceHigh)
	}
	if !containsString(result.Resource.MatchedRules, "resource-auth-status") {
		t.Fatalf("resource matched rules %v should contain resource-auth-status", result.Resource.MatchedRules)
	}
}

func TestClassifyError_冲突_无强证据规则时保守降级到模型(t *testing.T) {
	classifier := NewClassifier([]ClassificationRule{
		{
			ID:       "source-fallback",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 1,
			Conditions: RuleConditions{
				AnyContains: []string{"x"},
			},
			Decision:   RuleDecision{Source: ErrorFromServer},
			Confidence: ConfidenceLow,
			Reason:     "source fallback",
		},
		{
			ID:       "resource-key",
			Enabled:  true,
			Stage:    ClassificationStageResource,
			Priority: 100,
			Conditions: RuleConditions{
				AnyContains: []string{"x"},
			},
			Decision:   RuleDecision{Resource: ErrorResourceAPIKey},
			Confidence: ConfidenceHigh,
			Reason:     "resource key",
		},
		{
			ID:       "resource-platform",
			Enabled:  true,
			Stage:    ClassificationStageResource,
			Priority: 90,
			Conditions: RuleConditions{
				AnyContains: []string{"x"},
			},
			Decision:   RuleDecision{Resource: ErrorResourcePlatform},
			Confidence: ConfidenceHigh,
			Reason:     "resource platform",
		},
	})

	result := classifier.Classify(ClassifierInput{Message: "x"})

	if result.Resource.Value != ErrorResourceModel {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourceModel)
	}
	if result.Resource.Confidence != ConfidenceLow {
		t.Fatalf("resource confidence = %q, want %q", result.Resource.Confidence, ConfidenceLow)
	}
	if len(result.MatchedRules) == 0 {
		t.Fatalf("matched rules is empty")
	}
	if strings.TrimSpace(result.Explain) == "" {
		t.Fatalf("explain should not be empty")
	}
}

func TestClassifyError_冲突_强证据规则压过非强证据规则(t *testing.T) {
	classifier := NewClassifier([]ClassificationRule{
		{
			ID:       "source-fallback",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 1,
			Conditions: RuleConditions{
				AnyContains: []string{"x"},
			},
			Decision:   RuleDecision{Source: ErrorFromServer},
			Confidence: ConfidenceLow,
			Reason:     "source fallback",
		},
		{
			ID:             "resource-platform-strong",
			Enabled:        true,
			Stage:          ClassificationStageResource,
			Priority:       50,
			StrongEvidence: true,
			Conditions: RuleConditions{
				AnyContains: []string{"x"},
			},
			Decision:   RuleDecision{Resource: ErrorResourcePlatform},
			Confidence: ConfidenceMedium,
			Reason:     "强证据平台规则",
		},
		{
			ID:       "resource-model-weak",
			Enabled:  true,
			Stage:    ClassificationStageResource,
			Priority: 90,
			Conditions: RuleConditions{
				AnyContains: []string{"x"},
			},
			Decision:   RuleDecision{Resource: ErrorResourceModel},
			Confidence: ConfidenceHigh,
			Reason:     "弱证据模型规则",
		},
	})

	result := classifier.Classify(ClassifierInput{Message: "x"})

	if result.Resource.Value != ErrorResourcePlatform {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourcePlatform)
	}
	if result.Resource.Confidence != ConfidenceMedium {
		t.Fatalf("resource confidence = %q, want %q", result.Resource.Confidence, ConfidenceMedium)
	}
}

func TestClassifyError_冲突_多个强证据规则高分胜出(t *testing.T) {
	classifier := NewClassifier([]ClassificationRule{
		{
			ID:       "source-fallback",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 1,
			Conditions: RuleConditions{
				AnyContains: []string{"x"},
			},
			Decision:   RuleDecision{Source: ErrorFromServer},
			Confidence: ConfidenceLow,
			Reason:     "source fallback",
		},
		{
			ID:             "resource-key-strong",
			Enabled:        true,
			Stage:          ClassificationStageResource,
			Priority:       80,
			StrongEvidence: true,
			Conditions: RuleConditions{
				AnyContains: []string{"x"},
			},
			Decision:   RuleDecision{Resource: ErrorResourceAPIKey},
			Confidence: ConfidenceHigh,
			Reason:     "强证据密钥规则",
		},
		{
			ID:             "resource-platform-strong",
			Enabled:        true,
			Stage:          ClassificationStageResource,
			Priority:       90,
			StrongEvidence: true,
			Conditions: RuleConditions{
				AnyContains: []string{"x"},
			},
			Decision:   RuleDecision{Resource: ErrorResourcePlatform},
			Confidence: ConfidenceHigh,
			Reason:     "强证据平台规则",
		},
	})

	result := classifier.Classify(ClassifierInput{Message: "x"})

	// platform 规则优先级更高，应胜出
	if result.Resource.Value != ErrorResourcePlatform {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourcePlatform)
	}
}

func TestClassifyError_冲突_仅兜底规则命中时返回兜底(t *testing.T) {
	classifier := NewClassifier([]ClassificationRule{
		{
			ID:       "source-fallback",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 0,
			Conditions: RuleConditions{
				HTTPResponseReceived: boolPtr(false),
			},
			Decision:   RuleDecision{Source: ErrorFromGateway},
			Confidence: ConfidenceLow,
			Reason:     "source fallback",
		},
		{
			ID:         "resource-fallback",
			Enabled:    true,
			Stage:      ClassificationStageResource,
			Priority:   0,
			Fallback:   true,
			Conditions: RuleConditions{},
			Decision:   RuleDecision{Resource: ErrorResourceModel},
			Confidence: ConfidenceLow,
			Reason:     "兜底规则",
		},
	})

	result := classifier.Classify(ClassifierInput{})

	if result.Resource.Value != ErrorResourceModel {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourceModel)
	}
	if result.Resource.Confidence != ConfidenceLow {
		t.Fatalf("resource confidence = %q, want %q", result.Resource.Confidence, ConfidenceLow)
	}
}

func TestClassifyError_冲突_强证据平台规则压过兜底模型规则(t *testing.T) {
	classifier := NewClassifier([]ClassificationRule{
		{
			ID:       "source-fallback",
			Enabled:  true,
			Stage:    ClassificationStageSource,
			Priority: 1,
			Conditions: RuleConditions{
				AnyContains: []string{"x"},
			},
			Decision:   RuleDecision{Source: ErrorFromServer},
			Confidence: ConfidenceLow,
			Reason:     "source fallback",
		},
		{
			ID:             "resource-platform-strong",
			Enabled:        true,
			Stage:          ClassificationStageResource,
			Priority:       15,
			StrongEvidence: true,
			Conditions: RuleConditions{
				AnyContains: []string{"x"},
			},
			Decision:   RuleDecision{Resource: ErrorResourcePlatform},
			Confidence: ConfidenceMedium,
			Reason:     "强证据平台规则",
		},
		{
			ID:         "resource-model-fallback",
			Enabled:    true,
			Stage:      ClassificationStageResource,
			Priority:   0,
			Fallback:   true,
			Conditions: RuleConditions{},
			Decision:   RuleDecision{Resource: ErrorResourceModel},
			Confidence: ConfidenceLow,
			Reason:     "兜底模型规则",
		},
	})

	result := classifier.Classify(ClassifierInput{Message: "x"})

	if result.Resource.Value != ErrorResourcePlatform {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourcePlatform)
	}
}

func TestClassifyError_兜底_无信号默认模型承接(t *testing.T) {
	result := ClassifyError(ClassifierInput{})

	if result.Source.Value != ErrorFromGateway {
		t.Fatalf("source = %q, want %q", result.Source.Value, ErrorFromGateway)
	}
	if result.Resource.Value != ErrorResourceModel {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourceModel)
	}

	if result.Explain == "" {
		t.Fatalf("explain should not be empty")
	}
	if !strings.Contains(result.Explain, "source=") || !strings.Contains(result.Explain, "resource=") {
		t.Fatalf("explain format invalid: %q", result.Explain)
	}
	if result.Signals == nil {
		t.Fatalf("signals should not be nil")
	}
	if len(result.MatchedRules) == 0 {
		t.Fatalf("matched rules is empty")
	}
}

func TestClassifyError_Gateway网络不可用_归类Platform(t *testing.T) {
	result := ClassifyError(ClassifierInput{
		Code:      ErrCodeUnavailable,
		Message:   "lookup codex.sakurapy.de on 10.89.13.1:53: no such host",
		ErrorFrom: ErrorFromGateway,
	})

	if result.Source.Value != ErrorFromGateway {
		t.Fatalf("source = %q, want %q", result.Source.Value, ErrorFromGateway)
	}
	if result.Resource.Value != ErrorResourcePlatform {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourcePlatform)
	}
	if !containsString(result.Resource.MatchedRules, "resource-gateway-network-platform") {
		t.Fatalf("resource matched rules %v should contain resource-gateway-network-platform", result.Resource.MatchedRules)
	}
}

func TestClassifyError_Gateway网络不可用_非Gateway来源不命中(t *testing.T) {
	result := ClassifyError(ClassifierInput{
		Code:      ErrCodeUnavailable,
		Message:   "lookup codex.sakurapy.de on 10.89.13.1:53: no such host",
		ErrorFrom: ErrorFromUpstream,
	})

	// upstream + UNAVAILABLE 应走 resource-upstream-default-model 而非新规则
	if result.Resource.Value != ErrorResourceModel {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourceModel)
	}
	if containsString(result.Resource.MatchedRules, "resource-gateway-network-platform") {
		t.Fatalf("resource matched rules should NOT contain resource-gateway-network-platform for upstream source")
	}
}

func containsString(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
