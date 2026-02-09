package converter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MeowSalty/portal/logger"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// TestResponseToContract_BasicConversion 测试基本响应的转换
func TestResponseToContract_BasicConversion(t *testing.T) {
	log := logger.NewNopLogger()

	// 准备测试数据
	model := "claude-3-5-sonnet-20241022"
	text := "Hello, world!"
	inputTokens := 10
	outputTokens := 5
	stopReason := anthropicTypes.StopReasonEndTurn

	resp := &anthropicTypes.Response{
		ID:    "msg_123",
		Type:  anthropicTypes.ResponseTypeMessage,
		Role:  anthropicTypes.RoleAssistant,
		Model: model,
		Content: []anthropicTypes.ResponseContentBlock{
			{
				Text: &anthropicTypes.TextBlock{
					Type: anthropicTypes.ResponseContentBlockText,
					Text: text,
				},
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  &inputTokens,
			OutputTokens: &outputTokens,
		},
	}

	// 执行转换
	contract, err := ResponseToContract(resp, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract == nil {
		t.Fatal("contract 不应为 nil")
	}
	if contract.Source != types.VendorSourceAnthropic {
		t.Errorf("Source = %v, 期望 %v", contract.Source, types.VendorSourceAnthropic)
	}
	if contract.ID != "msg_123" {
		t.Errorf("ID = %v, 期望 msg_123", contract.ID)
	}
	if contract.Model == nil || *contract.Model != model {
		t.Errorf("Model = %v, 期望 %v", contract.Model, model)
	}
	if len(contract.Choices) != 1 {
		t.Fatalf("Choices 长度 = %d, 期望 1", len(contract.Choices))
	}
	if contract.Choices[0].Message == nil {
		t.Fatal("Message 不应为 nil")
	}
	if contract.Choices[0].Message.Content == nil || *contract.Choices[0].Message.Content != text {
		t.Errorf("Content = %v, 期望 %v", contract.Choices[0].Message.Content, text)
	}
}

// TestResponseToContract_FinishReasonMapping 测试完成原因的映射
func TestResponseToContract_FinishReasonMapping(t *testing.T) {
	log := logger.NewNopLogger()

	tests := []struct {
		name           string
		stopReason     anthropicTypes.StopReason
		expectedReason types.ResponseFinishReason
	}{
		{"EndTurn", anthropicTypes.StopReasonEndTurn, types.ResponseFinishReasonStop},
		{"StopSeq", anthropicTypes.StopReasonStopSeq, types.ResponseFinishReasonStop},
		{"MaxTokens", anthropicTypes.StopReasonMaxTokens, types.ResponseFinishReasonLength},
		{"ToolUse", anthropicTypes.StopReasonToolUse, types.ResponseFinishReasonToolCalls},
		{"Refusal", anthropicTypes.StopReasonRefusal, types.ResponseFinishReasonContentFilter},
		{"PauseTurn", anthropicTypes.StopReasonPauseTurn, types.ResponseFinishReasonUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &anthropicTypes.Response{
				ID:         "msg_test",
				Type:       anthropicTypes.ResponseTypeMessage,
				Role:       anthropicTypes.RoleAssistant,
				Model:      "claude-3-5-sonnet-20241022",
				Content:    []anthropicTypes.ResponseContentBlock{},
				StopReason: &tt.stopReason,
			}

			contract, err := ResponseToContract(resp, log)
			if err != nil {
				t.Fatalf("转换失败: %v", err)
			}

			if len(contract.Choices) == 0 {
				t.Fatal("Choices 不应为空")
			}
			if contract.Choices[0].FinishReason == nil {
				t.Fatal("FinishReason 不应为 nil")
			}
			if *contract.Choices[0].FinishReason != tt.expectedReason {
				t.Errorf("FinishReason = %v, 期望 %v", *contract.Choices[0].FinishReason, tt.expectedReason)
			}
		})
	}
}

// TestResponseToContract_ToolCalls 测试工具调用的转换
func TestResponseToContract_ToolCalls(t *testing.T) {
	log := logger.NewNopLogger()

	toolID := "toolu_123"
	toolName := "get_weather"
	stopReason := anthropicTypes.StopReasonToolUse

	resp := &anthropicTypes.Response{
		ID:    "msg_123",
		Type:  anthropicTypes.ResponseTypeMessage,
		Role:  anthropicTypes.RoleAssistant,
		Model: "claude-3-5-sonnet-20241022",
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
		StopReason: &stopReason,
	}

	contract, err := ResponseToContract(resp, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if len(contract.Choices) == 0 {
		t.Fatal("Choices 不应为空")
	}
	if contract.Choices[0].Message == nil {
		t.Fatal("Message 不应为 nil")
	}
	if len(contract.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("ToolCalls 长度 = %d, 期望 1", len(contract.Choices[0].Message.ToolCalls))
	}

	toolCall := contract.Choices[0].Message.ToolCalls[0]
	if toolCall.ID == nil || *toolCall.ID != toolID {
		t.Errorf("ToolCall ID = %v, 期望 %v", toolCall.ID, toolID)
	}
	if toolCall.Name == nil || *toolCall.Name != toolName {
		t.Errorf("ToolCall Name = %v, 期望 %v", toolCall.Name, toolName)
	}
}

// TestResponseToContract_Usage 测试使用统计的转换
func TestResponseToContract_Usage(t *testing.T) {
	log := logger.NewNopLogger()

	inputTokens := 100
	outputTokens := 50
	cacheCreation := 10
	cacheRead := 20

	resp := &anthropicTypes.Response{
		ID:    "msg_123",
		Type:  anthropicTypes.ResponseTypeMessage,
		Role:  anthropicTypes.RoleAssistant,
		Model: "claude-3-5-sonnet-20241022",
		Content: []anthropicTypes.ResponseContentBlock{
			{
				Text: &anthropicTypes.TextBlock{
					Type: anthropicTypes.ResponseContentBlockText,
					Text: "Test",
				},
			},
		},
		Usage: &anthropicTypes.Usage{
			InputTokens:              &inputTokens,
			OutputTokens:             &outputTokens,
			CacheCreationInputTokens: &cacheCreation,
			CacheReadInputTokens:     &cacheRead,
		},
	}

	contract, err := ResponseToContract(resp, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
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
	if contract.Usage.TotalTokens == nil || *contract.Usage.TotalTokens != inputTokens+outputTokens {
		t.Errorf("TotalTokens = %v, 期望 %v", contract.Usage.TotalTokens, inputTokens+outputTokens)
	}

	// 验证 Extras 中的缓存字段
	if val, ok := contract.Usage.Extras["anthropic.cache_creation_input_tokens"]; !ok || val != cacheCreation {
		t.Errorf("cache_creation_input_tokens = %v, 期望 %v", val, cacheCreation)
	}
	if val, ok := contract.Usage.Extras["anthropic.cache_read_input_tokens"]; !ok || val != cacheRead {
		t.Errorf("cache_read_input_tokens = %v, 期望 %v", val, cacheRead)
	}
}

// TestResponseToContract_RoundTrip 测试双向转换的可逆性
func TestResponseToContract_RoundTrip(t *testing.T) {
	log := logger.NewNopLogger()

	// 准备原始响应
	model := "claude-3-5-sonnet-20241022"
	text := "Hello, world!"
	inputTokens := 10
	outputTokens := 5
	stopReason := anthropicTypes.StopReasonEndTurn

	original := &anthropicTypes.Response{
		ID:    "msg_123",
		Type:  anthropicTypes.ResponseTypeMessage,
		Role:  anthropicTypes.RoleAssistant,
		Model: model,
		Content: []anthropicTypes.ResponseContentBlock{
			{
				Text: &anthropicTypes.TextBlock{
					Type: anthropicTypes.ResponseContentBlockText,
					Text: text,
				},
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  &inputTokens,
			OutputTokens: &outputTokens,
		},
	}

	// RequestToContract
	contract, err := ResponseToContract(original, log)
	if err != nil {
		t.Fatalf("ToContract 失败: %v", err)
	}

	// FromContract
	restored, err := ResponseFromContract(contract, log)
	if err != nil {
		t.Fatalf("FromContract 失败: %v", err)
	}

	// 验证关键字段
	if restored.ID != original.ID {
		t.Errorf("ID = %v, 期望 %v", restored.ID, original.ID)
	}
	if restored.Model != original.Model {
		t.Errorf("Model = %v, 期望 %v", restored.Model, original.Model)
	}
	if restored.StopReason == nil || *restored.StopReason != *original.StopReason {
		t.Errorf("StopReason = %v, 期望 %v", restored.StopReason, original.StopReason)
	}
	if len(restored.Content) != len(original.Content) {
		t.Errorf("Content 长度 = %d, 期望 %d", len(restored.Content), len(original.Content))
	}
}

// TestResponseErrorToContract 测试错误响应的转换
func TestResponseToContract_VendorExtrasAndRawFidelity(t *testing.T) {
	log := logger.NewNopLogger()
	stopSequence := "stop_here"
	stopReason := anthropicTypes.StopReasonStopSeq
	citation := anthropicTypes.TextCitation{
		CharLocation: &anthropicTypes.CitationCharLocation{
			Type:           anthropicTypes.TextCitationTypeCharLocation,
			CitedText:      "quote",
			DocumentIndex:  2,
			DocumentTitle:  "doc-title",
			FileID:         "file_123",
			StartCharIndex: 3,
			EndCharIndex:   8,
		},
	}

	resp := &anthropicTypes.Response{
		ID:    "msg_456",
		Type:  anthropicTypes.ResponseTypeMessage,
		Role:  anthropicTypes.RoleAssistant,
		Model: "claude-3-5-sonnet-20241022",
		Content: []anthropicTypes.ResponseContentBlock{
			{
				Text: &anthropicTypes.TextBlock{
					Type:      anthropicTypes.ResponseContentBlockText,
					Text:      "hello",
					Citations: []anthropicTypes.TextCitation{citation},
				},
			},
			{
				Thinking: &anthropicTypes.ThinkingBlock{
					Type:      anthropicTypes.ResponseContentBlockThinking,
					Thinking:  "thinking",
					Signature: "sig_abc",
				},
			},
			{
				RedactedThinking: &anthropicTypes.RedactedThinkingBlock{
					Type: anthropicTypes.ResponseContentBlockRedactedThinking,
					Data: "redacted_data",
				},
			},
			{
				ToolUse: &anthropicTypes.ToolUseBlock{
					Type:  anthropicTypes.ResponseContentBlockToolUse,
					ID:    "tool_1",
					Name:  "calc",
					Input: map[string]interface{}{"a": 1},
				},
			},
			{
				ServerToolUse: &anthropicTypes.ServerToolUseBlock{
					Type:  anthropicTypes.ResponseContentBlockServerToolUse,
					ID:    "server_tool_1",
					Name:  "server_search",
					Input: map[string]interface{}{"query": "q"},
				},
			},
			{
				WebSearchToolResult: &anthropicTypes.WebSearchToolResultBlock{
					Type:      anthropicTypes.ResponseContentBlockWebSearchToolResult,
					ToolUseID: "tool_1",
					Content: anthropicTypes.WebSearchToolResultBlockContent{
						Error: &anthropicTypes.WebSearchToolResultError{
							Type:      "web_search_tool_result_error",
							ErrorCode: "bad_request",
						},
					},
				},
			},
		},
		StopReason:   &stopReason,
		StopSequence: &stopSequence,
	}

	contract, err := ResponseToContract(resp, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	assert.NotNil(t, contract.Extras, "响应 Extras 不应为 nil")
	stopSequenceRaw, ok := GetVendorExtraRaw("anthropic.stop_sequence", contract.Extras)
	assert.True(t, ok, "StopSequence 扩展应存在")
	assert.Equal(t, stopSequence, string(stopSequenceRaw), "StopSequence 应保持一致")

	if len(contract.Choices) == 0 || contract.Choices[0].Message == nil {
		t.Fatalf("缺少消息")
	}
	message := contract.Choices[0].Message
	assert.NotNil(t, message.Extras, "Message.Extras 不应为 nil")
	contentBlocksRaw, ok := GetVendorExtraRaw("anthropic.content_blocks", message.Extras)
	assert.True(t, ok, "应保存原始 content_blocks")
	var contentBlocks []anthropicTypes.ResponseContentBlock
	assert.NoError(t, json.Unmarshal(contentBlocksRaw, &contentBlocks), "反序列化 content_blocks 失败")
	assert.Equal(t, len(resp.Content), len(contentBlocks), "content_blocks 长度应一致")

	var thinkingPart *types.ResponseContentPart
	var redactedPart *types.ResponseContentPart
	var textPart *types.ResponseContentPart
	for i := range message.Parts {
		part := &message.Parts[i]
		if part.Type == "text" && textPart == nil {
			textPart = part
		}
		var signature string
		if part.Type == "thinking" && part.Extras != nil {
			if found, _ := GetVendorExtra("anthropic.signature", part.Extras, &signature); found {
				thinkingPart = part
			}
			var redacted bool
			if found, _ := GetVendorExtra("anthropic.redacted", part.Extras, &redacted); found && redacted {
				redactedPart = part
			}
		}
	}
	if assert.NotNil(t, thinkingPart, "思考块扩展应存在") {
		signatureRaw, ok := GetVendorExtraRaw("anthropic.signature", thinkingPart.Extras)
		assert.True(t, ok, "思考签名应存在")
		assert.Equal(t, "sig_abc", string(signatureRaw), "思考签名应保持一致")
	}
	if assert.NotNil(t, redactedPart, "脱敏思考扩展应存在") {
		var redacted bool
		found, err := GetVendorExtra("anthropic.redacted", redactedPart.Extras, &redacted)
		assert.NoError(t, err, "读取脱敏标记失败")
		assert.True(t, found, "脱敏标记应存在")
		assert.True(t, redacted, "脱敏标记应为 true")
		dataRaw, ok := GetVendorExtraRaw("anthropic.data", redactedPart.Extras)
		assert.True(t, ok, "脱敏数据应存在")
		assert.Equal(t, "redacted_data", string(dataRaw), "脱敏数据应保持一致")
	}
	if assert.NotNil(t, textPart, "文本块应存在") {
		if assert.NotEmpty(t, textPart.Annotations, "文本块注释应存在") {
			annotation := textPart.Annotations[0]
			assert.NotNil(t, annotation.Extras, "Annotation.Extras 不应为 nil")
			citationTypeRaw, ok := GetVendorExtraRaw("anthropic.citation_type", annotation.Extras)
			assert.True(t, ok, "citation_type 应存在")
			assert.Equal(t, "char_location", string(citationTypeRaw), "citation_type 应为 char_location")
			docTitleRaw, ok := GetVendorExtraRaw("anthropic.document_title", annotation.Extras)
			assert.True(t, ok, "document_title 应存在")
			assert.Equal(t, "doc-title", string(docTitleRaw), "document_title 应保持一致")
		}
	}

	var serverToolCall *types.ResponseToolCall
	for i := range message.ToolCalls {
		call := &message.ToolCalls[i]
		var isServerTool bool
		if call.Extras != nil {
			if found, _ := GetVendorExtra("anthropic.server_tool", call.Extras, &isServerTool); found && isServerTool {
				serverToolCall = call
				break
			}
		}
	}
	assert.NotNil(t, serverToolCall, "服务器工具调用扩展应存在")

	var webSearchResult *types.ResponseToolResult
	for i := range message.ToolResults {
		result := &message.ToolResults[i]
		var isWebSearch bool
		if result.Extras != nil {
			if found, _ := GetVendorExtra("anthropic.web_search_result", result.Extras, &isWebSearch); found && isWebSearch {
				webSearchResult = result
				break
			}
		}
	}
	if assert.NotNil(t, webSearchResult, "WebSearch 工具结果扩展应存在") {
		var hasError bool
		found, err := GetVendorExtra("anthropic.has_error", webSearchResult.Extras, &hasError)
		assert.NoError(t, err, "读取 WebSearch 错误标记失败")
		assert.True(t, found, "WebSearch 错误标记应存在")
		assert.True(t, hasError, "WebSearch 错误标记应为 true")
	}
}

func TestResponseErrorToContract(t *testing.T) {
	errResp := &anthropicTypes.ErrorResponse{
		Type: "error",
		Error: anthropicTypes.Error{
			Type:    "invalid_request_error",
			Message: "Invalid API key",
		},
	}

	contract, err := ResponseErrorToContract(errResp)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if contract == nil {
		t.Fatal("contract 不应为 nil")
	}
	if contract.Source != types.VendorSourceAnthropic {
		t.Errorf("Source = %v, 期望 %v", contract.Source, types.VendorSourceAnthropic)
	}
	if contract.Error == nil {
		t.Fatal("Error 不应为 nil")
	}
	if contract.Error.Type == nil || *contract.Error.Type != "invalid_request_error" {
		t.Errorf("Error.Type = %v, 期望 invalid_request_error", contract.Error.Type)
	}
	if contract.Error.Message == nil || *contract.Error.Message != "Invalid API key" {
		t.Errorf("Error.Message = %v, 期望 Invalid API key", contract.Error.Message)
	}
}
