package converter_test

import (
	"encoding/json"
	"testing"

	"github.com/MeowSalty/portal/request/adapter/anthropic/converter"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
)

// TestConvertRequest_SimpleTextMessage 测试简单文本消息转换
func TestConvertRequest_SimpleTextMessage(t *testing.T) {
	// 准备测试数据：一个包含简单文本消息的 JSON
	coreReqJSON := `{
		"model": "claude-sonnet-3",
		"messages": [
			{
				"role": "user",
				"content": "你好，请介绍一下你自己"
			}
		],
		"temperature": 0.7,
		"max_tokens": 500
	}`

	// 解析 JSON 数据到核心请求对象
	var request coreTypes.Request
	err := json.Unmarshal([]byte(coreReqJSON), &request)
	if err != nil {
		t.Fatalf("解析 JSON 失败: %v", err)
	}

	// 创建一个模拟的通道对象
	channel := &routing.Channel{
		ModelName: "claude-3-sonnet-20240229",
	}

	// 执行转换
	anthropicReq := converter.ConvertRequest(&request, channel)

	// 验证结果
	if anthropicReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// 检查基本字段
	if anthropicReq.Model != "claude-3-sonnet-20240229" {
		t.Errorf("期望模型 'claude-3-sonnet-20240229', 得到 '%s'", anthropicReq.Model)
	}

	if anthropicReq.MaxTokens != 500 {
		t.Errorf("期望最大令牌数 500, 得到 %d", anthropicReq.MaxTokens)
	}

	// 检查温度参数
	if anthropicReq.Temperature == nil {
		t.Error("期望设置 Temperature")
	} else if *anthropicReq.Temperature != 0.7 {
		t.Errorf("期望温度 0.7, 得到 %f", *anthropicReq.Temperature)
	}

	// 检查消息
	if len(anthropicReq.Messages) != 1 {
		t.Errorf("期望 1 条消息，得到 %d", len(anthropicReq.Messages))
	}

	message := anthropicReq.Messages[0]
	if message.Role != "user" {
		t.Errorf("期望角色 'user', 得到 '%s'", message.Role)
	}

	// 检查消息内容
	content, ok := message.Content.(string)
	if !ok {
		t.Fatal("期望消息内容为字符串")
	}
	if content != "你好，请介绍一下你自己" {
		t.Errorf("期望内容 '你好，请介绍一下你自己', 得到 '%s'", content)
	}

	// 检查系统消息（应该为 nil）
	if anthropicReq.System != nil {
		t.Error("期望系统消息为 nil")
	}
}

