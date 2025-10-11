package converter_test

import (
	"encoding/json"
	"testing"

	"github.com/MeowSalty/portal/request/adapter/openai/converter"
	"github.com/MeowSalty/portal/request/adapter/openai/types"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
)

// TestConvertRequest_TextMessage 测试文本消息请求转换
func TestConvertRequest_TextMessage(t *testing.T) {
	// 构造输入的 JSON
	coreReqJSON := `{
		"model": "gpt-3.5-turbo",
		"messages": [
			{
				"role": "user",
				"content": "Hello, how are you?"
			},
			{
				"role": "assistant",
				"content": "I'm fine, thank you! How can I help you today?"
			}
		],
		"temperature": 0.7,
		"max_tokens": 150
	}`

	// 解析核心请求
	var coreRequest coreTypes.Request
	err := json.Unmarshal([]byte(coreReqJSON), &coreRequest)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 创建通道信息
	channel := &routing.Channel{
		ModelName: "gpt-3.5-turbo",
	}

	// 调用转换函数
	result := converter.ConvertRequest(&coreRequest, channel)

	// 将结果转换为 OpenAI 请求类型
	openaiReq, ok := result.(*types.Request)
	if !ok {
		t.Fatalf("期望结果类型为*types.Request，实际为 %T", result)
	}

	// 验证模型名称
	if openaiReq.Model != "gpt-3.5-turbo" {
		t.Errorf("期望模型为 'gpt-3.5-turbo'，实际为 '%s'", openaiReq.Model)
	}

	// 验证消息数量
	if len(openaiReq.Messages) != 2 {
		t.Fatalf("期望有 2 条消息，实际为 %d", len(openaiReq.Messages))
	}

	// 验证第一条消息
	userMsg := openaiReq.Messages[0]
	if userMsg.Role != "user" {
		t.Errorf("期望第一条消息角色为 'user'，实际为 '%s'", userMsg.Role)
	}

	content, ok := userMsg.Content.(string)
	if !ok {
		t.Errorf("期望第一条消息内容为字符串类型，实际为 %T", userMsg.Content)
	}
	if content != "Hello, how are you?" {
		t.Errorf("期望第一条消息内容为 'Hello, how are you?'，实际为 '%s'", content)
	}

	// 验证第二条消息
	assistantMsg := openaiReq.Messages[1]
	if assistantMsg.Role != "assistant" {
		t.Errorf("期望第二条消息角色为 'assistant'，实际为 '%s'", assistantMsg.Role)
	}

	assistantContent, ok := assistantMsg.Content.(string)
	if !ok {
		t.Errorf("期望第二条消息内容为字符串类型，实际为 %T", assistantMsg.Content)
	}
	if assistantContent != "I'm fine, thank you! How can I help you today?" {
		t.Errorf("期望第二条消息内容为 'I'm fine, thank you! How can I help you today?'，实际为 '%s'", assistantContent)
	}

	// 验证温度参数
	if openaiReq.Temperature == nil {
		t.Error("期望温度参数已设置")
	} else if *openaiReq.Temperature != 0.7 {
		t.Errorf("期望温度参数为 0.7，实际为 %f", *openaiReq.Temperature)
	}

	// 验证最大 token 数
	if openaiReq.MaxTokens == nil {
		t.Error("期望最大 token 数已设置")
	} else if *openaiReq.MaxTokens != 150 {
		t.Errorf("期望最大 token 数为 150，实际为 %d", *openaiReq.MaxTokens)
	}
}

