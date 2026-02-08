package types

// ModelStatus 表示模型状态。
type ModelStatus struct {
	ModelStage     ModelStage `json:"modelStage,omitempty"`     // 模型阶段
	RetirementTime string     `json:"retirementTime,omitempty"` // 退役时间
	Message        string     `json:"message,omitempty"`        // 状态信息
}
