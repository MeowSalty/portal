package responses

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

// marshalToMap 将 JSON 反序列化为 map[string]any，用于比较。
func marshalToMap(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("反序列化为 map 失败: %v", err)
	}
	return m
}

// assertOnlyFieldSet 断言 StreamEvent 中只有一个指针字段被设置。
func assertOnlyFieldSet(t *testing.T, event *StreamEvent, fieldName string) {
	t.Helper()
	v := reflect.ValueOf(event).Elem()
	count := 0
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).IsNil() {
			count++
			if v.Type().Field(i).Name != fieldName {
				t.Errorf("期望字段 %s 被设置，但实际设置了 %s", fieldName, v.Type().Field(i).Name)
			}
		}
	}
	if count != 1 {
		t.Errorf("期望只有一个字段被设置，实际设置了 %d 个字段", count)
	}
}

// roundTripEvent 验证事件的往返一致性。
func roundTripEvent(t *testing.T, payload []byte, expected any) {
	t.Helper()

	// JSON -> StreamEvent -> 具体事件类型
	var event StreamEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatalf("反序列化到 StreamEvent 失败: %v", err)
	}

	// 获取具体事件类型
	eventValue := reflect.ValueOf(event)
	var concreteEvent any
	for i := 0; i < eventValue.NumField(); i++ {
		if !eventValue.Field(i).IsNil() {
			concreteEvent = eventValue.Field(i).Interface()
			break
		}
	}

	if concreteEvent == nil {
		t.Fatal("没有设置任何事件字段")
	}

	// 比较字段
	if !reflect.DeepEqual(concreteEvent, expected) {
		t.Errorf("往返后字段不一致:\n期望：%+v\n实际：%+v", expected, concreteEvent)
	}

	// StreamEvent -> JSON -> 具体事件类型
	marshaled, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("序列化 StreamEvent 失败: %v", err)
	}

	var unmarshaled any
	switch expected.(type) {
	case *ResponseAudioDeltaEvent:
		var v ResponseAudioDeltaEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseAudioDoneEvent:
		var v ResponseAudioDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseAudioTranscriptDeltaEvent:
		var v ResponseAudioTranscriptDeltaEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseAudioTranscriptDoneEvent:
		var v ResponseAudioTranscriptDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseCodeInterpreterCallCodeDeltaEvent:
		var v ResponseCodeInterpreterCallCodeDeltaEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseCodeInterpreterCallCodeDoneEvent:
		var v ResponseCodeInterpreterCallCodeDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseCodeInterpreterCallCompletedEvent:
		var v ResponseCodeInterpreterCallCompletedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseCodeInterpreterCallInProgressEvent:
		var v ResponseCodeInterpreterCallInProgressEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseCodeInterpreterCallInterpretingEvent:
		var v ResponseCodeInterpreterCallInterpretingEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseFileSearchCallCompletedEvent:
		var v ResponseFileSearchCallCompletedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseFileSearchCallInProgressEvent:
		var v ResponseFileSearchCallInProgressEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseFileSearchCallSearchingEvent:
		var v ResponseFileSearchCallSearchingEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseWebSearchCallCompletedEvent:
		var v ResponseWebSearchCallCompletedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseWebSearchCallInProgressEvent:
		var v ResponseWebSearchCallInProgressEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseWebSearchCallSearchingEvent:
		var v ResponseWebSearchCallSearchingEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseImageGenCallCompletedEvent:
		var v ResponseImageGenCallCompletedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseImageGenCallGeneratingEvent:
		var v ResponseImageGenCallGeneratingEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseImageGenCallInProgressEvent:
		var v ResponseImageGenCallInProgressEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseImageGenCallPartialImageEvent:
		var v ResponseImageGenCallPartialImageEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseMCPCallArgumentsDeltaEvent:
		var v ResponseMCPCallArgumentsDeltaEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseMCPCallArgumentsDoneEvent:
		var v ResponseMCPCallArgumentsDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseMCPCallCompletedEvent:
		var v ResponseMCPCallCompletedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseMCPCallFailedEvent:
		var v ResponseMCPCallFailedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseMCPCallInProgressEvent:
		var v ResponseMCPCallInProgressEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseMCPListToolsCompletedEvent:
		var v ResponseMCPListToolsCompletedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseMCPListToolsFailedEvent:
		var v ResponseMCPListToolsFailedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseMCPListToolsInProgressEvent:
		var v ResponseMCPListToolsInProgressEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseCustomToolCallInputDeltaEvent:
		var v ResponseCustomToolCallInputDeltaEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseCustomToolCallInputDoneEvent:
		var v ResponseCustomToolCallInputDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseFunctionCallArgumentsDeltaEvent:
		var v ResponseFunctionCallArgumentsDeltaEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseFunctionCallArgumentsDoneEvent:
		var v ResponseFunctionCallArgumentsDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseCreatedEvent:
		var v ResponseCreatedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseInProgressEvent:
		var v ResponseInProgressEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseCompletedEvent:
		var v ResponseCompletedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseFailedEvent:
		var v ResponseFailedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseIncompleteEvent:
		var v ResponseIncompleteEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseQueuedEvent:
		var v ResponseQueuedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseOutputItemAddedEvent:
		var v ResponseOutputItemAddedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseOutputItemDoneEvent:
		var v ResponseOutputItemDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseContentPartAddedEvent:
		var v ResponseContentPartAddedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseContentPartDoneEvent:
		var v ResponseContentPartDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseOutputTextDeltaEvent:
		var v ResponseOutputTextDeltaEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseOutputTextDoneEvent:
		var v ResponseOutputTextDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseOutputTextAnnotationAddedEvent:
		var v ResponseOutputTextAnnotationAddedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseRefusalDeltaEvent:
		var v ResponseRefusalDeltaEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseRefusalDoneEvent:
		var v ResponseRefusalDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseReasoningTextDeltaEvent:
		var v ResponseReasoningTextDeltaEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseReasoningTextDoneEvent:
		var v ResponseReasoningTextDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseReasoningSummaryPartAddedEvent:
		var v ResponseReasoningSummaryPartAddedEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseReasoningSummaryPartDoneEvent:
		var v ResponseReasoningSummaryPartDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseReasoningSummaryTextDeltaEvent:
		var v ResponseReasoningSummaryTextDeltaEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseReasoningSummaryTextDoneEvent:
		var v ResponseReasoningSummaryTextDoneEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseErrorEvent:
		var v ResponseErrorEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	case *ResponseKeepaliveEvent:
		var v ResponseKeepaliveEvent
		if err := json.Unmarshal(marshaled, &v); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}
		unmarshaled = &v
	default:
		t.Fatalf("未知的事件类型：%T", expected)
	}

	if !reflect.DeepEqual(unmarshaled, expected) {
		t.Errorf("往返后字段不一致:\n期望：%+v\n实际：%+v", expected, unmarshaled)
	}

	// 比较原始 JSON 和往返后的 JSON（使用 map 比较，避免字段顺序问题）
	originalMap := marshalToMap(t, payload)
	roundTripMap := marshalToMap(t, marshaled)
	if !reflect.DeepEqual(originalMap, roundTripMap) {
		t.Errorf("往返后 JSON 不一致:\n原始：%+v\n往返：%+v", originalMap, roundTripMap)
	}
}

