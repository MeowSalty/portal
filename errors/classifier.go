package errors

import (
	"fmt"
	"sort"
	"strings"
)

// Classifier 定义统一错误分类器。
type Classifier struct {
	rules []ClassificationRule
}

// NewClassifier 创建分类器。
func NewClassifier(rules []ClassificationRule) *Classifier {
	cloned := make([]ClassificationRule, len(rules))
	copy(cloned, rules)

	sort.SliceStable(cloned, func(i, j int) bool {
		if cloned[i].Priority != cloned[j].Priority {
			return cloned[i].Priority > cloned[j].Priority
		}
		return confidenceWeight(cloned[i].Confidence) > confidenceWeight(cloned[j].Confidence)
	})

	return &Classifier{rules: cloned}
}

// DefaultClassifier 创建带默认规则的分类器。
func DefaultClassifier() *Classifier {
	return NewClassifier(DefaultClassificationRules())
}

// ClassifyError 使用默认分类器进行分类。
func ClassifyError(input ClassifierInput) ClassificationResult {
	return DefaultClassifier().Classify(input)
}

// Classify 执行分类。
func (c *Classifier) Classify(input ClassifierInput) ClassificationResult {
	sourceDecision := c.classifySource(input)
	resourceDecision := c.classifyResource(input)

	matched := make([]string, 0, len(sourceDecision.MatchedRules)+len(resourceDecision.MatchedRules))
	matched = append(matched, sourceDecision.MatchedRules...)
	matched = append(matched, resourceDecision.MatchedRules...)

	explain := strings.Join([]string{
		fmt.Sprintf("source=%s(%s)", sourceDecision.Value, sourceDecision.Confidence),
		fmt.Sprintf("resource=%s(%s)", resourceDecision.Value, resourceDecision.Confidence),
	}, "; ")

	return ClassificationResult{
		Source:       sourceDecision,
		Resource:     resourceDecision,
		MatchedRules: matched,
		Explain:      explain,
		Signals:      buildSignals(input),
	}
}

func (c *Classifier) classifySource(input ClassifierInput) ClassificationDecision[ErrorFromValue] {
	matches := c.matchedRules(ClassificationStageSource, input)
	if len(matches) == 0 {
		return ClassificationDecision[ErrorFromValue]{
			Value:      ErrorFromGateway,
			Confidence: ConfidenceLow,
			Explain:    "无来源命中规则，兜底 gateway",
		}
	}

	best := matches[0]
	return ClassificationDecision[ErrorFromValue]{
		Value:        best.Decision.Source,
		Confidence:   best.Confidence,
		MatchedRules: collectRuleIDs(matches),
		Explain:      best.Reason,
	}
}

func (c *Classifier) classifyResource(input ClassifierInput) ClassificationDecision[ErrorResourceType] {
	matches := c.matchedRules(ClassificationStageResource, input)
	if len(matches) == 0 {
		return ClassificationDecision[ErrorResourceType]{
			Value:      ErrorResourceModel,
			Confidence: ConfidenceLow,
			Explain:    "无资源命中规则，兜底 model",
		}
	}

	value, confidence, reason := resolveConservativeResource(matches, input)
	return ClassificationDecision[ErrorResourceType]{
		Value:        value,
		Confidence:   confidence,
		MatchedRules: collectRuleIDs(matches),
		Explain:      reason,
	}
}

func (c *Classifier) matchedRules(stage ClassificationStage, input ClassifierInput) []ClassificationRule {
	matched := make([]ClassificationRule, 0)
	for _, rule := range c.rules {
		if !rule.Enabled || rule.Stage != stage {
			continue
		}
		if matchRuleConditions(rule.Conditions, input) {
			matched = append(matched, rule)
		}
	}
	return matched
}

func resolveConservativeResource(matches []ClassificationRule, input ClassifierInput) (ErrorResourceType, ClassificationConfidence, string) {
	best := matches[0]

	hasKey := false
	hasPlatform := false
	hasModel := false
	for _, m := range matches {
		switch m.Decision.Resource {
		case ErrorResourceAPIKey:
			hasKey = true
		case ErrorResourcePlatform:
			hasPlatform = true
		case ErrorResourceModel:
			hasModel = true
		}
	}

	if hasKey && hasPlatform && !hasStrongAuthEvidence(input) {
		return ErrorResourceModel, ConfidenceLow, "密钥与平台资源规则冲突且缺少强认证证据，保守降级 model"
	}
	if hasPlatform && hasModel && !hasStrongPlatformEvidence(input) {
		return ErrorResourceModel, ConfidenceLow, "平台与模型资源规则冲突且平台证据不足，保守选择 model"
	}
	if hasKey && hasModel && !hasStrongAuthEvidence(input) {
		return ErrorResourceModel, ConfidenceLow, "密钥与模型资源规则冲突且认证证据不足，保守选择 model"
	}

	return best.Decision.Resource, best.Confidence, best.Reason
}