// TestConvertRequest_MultiRoundMessages 测试多轮对话消息转换
func TestConvertRequest_MultiRoundMessages(t *testing.T) {
	// 准备测试数据：多轮对话
	coreReqJSON := `{
		"model": "claude-3-sonnet-20240229",
		"messages": [
			{
				"role": "system",
				"content": "你是一个有帮助的助手"
			},
			{
				"role": "user",
				"content": "什么是人工智能？"
			},
			{
				"role": "assistant",
				"content": "人工智能是..."
			},
			{
				"role": "user",
				"content": "能详细解释一下吗？"
			}
		],
		"max_tokens": 1000
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
	anthropicReq := converter.ConvertRequest(&request, channel)

	// 验证结果
	if anthropicReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// 检查系统消息
	if anthropicReq.System == nil {
		t.Error("期望设置系统消息")
	} else {
		systemContent, ok := anthropicReq.System.(string)
		if !ok {
			t.Fatal("期望系统消息为字符串")
		}
		if systemContent != "你是一个有帮助的助手" {
			t.Errorf("期望系统消息 '你是一个有帮助的助手', 得到 '%s'", systemContent)
		}
	}

	// 检查消息数量（系统消息应该被提取，剩下 3 条）
	if len(anthropicReq.Messages) != 3 {
		t.Errorf("期望 3 条消息（除去系统消息），得到 %d", len(anthropicReq.Messages))
	}

	// 检查第一条用户消息
	if anthropicReq.Messages[0].Role != "user" {
		t.Errorf("期望第一条消息角色 'user', 得到 '%s'", anthropicReq.Messages[0].Role)
	}
	if anthropicReq.Messages[0].Content != "什么是人工智能？" {
		t.Errorf("期望第一条消息内容 '什么是人工智能？', 得到 '%v'", anthropicReq.Messages[0].Content)
	}

	// 检查助手消息
	if anthropicReq.Messages[1].Role != "assistant" {
		t.Errorf("期望第二条消息角色 'assistant', 得到 '%s'", anthropicReq.Messages[1].Role)
	}

	// 检查第二条用户消息
	if anthropicReq.Messages[2].Role != "user" {
		t.Errorf("期望第三条消息角色 'user', 得到 '%s'", anthropicReq.Messages[2].Role)
	}
}

// TestConvertRequest_MultiRoundMessagesWithContentParts 测试包含内容部分的多轮消息
func TestConvertRequest_MultiRoundMessagesWithContentParts(t *testing.T) {
	// 准备测试数据：包含文本和图像的多轮对话
	coreReqJSON := `{
		"model": "claude-3-sonnet-20240229",
		"messages": [
			{
				"role": "user",
				"content": [
					{
						"type": "text",
						"text": "这张图片里有什么？"
					},
					{
						"type": "image_url",
						"image_url": {
							"url": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
						}
					}
				]
			},
			{
				"role": "assistant",
				"content": "这是一张测试图片"
			},
			{
				"role": "user",
				"content": [
					{
						"type": "text",
						"text": "能详细描述一下吗？"
					}
				]
			}
		],
		"max_tokens": 1500
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
	anthropicReq := converter.ConvertRequest(&request, channel)

	// 验证结果
	if anthropicReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// 检查消息数量
	if len(anthropicReq.Messages) != 3 {
		t.Errorf("期望 3 条消息，得到 %d", len(anthropicReq.Messages))
	}

	// 检查第一条消息（包含文本和图像）
	firstMessage := anthropicReq.Messages[0]
	if firstMessage.Role != "user" {
		t.Errorf("期望第一条消息角色 'user', 得到 '%s'", firstMessage.Role)
	}

	// 检查内容块
	contentBlocks, ok := firstMessage.Content.([]anthropicTypes.ContentBlock)
	if !ok {
		t.Fatal("期望第一条消息内容为 ContentBlock 数组")
	}
	if len(contentBlocks) != 2 {
		t.Errorf("期望 2 个内容块，得到 %d", len(contentBlocks))
	}

	// 检查文本块
	if contentBlocks[0].Type != "text" {
		t.Errorf("期望第一个块类型 'text', 得到 '%s'", contentBlocks[0].Type)
	}
	if contentBlocks[0].Text == nil || *contentBlocks[0].Text != "这张图片里有什么？" {
		t.Errorf("期望文本内容 '这张图片里有什么？', 得到 '%v'", contentBlocks[0].Text)
	}

	// 检查图像块
	if contentBlocks[1].Type != "image" {
		t.Errorf("期望第二个块类型 'image', 得到 '%s'", contentBlocks[1].Type)
	}
	if contentBlocks[1].Source == nil {
		t.Fatal("期望图像源不为 nil")
	}
	if contentBlocks[1].Source.Type != "base64" {
		t.Errorf("期望图像源类型 'base64', 得到 '%s'", contentBlocks[1].Source.Type)
	}
	if contentBlocks[1].Source.MediaType != "image/png" {
		t.Errorf("期望媒体类型 'image/png', 得到 '%s'", contentBlocks[1].Source.MediaType)
	}

	// 检查第二条消息（助手回复）
	secondMessage := anthropicReq.Messages[1]
	if secondMessage.Role != "assistant" {
		t.Errorf("期望第二条消息角色 'assistant', 得到 '%s'", secondMessage.Role)
	}
	if secondMessage.Content != "这是一张测试图片" {
		t.Errorf("期望助手回复 '这是一张测试图片', 得到 '%v'", secondMessage.Content)
	}

	// 检查第三条消息（用户追问）
	thirdMessage := anthropicReq.Messages[2]
	if thirdMessage.Role != "user" {
		t.Errorf("期望第三条消息角色 'user', 得到 '%s'", thirdMessage.Role)
	}

	// 检查内容块
	thirdContentBlocks, ok := thirdMessage.Content.([]anthropicTypes.ContentBlock)
	if !ok {
		t.Fatal("期望第三条消息内容为 ContentBlock 数组")
	}
	if len(thirdContentBlocks) != 1 {
		t.Errorf("期望 1 个内容块，得到 %d", len(thirdContentBlocks))
	}
	if thirdContentBlocks[0].Type != "text" {
		t.Errorf("期望块类型 'text', 得到 '%s'", thirdContentBlocks[0].Type)
	}
}

// TestConvertRequest_WithTools 测试包含工具调用的请求转换
func TestConvertRequest_WithTools(t *testing.T) {
	// 准备测试数据：包含工具定义和工具选择
	coreReqJSON := `{
		"model": "claude-3-sonnet-20240229",
		"messages": [
			{
				"role": "user",
				"content": "现在北京天气怎么样？"
			}
		],
		"tools": [
			{
				"type": "function",
				"function": {
					"name": "get_weather",
					"description": "获取指定城市的天气信息",
					"parameters": {
						"type": "object",
						"properties": {
							"city": {
								"type": "string",
								"description": "城市名称"
							},
							"unit": {
								"type": "string",
								"enum": ["celsius", "fahrenheit"],
								"description": "温度单位"
							}
						},
						"required": ["city"]
					}
				}
			}
		],
		"tool_choice": "auto",
		"max_tokens": 1000
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
	anthropicReq := converter.ConvertRequest(&request, channel)

	// 验证结果
	if anthropicReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// 检查工具
	if len(anthropicReq.Tools) != 1 {
		t.Errorf("期望 1 个工具，得到 %d", len(anthropicReq.Tools))
	}

	tool := anthropicReq.Tools[0]
	if tool.Name != "get_weather" {
		t.Errorf("期望工具名称 'get_weather', 得到 '%s'", tool.Name)
	}
	if tool.Description == nil || *tool.Description != "获取指定城市的天气信息" {
		t.Errorf("期望工具描述 '获取指定城市的天气信息', 得到 '%v'", tool.Description)
	}

	// 检查工具参数 schema
	if tool.InputSchema.Type != "object" {
		t.Errorf("期望 schema 类型 'object', 得到 '%s'", tool.InputSchema.Type)
	}
	if tool.InputSchema.Properties == nil {
		t.Error("期望 schema 包含 properties")
	} else {
		if _, exists := tool.InputSchema.Properties["city"]; !exists {
			t.Error("期望 properties 中包含 'city' 字段")
		}
		if _, exists := tool.InputSchema.Properties["unit"]; !exists {
			t.Error("期望 properties 中包含 'unit' 字段")
		}
	}
	if len(tool.InputSchema.Required) != 1 || tool.InputSchema.Required[0] != "city" {
		t.Errorf("期望 required 字段为 ['city'], 得到 %v", tool.InputSchema.Required)
	}

	// 检查工具选择
	if anthropicReq.ToolChoice == nil {
		t.Fatal("期望设置工具选择")
	}
	toolChoice, ok := anthropicReq.ToolChoice.(anthropicTypes.ToolChoiceAuto)
	if !ok {
		t.Fatal("期望工具选择为 ToolChoiceAuto 类型")
	}
	if toolChoice.Type != "auto" {
		t.Errorf("期望工具选择类型 'auto', 得到 '%s'", toolChoice.Type)
	}
}

// TestConvertRequest_WithSpecificToolChoice 测试指定特定工具的选择
func TestConvertRequest_WithSpecificToolChoice(t *testing.T) {
	// 准备测试数据：指定特定工具
	coreReqJSON := `{
		"model": "claude-3-sonnet-20240229",
		"messages": [
			{
				"role": "user",
				"content": "获取北京天气"
			}
		],
		"tools": [
			{
				"type": "function",
				"function": {
					"name": "get_weather",
					"description": "获取天气信息",
					"parameters": {
						"type": "object",
						"properties": {
							"city": {
								"type": "string"
							}
						}
					}
				}
			}
		],
		"tool_choice": {
			"type": "function",
			"function": {
				"name": "get_weather"
			}
		}
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
	anthropicReq := converter.ConvertRequest(&request, channel)

	// 验证结果
	if anthropicReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// 检查工具选择
	if anthropicReq.ToolChoice == nil {
		t.Fatal("期望设置工具选择")
	}
	toolChoice, ok := anthropicReq.ToolChoice.(anthropicTypes.ToolChoiceTool)
	if !ok {
		t.Fatal("期望工具选择为 ToolChoiceTool 类型")
	}
	if toolChoice.Type != "tool" {
		t.Errorf("期望工具选择类型 'tool', 得到 '%s'", toolChoice.Type)
	}
	if toolChoice.Name != "get_weather" {
		t.Errorf("期望工具名称 'get_weather', 得到 '%s'", toolChoice.Name)
	}
}

// TestConvertRequest_WithParameters 测试各种请求参数转换
func TestConvertRequest_WithParameters(t *testing.T) {
	// 准备测试数据：包含各种参数
	temperature := 0.8
	topP := 0.9
	topK := 40
	maxTokens := 2000
	stopSequence := "END"
	userID := "user123"

	coreReq := &coreTypes.Request{
		Model:       "claude-opus-3",
		Temperature: &temperature,
		TopP:        &topP,
		TopK:        &topK,
		MaxTokens:   &maxTokens,
		Messages: []coreTypes.Message{
			{
				Role: "user",
				Content: coreTypes.MessageContent{
					StringValue: &[]string{"测试消息"}[0],
				},
			},
		},
		Stop: coreTypes.Stop{
			StringValue: &stopSequence,
		},
		User:   &userID,
		Stream: &[]bool{true}[0],
	}

	// 创建一个模拟的通道对象
	channel := &routing.Channel{
		ModelName: "claude-3-opus-20240229",
	}

	// 执行转换
	anthropicReq := converter.ConvertRequest(coreReq, channel)

	// 验证结果
	if anthropicReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// 检查基本参数
	if anthropicReq.Model != "claude-3-opus-20240229" {
		t.Errorf("期望模型 'claude-3-opus-20240229', 得到 '%s'", anthropicReq.Model)
	}
	if anthropicReq.MaxTokens != 2000 {
		t.Errorf("期望最大令牌数 2000, 得到 %d", anthropicReq.MaxTokens)
	}

	// 检查温度
	if anthropicReq.Temperature == nil {
		t.Error("期望设置 Temperature")
	} else if *anthropicReq.Temperature != 0.8 {
		t.Errorf("期望温度 0.8, 得到 %f", *anthropicReq.Temperature)
	}

	// 检查 TopP
	if anthropicReq.TopP == nil {
		t.Error("期望设置 TopP")
	} else if *anthropicReq.TopP != 0.9 {
		t.Errorf("期望 TopP 0.9, 得到 %f", *anthropicReq.TopP)
	}

	// 检查 TopK
	if anthropicReq.TopK == nil {
		t.Error("期望设置 TopK")
	} else if *anthropicReq.TopK != 40 {
		t.Errorf("期望 TopK 40, 得到 %d", *anthropicReq.TopK)
	}

	// 检查停止序列
	if len(anthropicReq.StopSequences) != 1 {
		t.Errorf("期望 1 个停止序列，得到 %d", len(anthropicReq.StopSequences))
	} else if anthropicReq.StopSequences[0] != "END" {
		t.Errorf("期望停止序列 'END', 得到 '%s'", anthropicReq.StopSequences[0])
	}

	// 检查流式传输
	if anthropicReq.Stream == nil {
		t.Error("期望设置 Stream")
	} else if *anthropicReq.Stream != true {
		t.Error("期望 Stream 为 true")
	}

	// 检查用户元数据
	if anthropicReq.Metadata == nil {
		t.Error("期望设置 Metadata")
	} else if anthropicReq.Metadata.UserID == nil || *anthropicReq.Metadata.UserID != "user123" {
		t.Errorf("期望用户 ID 'user123', 得到 '%v'", anthropicReq.Metadata.UserID)
	}
}

// TestConvertRequest_WithMultipleStopSequences 测试多个停止序列
func TestConvertRequest_WithMultipleStopSequences(t *testing.T) {
	stopSequences := []string{"END", "STOP", "完成"}
	coreReq := &coreTypes.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []coreTypes.Message{
			{
				Role: "user",
				Content: coreTypes.MessageContent{
					StringValue: &[]string{"测试"}[0],
				},
			},
		},
		Stop: coreTypes.Stop{
			StringArray: stopSequences,
		},
	}

	// 创建一个模拟的通道对象
	channel := &routing.Channel{}

	// 执行转换
	anthropicReq := converter.ConvertRequest(coreReq, channel)

	// 验证结果
	if anthropicReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// 检查停止序列
	if len(anthropicReq.StopSequences) != 3 {
		t.Errorf("期望 3 个停止序列，得到 %d", len(anthropicReq.StopSequences))
	}
	for i, seq := range stopSequences {
		if anthropicReq.StopSequences[i] != seq {
			t.Errorf("期望停止序列[%d] '%s', 得到 '%s'", i, seq, anthropicReq.StopSequences[i])
		}
	}
}

// TestConvertRequest_WithDefaultMaxTokens 测试默认最大令牌数
func TestConvertRequest_WithDefaultMaxTokens(t *testing.T) {
	coreReq := &coreTypes.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []coreTypes.Message{
			{
				Role: "user",
				Content: coreTypes.MessageContent{
					StringValue: &[]string{"测试"}[0],
				},
			},
		},
		// 不设置 MaxTokens
	}

	// 创建一个模拟的通道对象
	channel := &routing.Channel{}

	// 执行转换
	anthropicReq := converter.ConvertRequest(coreReq, channel)

	// 验证结果
	if anthropicReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// 检查默认最大令牌数
	if anthropicReq.MaxTokens != 1024 {
		t.Errorf("期望默认最大令牌数 1024, 得到 %d", anthropicReq.MaxTokens)
	}
}