// TestStreamEvent_Unmarshal_AllTypes 测试所有事件类型的反序列化。
func TestStreamEvent_Unmarshal_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		payload  []byte
		expected any
		field    string
	}{
		// Audio 事件
		{
			name: "response.audio.delta",
			payload: []byte(`{
				"type": "response.audio.delta",
				"response_id": "resp-1",
				"delta": "audio-data",
				"sequence_number": 1
			}`),
			expected: &ResponseAudioDeltaEvent{
				Type:           StreamEventAudioDelta,
				ResponseID:     "resp-1",
				Delta:          "audio-data",
				SequenceNumber: 1,
			},
			field: "AudioDelta",
		},
		{
			name: "response.audio.done",
			payload: []byte(`{
				"type": "response.audio.done",
				"response_id": "resp-1",
				"sequence_number": 2
			}`),
			expected: &ResponseAudioDoneEvent{
				Type:           StreamEventAudioDone,
				ResponseID:     "resp-1",
				SequenceNumber: 2,
			},
			field: "AudioDone",
		},
		{
			name: "response.audio.transcript.delta",
			payload: []byte(`{
				"type": "response.audio.transcript.delta",
				"response_id": "resp-1",
				"delta": "transcript-data",
				"sequence_number": 3
			}`),
			expected: &ResponseAudioTranscriptDeltaEvent{
				Type:           StreamEventAudioTranscriptDelta,
				ResponseID:     "resp-1",
				Delta:          "transcript-data",
				SequenceNumber: 3,
			},
			field: "AudioTranscriptDelta",
		},
		{
			name: "response.audio.transcript.done",
			payload: []byte(`{
				"type": "response.audio.transcript.done",
				"response_id": "resp-1",
				"sequence_number": 4
			}`),
			expected: &ResponseAudioTranscriptDoneEvent{
				Type:           StreamEventAudioTranscriptDone,
				ResponseID:     "resp-1",
				SequenceNumber: 4,
			},
			field: "AudioTranscriptDone",
		},
		// Code Interpreter 事件
		{
			name: "response.code_interpreter_call_code.delta",
			payload: []byte(`{
				"type": "response.code_interpreter_call_code.delta",
				"output_index": 0,
				"item_id": "item-1",
				"delta": "code-delta",
				"sequence_number": 5
			}`),
			expected: &ResponseCodeInterpreterCallCodeDeltaEvent{
				Type:           StreamEventCodeInterpreterCallCodeDelta,
				OutputIndex:    0,
				ItemID:         "item-1",
				Delta:          "code-delta",
				SequenceNumber: 5,
			},
			field: "CodeInterpreterCallCodeDelta",
		},
		{
			name: "response.code_interpreter_call_code.done",
			payload: []byte(`{
				"type": "response.code_interpreter_call_code.done",
				"output_index": 0,
				"item_id": "item-1",
				"code": "print('hello')",
				"sequence_number": 6
			}`),
			expected: &ResponseCodeInterpreterCallCodeDoneEvent{
				Type:           StreamEventCodeInterpreterCallCodeDone,
				OutputIndex:    0,
				ItemID:         "item-1",
				Code:           "print('hello')",
				SequenceNumber: 6,
			},
			field: "CodeInterpreterCallCodeDone",
		},
		{
			name: "response.code_interpreter_call.completed",
			payload: []byte(`{
				"type": "response.code_interpreter_call.completed",
				"output_index": 0,
				"item_id": "item-1",
				"sequence_number": 7
			}`),
			expected: &ResponseCodeInterpreterCallCompletedEvent{
				Type:           StreamEventCodeInterpreterCallCompleted,
				OutputIndex:    0,
				ItemID:         "item-1",
				SequenceNumber: 7,
			},
			field: "CodeInterpreterCallCompleted",
		},
		{
			name: "response.code_interpreter_call.in_progress",
			payload: []byte(`{
				"type": "response.code_interpreter_call.in_progress",
				"output_index": 0,
				"item_id": "item-1",
				"sequence_number": 8
			}`),
			expected: &ResponseCodeInterpreterCallInProgressEvent{
				Type:           StreamEventCodeInterpreterCallInProgress,
				OutputIndex:    0,
				ItemID:         "item-1",
				SequenceNumber: 8,
			},
			field: "CodeInterpreterCallInProgress",
		},
		{
			name: "response.code_interpreter_call.interpreting",
			payload: []byte(`{
				"type": "response.code_interpreter_call.interpreting",
				"output_index": 0,
				"item_id": "item-1",
				"sequence_number": 9
			}`),
			expected: &ResponseCodeInterpreterCallInterpretingEvent{
				Type:           StreamEventCodeInterpreterCallInterpreting,
				OutputIndex:    0,
				ItemID:         "item-1",
				SequenceNumber: 9,
			},
			field: "CodeInterpreterCallInterpreting",
		},
		// File Search 事件
		{
			name: "response.file_search_call.completed",
			payload: []byte(`{
				"type": "response.file_search_call.completed",
				"output_index": 0,
				"item_id": "item-2",
				"sequence_number": 10
			}`),
			expected: &ResponseFileSearchCallCompletedEvent{
				Type:           StreamEventFileSearchCallCompleted,
				OutputIndex:    0,
				ItemID:         "item-2",
				SequenceNumber: 10,
			},
			field: "FileSearchCallCompleted",
		},
		{
			name: "response.file_search_call.in_progress",
			payload: []byte(`{
				"type": "response.file_search_call.in_progress",
				"output_index": 0,
				"item_id": "item-2",
				"sequence_number": 11
			}`),
			expected: &ResponseFileSearchCallInProgressEvent{
				Type:           StreamEventFileSearchCallInProgress,
				OutputIndex:    0,
				ItemID:         "item-2",
				SequenceNumber: 11,
			},
			field: "FileSearchCallInProgress",
		},
		{
			name: "response.file_search_call.searching",
			payload: []byte(`{
				"type": "response.file_search_call.searching",
				"output_index": 0,
				"item_id": "item-2",
				"sequence_number": 12
			}`),
			expected: &ResponseFileSearchCallSearchingEvent{
				Type:           StreamEventFileSearchCallSearching,
				OutputIndex:    0,
				ItemID:         "item-2",
				SequenceNumber: 12,
			},
			field: "FileSearchCallSearching",
		},
		// Web Search 事件
		{
			name: "response.web_search_call.completed",
			payload: []byte(`{
				"type": "response.web_search_call.completed",
				"output_index": 0,
				"item_id": "item-3",
				"sequence_number": 13
			}`),
			expected: &ResponseWebSearchCallCompletedEvent{
				Type:           StreamEventWebSearchCallCompleted,
				OutputIndex:    0,
				ItemID:         "item-3",
				SequenceNumber: 13,
			},
			field: "WebSearchCallCompleted",
		},
		{
			name: "response.web_search_call.in_progress",
			payload: []byte(`{
				"type": "response.web_search_call.in_progress",
				"output_index": 0,
				"item_id": "item-3",
				"sequence_number": 14
			}`),
			expected: &ResponseWebSearchCallInProgressEvent{
				Type:           StreamEventWebSearchCallInProgress,
				OutputIndex:    0,
				ItemID:         "item-3",
				SequenceNumber: 14,
			},
			field: "WebSearchCallInProgress",
		},
		{
			name: "response.web_search_call.searching",
			payload: []byte(`{
				"type": "response.web_search_call.searching",
				"output_index": 0,
				"item_id": "item-3",
				"sequence_number": 15
			}`),
			expected: &ResponseWebSearchCallSearchingEvent{
				Type:           StreamEventWebSearchCallSearching,
				OutputIndex:    0,
				ItemID:         "item-3",
				SequenceNumber: 15,
			},
			field: "WebSearchCallSearching",
		},
		// Image Generation 事件
		{
			name: "response.image_generation_call.completed",
			payload: []byte(`{
				"type": "response.image_generation_call.completed",
				"output_index": 0,
				"item_id": "item-4",
				"sequence_number": 16
			}`),
			expected: &ResponseImageGenCallCompletedEvent{
				Type:           StreamEventImageGenCallCompleted,
				OutputIndex:    0,
				ItemID:         "item-4",
				SequenceNumber: 16,
			},
			field: "ImageGenCallCompleted",
		},
		{
			name: "response.image_generation_call.generating",
			payload: []byte(`{
				"type": "response.image_generation_call.generating",
				"output_index": 0,
				"item_id": "item-4",
				"sequence_number": 17
			}`),
			expected: &ResponseImageGenCallGeneratingEvent{
				Type:           StreamEventImageGenCallGenerating,
				OutputIndex:    0,
				ItemID:         "item-4",
				SequenceNumber: 17,
			},
			field: "ImageGenCallGenerating",
		},
		{
			name: "response.image_generation_call.in_progress",
			payload: []byte(`{
				"type": "response.image_generation_call.in_progress",
				"output_index": 0,
				"item_id": "item-4",
				"sequence_number": 18
			}`),
			expected: &ResponseImageGenCallInProgressEvent{
				Type:           StreamEventImageGenCallInProgress,
				OutputIndex:    0,
				ItemID:         "item-4",
				SequenceNumber: 18,
			},
			field: "ImageGenCallInProgress",
		},
		{
			name: "response.image_generation_call.partial_image",
			payload: []byte(`{
				"type": "response.image_generation_call.partial_image",
				"output_index": 0,
				"item_id": "item-4",
				"sequence_number": 19,
				"partial_image_index": 0,
				"partial_image_b64": "base64data"
			}`),
			expected: &ResponseImageGenCallPartialImageEvent{
				Type:              StreamEventImageGenCallPartialImage,
				OutputIndex:       0,
				ItemID:            "item-4",
				SequenceNumber:    19,
				PartialImageIndex: 0,
				PartialImageB64:   "base64data",
			},
			field: "ImageGenCallPartialImage",
		},
		// MCP 事件
		{
			name: "response.mcp_call_arguments.delta",
			payload: []byte(`{
				"type": "response.mcp_call_arguments.delta",
				"output_index": 0,
				"item_id": "item-5",
				"delta": "args-delta",
				"sequence_number": 20
			}`),
			expected: &ResponseMCPCallArgumentsDeltaEvent{
				Type:           StreamEventMCPCallArgumentsDelta,
				OutputIndex:    0,
				ItemID:         "item-5",
				Delta:          "args-delta",
				SequenceNumber: 20,
			},
			field: "MCPCallArgumentsDelta",
		},
		{
			name: "response.mcp_call_arguments.done",
			payload: []byte(`{
				"type": "response.mcp_call_arguments.done",
				"output_index": 0,
				"item_id": "item-5",
				"arguments": "{\"key\":\"value\"}",
				"sequence_number": 21
			}`),
			expected: &ResponseMCPCallArgumentsDoneEvent{
				Type:           StreamEventMCPCallArgumentsDone,
				OutputIndex:    0,
				ItemID:         "item-5",
				Arguments:      `{"key":"value"}`,
				SequenceNumber: 21,
			},
			field: "MCPCallArgumentsDone",
		},
		{
			name: "response.mcp_call.completed",
			payload: []byte(`{
				"type": "response.mcp_call.completed",
				"output_index": 0,
				"item_id": "item-5",
				"sequence_number": 22
			}`),
			expected: &ResponseMCPCallCompletedEvent{
				Type:           StreamEventMCPCallCompleted,
				OutputIndex:    0,
				ItemID:         "item-5",
				SequenceNumber: 22,
			},
			field: "MCPCallCompleted",
		},
		{
			name: "response.mcp_call.failed",
			payload: []byte(`{
				"type": "response.mcp_call.failed",
				"output_index": 0,
				"item_id": "item-5",
				"sequence_number": 23
			}`),
			expected: &ResponseMCPCallFailedEvent{
				Type:           StreamEventMCPCallFailed,
				OutputIndex:    0,
				ItemID:         "item-5",
				SequenceNumber: 23,
			},
			field: "MCPCallFailed",
		},
		{
			name: "response.mcp_call.in_progress",
			payload: []byte(`{
				"type": "response.mcp_call.in_progress",
				"output_index": 0,
				"item_id": "item-5",
				"sequence_number": 24
			}`),
			expected: &ResponseMCPCallInProgressEvent{
				Type:           StreamEventMCPCallInProgress,
				OutputIndex:    0,
				ItemID:         "item-5",
				SequenceNumber: 24,
			},
			field: "MCPCallInProgress",
		},
		{
			name: "response.mcp_list_tools.completed",
			payload: []byte(`{
				"type": "response.mcp_list_tools.completed",
				"output_index": 0,
				"item_id": "item-6",
				"sequence_number": 25
			}`),
			expected: &ResponseMCPListToolsCompletedEvent{
				Type:           StreamEventMCPListToolsCompleted,
				OutputIndex:    0,
				ItemID:         "item-6",
				SequenceNumber: 25,
			},
			field: "MCPListToolsCompleted",
		},
		{
			name: "response.mcp_list_tools.failed",
			payload: []byte(`{
				"type": "response.mcp_list_tools.failed",
				"output_index": 0,
				"item_id": "item-6",
				"sequence_number": 26
			}`),
			expected: &ResponseMCPListToolsFailedEvent{
				Type:           StreamEventMCPListToolsFailed,
				OutputIndex:    0,
				ItemID:         "item-6",
				SequenceNumber: 26,
			},
			field: "MCPListToolsFailed",
		},
		{
			name: "response.mcp_list_tools.in_progress",
			payload: []byte(`{
				"type": "response.mcp_list_tools.in_progress",
				"output_index": 0,
				"item_id": "item-6",
				"sequence_number": 27
			}`),
			expected: &ResponseMCPListToolsInProgressEvent{
				Type:           StreamEventMCPListToolsInProgress,
				OutputIndex:    0,
				ItemID:         "item-6",
				SequenceNumber: 27,
			},
			field: "MCPListToolsInProgress",
		},
		// Custom Tool Call 事件
		{
			name: "response.custom_tool_call_input.delta",
			payload: []byte(`{
				"type": "response.custom_tool_call_input.delta",
				"output_index": 0,
				"item_id": "item-7",
				"delta": "input-delta",
				"sequence_number": 28
			}`),
			expected: &ResponseCustomToolCallInputDeltaEvent{
				Type:           StreamEventCustomToolCallInputDelta,
				OutputIndex:    0,
				ItemID:         "item-7",
				Delta:          "input-delta",
				SequenceNumber: 28,
			},
			field: "CustomToolCallInputDelta",
		},
		{
			name: "response.custom_tool_call_input.done",
			payload: []byte(`{
				"type": "response.custom_tool_call_input.done",
				"output_index": 0,
				"item_id": "item-7",
				"input": "tool-input",
				"sequence_number": 29
			}`),
			expected: &ResponseCustomToolCallInputDoneEvent{
				Type:           StreamEventCustomToolCallInputDone,
				OutputIndex:    0,
				ItemID:         "item-7",
				Input:          "tool-input",
				SequenceNumber: 29,
			},
			field: "CustomToolCallInputDone",
		},
		// Function Call 事件
		{
			name: "response.function_call_arguments.delta",
			payload: []byte(`{
				"type": "response.function_call_arguments.delta",
				"output_index": 0,
				"item_id": "item-8",
				"delta": "args-delta",
				"sequence_number": 30
			}`),
			expected: &ResponseFunctionCallArgumentsDeltaEvent{
				Type:           StreamEventFunctionCallArgumentsDelta,
				OutputIndex:    0,
				ItemID:         "item-8",
				Delta:          "args-delta",
				SequenceNumber: 30,
			},
			field: "FunctionCallArgumentsDelta",
		},
		{
			name: "response.function_call_arguments.done",
			payload: []byte(`{
				"type": "response.function_call_arguments.done",
				"output_index": 0,
				"item_id": "item-8",
				"name": "function_name",
				"arguments": "{\"arg\":\"value\"}",
				"sequence_number": 31
			}`),
			expected: &ResponseFunctionCallArgumentsDoneEvent{
				Type:           StreamEventFunctionCallArgumentsDone,
				OutputIndex:    0,
				ItemID:         "item-8",
				Name:           "function_name",
				Arguments:      `{"arg":"value"}`,
				SequenceNumber: 31,
			},
			field: "FunctionCallArgumentsDone",
		},
		// Response 生命周期事件
		{
			name: "response.created",
			payload: []byte(`{
				"type": "response.created",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 32
			}`),
			expected: &ResponseCreatedEvent{
				Type: StreamEventCreated,
				Response: Response{
					ID:                "resp-1",
					Object:            "response",
					CreatedAt:         1234567890,
					Model:             "gpt-4",
					Output:            []OutputItem{},
					ParallelToolCalls: false,
					Metadata:          map[string]string{},
					Tools:             []shared.ToolUnion{},
					ToolChoice:        nil,
				},
				SequenceNumber: 32,
			},
			field: "Created",
		},
		{
			name: "response.in_progress",
			payload: []byte(`{
				"type": "response.in_progress",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 33
			}`),
			expected: &ResponseInProgressEvent{
				Type: StreamEventInProgress,
				Response: Response{
					ID:                "resp-1",
					Object:            "response",
					CreatedAt:         1234567890,
					Model:             "gpt-4",
					Output:            []OutputItem{},
					ParallelToolCalls: false,
					Metadata:          map[string]string{},
					Tools:             []shared.ToolUnion{},
					ToolChoice:        nil,
				},
				SequenceNumber: 33,
			},
			field: "InProgress",
		},
		{
			name: "response.completed",
			payload: []byte(`{
				"type": "response.completed",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 34
			}`),
			expected: &ResponseCompletedEvent{
				Type: StreamEventCompleted,
				Response: Response{
					ID:                "resp-1",
					Object:            "response",
					CreatedAt:         1234567890,
					Model:             "gpt-4",
					Output:            []OutputItem{},
					ParallelToolCalls: false,
					Metadata:          map[string]string{},
					Tools:             []shared.ToolUnion{},
					ToolChoice:        nil,
				},
				SequenceNumber: 34,
			},
			field: "Completed",
		},
		{
			name: "response.failed",
			payload: []byte(`{
				"type": "response.failed",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 35
			}`),
			expected: &ResponseFailedEvent{
				Type: StreamEventFailed,
				Response: Response{
					ID:                "resp-1",
					Object:            "response",
					CreatedAt:         1234567890,
					Model:             "gpt-4",
					Output:            []OutputItem{},
					ParallelToolCalls: false,
					Metadata:          map[string]string{},
					Tools:             []shared.ToolUnion{},
					ToolChoice:        nil,
				},
				SequenceNumber: 35,
			},
			field: "Failed",
		},
		{
			name: "response.incomplete",
			payload: []byte(`{
				"type": "response.incomplete",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 36
			}`),
			expected: &ResponseIncompleteEvent{
				Type: StreamEventIncomplete,
				Response: Response{
					ID:                "resp-1",
					Object:            "response",
					CreatedAt:         1234567890,
					Model:             "gpt-4",
					Output:            []OutputItem{},
					ParallelToolCalls: false,
					Metadata:          map[string]string{},
					Tools:             []shared.ToolUnion{},
					ToolChoice:        nil,
				},
				SequenceNumber: 36,
			},
			field: "Incomplete",
		},
		{
			name: "response.queued",
			payload: []byte(`{
				"type": "response.queued",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 37
			}`),
			expected: &ResponseQueuedEvent{
				Type: StreamEventQueued,
				Response: Response{
					ID:                "resp-1",
					Object:            "response",
					CreatedAt:         1234567890,
					Model:             "gpt-4",
					Output:            []OutputItem{},
					ParallelToolCalls: false,
					Metadata:          map[string]string{},
					Tools:             []shared.ToolUnion{},
					ToolChoice:        nil,
				},
				SequenceNumber: 37,
			},
			field: "Queued",
		},
		// Output Item 事件
		{
			name: "response.output_item.added",
			payload: []byte(`{
				"type": "response.output_item.added",
				"output_index": 0,
				"item": {
					"type": "message",
					"id": "msg_1",
					"role": "assistant",
					"content": [],
					"status": "in_progress"
				},
				"sequence_number": 38
			}`),
			expected: &ResponseOutputItemAddedEvent{
				Type:        StreamEventOutputItemAdded,
				OutputIndex: 0,
				Item: OutputItem{
					Message: &OutputMessage{
						Type:    "message",
						ID:      "msg_1",
						Role:    "assistant",
						Content: []OutputMessageContent{},
						Status:  "in_progress",
					},
				},
				SequenceNumber: 38,
			},
			field: "OutputItemAdded",
		},
		{
			name: "response.output_item.done",
			payload: []byte(`{
				"type": "response.output_item.done",
				"output_index": 0,
				"item": {
					"type": "message",
					"id": "msg_1",
					"role": "assistant",
					"content": [],
					"status": "completed"
				},
				"sequence_number": 39
			}`),
			expected: &ResponseOutputItemDoneEvent{
				Type:        StreamEventOutputItemDone,
				OutputIndex: 0,
				Item: OutputItem{
					Message: &OutputMessage{
						Type:    "message",
						ID:      "msg_1",
						Role:    "assistant",
						Content: []OutputMessageContent{},
						Status:  "completed",
					},
				},
				SequenceNumber: 39,
			},
			field: "OutputItemDone",
		},
		// Content Part 事件
		{
			name: "response.content_part.added",
			payload: []byte(`{
				"type": "response.content_part.added",
				"item_id": "item-9",
				"output_index": 0,
				"content_index": 0,
				"part": {
					"type": "output_text",
					"text": "test content",
					"annotations": []
				},
				"sequence_number": 40
			}`),
			expected: &ResponseContentPartAddedEvent{
				Type:         StreamEventContentPartAdded,
				ItemID:       "item-9",
				OutputIndex:  0,
				ContentIndex: 0,
				Part: OutputContentPart{
					OutputText: &OutputTextContent{
						Type:        OutputMessageContentTypeOutputText,
						Text:        "test content",
						Annotations: []Annotation{},
					},
				},
				SequenceNumber: 40,
			},
			field: "ContentPartAdded",
		},
		{
			name: "response.content_part.done",
			payload: []byte(`{
				"type": "response.content_part.done",
				"item_id": "item-9",
				"output_index": 0,
				"content_index": 0,
				"part": {
					"type": "output_text",
					"text": "test content",
					"annotations": []
				},
				"sequence_number": 41
			}`),
			expected: &ResponseContentPartDoneEvent{
				Type:         StreamEventContentPartDone,
				ItemID:       "item-9",
				OutputIndex:  0,
				ContentIndex: 0,
				Part: OutputContentPart{
					OutputText: &OutputTextContent{
						Type:        OutputMessageContentTypeOutputText,
						Text:        "test content",
						Annotations: []Annotation{},
					},
				},
				SequenceNumber: 41,
			},
			field: "ContentPartDone",
		},
		// Output Text 事件
		{
			name: "response.output_text.delta",
			payload: []byte(`{
				"type": "response.output_text.delta",
				"item_id": "item-9",
				"output_index": 0,
				"content_index": 0,
				"delta": "text-delta",
				"sequence_number": 42,
				"logprobs": []
			}`),
			expected: &ResponseOutputTextDeltaEvent{
				Type:           StreamEventOutputTextDelta,
				ItemID:         "item-9",
				OutputIndex:    0,
				ContentIndex:   0,
				Delta:          "text-delta",
				SequenceNumber: 42,
				Logprobs:       []ResponseLogProb{},
			},
			field: "OutputTextDelta",
		},
		{
			name: "response.output_text.done",
			payload: []byte(`{
				"type": "response.output_text.done",
				"item_id": "item-9",
				"output_index": 0,
				"content_index": 0,
				"text": "full text",
				"sequence_number": 43,
				"logprobs": []
			}`),
			expected: &ResponseOutputTextDoneEvent{
				Type:           StreamEventOutputTextDone,
				ItemID:         "item-9",
				OutputIndex:    0,
				ContentIndex:   0,
				Text:           "full text",
				SequenceNumber: 43,
				Logprobs:       []ResponseLogProb{},
			},
			field: "OutputTextDone",
		},
		{
			name: "response.output_text.annotation.added",
			payload: []byte(`{
				"type": "response.output_text.annotation.added",
				"item_id": "item-9",
				"output_index": 0,
				"content_index": 0,
				"annotation_index": 0,
				"annotation": {
					"type": "file_citation",
					"file_id": "file-1",
					"index": 0,
					"filename": "example.txt"
				},
				"sequence_number": 44
			}`),
			expected: &ResponseOutputTextAnnotationAddedEvent{
				Type:            StreamEventOutputTextAnnotationAdded,
				ItemID:          "item-9",
				OutputIndex:     0,
				ContentIndex:    0,
				AnnotationIndex: 0,
				Annotation: Annotation{
					FileCitation: &FileCitationAnnotation{
						Type:     AnnotationTypeFileCitation,
						FileID:   "file-1",
						Index:    0,
						Filename: "example.txt",
					},
				},
				SequenceNumber: 44,
			},
			field: "OutputTextAnnotationAdded",
		},
		// Refusal 事件
		{
			name: "response.refusal.delta",
			payload: []byte(`{
				"type": "response.refusal.delta",
				"item_id": "item-10",
				"output_index": 0,
				"content_index": 0,
				"delta": "refusal-delta",
				"sequence_number": 45
			}`),
			expected: &ResponseRefusalDeltaEvent{
				Type:           StreamEventRefusalDelta,
				ItemID:         "item-10",
				OutputIndex:    0,
				ContentIndex:   0,
				Delta:          "refusal-delta",
				SequenceNumber: 45,
			},
			field: "RefusalDelta",
		},
		{
			name: "response.refusal.done",
			payload: []byte(`{
				"type": "response.refusal.done",
				"item_id": "item-10",
				"output_index": 0,
				"content_index": 0,
				"refusal": "full refusal",
				"sequence_number": 46
			}`),
			expected: &ResponseRefusalDoneEvent{
				Type:           StreamEventRefusalDone,
				ItemID:         "item-10",
				OutputIndex:    0,
				ContentIndex:   0,
				Refusal:        "full refusal",
				SequenceNumber: 46,
			},
			field: "RefusalDone",
		},
		// Reasoning Text 事件
		{
			name: "response.reasoning_text.delta",
			payload: []byte(`{
				"type": "response.reasoning_text.delta",
				"item_id": "item-11",
				"output_index": 0,
				"content_index": 0,
				"delta": "reasoning-delta",
				"sequence_number": 47
			}`),
			expected: &ResponseReasoningTextDeltaEvent{
				Type:           StreamEventReasoningTextDelta,
				ItemID:         "item-11",
				OutputIndex:    0,
				ContentIndex:   0,
				Delta:          "reasoning-delta",
				SequenceNumber: 47,
			},
			field: "ReasoningTextDelta",
		},
		{
			name: "response.reasoning_text.done",
			payload: []byte(`{
				"type": "response.reasoning_text.done",
				"item_id": "item-11",
				"output_index": 0,
				"content_index": 0,
				"text": "full reasoning",
				"sequence_number": 48
			}`),
			expected: &ResponseReasoningTextDoneEvent{
				Type:           StreamEventReasoningTextDone,
				ItemID:         "item-11",
				OutputIndex:    0,
				ContentIndex:   0,
				Text:           "full reasoning",
				SequenceNumber: 48,
			},
			field: "ReasoningTextDone",
		},
		// Reasoning Summary 事件
		{
			name: "response.reasoning_summary_part.added",
			payload: []byte(`{
				"type": "response.reasoning_summary_part.added",
				"item_id": "item-12",
				"output_index": 0,
				"summary_index": 0,
				"part": {
					"type": "summary_text",
					"text": "summary text"
				},
				"sequence_number": 49
			}`),
			expected: &ResponseReasoningSummaryPartAddedEvent{
				Type:         StreamEventReasoningSummaryPartAdded,
				ItemID:       "item-12",
				OutputIndex:  0,
				SummaryIndex: 0,
				Part: OutputSummaryPart{
					Type: "summary_text",
					Text: "summary text",
				},
				SequenceNumber: 49,
			},
			field: "ReasoningSummaryPartAdded",
		},
		{
			name: "response.reasoning_summary_part.done",
			payload: []byte(`{
				"type": "response.reasoning_summary_part.done",
				"item_id": "item-12",
				"output_index": 0,
				"summary_index": 0,
				"part": {
					"type": "summary_text",
					"text": "summary text"
				},
				"sequence_number": 50
			}`),
			expected: &ResponseReasoningSummaryPartDoneEvent{
				Type:         StreamEventReasoningSummaryPartDone,
				ItemID:       "item-12",
				OutputIndex:  0,
				SummaryIndex: 0,
				Part: OutputSummaryPart{
					Type: "summary_text",
					Text: "summary text",
				},
				SequenceNumber: 50,
			},
			field: "ReasoningSummaryPartDone",
		},
		{
			name: "response.reasoning_summary_text.delta",
			payload: []byte(`{
				"type": "response.reasoning_summary_text.delta",
				"item_id": "item-12",
				"output_index": 0,
				"summary_index": 0,
				"delta": "summary-delta",
				"sequence_number": 51
			}`),
			expected: &ResponseReasoningSummaryTextDeltaEvent{
				Type:           StreamEventReasoningSummaryTextDelta,
				ItemID:         "item-12",
				OutputIndex:    0,
				SummaryIndex:   0,
				Delta:          "summary-delta",
				SequenceNumber: 51,
			},
			field: "ReasoningSummaryTextDelta",
		},
		{
			name: "response.reasoning_summary_text.done",
			payload: []byte(`{
				"type": "response.reasoning_summary_text.done",
				"item_id": "item-12",
				"output_index": 0,
				"summary_index": 0,
				"text": "full summary",
				"sequence_number": 52
			}`),
			expected: &ResponseReasoningSummaryTextDoneEvent{
				Type:           StreamEventReasoningSummaryTextDone,
				ItemID:         "item-12",
				OutputIndex:    0,
				SummaryIndex:   0,
				Text:           "full summary",
				SequenceNumber: 52,
			},
			field: "ReasoningSummaryTextDone",
		},
		// Error 事件
		{
			name: "error",
			payload: []byte(`{
				"type": "error",
				"code": "invalid_request",
				"message": "Invalid request",
				"sequence_number": 1
			}`),
			expected: &ResponseErrorEvent{
				Type:           StreamEventError,
				Code:           strPtr("invalid_request"),
				Message:        "Invalid request",
				SequenceNumber: 1,
			},
			field: "Error",
		},
		// Keepalive 事件
		{
			name: "keepalive",
			payload: []byte(`{
				"type": "keepalive",
				"sequence_number": 2
			}`),
			expected: &ResponseKeepaliveEvent{
				Type:           StreamEventKeepalive,
				SequenceNumber: 2,
			},
			field: "Keepalive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event StreamEvent
			if err := json.Unmarshal(tt.payload, &event); err != nil {
				t.Fatalf("反序列化失败: %v", err)
			}

			assertOnlyFieldSet(t, &event, tt.field)

			// 验证具体事件类型
			eventValue := reflect.ValueOf(event)
			for i := 0; i < eventValue.NumField(); i++ {
				if !eventValue.Field(i).IsNil() {
					concreteEvent := eventValue.Field(i).Interface()
					if !reflect.DeepEqual(concreteEvent, tt.expected) {
						t.Errorf("字段不一致:\n期望：%+v\n实际：%+v", tt.expected, concreteEvent)
					}
					break
				}
			}
		})
	}
}

