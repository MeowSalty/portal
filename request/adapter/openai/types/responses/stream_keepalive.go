package responses

// ResponseKeepaliveEvent 表示 keepalive 流事件。
type ResponseKeepaliveEvent struct {
	Type           StreamEventType `json:"type"`
	SequenceNumber int             `json:"sequence_number"`
}
