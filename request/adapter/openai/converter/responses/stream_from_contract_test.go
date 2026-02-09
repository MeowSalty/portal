package responses

import (
	"testing"

	"github.com/MeowSalty/portal/logger"
	responsesTypes "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockIndexContext 是 StreamIndexContext 的模拟实现，用于测试索引补齐逻辑
type mockIndexContext struct {
	sequence      int
	itemIDMap     map[string]string
	outputIndex   int
	contentIndex  map[string]int
	annotationIdx map[string]int
	currentItemID string
}

func newMockIndexContext() *mockIndexContext {
	return &mockIndexContext{
		sequence:      0,
		itemIDMap:     make(map[string]string),
		outputIndex:   0,
		contentIndex:  make(map[string]int),
		annotationIdx: make(map[string]int),
	}
}

func (m *mockIndexContext) NextSequence() int {
	m.sequence++
	return m.sequence
}

func (m *mockIndexContext) EnsureItemID(key string) string {
	if id, ok := m.itemIDMap[key]; ok {
		return id
	}
	id := "item_" + key
	m.itemIDMap[key] = id
	return id
}

func (m *mockIndexContext) EnsureOutputIndex(responseID string) int {
	m.outputIndex++
	return m.outputIndex
}

func (m *mockIndexContext) EnsureContentIndex(itemID string, index int) int {
	key := itemID + "_content"
	if idx, ok := m.contentIndex[key]; ok {
		return idx
	}
	m.contentIndex[key] = 1
	return 1
}

func (m *mockIndexContext) EnsureAnnotationIndex(itemID string, index int) int {
	key := itemID + "_annotation"
	if idx, ok := m.annotationIdx[key]; ok {
		return idx
	}
	m.annotationIdx[key] = 1
	return 1
}

func (m *mockIndexContext) GetItemID() string {
	return m.currentItemID
}

func (m *mockIndexContext) GetMessageID() string {
	return ""
}

func (m *mockIndexContext) SetMessageID(messageID string) {
	// 空实现，测试中不需要
}

func (m *mockIndexContext) SetItemID(itemID string) {
	m.currentItemID = itemID
}

// TestStreamEventFormContract_NilInput 测试空输入处理
func TestStreamEventFormContract_NilInput(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	events, err := StreamEventFormContract(nil, log, indexCtx)

	assert.NoError(t, err)
	assert.Nil(t, events)
}

// TestStreamEventFormContract_UnknownEventType 测试未知事件类型被忽略
func TestStreamEventFormContract_UnknownEventType(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	// 创建一个不存在的事件类型（通过强制设置无效的 Type）
	contract := &adapterTypes.StreamEventContract{
		Type:   "unknown_event_type",
		Source: adapterTypes.StreamSourceOpenAIResponse,
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)

	// 未知事件类型应该返回 nil 和 nil error（被忽略）
	assert.NoError(t, err)
	assert.Nil(t, events)
}

// ============================================================
// 核心文本输出路径：事件顺序与完整性测试
// ============================================================

// TestStreamEventFormContract_TextOutputFlow 完整的文本输出流程测试
// 验证：response.created -> response.in_progress -> output_item.added -> content_part.added ->
//
//	output_text.delta* -> output_text.done -> content_part.done -> output_item.done -> response.completed
func TestStreamEventFormContract_TextOutputFlow(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()
	responseID := "resp_123"
	itemID := "msg_456"

	tests := []struct {
		name           string
		contract       *adapterTypes.StreamEventContract
		expectedType   string // 转换后的事件类型
		expectedSeq    int
		expectedFields map[string]interface{}
	}{
		{
			name: "response.created",
			contract: &adapterTypes.StreamEventContract{
				Type:           adapterTypes.StreamEventResponseCreated,
				Source:         adapterTypes.StreamSourceOpenAIResponse,
				ResponseID:     responseID,
				SequenceNumber: 1,
			},
			expectedType: "response.created",
			expectedSeq:  1,
			expectedFields: map[string]interface{}{
				"response_id": responseID,
			},
		},
		{
			name: "response.in_progress",
			contract: &adapterTypes.StreamEventContract{
				Type:           adapterTypes.StreamEventResponseInProgress,
				Source:         adapterTypes.StreamSourceOpenAIResponse,
				ResponseID:     responseID,
				SequenceNumber: 2,
			},
			expectedType: "response.in_progress",
			expectedSeq:  2,
			expectedFields: map[string]interface{}{
				"response_id": responseID,
			},
		},
		{
			name: "output_item.added",
			contract: &adapterTypes.StreamEventContract{
				Type:           adapterTypes.StreamEventOutputItemAdded,
				Source:         adapterTypes.StreamSourceOpenAIResponse,
				ResponseID:     responseID,
				ItemID:         itemID,
				SequenceNumber: 3,
				OutputIndex:    0,
				Content: &adapterTypes.StreamContentPayload{
					Raw: map[string]interface{}{
						"item": responsesTypes.OutputItem{
							Message: &responsesTypes.OutputMessage{
								ID:   itemID,
								Type: responsesTypes.OutputItemTypeMessage,
							},
						},
					},
				},
			},
			expectedType: "response.output_item.added",
			expectedSeq:  3,
			expectedFields: map[string]interface{}{
				"item_id":      itemID,
				"output_index": 0,
			},
		},
		{
			name: "content_part.added",
			contract: &adapterTypes.StreamEventContract{
				Type:           adapterTypes.StreamEventContentPartAdded,
				Source:         adapterTypes.StreamSourceOpenAIResponse,
				ResponseID:     responseID,
				ItemID:         itemID,
				SequenceNumber: 4,
				OutputIndex:    0,
				ContentIndex:   0,
			},
			expectedType: "response.content_part.added",
			expectedSeq:  4,
			expectedFields: map[string]interface{}{
				"item_id":       itemID,
				"output_index":  0,
				"content_index": 0,
			},
		},
		{
			name: "output_text.delta (1)",
			contract: &adapterTypes.StreamEventContract{
				Type:           adapterTypes.StreamEventOutputTextDelta,
				Source:         adapterTypes.StreamSourceOpenAIResponse,
				ResponseID:     responseID,
				ItemID:         itemID,
				SequenceNumber: 5,
				OutputIndex:    0,
				ContentIndex:   0,
				Delta: &adapterTypes.StreamDeltaPayload{
					Text: func() *string { s := "Hello"; return &s }(),
				},
			},
			expectedType: "response.output_text.delta",
			expectedSeq:  5,
			expectedFields: map[string]interface{}{
				"item_id":       itemID,
				"delta":         "Hello",
				"output_index":  0,
				"content_index": 0,
			},
		},
		{
			name: "output_text.delta (2)",
			contract: &adapterTypes.StreamEventContract{
				Type:           adapterTypes.StreamEventOutputTextDelta,
				Source:         adapterTypes.StreamSourceOpenAIResponse,
				ResponseID:     responseID,
				ItemID:         itemID,
				SequenceNumber: 6,
				OutputIndex:    0,
				ContentIndex:   0,
				Delta: &adapterTypes.StreamDeltaPayload{
					Text: func() *string { s := " world!"; return &s }(),
				},
			},
			expectedType: "response.output_text.delta",
			expectedSeq:  6,
			expectedFields: map[string]interface{}{
				"item_id":       itemID,
				"delta":         " world!",
				"output_index":  0,
				"content_index": 0,
			},
		},
		{
			name: "output_text.done",
			contract: &adapterTypes.StreamEventContract{
				Type:           adapterTypes.StreamEventOutputTextDone,
				Source:         adapterTypes.StreamSourceOpenAIResponse,
				ResponseID:     responseID,
				ItemID:         itemID,
				SequenceNumber: 7,
				OutputIndex:    0,
				ContentIndex:   0,
				Content: &adapterTypes.StreamContentPayload{
					Text: func() *string { s := "Hello world!"; return &s }(),
				},
			},
			expectedType: "response.output_text.done",
			expectedSeq:  7,
			expectedFields: map[string]interface{}{
				"item_id":       itemID,
				"text":          "Hello world!",
				"output_index":  0,
				"content_index": 0,
			},
		},
		{
			name: "content_part.done",
			contract: &adapterTypes.StreamEventContract{
				Type:           adapterTypes.StreamEventContentPartDone,
				Source:         adapterTypes.StreamSourceOpenAIResponse,
				ResponseID:     responseID,
				ItemID:         itemID,
				SequenceNumber: 8,
				OutputIndex:    0,
				ContentIndex:   0,
			},
			expectedType: "response.content_part.done",
			expectedSeq:  8,
			expectedFields: map[string]interface{}{
				"item_id":       itemID,
				"output_index":  0,
				"content_index": 0,
			},
		},
		{
			name: "output_item.done",
			contract: &adapterTypes.StreamEventContract{
				Type:           adapterTypes.StreamEventOutputItemDone,
				Source:         adapterTypes.StreamSourceOpenAIResponse,
				ResponseID:     responseID,
				ItemID:         itemID,
				SequenceNumber: 9,
				OutputIndex:    0,
				Content: &adapterTypes.StreamContentPayload{
					Raw: map[string]interface{}{
						"item": responsesTypes.OutputItem{
							Message: &responsesTypes.OutputMessage{
								ID:   itemID,
								Type: responsesTypes.OutputItemTypeMessage,
							},
						},
					},
				},
			},
			expectedType: "response.output_item.done",
			expectedSeq:  9,
			expectedFields: map[string]interface{}{
				"item_id":      itemID,
				"output_index": 0,
			},
		},
		{
			name: "response.completed",
			contract: &adapterTypes.StreamEventContract{
				Type:           adapterTypes.StreamEventResponseCompleted,
				Source:         adapterTypes.StreamSourceOpenAIResponse,
				ResponseID:     responseID,
				SequenceNumber: 10,
				Usage: &adapterTypes.StreamUsagePayload{
					InputTokens:  func() *int { i := 10; return &i }(),
					OutputTokens: func() *int { i := 5; return &i }(),
					TotalTokens:  func() *int { i := 15; return &i }(),
				},
			},
			expectedType: "response.completed",
			expectedSeq:  10,
			expectedFields: map[string]interface{}{
				"response_id": responseID,
			},
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			events, err := StreamEventFormContract(tc.contract, log, indexCtx)

			assert.NoError(t, err, "测试用例 %d (%s) 不应返回错误", i, tc.name)
			require.NotNil(t, events, "测试用例 %d (%s) 应返回事件", i, tc.name)
			require.Len(t, events, 1, "测试用例 %d (%s) 应返回单个事件", i, tc.name)

			event := events[0]

			// 验证事件类型
			var eventType string
			switch {
			case event.Created != nil:
				eventType = "response.created"
			case event.InProgress != nil:
				eventType = "response.in_progress"
			case event.Completed != nil:
				eventType = "response.completed"
			case event.OutputItemAdded != nil:
				eventType = "response.output_item.added"
			case event.OutputItemDone != nil:
				eventType = "response.output_item.done"
			case event.ContentPartAdded != nil:
				eventType = "response.content_part.added"
			case event.ContentPartDone != nil:
				eventType = "response.content_part.done"
			case event.OutputTextDelta != nil:
				eventType = "response.output_text.delta"
			case event.OutputTextDone != nil:
				eventType = "response.output_text.done"
			}

			assert.Equal(t, tc.expectedType, eventType, "事件类型不匹配")

			// 验证序列号
			assert.Equal(t, tc.expectedSeq, getEventSequenceNumber(event), "序列号不匹配")

			// 验证字段
			for field, expectedValue := range tc.expectedFields {
				actualValue := getEventField(event, field)
				assert.Equal(t, expectedValue, actualValue, "字段 %s 不匹配", field)
			}
		})
	}
}

