package responses

// ResponseOutputItemAddedEvent response.output_item.added 事件。
type ResponseOutputItemAddedEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	Item           OutputItem      `json:"item"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseOutputItemDoneEvent response.output_item.done 事件。
type ResponseOutputItemDoneEvent struct {
	Type           StreamEventType `json:"type"`
	OutputIndex    int             `json:"output_index"`
	Item           OutputItem      `json:"item"`
	SequenceNumber int             `json:"sequence_number"`
}

// ResponseContentPartAddedEvent response.content_part.added 事件。
type ResponseContentPartAddedEvent struct {
	Type           StreamEventType   `json:"type"`
	ItemID         string            `json:"item_id"`
	OutputIndex    int               `json:"output_index"`
	ContentIndex   int               `json:"content_index"`
	Part           OutputContentPart `json:"part"`
	SequenceNumber int               `json:"sequence_number"`
}

// ResponseContentPartDoneEvent response.content_part.done 事件。
type ResponseContentPartDoneEvent struct {
	Type           StreamEventType   `json:"type"`
	ItemID         string            `json:"item_id"`
	OutputIndex    int               `json:"output_index"`
	ContentIndex   int               `json:"content_index"`
	Part           OutputContentPart `json:"part"`
	SequenceNumber int               `json:"sequence_number"`
}
