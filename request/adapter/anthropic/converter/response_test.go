package converter_test

import (
	"encoding/json"
	"testing"

	"github.com/MeowSalty/portal/request/adapter/anthropic/converter"
	"github.com/MeowSalty/portal/request/adapter/anthropic/types"
	coreTypes "github.com/MeowSalty/portal/types"
)

// TestConvertResponse_SimpleText 测试简单文本消息的转换
func TestConvertResponse_SimpleText(t *testing.T) {
	// 构造核心响应
	content := "你好，我是 Claude。"
	finishReason := "stop"
	coreResp := &coreTypes.Response{
		ID:    "msg_01AbCdEfGhIjKlMnOpQr",
		Model: "claude-3-5-sonnet-20241022",
		Choices: []coreTypes.Choice{
			{
				Message: &coreTypes.ResponseMessage{
					Role:    "assistant",
					Content: &content,
				},
				FinishReason: &finishReason,
			},
		},
		Usage: &coreTypes.ResponseUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	// 调用转换函数
	result := converter.ConvertResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	// 验证基本字段
	if result.ID != "msg_01AbCdEfGhIjKlMnOpQr" {
		t.Errorf("期望ID为 'msg_01AbCdEfGhIjKlMnOpQr'，实际为 '%s'", result.ID)
	}

	if result.Type != "message" {
		t.Errorf("期望Type为 'message'，实际为 '%s'", result.Type)
	}

	if result.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("期望模型为 'claude-3-5-sonnet-20241022'，实际为 '%s'", result.Model)
	}

	if result.Role != "assistant" {
		t.Errorf("期望角色为 'assistant'，实际为 '%s'", result.Role)
	}

	// 验证内容
	if len(result.Content) != 1 {
		t.Fatalf("期望有 1 个内容块，实际为 %d", len(result.Content))
	}

	if result.Content[0].Type != "text" {
		t.Errorf("期望内容类型为 'text'，实际为 '%s'", result.Content[0].Type)
	}

	if result.Content[0].Text == nil {
		t.Fatal("期望文本内容已设置")
	}

	if *result.Content[0].Text != "你好，我是 Claude。" {
		t.Errorf("期望文本内容为 '你好，我是 Claude。'，实际为 '%s'", *result.Content[0].Text)
	}

	// 验证停止原因
	if result.StopReason == nil {
		t.Fatal("期望停止原因已设置")
	}

	if *result.StopReason != "end_turn" {
		t.Errorf("期望停止原因为 'end_turn'，实际为 '%s'", *result.StopReason)
	}

	// 验证使用统计
	if result.Usage == nil {
		t.Fatal("期望使用统计已设置")
	}

	if result.Usage.InputTokens != 10 {
		t.Errorf("期望输入 token 数为 10，实际为 %d", result.Usage.InputTokens)
	}

	if result.Usage.OutputTokens != 20 {
		t.Errorf("期望输出 token 数为 20，实际为 %d", result.Usage.OutputTokens)
	}
}

// TestConvertResponse_EmptyContent 测试空内容的转换
func TestConvertResponse_EmptyContent(t *testing.T) {
	// 构造核心响应（无内容）
	finishReason := "stop"
	coreResp := &coreTypes.Response{
		ID:    "msg_02XyZaBcDeFgHiJkLmNo",
		Model: "claude-3-opus-20240229",
		Choices: []coreTypes.Choice{
			{
				Message: &coreTypes.ResponseMessage{
					Role:    "assistant",
					Content: nil,
				},
				FinishReason: &finishReason,
			},
		},
	}

	// 调用转换函数
	result := converter.ConvertResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	// 验证内容为空数组
	if len(result.Content) != 0 {
		t.Errorf("期望内容为空数组，实际有 %d 个元素", len(result.Content))
	}
}

