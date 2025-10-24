package types_test

import (
	"testing"

	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
)

// TestConvertCoreRequest_SimpleTextMessage 测试简单文本消息转换
func TestConvertCoreRequest_SimpleTextMessage(t *testing.T) {
	// 准备测试数据：一个包含简单文本消息的 Anthropic 请求
	text := "你好，请介绍一下你自己"
	temperature := 0.7
	anthropicReq := &anthropicTypes.Request{
		Model:     "claude-3-sonnet-20240229",
		MaxTokens: 500,
		Messages: []anthropicTypes.InputMessage{
			{
				Role:    "user",
				Content: text,
			},
		},
		Temperature: &temperature,
	}

	// 执行转换
	coreReq := anthropicReq.ConvertCoreRequest()

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查基本字段
	if coreReq.Model != "claude-3-sonnet-20240229" {
		t.Errorf("期望模型 'claude-3-sonnet-20240229', 得到 '%s'", coreReq.Model)
	}

	if coreReq.MaxTokens == nil || *coreReq.MaxTokens != 500 {
		t.Errorf("期望最大令牌数 500, 得到 %v", coreReq.MaxTokens)
	}

	// 检查温度参数
	if coreReq.Temperature == nil {
		t.Error("期望设置 Temperature")
	} else if *coreReq.Temperature != 0.7 {
		t.Errorf("期望温度 0.7, 得到 %f", *coreReq.Temperature)
	}

	// 检查消息
	if len(coreReq.Messages) != 1 {
		t.Fatalf("期望 1 条消息，得到 %d", len(coreReq.Messages))
	}

	message := coreReq.Messages[0]
	if message.Role != "user" {
		t.Errorf("期望角色 'user', 得到 '%s'", message.Role)
	}

	// 检查消息内容
	if message.Content.StringValue == nil {
		t.Fatal("期望消息内容为字符串")
	}
	if *message.Content.StringValue != text {
		t.Errorf("期望内容 '%s', 得到 '%s'", text, *message.Content.StringValue)
	}
}

// TestConvertCoreRequest_MultiRoundTextMessages 测试多轮文本消息转换
func TestConvertCoreRequest_MultiRoundTextMessages(t *testing.T) {
	// 准备测试数据：多轮对话
	systemMsg := "你是一个有帮助的助手"
	anthropicReq := &anthropicTypes.Request{
		Model:     "claude-3-sonnet-20240229",
		MaxTokens: 1000,
		System:    systemMsg,
		Messages: []anthropicTypes.InputMessage{
			{
				Role:    "user",
				Content: "什么是人工智能？",
			},
			{
				Role:    "assistant",
				Content: "人工智能是计算机科学的一个分支...",
			},
			{
				Role:    "user",
				Content: "能详细解释一下吗？",
			},
		},
	}

	// 执行转换
	coreReq := anthropicReq.ConvertCoreRequest()

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查消息数量（系统消息 + 3 条对话消息）
	if len(coreReq.Messages) != 4 {
		t.Fatalf("期望 4 条消息（包含系统消息），得到 %d", len(coreReq.Messages))
	}

	// 检查系统消息
	systemMessage := coreReq.Messages[0]
	if systemMessage.Role != "system" {
		t.Errorf("期望第一条消息角色 'system', 得到 '%s'", systemMessage.Role)
	}
	if systemMessage.Content.StringValue == nil || *systemMessage.Content.StringValue != systemMsg {
		t.Errorf("期望系统消息 '%s', 得到 '%v'", systemMsg, systemMessage.Content.StringValue)
	}

	// 检查第一条用户消息
	if coreReq.Messages[1].Role != "user" {
		t.Errorf("期望第二条消息角色 'user', 得到 '%s'", coreReq.Messages[1].Role)
	}
	if coreReq.Messages[1].Content.StringValue == nil || *coreReq.Messages[1].Content.StringValue != "什么是人工智能？" {
		t.Errorf("期望用户消息 '什么是人工智能？', 得到 '%v'", coreReq.Messages[1].Content.StringValue)
	}

	// 检查助手消息
	if coreReq.Messages[2].Role != "assistant" {
		t.Errorf("期望第三条消息角色 'assistant', 得到 '%s'", coreReq.Messages[2].Role)
	}

	// 检查第二条用户消息
	if coreReq.Messages[3].Role != "user" {
		t.Errorf("期望第四条消息角色 'user', 得到 '%s'", coreReq.Messages[3].Role)
	}
}