// TestStreamEventFormContract_TextOutputFlow_SequenceMonotonic 测试序列号单调递增
func TestStreamEventFormContract_TextOutputFlow_SequenceMonotonic(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contracts := []adapterTypes.StreamEventContract{
		{Type: adapterTypes.StreamEventResponseCreated, SequenceNumber: 1},
		{Type: adapterTypes.StreamEventResponseInProgress, SequenceNumber: 2},
		{Type: adapterTypes.StreamEventOutputItemAdded, SequenceNumber: 3},
		{Type: adapterTypes.StreamEventContentPartAdded, SequenceNumber: 4},
		{Type: adapterTypes.StreamEventOutputTextDelta, SequenceNumber: 5},
		{Type: adapterTypes.StreamEventOutputTextDone, SequenceNumber: 6},
		{Type: adapterTypes.StreamEventContentPartDone, SequenceNumber: 7},
		{Type: adapterTypes.StreamEventOutputItemDone, SequenceNumber: 8},
		{Type: adapterTypes.StreamEventResponseCompleted, SequenceNumber: 9},
	}

	var previousSeq int
	for i, contract := range contracts {
		events, err := StreamEventFormContract(&contract, log, indexCtx)
		require.NoError(t, err)
		require.Len(t, events, 1)

		currentSeq := getEventSequenceNumber(events[0])

		if i > 0 {
			assert.Greater(t, currentSeq, previousSeq,
				"事件 %d 的序列号 %d 应大于前一个事件的序列号 %d", i, currentSeq, previousSeq)
		}
		previousSeq = currentSeq
	}
}

