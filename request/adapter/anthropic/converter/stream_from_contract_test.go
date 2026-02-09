package converter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// TestStreamEventFromContract_MessageStart 测试 message_start 事件的转换
func TestStreamEventFromContract_MessageStart(t *testing.T) {
	log := logger.NewNopLogger()

	// 准备测试数据
	messageID := "msg_123"
	role := "assistant"
	text := "Hello, world!"

	contract := &types.StreamEventContract{
		Type:       types.StreamEventMessageStart,
		Source:     types.StreamSourceAnthropic,
		ResponseID: "resp_123",
		MessageID:  messageID,
		Message: &types.StreamMessagePayload{
			Role: role,
			Parts: []types.StreamContentPart{
				{
					Type: "text",
					Text: text,
				},
			},
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event == nil {
		t.Fatal("event 不应为 nil")
	}
	if event.MessageStart == nil {
		t.Fatal("MessageStart 不应为 nil")
	}
	if event.MessageStart.Type != anthropicTypes.StreamEventMessageStart {
		t.Errorf("Type = %v, 期望 %v", event.MessageStart.Type, anthropicTypes.StreamEventMessageStart)
	}
	if event.MessageStart.Message.ID != messageID {
		t.Errorf("Message.ID = %v, 期望 %v", event.MessageStart.Message.ID, messageID)
	}
	if string(event.MessageStart.Message.Role) != role {
		t.Errorf("Message.Role = %v, 期望 %v", event.MessageStart.Message.Role, role)
	}
	if len(event.MessageStart.Message.Content) != 1 {
		t.Fatalf("Content 长度 = %d, 期望 1", len(event.MessageStart.Message.Content))
	}
	if event.MessageStart.Message.Content[0].Text == nil || event.MessageStart.Message.Content[0].Text.Text != text {
		t.Errorf("Content[0].Text.Text = %v, 期望 %v", event.MessageStart.Message.Content[0].Text, text)
	}
}

// TestStreamEventFromContract_MessageStart_WithRawMessage 测试使用原始 message 的转换
func TestStreamEventFromContract_MessageStart_WithRawMessage(t *testing.T) {
	log := logger.NewNopLogger()

	// 准备原始 message
	originalMessage := anthropicTypes.Response{
		ID:   "msg_original",
		Type: anthropicTypes.ResponseTypeMessage,
		Role: anthropicTypes.RoleAssistant,
		Content: []anthropicTypes.ResponseContentBlock{
			{
				Text: &anthropicTypes.TextBlock{
					Type: anthropicTypes.ResponseContentBlockText,
					Text: "Original text",
				},
			},
		},
	}

	// 序列化原始 message
	msgJSON, _ := json.Marshal(originalMessage)

	contract := &types.StreamEventContract{
		Type:       types.StreamEventMessageStart,
		Source:     types.StreamSourceAnthropic,
		ResponseID: "resp_123",
		MessageID:  "msg_123", // 这个会被原始 message 覆盖
		Message: &types.StreamMessagePayload{
			Role: "assistant",
			Parts: []types.StreamContentPart{
				{
					Type: "text",
					Text: "Different text", // 这个会被原始 message 覆盖
				},
			},
		},
		Extensions: map[string]interface{}{
			"anthropic": map[string]interface{}{
				"message": json.RawMessage(msgJSON),
			},
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果 - 应该使用原始 message
	if event.MessageStart.Message.ID != "msg_original" {
		t.Errorf("Message.ID = %v, 期望 msg_original (应使用原始 message)", event.MessageStart.Message.ID)
	}
	if event.MessageStart.Message.Content[0].Text.Text != "Original text" {
		t.Errorf("Content[0].Text.Text = %v, 期望 Original text (应使用原始 message)", event.MessageStart.Message.Content[0].Text.Text)
	}
}

// TestStreamEventFromContract_MessageStart_WithToolCalls 测试包含工具调用的 message_start
func TestStreamEventFromContract_MessageStart_WithToolCalls(t *testing.T) {
	log := logger.NewNopLogger()

	toolID := "toolu_123"
	toolName := "get_weather"
	toolArgs := `{"location":"Beijing"}`

	contract := &types.StreamEventContract{
		Type:       types.StreamEventMessageStart,
		Source:     types.StreamSourceAnthropic,
		ResponseID: "resp_123",
		MessageID:  "msg_123",
		Message: &types.StreamMessagePayload{
			Role: "assistant",
			ToolCalls: []types.StreamToolCall{
				{
					ID:        toolID,
					Type:      "tool_use",
					Name:      toolName,
					Arguments: toolArgs,
				},
			},
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if len(event.MessageStart.Message.Content) != 1 {
		t.Fatalf("Content 长度 = %d, 期望 1", len(event.MessageStart.Message.Content))
	}
	if event.MessageStart.Message.Content[0].ToolUse == nil {
		t.Fatal("ToolUse 不应为 nil")
	}
	if event.MessageStart.Message.Content[0].ToolUse.ID != toolID {
		t.Errorf("ToolUse.ID = %v, 期望 %v", event.MessageStart.Message.Content[0].ToolUse.ID, toolID)
	}
	if event.MessageStart.Message.Content[0].ToolUse.Name != toolName {
		t.Errorf("ToolUse.Name = %v, 期望 %v", event.MessageStart.Message.Content[0].ToolUse.Name, toolName)
	}
}

// TestStreamEventFromContract_MessageDelta 测试 message_delta 事件的转换
func TestStreamEventFromContract_MessageDelta(t *testing.T) {
	log := logger.NewNopLogger()

	inputTokens := 100
	outputTokens := 50
	stopReason := "end_turn"

	contract := &types.StreamEventContract{
		Type:   types.StreamEventMessageDelta,
		Source: types.StreamSourceAnthropic,
		Delta: &types.StreamDeltaPayload{
			DeltaType: "other",
			Raw: map[string]interface{}{
				"stop_reason": stopReason,
			},
		},
		Usage: &types.StreamUsagePayload{
			InputTokens:  &inputTokens,
			OutputTokens: &outputTokens,
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.MessageDelta == nil {
		t.Fatal("MessageDelta 不应为 nil")
	}
	if event.MessageDelta.Type != anthropicTypes.StreamEventMessageDelta {
		t.Errorf("Type = %v, 期望 %v", event.MessageDelta.Type, anthropicTypes.StreamEventMessageDelta)
	}
	if event.MessageDelta.Delta.StopReason == nil || *event.MessageDelta.Delta.StopReason != anthropicTypes.StopReason(stopReason) {
		t.Errorf("StopReason = %v, 期望 %v", event.MessageDelta.Delta.StopReason, stopReason)
	}
	if event.MessageDelta.Usage == nil {
		t.Fatal("Usage 不应为 nil")
	}
	if event.MessageDelta.Usage.InputTokens == nil || *event.MessageDelta.Usage.InputTokens != inputTokens {
		t.Errorf("InputTokens = %v, 期望 %v", event.MessageDelta.Usage.InputTokens, inputTokens)
	}
	if event.MessageDelta.Usage.OutputTokens == nil || *event.MessageDelta.Usage.OutputTokens != outputTokens {
		t.Errorf("OutputTokens = %v, 期望 %v", event.MessageDelta.Usage.OutputTokens, outputTokens)
	}
}

// TestStreamEventFromContract_MessageDelta_WithCacheTokens 测试包含缓存 token 的 message_delta
func TestStreamEventFromContract_MessageDelta_WithCacheTokens(t *testing.T) {
	log := logger.NewNopLogger()

	inputTokens := 100
	cacheCreation := 10
	cacheRead := 20

	contract := &types.StreamEventContract{
		Type:   types.StreamEventMessageDelta,
		Source: types.StreamSourceAnthropic,
		Usage: &types.StreamUsagePayload{
			InputTokens: &inputTokens,
			Raw: map[string]interface{}{
				"cache_creation_input_tokens": float64(cacheCreation),
				"cache_read_input_tokens":     float64(cacheRead),
			},
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.MessageDelta.Usage.CacheCreationInputTokens == nil || *event.MessageDelta.Usage.CacheCreationInputTokens != cacheCreation {
		t.Errorf("CacheCreationInputTokens = %v, 期望 %v", event.MessageDelta.Usage.CacheCreationInputTokens, cacheCreation)
	}
	if event.MessageDelta.Usage.CacheReadInputTokens == nil || *event.MessageDelta.Usage.CacheReadInputTokens != cacheRead {
		t.Errorf("CacheReadInputTokens = %v, 期望 %v", event.MessageDelta.Usage.CacheReadInputTokens, cacheRead)
	}
}

// TestStreamEventFromContract_MessageStop 测试 message_stop 事件的转换
func TestStreamEventFromContract_MessageStop(t *testing.T) {
	log := logger.NewNopLogger()

	contract := &types.StreamEventContract{
		Type:   types.StreamEventMessageStop,
		Source: types.StreamSourceAnthropic,
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.MessageStop == nil {
		t.Fatal("MessageStop 不应为 nil")
	}
	if event.MessageStop.Type != anthropicTypes.StreamEventMessageStop {
		t.Errorf("Type = %v, 期望 %v", event.MessageStop.Type, anthropicTypes.StreamEventMessageStop)
	}
}

// TestStreamEventFromContract_ContentBlockStart_Text 测试文本内容块的转换
func TestStreamEventFromContract_ContentBlockStart_Text(t *testing.T) {
	log := logger.NewNopLogger()

	text := "Hello, world!"

	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockStart,
		Source:       types.StreamSourceAnthropic,
		ContentIndex: 0,
		Content: &types.StreamContentPayload{
			Kind: "text",
			Text: &text,
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.ContentBlockStart == nil {
		t.Fatal("ContentBlockStart 不应为 nil")
	}
	if event.ContentBlockStart.Index != 0 {
		t.Errorf("Index = %d, 期望 0", event.ContentBlockStart.Index)
	}
	if event.ContentBlockStart.ContentBlock.Text == nil {
		t.Fatal("Text 不应为 nil")
	}
	if event.ContentBlockStart.ContentBlock.Text.Text != text {
		t.Errorf("Text.Text = %v, 期望 %v", event.ContentBlockStart.ContentBlock.Text.Text, text)
	}
}

// TestStreamEventFromContract_ContentBlockStart_ToolUse 测试工具使用内容块的转换
func TestStreamEventFromContract_ContentBlockStart_ToolUse(t *testing.T) {
	log := logger.NewNopLogger()

	toolID := "toolu_123"
	toolName := "get_weather"
	toolArgs := `{"location":"Beijing"}`

	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockStart,
		Source:       types.StreamSourceAnthropic,
		ContentIndex: 0,
		Content: &types.StreamContentPayload{
			Kind: "tool_use",
			Tool: &types.StreamToolCall{
				ID:        toolID,
				Type:      "tool_use",
				Name:      toolName,
				Arguments: toolArgs,
			},
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.ContentBlockStart.ContentBlock.ToolUse == nil {
		t.Fatal("ToolUse 不应为 nil")
	}
	if event.ContentBlockStart.ContentBlock.ToolUse.ID != toolID {
		t.Errorf("ToolUse.ID = %v, 期望 %v", event.ContentBlockStart.ContentBlock.ToolUse.ID, toolID)
	}
	if event.ContentBlockStart.ContentBlock.ToolUse.Name != toolName {
		t.Errorf("ToolUse.Name = %v, 期望 %v", event.ContentBlockStart.ContentBlock.ToolUse.Name, toolName)
	}
	if event.ContentBlockStart.ContentBlock.ToolUse.Input == nil {
		t.Fatal("ToolUse.Input 不应为 nil")
	}
	if location, ok := event.ContentBlockStart.ContentBlock.ToolUse.Input["location"].(string); !ok || location != "Beijing" {
		t.Errorf("ToolUse.Input[location] = %v, 期望 Beijing", event.ContentBlockStart.ContentBlock.ToolUse.Input["location"])
	}
}

// TestStreamEventFromContract_ContentBlockStart_RawContentBlock 测试 raw content_block 还原
func TestStreamEventFromContract_ContentBlockStart_RawContentBlock(t *testing.T) {
	log := logger.NewNopLogger()

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
	textBlock := anthropicTypes.TextBlock{
		Type:      anthropicTypes.ResponseContentBlockText,
		Text:      "hello",
		Citations: []anthropicTypes.TextCitation{citation},
	}
	blockJSON, err := json.Marshal(textBlock)
	if err != nil {
		t.Fatalf("序列化 raw content_block 失败: %v", err)
	}

	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockStart,
		Source:       types.StreamSourceAnthropic,
		ContentIndex: 3,
		Content: &types.StreamContentPayload{
			Kind: "text",
			Raw: map[string]interface{}{
				"content_block": json.RawMessage(blockJSON),
			},
		},
	}

	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if assert.NotNil(t, event.ContentBlockStart, "ContentBlockStart 不应为 nil") {
		assert.Equal(t, 3, event.ContentBlockStart.Index, "content_index 应保持一致")
		if assert.NotNil(t, event.ContentBlockStart.ContentBlock.Text, "TextBlock 不应为 nil") {
			assert.Equal(t, "hello", event.ContentBlockStart.ContentBlock.Text.Text, "Text 文本应保持一致")
			if assert.Len(t, event.ContentBlockStart.ContentBlock.Text.Citations, 1, "Citations 应保留") {
				assert.Equal(t, "quote", event.ContentBlockStart.ContentBlock.Text.Citations[0].CharLocation.CitedText, "Citation 应保持一致")
			}
		}
	}
}

// TestStreamEventFromContract_ContentBlockStart_Thinking 测试思考内容块的转换
func TestStreamEventFromContract_ContentBlockStart_Thinking(t *testing.T) {
	log := logger.NewNopLogger()

	thinking := "Let me think about this..."
	signature := "abc123"

	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockStart,
		Source:       types.StreamSourceAnthropic,
		ContentIndex: 0,
		Content: &types.StreamContentPayload{
			Kind: "thinking",
			Text: &thinking,
			Raw: map[string]interface{}{
				"signature": signature,
			},
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.ContentBlockStart.ContentBlock.Thinking == nil {
		t.Fatal("Thinking 不应为 nil")
	}
	if event.ContentBlockStart.ContentBlock.Thinking.Thinking != thinking {
		t.Errorf("Thinking.Thinking = %v, 期望 %v", event.ContentBlockStart.ContentBlock.Thinking.Thinking, thinking)
	}
	if event.ContentBlockStart.ContentBlock.Thinking.Signature != signature {
		t.Errorf("Thinking.Signature = %v, 期望 %v", event.ContentBlockStart.ContentBlock.Thinking.Signature, signature)
	}
}

// TestStreamEventFromContract_ContentBlockDelta_TextDelta 测试文本增量事件的转换
func TestStreamEventFromContract_ContentBlockDelta_TextDelta(t *testing.T) {
	log := logger.NewNopLogger()

	text := "Hello"

	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockDelta,
		Source:       types.StreamSourceAnthropic,
		ContentIndex: 0,
		Delta: &types.StreamDeltaPayload{
			DeltaType: string(anthropicTypes.DeltaTypeText),
			Text:      &text,
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.ContentBlockDelta == nil {
		t.Fatal("ContentBlockDelta 不应为 nil")
	}
	if event.ContentBlockDelta.Index != 0 {
		t.Errorf("Index = %d, 期望 0", event.ContentBlockDelta.Index)
	}
	if event.ContentBlockDelta.Delta.Text == nil {
		t.Fatal("Delta.Text 不应为 nil")
	}
	if event.ContentBlockDelta.Delta.Text.Text != text {
		t.Errorf("Delta.Text.Text = %v, 期望 %v", event.ContentBlockDelta.Delta.Text.Text, text)
	}
}

// TestStreamEventFromContract_ContentBlockDelta_InputJSONDelta 测试工具输入 JSON 增量事件的转换
func TestStreamEventFromContract_ContentBlockDelta_InputJSONDelta(t *testing.T) {
	log := logger.NewNopLogger()

	partialJSON := `{"location":"`

	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockDelta,
		Source:       types.StreamSourceAnthropic,
		ContentIndex: 0,
		Delta: &types.StreamDeltaPayload{
			DeltaType:   string(anthropicTypes.DeltaTypeInputJSON),
			PartialJSON: &partialJSON,
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.ContentBlockDelta.Delta.InputJSON == nil {
		t.Fatal("Delta.InputJSON 不应为 nil")
	}
	if event.ContentBlockDelta.Delta.InputJSON.PartialJSON != partialJSON {
		t.Errorf("Delta.InputJSON.PartialJSON = %v, 期望 %v", event.ContentBlockDelta.Delta.InputJSON.PartialJSON, partialJSON)
	}
}

// TestStreamEventFromContract_ContentBlockDelta_ThinkingDelta 测试思考增量事件的转换
func TestStreamEventFromContract_ContentBlockDelta_ThinkingDelta(t *testing.T) {
	log := logger.NewNopLogger()

	thinking := "Let me think..."

	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockDelta,
		Source:       types.StreamSourceAnthropic,
		ContentIndex: 0,
		Delta: &types.StreamDeltaPayload{
			DeltaType: string(anthropicTypes.DeltaTypeThinking),
			Thinking:  &thinking,
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.ContentBlockDelta.Delta.Thinking == nil {
		t.Fatal("Delta.Thinking 不应为 nil")
	}
	if event.ContentBlockDelta.Delta.Thinking.Thinking != thinking {
		t.Errorf("Delta.Thinking.Thinking = %v, 期望 %v", event.ContentBlockDelta.Delta.Thinking.Thinking, thinking)
	}
}

// TestStreamEventFromContract_ContentBlockDelta_SignatureDelta 测试签名增量事件的转换
func TestStreamEventFromContract_ContentBlockDelta_SignatureDelta(t *testing.T) {
	log := logger.NewNopLogger()

	signature := "abc123"

	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockDelta,
		Source:       types.StreamSourceAnthropic,
		ContentIndex: 0,
		Delta: &types.StreamDeltaPayload{
			DeltaType: string(anthropicTypes.DeltaTypeSignature),
			Signature: &signature,
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.ContentBlockDelta.Delta.Signature == nil {
		t.Fatal("Delta.Signature 不应为 nil")
	}
	if event.ContentBlockDelta.Delta.Signature.Signature != signature {
		t.Errorf("Delta.Signature.Signature = %v, 期望 %v", event.ContentBlockDelta.Delta.Signature.Signature, signature)
	}
}

// TestStreamEventFromContract_ContentBlockStop 测试 content_block_stop 事件的转换
func TestStreamEventFromContract_ContentBlockStop(t *testing.T) {
	log := logger.NewNopLogger()

	contract := &types.StreamEventContract{
		Type:         types.StreamEventContentBlockStop,
		Source:       types.StreamSourceAnthropic,
		ContentIndex: 0,
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.ContentBlockStop == nil {
		t.Fatal("ContentBlockStop 不应为 nil")
	}
	if event.ContentBlockStop.Index != 0 {
		t.Errorf("Index = %d, 期望 0", event.ContentBlockStop.Index)
	}
}

// TestStreamEventFromContract_Ping 测试 ping 事件的转换
func TestStreamEventFromContract_Ping(t *testing.T) {
	log := logger.NewNopLogger()

	contract := &types.StreamEventContract{
		Type:   types.StreamEventPing,
		Source: types.StreamSourceAnthropic,
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.Ping == nil {
		t.Fatal("Ping 不应为 nil")
	}
	if event.Ping.Type != anthropicTypes.StreamEventPing {
		t.Errorf("Type = %v, 期望 %v", event.Ping.Type, anthropicTypes.StreamEventPing)
	}
}

// TestStreamEventFromContract_Error 测试 error 事件的转换
func TestStreamEventFromContract_Error(t *testing.T) {
	log := logger.NewNopLogger()

	contract := &types.StreamEventContract{
		Type:   types.StreamEventError,
		Source: types.StreamSourceAnthropic,
		Error: &types.StreamErrorPayload{
			Type:    "invalid_request_error",
			Message: "Invalid API key",
		},
	}

	// 执行转换
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if event.Error == nil {
		t.Fatal("Error 不应为 nil")
	}
	if event.Error.Type != anthropicTypes.StreamEventError {
		t.Errorf("Type = %v, 期望 %v", event.Error.Type, anthropicTypes.StreamEventError)
	}
	if event.Error.Error.Error.Type != "invalid_request_error" {
		t.Errorf("Error.Type = %v, 期望 invalid_request_error", event.Error.Error.Error.Type)
	}
	if event.Error.Error.Error.Message != "Invalid API key" {
		t.Errorf("Error.Message = %v, 期望 Invalid API key", event.Error.Error.Error.Message)
	}
}

// TestStreamEventFromContract_InvalidSource 测试非 Anthropic 来源的转换
// Source 字段不参与转换逻辑，仅用于标识来源，因此不应返回错误
func TestStreamEventFromContract_InvalidSource(t *testing.T) {
	log := logger.NewNopLogger()

	contract := &types.StreamEventContract{
		Type:      types.StreamEventMessageStart,
		Source:    types.StreamSourceOpenAIChat, // 非预期的来源，但不应影响转换
		MessageID: "msg_123",
		Message: &types.StreamMessagePayload{
			Role: "assistant",
			Parts: []types.StreamContentPart{
				{
					Type: "text",
					Text: "Hello, world!",
				},
			},
		},
	}

	// 执行转换 - 不应返回错误
	event, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证转换结果 - 应按 Type 正常转换
	if event == nil {
		t.Fatal("event 不应为 nil")
	}
	if event.MessageStart == nil {
		t.Fatal("MessageStart 不应为 nil")
	}
	if event.MessageStart.Message.ID != "msg_123" {
		t.Errorf("Message.ID = %v, 期望 msg_123", event.MessageStart.Message.ID)
	}
}

// TestStreamEventFromContract_InvalidType 测试无效事件类型的错误处理
func TestStreamEventFromContract_InvalidType(t *testing.T) {
	log := logger.NewNopLogger()

	contract := &types.StreamEventContract{
		Type:   "invalid_type",
		Source: types.StreamSourceAnthropic,
	}

	// 执行转换
	_, err := StreamEventFromContract(contract, log)
	if err == nil {
		t.Fatal("期望返回错误")
	}
}

// TestStreamEventFromContract_NilEvent 测试 nil 事件的处理
func TestStreamEventFromContract_NilEvent(t *testing.T) {
	log := logger.NewNopLogger()

	// 执行转换
	event, err := StreamEventFromContract(nil, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if event != nil {
		t.Fatal("nil 事件应返回 nil")
	}
}

// TestStreamEventFromContract_RoundTrip 测试双向转换的可逆性
func TestStreamEventFromContract_RoundTrip(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	// 准备原始 Anthropic 事件
	original := &anthropicTypes.StreamEvent{
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
							Text: "Hello, world!",
						},
					},
				},
			},
		},
	}

	// Anthropic -> Contract
	contract, err := StreamEventToContract(original, ctx, log)
	if err != nil {
		t.Fatalf("ToContract 失败: %v", err)
	}

	// Contract -> Anthropic
	restored, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("FromContract 失败: %v", err)
	}

	// 验证关键字段
	if restored.MessageStart == nil {
		t.Fatal("MessageStart 不应为 nil")
	}
	if restored.MessageStart.Message.ID != original.MessageStart.Message.ID {
		t.Errorf("ID = %v, 期望 %v", restored.MessageStart.Message.ID, original.MessageStart.Message.ID)
	}
	if restored.MessageStart.Message.Role != original.MessageStart.Message.Role {
		t.Errorf("Role = %v, 期望 %v", restored.MessageStart.Message.Role, original.MessageStart.Message.Role)
	}
	if len(restored.MessageStart.Message.Content) != len(original.MessageStart.Message.Content) {
		t.Errorf("Content 长度 = %d, 期望 %d", len(restored.MessageStart.Message.Content), len(original.MessageStart.Message.Content))
	}
}

// TestStreamEventFromContract_RoundTrip_ToolUse 测试工具调用的双向转换
func TestStreamEventFromContract_RoundTrip_ToolUse(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	// 准备原始 Anthropic 事件
	original := &anthropicTypes.StreamEvent{
		ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
			Type:  anthropicTypes.StreamEventContentBlockStart,
			Index: 0,
			ContentBlock: anthropicTypes.ResponseContentBlock{
				ToolUse: &anthropicTypes.ToolUseBlock{
					Type:  anthropicTypes.ResponseContentBlockToolUse,
					ID:    "toolu_123",
					Name:  "get_weather",
					Input: map[string]interface{}{"location": "Beijing"},
				},
			},
		},
	}

	// Anthropic -> Contract
	contract, err := StreamEventToContract(original, ctx, log)
	if err != nil {
		t.Fatalf("ToContract 失败: %v", err)
	}

	// Contract -> Anthropic
	restored, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("FromContract 失败: %v", err)
	}

	// 验证关键字段
	if restored.ContentBlockStart == nil {
		t.Fatal("ContentBlockStart 不应为 nil")
	}
	if restored.ContentBlockStart.ContentBlock.ToolUse == nil {
		t.Fatal("ToolUse 不应为 nil")
	}
	if restored.ContentBlockStart.ContentBlock.ToolUse.ID != original.ContentBlockStart.ContentBlock.ToolUse.ID {
		t.Errorf("ToolUse.ID = %v, 期望 %v", restored.ContentBlockStart.ContentBlock.ToolUse.ID, original.ContentBlockStart.ContentBlock.ToolUse.ID)
	}
	if restored.ContentBlockStart.ContentBlock.ToolUse.Name != original.ContentBlockStart.ContentBlock.ToolUse.Name {
		t.Errorf("ToolUse.Name = %v, 期望 %v", restored.ContentBlockStart.ContentBlock.ToolUse.Name, original.ContentBlockStart.ContentBlock.ToolUse.Name)
	}
}

// TestStreamEventFromContract_RoundTrip_Thinking 测试思考内容的双向转换
func TestStreamEventFromContract_RoundTrip_Thinking(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := types.NewStreamIndexContext()

	// 准备原始 Anthropic 事件
	original := &anthropicTypes.StreamEvent{
		ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
			Type:  anthropicTypes.StreamEventContentBlockStart,
			Index: 0,
			ContentBlock: anthropicTypes.ResponseContentBlock{
				Thinking: &anthropicTypes.ThinkingBlock{
					Type:      anthropicTypes.ResponseContentBlockThinking,
					Thinking:  "Let me think...",
					Signature: "abc123",
				},
			},
		},
	}

	// Anthropic -> Contract
	contract, err := StreamEventToContract(original, ctx, log)
	if err != nil {
		t.Fatalf("ToContract 失败: %v", err)
	}

	// Contract -> Anthropic
	restored, err := StreamEventFromContract(contract, log)
	if err != nil {
		t.Fatalf("FromContract 失败: %v", err)
	}

	// 验证关键字段
	if restored.ContentBlockStart == nil {
		t.Fatal("ContentBlockStart 不应为 nil")
	}
	if restored.ContentBlockStart.ContentBlock.Thinking == nil {
		t.Fatal("Thinking 不应为 nil")
	}
	if restored.ContentBlockStart.ContentBlock.Thinking.Thinking != original.ContentBlockStart.ContentBlock.Thinking.Thinking {
		t.Errorf("Thinking.Thinking = %v, 期望 %v", restored.ContentBlockStart.ContentBlock.Thinking.Thinking, original.ContentBlockStart.ContentBlock.Thinking.Thinking)
	}
	if restored.ContentBlockStart.ContentBlock.Thinking.Signature != original.ContentBlockStart.ContentBlock.Thinking.Signature {
		t.Errorf("Thinking.Signature = %v, 期望 %v", restored.ContentBlockStart.ContentBlock.Thinking.Signature, original.ContentBlockStart.ContentBlock.Thinking.Signature)
	}
}
