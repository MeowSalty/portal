package chat_test

import (
	"encoding/json"
	"testing"

	openaiChatConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/chat"
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	coreTypes "github.com/MeowSalty/portal/types"
)

// TestConvertCoreResponse_TextMessage 测试文本消息的 OpenAI 响应转换
func TestConvertCoreResponse_TextMessage(t *testing.T) {
	// 构造输入的 OpenAI 响应 JSON
	openaiRespJSON := `{
		"id": "chatcmpl-1234567890",
		"object": "chat.completion",
		"created": 1700000000,
		"model": "gpt-3.5-turbo",
		"choices": [
			{
				"index": 0,
				"finish_reason": "stop",
				"message": {
					"role": "assistant",
					"content": "Hello! How can I help you today?"
				}
			}
		],
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 20,
			"total_tokens": 30
		}
	}`

	// 解析 OpenAI 响应
	var openaiResp openaiChat.Response
	err := json.Unmarshal([]byte(openaiRespJSON), &openaiResp)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := openaiChatConverter.ConvertCoreResponse(&openaiResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证基本字段
	if result.ID != "chatcmpl-1234567890" {
		t.Errorf("期望ID为 'chatcmpl-1234567890'，实际为 '%s'", result.ID)
	}

	if result.Model != "gpt-3.5-turbo" {
		t.Errorf("期望模型为 'gpt-3.5-turbo'，实际为 '%s'", result.Model)
	}

	if result.Object != "chat.completion" {
		t.Errorf("期望对象类型为 'chat.completion'，实际为 '%s'", result.Object)
	}

	if result.Created != 1700000000 {
		t.Errorf("期望创建时间为 1700000000，实际为 %d", result.Created)
	}

	// 验证使用情况
	if result.Usage == nil {
		t.Fatal("期望使用情况数据已设置")
	}

	if result.Usage.PromptTokens != 10 {
		t.Errorf("期望提示 token 数为 10，实际为 %d", result.Usage.PromptTokens)
	}

	if result.Usage.CompletionTokens != 20 {
		t.Errorf("期望完成 token 数为 20，实际为 %d", result.Usage.CompletionTokens)
	}

	if result.Usage.TotalTokens != 30 {
		t.Errorf("期望总 token 数为 30，实际为 %d", result.Usage.TotalTokens)
	}

	// 验证选择项
	if len(result.Choices) != 1 {
		t.Fatalf("期望有 1 个选择项，实际为 %d", len(result.Choices))
	}

	choice := result.Choices[0]
	if choice.FinishReason == nil {
		t.Error("期望完成原因已设置")
	} else if *choice.FinishReason != "stop" {
		t.Errorf("期望完成原因为 'stop'，实际为 '%s'", *choice.FinishReason)
	}

	// 验证消息内容
	if choice.Message == nil {
		t.Fatal("期望消息已设置")
	}

	if choice.Message.Role != "assistant" {
		t.Errorf("期望消息角色为 'assistant'，实际为 '%s'", choice.Message.Role)
	}

	if choice.Message.Content == nil {
		t.Fatal("期望消息内容已设置")
	} else if *choice.Message.Content != "Hello! How can I help you today?" {
		t.Errorf("期望消息内容为 'Hello! How can I help you today?'，实际为 '%s'", *choice.Message.Content)
	}
}

// TestConvertCoreResponse_TextMessageWithoutUsage 测试没有使用情况的文本消息转换
func TestConvertCoreResponse_TextMessageWithoutUsage(t *testing.T) {
	// 构造输入的 OpenAI 响应 JSON（没有使用情况数据）
	openaiRespJSON := `{
		"id": "chatcmpl-9876543210",
		"object": "chat.completion",
		"created": 1700000001,
		"model": "gpt-4",
		"choices": [
			{
				"index": 0,
				"finish_reason": "length",
				"message": {
					"role": "assistant",
					"content": "This is a test response."
				}
			}
		]
	}`

	// 解析 OpenAI 响应
	var openaiResp openaiChat.Response
	err := json.Unmarshal([]byte(openaiRespJSON), &openaiResp)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := openaiChatConverter.ConvertCoreResponse(&openaiResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证使用情况为空
	if result.Usage != nil {
		t.Error("期望使用情况数据为空")
	}

	// 验证选择项
	choice := result.Choices[0]
	if choice.FinishReason == nil {
		t.Error("期望完成原因已设置")
	} else if *choice.FinishReason != "length" {
		t.Errorf("期望完成原因为 'length'，实际为 '%s'", *choice.FinishReason)
	}

	// 验证消息内容
	if choice.Message.Content == nil {
		t.Fatal("期望消息内容已设置")
	} else if *choice.Message.Content != "This is a test response." {
		t.Errorf("期望消息内容为 'This is a test response.'，实际为 '%s'", *choice.Message.Content)
	}
}

// TestConvertCoreResponse_EmptyContent 测试空内容消息的转换
func TestConvertCoreResponse_EmptyContent(t *testing.T) {
	// 构造输入的 OpenAI 响应 JSON（内容为空）
	openaiRespJSON := `{
		"id": "chatcmpl-1111111111",
		"object": "chat.completion",
		"created": 1700000002,
		"model": "gpt-3.5-turbo",
		"choices": [
			{
				"index": 0,
				"finish_reason": "stop",
				"message": {
					"role": "assistant"
				}
			}
		]
	}`

	// 解析 OpenAI 响应
	var openaiResp openaiChat.Response
	err := json.Unmarshal([]byte(openaiRespJSON), &openaiResp)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := openaiChatConverter.ConvertCoreResponse(&openaiResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	// 验证消息内容为空
	choice := result.Choices[0]
	if choice.Message.Content != nil {
		t.Error("期望消息内容为空")
	}
}

// TestConvertCoreResponse_NilInput 测试空输入的处理
func TestConvertCoreResponse_NilInput(t *testing.T) {
	// 调用转换函数，传入 nil
	result := openaiChatConverter.ConvertCoreResponse(nil)

	// 验证返回结果为 nil
	if result != nil {
		t.Errorf("期望返回 nil，实际为 %v", result)
	}
}

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
	var openaiResp openaiChat.Response
	err := json.Unmarshal([]byte(openaiRespJSON), &openaiResp)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := openaiChatConverter.ConvertCoreResponse(&openaiResp)

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

// TestConvertCoreResponse_ExtraFields_NonStreaming 测试非流式响应中 ExtraFields 的传递
func TestConvertCoreResponse_ExtraFields_NonStreaming(t *testing.T) {
	// 构造包含 ExtraFields 的 OpenAI 非流式响应 JSON
	openaiRespJSON := `{
		"id": "chatcmpl-nonstream-extra",
		"object": "chat.completion",
		"created": 1700000004,
		"model": "gpt-3.5-turbo",
		"choices": [
			{
				"index": 0,
				"finish_reason": "stop",
				"message": {
					"role": "assistant",
					"content": "Non-streaming with extras",
					"nonstream_extra1": "nonstream_value1",
					"nonstream_extra2": 666
				}
			}
		]
	}`

	// 解析 OpenAI 响应
	var openaiResp openaiChat.Response
	err := json.Unmarshal([]byte(openaiRespJSON), &openaiResp)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := openaiChatConverter.ConvertCoreResponse(&openaiResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreResponse 返回非空结果")
	}

	choice := result.Choices[0]

	// 验证非流式响应的 Message 字段
	if choice.Message == nil {
		t.Fatal("期望 Message 字段已设置")
	}

	// 验证 ExtraFields 被正确传递
	if len(choice.Message.ExtraFields) != 2 {
		t.Fatalf("期望 Message 有 2 个 ExtraFields，实际为 %d", len(choice.Message.ExtraFields))
	}

	// 验证 ExtraFields 的值
	if choice.Message.ExtraFields["nonstream_extra1"] != "nonstream_value1" {
		t.Errorf("期望 nonstream_extra1 为 'nonstream_value1'，实际为 '%v'", choice.Message.ExtraFields["nonstream_extra1"])
	}

	if choice.Message.ExtraFields["nonstream_extra2"] != float64(666) {
		t.Errorf("期望 nonstream_extra2 为 666，实际为 '%v'", choice.Message.ExtraFields["nonstream_extra2"])
	}

	// 验证来源格式
	if choice.Message.ExtraFieldsFormat != "openai" {
		t.Errorf("期望来源格式为 'openai'，实际为 '%s'", choice.Message.ExtraFieldsFormat)
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
	result := openaiChatConverter.ConvertResponse(coreResp)

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

// TestConvertResponse_ExtraFields_NonStreaming 测试将核心非流式响应转换为 OpenAI 响应时 ExtraFields 的传递
func TestConvertResponse_ExtraFields_NonStreaming(t *testing.T) {
	// 构造包含 ExtraFields 的核心非流式响应
	coreResp := &coreTypes.Response{
		ID:      "core-nonstream-extra",
		Model:   "gpt-3.5-turbo",
		Object:  "chat.completion",
		Created: 1700000006,
		Choices: []coreTypes.Choice{
			{
				FinishReason: stringPtr("stop"),
				Message: &coreTypes.ResponseMessage{
					Role:    "assistant",
					Content: stringPtr("Core non-streaming with extras"),
					ExtraFields: map[string]interface{}{
						"core_nonstream_extra1": "core_nonstream_value1",
						"core_nonstream_extra2": 888,
						"core_nonstream_extra3": 33.3,
					},
				},
			},
		},
	}

	// 调用转换函数
	result := openaiChatConverter.ConvertResponse(coreResp)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertResponse 返回非空结果")
	}

	choice := result.Choices[0]

	// 验证非流式响应的 Message 字段
	if choice.Message == nil {
		t.Fatal("期望 Message 字段已设置")
	}

	// 验证 ExtraFields 被正确传递
	if len(choice.Message.ExtraFields) != 3 {
		t.Fatalf("期望 Message 有 3 个 ExtraFields，实际为 %d", len(choice.Message.ExtraFields))
	}

	// 验证 ExtraFields 的值
	if choice.Message.ExtraFields["core_nonstream_extra1"] != "core_nonstream_value1" {
		t.Errorf("期望 core_nonstream_extra1 为 'core_nonstream_value1'，实际为 '%v'", choice.Message.ExtraFields["core_nonstream_extra1"])
	}

	if choice.Message.ExtraFields["core_nonstream_extra2"] != int(888) {
		t.Errorf("期望 core_nonstream_extra2 为 888，实际为 '%v'", choice.Message.ExtraFields["core_nonstream_extra2"])
	}

	if choice.Message.ExtraFields["core_nonstream_extra3"] != float64(33.3) {
		t.Errorf("期望 core_nonstream_extra3 为 33.3，实际为 '%v'", choice.Message.ExtraFields["core_nonstream_extra2"])
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
	var openaiResp openaiChat.Response
	err := json.Unmarshal([]byte(openaiRespJSON), &openaiResp)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := openaiChatConverter.ConvertCoreResponse(&openaiResp)

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

// 辅助函数：创建字符串指针
func stringPtr(s string) *string {
	return &s
}