// ============================================================
// 降级拆分事件：序列号递增校验测试
// ============================================================

// TestStreamEventFormContract_DegradedMessageStop 测试 message_stop 降级为两个事件
func TestStreamEventFormContract_DegradedMessageStop(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventMessageStop,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     "resp_123",
		SequenceNumber: 10,
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Len(t, events, 2, "message_stop 应降级为 2 个事件")

	// 第一个事件：output_text.done
	assert.NotNil(t, events[0].OutputTextDone, "第一个事件应为 output_text.done")
	seq1 := getEventSequenceNumber(events[0])

	// 第二个事件：response.completed
	assert.NotNil(t, events[1].Completed, "第二个事件应为 response.completed")
	seq2 := getEventSequenceNumber(events[1])

	// 验证序列号递增
	assert.Equal(t, 10, seq1, "第一个事件的序列号应为 10")
	assert.Equal(t, 11, seq2, "第二个事件的序列号应为 11（递增）")
}

// TestStreamEventFormContract_DegradedContentBlockStart 测试 content_block_start 降级为两个事件
func TestStreamEventFormContract_DegradedContentBlockStart(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventContentBlockStart,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     "resp_123",
		SequenceNumber: 5,
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Len(t, events, 2, "content_block_start 应降级为 2 个事件")

	// 第一个事件：output_item.added
	assert.NotNil(t, events[0].OutputItemAdded, "第一个事件应为 output_item.added")
	seq1 := getEventSequenceNumber(events[0])

	// 第二个事件：content_part.added
	assert.NotNil(t, events[1].ContentPartAdded, "第二个事件应为 content_part.added")
	seq2 := getEventSequenceNumber(events[1])

	// 验证序列号递增
	assert.Equal(t, 5, seq1, "第一个事件的序列号应为 5")
	assert.Equal(t, 6, seq2, "第二个事件的序列号应为 6（递增）")
}