// TestConvertRequest_WithToolChoiceNone 测试工具选择为 none
func TestConvertRequest_WithToolChoiceNone(t *testing.T) {
	mode := "none"
	coreReq := &coreTypes.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []coreTypes.Message{
			{
				Role: "user",
				Content: coreTypes.MessageContent{
					StringValue: &[]string{"测试"}[0],
				},
			},
		},
		ToolChoice: coreTypes.ToolChoice{
			Mode: &mode,
		},
	}

	// 创建一个模拟的通道对象
	channel := &routing.Channel{}

	// 执行转换
	anthropicReq := converter.ConvertRequest(coreReq, channel)

	// 验证结果
	if anthropicReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// Anthropic 没有 "none" 选项，应该返回 nil
	if anthropicReq.ToolChoice != nil {
		t.Error("工具选择为 'none' 时，Anthropic 应该返回 nil")
	}
}

// TestConvertRequest_WithRegularImageURL 测试普通图像 URL（非 base64）
func TestConvertRequest_WithRegularImageURL(t *testing.T) {
	coreReq := &coreTypes.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []coreTypes.Message{
			{
				Role: "user",
				Content: coreTypes.MessageContent{
					ContentParts: []coreTypes.ContentPart{
						{
							Type: "image_url",
							ImageURL: &coreTypes.ImageURL{
								URL: "https://example.com/image.jpg",
							},
						},
					},
				},
			},
		},
	}

	// 创建一个模拟的通道对象
	channel := &routing.Channel{}

	// 执行转换
	anthropicReq := converter.ConvertRequest(coreReq, channel)

	// 验证结果
	if anthropicReq == nil {
		t.Fatal("ConvertRequest 返回 nil")
	}

	// 检查消息
	if len(anthropicReq.Messages) != 1 {
		t.Errorf("期望 1 条消息，得到 %d", len(anthropicReq.Messages))
	}

	message := anthropicReq.Messages[0]
	contentBlocks, ok := message.Content.([]anthropicTypes.ContentBlock)
	if !ok {
		t.Fatal("期望消息内容为 ContentBlock 数组")
	}
	if len(contentBlocks) != 1 {
		t.Errorf("期望 1 个内容块，得到 %d", len(contentBlocks))
	}

	// 普通 URL 应该被忽略，Source 为 nil
	if contentBlocks[0].Source != nil {
		t.Error("普通图像 URL 的 Source 应该为 nil")
	}
}
