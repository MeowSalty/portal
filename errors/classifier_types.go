package errors

// ClassificationStage 定义分类阶段。
type ClassificationStage string

const (
	// ClassificationStageSource 来源分类阶段。
	ClassificationStageSource ClassificationStage = "source"
	// ClassificationStageResource 资源归属分类阶段。
	ClassificationStageResource ClassificationStage = "resource"
)

// ClassificationConfidence 定义分类置信度。
type ClassificationConfidence string

const (
	// ConfidenceHigh 高置信度。
	ConfidenceHigh ClassificationConfidence = "high"
	// ConfidenceMedium 中置信度。
	ConfidenceMedium ClassificationConfidence = "medium"
	// ConfidenceLow 低置信度。
	ConfidenceLow ClassificationConfidence = "low"
)

// ErrorResourceType 定义健康惩罚资源类型。
type ErrorResourceType string

const (
	// ErrorResourcePlatform 平台资源。
	ErrorResourcePlatform ErrorResourceType = "platform"
	// ErrorResourceAPIKey 密钥资源。
	ErrorResourceAPIKey ErrorResourceType = "api_key"
	// ErrorResourceModel 模型资源。
	ErrorResourceModel ErrorResourceType = "model"
)

// ClassifierInput 定义统一分类输入。
type ClassifierInput struct {
	// 基础字段
	Code                 ErrorCode
	Message              string
	HTTPStatus           int
	ErrorFrom            ErrorFromValue
	HTTPResponseReceived bool

	// 结构化上游字段
	ErrorType    string
	VendorCode   string
	ErrorParam   string
	ErrorMessage string

	// 非结构化文本
	ResponseBody string
	RawText      string
	CauseMessage string

	// 请求上下文
	Provider          string
	ModelName         string
	OriginalModelName string
	APIVariant        string
	Endpoint          string
}

// ClassificationDecision 定义单阶段分类决策。
type ClassificationDecision[T any] struct {
	Value        T
	Confidence   ClassificationConfidence
	MatchedRules []string
	Explain      string
}

// ClassificationResult 定义分类结果。
type ClassificationResult struct {
	Source   ClassificationDecision[ErrorFromValue]
	Resource ClassificationDecision[ErrorResourceType]

	MatchedRules []string
	Explain      string
	Signals      map[string]any
}

// ClassificationRule 定义单条分类规则。
type ClassificationRule struct {
	ID         string
	Enabled    bool
	Stage      ClassificationStage
	Priority   int
	Conditions RuleConditions
	Decision   RuleDecision
	Confidence ClassificationConfidence
	Reason     string

	// StrongEvidence 标记该规则为强证据规则。
	// 在资源冲突裁决中，强证据规则可以压过非强证据规则和兜底规则。
	StrongEvidence bool

	// Fallback 标记该规则为兜底规则。
	// 兜底规则在资源冲突裁决中不参与竞争，仅在没有其他规则命中时生效。
	Fallback bool
}

// RuleConditions 定义规则匹配条件。
//
// 说明：
// - 各字段之间为“与”关系。
// - 文本类字段中，切片内部为“任一命中”。
// - AllContains 为“全部命中”。
type RuleConditions struct {
	Codes []ErrorCode

	HTTPStatuses []int
	ErrorFrom    []ErrorFromValue

	ErrorTypes  []string
	VendorCodes []string

	MessageContains      []string
	ErrorMessageContains []string
	ResponseBodyContains []string
	CauseMessageContains []string
	RawTextContains      []string

	AnyContains []string
	AllContains []string

	HTTPResponseReceived *bool
}

// RuleDecision 定义规则决策结果。
type RuleDecision struct {
	Source   ErrorFromValue
	Resource ErrorResourceType
}
