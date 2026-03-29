package errors

import (
	"strings"
	"testing"
)

func TestClassifyError_保守判定_GatewayTimeout不升级平台或密钥(t *testing.T) {
	result := ClassifyError(ClassifierInput{
		Message:              "Gateway Timeout while waiting response",
		HTTPStatus:           504,
		HTTPResponseReceived: true,
	})

	if result.Source.Value != ErrorFromServer {
		t.Fatalf("source = %q, want %q", result.Source.Value, ErrorFromServer)
	}
	if result.Resource.Value != ErrorResourceModel {
		t.Fatalf("resource = %q, want %q", result.Resource.Value, ErrorResourceModel)
	}
	if result.Resource.Confidence != ConfidenceLow {
		t.Fatalf("resource confidence = %q, want %q", result.Resource.Confidence, ConfidenceLow)
	}
	if len(result.MatchedRules) == 0 {
		t.Fatalf("matched rules is empty")
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

func TestClassifyError_冲突_构造规则冲突时保守降级到模型(t *testing.T) {
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

func containsString(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
