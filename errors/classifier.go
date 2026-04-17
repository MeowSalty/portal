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

	value, confidence, reason := resolveResourceByScore(matches)
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

// resolveResourceByScore 基于规则评分的资源冲突裁决。
//
// 评分模型：
//   - 每条规则得分 = Priority*10 + confidenceWeight(Confidence)
//   - StrongEvidence 规则额外 +10000
//   - Fallback 规则不参与竞争，仅在没有其他规则命中时生效
//
// 裁决策略：
//  1. 分离兜底规则与非兜底规则
//  2. 若无非兜底规则命中，使用兜底规则
//  3. 若仅一种资源类型有非兜底规则，直接返回
//  4. 若多种资源类型冲突：
//     a. 仅一种有强证据 → 强证据胜出
//     b. 多种有强证据 → 最高分胜出
//     c. 均无强证据 → 保守降级到 model
func resolveResourceByScore(matches []ClassificationRule) (ErrorResourceType, ClassificationConfidence, string) {
	// 分离兜底规则与非兜底规则
	var nonFallback, fallbackRules []ClassificationRule
	for _, m := range matches {
		if m.Fallback {
			fallbackRules = append(fallbackRules, m)
		} else {
			nonFallback = append(nonFallback, m)
		}
	}

	// 没有非兜底规则命中，使用兜底规则
	if len(nonFallback) == 0 {
		if len(fallbackRules) > 0 {
			return fallbackRules[0].Decision.Resource, fallbackRules[0].Confidence, fallbackRules[0].Reason
		}
		return ErrorResourceModel, ConfidenceLow, "无资源命中规则，兜底 model"
	}

	// 按资源类型分组，计算每组的强证据标记与最高分
	type resourceGroup struct {
		resource  ErrorResourceType
		hasStrong bool
		bestRule  ClassificationRule
		bestScore int
	}

	groups := make(map[ErrorResourceType]*resourceGroup)
	var resourceOrder []ErrorResourceType

	for _, m := range nonFallback {
		r := m.Decision.Resource
		if _, exists := groups[r]; !exists {
			groups[r] = &resourceGroup{resource: r}
			resourceOrder = append(resourceOrder, r)
		}
		score := ruleScore(m)
		if m.StrongEvidence {
			groups[r].hasStrong = true
		}
		if score > groups[r].bestScore {
			groups[r].bestScore = score
			groups[r].bestRule = m
		}
	}

	// 只有一种资源类型，直接返回
	if len(groups) == 1 {
		g := groups[resourceOrder[0]]
		return g.bestRule.Decision.Resource, g.bestRule.Confidence, g.bestRule.Reason
	}

	// 多种资源类型冲突，检查强证据
	var strongGroups []ErrorResourceType
	for _, r := range resourceOrder {
		if groups[r].hasStrong {
			strongGroups = append(strongGroups, r)
		}
	}

	// 仅一种资源类型有强证据，强证据胜出
	if len(strongGroups) == 1 {
		g := groups[strongGroups[0]]
		return g.bestRule.Decision.Resource, g.bestRule.Confidence, g.bestRule.Reason
	}

	// 多种资源类型都有强证据，按分数选最高
	if len(strongGroups) > 1 {
		best := groups[strongGroups[0]]
		for _, r := range strongGroups[1:] {
			if groups[r].bestScore > best.bestScore {
				best = groups[r]
			}
		}
		return best.bestRule.Decision.Resource, best.bestRule.Confidence, best.bestRule.Reason
	}

	// 均无强证据规则，多种资源类型冲突，保守降级到 model
	return ErrorResourceModel, ConfidenceLow, "多种资源类型冲突且缺少强证据规则，保守降级 model"
}

// ruleScore 计算单条规则的评分。
func ruleScore(m ClassificationRule) int {
	score := m.Priority*10 + confidenceWeight(m.Confidence)
	if m.StrongEvidence {
		score += 10000
	}
	return score
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