// TestConvertResponse_WithToolCalls 测试包含工具调用的转换
func TestConvertResponse_WithToolCalls(t *testing.T) {
	// 构造核心响应（包含工具调用）
	finishReason := "tool_calls"
	coreResp := &coreTypes.Response{
		ID:    "msg_03ToolCallTest",
		Model: "claude-3-5-sonnet-20241022",
		Choices: []coreTypes.Choice{
			{
				Message: &coreTypes.ResponseMessage{
					Role: "assistant",
					ToolCalls: []coreTypes.ToolCall{
						{
							ID:   "toolu_01A2B3C4D5E6F7G8H9I0",
							Type: "function",
							Function: coreTypes.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "北京", "unit": "celsius"}`,
							},
						},
					},
				},
				FinishReason: &finishReason,
			},
		},
		Usage: &coreTypes.ResponseUsage{
			PromptTokens:     15,
			CompletionTokens: 25,
			TotalTokens:      40,
		},
	}

	// 调用转换函数
	result := converter.ConvertResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	// 验证内容包含工具调用
	if len(result.Content) != 1 {
		t.Fatalf("期望有 1 个内容块，实际为 %d", len(result.Content))
	}

	toolContent := result.Content[0]
	if toolContent.Type != "tool_use" {
		t.Errorf("期望内容类型为 'tool_use'，实际为 '%s'", toolContent.Type)
	}

	if toolContent.ID == nil {
		t.Fatal("期望工具调用 ID 已设置")
	}

	if *toolContent.ID != "toolu_01A2B3C4D5E6F7G8H9I0" {
		t.Errorf("期望工具调用ID为 'toolu_01A2B3C4D5E6F7G8H9I0'，实际为 '%s'", *toolContent.ID)
	}

	if toolContent.Name == nil {
		t.Fatal("期望工具名称已设置")
	}

	if *toolContent.Name != "get_weather" {
		t.Errorf("期望工具名称为 'get_weather'，实际为 '%s'", *toolContent.Name)
	}

	// 验证工具输入参数
	if toolContent.Input == nil {
		t.Fatal("期望工具输入参数已设置")
	}

	inputMap, ok := toolContent.Input.(map[string]interface{})
	if !ok {
		t.Fatal("期望工具输入参数为 map[string]interface{}")
	}

	if inputMap["location"] != "北京" {
		t.Errorf("期望 location 为 '北京'，实际为 '%v'", inputMap["location"])
	}

	if inputMap["unit"] != "celsius" {
		t.Errorf("期望 unit 为 'celsius'，实际为 '%v'", inputMap["unit"])
	}

	// 验证停止原因
	if result.StopReason == nil {
		t.Fatal("期望停止原因已设置")
	}

	if *result.StopReason != "tool_use" {
		t.Errorf("期望停止原因为 'tool_use'，实际为 '%s'", *result.StopReason)
	}
}

// TestConvertResponse_TextAndToolCalls 测试同时包含文本和工具调用的转换
func TestConvertResponse_TextAndToolCalls(t *testing.T) {
	// 构造核心响应（包含文本和工具调用）
	content := "让我帮你查询北京的天气。"
	finishReason := "tool_calls"
	coreResp := &coreTypes.Response{
		ID:    "msg_04MixedContent",
		Model: "claude-3-5-sonnet-20241022",
		Choices: []coreTypes.Choice{
			{
				Message: &coreTypes.ResponseMessage{
					Role:    "assistant",
					Content: &content,
					ToolCalls: []coreTypes.ToolCall{
						{
							ID:   "toolu_01WeatherQuery",
							Type: "function",
							Function: coreTypes.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "北京"}`,
							},
						},
					},
				},
				FinishReason: &finishReason,
			},
		},
	}

	// 调用转换函数
	result := converter.ConvertResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	// 验证内容包含文本和工具调用
	if len(result.Content) != 2 {
		t.Fatalf("期望有 2 个内容块，实际为 %d", len(result.Content))
	}

	// 验证第一个内容块（文本）
	if result.Content[0].Type != "text" {
		t.Errorf("期望第一个内容块类型为 'text'，实际为 '%s'", result.Content[0].Type)
	}

	if result.Content[0].Text == nil || *result.Content[0].Text != "让我帮你查询北京的天气。" {
		t.Errorf("期望文本内容为 '让我帮你查询北京的天气。'")
	}

	// 验证第二个内容块（工具调用）
	if result.Content[1].Type != "tool_use" {
		t.Errorf("期望第二个内容块类型为 'tool_use'，实际为 '%s'", result.Content[1].Type)
	}

	if result.Content[1].Name == nil || *result.Content[1].Name != "get_weather" {
		t.Errorf("期望工具名称为 'get_weather'")
	}
}

