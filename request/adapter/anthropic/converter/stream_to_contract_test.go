package converter

import (
	"testing"

	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/stretchr/testify/assert"
)

// TestStreamEventToContract_NilEvent 测试 nil 事件的处理
func TestStreamEventToContract_NilEvent(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	contract, err := StreamEventToContract(nil, ctx, log)
	if err != nil {
		t.Fatalf("期望返回 nil, 但得到错误: %v", err)
	}
	if contract != nil {
		t.Fatal("nil 事件应返回 nil")
	}
}

// TestStreamEventToContract_SequenceNumberPositive 测试 sequence_number > 0
func TestStreamEventToContract_SequenceNumberPositive(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	messageID := "msg_123"
	text := "Hello, world!"

	event := &anthropicTypes.StreamEvent{
		MessageStart: &anthropicTypes.MessageStartEvent{
			Type: anthropicTypes.StreamEventMessageStart,
			Message: anthropicTypes.Response{
				ID:   messageID,
				Type: anthropicTypes.ResponseTypeMessage,
				Role: anthropicTypes.RoleAssistant,
				Content: []anthropicTypes.ResponseContentBlock{
					{
						Text: &anthropicTypes.TextBlock{
							Type: anthropicTypes.ResponseContentBlockText,
							Text: text,
						},
					},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证 sequence_number > 0
	if contract.SequenceNumber <= 0 {
		t.Errorf("SequenceNumber = %d, 期望 > 0", contract.SequenceNumber)
	}
}

// TestStreamEventToContract_ItemIDConsistency 测试同一流内 item_id 一致性
func TestStreamEventToContract_ItemIDConsistency(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	messageID := "msg_123"

	// 模拟同一流内的多个事件
	events := []*anthropicTypes.StreamEvent{
		{
			MessageStart: &anthropicTypes.MessageStartEvent{
				Type: anthropicTypes.StreamEventMessageStart,
				Message: anthropicTypes.Response{
					ID:   messageID,
					Type: anthropicTypes.ResponseTypeMessage,
					Role: anthropicTypes.RoleAssistant,
					Content: []anthropicTypes.ResponseContentBlock{
						{
							Text: &anthropicTypes.TextBlock{
								Type: anthropicTypes.ResponseContentBlockText,
								Text: "Hello",
							},
						},
					},
				},
			},
		},
		{
			MessageDelta: &anthropicTypes.MessageDeltaEvent{
				Type: anthropicTypes.StreamEventMessageDelta,
			},
		},
		{
			MessageStop: &anthropicTypes.MessageStopEvent{
				Type: anthropicTypes.StreamEventMessageStop,
			},
		},
	}

	var itemIDs []string
	for _, event := range events {
		contract, err := StreamEventToContract(event, ctx, log)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}
		if contract != nil {
			itemIDs = append(itemIDs, contract.ItemID)
		}
	}

	// 验证 item_id 非空
	for i, itemID := range itemIDs {
		if itemID == "" {
			t.Errorf("事件 %d 的 ItemID 为空", i)
		}
	}

	// 验证同一流内 item_id 一致（message_start 和 message_stop）
	if len(itemIDs) >= 2 {
		if itemIDs[0] != itemIDs[len(itemIDs)-1] {
			t.Errorf("item_id 不一致: 首个事件 = %s, 最后事件 = %s",
				itemIDs[0], itemIDs[len(itemIDs)-1])
		}
	}
}

// TestStreamEventToContract_ContentIndexNoRegression 测试 content_index 不发生回退
func TestStreamEventToContract_ContentIndexNoRegression(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	// 模拟包含多个 content block 的事件流
	events := []*anthropicTypes.StreamEvent{
		{
			ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
				Type:  anthropicTypes.StreamEventContentBlockStart,
				Index: 0,
				ContentBlock: anthropicTypes.ResponseContentBlock{
					Text: &anthropicTypes.TextBlock{
						Type: anthropicTypes.ResponseContentBlockText,
						Text: "Block 1",
					},
				},
			},
		},
		{
			ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
				Type:  anthropicTypes.StreamEventContentBlockStart,
				Index: 1,
				ContentBlock: anthropicTypes.ResponseContentBlock{
					Text: &anthropicTypes.TextBlock{
						Type: anthropicTypes.ResponseContentBlockText,
						Text: "Block 2",
					},
				},
			},
		},
		{
			ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
				Type:  anthropicTypes.StreamEventContentBlockStart,
				Index: 2,
				ContentBlock: anthropicTypes.ResponseContentBlock{
					Text: &anthropicTypes.TextBlock{
						Type: anthropicTypes.ResponseContentBlockText,
						Text: "Block 3",
					},
				},
			},
		},
	}

	previousContentIndex := -1
	for i, event := range events {
		contract, err := StreamEventToContract(event, ctx, log)
		if err != nil {
			t.Fatalf("转换事件 %d 失败: %v", i, err)
		}

		// 验证 content_index >= 0
		if contract.ContentIndex < 0 {
			t.Errorf("事件 %d 的 ContentIndex = %d, 应 >= 0", i, contract.ContentIndex)
		}

		// 验证不发生回退（除了第一次）
		if i > 0 && contract.ContentIndex < previousContentIndex {
			t.Errorf("事件 %d 的 ContentIndex (%d) 回退，前值为 %d",
				i, contract.ContentIndex, previousContentIndex)
		}

		previousContentIndex = contract.ContentIndex
	}
}

// TestStreamEventToContract_NilLogger 测试 nil 日志记录器的处理
func TestStreamEventToContract_NilLogger(t *testing.T) {
	ctx := types.NewStreamIndexContext()
	event := &anthropicTypes.StreamEvent{
		MessageStop: &anthropicTypes.MessageStopEvent{
			Type: anthropicTypes.StreamEventMessageStop,
		},
	}

	contract, err := StreamEventToContract(event, ctx, nil)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if contract == nil {
		t.Fatal("contract 不应为 nil")
	}
}

// TestStreamEventToContract_UnknownEventType 测试未知事件类型的错误处理
func TestStreamEventToContract_UnknownEventType(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	// 创建一个空事件（没有设置任何类型）
	event := &anthropicTypes.StreamEvent{}

	_, err := StreamEventToContract(event, ctx, log)
	if err == nil {
		t.Fatal("期望返回错误")
	}
}

