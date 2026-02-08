package responses

// InputItemCompaction 表示压缩项（type == "compaction"）
type InputItemCompaction struct {
	ID               string        `json:"id"`
	Type             InputItemType `json:"type"`
	CreatedBy        string        `json:"created_by"`
	EncryptedContent *string       `json:"encrypted_content,omitempty"`
}

// InputItemMCPApprovalRequest 表示 MCP 审批请求（type == "mcp_approval_request"）
type InputItemMCPApprovalRequest struct {
	Type        InputItemType `json:"type"`
	ID          string        `json:"id"`
	ServerLabel string        `json:"server_label"`
	Name        string        `json:"name"`
	Arguments   string        `json:"arguments"`
}

// InputItemMCPApprovalResponse 表示 MCP 审批响应（type == "mcp_approval_response"）
type InputItemMCPApprovalResponse struct {
	Type              InputItemType `json:"type"`
	ID                *string       `json:"id,omitempty"`
	ApprovalRequestID string        `json:"approval_request_id"`
	Approve           bool          `json:"approve"`
	Reason            *string       `json:"reason,omitempty"`
}