// TestStreamEventFormContract_DegradedContentBlockStop 测试 content_block_stop 降级为两个事件
func TestStreamEventFormContract_DegradedContentBlockStop(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventContentBlockStop,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     "resp_123",
		SequenceNumber: 8,
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Len(t, events, 2, "content_block_stop 应降级为 2 个事件")

	// 第一个事件：content_part.done
	assert.NotNil(t, events[0].ContentPartDone, "第一个事件应为 content_part.done")
	seq1 := getEventSequenceNumber(events[0])

	// 第二个事件：output_item.done
	assert.NotNil(t, events[1].OutputItemDone, "第二个事件应为 output_item.done")
	seq2 := getEventSequenceNumber(events[1])

	// 验证序列号递增
	assert.Equal(t, 8, seq1, "第一个事件的序列号应为 8")
	assert.Equal(t, 9, seq2, "第二个事件的序列号应为 9（递增）")
}

// TestStreamEventFormContract_DegradedSequenceWithZeroInput 测试序列号为 0 时的补齐
func TestStreamEventFormContract_DegradedSequenceWithZeroInput(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventContentBlockStart,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     "resp_123",
		SequenceNumber: 0, // 序列号为 0
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Len(t, events, 2)

	// 验证两个事件的序列号都来自 indexCtx.NextSequence()
	seq1 := getEventSequenceNumber(events[0])
	seq2 := getEventSequenceNumber(events[1])

	assert.Equal(t, 1, seq1, "第一个事件的序列号应由 indexCtx 生成，从 1 开始")
	assert.Equal(t, 2, seq2, "第二个事件的序列号应递增为 2")
}

// ============================================================
// 索引补齐测试
// ============================================================

// TestStreamEventFormContract_EnsureSequenceNumber 测试序列号补齐
func TestStreamEventFormContract_EnsureSequenceNumber(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventResponseCreated,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     "resp_123",
		SequenceNumber: 0, // 缺失序列号
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	seq := getEventSequenceNumber(events[0])
	assert.Equal(t, 1, seq, "序列号应被补齐为 1")
}

// TestStreamEventFormContract_EnsureItemID 测试 item_id 补齐
func TestStreamEventFormContract_EnsureItemID(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputItemAdded,
		Source:         adapterTypes.StreamSourceGemini,
		ResponseID:     "resp_123",
		SequenceNumber: 1,
		ItemID:         "", // 缺失 item_id
		OutputIndex:    0,
		Content: &adapterTypes.StreamContentPayload{
			Raw: map[string]interface{}{
				"item": responsesTypes.OutputItem{
					Message: &responsesTypes.OutputMessage{
						Type: responsesTypes.OutputItemTypeMessage,
					},
				},
			},
		},
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputItemAdded)
	itemID := events[0].OutputItemAdded.Item.Message.ID
	assert.NotEmpty(t, itemID, "item_id 应被补齐")
}

