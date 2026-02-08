package types

// FinishReason 表示响应生成结束原因。
type FinishReason = string

// FinishReason 枚举值。
const (
	FinishReasonUnspecified        FinishReason = "FINISH_REASON_UNSPECIFIED"
	FinishReasonStop               FinishReason = "STOP"
	FinishReasonMaxTokens          FinishReason = "MAX_TOKENS"
	FinishReasonSafety             FinishReason = "SAFETY"
	FinishReasonRecitation         FinishReason = "RECITATION"
	FinishReasonLanguage           FinishReason = "LANGUAGE"
	FinishReasonOther              FinishReason = "OTHER"
	FinishReasonBlocklist          FinishReason = "BLOCKLIST"
	FinishReasonProhibitedContent  FinishReason = "PROHIBITED_CONTENT"
	FinishReasonSPII               FinishReason = "SPII"
	FinishReasonMalformedFunction  FinishReason = "MALFORMED_FUNCTION_CALL"
	FinishReasonImageSafety        FinishReason = "IMAGE_SAFETY"
	FinishReasonImageProhibited    FinishReason = "IMAGE_PROHIBITED_CONTENT"
	FinishReasonImageOther         FinishReason = "IMAGE_OTHER"
	FinishReasonNoImage            FinishReason = "NO_IMAGE"
	FinishReasonImageRecitation    FinishReason = "IMAGE_RECITATION"
	FinishReasonUnexpectedToolCall FinishReason = "UNEXPECTED_TOOL_CALL"
	FinishReasonTooManyToolCalls   FinishReason = "TOO_MANY_TOOL_CALLS"
	FinishReasonMissingThoughtSig  FinishReason = "MISSING_THOUGHT_SIGNATURE"
)

// BlockReason 表示提示被阻止原因。
type BlockReason = string

// BlockReason 枚举值。
const (
	BlockReasonUnspecified BlockReason = "BLOCK_REASON_UNSPECIFIED"
	BlockReasonSafety      BlockReason = "SAFETY"
	BlockReasonOther       BlockReason = "OTHER"
	BlockReasonBlocklist   BlockReason = "BLOCKLIST"
	BlockReasonProhibited  BlockReason = "PROHIBITED_CONTENT"
	BlockReasonImageSafety BlockReason = "IMAGE_SAFETY"
)

// HarmCategory 表示安全分类。
type HarmCategory = string

// HarmCategory 枚举值。
const (
	HarmCategoryUnspecified      HarmCategory = "HARM_CATEGORY_UNSPECIFIED"
	HarmCategoryDerogatory       HarmCategory = "HARM_CATEGORY_DEROGATORY"
	HarmCategoryToxicity         HarmCategory = "HARM_CATEGORY_TOXICITY"
	HarmCategoryViolence         HarmCategory = "HARM_CATEGORY_VIOLENCE"
	HarmCategorySexual           HarmCategory = "HARM_CATEGORY_SEXUAL"
	HarmCategoryMedical          HarmCategory = "HARM_CATEGORY_MEDICAL"
	HarmCategoryDangerous        HarmCategory = "HARM_CATEGORY_DANGEROUS"
	HarmCategoryHarassment       HarmCategory = "HARM_CATEGORY_HARASSMENT"
	HarmCategoryHateSpeech       HarmCategory = "HARM_CATEGORY_HATE_SPEECH"
	HarmCategorySexuallyExplicit HarmCategory = "HARM_CATEGORY_SEXUALLY_EXPLICIT"
	HarmCategoryDangerousContent HarmCategory = "HARM_CATEGORY_DANGEROUS_CONTENT"
	HarmCategoryCivicIntegrity   HarmCategory = "HARM_CATEGORY_CIVIC_INTEGRITY"
)

// HarmProbability 表示安全概率。
type HarmProbability = string

// HarmProbability 枚举值。
const (
	HarmProbabilityUnspecified HarmProbability = "HARM_PROBABILITY_UNSPECIFIED"
	HarmProbabilityNegligible  HarmProbability = "NEGLIGIBLE"
	HarmProbabilityLow         HarmProbability = "LOW"
	HarmProbabilityMedium      HarmProbability = "MEDIUM"
	HarmProbabilityHigh        HarmProbability = "HIGH"
)

// Modality 表示模态枚举。
type Modality = string

// Modality 枚举值。
const (
	ModalityUnspecified Modality = "MODALITY_UNSPECIFIED"
	ModalityText        Modality = "TEXT"
	ModalityImage       Modality = "IMAGE"
	ModalityVideo       Modality = "VIDEO"
	ModalityAudio       Modality = "AUDIO"
	ModalityDocument    Modality = "DOCUMENT"
)

// URLRetrievalStatus 表示 URL 检索状态。
type URLRetrievalStatus = string

// URLRetrievalStatus 枚举值。
const (
	URLRetrievalStatusUnspecified URLRetrievalStatus = "URL_RETRIEVAL_STATUS_UNSPECIFIED"
	URLRetrievalStatusSuccess     URLRetrievalStatus = "URL_RETRIEVAL_STATUS_SUCCESS"
	URLRetrievalStatusError       URLRetrievalStatus = "URL_RETRIEVAL_STATUS_ERROR"
	URLRetrievalStatusPaywall     URLRetrievalStatus = "URL_RETRIEVAL_STATUS_PAYWALL"
	URLRetrievalStatusUnsafe      URLRetrievalStatus = "URL_RETRIEVAL_STATUS_UNSAFE"
)

// ModelStage 表示模型阶段。
type ModelStage = string

// ModelStage 枚举值。
const (
	ModelStageUnspecified          ModelStage = "MODEL_STAGE_UNSPECIFIED"
	ModelStageUnstableExperimental ModelStage = "UNSTABLE_EXPERIMENTAL"
	ModelStageExperimental         ModelStage = "EXPERIMENTAL"
	ModelStagePreview              ModelStage = "PREVIEW"
	ModelStageStable               ModelStage = "STABLE"
	ModelStageLegacy               ModelStage = "LEGACY"
	ModelStageDeprecated           ModelStage = "DEPRECATED"
	ModelStageRetired              ModelStage = "RETIRED"
)