// TestConvertCoreRequest_MultiRoundMessagesWithContentBlocks 测试包含内容块的多轮消息
func TestConvertCoreRequest_MultiRoundMessagesWithContentBlocks(t *testing.T) {
	// 准备测试数据：包含文本和图像的多轮对话
	textContent1 := "这张图片里有什么？"
	textContent2 := "能详细描述一下吗？"
	assistantReply := "这是一张测试图片"

	anthropicReq := &anthropicTypes.Request{
		Model:     "claude-3-sonnet-20240229",
		MaxTokens: 1500,
		Messages: []anthropicTypes.InputMessage{
			{
				Role: "user",
				Content: []anthropicTypes.ContentBlock{
					{
						Type: "text",
						Text: &textContent1,
					},
					{
						Type: "image",
						Source: &anthropicTypes.ImageSource{
							Type:      "base64",
							MediaType: "image/png",
							Data:      "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
						},
					},
				},
			},
			{
				Role:    "assistant",
				Content: assistantReply,
			},
			{
				Role: "user",
				Content: []anthropicTypes.ContentBlock{
					{
						Type: "text",
						Text: &textContent2,
					},
				},
			},
		},
	}

	// 执行转换
	coreReq := anthropicReq.ConvertCoreRequest()

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查消息数量
	if len(coreReq.Messages) != 3 {
		t.Fatalf("期望 3 条消息，得到 %d", len(coreReq.Messages))
	}

	// 检查第一条消息（包含文本和图像）
	firstMessage := coreReq.Messages[0]
	if firstMessage.Role != "user" {
		t.Errorf("期望第一条消息角色 'user', 得到 '%s'", firstMessage.Role)
	}

	// 检查内容部分
	if len(firstMessage.Content.ContentParts) != 2 {
		t.Fatalf("期望 2 个内容部分，得到 %d", len(firstMessage.Content.ContentParts))
	}

	// 检查文本部分
	textPart := firstMessage.Content.ContentParts[0]
	if textPart.Type != "text" {
		t.Errorf("期望第一个部分类型 'text', 得到 '%s'", textPart.Type)
	}
	if textPart.Text == nil || *textPart.Text != textContent1 {
		t.Errorf("期望文本内容 '%s', 得到 '%v'", textContent1, textPart.Text)
	}

	// 检查图像部分
	imagePart := firstMessage.Content.ContentParts[1]
	if imagePart.Type != "image_url" {
		t.Errorf("期望第二个部分类型 'image_url', 得到 '%s'", imagePart.Type)
	}
	if imagePart.ImageURL == nil {
		t.Fatal("期望图像 URL 不为 nil")
	}
	expectedURL := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
	if imagePart.ImageURL.URL != expectedURL {
		t.Errorf("期望图像 URL '%s', 得到 '%s'", expectedURL, imagePart.ImageURL.URL)
	}

	// 检查第二条消息（助手回复）
	secondMessage := coreReq.Messages[1]
	if secondMessage.Role != "assistant" {
		t.Errorf("期望第二条消息角色 'assistant', 得到 '%s'", secondMessage.Role)
	}
	if secondMessage.Content.StringValue == nil || *secondMessage.Content.StringValue != assistantReply {
		t.Errorf("期望助手回复 '%s', 得到 '%v'", assistantReply, secondMessage.Content.StringValue)
	}

	// 检查第三条消息（用户追问）
	thirdMessage := coreReq.Messages[2]
	if thirdMessage.Role != "user" {
		t.Errorf("期望第三条消息角色 'user', 得到 '%s'", thirdMessage.Role)
	}

	if len(thirdMessage.Content.ContentParts) != 1 {
		t.Fatalf("期望 1 个内容部分，得到 %d", len(thirdMessage.Content.ContentParts))
	}
	if thirdMessage.Content.ContentParts[0].Type != "text" {
		t.Errorf("期望部分类型 'text', 得到 '%s'", thirdMessage.Content.ContentParts[0].Type)
	}
}