func matchRuleConditions(cond RuleConditions, input ClassifierInput) bool {
	if len(cond.Codes) > 0 && !containsErrorCode(cond.Codes, input.Code) {
		return false
	}
	if len(cond.HTTPStatuses) > 0 && !containsInt(cond.HTTPStatuses, input.HTTPStatus) {
		return false
	}
	if len(cond.ErrorFrom) > 0 && !containsErrorFrom(cond.ErrorFrom, input.ErrorFrom) {
		return false
	}
	if len(cond.ErrorTypes) > 0 && !containsFold(cond.ErrorTypes, input.ErrorType) {
		return false
	}
	if len(cond.VendorCodes) > 0 && !containsFold(cond.VendorCodes, input.VendorCode) {
		return false
	}
	if cond.HTTPResponseReceived != nil && input.HTTPResponseReceived != *cond.HTTPResponseReceived {
		return false
	}

	if len(cond.MessageContains) > 0 && !containsAnyFold(input.Message, cond.MessageContains) {
		return false
	}
	if len(cond.ErrorMessageContains) > 0 && !containsAnyFold(input.ErrorMessage, cond.ErrorMessageContains) {
		return false
	}
	if len(cond.ResponseBodyContains) > 0 && !containsAnyFold(input.ResponseBody, cond.ResponseBodyContains) {
		return false
	}
	if len(cond.CauseMessageContains) > 0 && !containsAnyFold(input.CauseMessage, cond.CauseMessageContains) {
		return false
	}
	if len(cond.RawTextContains) > 0 && !containsAnyFold(input.RawText, cond.RawTextContains) {
		return false
	}

	blob := combinedText(input)
	if len(cond.AnyContains) > 0 && !containsAnyFold(blob, cond.AnyContains) {
		return false
	}
	if len(cond.AllContains) > 0 && !containsAllFold(blob, cond.AllContains) {
		return false
	}

	return true
}

func hasStrongAuthEvidence(input ClassifierInput) bool {
	if input.HTTPStatus == 401 || input.HTTPStatus == 403 {
		return true
	}
	if input.Code == ErrCodeAuthenticationFailed || input.Code == ErrCodePermissionDenied {
		return true
	}
	return containsAnyFold(combinedText(input), []string{"api key", "invalid key", "authentication", "permission", "unauthorized", "token", "鉴权", "认证", "权限", "密钥"})
}

func hasStrongPlatformEvidence(input ClassifierInput) bool {
	// 上游服务端状态码属于强平台证据
	switch input.HTTPStatus {
	case 502, 503, 504, 521, 522, 524:
		return true
	}
	return containsAnyFold(combinedText(input), []string{
		"渠道", "路由", "节点", "平台内部",
		"channel unavailable", "route", "proxy", "backend", "platform",
		"system", "service", "gateway", "overloaded", "unavailable",
	})
}

func combinedText(input ClassifierInput) string {
	parts := []string{
		input.Message,
		input.ErrorMessage,
		input.ResponseBody,
		input.CauseMessage,
		input.RawText,
		input.ErrorType,
		input.VendorCode,
		input.ErrorParam,
		input.Provider,
		input.ModelName,
		input.OriginalModelName,
		input.APIVariant,
		input.Endpoint,
	}
	return strings.ToLower(strings.Join(parts, "\n"))
}

func buildSignals(input ClassifierInput) map[string]any {
	return map[string]any{
		"code":                   input.Code,
		"http_status":            input.HTTPStatus,
		"error_from":             input.ErrorFrom,
		"http_response_received": input.HTTPResponseReceived,
		"provider":               input.Provider,
		"model_name":             input.ModelName,
		"original_model_name":    input.OriginalModelName,
		"api_variant":            input.APIVariant,
		"endpoint":               input.Endpoint,
	}
}

func collectRuleIDs(rules []ClassificationRule) []string {
	ids := make([]string, 0, len(rules))
	for _, rule := range rules {
		ids = append(ids, rule.ID)
	}
	return ids
}

func confidenceWeight(c ClassificationConfidence) int {
	switch c {
	case ConfidenceHigh:
		return 3
	case ConfidenceMedium:
		return 2
	default:
		return 1
	}
}

func containsErrorCode(values []ErrorCode, target ErrorCode) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func containsErrorFrom(values []ErrorFromValue, target ErrorFromValue) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func containsInt(values []int, target int) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func containsFold(values []string, target string) bool {
	t := strings.ToLower(target)
	for _, v := range values {
		if strings.ToLower(v) == t {
			return true
		}
	}
	return false
}

func containsAnyFold(text string, keywords []string) bool {
	t := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(t, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func containsAllFold(text string, keywords []string) bool {
	t := strings.ToLower(text)
	for _, kw := range keywords {
		if !strings.Contains(t, strings.ToLower(kw)) {
			return false
		}
	}
	return true
}
