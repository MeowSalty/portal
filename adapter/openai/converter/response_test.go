package converter_test

import (
	"encoding/json"
	"testing"

	"github.com/MeowSalty/portal/adapter/openai/converter"
	"github.com/MeowSalty/portal/adapter/openai/types"
	// coreTypes "github.com/MeowSalty/portal/types"
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
				"delta": {
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
	var openaiResp types.Response
	err := json.Unmarshal([]byte(openaiRespJSON), &openaiResp)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := converter.ConvertCoreResponse(&openaiResp)

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
				"delta": {
					"role": "assistant",
					"content": "This is a test response."
				}
			}
		]
	}`

	// 解析 OpenAI 响应
	var openaiResp types.Response
	err := json.Unmarshal([]byte(openaiRespJSON), &openaiResp)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := converter.ConvertCoreResponse(&openaiResp)

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
				"delta": {
					"role": "assistant"
				}
			}
		]
	}`

	// 解析 OpenAI 响应
	var openaiResp types.Response
	err := json.Unmarshal([]byte(openaiRespJSON), &openaiResp)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := converter.ConvertCoreResponse(&openaiResp)

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
	result := converter.ConvertCoreResponse(nil)

	// 验证返回结果为 nil
	if result != nil {
		t.Errorf("期望返回 nil，实际为 %v", result)
	}
}
