package types

// SafetySetting 表示安全设置
type SafetySetting struct {
	Category  string `json:"category"`  // 安全类别
	Threshold string `json:"threshold"` // 阈值
}