// TestConvertCoreRequest_WithToolCalls 测试包含工具调用的多轮消息
func TestConvertCoreRequest_WithToolCalls(t *testing.T) {
	// 准备测试数据：包含工具定义和工具选择
	toolDesc := "获取指定城市的天气信息"
	anthropicReq := &anthropicTypes.Request{
		Model:     "claude-3-sonnet-20240229",
		MaxTokens: 1000,
		Messages: []anthropicTypes.InputMessage{
			{
				Role:    "user",
				Content: "现在北京天气怎么样？",
			},
		},
		Tools: []anthropicTypes.Tool{
			{
				Name:        "get_weather",
				Description: &toolDesc,
				InputSchema: anthropicTypes.InputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"city": map[string]interface{}{
							"type":        "string",
							"description": "城市名称",
						},
						"unit": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"celsius", "fahrenheit"},
							"description": "温度单位",
						},
					},
					Required: []string{"city"},
				},
			},
		},
		ToolChoice: anthropicTypes.ToolChoiceAuto{
			Type: "auto",
		},
	}

	// 执行转换
	coreReq := anthropicReq.ConvertCoreRequest()

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查工具
	if len(coreReq.Tools) != 1 {
		t.Fatalf("期望 1 个工具，得到 %d", len(coreReq.Tools))
	}

	tool := coreReq.Tools[0]
	if tool.Type != "function" {
		t.Errorf("期望工具类型 'function', 得到 '%s'", tool.Type)
	}
	if tool.Function.Name != "get_weather" {
		t.Errorf("期望工具名称 'get_weather', 得到 '%s'", tool.Function.Name)
	}
	if tool.Function.Description == nil || *tool.Function.Description != toolDesc {
		t.Errorf("期望工具描述 '%s', 得到 '%v'", toolDesc, tool.Function.Description)
	}

	// 检查工具参数
	params, ok := tool.Function.Parameters.(map[string]interface{})
	if !ok {
		t.Fatal("期望工具参数为 map[string]interface{}")
	}
	if params["type"] != "object" {
		t.Errorf("期望参数类型 'object', 得到 '%v'", params["type"])
	}

	properties, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("期望 properties 为 map[string]interface{}")
	}
	if _, exists := properties["city"]; !exists {
		t.Error("期望 properties 中包含 'city' 字段")
	}

	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("期望 required 为 []string")
	}
	if len(required) != 1 || required[0] != "city" {
		t.Errorf("期望 required 字段为 ['city'], 得到 %v", required)
	}

	// 检查工具选择
	if coreReq.ToolChoice.Mode == nil {
		t.Fatal("期望设置工具选择模式")
	}
	if *coreReq.ToolChoice.Mode != "auto" {
		t.Errorf("期望工具选择模式 'auto', 得到 '%s'", *coreReq.ToolChoice.Mode)
	}
}

