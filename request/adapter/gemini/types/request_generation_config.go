package types

// GenerationConfig 表示生成配置
type GenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`     // 温度
	TopP            *float64 `json:"topP,omitempty"`            // Top-p 采样
	TopK            *int     `json:"topK,omitempty"`            // Top-k 采样
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"` // 最大输出 token 数
	StopSequences   []string `json:"stopSequences,omitempty"`   // 停止序列
	CandidateCount  *int     `json:"candidateCount,omitempty"`  // 候选数量
	Seed            *int     `json:"seed,omitempty"`            // 随机种子
	// 响应 MIME 类型
	ResponseMimeType *string `json:"responseMimeType,omitempty"`
	// 响应 schema
	ResponseSchema *Schema `json:"responseSchema,omitempty"`
	// JSON Schema（兼容 JSON Schema 格式）
	ResponseJSONSchemaRaw interface{} `json:"_responseJsonSchema,omitempty"`
	// 内部字段
	ResponseJsonSchema interface{} `json:"responseJsonSchema,omitempty"`
	// 出现惩罚
	PresencePenalty *float64 `json:"presencePenalty,omitempty"`
	// 频率惩罚
	FrequencyPenalty *float64 `json:"frequencyPenalty,omitempty"`
	// 是否返回 logprobs
	ResponseLogprobs *bool `json:"responseLogprobs,omitempty"`
	// logprobs 数量
	Logprobs *int `json:"logprobs,omitempty"`
	// 增强公民答案
	EnableEnhancedCivicAnswers *bool `json:"enableEnhancedCivicAnswers,omitempty"`
	// 响应模态
	ResponseModalities []string `json:"responseModalities,omitempty"`
	// 语音配置
	SpeechConfig *SpeechConfig `json:"speechConfig,omitempty"`
	// 思考配置
	ThinkingConfig *ThinkingConfig `json:"thinkingConfig,omitempty"`
	// 图像配置
	ImageConfig *ImageConfig `json:"imageConfig,omitempty"`
	// 媒体分辨率
	MediaResolution *string `json:"mediaResolution,omitempty"`
}

// SpeechConfig 表示语音配置
type SpeechConfig struct {
	VoiceConfig             *VoiceConfig             `json:"voiceConfig,omitempty"`
	MultiSpeakerVoiceConfig *MultiSpeakerVoiceConfig `json:"multiSpeakerVoiceConfig,omitempty"`
	LanguageCode            *string                  `json:"languageCode,omitempty"`
}

// VoiceConfig 表示语音配置
type VoiceConfig struct {
	PrebuiltVoiceConfig *PrebuiltVoiceConfig `json:"prebuiltVoiceConfig,omitempty"`
}

// PrebuiltVoiceConfig 表示预置语音配置
type PrebuiltVoiceConfig struct {
	VoiceName string `json:"voiceName"`
}

// MultiSpeakerVoiceConfig 表示多说话人配置
type MultiSpeakerVoiceConfig struct {
	SpeakerVoiceConfigs []SpeakerVoiceConfig `json:"speakerVoiceConfigs"`
}

// SpeakerVoiceConfig 表示单个说话人配置
type SpeakerVoiceConfig struct {
	Speaker     string      `json:"speaker"`
	VoiceConfig VoiceConfig `json:"voiceConfig"`
}

// ThinkingConfig 表示思考配置
type ThinkingConfig struct {
	IncludeThoughts *bool   `json:"includeThoughts,omitempty"`
	ThinkingBudget  *int    `json:"thinkingBudget,omitempty"`
	ThinkingLevel   *string `json:"thinkingLevel,omitempty"`
}

// ImageConfig 表示图像配置
type ImageConfig struct {
	AspectRatio *string `json:"aspectRatio,omitempty"`
	ImageSize   *string `json:"imageSize,omitempty"`
}
