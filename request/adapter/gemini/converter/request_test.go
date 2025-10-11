package converter_test

import (
	"encoding/json"
	"testing"

	"github.com/MeowSalty/portal/request/adapter/gemini/converter"
	"github.com/MeowSalty/portal/request/adapter/gemini/types"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
)

func TestConvertRequest_TextMessage(t *testing.T) {
	// 准备测试数据：一个包含文本消息的 JSON
	coreReqJSON := `{
		"model": "gemini-pro",
		"messages": [
			{
				"role": "user",
				"content": "Hello, how are you?"
			}
		],
		"temperature": 0.8,
		"max_tokens": 100
	}`

	// 解析 JSON 数据到核心请求对象
	var request coreTypes.Request
	err := json.Unmarshal([]byte(coreReqJSON), &request)
	if err != nil {
		t.Fatalf("解析 JSON 失败: %v", err)
	}

	// 创建一个模拟的通道对象
	channel := &routing.Channel{}

	// 执行转换
	geminiReq := converter.ConvertRequest(&request, channel)

	// 验证结果
	if geminiReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// 检查基本结构
	if len(geminiReq.Contents) != 1 {
		t.Errorf("期望 1 个内容，得到 %d", len(geminiReq.Contents))
	}

	// 检查用户消息
	content := geminiReq.Contents[0]
	if content.Role != "user" {
		t.Errorf("期望角色 'user', 得到 '%s'", content.Role)
	}

	// 检查消息部分
	if len(content.Parts) != 1 {
		t.Errorf("期望 1 个部分，得到 %d", len(content.Parts))
	}

	// 检查文本内容
	part := content.Parts[0]
	if part.Text == nil {
		t.Fatal("期望文本部分，得到 nil")
	}
	if *part.Text != "Hello, how are you?" {
		t.Errorf("期望文本 'Hello, how are you?', 得到 '%s'", *part.Text)
	}

	// 检查生成配置
	if geminiReq.GenerationConfig == nil {
		t.Fatal("期望 GenerationConfig, 得到 nil")
	}

	// 检查温度参数
	if geminiReq.GenerationConfig.Temperature == nil {
		t.Error("期望设置 Temperature")
	} else if *geminiReq.GenerationConfig.Temperature != 0.8 {
		t.Errorf("期望温度 0.8, 得到 %f", *geminiReq.GenerationConfig.Temperature)
	}

	// 检查最大输出令牌数
	if geminiReq.GenerationConfig.MaxOutputTokens == nil {
		t.Error("期望设置 MaxOutputTokens")
	} else if *geminiReq.GenerationConfig.MaxOutputTokens != 100 {
		t.Errorf("期望最大令牌数 100, 得到 %d", *geminiReq.GenerationConfig.MaxOutputTokens)
	}

	// 将转换后的结果序列化为 JSON 以进行验证
	resultJSON, err := json.Marshal(geminiReq)
	if err != nil {
		t.Fatalf("无法将结果序列化为 JSON: %v", err)
	}

	// 解析结果 JSON 以进行详细检查
	var resultObj map[string]interface{}
	err = json.Unmarshal(resultJSON, &resultObj)
	if err != nil {
		t.Fatalf("无法解析结果 JSON: %v", err)
	}

	// 验证结构
	if _, exists := resultObj["contents"]; !exists {
		t.Error("结果中应包含 'contents' 字段")
	}

	if _, exists := resultObj["generationConfig"]; !exists {
		t.Error("结果中应包含 'generationConfig' 字段")
	}
}

func TestConvertRequest_TextMessageWithMultipleParts(t *testing.T) {
	// 准备测试数据：一个包含多个文本部分的 JSON
	coreReqJSON := `{
		"model": "gemini-pro",
		"messages": [
			{
				"role": "user",
				"content": [
					{
						"type": "text",
						"text": "What is in this image?"
					}
				]
			}
		]
	}`

	// 解析 JSON 数据到核心请求对象
	var request coreTypes.Request
	err := json.Unmarshal([]byte(coreReqJSON), &request)
	if err != nil {
		t.Fatalf("解析 JSON 失败: %v", err)
	}

	// 创建一个模拟的通道对象
	channel := &routing.Channel{}

	// 执行转换
	geminiReq := converter.ConvertRequest(&request, channel)

	// 验证结果
	if geminiReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// 检查基本结构
	if len(geminiReq.Contents) != 1 {
		t.Errorf("期望 1 个内容，得到 %d", len(geminiReq.Contents))
	}

	// 检查用户消息
	content := geminiReq.Contents[0]
	if content.Role != "user" {
		t.Errorf("期望角色 'user', 得到 '%s'", content.Role)
	}

	// 检查消息部分
	if len(content.Parts) != 1 {
		t.Errorf("期望 1 个部分，得到 %d", len(content.Parts))
	}

	// 检查文本内容
	part := content.Parts[0]
	if part.Text == nil {
		t.Fatal("期望文本部分，得到 nil")
	}
	if *part.Text != "What is in this image?" {
		t.Errorf("期望文本 'What is in this image?', 得到 '%s'", *part.Text)
	}
}