// TestConvertCoreRequest_WithSpecificToolChoice 测试指定特定工具的选择
func TestConvertCoreRequest_WithSpecificToolChoice(t *testing.T) {
	// 准备测试数据：指定特定工具
	toolDesc := "获取天气信息"
	anthropicReq := &anthropicTypes.Request{
		Model:     "claude-3-sonnet-20240229",
		MaxTokens: 1000,
		Messages: []anthropicTypes.InputMessage{
			{
				Role:    "user",
				Content: "获取北京天气",
			},
		},
		Tools: []anthropicTypes.Tool{
			{
				Name:        "get_weather",
				Description: &toolDesc,
				InputSchema: anthropicTypes.InputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"city": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
		ToolChoice: anthropicTypes.ToolChoiceTool{
			Type: "tool",
			Name: "get_weather",
		},
	}

	// 执行转换
	coreReq := anthropicReq.ConvertCoreRequest()

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查工具选择
	if coreReq.ToolChoice.Function == nil {
		t.Fatal("期望设置特定工具选择")
	}
	if coreReq.ToolChoice.Function.Type != "function" {
		t.Errorf("期望工具选择类型 'function', 得到 '%s'", coreReq.ToolChoice.Function.Type)
	}
	if coreReq.ToolChoice.Function.Function.Name != "get_weather" {
		t.Errorf("期望工具名称 'get_weather', 得到 '%s'", coreReq.ToolChoice.Function.Function.Name)
	}
}

// TestConvertCoreRequest_WithParameters 测试各种请求参数转换
func TestConvertCoreRequest_WithParameters(t *testing.T) {
	// 准备测试数据：包含各种参数
	temperature := 0.8
	topP := 0.9
	topK := 40
	stream := true
	userID := "user123"

	anthropicReq := &anthropicTypes.Request{
		Model:       "claude-3-opus-20240229",
		MaxTokens:   2000,
		Temperature: &temperature,
		TopP:        &topP,
		TopK:        &topK,
		Stream:      &stream,
		Messages: []anthropicTypes.InputMessage{
			{
				Role:    "user",
				Content: "测试消息",
			},
		},
		StopSequences: []string{"END"},
		Metadata: &anthropicTypes.Metadata{
			UserID: &userID,
		},
	}

	// 执行转换
	coreReq := anthropicReq.ConvertCoreRequest()

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查基本参数
	if coreReq.Model != "claude-3-opus-20240229" {
		t.Errorf("期望模型 'claude-3-opus-20240229', 得到 '%s'", coreReq.Model)
	}
	if coreReq.MaxTokens == nil || *coreReq.MaxTokens != 2000 {
		t.Errorf("期望最大令牌数 2000, 得到 %v", coreReq.MaxTokens)
	}

	// 检查温度
	if coreReq.Temperature == nil {
		t.Error("期望设置 Temperature")
	} else if *coreReq.Temperature != 0.8 {
		t.Errorf("期望温度 0.8, 得到 %f", *coreReq.Temperature)
	}

	// 检查 TopP
	if coreReq.TopP == nil {
		t.Error("期望设置 TopP")
	} else if *coreReq.TopP != 0.9 {
		t.Errorf("期望 TopP 0.9, 得到 %f", *coreReq.TopP)
	}

	// 检查 TopK
	if coreReq.TopK == nil {
		t.Error("期望设置 TopK")
	} else if *coreReq.TopK != 40 {
		t.Errorf("期望 TopK 40, 得到 %d", *coreReq.TopK)
	}

	// 检查停止序列
	if coreReq.Stop.StringValue == nil || *coreReq.Stop.StringValue != "END" {
		t.Errorf("期望停止序列 'END', 得到 '%v'", coreReq.Stop.StringValue)
	}

	// 检查流式传输
	if coreReq.Stream == nil {
		t.Error("期望设置 Stream")
	} else if *coreReq.Stream != true {
		t.Errorf("期望 Stream 为 true, 得到 %v", *coreReq.Stream)
	}

	// 检查用户元数据
	if coreReq.User == nil || *coreReq.User != "user123" {
		t.Errorf("期望用户 ID 'user123', 得到 '%v'", coreReq.User)
	}
}

// TestConvertCoreRequest_WithMultipleStopSequences 测试多个停止序列
func TestConvertCoreRequest_WithMultipleStopSequences(t *testing.T) {
	anthropicReq := &anthropicTypes.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []anthropicTypes.InputMessage{
			{
				Role:    "user",
				Content: "测试",
			},
		},
		StopSequences: []string{"END", "STOP", "完成"},
	}

	// 执行转换
	coreReq := anthropicReq.ConvertCoreRequest()

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查停止序列
	if coreReq.Stop.StringArray == nil {
		t.Fatal("期望停止序列数组不为 nil")
	}
	if len(coreReq.Stop.StringArray) != 3 {
		t.Fatalf("期望 3 个停止序列，得到 %d", len(coreReq.Stop.StringArray))
	}
	expectedSequences := []string{"END", "STOP", "完成"}
	for i, seq := range expectedSequences {
		if coreReq.Stop.StringArray[i] != seq {
			t.Errorf("期望停止序列[%d] '%s', 得到 '%s'", i, seq, coreReq.Stop.StringArray[i])
		}
	}
}

// TestConvertCoreRequest_WithSystemAsContentBlocks 测试系统消息作为内容块数组
func TestConvertCoreRequest_WithSystemAsContentBlocks(t *testing.T) {
	// 准备测试数据：系统消息作为内容块数组
	systemText := "你是一个专业的助手"
	systemImage := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="

	anthropicReq := &anthropicTypes.Request{
		Model: "claude-3-sonnet-20240229",
		System: []anthropicTypes.ContentBlock{
			{
				Type: "text",
				Text: &systemText,
			},
			{
				Type: "image",
				Source: &anthropicTypes.ImageSource{
					Type:      "base64",
					MediaType: "image/jpeg",
					Data:      systemImage,
				},
			},
		},
		Messages: []anthropicTypes.InputMessage{
			{
				Role:    "user",
				Content: "你好",
			},
		},
	}

	// 执行转换
	coreReq := anthropicReq.ConvertCoreRequest()

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查消息数量（系统消息 + 1 条用户消息）
	if len(coreReq.Messages) != 2 {
		t.Fatalf("期望 2 条消息（包含系统消息），得到 %d", len(coreReq.Messages))
	}

	// 检查系统消息
	systemMessage := coreReq.Messages[0]
	if systemMessage.Role != "system" {
		t.Errorf("期望系统消息角色 'system', 得到 '%s'", systemMessage.Role)
	}

	// 检查系统消息内容部分
	if len(systemMessage.Content.ContentParts) != 2 {
		t.Fatalf("期望系统消息有 2 个内容部分，得到 %d", len(systemMessage.Content.ContentParts))
	}

	// 检查文本部分
	textPart := systemMessage.Content.ContentParts[0]
	if textPart.Type != "text" {
		t.Errorf("期望文本部分类型 'text', 得到 '%s'", textPart.Type)
	}
	if textPart.Text == nil || *textPart.Text != systemText {
		t.Errorf("期望文本内容 '%s', 得到 '%v'", systemText, textPart.Text)
	}

	// 检查图像部分
	imagePart := systemMessage.Content.ContentParts[1]
	if imagePart.Type != "image_url" {
		t.Errorf("期望图像部分类型 'image_url', 得到 '%s'", imagePart.Type)
	}
	if imagePart.ImageURL == nil {
		t.Fatal("期望图像 URL 不为 nil")
	}
	expectedURL := "data:image/jpeg;base64," + systemImage
	if imagePart.ImageURL.URL != expectedURL {
		t.Errorf("期望图像 URL '%s', 得到 '%s'", expectedURL, imagePart.ImageURL.URL)
	}
}

// TestConvertCoreRequest_WithToolChoiceAny 测试工具选择为 "any"
func TestConvertCoreRequest_WithToolChoiceAny(t *testing.T) {
	anthropicReq := &anthropicTypes.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []anthropicTypes.InputMessage{
			{
				Role:    "user",
				Content: "测试",
			},
		},
		Tools: []anthropicTypes.Tool{
			{
				Name: "test_tool",
				InputSchema: anthropicTypes.InputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"param": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
		ToolChoice: anthropicTypes.ToolChoiceAny{
			Type: "any",
		},
	}

	// 执行转换
	coreReq := anthropicReq.ConvertCoreRequest()

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查工具选择
	if coreReq.ToolChoice.Mode == nil {
		t.Fatal("期望设置工具选择模式")
	}
	if *coreReq.ToolChoice.Mode != "auto" {
		t.Errorf("期望工具选择模式 'auto'（'any' 映射为 'auto'），得到 '%s'", *coreReq.ToolChoice.Mode)
	}
}

// TestConvertCoreRequest_WithToolChoiceAsMap 测试工具选择为 map 格式
func TestConvertCoreRequest_WithToolChoiceAsMap(t *testing.T) {
	// 准备测试数据：工具选择为 map 格式（模拟 JSON 解析结果）
	toolChoiceMap := map[string]interface{}{
		"type": "tool",
		"name": "get_weather",
	}

	anthropicReq := &anthropicTypes.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []anthropicTypes.InputMessage{
			{
				Role:    "user",
				Content: "测试",
			},
		},
		Tools: []anthropicTypes.Tool{
			{
				Name: "get_weather",
				InputSchema: anthropicTypes.InputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"city": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
		ToolChoice: toolChoiceMap,
	}

	// 执行转换
	coreReq := anthropicReq.ConvertCoreRequest()

	// 验证结果
	if coreReq == nil {
		t.Fatal("ConvertCoreRequest 返回 nil")
	}

	// 检查工具选择
	if coreReq.ToolChoice.Function == nil {
		t.Fatal("期望设置特定工具选择")
	}
	if coreReq.ToolChoice.Function.Function.Name != "get_weather" {
		t.Errorf("期望工具名称 'get_weather', 得到 '%s'", coreReq.ToolChoice.Function.Function.Name)
	}
}
