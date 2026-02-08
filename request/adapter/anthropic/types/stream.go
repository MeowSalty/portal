package types

import (
	"encoding/json"
	"fmt"
)

// StreamEvent 流式事件联合类型。
// 用于解析所有类型的 SSE 消息。
type StreamEvent struct {
	MessageStart      *MessageStartEvent
	MessageDelta      *MessageDeltaEvent
	MessageStop       *MessageStopEvent
	ContentBlockStart *ContentBlockStartEvent
	ContentBlockDelta *ContentBlockDeltaEvent
	ContentBlockStop  *ContentBlockStopEvent
	Ping              *PingEvent
	Error             *ErrorEvent
}

// StreamEventType 流式事件类型。
type StreamEventType string

const (
	StreamEventMessageStart      StreamEventType = "message_start"
	StreamEventMessageDelta      StreamEventType = "message_delta"
	StreamEventMessageStop       StreamEventType = "message_stop"
	StreamEventContentBlockStart StreamEventType = "content_block_start"
	StreamEventContentBlockDelta StreamEventType = "content_block_delta"
	StreamEventContentBlockStop  StreamEventType = "content_block_stop"
	StreamEventPing              StreamEventType = "ping"
	StreamEventError             StreamEventType = "error"
)

// MessageStartEvent message_start 事件。
type MessageStartEvent struct {
	Type    StreamEventType `json:"type"`
	Message Response        `json:"message"`
}

// ContentBlockStartEvent content_block_start 事件。
type ContentBlockStartEvent struct {
	Type         StreamEventType      `json:"type"`
	Index        int                  `json:"index"`
	ContentBlock ResponseContentBlock `json:"content_block"`
}

// ContentBlockDeltaEvent content_block_delta 事件。
type ContentBlockDeltaEvent struct {
	Type  StreamEventType   `json:"type"`
	Index int               `json:"index"`
	Delta ContentBlockDelta `json:"delta"`
}

// ContentBlockStopEvent content_block_stop 事件。
type ContentBlockStopEvent struct {
	Type  StreamEventType `json:"type"`
	Index int             `json:"index"`
}

// MessageDeltaEvent message_delta 事件。
type MessageDeltaEvent struct {
	Type  StreamEventType    `json:"type"`
	Delta MessageDelta       `json:"delta"`
	Usage *MessageDeltaUsage `json:"usage,omitempty"`
}

// MessageStopEvent message_stop 事件。
type MessageStopEvent struct {
	Type StreamEventType `json:"type"`
}

// PingEvent ping 事件。
type PingEvent struct {
	Type StreamEventType `json:"type"`
}

// ErrorEvent error 事件。
type ErrorEvent struct {
	Type  StreamEventType `json:"type"`
	Error ErrorResponse   `json:"error"`
}

// MessageDelta message_delta 的 delta 结构。
type MessageDelta struct {
	StopReason   *StopReason `json:"stop_reason,omitempty"`
	StopSequence *string     `json:"stop_sequence,omitempty"`
}

// UnmarshalJSON 实现 StreamEvent 的反序列化。
func (e *StreamEvent) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type typeHolder struct {
		Type StreamEventType `json:"type"`
	}
	var t typeHolder
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("流式事件解析失败：%w", err)
	}

	switch t.Type {
	case StreamEventMessageStart:
		var v MessageStartEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("message_start 事件解析失败：%w", err)
		}
		e.MessageStart = &v
	case StreamEventMessageDelta:
		var v MessageDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("message_delta 事件解析失败：%w", err)
		}
		e.MessageDelta = &v
	case StreamEventMessageStop:
		var v MessageStopEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("message_stop 事件解析失败：%w", err)
		}
		e.MessageStop = &v
	case StreamEventContentBlockStart:
		var v ContentBlockStartEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("content_block_start 事件解析失败：%w", err)
		}
		e.ContentBlockStart = &v
	case StreamEventContentBlockDelta:
		var v ContentBlockDeltaEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("content_block_delta 事件解析失败：%w", err)
		}
		e.ContentBlockDelta = &v
	case StreamEventContentBlockStop:
		var v ContentBlockStopEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("content_block_stop 事件解析失败：%w", err)
		}
		e.ContentBlockStop = &v
	case StreamEventPing:
		var v PingEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("ping 事件解析失败：%w", err)
		}
		e.Ping = &v
	case StreamEventError:
		var v ErrorEvent
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("error 事件解析失败：%w", err)
		}
		e.Error = &v
	default:
		return fmt.Errorf("不支持的流式事件类型: %s", t.Type)
	}

	return nil
}

// MarshalJSON 实现 StreamEvent 的序列化。
func (e StreamEvent) MarshalJSON() ([]byte, error) {
	count := 0
	var payload any
	set := func(v any) {
		count++
		payload = v
	}
	if e.MessageStart != nil {
		set(e.MessageStart)
	}
	if e.MessageDelta != nil {
		set(e.MessageDelta)
	}
	if e.MessageStop != nil {
		set(e.MessageStop)
	}
	if e.ContentBlockStart != nil {
		set(e.ContentBlockStart)
	}
	if e.ContentBlockDelta != nil {
		set(e.ContentBlockDelta)
	}
	if e.ContentBlockStop != nil {
		set(e.ContentBlockStop)
	}
	if e.Ping != nil {
		set(e.Ping)
	}
	if e.Error != nil {
		set(e.Error)
	}
	if count == 0 {
		return json.Marshal(nil)
	}
	if count > 1 {
		return nil, fmt.Errorf("流式事件只能设置一种类型")
	}
	return json.Marshal(payload)
}