// TestStreamEventFormContract_EnsureOutputIndex 测试 output_index 补齐
func TestStreamEventFormContract_EnsureOutputIndex(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputTextDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     "resp_123",
		ItemID:         "msg_456",
		SequenceNumber: 1,
		OutputIndex:    0, // 缺失 output_index（但类型明确是 output_text，会自动补齐）
		ContentIndex:   0,
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputTextDelta)
	// output_index 应该被保留或补齐
	outputIndex := events[0].OutputTextDelta.OutputIndex
	assert.True(t, outputIndex >= 0, "output_index 应被有效设置")
}

// TestStreamEventFormContract_EnsureContentIndex 测试 content_index 补齐
func TestStreamEventFormContract_EnsureContentIndex(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputTextDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     "resp_123",
		ItemID:         "msg_456",
		SequenceNumber: 1,
		OutputIndex:    0,
		ContentIndex:   0, // 缺失 content_index（但类型明确是 output_text，会自动补齐）
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputTextDelta)
	contentIndex := events[0].OutputTextDelta.ContentIndex
	assert.True(t, contentIndex >= 0, "content_index 应被有效设置")
}

// ============================================================
// 工具函数
// ============================================================

// getEventSequenceNumber 从事件中提取序列号
func getEventSequenceNumber(event *responsesTypes.StreamEvent) int {
	switch {
	case event.Created != nil:
		return event.Created.SequenceNumber
	case event.InProgress != nil:
		return event.InProgress.SequenceNumber
	case event.Completed != nil:
		return event.Completed.SequenceNumber
	case event.Failed != nil:
		return event.Failed.SequenceNumber
	case event.Incomplete != nil:
		return event.Incomplete.SequenceNumber
	case event.Queued != nil:
		return event.Queued.SequenceNumber
	case event.OutputItemAdded != nil:
		return event.OutputItemAdded.SequenceNumber
	case event.OutputItemDone != nil:
		return event.OutputItemDone.SequenceNumber
	case event.ContentPartAdded != nil:
		return event.ContentPartAdded.SequenceNumber
	case event.ContentPartDone != nil:
		return event.ContentPartDone.SequenceNumber
	case event.OutputTextDelta != nil:
		return event.OutputTextDelta.SequenceNumber
	case event.OutputTextDone != nil:
		return event.OutputTextDone.SequenceNumber
	case event.OutputTextAnnotationAdded != nil:
		return event.OutputTextAnnotationAdded.SequenceNumber
	case event.RefusalDelta != nil:
		return event.RefusalDelta.SequenceNumber
	case event.RefusalDone != nil:
		return event.RefusalDone.SequenceNumber
	case event.ReasoningTextDelta != nil:
		return event.ReasoningTextDelta.SequenceNumber
	case event.ReasoningTextDone != nil:
		return event.ReasoningTextDone.SequenceNumber
	case event.ReasoningSummaryPartAdded != nil:
		return event.ReasoningSummaryPartAdded.SequenceNumber
	case event.ReasoningSummaryPartDone != nil:
		return event.ReasoningSummaryPartDone.SequenceNumber
	case event.ReasoningSummaryTextDelta != nil:
		return event.ReasoningSummaryTextDelta.SequenceNumber
	case event.ReasoningSummaryTextDone != nil:
		return event.ReasoningSummaryTextDone.SequenceNumber
	case event.FunctionCallArgumentsDelta != nil:
		return event.FunctionCallArgumentsDelta.SequenceNumber
	case event.FunctionCallArgumentsDone != nil:
		return event.FunctionCallArgumentsDone.SequenceNumber
	case event.CustomToolCallInputDelta != nil:
		return event.CustomToolCallInputDelta.SequenceNumber
	case event.CustomToolCallInputDone != nil:
		return event.CustomToolCallInputDone.SequenceNumber
	case event.MCPCallArgumentsDelta != nil:
		return event.MCPCallArgumentsDelta.SequenceNumber
	case event.MCPCallArgumentsDone != nil:
		return event.MCPCallArgumentsDone.SequenceNumber
	case event.MCPCallCompleted != nil:
		return event.MCPCallCompleted.SequenceNumber
	case event.MCPCallFailed != nil:
		return event.MCPCallFailed.SequenceNumber
	case event.MCPCallInProgress != nil:
		return event.MCPCallInProgress.SequenceNumber
	case event.MCPListToolsCompleted != nil:
		return event.MCPListToolsCompleted.SequenceNumber
	case event.MCPListToolsFailed != nil:
		return event.MCPListToolsFailed.SequenceNumber
	case event.MCPListToolsInProgress != nil:
		return event.MCPListToolsInProgress.SequenceNumber
	case event.AudioDelta != nil:
		return event.AudioDelta.SequenceNumber
	case event.AudioDone != nil:
		return event.AudioDone.SequenceNumber
	case event.AudioTranscriptDelta != nil:
		return event.AudioTranscriptDelta.SequenceNumber
	case event.AudioTranscriptDone != nil:
		return event.AudioTranscriptDone.SequenceNumber
	case event.CodeInterpreterCallCodeDelta != nil:
		return event.CodeInterpreterCallCodeDelta.SequenceNumber
	case event.CodeInterpreterCallCodeDone != nil:
		return event.CodeInterpreterCallCodeDone.SequenceNumber
	case event.CodeInterpreterCallCompleted != nil:
		return event.CodeInterpreterCallCompleted.SequenceNumber
	case event.CodeInterpreterCallInProgress != nil:
		return event.CodeInterpreterCallInProgress.SequenceNumber
	case event.CodeInterpreterCallInterpreting != nil:
		return event.CodeInterpreterCallInterpreting.SequenceNumber
	case event.FileSearchCallCompleted != nil:
		return event.FileSearchCallCompleted.SequenceNumber
	case event.FileSearchCallInProgress != nil:
		return event.FileSearchCallInProgress.SequenceNumber
	case event.FileSearchCallSearching != nil:
		return event.FileSearchCallSearching.SequenceNumber
	case event.WebSearchCallCompleted != nil:
		return event.WebSearchCallCompleted.SequenceNumber
	case event.WebSearchCallInProgress != nil:
		return event.WebSearchCallInProgress.SequenceNumber
	case event.WebSearchCallSearching != nil:
		return event.WebSearchCallSearching.SequenceNumber
	case event.ImageGenCallCompleted != nil:
		return event.ImageGenCallCompleted.SequenceNumber
	case event.ImageGenCallGenerating != nil:
		return event.ImageGenCallGenerating.SequenceNumber
	case event.ImageGenCallInProgress != nil:
		return event.ImageGenCallInProgress.SequenceNumber
	case event.ImageGenCallPartialImage != nil:
		return event.ImageGenCallPartialImage.SequenceNumber
	case event.Error != nil:
		return event.Error.SequenceNumber
	default:
		return 0
	}
}

