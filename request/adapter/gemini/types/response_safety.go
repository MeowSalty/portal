package types

// SafetyRating 表示安全评级。
type SafetyRating struct {
	Category    HarmCategory    `json:"category"`    // 安全类别
	Probability HarmProbability `json:"probability"` // 概率
	Blocked     bool            `json:"blocked"`     // 是否被阻止
}