// TestStreamEvent_Marshal_AllTypes 测试所有事件类型的序列化。
func TestStreamEvent_Marshal_AllTypes(t *testing.T) {
	tests := []struct {
		name    string
		event   StreamEvent
		payload []byte
	}{
		// Audio 事件
		{
			name: "response.audio.delta",
			event: StreamEvent{
				AudioDelta: &ResponseAudioDeltaEvent{
					Type:           StreamEventAudioDelta,
					ResponseID:     "resp-1",
					Delta:          "audio-data",
					SequenceNumber: 1,
				},
			},
			payload: []byte(`{
				"type": "response.audio.delta",
				"response_id": "resp-1",
				"delta": "audio-data",
				"sequence_number": 1
			}`),
		},
		{
			name: "response.audio.done",
			event: StreamEvent{
				AudioDone: &ResponseAudioDoneEvent{
					Type:           StreamEventAudioDone,
					ResponseID:     "resp-1",
					SequenceNumber: 2,
				},
			},
			payload: []byte(`{
				"type": "response.audio.done",
				"response_id": "resp-1",
				"sequence_number": 2
			}`),
		},
		{
			name: "response.audio.transcript.delta",
			event: StreamEvent{
				AudioTranscriptDelta: &ResponseAudioTranscriptDeltaEvent{
					Type:           StreamEventAudioTranscriptDelta,
					ResponseID:     "resp-1",
					Delta:          "transcript-data",
					SequenceNumber: 3,
				},
			},
			payload: []byte(`{
				"type": "response.audio.transcript.delta",
				"response_id": "resp-1",
				"delta": "transcript-data",
				"sequence_number": 3
			}`),
		},
		{
			name: "response.audio.transcript.done",
			event: StreamEvent{
				AudioTranscriptDone: &ResponseAudioTranscriptDoneEvent{
					Type:           StreamEventAudioTranscriptDone,
					ResponseID:     "resp-1",
					SequenceNumber: 4,
				},
			},
			payload: []byte(`{
				"type": "response.audio.transcript.done",
				"response_id": "resp-1",
				"sequence_number": 4
			}`),
		},
		// Code Interpreter 事件
		{
			name: "response.code_interpreter_call_code.delta",
			event: StreamEvent{
				CodeInterpreterCallCodeDelta: &ResponseCodeInterpreterCallCodeDeltaEvent{
					Type:           StreamEventCodeInterpreterCallCodeDelta,
					OutputIndex:    0,
					ItemID:         "item-1",
					Delta:          "code-delta",
					SequenceNumber: 5,
				},
			},
			payload: []byte(`{
				"type": "response.code_interpreter_call_code.delta",
				"output_index": 0,
				"item_id": "item-1",
				"delta": "code-delta",
				"sequence_number": 5
			}`),
		},
		{
			name: "response.code_interpreter_call_code.done",
			event: StreamEvent{
				CodeInterpreterCallCodeDone: &ResponseCodeInterpreterCallCodeDoneEvent{
					Type:           StreamEventCodeInterpreterCallCodeDone,
					OutputIndex:    0,
					ItemID:         "item-1",
					Code:           "print('hello')",
					SequenceNumber: 6,
				},
			},
			payload: []byte(`{
				"type": "response.code_interpreter_call_code.done",
				"output_index": 0,
				"item_id": "item-1",
				"code": "print('hello')",
				"sequence_number": 6
			}`),
		},
		{
			name: "response.code_interpreter_call.completed",
			event: StreamEvent{
				CodeInterpreterCallCompleted: &ResponseCodeInterpreterCallCompletedEvent{
					Type:           StreamEventCodeInterpreterCallCompleted,
					OutputIndex:    0,
					ItemID:         "item-1",
					SequenceNumber: 7,
				},
			},
			payload: []byte(`{
				"type": "response.code_interpreter_call.completed",
				"output_index": 0,
				"item_id": "item-1",
				"sequence_number": 7
			}`),
		},
		{
			name: "response.code_interpreter_call.in_progress",
			event: StreamEvent{
				CodeInterpreterCallInProgress: &ResponseCodeInterpreterCallInProgressEvent{
					Type:           StreamEventCodeInterpreterCallInProgress,
					OutputIndex:    0,
					ItemID:         "item-1",
					SequenceNumber: 8,
				},
			},
			payload: []byte(`{
				"type": "response.code_interpreter_call.in_progress",
				"output_index": 0,
				"item_id": "item-1",
				"sequence_number": 8
			}`),
		},
		{
			name: "response.code_interpreter_call.interpreting",
			event: StreamEvent{
				CodeInterpreterCallInterpreting: &ResponseCodeInterpreterCallInterpretingEvent{
					Type:           StreamEventCodeInterpreterCallInterpreting,
					OutputIndex:    0,
					ItemID:         "item-1",
					SequenceNumber: 9,
				},
			},
			payload: []byte(`{
				"type": "response.code_interpreter_call.interpreting",
				"output_index": 0,
				"item_id": "item-1",
				"sequence_number": 9
			}`),
		},
		// File Search 事件
		{
			name: "response.file_search_call.completed",
			event: StreamEvent{
				FileSearchCallCompleted: &ResponseFileSearchCallCompletedEvent{
					Type:           StreamEventFileSearchCallCompleted,
					OutputIndex:    0,
					ItemID:         "item-2",
					SequenceNumber: 10,
				},
			},
			payload: []byte(`{
				"type": "response.file_search_call.completed",
				"output_index": 0,
				"item_id": "item-2",
				"sequence_number": 10
			}`),
		},
		{
			name: "response.file_search_call.in_progress",
			event: StreamEvent{
				FileSearchCallInProgress: &ResponseFileSearchCallInProgressEvent{
					Type:           StreamEventFileSearchCallInProgress,
					OutputIndex:    0,
					ItemID:         "item-2",
					SequenceNumber: 11,
				},
			},
			payload: []byte(`{
				"type": "response.file_search_call.in_progress",
				"output_index": 0,
				"item_id": "item-2",
				"sequence_number": 11
			}`),
		},
		{
			name: "response.file_search_call.searching",
			event: StreamEvent{
				FileSearchCallSearching: &ResponseFileSearchCallSearchingEvent{
					Type:           StreamEventFileSearchCallSearching,
					OutputIndex:    0,
					ItemID:         "item-2",
					SequenceNumber: 12,
				},
			},
			payload: []byte(`{
				"type": "response.file_search_call.searching",
				"output_index": 0,
				"item_id": "item-2",
				"sequence_number": 12
			}`),
		},
		// Web Search 事件
		{
			name: "response.web_search_call.completed",
			event: StreamEvent{
				WebSearchCallCompleted: &ResponseWebSearchCallCompletedEvent{
					Type:           StreamEventWebSearchCallCompleted,
					OutputIndex:    0,
					ItemID:         "item-3",
					SequenceNumber: 13,
				},
			},
			payload: []byte(`{
				"type": "response.web_search_call.completed",
				"output_index": 0,
				"item_id": "item-3",
				"sequence_number": 13
			}`),
		},
		{
			name: "response.web_search_call.in_progress",
			event: StreamEvent{
				WebSearchCallInProgress: &ResponseWebSearchCallInProgressEvent{
					Type:           StreamEventWebSearchCallInProgress,
					OutputIndex:    0,
					ItemID:         "item-3",
					SequenceNumber: 14,
				},
			},
			payload: []byte(`{
				"type": "response.web_search_call.in_progress",
				"output_index": 0,
				"item_id": "item-3",
				"sequence_number": 14
			}`),
		},
		{
			name: "response.web_search_call.searching",
			event: StreamEvent{
				WebSearchCallSearching: &ResponseWebSearchCallSearchingEvent{
					Type:           StreamEventWebSearchCallSearching,
					OutputIndex:    0,
					ItemID:         "item-3",
					SequenceNumber: 15,
				},
			},
			payload: []byte(`{
				"type": "response.web_search_call.searching",
				"output_index": 0,
				"item_id": "item-3",
				"sequence_number": 15
			}`),
		},
		// Image Generation 事件
		{
			name: "response.image_generation_call.completed",
			event: StreamEvent{
				ImageGenCallCompleted: &ResponseImageGenCallCompletedEvent{
					Type:           StreamEventImageGenCallCompleted,
					OutputIndex:    0,
					ItemID:         "item-4",
					SequenceNumber: 16,
				},
			},
			payload: []byte(`{
				"type": "response.image_generation_call.completed",
				"output_index": 0,
				"item_id": "item-4",
				"sequence_number": 16
			}`),
		},
		{
			name: "response.image_generation_call.generating",
			event: StreamEvent{
				ImageGenCallGenerating: &ResponseImageGenCallGeneratingEvent{
					Type:           StreamEventImageGenCallGenerating,
					OutputIndex:    0,
					ItemID:         "item-4",
					SequenceNumber: 17,
				},
			},
			payload: []byte(`{
				"type": "response.image_generation_call.generating",
				"output_index": 0,
				"item_id": "item-4",
				"sequence_number": 17
			}`),
		},
		{
			name: "response.image_generation_call.in_progress",
			event: StreamEvent{
				ImageGenCallInProgress: &ResponseImageGenCallInProgressEvent{
					Type:           StreamEventImageGenCallInProgress,
					OutputIndex:    0,
					ItemID:         "item-4",
					SequenceNumber: 18,
				},
			},
			payload: []byte(`{
				"type": "response.image_generation_call.in_progress",
				"output_index": 0,
				"item_id": "item-4",
				"sequence_number": 18
			}`),
		},
		{
			name: "response.image_generation_call.partial_image",
			event: StreamEvent{
				ImageGenCallPartialImage: &ResponseImageGenCallPartialImageEvent{
					Type:              StreamEventImageGenCallPartialImage,
					OutputIndex:       0,
					ItemID:            "item-4",
					SequenceNumber:    19,
					PartialImageIndex: 0,
					PartialImageB64:   "base64data",
				},
			},
			payload: []byte(`{
				"type": "response.image_generation_call.partial_image",
				"output_index": 0,
				"item_id": "item-4",
				"sequence_number": 19,
				"partial_image_index": 0,
				"partial_image_b64": "base64data"
			}`),
		},
		// MCP 事件
		{
			name: "response.mcp_call_arguments.delta",
			event: StreamEvent{
				MCPCallArgumentsDelta: &ResponseMCPCallArgumentsDeltaEvent{
					Type:           StreamEventMCPCallArgumentsDelta,
					OutputIndex:    0,
					ItemID:         "item-5",
					Delta:          "args-delta",
					SequenceNumber: 20,
				},
			},
			payload: []byte(`{
				"type": "response.mcp_call_arguments.delta",
				"output_index": 0,
				"item_id": "item-5",
				"delta": "args-delta",
				"sequence_number": 20
			}`),
		},
		{
			name: "response.mcp_call_arguments.done",
			event: StreamEvent{
				MCPCallArgumentsDone: &ResponseMCPCallArgumentsDoneEvent{
					Type:           StreamEventMCPCallArgumentsDone,
					OutputIndex:    0,
					ItemID:         "item-5",
					Arguments:      `{"key":"value"}`,
					SequenceNumber: 21,
				},
			},
			payload: []byte(`{
				"type": "response.mcp_call_arguments.done",
				"output_index": 0,
				"item_id": "item-5",
				"arguments": "{\"key\":\"value\"}",
				"sequence_number": 21
			}`),
		},
		{
			name: "response.mcp_call.completed",
			event: StreamEvent{
				MCPCallCompleted: &ResponseMCPCallCompletedEvent{
					Type:           StreamEventMCPCallCompleted,
					OutputIndex:    0,
					ItemID:         "item-5",
					SequenceNumber: 22,
				},
			},
			payload: []byte(`{
				"type": "response.mcp_call.completed",
				"output_index": 0,
				"item_id": "item-5",
				"sequence_number": 22
			}`),
		},
		{
			name: "response.mcp_call.failed",
			event: StreamEvent{
				MCPCallFailed: &ResponseMCPCallFailedEvent{
					Type:           StreamEventMCPCallFailed,
					OutputIndex:    0,
					ItemID:         "item-5",
					SequenceNumber: 23,
				},
			},
			payload: []byte(`{
				"type": "response.mcp_call.failed",
				"output_index": 0,
				"item_id": "item-5",
				"sequence_number": 23
			}`),
		},
		{
			name: "response.mcp_call.in_progress",
			event: StreamEvent{
				MCPCallInProgress: &ResponseMCPCallInProgressEvent{
					Type:           StreamEventMCPCallInProgress,
					OutputIndex:    0,
					ItemID:         "item-5",
					SequenceNumber: 24,
				},
			},
			payload: []byte(`{
				"type": "response.mcp_call.in_progress",
				"output_index": 0,
				"item_id": "item-5",
				"sequence_number": 24
			}`),
		},
		{
			name: "response.mcp_list_tools.completed",
			event: StreamEvent{
				MCPListToolsCompleted: &ResponseMCPListToolsCompletedEvent{
					Type:           StreamEventMCPListToolsCompleted,
					OutputIndex:    0,
					ItemID:         "item-6",
					SequenceNumber: 25,
				},
			},
			payload: []byte(`{
				"type": "response.mcp_list_tools.completed",
				"output_index": 0,
				"item_id": "item-6",
				"sequence_number": 25
			}`),
		},
		{
			name: "response.mcp_list_tools.failed",
			event: StreamEvent{
				MCPListToolsFailed: &ResponseMCPListToolsFailedEvent{
					Type:           StreamEventMCPListToolsFailed,
					OutputIndex:    0,
					ItemID:         "item-6",
					SequenceNumber: 26,
				},
			},
			payload: []byte(`{
				"type": "response.mcp_list_tools.failed",
				"output_index": 0,
				"item_id": "item-6",
				"sequence_number": 26
			}`),
		},
		{
			name: "response.mcp_list_tools.in_progress",
			event: StreamEvent{
				MCPListToolsInProgress: &ResponseMCPListToolsInProgressEvent{
					Type:           StreamEventMCPListToolsInProgress,
					OutputIndex:    0,
					ItemID:         "item-6",
					SequenceNumber: 27,
				},
			},
			payload: []byte(`{
				"type": "response.mcp_list_tools.in_progress",
				"output_index": 0,
				"item_id": "item-6",
				"sequence_number": 27
			}`),
		},
		// Custom Tool Call 事件
		{
			name: "response.custom_tool_call_input.delta",
			event: StreamEvent{
				CustomToolCallInputDelta: &ResponseCustomToolCallInputDeltaEvent{
					Type:           StreamEventCustomToolCallInputDelta,
					OutputIndex:    0,
					ItemID:         "item-7",
					Delta:          "input-delta",
					SequenceNumber: 28,
				},
			},
			payload: []byte(`{
				"type": "response.custom_tool_call_input.delta",
				"output_index": 0,
				"item_id": "item-7",
				"delta": "input-delta",
				"sequence_number": 28
			}`),
		},
		{
			name: "response.custom_tool_call_input.done",
			event: StreamEvent{
				CustomToolCallInputDone: &ResponseCustomToolCallInputDoneEvent{
					Type:           StreamEventCustomToolCallInputDone,
					OutputIndex:    0,
					ItemID:         "item-7",
					Input:          "tool-input",
					SequenceNumber: 29,
				},
			},
			payload: []byte(`{
				"type": "response.custom_tool_call_input.done",
				"output_index": 0,
				"item_id": "item-7",
				"input": "tool-input",
				"sequence_number": 29
			}`),
		},
		// Function Call 事件
		{
			name: "response.function_call_arguments.delta",
			event: StreamEvent{
				FunctionCallArgumentsDelta: &ResponseFunctionCallArgumentsDeltaEvent{
					Type:           StreamEventFunctionCallArgumentsDelta,
					OutputIndex:    0,
					ItemID:         "item-8",
					Delta:          "args-delta",
					SequenceNumber: 30,
				},
			},
			payload: []byte(`{
				"type": "response.function_call_arguments.delta",
				"output_index": 0,
				"item_id": "item-8",
				"delta": "args-delta",
				"sequence_number": 30
			}`),
		},
		{
			name: "response.function_call_arguments.done",
			event: StreamEvent{
				FunctionCallArgumentsDone: &ResponseFunctionCallArgumentsDoneEvent{
					Type:           StreamEventFunctionCallArgumentsDone,
					OutputIndex:    0,
					ItemID:         "item-8",
					Name:           "function_name",
					Arguments:      `{"arg":"value"}`,
					SequenceNumber: 31,
				},
			},
			payload: []byte(`{
				"type": "response.function_call_arguments.done",
				"output_index": 0,
				"item_id": "item-8",
				"name": "function_name",
				"arguments": "{\"arg\":\"value\"}",
				"sequence_number": 31
			}`),
		},
		// Response 生命周期事件
		{
			name: "response.created",
			event: StreamEvent{
				Created: &ResponseCreatedEvent{
					Type: StreamEventCreated,
					Response: Response{
						ID:                "resp-1",
						Object:            "response",
						CreatedAt:         1234567890,
						Model:             "gpt-4",
						Output:            []OutputItem{},
						ParallelToolCalls: false,
						Metadata:          map[string]string{},
						Tools:             []shared.ToolUnion{},
						ToolChoice:        nil,
					},
					SequenceNumber: 32,
				},
			},
			payload: []byte(`{
				"type": "response.created",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 32
			}`),
		},
		{
			name: "response.in_progress",
			event: StreamEvent{
				InProgress: &ResponseInProgressEvent{
					Type: StreamEventInProgress,
					Response: Response{
						ID:                "resp-1",
						Object:            "response",
						CreatedAt:         1234567890,
						Model:             "gpt-4",
						Output:            []OutputItem{},
						ParallelToolCalls: false,
						Metadata:          map[string]string{},
						Tools:             []shared.ToolUnion{},
						ToolChoice:        nil,
					},
					SequenceNumber: 33,
				},
			},
			payload: []byte(`{
				"type": "response.in_progress",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 33
			}`),
		},
		{
			name: "response.completed",
			event: StreamEvent{
				Completed: &ResponseCompletedEvent{
					Type: StreamEventCompleted,
					Response: Response{
						ID:                "resp-1",
						Object:            "response",
						CreatedAt:         1234567890,
						Model:             "gpt-4",
						Output:            []OutputItem{},
						ParallelToolCalls: false,
						Metadata:          map[string]string{},
						Tools:             []shared.ToolUnion{},
						ToolChoice:        nil,
					},
					SequenceNumber: 34,
				},
			},
			payload: []byte(`{
				"type": "response.completed",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 34
			}`),
		},
		{
			name: "response.failed",
			event: StreamEvent{
				Failed: &ResponseFailedEvent{
					Type: StreamEventFailed,
					Response: Response{
						ID:                "resp-1",
						Object:            "response",
						CreatedAt:         1234567890,
						Model:             "gpt-4",
						Output:            []OutputItem{},
						ParallelToolCalls: false,
						Metadata:          map[string]string{},
						Tools:             []shared.ToolUnion{},
						ToolChoice:        nil,
					},
					SequenceNumber: 35,
				},
			},
			payload: []byte(`{
				"type": "response.failed",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 35
			}`),
		},
		{
			name: "response.incomplete",
			event: StreamEvent{
				Incomplete: &ResponseIncompleteEvent{
					Type: StreamEventIncomplete,
					Response: Response{
						ID:                "resp-1",
						Object:            "response",
						CreatedAt:         1234567890,
						Model:             "gpt-4",
						Output:            []OutputItem{},
						ParallelToolCalls: false,
						Metadata:          map[string]string{},
						Tools:             []shared.ToolUnion{},
						ToolChoice:        nil,
					},
					SequenceNumber: 36,
				},
			},
			payload: []byte(`{
				"type": "response.incomplete",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 36
			}`),
		},
		{
			name: "response.queued",
			event: StreamEvent{
				Queued: &ResponseQueuedEvent{
					Type: StreamEventQueued,
					Response: Response{
						ID:                "resp-1",
						Object:            "response",
						CreatedAt:         1234567890,
						Model:             "gpt-4",
						Output:            []OutputItem{},
						ParallelToolCalls: false,
						Metadata:          map[string]string{},
						Tools:             []shared.ToolUnion{},
						ToolChoice:        nil,
					},
					SequenceNumber: 37,
				},
			},
			payload: []byte(`{
				"type": "response.queued",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 37
			}`),
		},
		// Output Item 事件
		{
			name: "response.output_item.added",
			event: StreamEvent{
				OutputItemAdded: &ResponseOutputItemAddedEvent{
					Type:        StreamEventOutputItemAdded,
					OutputIndex: 0,
					Item: OutputItem{
						Message: &OutputMessage{
							Type:    "message",
							ID:      "msg_1",
							Role:    "assistant",
							Content: []OutputMessageContent{},
							Status:  "in_progress",
						},
					},
					SequenceNumber: 38,
				},
			},
			payload: []byte(`{
				"type": "response.output_item.added",
				"output_index": 0,
				"item": {
					"type": "message",
					"id": "msg_1",
					"role": "assistant",
					"content": [],
					"status": "in_progress"
				},
				"sequence_number": 38
			}`),
		},
		{
			name: "response.output_item.done",
			event: StreamEvent{
				OutputItemDone: &ResponseOutputItemDoneEvent{
					Type:        StreamEventOutputItemDone,
					OutputIndex: 0,
					Item: OutputItem{
						Message: &OutputMessage{
							Type:    "message",
							ID:      "msg_1",
							Role:    "assistant",
							Content: []OutputMessageContent{},
							Status:  "completed",
						},
					},
					SequenceNumber: 39,
				},
			},
			payload: []byte(`{
				"type": "response.output_item.done",
				"output_index": 0,
				"item": {
					"type": "message",
					"id": "msg_1",
					"role": "assistant",
					"content": [],
					"status": "completed"
				},
				"sequence_number": 39
			}`),
		},
		// Content Part 事件
		{
			name: "response.content_part.added",
			event: StreamEvent{
				ContentPartAdded: &ResponseContentPartAddedEvent{
					Type:         StreamEventContentPartAdded,
					ItemID:       "item-9",
					OutputIndex:  0,
					ContentIndex: 0,
					Part: OutputContentPart{
						OutputText: &OutputTextContent{
							Type:        OutputMessageContentTypeOutputText,
							Text:        "test content",
							Annotations: []Annotation{},
						},
					},
					SequenceNumber: 40,
				},
			},
			payload: []byte(`{
				"type": "response.content_part.added",
				"item_id": "item-9",
				"output_index": 0,
				"content_index": 0,
				"part": {
					"type": "output_text",
					"text": "test content",
					"annotations": []
				},
				"sequence_number": 40
			}`),
		},
		{
			name: "response.content_part.done",
			event: StreamEvent{
				ContentPartDone: &ResponseContentPartDoneEvent{
					Type:         StreamEventContentPartDone,
					ItemID:       "item-9",
					OutputIndex:  0,
					ContentIndex: 0,
					Part: OutputContentPart{
						OutputText: &OutputTextContent{
							Type:        OutputMessageContentTypeOutputText,
							Text:        "test content",
							Annotations: []Annotation{},
						},
					},
					SequenceNumber: 41,
				},
			},
			payload: []byte(`{
				"type": "response.content_part.done",
				"item_id": "item-9",
				"output_index": 0,
				"content_index": 0,
				"part": {
					"type": "output_text",
					"text": "test content",
					"annotations": []
				},
				"sequence_number": 41
			}`),
		},
		// Output Text 事件
		{
			name: "response.output_text.delta",
			event: StreamEvent{
				OutputTextDelta: &ResponseOutputTextDeltaEvent{
					Type:           StreamEventOutputTextDelta,
					ItemID:         "item-9",
					OutputIndex:    0,
					ContentIndex:   0,
					Delta:          "text-delta",
					SequenceNumber: 42,
					Logprobs:       []ResponseLogProb{},
				},
			},
			payload: []byte(`{
				"type": "response.output_text.delta",
				"item_id": "item-9",
				"output_index": 0,
				"content_index": 0,
				"delta": "text-delta",
				"sequence_number": 42,
				"logprobs": []
			}`),
		},
		{
			name: "response.output_text.done",
			event: StreamEvent{
				OutputTextDone: &ResponseOutputTextDoneEvent{
					Type:           StreamEventOutputTextDone,
					ItemID:         "item-9",
					OutputIndex:    0,
					ContentIndex:   0,
					Text:           "full text",
					SequenceNumber: 43,
					Logprobs:       []ResponseLogProb{},
				},
			},
			payload: []byte(`{
				"type": "response.output_text.done",
				"item_id": "item-9",
				"output_index": 0,
				"content_index": 0,
				"text": "full text",
				"sequence_number": 43,
				"logprobs": []
			}`),
		},
		{
			name: "response.output_text.annotation.added",
			event: StreamEvent{
				OutputTextAnnotationAdded: &ResponseOutputTextAnnotationAddedEvent{
					Type:            StreamEventOutputTextAnnotationAdded,
					ItemID:          "item-9",
					OutputIndex:     0,
					ContentIndex:    0,
					AnnotationIndex: 0,
					Annotation: Annotation{
						FileCitation: &FileCitationAnnotation{
							Type:     AnnotationTypeFileCitation,
							FileID:   "file-1",
							Index:    0,
							Filename: "example.txt",
						},
					},
					SequenceNumber: 44,
				},
			},
			payload: []byte(`{
				"type": "response.output_text.annotation.added",
				"item_id": "item-9",
				"output_index": 0,
				"content_index": 0,
				"annotation_index": 0,
				"annotation": {
					"type": "file_citation",
					"file_id": "file-1",
					"index": 0,
					"filename": "example.txt"
				},
				"sequence_number": 44
			}`),
		},
		// Refusal 事件
		{
			name: "response.refusal.delta",
			event: StreamEvent{
				RefusalDelta: &ResponseRefusalDeltaEvent{
					Type:           StreamEventRefusalDelta,
					ItemID:         "item-10",
					OutputIndex:    0,
					ContentIndex:   0,
					Delta:          "refusal-delta",
					SequenceNumber: 45,
				},
			},
			payload: []byte(`{
				"type": "response.refusal.delta",
				"item_id": "item-10",
				"output_index": 0,
				"content_index": 0,
				"delta": "refusal-delta",
				"sequence_number": 45
			}`),
		},
		{
			name: "response.refusal.done",
			event: StreamEvent{
				RefusalDone: &ResponseRefusalDoneEvent{
					Type:           StreamEventRefusalDone,
					ItemID:         "item-10",
					OutputIndex:    0,
					ContentIndex:   0,
					Refusal:        "full refusal",
					SequenceNumber: 46,
				},
			},
			payload: []byte(`{
				"type": "response.refusal.done",
				"item_id": "item-10",
				"output_index": 0,
				"content_index": 0,
				"refusal": "full refusal",
				"sequence_number": 46
			}`),
		},
		// Reasoning Text 事件
		{
			name: "response.reasoning_text.delta",
			event: StreamEvent{
				ReasoningTextDelta: &ResponseReasoningTextDeltaEvent{
					Type:           StreamEventReasoningTextDelta,
					ItemID:         "item-11",
					OutputIndex:    0,
					ContentIndex:   0,
					Delta:          "reasoning-delta",
					SequenceNumber: 47,
				},
			},
			payload: []byte(`{
				"type": "response.reasoning_text.delta",
				"item_id": "item-11",
				"output_index": 0,
				"content_index": 0,
				"delta": "reasoning-delta",
				"sequence_number": 47
			}`),
		},
		{
			name: "response.reasoning_text.done",
			event: StreamEvent{
				ReasoningTextDone: &ResponseReasoningTextDoneEvent{
					Type:           StreamEventReasoningTextDone,
					ItemID:         "item-11",
					OutputIndex:    0,
					ContentIndex:   0,
					Text:           "full reasoning",
					SequenceNumber: 48,
				},
			},
			payload: []byte(`{
				"type": "response.reasoning_text.done",
				"item_id": "item-11",
				"output_index": 0,
				"content_index": 0,
				"text": "full reasoning",
				"sequence_number": 48
			}`),
		},
		// Reasoning Summary 事件
		{
			name: "response.reasoning_summary_part.added",
			event: StreamEvent{
				ReasoningSummaryPartAdded: &ResponseReasoningSummaryPartAddedEvent{
					Type:         StreamEventReasoningSummaryPartAdded,
					ItemID:       "item-12",
					OutputIndex:  0,
					SummaryIndex: 0,
					Part: OutputSummaryPart{
						Type: "summary_text",
						Text: "summary text",
					},
					SequenceNumber: 49,
				},
			},
			payload: []byte(`{
				"type": "response.reasoning_summary_part.added",
				"item_id": "item-12",
				"output_index": 0,
				"summary_index": 0,
				"part": {
					"type": "summary_text",
					"text": "summary text"
				},
				"sequence_number": 49
			}`),
		},
		{
			name: "response.reasoning_summary_part.done",
			event: StreamEvent{
				ReasoningSummaryPartDone: &ResponseReasoningSummaryPartDoneEvent{
					Type:         StreamEventReasoningSummaryPartDone,
					ItemID:       "item-12",
					OutputIndex:  0,
					SummaryIndex: 0,
					Part: OutputSummaryPart{
						Type: "summary_text",
						Text: "summary text",
					},
					SequenceNumber: 50,
				},
			},
			payload: []byte(`{
				"type": "response.reasoning_summary_part.done",
				"item_id": "item-12",
				"output_index": 0,
				"summary_index": 0,
				"part": {
					"type": "summary_text",
					"text": "summary text"
				},
				"sequence_number": 50
			}`),
		},
		{
			name: "response.reasoning_summary_text.delta",
			event: StreamEvent{
				ReasoningSummaryTextDelta: &ResponseReasoningSummaryTextDeltaEvent{
					Type:           StreamEventReasoningSummaryTextDelta,
					ItemID:         "item-12",
					OutputIndex:    0,
					SummaryIndex:   0,
					Delta:          "summary-delta",
					SequenceNumber: 51,
				},
			},
			payload: []byte(`{
				"type": "response.reasoning_summary_text.delta",
				"item_id": "item-12",
				"output_index": 0,
				"summary_index": 0,
				"delta": "summary-delta",
				"sequence_number": 51
			}`),
		},
		{
			name: "response.reasoning_summary_text.done",
			event: StreamEvent{
				ReasoningSummaryTextDone: &ResponseReasoningSummaryTextDoneEvent{
					Type:           StreamEventReasoningSummaryTextDone,
					ItemID:         "item-12",
					OutputIndex:    0,
					SummaryIndex:   0,
					Text:           "full summary",
					SequenceNumber: 52,
				},
			},
			payload: []byte(`{
				"type": "response.reasoning_summary_text.done",
				"item_id": "item-12",
				"output_index": 0,
				"summary_index": 0,
				"text": "full summary",
				"sequence_number": 52
			}`),
		},
		// Error 事件
		{
			name: "error",
			event: StreamEvent{
				Error: &ResponseErrorEvent{
					Type:           StreamEventError,
					Code:           strPtr("invalid_request"),
					Message:        "Invalid request",
					SequenceNumber: 1,
				},
			},
			payload: []byte(`{
				"type": "error",
				"code": "invalid_request",
				"message": "Invalid request",
				"sequence_number": 1
			}`),
		},
		// Keepalive 事件
		{
			name: "keepalive",
			event: StreamEvent{
				Keepalive: &ResponseKeepaliveEvent{
					Type:           StreamEventKeepalive,
					SequenceNumber: 2,
				},
			},
			payload: []byte(`{
				"type": "keepalive",
				"sequence_number": 2
			}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := json.Marshal(tt.event)
			if err != nil {
				t.Fatalf("序列化失败: %v", err)
			}

			// 比较原始 JSON 和序列化后的 JSON（使用 map 比较，避免字段顺序问题）
			originalMap := marshalToMap(t, tt.payload)
			marshaledMap := marshalToMap(t, marshaled)
			if !reflect.DeepEqual(originalMap, marshaledMap) {
				t.Errorf("序列化后 JSON 不一致:\n期望：%+v\n实际：%+v", originalMap, marshaledMap)
			}
		})
	}
}

// TestStreamEvent_Unmarshal_Null 测试 null 输入。
func TestStreamEvent_Unmarshal_Null(t *testing.T) {
	var event StreamEvent
	err := json.Unmarshal([]byte("null"), &event)
	if err != nil {
		t.Fatalf("反序列化 null 失败: %v", err)
	}

	// 验证所有字段都为 nil
	v := reflect.ValueOf(&event).Elem()
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).IsNil() {
			t.Errorf("期望所有字段为 nil，但字段 %s 不为 nil", v.Type().Field(i).Name)
		}
	}
}

// TestStreamEvent_Unmarshal_UnknownType 测试未知类型。
func TestStreamEvent_Unmarshal_UnknownType(t *testing.T) {
	var event StreamEvent
	err := json.Unmarshal([]byte(`{"type": "unknown.type"}`), &event)
	if err == nil {
		t.Fatal("期望返回错误，但返回 nil")
	}

	expectedMsg := "不支持的流式事件类型"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("期望错误消息包含 '%s'，实际: %v", expectedMsg, err)
	}
}

// TestStreamEvent_Unmarshal_InvalidJSON 测试无效 JSON。
// 注意：无效 JSON 在进入 StreamEvent.UnmarshalJSON 前就被标准库 encoding/json 拦截并返回语法错误，
// 因此不会触发"流式事件解析失败"包装。这里测试的是标准库的 JSON 语法错误。
func TestStreamEvent_Unmarshal_InvalidJSON(t *testing.T) {
	var event StreamEvent
	err := json.Unmarshal([]byte("{invalid json}"), &event)
	if err == nil {
		t.Fatal("期望返回错误，但返回 nil")
	}

	// 标准库会返回 JSON 语法错误，而不是自定义的"流式事件解析失败"
	expectedMsg := "invalid character"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("期望错误消息包含 '%s'，实际: %v", expectedMsg, err)
	}
}

// TestStreamEvent_Marshal_ZeroField 测试零字段序列化。
func TestStreamEvent_Marshal_ZeroField(t *testing.T) {
	event := StreamEvent{}
	result, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 零字段应该序列化为 null
	if string(result) != "null" {
		t.Errorf("期望序列化为 null，实际: %s", string(result))
	}
}

// TestStreamEvent_Marshal_MultipleFields 测试多字段序列化。
func TestStreamEvent_Marshal_MultipleFields(t *testing.T) {
	event := StreamEvent{
		AudioDelta: &ResponseAudioDeltaEvent{
			Type:           StreamEventAudioDelta,
			ResponseID:     "resp-1",
			Delta:          "audio-data",
			SequenceNumber: 1,
		},
		AudioDone: &ResponseAudioDoneEvent{
			Type:           StreamEventAudioDone,
			ResponseID:     "resp-1",
			SequenceNumber: 2,
		},
	}

	_, err := json.Marshal(event)
	if err == nil {
		t.Fatal("期望返回错误，但返回 nil")
	}

	expectedMsg := "流式事件只能设置一种类型"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("期望错误消息包含 '%s'，实际: %v", expectedMsg, err)
	}
}

// TestStreamEvent_RoundTrip 测试往返一致性。
func TestStreamEvent_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		payload  []byte
		expected any
	}{
		{
			name: "response.audio.delta",
			payload: []byte(`{
				"type": "response.audio.delta",
				"response_id": "resp-1",
				"delta": "audio-data",
				"sequence_number": 1
			}`),
			expected: &ResponseAudioDeltaEvent{
				Type:           StreamEventAudioDelta,
				ResponseID:     "resp-1",
				Delta:          "audio-data",
				SequenceNumber: 1,
			},
		},
		{
			name: "response.created",
			payload: []byte(`{
				"type": "response.created",
				"response": {
					"id": "resp-1",
					"object": "response",
					"created_at": 1234567890,
					"model": "gpt-4",
					"output": [],
					"parallel_tool_calls": false,
					"metadata": {},
					"tools": [],
					"tool_choice": null
				},
				"sequence_number": 32
			}`),
			expected: &ResponseCreatedEvent{
				Type: StreamEventCreated,
				Response: Response{
					ID:                "resp-1",
					Object:            "response",
					CreatedAt:         1234567890,
					Model:             "gpt-4",
					Output:            []OutputItem{},
					ParallelToolCalls: false,
					Metadata:          map[string]string{},
					Tools:             []shared.ToolUnion{},
					ToolChoice:        nil,
				},
				SequenceNumber: 32,
			},
		},
		{
			name: "error",
			payload: []byte(`{
				"type": "error",
				"code": "invalid_request",
				"message": "Invalid request",
				"sequence_number": 1
			}`),
			expected: &ResponseErrorEvent{
				Type:           StreamEventError,
				Code:           strPtr("invalid_request"),
				Message:        "Invalid request",
				SequenceNumber: 1,
			},
		},
		{
			name: "keepalive",
			payload: []byte(`{
				"type": "keepalive",
				"sequence_number": 2
			}`),
			expected: &ResponseKeepaliveEvent{
				Type:           StreamEventKeepalive,
				SequenceNumber: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roundTripEvent(t, tt.payload, tt.expected)
		})
	}
}

// TestStreamEvent_MarshalJSON_NilPointer 测试 nil 指针字段。
func TestStreamEvent_MarshalJSON_NilPointer(t *testing.T) {
	event := StreamEvent{
		AudioDelta: nil,
	}

	result, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 所有字段为 nil 应该序列化为 null
	if string(result) != "null" {
		t.Errorf("期望序列化为 null，实际: %s", string(result))
	}
}

// TestStreamEvent_UnmarshalJSON_WithOptionalFields 测试包含可选字段的反序列化。
func TestStreamEvent_UnmarshalJSON_WithOptionalFields(t *testing.T) {
	payload := []byte(`{
		"type": "response.output_text.delta",
		"item_id": "item-1",
		"output_index": 0,
		"content_index": 0,
		"delta": "text-delta",
		"sequence_number": 1,
		"logprobs": [],
		"obfuscation": "redacted"
	}`)

	var event StreamEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	if event.OutputTextDelta == nil {
		t.Fatal("期望 OutputTextDelta 被设置")
	}

	if event.OutputTextDelta.Obfuscation == nil {
		t.Fatal("期望 Obfuscation 被设置")
	}

	if *event.OutputTextDelta.Obfuscation != "redacted" {
		t.Errorf("期望 Obfuscation 为 'redacted'，实际: %s", *event.OutputTextDelta.Obfuscation)
	}
}

// TestStreamEvent_MarshalJSON_WithOptionalFields 测试包含可选字段的序列化。
func TestStreamEvent_MarshalJSON_WithOptionalFields(t *testing.T) {
	obfuscation := "redacted"
	event := StreamEvent{
		OutputTextDelta: &ResponseOutputTextDeltaEvent{
			Type:           StreamEventOutputTextDelta,
			ItemID:         "item-1",
			OutputIndex:    0,
			ContentIndex:   0,
			Delta:          "text-delta",
			SequenceNumber: 1,
			Logprobs:       []ResponseLogProb{},
			Obfuscation:    &obfuscation,
		},
	}

	result, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	resultMap := marshalToMap(t, result)
	if resultMap["obfuscation"] != "redacted" {
		t.Errorf("期望 obfuscation 为 'redacted'，实际: %v", resultMap["obfuscation"])
	}
}

// TestStreamEvent_ErrorHandling 测试错误处理。
func TestStreamEvent_ErrorHandling(t *testing.T) {
	t.Run("UnmarshalJSON 返回包装错误", func(t *testing.T) {
		var event StreamEvent
		err := json.Unmarshal([]byte(`{"type": "response.audio.delta", "response_id": 123}`), &event)
		if err == nil {
			t.Fatal("期望返回错误，但返回 nil")
		}

		// 验证错误是包装的
		expectedMsg := "response.audio.delta 事件解析失败"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("期望错误消息包含 '%s'，实际: %v", expectedMsg, err)
		}
	})

	t.Run("MarshalJSON 返回多字段错误", func(t *testing.T) {
		event := StreamEvent{
			AudioDelta: &ResponseAudioDeltaEvent{
				Type:           StreamEventAudioDelta,
				ResponseID:     "resp-1",
				Delta:          "audio-data",
				SequenceNumber: 1,
			},
			AudioDone: &ResponseAudioDoneEvent{
				Type:           StreamEventAudioDone,
				ResponseID:     "resp-1",
				SequenceNumber: 2,
			},
		}

		_, err := json.Marshal(event)
		if err == nil {
			t.Fatal("期望返回错误，但返回 nil")
		}

		// 验证错误消息
		expectedMsg := "流式事件只能设置一种类型"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("期望错误消息包含 '%s'，实际: %v", expectedMsg, err)
		}
	})
}

// TestStreamEvent_TypeConsistency 测试类型一致性。
func TestStreamEvent_TypeConsistency(t *testing.T) {
	tests := []struct {
		name         string
		payload      []byte
		expectedType StreamEventType
	}{
		{"response.audio.delta", []byte(`{"type": "response.audio.delta", "response_id": "resp-1", "delta": "data", "sequence_number": 1}`), StreamEventAudioDelta},
		{"response.created", []byte(`{"type": "response.created", "response": {"id": "resp-1", "object": "response", "created_at": 1234567890, "model": "gpt-4", "output": [], "parallel_tool_calls": false, "metadata": {}, "tools": [], "tool_choice": null}, "sequence_number": 1}`), StreamEventCreated},
		{"error", []byte(`{"type": "error", "code": "invalid_request", "message": "Invalid request", "sequence_number": 1}`), StreamEventError},
		{"keepalive", []byte(`{"type": "keepalive", "sequence_number": 2}`), StreamEventKeepalive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event StreamEvent
			if err := json.Unmarshal(tt.payload, &event); err != nil {
				t.Fatalf("反序列化失败: %v", err)
			}

			// 验证类型字段
			v := reflect.ValueOf(&event).Elem()
			for i := 0; i < v.NumField(); i++ {
				if !v.Field(i).IsNil() {
					// 获取嵌套结构的 Type 字段
					nestedValue := v.Field(i).Elem()
					typeField := nestedValue.FieldByName("Type")
					if typeField.IsValid() {
						actualType := typeField.Interface().(StreamEventType)
						if actualType != tt.expectedType {
							t.Errorf("期望类型为 %s，实际: %s", tt.expectedType, actualType)
						}
					}
					break
				}
			}
		})
	}
}