// TestStreamEventToContract_MessageStart 测试 message_start 事件的转换
func TestStreamEventToContract_MessageStart(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	messageID := "msg_123"
	text := "Hello, world!"

	event := &anthropicTypes.StreamEvent{
		MessageStart: &anthropicTypes.MessageStartEvent{
			Type: anthropicTypes.StreamEventMessageStart,
			Message: anthropicTypes.Response{
				ID:   messageID,
				Type: anthropicTypes.ResponseTypeMessage,
				Role: anthropicTypes.RoleAssistant,
				Content: []anthropicTypes.ResponseContentBlock{
					{
						Text: &anthropicTypes.TextBlock{
							Type: anthropicTypes.ResponseContentBlockText,
							Text: text,
						},
					},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID == "" {
		t.Error("ItemID 应被补齐")
	}

	// 验证结果
	if contract == nil {
		t.Fatal("contract 不应为 nil")
	}
	if contract.Type != types.StreamEventMessageStart {
		t.Errorf("Type = %v, 期望 %v", contract.Type, types.StreamEventMessageStart)
	}
	if contract.Source != types.StreamSourceAnthropic {
		t.Errorf("Source = %v, 期望 %v", contract.Source, types.StreamSourceAnthropic)
	}
	if contract.ResponseID != messageID {
		t.Errorf("ResponseID = %v, 期望 %v", contract.ResponseID, messageID)
	}
	if contract.MessageID != messageID {
		t.Errorf("MessageID = %v, 期望 %v", contract.MessageID, messageID)
	}
	if contract.Message == nil {
		t.Fatal("Message 不应为 nil")
	}
	if contract.Message.Role != "assistant" {
		t.Errorf("Message.Role = %v, 期望 assistant", contract.Message.Role)
	}
	if len(contract.Message.Parts) != 1 {
		t.Fatalf("Message.Parts 长度 = %d, 期望 1", len(contract.Message.Parts))
	}
	if contract.Message.Parts[0].Type != "text" {
		t.Errorf("Parts[0].Type = %v, 期望 text", contract.Message.Parts[0].Type)
	}
	if contract.Message.Parts[0].Text != text {
		t.Errorf("Parts[0].Text = %v, 期望 %v", contract.Message.Parts[0].Text, text)
	}
}

// TestStreamEventToContract_MessageStart_WithToolUse 测试包含工具调用的 message_start
func TestStreamEventToContract_MessageStart_WithToolUse(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	toolID := "toolu_123"
	toolName := "get_weather"

	event := &anthropicTypes.StreamEvent{
		MessageStart: &anthropicTypes.MessageStartEvent{
			Type: anthropicTypes.StreamEventMessageStart,
			Message: anthropicTypes.Response{
				ID:   "msg_123",
				Type: anthropicTypes.ResponseTypeMessage,
				Role: anthropicTypes.RoleAssistant,
				Content: []anthropicTypes.ResponseContentBlock{
					{
						ToolUse: &anthropicTypes.ToolUseBlock{
							Type:  anthropicTypes.ResponseContentBlockToolUse,
							ID:    toolID,
							Name:  toolName,
							Input: map[string]interface{}{"location": "Beijing"},
						},
					},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID != toolID {
		t.Errorf("ItemID = %v, 期望 %v（应使用 tool_use.id）", contract.ItemID, toolID)
	}

	// 验证结果
	if len(contract.Message.ToolCalls) != 1 {
		t.Fatalf("ToolCalls 长度 = %d, 期望 1", len(contract.Message.ToolCalls))
	}
	if contract.Message.ToolCalls[0].ID != toolID {
		t.Errorf("ToolCalls[0].ID = %v, 期望 %v", contract.Message.ToolCalls[0].ID, toolID)
	}
	if contract.Message.ToolCalls[0].Name != toolName {
		t.Errorf("ToolCalls[0].Name = %v, 期望 %v", contract.Message.ToolCalls[0].Name, toolName)
	}
	if contract.Message.ToolCalls[0].Type != "tool_use" {
		t.Errorf("ToolCalls[0].Type = %v, 期望 tool_use", contract.Message.ToolCalls[0].Type)
	}
}

// TestStreamEventToContract_MessageStart_WithThinking 测试包含思考内容的 message_start
func TestStreamEventToContract_MessageStart_WithThinking(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	thinking := "Let me think about this..."

	event := &anthropicTypes.StreamEvent{
		MessageStart: &anthropicTypes.MessageStartEvent{
			Type: anthropicTypes.StreamEventMessageStart,
			Message: anthropicTypes.Response{
				ID:   "msg_123",
				Type: anthropicTypes.ResponseTypeMessage,
				Role: anthropicTypes.RoleAssistant,
				Content: []anthropicTypes.ResponseContentBlock{
					{
						Thinking: &anthropicTypes.ThinkingBlock{
							Type:      anthropicTypes.ResponseContentBlockThinking,
							Thinking:  thinking,
							Signature: "abc123",
						},
					},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID == "" {
		t.Error("ItemID 应被补齐")
	}

	// 验证结果
	if len(contract.Message.Parts) != 1 {
		t.Fatalf("Parts 长度 = %d, 期望 1", len(contract.Message.Parts))
	}
	if contract.Message.Parts[0].Type != "thinking" {
		t.Errorf("Parts[0].Type = %v, 期望 thinking", contract.Message.Parts[0].Type)
	}
	if contract.Message.Parts[0].Text != thinking {
		t.Errorf("Parts[0].Text = %v, 期望 %v", contract.Message.Parts[0].Text, thinking)
	}
}

// TestStreamEventToContract_MessageStart_WithCitations 测试包含引用的 message_start
func TestStreamEventToContract_MessageStart_WithCitations(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	event := &anthropicTypes.StreamEvent{
		MessageStart: &anthropicTypes.MessageStartEvent{
			Type: anthropicTypes.StreamEventMessageStart,
			Message: anthropicTypes.Response{
				ID:   "msg_123",
				Type: anthropicTypes.ResponseTypeMessage,
				Role: anthropicTypes.RoleAssistant,
				Content: []anthropicTypes.ResponseContentBlock{
					{
						Text: &anthropicTypes.TextBlock{
							Type: anthropicTypes.ResponseContentBlockText,
							Text: "Hello",
							Citations: []anthropicTypes.TextCitation{
								{
									CharLocation: &anthropicTypes.CitationCharLocation{
										Type:           anthropicTypes.TextCitationTypeCharLocation,
										CitedText:      "Hello",
										DocumentIndex:  0,
										DocumentTitle:  "doc.pdf",
										FileID:         "file_123",
										StartCharIndex: 0,
										EndCharIndex:   5,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if len(contract.Message.Parts[0].Annotations) != 1 {
		t.Fatalf("Annotations 长度 = %d, 期望 1", len(contract.Message.Parts[0].Annotations))
	}
}

// TestStreamEventToContract_MessageDelta 测试 message_delta 事件的转换
func TestStreamEventToContract_MessageDelta(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	inputTokens := 100
	outputTokens := 50
	stopReason := anthropicTypes.StopReasonEndTurn

	event := &anthropicTypes.StreamEvent{
		MessageDelta: &anthropicTypes.MessageDeltaEvent{
			Type: anthropicTypes.StreamEventMessageDelta,
			Delta: anthropicTypes.MessageDelta{
				StopReason: &stopReason,
			},
			Usage: &anthropicTypes.MessageDeltaUsage{
				InputTokens:  &inputTokens,
				OutputTokens: &outputTokens,
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}

	// 验证结果
	if contract.Type != types.StreamEventMessageDelta {
		t.Errorf("Type = %v, 期望 %v", contract.Type, types.StreamEventMessageDelta)
	}
	if contract.Delta == nil {
		t.Fatal("Delta 不应为 nil")
	}
	if contract.Delta.Raw["stop_reason"] != string(stopReason) {
		t.Errorf("Delta.Raw[stop_reason] = %v, 期望 %v", contract.Delta.Raw["stop_reason"], stopReason)
	}
	if contract.Usage == nil {
		t.Fatal("Usage 不应为 nil")
	}
	if contract.Usage.InputTokens == nil || *contract.Usage.InputTokens != inputTokens {
		t.Errorf("InputTokens = %v, 期望 %v", contract.Usage.InputTokens, inputTokens)
	}
	if contract.Usage.OutputTokens == nil || *contract.Usage.OutputTokens != outputTokens {
		t.Errorf("OutputTokens = %v, 期望 %v", contract.Usage.OutputTokens, outputTokens)
	}
	if contract.Usage.TotalTokens == nil || *contract.Usage.TotalTokens != 150 {
		t.Errorf("TotalTokens = %v, 期望 150", contract.Usage.TotalTokens)
	}
}

// TestStreamEventToContract_MessageDelta_WithCacheTokens 测试包含缓存 token 的 message_delta
func TestStreamEventToContract_MessageDelta_WithCacheTokens(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	cacheCreation := 10
	cacheRead := 20

	event := &anthropicTypes.StreamEvent{
		MessageDelta: &anthropicTypes.MessageDeltaEvent{
			Type: anthropicTypes.StreamEventMessageDelta,
			Usage: &anthropicTypes.MessageDeltaUsage{
				CacheCreationInputTokens: &cacheCreation,
				CacheReadInputTokens:     &cacheRead,
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract.Usage.Raw["cache_creation_input_tokens"] != cacheCreation {
		t.Errorf("Raw[cache_creation_input_tokens] = %v, 期望 %v", contract.Usage.Raw["cache_creation_input_tokens"], cacheCreation)
	}
	if contract.Usage.Raw["cache_read_input_tokens"] != cacheRead {
		t.Errorf("Raw[cache_read_input_tokens] = %v, 期望 %v", contract.Usage.Raw["cache_read_input_tokens"], cacheRead)
	}
}

// TestStreamEventToContract_MessageDelta_WithServerToolUse 测试包含服务器工具使用的 message_delta
func TestStreamEventToContract_MessageDelta_WithServerToolUse(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	webSearchRequests := 5

	event := &anthropicTypes.StreamEvent{
		MessageDelta: &anthropicTypes.MessageDeltaEvent{
			Type: anthropicTypes.StreamEventMessageDelta,
			Usage: &anthropicTypes.MessageDeltaUsage{
				ServerToolUse: &anthropicTypes.ServerToolUsage{
					WebSearchRequests: &webSearchRequests,
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract.Usage.Raw["server_tool_use"] == nil {
		t.Fatal("Raw[server_tool_use] 不应为 nil")
	}
}

// TestStreamEventToContract_MessageStop 测试 message_stop 事件的转换
func TestStreamEventToContract_MessageStop(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	event := &anthropicTypes.StreamEvent{
		MessageStop: &anthropicTypes.MessageStopEvent{
			Type: anthropicTypes.StreamEventMessageStop,
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}

	// 验证结果
	if contract.Type != types.StreamEventMessageStop {
		t.Errorf("Type = %v, 期望 %v", contract.Type, types.StreamEventMessageStop)
	}
	if contract.Source != types.StreamSourceAnthropic {
		t.Errorf("Source = %v, 期望 %v", contract.Source, types.StreamSourceAnthropic)
	}
}

// TestStreamEventToContract_ContentBlockStart_Text 测试文本内容块的转换
func TestStreamEventToContract_ContentBlockStart_Text(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	text := "Hello, world!"
	citation := anthropicTypes.TextCitation{
		CharLocation: &anthropicTypes.CitationCharLocation{
			Type:           anthropicTypes.TextCitationTypeCharLocation,
			CitedText:      "quote",
			DocumentIndex:  2,
			DocumentTitle:  "doc",
			FileID:         "file_1",
			StartCharIndex: 1,
			EndCharIndex:   2,
		},
	}

	event := &anthropicTypes.StreamEvent{
		ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
			Type:  anthropicTypes.StreamEventContentBlockStart,
			Index: 0,
			ContentBlock: anthropicTypes.ResponseContentBlock{
				Text: &anthropicTypes.TextBlock{
					Type:      anthropicTypes.ResponseContentBlockText,
					Text:      text,
					Citations: []anthropicTypes.TextCitation{citation},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID == "" {
		t.Error("ItemID 应被补齐")
	}

	// 验证结果
	if contract.Type != types.StreamEventContentBlockStart {
		t.Errorf("Type = %v, 期望 %v", contract.Type, types.StreamEventContentBlockStart)
	}
	if contract.ContentIndex != 0 {
		t.Errorf("ContentIndex = %d, 期望 0", contract.ContentIndex)
	}
	if contract.Content == nil {
		t.Fatal("Content 不应为 nil")
	}
	if contract.Content.Kind != "text" {
		t.Errorf("Content.Kind = %v, 期望 text", contract.Content.Kind)
	}
	if contract.Content.Text == nil || *contract.Content.Text != text {
		t.Errorf("Content.Text = %v, 期望 %v", contract.Content.Text, text)
	}
	if assert.Len(t, contract.Content.Annotations, 1, "Annotations 应包含引用") {
		annotation, ok := contract.Content.Annotations[0].(types.ResponseAnnotation)
		if !ok {
			t.Fatalf("Annotations[0] 类型应为 ResponseAnnotation")
		}
		assert.NotNil(t, annotation.Extras, "Annotation.Extras 不应为 nil")
		citationTypeRaw, ok := GetVendorExtraRaw("anthropic.citation_type", annotation.Extras)
		assert.True(t, ok, "citation_type 应存在")
		assert.Equal(t, "char_location", string(citationTypeRaw), "citation_type 应为 char_location")
		docIndexRaw, ok := GetVendorExtraRaw("anthropic.document_index", annotation.Extras)
		assert.True(t, ok, "document_index 应存在")
		assert.Equal(t, "2", string(docIndexRaw), "document_index 应为 2")
		citedTextRaw, ok := GetVendorExtraRaw("anthropic.cited_text", annotation.Extras)
		assert.True(t, ok, "cited_text 应存在")
		assert.Equal(t, "quote", string(citedTextRaw), "cited_text 应为 quote")
	}
}

// TestStreamEventToContract_ContentBlockStart_ToolUse 测试工具使用内容块的转换
func TestStreamEventToContract_ContentBlockStart_ToolUse(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	toolID := "toolu_123"
	toolName := "get_weather"

	event := &anthropicTypes.StreamEvent{
		ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
			Type:  anthropicTypes.StreamEventContentBlockStart,
			Index: 0,
			ContentBlock: anthropicTypes.ResponseContentBlock{
				ToolUse: &anthropicTypes.ToolUseBlock{
					Type:  anthropicTypes.ResponseContentBlockToolUse,
					ID:    toolID,
					Name:  toolName,
					Input: map[string]interface{}{"location": "Beijing"},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID != toolID {
		t.Errorf("ItemID = %v, 期望 %v（应使用 tool_use.id）", contract.ItemID, toolID)
	}

	// 验证结果
	if contract.Content.Kind != "tool_use" {
		t.Errorf("Content.Kind = %v, 期望 tool_use", contract.Content.Kind)
	}
	if contract.Content.Tool == nil {
		t.Fatal("Content.Tool 不应为 nil")
	}
	if contract.Content.Tool.ID != toolID {
		t.Errorf("Tool.ID = %v, 期望 %v", contract.Content.Tool.ID, toolID)
	}
	if contract.Content.Tool.Name != toolName {
		t.Errorf("Tool.Name = %v, 期望 %v", contract.Content.Tool.Name, toolName)
	}
	if contract.Content.Tool.Type != "tool_use" {
		t.Errorf("Tool.Type = %v, 期望 tool_use", contract.Content.Tool.Type)
	}
}

// TestStreamEventToContract_ContentBlockStart_Thinking 测试思考内容块的转换
func TestStreamEventToContract_ContentBlockStart_Thinking(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	thinking := "Let me think about this..."

	event := &anthropicTypes.StreamEvent{
		ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
			Type:  anthropicTypes.StreamEventContentBlockStart,
			Index: 0,
			ContentBlock: anthropicTypes.ResponseContentBlock{
				Thinking: &anthropicTypes.ThinkingBlock{
					Type:      anthropicTypes.ResponseContentBlockThinking,
					Thinking:  thinking,
					Signature: "abc123",
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID == "" {
		t.Error("ItemID 应被补齐")
	}

	// 验证结果
	if contract.Content.Kind != "thinking" {
		t.Errorf("Content.Kind = %v, 期望 thinking", contract.Content.Kind)
	}
	if contract.Content.Text == nil || *contract.Content.Text != thinking {
		t.Errorf("Content.Text = %v, 期望 %v", contract.Content.Text, thinking)
	}
}

// TestStreamEventToContract_ContentBlockStart_RedactedThinking 测试脱敏思考内容块的转换
func TestStreamEventToContract_ContentBlockStart_RedactedThinking(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	data := "redacted_data"

	event := &anthropicTypes.StreamEvent{
		ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
			Type:  anthropicTypes.StreamEventContentBlockStart,
			Index: 0,
			ContentBlock: anthropicTypes.ResponseContentBlock{
				RedactedThinking: &anthropicTypes.RedactedThinkingBlock{
					Type: anthropicTypes.ResponseContentBlockRedactedThinking,
					Data: data,
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID == "" {
		t.Error("ItemID 应被补齐")
	}

	// 验证结果
	if contract.Content.Kind != "redacted_thinking" {
		t.Errorf("Content.Kind = %v, 期望 redacted_thinking", contract.Content.Kind)
	}
}

// TestStreamEventToContract_ContentBlockStart_ServerToolUse 测试服务器工具使用内容块的转换
func TestStreamEventToContract_ContentBlockStart_ServerToolUse(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	toolID := "server_tool_123"
	toolName := "server_search"

	event := &anthropicTypes.StreamEvent{
		ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
			Type:  anthropicTypes.StreamEventContentBlockStart,
			Index: 0,
			ContentBlock: anthropicTypes.ResponseContentBlock{
				ServerToolUse: &anthropicTypes.ServerToolUseBlock{
					Type:  anthropicTypes.ResponseContentBlockServerToolUse,
					ID:    toolID,
					Name:  toolName,
					Input: map[string]interface{}{"query": "test"},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID != toolID {
		t.Errorf("ItemID = %v, 期望 %v（应使用 server_tool_use.id）", contract.ItemID, toolID)
	}

	// 验证结果
	if contract.Content.Kind != "server_tool_use" {
		t.Errorf("Content.Kind = %v, 期望 server_tool_use", contract.Content.Kind)
	}
	if contract.Content.Tool == nil {
		t.Fatal("Content.Tool 不应为 nil")
	}
	if contract.Content.Tool.ID != toolID {
		t.Errorf("Tool.ID = %v, 期望 %v", contract.Content.Tool.ID, toolID)
	}
	if contract.Content.Tool.Type != "server_tool_use" {
		t.Errorf("Tool.Type = %v, 期望 server_tool_use", contract.Content.Tool.Type)
	}
}

// TestStreamEventToContract_ContentBlockStart_WebSearchToolResult 测试 Web 搜索结果内容块的转换
func TestStreamEventToContract_ContentBlockStart_WebSearchToolResult(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	event := &anthropicTypes.StreamEvent{
		ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
			Type:  anthropicTypes.StreamEventContentBlockStart,
			Index: 0,
			ContentBlock: anthropicTypes.ResponseContentBlock{
				WebSearchToolResult: &anthropicTypes.WebSearchToolResultBlock{
					Type:      anthropicTypes.ResponseContentBlockWebSearchToolResult,
					ToolUseID: "toolu_123",
					Content: anthropicTypes.WebSearchToolResultBlockContent{
						Results: []anthropicTypes.WebSearchResultBlock{
							{
								Type:             "web_search_result",
								Title:            "Test Result",
								URL:              "https://example.com",
								EncryptedContent: "encrypted",
								PageAge:          "1 day",
							},
						},
					},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID == "" {
		t.Error("ItemID 应被补齐")
	}

	// 验证结果
	if contract.Content.Kind != "web_search_tool_result" {
		t.Errorf("Content.Kind = %v, 期望 web_search_tool_result", contract.Content.Kind)
	}
}

// TestStreamEventToContract_ContentBlockStart_Other 测试未知类型内容块的转换
func TestStreamEventToContract_ContentBlockStart_Other(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	event := &anthropicTypes.StreamEvent{
		ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
			Type:         anthropicTypes.StreamEventContentBlockStart,
			Index:        0,
			ContentBlock: anthropicTypes.ResponseContentBlock{
				// 空内容块，应映射为 "other"
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract.Content.Kind != "other" {
		t.Errorf("Content.Kind = %v, 期望 other", contract.Content.Kind)
	}
}

// TestStreamEventToContract_ContentBlockDelta_TextDelta 测试文本增量事件的转换
func TestStreamEventToContract_ContentBlockDelta_TextDelta(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	text := "Hello"

	event := &anthropicTypes.StreamEvent{
		ContentBlockDelta: &anthropicTypes.ContentBlockDeltaEvent{
			Type:  anthropicTypes.StreamEventContentBlockDelta,
			Index: 0,
			Delta: anthropicTypes.ContentBlockDelta{
				Text: &anthropicTypes.TextDelta{
					Type: anthropicTypes.DeltaTypeText,
					Text: text,
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}

	// 验证结果
	if contract.Type != types.StreamEventContentBlockDelta {
		t.Errorf("Type = %v, 期望 %v", contract.Type, types.StreamEventContentBlockDelta)
	}
	if contract.Delta == nil {
		t.Fatal("Delta 不应为 nil")
	}
	if contract.Delta.DeltaType != string(anthropicTypes.DeltaTypeText) {
		t.Errorf("Delta.DeltaType = %v, 期望 %v", contract.Delta.DeltaType, anthropicTypes.DeltaTypeText)
	}
	if contract.Delta.Text == nil || *contract.Delta.Text != text {
		t.Errorf("Delta.Text = %v, 期望 %v", contract.Delta.Text, text)
	}
}

// TestStreamEventToContract_ContentBlockDelta_InputJSONDelta 测试工具输入 JSON 增量事件的转换
func TestStreamEventToContract_ContentBlockDelta_InputJSONDelta(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	partialJSON := `{"location":"`

	event := &anthropicTypes.StreamEvent{
		ContentBlockDelta: &anthropicTypes.ContentBlockDeltaEvent{
			Type:  anthropicTypes.StreamEventContentBlockDelta,
			Index: 0,
			Delta: anthropicTypes.ContentBlockDelta{
				InputJSON: &anthropicTypes.InputJSONDelta{
					Type:        anthropicTypes.DeltaTypeInputJSON,
					PartialJSON: partialJSON,
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract.Delta.DeltaType != string(anthropicTypes.DeltaTypeInputJSON) {
		t.Errorf("Delta.DeltaType = %v, 期望 %v", contract.Delta.DeltaType, anthropicTypes.DeltaTypeInputJSON)
	}
	if contract.Delta.PartialJSON == nil || *contract.Delta.PartialJSON != partialJSON {
		t.Errorf("Delta.PartialJSON = %v, 期望 %v", contract.Delta.PartialJSON, partialJSON)
	}
}

// TestStreamEventToContract_ContentBlockDelta_ThinkingDelta 测试思考增量事件的转换
func TestStreamEventToContract_ContentBlockDelta_ThinkingDelta(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	thinking := "Let me think..."

	event := &anthropicTypes.StreamEvent{
		ContentBlockDelta: &anthropicTypes.ContentBlockDeltaEvent{
			Type:  anthropicTypes.StreamEventContentBlockDelta,
			Index: 0,
			Delta: anthropicTypes.ContentBlockDelta{
				Thinking: &anthropicTypes.ThinkingDelta{
					Type:     anthropicTypes.DeltaTypeThinking,
					Thinking: thinking,
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract.Delta.DeltaType != string(anthropicTypes.DeltaTypeThinking) {
		t.Errorf("Delta.DeltaType = %v, 期望 %v", contract.Delta.DeltaType, anthropicTypes.DeltaTypeThinking)
	}
	if contract.Delta.Thinking == nil || *contract.Delta.Thinking != thinking {
		t.Errorf("Delta.Thinking = %v, 期望 %v", contract.Delta.Thinking, thinking)
	}
}

// TestStreamEventToContract_ContentBlockDelta_SignatureDelta 测试签名增量事件的转换
func TestStreamEventToContract_ContentBlockDelta_SignatureDelta(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	signature := "abc123"

	event := &anthropicTypes.StreamEvent{
		ContentBlockDelta: &anthropicTypes.ContentBlockDeltaEvent{
			Type:  anthropicTypes.StreamEventContentBlockDelta,
			Index: 0,
			Delta: anthropicTypes.ContentBlockDelta{
				Signature: &anthropicTypes.SignatureDelta{
					Type:      anthropicTypes.DeltaTypeSignature,
					Signature: signature,
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract.Delta.DeltaType != string(anthropicTypes.DeltaTypeSignature) {
		t.Errorf("Delta.DeltaType = %v, 期望 %v", contract.Delta.DeltaType, anthropicTypes.DeltaTypeSignature)
	}
	if contract.Delta.Signature == nil || *contract.Delta.Signature != signature {
		t.Errorf("Delta.Signature = %v, 期望 %v", contract.Delta.Signature, signature)
	}
}

// TestStreamEventToContract_ContentBlockDelta_CitationsDelta 测试引用增量事件的转换
func TestStreamEventToContract_ContentBlockDelta_CitationsDelta(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	citation := anthropicTypes.TextCitation{
		CharLocation: &anthropicTypes.CitationCharLocation{
			Type:           anthropicTypes.TextCitationTypeCharLocation,
			CitedText:      "Hello",
			DocumentIndex:  0,
			DocumentTitle:  "doc.pdf",
			FileID:         "file_123",
			StartCharIndex: 0,
			EndCharIndex:   5,
		},
	}

	event := &anthropicTypes.StreamEvent{
		ContentBlockDelta: &anthropicTypes.ContentBlockDeltaEvent{
			Type:  anthropicTypes.StreamEventContentBlockDelta,
			Index: 0,
			Delta: anthropicTypes.ContentBlockDelta{
				Citations: &anthropicTypes.CitationsDelta{
					Type:     anthropicTypes.DeltaTypeCitations,
					Citation: citation,
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract.Delta.DeltaType != string(anthropicTypes.DeltaTypeCitations) {
		t.Errorf("Delta.DeltaType = %v, 期望 %v", contract.Delta.DeltaType, anthropicTypes.DeltaTypeCitations)
	}
	if contract.Delta.Citation == nil {
		t.Fatal("Delta.Citation 不应为 nil")
	}
}

// TestStreamEventToContract_ContentBlockDelta_Other 测试未知类型增量事件的转换
func TestStreamEventToContract_ContentBlockDelta_Other(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	event := &anthropicTypes.StreamEvent{
		ContentBlockDelta: &anthropicTypes.ContentBlockDeltaEvent{
			Type:  anthropicTypes.StreamEventContentBlockDelta,
			Index: 0,
			Delta: anthropicTypes.ContentBlockDelta{
				// 空增量，应映射为 "other"
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract.Delta.DeltaType != "other" {
		t.Errorf("Delta.DeltaType = %v, 期望 other", contract.Delta.DeltaType)
	}
}

// TestStreamEventToContract_ContentBlockStop 测试 content_block_stop 事件的转换
func TestStreamEventToContract_ContentBlockStop(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	event := &anthropicTypes.StreamEvent{
		ContentBlockStop: &anthropicTypes.ContentBlockStopEvent{
			Type:  anthropicTypes.StreamEventContentBlockStop,
			Index: 0,
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证索引字段
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}

	// 验证结果
	if contract.Type != types.StreamEventContentBlockStop {
		t.Errorf("Type = %v, 期望 %v", contract.Type, types.StreamEventContentBlockStop)
	}
	if contract.ContentIndex != 0 {
		t.Errorf("ContentIndex = %d, 期望 0", contract.ContentIndex)
	}
}

// TestStreamEventToContract_Ping 测试 ping 事件的转换
func TestStreamEventToContract_Ping(t *testing.T) {
	log := logger.NewNopLogger()

	event := &anthropicTypes.StreamEvent{
		Ping: &anthropicTypes.PingEvent{
			Type: anthropicTypes.StreamEventPing,
		},
	}

	contract, err := StreamEventToContract(event, nil, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract.Type != types.StreamEventPing {
		t.Errorf("Type = %v, 期望 %v", contract.Type, types.StreamEventPing)
	}
	if contract.Source != types.StreamSourceAnthropic {
		t.Errorf("Source = %v, 期望 %v", contract.Source, types.StreamSourceAnthropic)
	}
}

// TestStreamEventToContract_Error 测试 error 事件的转换
func TestStreamEventToContract_Error(t *testing.T) {
	log := logger.NewNopLogger()

	event := &anthropicTypes.StreamEvent{
		Error: &anthropicTypes.ErrorEvent{
			Type: anthropicTypes.StreamEventError,
			Error: anthropicTypes.ErrorResponse{
				Type: "error",
				Error: anthropicTypes.Error{
					Type:    "invalid_request_error",
					Message: "Invalid API key",
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, nil, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract.Type != types.StreamEventError {
		t.Errorf("Type = %v, 期望 %v", contract.Type, types.StreamEventError)
	}
	if contract.Error == nil {
		t.Fatal("Error 不应为 nil")
	}
	if contract.Error.Type != "invalid_request_error" {
		t.Errorf("Error.Type = %v, 期望 invalid_request_error", contract.Error.Type)
	}
	if contract.Error.Message != "Invalid API key" {
		t.Errorf("Error.Message = %v, 期望 Invalid API key", contract.Error.Message)
	}
}

// TestConvertResponseContentBlocksToStreamParts_Text 测试文本内容块的转换
func TestConvertResponseContentBlocksToStreamParts_Text(t *testing.T) {
	log := logger.NewNopLogger()

	blocks := []anthropicTypes.ResponseContentBlock{
		{
			Text: &anthropicTypes.TextBlock{
				Type: anthropicTypes.ResponseContentBlockText,
				Text: "Hello, world!",
			},
		},
	}

	parts, toolCalls, err := convertResponseContentBlocksToStreamParts(blocks, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if len(parts) != 1 {
		t.Fatalf("Parts 长度 = %d, 期望 1", len(parts))
	}
	if parts[0].Type != "text" {
		t.Errorf("Parts[0].Type = %v, 期望 text", parts[0].Type)
	}
	if parts[0].Text != "Hello, world!" {
		t.Errorf("Parts[0].Text = %v, 期望 Hello, world!", parts[0].Text)
	}
	if len(toolCalls) != 0 {
		t.Fatalf("ToolCalls 长度 = %d, 期望 0", len(toolCalls))
	}
}

// TestConvertResponseContentBlocksToStreamParts_ToolUse 测试工具使用内容块的转换
func TestConvertResponseContentBlocksToStreamParts_ToolUse(t *testing.T) {
	log := logger.NewNopLogger()

	blocks := []anthropicTypes.ResponseContentBlock{
		{
			ToolUse: &anthropicTypes.ToolUseBlock{
				Type:  anthropicTypes.ResponseContentBlockToolUse,
				ID:    "toolu_123",
				Name:  "get_weather",
				Input: map[string]interface{}{"location": "Beijing"},
			},
		},
	}

	parts, toolCalls, err := convertResponseContentBlocksToStreamParts(blocks, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if len(parts) != 0 {
		t.Fatalf("Parts 长度 = %d, 期望 0", len(parts))
	}
	if len(toolCalls) != 1 {
		t.Fatalf("ToolCalls 长度 = %d, 期望 1", len(toolCalls))
	}
	if toolCalls[0].ID != "toolu_123" {
		t.Errorf("ToolCalls[0].ID = %v, 期望 toolu_123", toolCalls[0].ID)
	}
	if toolCalls[0].Name != "get_weather" {
		t.Errorf("ToolCalls[0].Name = %v, 期望 get_weather", toolCalls[0].Name)
	}
	if toolCalls[0].Type != "tool_use" {
		t.Errorf("ToolCalls[0].Type = %v, 期望 tool_use", toolCalls[0].Type)
	}
}

// TestConvertResponseContentBlocksToStreamParts_Thinking 测试思考内容块的转换
func TestConvertResponseContentBlocksToStreamParts_Thinking(t *testing.T) {
	log := logger.NewNopLogger()

	blocks := []anthropicTypes.ResponseContentBlock{
		{
			Thinking: &anthropicTypes.ThinkingBlock{
				Type:      anthropicTypes.ResponseContentBlockThinking,
				Thinking:  "Let me think...",
				Signature: "abc123",
			},
		},
	}

	parts, _, err := convertResponseContentBlocksToStreamParts(blocks, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if len(parts) != 1 {
		t.Fatalf("Parts 长度 = %d, 期望 1", len(parts))
	}
	if parts[0].Type != "thinking" {
		t.Errorf("Parts[0].Type = %v, 期望 thinking", parts[0].Type)
	}
	if parts[0].Text != "Let me think..." {
		t.Errorf("Parts[0].Text = %v, 期望 Let me think...", parts[0].Text)
	}
	if parts[0].Raw["signature"] != "abc123" {
		t.Errorf("Parts[0].Raw[signature] = %v, 期望 abc123", parts[0].Raw["signature"])
	}
}

// TestConvertResponseContentBlocksToStreamParts_RedactedThinking 测试脱敏思考内容块的转换
func TestConvertResponseContentBlocksToStreamParts_RedactedThinking(t *testing.T) {
	log := logger.NewNopLogger()

	blocks := []anthropicTypes.ResponseContentBlock{
		{
			RedactedThinking: &anthropicTypes.RedactedThinkingBlock{
				Type: anthropicTypes.ResponseContentBlockRedactedThinking,
				Data: "redacted_data",
			},
		},
	}

	parts, _, err := convertResponseContentBlocksToStreamParts(blocks, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if len(parts) != 1 {
		t.Fatalf("Parts 长度 = %d, 期望 1", len(parts))
	}
	if parts[0].Type != "redacted_thinking" {
		t.Errorf("Parts[0].Type = %v, 期望 redacted_thinking", parts[0].Type)
	}
	if parts[0].Raw["data"] != "redacted_data" {
		t.Errorf("Parts[0].Raw[data] = %v, 期望 redacted_data", parts[0].Raw["data"])
	}
}

// TestConvertResponseContentBlocksToStreamParts_ServerToolUse 测试服务器工具使用内容块的转换
func TestConvertResponseContentBlocksToStreamParts_ServerToolUse(t *testing.T) {
	log := logger.NewNopLogger()

	blocks := []anthropicTypes.ResponseContentBlock{
		{
			ServerToolUse: &anthropicTypes.ServerToolUseBlock{
				Type:  anthropicTypes.ResponseContentBlockServerToolUse,
				ID:    "server_tool_123",
				Name:  "server_search",
				Input: map[string]interface{}{"query": "test"},
			},
		},
	}

	_, toolCalls, err := convertResponseContentBlocksToStreamParts(blocks, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if len(toolCalls) != 1 {
		t.Fatalf("ToolCalls 长度 = %d, 期望 1", len(toolCalls))
	}
	if toolCalls[0].ID != "server_tool_123" {
		t.Errorf("ToolCalls[0].ID = %v, 期望 server_tool_123", toolCalls[0].ID)
	}
	if toolCalls[0].Type != "server_tool_use" {
		t.Errorf("ToolCalls[0].Type = %v, 期望 server_tool_use", toolCalls[0].Type)
	}
}

// TestConvertResponseContentBlocksToStreamParts_WebSearchToolResult 测试 Web 搜索结果内容块的转换
func TestConvertResponseContentBlocksToStreamParts_WebSearchToolResult(t *testing.T) {
	log := logger.NewNopLogger()

	blocks := []anthropicTypes.ResponseContentBlock{
		{
			WebSearchToolResult: &anthropicTypes.WebSearchToolResultBlock{
				Type:      anthropicTypes.ResponseContentBlockWebSearchToolResult,
				ToolUseID: "toolu_123",
				Content: anthropicTypes.WebSearchToolResultBlockContent{
					Results: []anthropicTypes.WebSearchResultBlock{
						{
							Type:             "web_search_result",
							Title:            "Test Result",
							URL:              "https://example.com",
							EncryptedContent: "encrypted",
							PageAge:          "1 day",
						},
					},
				},
			},
		},
	}

	parts, _, err := convertResponseContentBlocksToStreamParts(blocks, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if len(parts) != 1 {
		t.Fatalf("Parts 长度 = %d, 期望 1", len(parts))
	}
	if parts[0].Type != "web_search_tool_result" {
		t.Errorf("Parts[0].Type = %v, 期望 web_search_tool_result", parts[0].Type)
	}
}

// TestConvertResponseContentBlocksToStreamParts_Mixed 测试混合内容块的转换
func TestConvertResponseContentBlocksToStreamParts_Mixed(t *testing.T) {
	log := logger.NewNopLogger()

	blocks := []anthropicTypes.ResponseContentBlock{
		{
			Text: &anthropicTypes.TextBlock{
				Type: anthropicTypes.ResponseContentBlockText,
				Text: "Hello",
			},
		},
		{
			ToolUse: &anthropicTypes.ToolUseBlock{
				Type:  anthropicTypes.ResponseContentBlockToolUse,
				ID:    "toolu_123",
				Name:  "get_weather",
				Input: map[string]interface{}{"location": "Beijing"},
			},
		},
		{
			Thinking: &anthropicTypes.ThinkingBlock{
				Type:      anthropicTypes.ResponseContentBlockThinking,
				Thinking:  "Thinking...",
				Signature: "sig123",
			},
		},
	}

	parts, toolCalls, err := convertResponseContentBlocksToStreamParts(blocks, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if len(parts) != 2 {
		t.Fatalf("Parts 长度 = %d, 期望 2", len(parts))
	}
	if len(toolCalls) != 1 {
		t.Fatalf("ToolCalls 长度 = %d, 期望 1", len(toolCalls))
	}
}

// TestStreamEventToContract_Extensions 测试扩展字段的保存
func TestStreamEventToContract_VendorSource(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	event := &anthropicTypes.StreamEvent{
		MessageStart: &anthropicTypes.MessageStartEvent{
			Type: anthropicTypes.StreamEventMessageStart,
			Message: anthropicTypes.Response{
				ID:   "msg_vendor",
				Type: anthropicTypes.ResponseTypeMessage,
				Role: anthropicTypes.RoleAssistant,
				Content: []anthropicTypes.ResponseContentBlock{
					{
						Text: &anthropicTypes.TextBlock{
							Type: anthropicTypes.ResponseContentBlockText,
							Text: "hello",
						},
					},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	assert.Equal(t, types.StreamSourceAnthropic, contract.Source, "Source 应为 anthropic")
}

func TestStreamEventToContract_Extensions(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	event := &anthropicTypes.StreamEvent{
		MessageStart: &anthropicTypes.MessageStartEvent{
			Type: anthropicTypes.StreamEventMessageStart,
			Message: anthropicTypes.Response{
				ID:   "msg_123",
				Type: anthropicTypes.ResponseTypeMessage,
				Role: anthropicTypes.RoleAssistant,
				Content: []anthropicTypes.ResponseContentBlock{
					{
						Text: &anthropicTypes.TextBlock{
							Type: anthropicTypes.ResponseContentBlockText,
							Text: "Hello",
						},
					},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证扩展字段
	if contract.Extensions == nil {
		t.Fatal("Extensions 不应为 nil")
	}
	if contract.Extensions["anthropic"] == nil {
		t.Fatal("Extensions[anthropic] 不应为 nil")
	}
}

// TestStreamEventToContract_ContentIndex 测试内容索引的传递
func TestStreamEventToContract_RawFidelity_ContentBlock(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	citation := anthropicTypes.TextCitation{
		CharLocation: &anthropicTypes.CitationCharLocation{
			Type:           anthropicTypes.TextCitationTypeCharLocation,
			CitedText:      "quote",
			DocumentIndex:  1,
			DocumentTitle:  "doc",
			FileID:         "file_1",
			StartCharIndex: 1,
			EndCharIndex:   2,
		},
	}

	event := &anthropicTypes.StreamEvent{
		ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
			Type:  anthropicTypes.StreamEventContentBlockStart,
			Index: 2,
			ContentBlock: anthropicTypes.ResponseContentBlock{
				Text: &anthropicTypes.TextBlock{
					Type:      anthropicTypes.ResponseContentBlockText,
					Text:      "hello",
					Citations: []anthropicTypes.TextCitation{citation},
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	assert.NotNil(t, contract.Content, "Content 不应为 nil")
	assert.NotNil(t, contract.Content.Raw, "Content.Raw 不应为 nil")
	rawBlock, ok := contract.Content.Raw["content_block"]
	assert.True(t, ok, "raw content_block 应存在")

	restored := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockStart,
		Source:       types.StreamSourceAnthropic,
		ContentIndex: contract.ContentIndex,
		Content: &types.StreamContentPayload{
			Kind: contract.Content.Kind,
			Raw: map[string]interface{}{
				"content_block": rawBlock,
			},
		},
	}

	streamEvent, err := StreamEventFromContract(restored, log)
	if err != nil {
		t.Fatalf("反向转换失败: %v", err)
	}
	if streamEvent.ContentBlockStart == nil || streamEvent.ContentBlockStart.ContentBlock.Text == nil {
		t.Fatalf("反向转换缺少 TextBlock")
	}
	assert.Equal(t, "hello", streamEvent.ContentBlockStart.ContentBlock.Text.Text, "TextBlock 文本应保持一致")
	if assert.Len(t, streamEvent.ContentBlockStart.ContentBlock.Text.Citations, 1, "应保留 Citations") {
		assert.Equal(t, "quote", streamEvent.ContentBlockStart.ContentBlock.Text.Citations[0].CharLocation.CitedText, "Citation 应保持一致")
	}
}

func TestStreamEventToContract_ContentIndex(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	event := &anthropicTypes.StreamEvent{
		ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
			Type:  anthropicTypes.StreamEventContentBlockStart,
			Index: 5,
			ContentBlock: anthropicTypes.ResponseContentBlock{
				Text: &anthropicTypes.TextBlock{
					Type: anthropicTypes.ResponseContentBlockText,
					Text: "Hello",
				},
			},
		},
	}

	contract, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证内容索引
	if contract.ContentIndex != 5 {
		t.Errorf("ContentIndex = %d, 期望 5", contract.ContentIndex)
	}
}