// TestConvertCoreRequest_TextMessage 测试从 Gemini 请求到核心请求的反向转换
func TestConvertCoreRequest_TextMessage(t *testing.T) {
	// 准备测试数据：一个包含文本内容的 Gemini 请求 JSON
	geminiReqJSON := `{
		"contents": [
			{
				"role": "user",
				"parts": [
					{
						"text": "Hello, how are you?"
					}
				]
			}
		],
		"generationConfig": {
			"temperature": 0.8,
			"maxOutputTokens": 100
		}
	}`

	// 解析 JSON 数据到 Gemini 请求对象
	var geminiReq types.Request
	err := json.Unmarshal([]byte(geminiReqJSON), &geminiReq)
	if err != nil {
		t.Fatalf("解析 JSON 失败: %v", err)
	}

	// 执行反向转换
	coreReq := converter.ConvertCoreRequest(&geminiReq)

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查消息
	if len(coreReq.Messages) != 1 {
		t.Errorf("期望 1 条消息，得到 %d", len(coreReq.Messages))
	}

	// 检查用户消息
	message := coreReq.Messages[0]
	if message.Role != "user" {
		t.Errorf("期望角色 'user', 得到 '%s'", message.Role)
	}

	// 检查文本内容
	if message.Content.StringValue == nil {
		t.Fatal("期望文本内容，得到 nil")
	}
	if *message.Content.StringValue != "Hello, how are you?" {
		t.Errorf("期望文本 'Hello, how are you?', 得到 '%s'", *message.Content.StringValue)
	}

	// 检查温度参数
	if coreReq.Temperature == nil {
		t.Error("期望设置 Temperature")
	} else if *coreReq.Temperature != 0.8 {
		t.Errorf("期望温度 0.8, 得到 %f", *coreReq.Temperature)
	}

	// 检查最大令牌数
	if coreReq.MaxTokens == nil {
		t.Error("期望设置 MaxTokens")
	} else if *coreReq.MaxTokens != 100 {
		t.Errorf("期望最大令牌数 100, 得到 %d", *coreReq.MaxTokens)
	}

	// 将结果序列化为 JSON 以进行验证
	resultJSON, err := json.Marshal(coreReq)
	if err != nil {
		t.Fatalf("无法将结果序列化为 JSON: %v", err)
	}

	// 解析结果 JSON 以进行详细检查
	var resultObj map[string]interface{}
	err = json.Unmarshal(resultJSON, &resultObj)
	if err != nil {
		t.Fatalf("无法解析结果 JSON: %v", err)
	}

	// 验证结构
	if _, exists := resultObj["messages"]; !exists {
		t.Error("结果中应包含 'messages' 字段")
	}

	if _, exists := resultObj["temperature"]; !exists {
		t.Error("结果中应包含 'temperature' 字段")
	}

	if _, exists := resultObj["max_tokens"]; !exists {
		t.Error("结果中应包含 'max_tokens' 字段")
	}
}

// TestConvertCoreRequest_WithTools 测试指定工具时的反向转换
func TestConvertCoreRequest_WithTools(t *testing.T) {
	// 准备测试数据：一个包含工具的 Gemini 请求
	geminiReqJSON := `{
		"contents": [
			{
				"role": "user",
				"parts": [
					{
						"text": "Test message"
					}
				]
			}
		],
		"tools": [
			{
				"functionDeclarations": [
					{
						"name": "testFunction",
						"description": "Test function description",
						"parameters": {
							"type": "object",
							"properties": {}
						}
					}
				]
			}
		]
	}`

	// 解析 JSON 数据到 Gemini 请求对象
	var geminiReq types.Request
	err := json.Unmarshal([]byte(geminiReqJSON), &geminiReq)
	if err != nil {
		t.Fatalf("解析 JSON 失败: %v", err)
	}

	// 执行反向转换
	coreReq := converter.ConvertCoreRequest(&geminiReq)

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查工具是否正确转换回来
	if len(coreReq.Tools) == 0 {
		t.Error("结果中应包含工具")
	}

	// 检查转换后的工具结构
	tool := coreReq.Tools[0]
	if tool.Type != "function" {
		t.Errorf("期望工具类型 'function', 得到 '%s'", tool.Type)
	}
	if tool.Function.Name != "testFunction" {
		t.Errorf("期望函数名 'testFunction', 得到 '%s'", tool.Function.Name)
	}
	if tool.Function.Description == nil || *tool.Function.Description != "Test function description" {
		t.Errorf("期望描述 'Test function description', 得到 '%v'", tool.Function.Description)
	}
}

// TestConvertCoreRequest_WithToolChoice 测试指定工具选择时的反向转换
func TestConvertCoreRequest_WithToolChoice(t *testing.T) {
	// 准备测试数据：一个包含工具选择的 Gemini 请求
	geminiReqJSON := `{
		"contents": [
			{
				"role": "user",
				"parts": [
					{
						"text": "Test message"
					}
				]
			}
		],
		"toolConfig": {
			"functionCallingConfig": {
				"mode": "AUTO"
			}
		}
	}`

	// 解析 JSON 数据到 Gemini 请求对象
	var geminiReq types.Request
	err := json.Unmarshal([]byte(geminiReqJSON), &geminiReq)
	if err != nil {
		t.Fatalf("解析 JSON 失败: %v", err)
	}

	// 执行反向转换
	coreReq := converter.ConvertCoreRequest(&geminiReq)

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查工具选择是否正确转换回来
	if coreReq.ToolChoice.Mode == nil {
		t.Error("期望设置工具选择模式")
	} else if *coreReq.ToolChoice.Mode != "auto" {
		t.Errorf("期望工具选择模式 'auto', 得到 '%s'", *coreReq.ToolChoice.Mode)
	}
}
