package shared

// ImageDetail 表示图像细节级别
type ImageDetail = string

const (
	ImageDetailAuto ImageDetail = "auto"
	ImageDetailLow  ImageDetail = "low"
	ImageDetailHigh ImageDetail = "high"
)

// ReasoningEffort 表示推理努力级别
type ReasoningEffort = string

const (
	ReasoningEffortNone    ReasoningEffort = "none"
	ReasoningEffortMinimal ReasoningEffort = "minimal"
	ReasoningEffortLow     ReasoningEffort = "low"
	ReasoningEffortMedium  ReasoningEffort = "medium"
	ReasoningEffortHigh    ReasoningEffort = "high"
	ReasoningEffortXHigh   ReasoningEffort = "xhigh"
)

// VerbosityLevel 表示输出详细度级别
type VerbosityLevel = string

const (
	VerbosityLow    VerbosityLevel = "low"
	VerbosityMedium VerbosityLevel = "medium"
	VerbosityHigh   VerbosityLevel = "high"
)

// ServiceTier 表示服务层级
type ServiceTier = string

const (
	ServiceTierAuto     ServiceTier = "auto"
	ServiceTierDefault  ServiceTier = "default"
	ServiceTierFlex     ServiceTier = "flex"
	ServiceTierScale    ServiceTier = "scale"
	ServiceTierPriority ServiceTier = "priority"
)
