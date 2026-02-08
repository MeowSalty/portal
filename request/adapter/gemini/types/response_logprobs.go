package types

// LogprobsResult 表示对数概率结果。
type LogprobsResult struct {
	LogProbabilitySum float32                   `json:"logProbabilitySum,omitempty"` // 对数概率和
	TopCandidates     []TopCandidates           `json:"topCandidates,omitempty"`     // Top 候选
	ChosenCandidates  []LogprobsResultCandidate `json:"chosenCandidates,omitempty"`  // 选中候选
}

// TopCandidates 表示解码步的候选集合。
type TopCandidates struct {
	Candidates []LogprobsResultCandidate `json:"candidates,omitempty"` // 候选
}

// LogprobsResultCandidate 表示对数概率候选。
type LogprobsResultCandidate struct {
	Token          string  `json:"token,omitempty"`          // token
	TokenID        int32   `json:"tokenId,omitempty"`        // token ID
	LogProbability float32 `json:"logProbability,omitempty"` // 对数概率
}
