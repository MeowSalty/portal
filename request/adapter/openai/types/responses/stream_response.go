package responses

// ResponseCreatedEvent response.created 事件。
type ResponseCreatedEvent struct {
	Type           StreamEventType `json:"type"`
	Response       Response        `json:"response"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseInProgressEvent response.in_progress 事件。
type ResponseInProgressEvent struct {
	Type           StreamEventType `json:"type"`
	Response       Response        `json:"response"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseCompletedEvent response.completed 事件。
type ResponseCompletedEvent struct {
	Type           StreamEventType `json:"type"`
	Response       Response        `json:"response"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseFailedEvent response.failed 事件。
type ResponseFailedEvent struct {
	Type           StreamEventType `json:"type"`
	Response       Response        `json:"response"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseIncompleteEvent response.incomplete 事件。
type ResponseIncompleteEvent struct {
	Type           StreamEventType `json:"type"`
	Response       Response        `json:"response"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseQueuedEvent response.queued 事件。
type ResponseQueuedEvent struct {
	Type           StreamEventType `json:"type"`
	Response       Response        `json:"response"`
	SequenceNumber int             `json:"sequence_number"`
}