// TestConvertResponse_MultipleToolCalls 测试多个工具调用的转换
func TestConvertResponse_MultipleToolCalls(t *testing.T) {
	// 构造核心响应（包含多个工具调用）
	finishReason := "tool_calls"
	coreResp := &coreTypes.Response{
		ID:    "msg_05MultiTools",
		Model: "claude-3-5-sonnet-20241022",
		Choices: []coreTypes.Choice{
			{
				Message: &coreTypes.ResponseMessage{
					Role: "assistant",
					ToolCalls: []coreTypes.ToolCall{
						{
							ID:   "toolu_01First",
							Type: "function",
							Function: coreTypes.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "北京"}`,
							},
						},
						{
							ID:   "toolu_02Second",
							Type: "function",
							Function: coreTypes.FunctionCall{
								Name:      "get_time",
								Arguments: `{"timezone": "Asia/Shanghai"}`,
							},
						},
					},
				},
				FinishReason: &finishReason,
			},
		},
	}

	// 调用转换函数
	result := converter.ConvertResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	// 验证内容包含两个工具调用
	if len(result.Content) != 2 {
		t.Fatalf("期望有 2 个内容块，实际为 %d", len(result.Content))
	}

	// 验证第一个工具调用
	if result.Content[0].Type != "tool_use" {
		t.Errorf("期望第一个内容块类型为 'tool_use'，实际为 '%s'", result.Content[0].Type)
	}

	if *result.Content[0].Name != "get_weather" {
		t.Errorf("期望第一个工具名称为 'get_weather'，实际为 '%s'", *result.Content[0].Name)
	}

	// 验证第二个工具调用
	if result.Content[1].Type != "tool_use" {
		t.Errorf("期望第二个内容块类型为 'tool_use'，实际为 '%s'", result.Content[1].Type)
	}

	if *result.Content[1].Name != "get_time" {
		t.Errorf("期望第二个工具名称为 'get_time'，实际为 '%s'", *result.Content[1].Name)
	}
}

// TestConvertResponse_NoMessage 测试没有消息的转换
func TestConvertResponse_NoMessage(t *testing.T) {
	// 构造核心响应（没有消息）
	coreResp := &coreTypes.Response{
		ID:    "msg_06NoMessage",
		Model: "claude-3-opus-20240229",
		Choices: []coreTypes.Choice{
			{
				Message: nil,
			},
		},
	}

	// 调用转换函数
	result := converter.ConvertResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	// 验证默认角色
	if result.Role != "assistant" {
		t.Errorf("期望默认角色为 'assistant'，实际为 '%s'", result.Role)
	}

	// 验证内容为空数组
	if len(result.Content) != 0 {
		t.Errorf("期望内容为空数组，实际有 %d 个元素", len(result.Content))
	}
}

// TestConvertResponse_NoChoices 测试没有选择项的转换
func TestConvertResponse_NoChoices(t *testing.T) {
	// 构造核心响应（没有选择项）
	coreResp := &coreTypes.Response{
		ID:      "msg_07NoChoices",
		Model:   "claude-3-opus-20240229",
		Choices: []coreTypes.Choice{},
	}

	// 调用转换函数
	result := converter.ConvertResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	// 验证基本字段仍然设置
	if result.ID != "msg_07NoChoices" {
		t.Errorf("期望ID为 'msg_07NoChoices'，实际为 '%s'", result.ID)
	}

	if result.Type != "message" {
		t.Errorf("期望Type为 'message'，实际为 '%s'", result.Type)
	}
}

// TestConvertResponse_FinishReasonMapping 测试完成原因映射
func TestConvertResponse_FinishReasonMapping(t *testing.T) {
	testCases := []struct {
		name               string
		coreFinishReason   string
		expectedStopReason string
	}{
		{
			name:               "stop to end_turn",
			coreFinishReason:   "stop",
			expectedStopReason: "end_turn",
		},
		{
			name:               "length to max_tokens",
			coreFinishReason:   "length",
			expectedStopReason: "max_tokens",
		},
		{
			name:               "tool_calls to tool_use",
			coreFinishReason:   "tool_calls",
			expectedStopReason: "tool_use",
		},
		{
			name:               "content_filter to end_turn",
			coreFinishReason:   "content_filter",
			expectedStopReason: "end_turn",
		},
		{
			name:               "unknown to end_turn",
			coreFinishReason:   "unknown_reason",
			expectedStopReason: "end_turn",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content := "测试内容"
			coreResp := &coreTypes.Response{
				ID:    "msg_08FinishReason",
				Model: "claude-3-5-sonnet-20241022",
				Choices: []coreTypes.Choice{
					{
						Message: &coreTypes.ResponseMessage{
							Role:    "assistant",
							Content: &content,
						},
						FinishReason: &tc.coreFinishReason,
					},
				},
			}

			result := converter.ConvertResponse(coreResp)

			if result.StopReason == nil {
				t.Fatal("期望停止原因已设置")
			}

			if *result.StopReason != tc.expectedStopReason {
				t.Errorf("期望停止原因为 '%s'，实际为 '%s'", tc.expectedStopReason, *result.StopReason)
			}
		})
	}
}

// TestConvertResponse_NoUsage 测试没有使用统计的转换
func TestConvertResponse_NoUsage(t *testing.T) {
	// 构造核心响应（没有使用统计）
	content := "测试内容"
	finishReason := "stop"
	coreResp := &coreTypes.Response{
		ID:    "msg_09NoUsage",
		Model: "claude-3-opus-20240229",
		Choices: []coreTypes.Choice{
			{
				Message: &coreTypes.ResponseMessage{
					Role:    "assistant",
					Content: &content,
				},
				FinishReason: &finishReason,
			},
		},
		Usage: nil,
	}

	// 调用转换函数
	result := converter.ConvertResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	// 验证使用统计为空
	if result.Usage != nil {
		t.Error("期望使用统计为空")
	}
}

// TestConvertResponse_InvalidToolCallArguments 测试无效的工具调用参数
func TestConvertResponse_InvalidToolCallArguments(t *testing.T) {
	// 构造核心响应（包含无效的 JSON 参数）
	finishReason := "tool_calls"
	coreResp := &coreTypes.Response{
		ID:    "msg_10InvalidArgs",
		Model: "claude-3-5-sonnet-20241022",
		Choices: []coreTypes.Choice{
			{
				Message: &coreTypes.ResponseMessage{
					Role: "assistant",
					ToolCalls: []coreTypes.ToolCall{
						{
							ID:   "toolu_01InvalidJSON",
							Type: "function",
							Function: coreTypes.FunctionCall{
								Name:      "test_function",
								Arguments: `{invalid json}`,
							},
						},
					},
				},
				FinishReason: &finishReason,
			},
		},
	}

	// 调用转换函数（不应崩溃）
	result := converter.ConvertResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	// 验证工具调用仍然存在
	if len(result.Content) != 1 {
		t.Fatalf("期望有 1 个内容块，实际为 %d", len(result.Content))
	}

	// 验证输入为 nil（因为 JSON 解析失败）
	if result.Content[0].Input != nil {
		t.Error("期望无效的 JSON 参数被处理为 nil")
	}
}

// TestConvertResponse_EmptyToolCallArguments 测试空工具调用参数
func TestConvertResponse_EmptyToolCallArguments(t *testing.T) {
	// 构造核心响应（包含空参数）
	finishReason := "tool_calls"
	coreResp := &coreTypes.Response{
		ID:    "msg_11EmptyArgs",
		Model: "claude-3-5-sonnet-20241022",
		Choices: []coreTypes.Choice{
			{
				Message: &coreTypes.ResponseMessage{
					Role: "assistant",
					ToolCalls: []coreTypes.ToolCall{
						{
							ID:   "toolu_01EmptyArgs",
							Type: "function",
							Function: coreTypes.FunctionCall{
								Name:      "no_params_function",
								Arguments: "",
							},
						},
					},
				},
				FinishReason: &finishReason,
			},
		},
	}

	// 调用转换函数
	result := converter.ConvertResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	// 验证工具调用仍然存在
	if len(result.Content) != 1 {
		t.Fatalf("期望有 1 个内容块，实际为 %d", len(result.Content))
	}

	// 验证输入为 nil（空参数时）
	if result.Content[0].Input != nil {
		t.Error("期望空参数被处理为 nil")
	}
}

// TestConvertResponse_RoundTrip 测试往返转换
func TestConvertResponse_RoundTrip(t *testing.T) {
	// 构造 Anthropic 响应
	text := "这是一个测试响应。"
	stopReason := "end_turn"
	anthropicResp := &types.Response{
		ID:    "msg_12RoundTrip",
		Type:  "message",
		Role:  "assistant",
		Model: "claude-3-5-sonnet-20241022",
		Content: []types.ResponseContent{
			{
				Type: "text",
				Text: &text,
			},
		},
		StopReason: &stopReason,
		Usage: &types.Usage{
			InputTokens:  15,
			OutputTokens: 25,
		},
	}

	// 将 Anthropic 响应转换为核心响应
	coreResp := anthropicResp.ConvertCoreResponse()

	// 再将核心响应转换回 Anthropic 响应
	result := converter.ConvertResponse(coreResp)

	// 验证往返转换的一致性
	if result.ID != anthropicResp.ID {
		t.Errorf("期望ID为 '%s'，实际为 '%s'", anthropicResp.ID, result.ID)
	}

	if result.Model != anthropicResp.Model {
		t.Errorf("期望模型为 '%s'，实际为 '%s'", anthropicResp.Model, result.Model)
	}

	if result.Role != anthropicResp.Role {
		t.Errorf("期望角色为 '%s'，实际为 '%s'", anthropicResp.Role, result.Role)
	}

	if len(result.Content) != len(anthropicResp.Content) {
		t.Fatalf("期望有 %d 个内容块，实际为 %d", len(anthropicResp.Content), len(result.Content))
	}

	if *result.Content[0].Text != *anthropicResp.Content[0].Text {
		t.Errorf("期望文本内容为 '%s'，实际为 '%s'", *anthropicResp.Content[0].Text, *result.Content[0].Text)
	}

	// 注意：stop_reason 会经过映射转换，所以可能不完全一致
	// end_turn -> stop -> end_turn（往返后保持一致）
}

// TestConvertResponse_JSONSerialization 测试 JSON 序列化
func TestConvertResponse_JSONSerialization(t *testing.T) {
	// 构造核心响应
	content := "JSON 序列化测试"
	finishReason := "stop"
	coreResp := &coreTypes.Response{
		ID:    "msg_13JSON",
		Model: "claude-3-5-sonnet-20241022",
		Choices: []coreTypes.Choice{
			{
				Message: &coreTypes.ResponseMessage{
					Role:    "assistant",
					Content: &content,
				},
				FinishReason: &finishReason,
			},
		},
		Usage: &coreTypes.ResponseUsage{
			PromptTokens:     10,
			CompletionTokens: 15,
			TotalTokens:      25,
		},
	}

	// 调用转换函数
	result := converter.ConvertResponse(coreResp)

	// 序列化为 JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	// 反序列化
	var deserialized types.Response
	err = json.Unmarshal(jsonData, &deserialized)
	if err != nil {
		t.Fatalf("JSON 反序列化失败: %v", err)
	}

	// 验证反序列化后的数据
	if deserialized.ID != result.ID {
		t.Errorf("期望ID为 '%s'，实际为 '%s'", result.ID, deserialized.ID)
	}

	if deserialized.Model != result.Model {
		t.Errorf("期望模型为 '%s'，实际为 '%s'", result.Model, deserialized.Model)
	}

	if len(deserialized.Content) != len(result.Content) {
		t.Fatalf("期望有 %d 个内容块，实际为 %d", len(result.Content), len(deserialized.Content))
	}

	if *deserialized.Content[0].Text != *result.Content[0].Text {
		t.Errorf("期望文本内容为 '%s'，实际为 '%s'", *result.Content[0].Text, *deserialized.Content[0].Text)
	}
}