// getEventField 从事件中提取指定字段的值
func getEventField(event *responsesTypes.StreamEvent, field string) interface{} {
	switch field {
	case "response_id":
		switch {
		case event.Created != nil:
			return event.Created.Response.ID
		case event.InProgress != nil:
			return event.InProgress.Response.ID
		case event.Completed != nil:
			return event.Completed.Response.ID
		case event.Failed != nil:
			return event.Failed.Response.ID
		case event.Incomplete != nil:
			return event.Incomplete.Response.ID
		case event.Queued != nil:
			return event.Queued.Response.ID
		}
	case "item_id":
		switch {
		case event.OutputItemAdded != nil:
			return event.OutputItemAdded.Item.Message.ID
		case event.OutputItemDone != nil:
			return event.OutputItemDone.Item.Message.ID
		case event.ContentPartAdded != nil:
			return event.ContentPartAdded.ItemID
		case event.ContentPartDone != nil:
			return event.ContentPartDone.ItemID
		case event.OutputTextDelta != nil:
			return event.OutputTextDelta.ItemID
		case event.OutputTextDone != nil:
			return event.OutputTextDone.ItemID
		}
	case "output_index":
		switch {
		case event.OutputItemAdded != nil:
			return event.OutputItemAdded.OutputIndex
		case event.OutputItemDone != nil:
			return event.OutputItemDone.OutputIndex
		case event.ContentPartAdded != nil:
			return event.ContentPartAdded.OutputIndex
		case event.ContentPartDone != nil:
			return event.ContentPartDone.OutputIndex
		case event.OutputTextDelta != nil:
			return event.OutputTextDelta.OutputIndex
		case event.OutputTextDone != nil:
			return event.OutputTextDone.OutputIndex
		}
	case "content_index":
		switch {
		case event.ContentPartAdded != nil:
			return event.ContentPartAdded.ContentIndex
		case event.ContentPartDone != nil:
			return event.ContentPartDone.ContentIndex
		case event.OutputTextDelta != nil:
			return event.OutputTextDelta.ContentIndex
		case event.OutputTextDone != nil:
			return event.OutputTextDone.ContentIndex
		}
	case "delta":
		if event.OutputTextDelta != nil {
			return event.OutputTextDelta.Delta
		}
		if event.RefusalDelta != nil {
			return event.RefusalDelta.Delta
		}
		if event.ReasoningTextDelta != nil {
			return event.ReasoningTextDelta.Delta
		}
	case "text":
		if event.OutputTextDone != nil {
			return event.OutputTextDone.Text
		}
		if event.RefusalDone != nil {
			return event.RefusalDone.Refusal
		}
		if event.ReasoningTextDone != nil {
			return event.ReasoningTextDone.Text
		}
	}
	return nil
}

