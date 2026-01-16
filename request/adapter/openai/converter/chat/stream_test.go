package chat_test

import (
	"encoding/json"
	"testing"

	openaiChatConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/chat"
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	coreTypes "github.com/MeowSalty/portal/types"
)

// TestConvertCoreResponse_ExtraFields_Streaming 测试流式响应中 ExtraFields 的传递
func TestConvertCoreResponse_ExtraFields_Streaming(t *testing.T) {
	// 构造包含 ExtraFields 的 OpenAI 流式响应 JSON
	openaiRespJSON := `{
		"id": "chatcmpl-extra-test",
		"object": "chat.completion.chunk",
		"created": 1700000003,
		"model": "gpt-4",
		"choices": [
			{
				"index": 0,
				"finish_reason": "stop",
				"delta": {
					"role": "assistant",
					"content": "Streaming with extras",
					"stream_extra1": "stream_value1",
					"stream_extra2": 555
				}
			}
		]
	}`

	// 解析 OpenAI 响应
	var openaiResp openaiChat.ResponseChunk
	err := json.Unmarshal([]byte(openaiRespJSON), &openaiResp)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := openaiChatConverter.ConvertCoreStreamResponse(&openaiResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证流式响应的 Delta 字段
	if choice.Delta == nil {
		t.Fatal("期望 Delta 字段已设置")
	}

	// 验证 ExtraFields 被正确传递
	if len(choice.Delta.ExtraFields) != 2 {
		t.Fatalf("期望 Delta 有 2 个 ExtraFields，实际为 %d", len(choice.Delta.ExtraFields))
	}

	// 验证 ExtraFields 的值
	if choice.Delta.ExtraFields["stream_extra1"] != "stream_value1" {
		t.Errorf("期望 stream_extra1 为 'stream_value1'，实际为 '%v'", choice.Delta.ExtraFields["stream_extra1"])
	}

	if choice.Delta.ExtraFields["stream_extra2"] != float64(555) {
		t.Errorf("期望 stream_extra2 为 555，实际为 '%v'", choice.Delta.ExtraFields["stream_extra2"])
	}

	// 验证来源格式
	if choice.Delta.ExtraFieldsFormat != "openai" {
		t.Errorf("期望来源格式为 'openai'，实际为 '%s'", choice.Delta.ExtraFieldsFormat)
	}
}

// TestConvertResponse_ExtraFields_Streaming 测试将核心响应转换为 OpenAI 响应时 ExtraFields 的传递
func TestConvertResponse_ExtraFields_Streaming(t *testing.T) {
	// 构造包含 ExtraFields 的核心流式响应
	coreResp := &coreTypes.Response{
		ID:      "core-extra-test",
		Model:   "gpt-4",
		Object:  "chat.completion.chunk",
		Created: 1700000005,
		Choices: []coreTypes.Choice{
			{
				FinishReason: stringPtr("stop"),
				Delta: &coreTypes.Delta{
					Role:    stringPtr("assistant"),
					Content: stringPtr("Core streaming with extras"),
					ExtraFields: map[string]interface{}{
						"core_stream_extra1": "core_stream_value1",
						"core_stream_extra2": 777,
						"core_stream_extra3": 3.14,
					},
				},
			},
		},
	}

	// 调用转换函数
	result := openaiChatConverter.ConvertStreamResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// 验证流式响应的 Delta 字段
	if choice.Delta == nil {
		t.Fatal("期望 Delta 字段已设置")
	}

	// 验证 ExtraFields 被正确传递
	if len(choice.Delta.ExtraFields) != 3 {
		t.Fatalf("期望 Delta 有 3 个 ExtraFields，实际为 %d", len(choice.Delta.ExtraFields))
	}

	// 验证 ExtraFields 的值
	if choice.Delta.ExtraFields["core_stream_extra1"] != "core_stream_value1" {
		t.Errorf("期望 core_stream_extra1 为 'core_stream_value1'，实际为 '%v'", choice.Delta.ExtraFields["core_stream_extra1"])
	}

	if choice.Delta.ExtraFields["core_stream_extra2"] != int(777) {
		t.Errorf("期望 core_stream_extra2 为 777，实际为 '%v'", choice.Delta.ExtraFields["core_stream_extra2"])
	}

	if choice.Delta.ExtraFields["core_stream_extra3"] != float64(3.14) {
		t.Errorf("期望 core_stream_extra3 为 3.14，实际为 '%v'", choice.Delta.ExtraFields["core_stream_extra3"])
	}
}

// TestConvertCoreResponse_ExtraFields_Empty 测试没有 ExtraFields 时的转换
func TestConvertCoreResponse_ExtraFields_Empty(t *testing.T) {
	// 构造不包含 ExtraFields 的 OpenAI 响应 JSON
	openaiRespJSON := `{
		"id": "chatcmpl-no-extra",
		"object": "chat.completion.chunk",
		"created": 1700000007,
		"model": "gpt-4",
		"choices": [
			{
				"index": 0,
				"finish_reason": "stop",
				"delta": {
					"role": "assistant",
					"content": "No extra fields"
				}
			}
		]
	}`

	// 解析 OpenAI 响应
	var openaiResp openaiChat.ResponseChunk
	err := json.Unmarshal([]byte(openaiRespJSON), &openaiResp)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := openaiChatConverter.ConvertCoreStreamResponse(&openaiResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	choice := result.Choices[0]

	// 验证没有 ExtraFields
	if len(choice.Delta.ExtraFields) > 0 {
		t.Errorf("期望没有 ExtraFields，实际有 %d 个", len(choice.Delta.ExtraFields))
	}
}