// TestConvertRequest_TextMessageWithStream 测试带流式传输的文本消息请求转换
func TestConvertRequest_TextMessageWithStream(t *testing.T) {
	// 构造输入的 JSON
	coreReqJSON := `{
		"model": "gpt-4",
		"messages": [
			{
				"role": "system",
				"content": "You are a helpful assistant."
			},
			{
				"role": "user",
				"content": "What's the weather like today?"
			}
		],
		"stream": true
	}`

	// 解析核心请求
	var coreRequest coreTypes.Request
	err := json.Unmarshal([]byte(coreReqJSON), &coreRequest)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 创建通道信息
	channel := &routing.Channel{
		ModelName: "gpt-4",
	}

	// 调用转换函数
	result := converter.ConvertRequest(&coreRequest, channel)

	// 将结果转换为 OpenAI 请求类型
	openaiReq, ok := result.(*types.Request)
	if !ok {
		t.Fatalf("期望结果类型为*types.Request，实际为 %T", result)
	}

	// 验证模型名称
	if openaiReq.Model != "gpt-4" {
		t.Errorf("期望模型为 'gpt-4'，实际为 '%s'", openaiReq.Model)
	}

	// 验证流参数
	if openaiReq.Stream == nil {
		t.Error("期望流参数已设置")
	} else if !*openaiReq.Stream {
		t.Error("期望流参数为 true")
	}

	// 验证消息数量
	if len(openaiReq.Messages) != 2 {
		t.Fatalf("期望有 2 条消息，实际为 %d", len(openaiReq.Messages))
	}

	// 验证系统消息
	systemMsg := openaiReq.Messages[0]
	if systemMsg.Role != "system" {
		t.Errorf("期望第一条消息角色为 'system'，实际为 '%s'", systemMsg.Role)
	}

	systemContent, ok := systemMsg.Content.(string)
	if !ok {
		t.Errorf("期望系统消息内容为字符串类型，实际为 %T", systemMsg.Content)
	}
	if systemContent != "You are a helpful assistant." {
		t.Errorf("期望系统消息内容为 'You are a helpful assistant.'，实际为 '%s'", systemContent)
	}

	// 验证用户消息
	userMsg := openaiReq.Messages[1]
	if userMsg.Role != "user" {
		t.Errorf("期望第二条消息角色为 'user'，实际为 '%s'", userMsg.Role)
	}

	userContent, ok := userMsg.Content.(string)
	if !ok {
		t.Errorf("期望用户消息内容为字符串类型，实际为 %T", userMsg.Content)
	}
	if userContent != "What's the weather like today?" {
		t.Errorf("期望用户消息内容为 'What's the weather like today?'，实际为 '%s'", userContent)
	}
}

// TestConvertCoreRequest_TextMessage 测试核心文本消息请求转换
func TestConvertCoreRequest_TextMessage(t *testing.T) {
	// 构造输入的 OpenAI 请求 JSON
	openaiReqJSON := `{
		"model": "gpt-3.5-turbo",
		"messages": [
			{
				"role": "user",
				"content": "Hello, how are you?"
			},
			{
				"role": "assistant", 
				"content": "I'm fine, thank you! How can I help you today?"
			}
		],
		"temperature": 0.7,
		"max_tokens": 150
	}`

	// 解析 OpenAI 请求
	var openaiRequest types.Request
	err := json.Unmarshal([]byte(openaiReqJSON), &openaiRequest)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := converter.ConvertCoreRequest(&openaiRequest)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreRequest 返回非空结果")
	}

	// 验证模型名称
	if result.Model != "gpt-3.5-turbo" {
		t.Errorf("期望模型为 'gpt-3.5-turbo'，实际为 '%s'", result.Model)
	}

	// 验证消息数量
	if len(result.Messages) != 2 {
		t.Fatalf("期望有 2 条消息，实际为 %d", len(result.Messages))
	}

	// 验证第一条消息
	userMsg := result.Messages[0]
	if userMsg.Role != "user" {
		t.Errorf("期望第一条消息角色为 'user'，实际为 '%s'", userMsg.Role)
	}
	if userMsg.Content.StringValue == nil {
		t.Error("期望第一条消息内容已设置")
	} else if *userMsg.Content.StringValue != "Hello, how are you?" {
		t.Errorf("期望第一条消息内容为 'Hello, how are you?'，实际为 '%s'", *userMsg.Content.StringValue)
	}

	// 验证第二条消息
	assistantMsg := result.Messages[1]
	if assistantMsg.Role != "assistant" {
		t.Errorf("期望第二条消息角色为 'assistant'，实际为 '%s'", assistantMsg.Role)
	}
	if assistantMsg.Content.StringValue == nil {
		t.Error("期望第二条消息内容已设置")
	} else if *assistantMsg.Content.StringValue != "I'm fine, thank you! How can I help you today?" {
		t.Errorf("期望第二条消息内容为 'I'm fine, thank you! How can I help you today?'，实际为 '%s'", *assistantMsg.Content.StringValue)
	}

	// 验证温度参数
	if result.Temperature == nil {
		t.Error("期望温度参数已设置")
	} else if *result.Temperature != 0.7 {
		t.Errorf("期望温度参数为 0.7，实际为 %f", *result.Temperature)
	}

	// 验证最大 token 数
	if result.MaxTokens == nil {
		t.Error("期望最大 token 数已设置")
	} else if *result.MaxTokens != 150 {
		t.Errorf("期望最大 token 数为 150，实际为 %d", *result.MaxTokens)
	}
}