// TestStreamEventFormContract_WithLogprobs 测试 logprobs 字段完整性
func TestStreamEventFormContract_WithLogprobs(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	logprobs := []responsesTypes.ResponseLogProb{
		{Token: "Hello", Logprob: -0.1, TopLogprobs: []responsesTypes.ResponseTopLogProb{}},
		{Token: " ", Logprob: -0.01, TopLogprobs: []responsesTypes.ResponseTopLogProb{}},
	}

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputTextDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     "resp_123",
		ItemID:         "msg_456",
		SequenceNumber: 1,
		OutputIndex:    0,
		ContentIndex:   0,
		Delta: &adapterTypes.StreamDeltaPayload{
			Text: func() *string { s := "Hello "; return &s }(),
		},
		Extensions: map[string]interface{}{
			"openai_responses": map[string]interface{}{
				"logprobs": logprobs,
			},
		},
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputTextDelta)
	assert.NotNil(t, events[0].OutputTextDelta.Logprobs, "logprobs 字段不应为 nil")
	assert.Len(t, events[0].OutputTextDelta.Logprobs, 2, "logprobs 应包含 2 个元素")
	assert.Equal(t, "Hello", events[0].OutputTextDelta.Logprobs[0].Token)
}

// TestStreamEventFormContract_LogprobsNilHandling 测试 logprobs 为 nil 时输出空数组
func TestStreamEventFormContract_LogprobsNilHandling(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputTextDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     "resp_123",
		ItemID:         "msg_456",
		SequenceNumber: 1,
		OutputIndex:    0,
		ContentIndex:   0,
		Delta: &adapterTypes.StreamDeltaPayload{
			Text: func() *string { s := "Hello"; return &s }(),
		},
		// 不提供 logprobs
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputTextDelta)
	// logprobs 应为空数组而非 nil
	assert.NotNil(t, events[0].OutputTextDelta.Logprobs, "logprobs 字段不应为 nil")
	assert.Empty(t, events[0].OutputTextDelta.Logprobs, "logprobs 应为空数组")
}

// ============================================================
// 索引修复逻辑测试
// ============================================================

// TestStreamEventFormContract_IndexZeroPreserved 测试原生事件的 0 值索引保持不变
func TestStreamEventFormContract_IndexZeroPreserved(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputTextDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse, // 原生事件
		ResponseID:     "resp_123",
		ItemID:         "msg_456",
		SequenceNumber: 1,
		OutputIndex:    0, // 合法的 0 值
		ContentIndex:   0, // 合法的 0 值
		Delta: &adapterTypes.StreamDeltaPayload{
			Text: func() *string { s := "Hello"; return &s }(),
		},
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputTextDelta)
	// 原生事件的 0 值应该保持不变
	assert.Equal(t, 0, events[0].OutputTextDelta.OutputIndex, "原生事件的 output_index 0 应保持不变")
	assert.Equal(t, 0, events[0].OutputTextDelta.ContentIndex, "原生事件的 content_index 0 应保持不变")
}

// TestStreamEventFormContract_NegativeIndexFilled 测试非原生事件的负值索引触发补齐
func TestStreamEventFormContract_NegativeIndexFilled(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputTextDelta,
		Source:         adapterTypes.StreamSourceGemini, // 非原生事件
		ResponseID:     "resp_123",
		ItemID:         "msg_456",
		SequenceNumber: 1,
		OutputIndex:    -1, // 负值，应触发补齐
		ContentIndex:   -1, // 负值，应触发补齐
		Delta: &adapterTypes.StreamDeltaPayload{
			Text: func() *string { s := "Hello"; return &s }(),
		},
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputTextDelta)
	// 负值应该被补齐为正数
	assert.Greater(t, events[0].OutputTextDelta.OutputIndex, 0, "负值 output_index 应被补齐为正数")
	assert.Greater(t, events[0].OutputTextDelta.ContentIndex, 0, "负值 content_index 应被补齐为正数")
}

