package types_test

import (
	"testing"

	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
)

// TestConvertCoreResponse_SimpleText 测试简单文本消息的转换
func TestConvertCoreResponse_SimpleText(t *testing.T) {
	// 构造 Anthropic 响应
	text := "你好，我是 Claude。"
	stopReason := "end_turn"
	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_01AbCdEfGhIjKlMnOpQr",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-5-sonnet-20241022",
		Content: []anthropicTypes.ResponseContent{
			{
				Type: "text",
				Text: &text,
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  10,
			OutputTokens: 20,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证基本字段
	if result.ID != "msg_01AbCdEfGhIjKlMnOpQr" {
		t.Errorf("期望 ID 为 'msg_01AbCdEfGhIjKlMnOpQr'，实际为 '%s'", result.ID)
	}

	if result.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("期望模型为 'claude-3-5-sonnet-20241022'，实际为 '%s'", result.Model)
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证消息
	if choice.Message == nil {
		t.Fatal("期望消息不为空")
	}

	if choice.Message.Role != "assistant" {
		t.Errorf("期望角色为 'assistant'，实际为 '%s'", choice.Message.Role)
	}

	if choice.Message.Content == nil {
		t.Fatal("期望内容不为空")
	}

	if *choice.Message.Content != "你好，我是 Claude。" {
		t.Errorf("期望内容为 '你好，我是 Claude。'，实际为 '%s'", *choice.Message.Content)
	}

	// 验证停止原因
	if choice.FinishReason == nil {
		t.Fatal("期望停止原因已设置")
	}

	if *choice.FinishReason != "stop" {
		t.Errorf("期望停止原因为 'stop'，实际为 '%s'", *choice.FinishReason)
	}

	// 验证使用统计
	if result.Usage == nil {
		t.Fatal("期望使用统计已设置")
	}

	if result.Usage.PromptTokens != 10 {
		t.Errorf("期望输入 token 数为 10，实际为 %d", result.Usage.PromptTokens)
	}

	if result.Usage.CompletionTokens != 20 {
		t.Errorf("期望输出 token 数为 20，实际为 %d", result.Usage.CompletionTokens)
	}

	if result.Usage.TotalTokens != 30 {
		t.Errorf("期望总 token 数为 30，实际为 %d", result.Usage.TotalTokens)
	}
}

// TestConvertCoreResponse_MultipleTextBlocks 测试多个文本块的转换
func TestConvertCoreResponse_MultipleTextBlocks(t *testing.T) {
	// 构造 Anthropic 响应（多个文本块）
	text1 := "这是第一段内容。"
	text2 := "这是第二段内容。"
	stopReason := "end_turn"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_02MultiText",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-opus-20240229",
		Content: []anthropicTypes.ResponseContent{
			{
				Type: "text",
				Text: &text1,
			},
			{
				Type: "text",
				Text: &text2,
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  15,
			OutputTokens: 25,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 对于多个文本块，应该返回字符串形式（而不是 ContentParts）
	// 这取决于实现，如果只有一个文本块返回字符串，多个则可能返回 nil
	// 根据代码逻辑，多个块不会转为字符串
	if choice.Message == nil {
		t.Fatal("期望消息不为空")
	}

	// 内容应该为空，因为有多个内容块
	if choice.Message.Content != nil {
		t.Error("期望内容为空（多个内容块时）")
	}
}

// TestConvertCoreResponse_WithToolUse 测试包含工具调用的转换
func TestConvertCoreResponse_WithToolUse(t *testing.T) {
	// 构造 Anthropic 响应（包含工具调用）
	text := "让我帮你查询北京的天气。"
	toolID := "toolu_01A2B3C4D5E6F7G8H9I0"
	toolName := "get_weather"
	stopReason := "tool_use"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_03ToolUse",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-5-sonnet-20241022",
		Content: []anthropicTypes.ResponseContent{
			{
				Type: "text",
				Text: &text,
			},
			{
				Type:  "tool_use",
				ID:    &toolID,
				Name:  &toolName,
				Input: map[string]interface{}{"location": "北京", "unit": "celsius"},
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  20,
			OutputTokens: 30,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证停止原因
	if choice.FinishReason == nil {
		t.Fatal("期望停止原因已设置")
	}

	if *choice.FinishReason != "tool_calls" {
		t.Errorf("期望停止原因为 'tool_calls'，实际为 '%s'", *choice.FinishReason)
	}

	// 验证消息不为空
	if choice.Message == nil {
		t.Fatal("期望消息不为空")
	}

	// 对于包含多个内容块（文本 + 工具调用），Content 应该为 nil
	if choice.Message.Content != nil {
		t.Error("期望内容为空（多个内容块时）")
	}
}

// TestConvertCoreResponse_OnlyToolUse 测试仅包含工具调用的转换
func TestConvertCoreResponse_OnlyToolUse(t *testing.T) {
	// 构造 Anthropic 响应（仅工具调用）
	toolID := "toolu_01WeatherQuery"
	toolName := "get_weather"
	stopReason := "tool_use"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_04OnlyTool",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-5-sonnet-20241022",
		Content: []anthropicTypes.ResponseContent{
			{
				Type:  "tool_use",
				ID:    &toolID,
				Name:  &toolName,
				Input: map[string]interface{}{"location": "上海"},
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  10,
			OutputTokens: 15,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证停止原因
	if choice.FinishReason == nil || *choice.FinishReason != "tool_calls" {
		t.Error("期望停止原因为 'tool_calls'")
	}

	// 验证消息
	if choice.Message == nil {
		t.Fatal("期望消息不为空")
	}

	// 单个工具调用不会转为字符串形式
	if choice.Message.Content != nil {
		t.Error("期望内容为空（非文本内容块时）")
	}
}

// TestConvertCoreResponse_MultipleToolUses 测试多个工具调用的转换
func TestConvertCoreResponse_MultipleToolUses(t *testing.T) {
	// 构造 Anthropic 响应（多个工具调用）
	toolID1 := "toolu_01First"
	toolName1 := "get_weather"
	toolID2 := "toolu_02Second"
	toolName2 := "get_time"
	stopReason := "tool_use"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_05MultiTools",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-5-sonnet-20241022",
		Content: []anthropicTypes.ResponseContent{
			{
				Type:  "tool_use",
				ID:    &toolID1,
				Name:  &toolName1,
				Input: map[string]interface{}{"location": "北京"},
			},
			{
				Type:  "tool_use",
				ID:    &toolID2,
				Name:  &toolName2,
				Input: map[string]interface{}{"timezone": "Asia/Shanghai"},
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  15,
			OutputTokens: 20,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证停止原因
	if choice.FinishReason == nil || *choice.FinishReason != "tool_calls" {
		t.Error("期望停止原因为 'tool_calls'")
	}
}

// TestConvertCoreResponse_WithThinking 测试包含思考内容的转换
func TestConvertCoreResponse_WithThinking(t *testing.T) {
	// 构造 Anthropic 响应（包含思考内容）
	text := "答案是 42。"
	thinking := "我需要仔细思考这个问题..."
	stopReason := "end_turn"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_06Thinking",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-5-sonnet-20241022",
		Content: []anthropicTypes.ResponseContent{
			{
				Type:     "thinking",
				Thinking: &thinking,
			},
			{
				Type: "text",
				Text: &text,
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  10,
			OutputTokens: 25,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证消息不为空
	if choice.Message == nil {
		t.Fatal("期望消息不为空")
	}

	// 包含多个内容块，Content 应该为 nil
	if choice.Message.Content != nil {
		t.Error("期望内容为空（多个内容块时）")
	}
}

// TestConvertCoreResponse_WithWebSearchResult 测试包含网络搜索结果的转换
func TestConvertCoreResponse_WithWebSearchResult(t *testing.T) {
	// 构造 Anthropic 响应（包含网络搜索结果）
	stopReason := "end_turn"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_07WebSearch",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-5-sonnet-20241022",
		Content: []anthropicTypes.ResponseContent{
			{
				Type: "web_search_tool_result",
				Content: []anthropicTypes.WebSearchToolResult{
					{
						Type:             "web_search_result",
						Title:            "测试标题",
						URL:              "https://example.com",
						EncryptedContent: "encrypted_data",
					},
				},
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  10,
			OutputTokens: 15,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证消息不为空
	if choice.Message == nil {
		t.Fatal("期望消息不为空")
	}
}

// TestConvertCoreResponse_EmptyContent 测试空内容的转换
func TestConvertCoreResponse_EmptyContent(t *testing.T) {
	// 构造 Anthropic 响应（空内容）
	stopReason := "end_turn"

	anthropicResp := &anthropicTypes.Response{
		ID:         "msg_08EmptyContent",
		Type:       "message",
		Role:       "assistant",
		Model:      "claude-3-opus-20240229",
		Content:    []anthropicTypes.ResponseContent{},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  5,
			OutputTokens: 0,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证消息不为空
	if choice.Message == nil {
		t.Fatal("期望消息不为空")
	}

	// 空内容应该返回 nil
	if choice.Message.Content != nil {
		t.Error("期望内容为空")
	}
}

// TestConvertCoreResponse_NilContent 测试 nil 内容的转换
func TestConvertCoreResponse_NilContent(t *testing.T) {
	// 构造 Anthropic 响应（nil 内容）
	stopReason := "end_turn"

	anthropicResp := &anthropicTypes.Response{
		ID:         "msg_09NilContent",
		Type:       "message",
		Role:       "assistant",
		Model:      "claude-3-opus-20240229",
		Content:    nil,
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  5,
			OutputTokens: 0,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证消息不为空
	if choice.Message == nil {
		t.Fatal("期望消息不为空")
	}

	// nil 内容应该返回 nil
	if choice.Message.Content != nil {
		t.Error("期望内容为空")
	}
}

// TestConvertCoreResponse_StopReasonMapping 测试停止原因映射
func TestConvertCoreResponse_StopReasonMapping(t *testing.T) {
	testCases := []struct {
		name                 string
		anthropicStopReason  string
		expectedFinishReason string
	}{
		{
			name:                 "end_turn to stop",
			anthropicStopReason:  "end_turn",
			expectedFinishReason: "stop",
		},
		{
			name:                 "max_tokens to length",
			anthropicStopReason:  "max_tokens",
			expectedFinishReason: "length",
		},
		{
			name:                 "tool_use to tool_calls",
			anthropicStopReason:  "tool_use",
			expectedFinishReason: "tool_calls",
		},
		{
			name:                 "stop_sequence to stop",
			anthropicStopReason:  "stop_sequence",
			expectedFinishReason: "stop",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			text := "测试内容"
			anthropicResp := &anthropicTypes.Response{
				ID:    "msg_10StopReason",
				Type:  "message",
				Role:  "assistant",
				Model: "claude-3-5-sonnet-20241022",
				Content: []anthropicTypes.ResponseContent{
					{
						Type: "text",
						Text: &text,
					},
				},
				StopReason: &tc.anthropicStopReason,
				Usage: &anthropicTypes.Usage{
					InputTokens:  10,
					OutputTokens: 15,
				},
			}

			result := anthropicResp.ConvertCoreResponse()

			if result == nil {
				t.Fatal("期望 ConvertCoreResponse 返回非空结果")
			}

			choice := result.Choices[0]

			if choice.FinishReason == nil {
				t.Fatal("期望停止原因已设置")
			}

			if *choice.FinishReason != tc.expectedFinishReason {
				t.Errorf("期望停止原因为 '%s'，实际为 '%s'", tc.expectedFinishReason, *choice.FinishReason)
			}
		})
	}
}

// TestConvertCoreResponse_NoStopReason 测试没有停止原因的转换
func TestConvertCoreResponse_NoStopReason(t *testing.T) {
	// 构造 Anthropic 响应（无停止原因）
	text := "测试内容"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_11NoStopReason",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-opus-20240229",
		Content: []anthropicTypes.ResponseContent{
			{
				Type: "text",
				Text: &text,
			},
		},
		StopReason: nil,
		Usage: &anthropicTypes.Usage{
			InputTokens:  10,
			OutputTokens: 15,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证停止原因为空
	if choice.FinishReason != nil {
		t.Error("期望停止原因为空")
	}
}

// TestConvertCoreResponse_NoUsage 测试没有使用统计的转换
func TestConvertCoreResponse_NoUsage(t *testing.T) {
	// 构造 Anthropic 响应（无使用统计）
	text := "测试内容"
	stopReason := "end_turn"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_12NoUsage",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-opus-20240229",
		Content: []anthropicTypes.ResponseContent{
			{
				Type: "text",
				Text: &text,
			},
		},
		StopReason: &stopReason,
		Usage:      nil,
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证使用统计为空
	if result.Usage != nil {
		t.Error("期望使用统计为空")
	}
}

// TestConvertCoreResponse_ZeroUsage 测试使用统计为零的转换
func TestConvertCoreResponse_ZeroUsage(t *testing.T) {
	// 构造 Anthropic 响应（使用统计为零）
	text := "测试内容"
	stopReason := "end_turn"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_13ZeroUsage",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-opus-20240229",
		Content: []anthropicTypes.ResponseContent{
			{
				Type: "text",
				Text: &text,
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  0,
			OutputTokens: 0,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证使用统计为空（因为都是 0）
	if result.Usage != nil {
		t.Error("期望使用统计为空（当 token 数为 0 时）")
	}
}

// TestConvertCoreResponse_ComplexMultiRound 测试复杂多轮对话场景
func TestConvertCoreResponse_ComplexMultiRound(t *testing.T) {
	// 构造 Anthropic 响应（复杂场景：文本 + 思考 + 工具调用）
	text1 := "让我思考一下..."
	thinking := "我需要获取天气信息"
	toolID := "toolu_01Complex"
	toolName := "get_weather"
	text2 := "好的，我来查询天气。"
	stopReason := "tool_use"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_14Complex",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-5-sonnet-20241022",
		Content: []anthropicTypes.ResponseContent{
			{
				Type: "text",
				Text: &text1,
			},
			{
				Type:     "thinking",
				Thinking: &thinking,
			},
			{
				Type: "text",
				Text: &text2,
			},
			{
				Type:  "tool_use",
				ID:    &toolID,
				Name:  &toolName,
				Input: map[string]interface{}{"location": "北京", "unit": "celsius"},
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  25,
			OutputTokens: 40,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证基本字段
	if result.ID != "msg_14Complex" {
		t.Errorf("期望 ID 为 'msg_14Complex'，实际为 '%s'", result.ID)
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证停止原因
	if choice.FinishReason == nil || *choice.FinishReason != "tool_calls" {
		t.Error("期望停止原因为 'tool_calls'")
	}

	// 验证消息不为空
	if choice.Message == nil {
		t.Fatal("期望消息不为空")
	}

	// 验证角色
	if choice.Message.Role != "assistant" {
		t.Errorf("期望角色为 'assistant'，实际为 '%s'", choice.Message.Role)
	}

	// 多个内容块时，Content 应该为 nil
	if choice.Message.Content != nil {
		t.Error("期望内容为空（多个内容块时）")
	}

	// 验证使用统计
	if result.Usage == nil {
		t.Fatal("期望使用统计已设置")
	}

	if result.Usage.PromptTokens != 25 {
		t.Errorf("期望输入 token 数为 25，实际为 %d", result.Usage.PromptTokens)
	}

	if result.Usage.CompletionTokens != 40 {
		t.Errorf("期望输出 token 数为 40，实际为 %d", result.Usage.CompletionTokens)
	}

	if result.Usage.TotalTokens != 65 {
		t.Errorf("期望总 token 数为 65，实际为 %d", result.Usage.TotalTokens)
	}
}

// TestConvertCoreResponse_ServerToolUse 测试服务器工具使用的转换
func TestConvertCoreResponse_ServerToolUse(t *testing.T) {
	// 构造 Anthropic 响应（服务器工具使用）
	toolID := "toolu_01ServerTool"
	toolName := "server_function"
	stopReason := "tool_use"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_15ServerTool",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-5-sonnet-20241022",
		Content: []anthropicTypes.ResponseContent{
			{
				Type:  "server_tool_use",
				ID:    &toolID,
				Name:  &toolName,
				Input: map[string]interface{}{"param": "value"},
			},
		},
		StopReason: &stopReason,
		Usage: &anthropicTypes.Usage{
			InputTokens:  10,
			OutputTokens: 15,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证停止原因
	if choice.FinishReason == nil || *choice.FinishReason != "tool_calls" {
		t.Error("期望停止原因为 'tool_calls'")
	}

	// 验证消息不为空
	if choice.Message == nil {
		t.Fatal("期望消息不为空")
	}
}

// TestConvertCoreResponse_AllFields 测试所有字段的转换
func TestConvertCoreResponse_AllFields(t *testing.T) {
	// 构造 Anthropic 响应（所有字段）
	text := "完整的响应测试"
	stopReason := "end_turn"
	stopSequence := "END"

	anthropicResp := &anthropicTypes.Response{
		ID:    "msg_16AllFields",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-5-sonnet-20241022",
		Content: []anthropicTypes.ResponseContent{
			{
				Type: "text",
				Text: &text,
			},
		},
		StopReason:   &stopReason,
		StopSequence: &stopSequence,
		Usage: &anthropicTypes.Usage{
			InputTokens:  20,
			OutputTokens: 30,
		},
	}

	// 调用转换函数
	result := anthropicResp.ConvertCoreResponse()

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证所有基本字段
	if result.ID != "msg_16AllFields" {
		t.Errorf("期望 ID 为 'msg_16AllFields'，实际为 '%s'", result.ID)
	}

	if result.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("期望模型为 'claude-3-5-sonnet-20241022'，实际为 '%s'", result.Model)
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证消息
	if choice.Message == nil {
		t.Fatal("期望消息不为空")
	}

	if choice.Message.Role != "assistant" {
		t.Errorf("期望角色为 'assistant'，实际为 '%s'", choice.Message.Role)
	}

	if choice.Message.Content == nil || *choice.Message.Content != text {
		t.Error("期望内容正确设置")
	}

	// 验证停止原因
	if choice.FinishReason == nil || *choice.FinishReason != "stop" {
		t.Error("期望停止原因为 'stop'")
	}

	// 验证使用统计
	if result.Usage == nil {
		t.Fatal("期望使用统计已设置")
	}

	if result.Usage.PromptTokens != 20 {
		t.Errorf("期望输入 token 数为 20，实际为 %d", result.Usage.PromptTokens)
	}

	if result.Usage.CompletionTokens != 30 {
		t.Errorf("期望输出 token 数为 30，实际为 %d", result.Usage.CompletionTokens)
	}

	if result.Usage.TotalTokens != 50 {
		t.Errorf("期望总 token 数为 50，实际为 %d", result.Usage.TotalTokens)
	}
}