// TestConvertCoreRequest_TextMessageWithStream 测试带流式传输的核心文本消息请求转换
func TestConvertCoreRequest_TextMessageWithStream(t *testing.T) {
	// 构造输入的 OpenAI 请求 JSON
	openaiReqJSON := `{
		"model": "gpt-4",
		"messages": [
			{
				"role": "system",
				"content": "You are a helpful assistant."
			},
			{
				"role": "user",
				"content": "What's the weather like today?"
			}
		],
		"stream": true
	}`

	// 解析 OpenAI 请求
	var openaiRequest types.Request
	err := json.Unmarshal([]byte(openaiReqJSON), &openaiRequest)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := converter.ConvertCoreRequest(&openaiRequest)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreRequest 返回非空结果")
	}

	// 验证模型名称
	if result.Model != "gpt-4" {
		t.Errorf("期望模型为 'gpt-4'，实际为 '%s'", result.Model)
	}

	// 验证流参数
	if result.Stream == nil {
		t.Error("期望流参数已设置")
	} else if !*result.Stream {
		t.Error("期望流参数为 true")
	}

	// 验证消息数量
	if len(result.Messages) != 2 {
		t.Fatalf("期望有 2 条消息，实际为 %d", len(result.Messages))
	}

	// 验证系统消息
	systemMsg := result.Messages[0]
	if systemMsg.Role != "system" {
		t.Errorf("期望第一条消息角色为 'system'，实际为 '%s'", systemMsg.Role)
	}
	if systemMsg.Content.StringValue == nil {
		t.Error("期望系统消息内容已设置")
	} else if *systemMsg.Content.StringValue != "You are a helpful assistant." {
		t.Errorf("期望系统消息内容为 'You are a helpful assistant.'，实际为 '%s'", *systemMsg.Content.StringValue)
	}

	// 验证用户消息
	userMsg := result.Messages[1]
	if userMsg.Role != "user" {
		t.Errorf("期望第二条消息角色为 'user'，实际为 '%s'", userMsg.Role)
	}
	if userMsg.Content.StringValue == nil {
		t.Error("期望用户消息内容已设置")
	} else if *userMsg.Content.StringValue != "What's the weather like today?" {
		t.Errorf("期望用户消息内容为 'What's the weather like today?'，实际为 '%s'", *userMsg.Content.StringValue)
	}
}

// TestConvertCoreRequest_WithAdditionalParameters 测试带附加参数的核心请求转换
func TestConvertCoreRequest_WithAdditionalParameters(t *testing.T) {
	// 构造输入的 OpenAI 请求 JSON，包含额外参数
	openaiReqJSON := `{
		"model": "gpt-3.5-turbo",
		"messages": [
			{
				"role": "user",
				"content": "Test message"
			}
		],
		"temperature": 0.8,
		"top_p": 0.9,
		"max_tokens": 100,
		"frequency_penalty": 0.5,
		"presence_penalty": 0.3,
		"seed": 42,
		"user": "test-user"
	}`

	// 解析 OpenAI 请求
	var openaiRequest types.Request
	err := json.Unmarshal([]byte(openaiReqJSON), &openaiRequest)
	if err != nil {
		t.Fatalf("解析输入JSON失败: %v", err)
	}

	// 调用转换函数
	result := converter.ConvertCoreRequest(&openaiRequest)

	// 验证结果不为空
	if result == nil {
		t.Fatal("期望 ConvertCoreRequest 返回非空结果")
	}

	// 验证温度参数
	if result.Temperature == nil {
		t.Error("期望温度参数已设置")
	} else if *result.Temperature != 0.8 {
		t.Errorf("期望温度参数为 0.8，实际为 %f", *result.Temperature)
	}

	// 验证 TopP 参数
	if result.TopP == nil {
		t.Error("期望 top_p 参数已设置")
	} else if *result.TopP != 0.9 {
		t.Errorf("期望 top_p 参数为 0.9，实际为 %f", *result.TopP)
	}

	// 验证最大 token 数
	if result.MaxTokens == nil {
		t.Error("期望最大 token 数已设置")
	} else if *result.MaxTokens != 100 {
		t.Errorf("期望最大 token 数为 100，实际为 %d", *result.MaxTokens)
	}

	// 验证频率惩罚
	if result.FrequencyPenalty == nil {
		t.Error("期望频率惩罚参数已设置")
	} else if *result.FrequencyPenalty != 0.5 {
		t.Errorf("期望频率惩罚参数为 0.5，实际为 %f", *result.FrequencyPenalty)
	}

	// 验证存在惩罚
	if result.PresencePenalty == nil {
		t.Error("期望存在惩罚参数已设置")
	} else if *result.PresencePenalty != 0.3 {
		t.Errorf("期望存在惩罚参数为 0.3，实际为 %f", *result.PresencePenalty)
	}

	// 验证种子
	if result.Seed == nil {
		t.Error("期望种子参数已设置")
	} else if *result.Seed != 42 {
		t.Errorf("期望种子参数为 42，实际为 %d", *result.Seed)
	}

	// 验证用户
	if result.User == nil {
		t.Error("期望用户参数已设置")
	} else if *result.User != "test-user" {
		t.Errorf("期望用户参数为 'test-user'，实际为 '%s'", *result.User)
	}
}