// TestStreamEventFormContract_NonNativeEventZeroFilled 测试非原生事件的负值索引触发补齐
func TestStreamEventFormContract_NonNativeEventZeroFilled(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputTextDelta,
		Source:         adapterTypes.StreamSourceGemini, // 非原生事件
		ResponseID:     "resp_123",
		ItemID:         "msg_456",
		SequenceNumber: 1,
		OutputIndex:    -1, // 非原生事件的负值应触发补齐
		ContentIndex:   -1, // 非原生事件的负值应触发补齐
		Delta: &adapterTypes.StreamDeltaPayload{
			Text: func() *string { s := "Hello"; return &s }(),
		},
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputTextDelta)
	// 负值应该被补齐为正数
	assert.Greater(t, events[0].OutputTextDelta.OutputIndex, 0, "负值 output_index 应被补齐为正数")
	assert.Greater(t, events[0].OutputTextDelta.ContentIndex, 0, "负值 content_index 应被补齐为正数")
}

// TestStreamEventFormContract_MessageIDBackfill 测试 Message 存在但 ID 为空时回填
func TestStreamEventFormContract_MessageIDBackfill(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputItemAdded,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     "resp_123",
		ItemID:         "msg_456",
		SequenceNumber: 1,
		OutputIndex:    0,
		Content: &adapterTypes.StreamContentPayload{
			Raw: map[string]interface{}{
				"item": responsesTypes.OutputItem{
					Message: &responsesTypes.OutputMessage{
						ID:   "", // ID 为空，应回填
						Type: responsesTypes.OutputItemTypeMessage,
					},
				},
			},
		},
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputItemAdded)
	// Message.ID 应被回填为 contract.ItemID
	assert.Equal(t, "msg_456", events[0].OutputItemAdded.Item.Message.ID, "Message.ID 应被回填为 contract.ItemID")
}

// TestStreamEventFormContract_MessageIDBackfillDone 测试 output_item.done 事件的 Message ID 回填
func TestStreamEventFormContract_MessageIDBackfillDone(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputItemDone,
		Source:         adapterTypes.StreamSourceOpenAIResponse,
		ResponseID:     "resp_123",
		ItemID:         "msg_456",
		SequenceNumber: 1,
		OutputIndex:    0,
		Content: &adapterTypes.StreamContentPayload{
			Raw: map[string]interface{}{
				"item": responsesTypes.OutputItem{
					Message: &responsesTypes.OutputMessage{
						ID:   "", // ID 为空，应回填
						Type: responsesTypes.OutputItemTypeMessage,
					},
				},
			},
		},
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputItemDone)
	// Message.ID 应被回填为 contract.ItemID
	assert.Equal(t, "msg_456", events[0].OutputItemDone.Item.Message.ID, "Message.ID 应被回填为 contract.ItemID")
}

// TestStreamEventFormContract_NativeEventItemIDNotFilled 测试原生事件的 item_id 不被补齐
func TestStreamEventFormContract_NativeEventItemIDNotFilled(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputTextDelta,
		Source:         adapterTypes.StreamSourceOpenAIResponse, // 原生事件
		ResponseID:     "resp_123",
		ItemID:         "", // item_id 为空
		SequenceNumber: 1,
		OutputIndex:    0,
		ContentIndex:   0,
		Delta: &adapterTypes.StreamDeltaPayload{
			Text: func() *string { s := "Hello"; return &s }(),
		},
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputTextDelta)
	// 原生事件的空 item_id 不应被补齐
	assert.Empty(t, events[0].OutputTextDelta.ItemID, "原生事件的空 item_id 不应被补齐")
}

// TestStreamEventFormContract_NonNativeEventItemIDFilled 测试非原生事件的 item_id 被补齐
func TestStreamEventFormContract_NonNativeEventItemIDFilled(t *testing.T) {
	log := logger.NewNopLogger()
	indexCtx := newMockIndexContext()

	contract := &adapterTypes.StreamEventContract{
		Type:           adapterTypes.StreamEventOutputTextDelta,
		Source:         adapterTypes.StreamSourceGemini, // 非原生事件
		ResponseID:     "resp_123",
		ItemID:         "", // item_id 为空，应被补齐
		SequenceNumber: 1,
		OutputIndex:    -1,
		ContentIndex:   -1,
		Delta: &adapterTypes.StreamDeltaPayload{
			Text: func() *string { s := "Hello"; return &s }(),
		},
	}

	events, err := StreamEventFormContract(contract, log, indexCtx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	require.NotNil(t, events[0].OutputTextDelta)
	// 非原生事件的空 item_id 应被补齐
	assert.NotEmpty(t, events[0].OutputTextDelta.ItemID, "非原生事件的空 item_id 应被补齐")
}
