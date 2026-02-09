package converter

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

func TestFromContractPromptAndOptions(t *testing.T) {
	prompt := "hi"
	maxTokens := 10
	temperature := 0.7
	stopText := "stop"
	stream := true
	userID := "u1"
	serviceTier := "auto"
	toolMode := "auto"
	parallel := false
	reasonMode := "enabled"
	budget := 7
	desc := "tool desc"

	contract := &types.RequestContract{
		Model:           "test-model",
		Prompt:          &prompt,
		MaxOutputTokens: &maxTokens,
		Temperature:     &temperature,
		Stop: &types.Stop{
			Text: &stopText,
		},
		Stream: &stream,
		Metadata: map[string]interface{}{
			"user_id": userID,
		},
		ServiceTier: &serviceTier,
		Tools: []types.Tool{
			{
				Type: "function",
				Function: &types.Function{
					Name:        "tool",
					Description: &desc,
					Parameters: map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
		ToolChoice: &types.ToolChoice{
			Mode: &toolMode,
		},
		ParallelToolCalls: &parallel,
		Reasoning: &types.Reasoning{
			Mode:   &reasonMode,
			Budget: &budget,
		},
	}

	req, err := RequestFromContract(contract)
	if err != nil {
		t.Fatalf("FromContract 错误: %v", err)
	}
	if req == nil {
		t.Fatalf("FromContract 返回了空请求")
	}
	if req.Model != "test-model" {
		t.Fatalf("意外的模型: %s", req.Model)
	}
	if req.MaxTokens != maxTokens {
		t.Fatalf("意外的最大令牌数：%d", req.MaxTokens)
	}
	if req.Temperature == nil || *req.Temperature != temperature {
		t.Fatalf("意外的温度设置")
	}
	if len(req.Messages) != 1 {
		t.Fatalf("意外的消息长度：%d", len(req.Messages))
	}
	if req.Messages[0].Role != anthropicTypes.RoleUser {
		t.Fatalf("意外的角色: %s", req.Messages[0].Role)
	}
	if req.Messages[0].Content.StringValue == nil || *req.Messages[0].Content.StringValue != prompt {
		t.Fatalf("意外的提示内容")
	}
	if len(req.StopSequences) != 1 || req.StopSequences[0] != stopText {
		t.Fatalf("意外的停止序列")
	}
	if req.Stream == nil || *req.Stream != stream {
		t.Fatalf("意外的流式传输设置")
	}
	if req.Metadata == nil || req.Metadata.UserID == nil || *req.Metadata.UserID != userID {
		t.Fatalf("意外的元数据用户 ID")
	}
	if req.ServiceTier == nil || *req.ServiceTier != anthropicTypes.ServiceTier(serviceTier) {
		t.Fatalf("意外的服务层级")
	}
	if len(req.Tools) != 1 || req.Tools[0].Custom == nil {
		t.Fatalf("意外的工具")
	}
	if req.Tools[0].Custom.InputSchema.Type != anthropicTypes.InputSchemaTypeObject {
		t.Fatalf("意外的工具架构类型")
	}
	if req.ToolChoice == nil || req.ToolChoice.Auto == nil || req.ToolChoice.Auto.DisableParallelToolUse == nil {
		t.Fatalf("意外的工具选择")
	}
	if *req.ToolChoice.Auto.DisableParallelToolUse != true {
		t.Fatalf("意外的禁用并行工具使用设置")
	}
	if req.Thinking == nil || req.Thinking.Enabled == nil || req.Thinking.Enabled.BudgetTokens != budget {
		t.Fatalf("意外的思考配置")
	}
}

func TestFromContractMessageBlocksAndToolCalls(t *testing.T) {
	text := "hello"
	id := "tool-id"
	name := "calc"
	args := "{\"a\":1}"
	cc := &anthropicTypes.CacheControlEphemeral{Type: anthropicTypes.CacheControlTypeEphemeral}
	citations := []anthropicTypes.TextCitationParam{
		{
			CharLocation: &anthropicTypes.CitationCharLocationParam{
				Type:           anthropicTypes.TextCitationTypeCharLocation,
				CitedText:      "a",
				DocumentIndex:  1,
				DocumentTitle:  "doc",
				StartCharIndex: 1,
				EndCharIndex:   2,
			},
		},
	}

	contract := &types.RequestContract{
		Model: "test-model",
		Messages: []types.Message{
			{
				Role: "user",
				Content: types.Content{
					Parts: []types.ContentPart{
						{
							Type: "text",
							Text: &text,
							VendorExtras: map[string]interface{}{
								"cache_control": cc,
								"citations":     citations,
							},
						},
					},
				},
				ToolCalls: []types.ToolCall{
					{
						ID:        &id,
						Name:      &name,
						Arguments: &args,
					},
				},
			},
		},
	}

	req, err := RequestFromContract(contract)
	if err != nil {
		t.Fatalf("FromContract 错误: %v", err)
	}
	if len(req.Messages) != 1 {
		t.Fatalf("意外的消息长度：%d", len(req.Messages))
	}
	blocks := req.Messages[0].Content.Blocks
	if len(blocks) != 2 {
		t.Fatalf("意外的块长度：%d", len(blocks))
	}
	if blocks[0].Text == nil || blocks[0].Text.CacheControl != cc || len(blocks[0].Text.Citations) != 1 {
		t.Fatalf("意外的文本块")
	}
	if blocks[1].ToolUse == nil || blocks[1].ToolUse.ID != id || blocks[1].ToolUse.Name != name {
		t.Fatalf("意外的工具使用块")
	}
	if blocks[1].ToolUse.Input == nil {
		t.Fatalf("意外的工具使用输入")
	}
	if v, ok := blocks[1].ToolUse.Input["a"]; !ok || v != float64(1) {
		t.Fatalf("意外的工具使用输入值")
	}
}

func TestToContractMessageBlocks(t *testing.T) {
	cc := &anthropicTypes.CacheControlEphemeral{Type: anthropicTypes.CacheControlTypeEphemeral}
	input := map[string]interface{}{"x": 1}
	msg := anthropicTypes.Message{
		Role: anthropicTypes.RoleUser,
		Content: anthropicTypes.MessageContentParam{
			Blocks: []anthropicTypes.ContentBlockParam{
				{
					Text: &anthropicTypes.TextBlockParam{
						Type:         anthropicTypes.ContentBlockTypeText,
						Text:         "hi",
						CacheControl: cc,
					},
				},
				{
					Image: &anthropicTypes.ImageBlockParam{
						Type: anthropicTypes.ContentBlockTypeImage,
						Source: anthropicTypes.ImageSource{
							Base64: &anthropicTypes.Base64ImageSource{
								Type:      anthropicTypes.ImageSourceTypeBase64,
								MediaType: anthropicTypes.ImageMediaTypePNG,
								Data:      "abc",
							},
						},
					},
				},
				{
					ToolUse: &anthropicTypes.ToolUseBlockParam{
						Type:  anthropicTypes.ContentBlockTypeToolUse,
						ID:    "id1",
						Name:  "func",
						Input: input,
					},
				},
			},
		},
	}

	req := &anthropicTypes.Request{
		Model:    "model",
		Messages: []anthropicTypes.Message{msg},
	}

	contract, err := RequestToContract(req)
	if err != nil {
		t.Fatalf("ToContract 错误: %v", err)
	}
	if len(contract.Messages) != 1 {
		t.Fatalf("意外的消息长度：%d", len(contract.Messages))
	}
	parts := contract.Messages[0].Content.Parts
	if len(parts) != 2 {
		t.Fatalf("意外的部分长度：%d", len(parts))
	}
	if parts[0].VendorExtras == nil || parts[0].VendorExtras["cache_control"] != cc {
		t.Fatalf("意外的文本部分供应商扩展")
	}
	if parts[1].Image == nil || parts[1].Image.Data == nil || *parts[1].Image.Data != "abc" {
		t.Fatalf("意外的图像部分")
	}
	if len(contract.Messages[0].ToolCalls) != 1 {
		t.Fatalf("意外的工具调用长度：%d", len(contract.Messages[0].ToolCalls))
	}
	if contract.Messages[0].ToolCalls[0].Arguments == nil {
		t.Fatalf("意外的工具调用参数")
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(*contract.Messages[0].ToolCalls[0].Arguments), &parsed); err != nil {
		t.Fatalf("解析工具调用参数错误: %v", err)
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("序列化输入参数错误: %v", err)
	}
	var expected map[string]interface{}
	if err := json.Unmarshal(inputJSON, &expected); err != nil {
		t.Fatalf("解析输入参数错误: %v", err)
	}
	if !reflect.DeepEqual(parsed, expected) {
		t.Fatalf("意外的工具调用参数")
	}
}

func TestToContractToolsToolChoiceThinking(t *testing.T) {
	desc := "desc"
	cc := &anthropicTypes.CacheControlEphemeral{Type: anthropicTypes.CacheControlTypeEphemeral}
	disableParallel := true
	req := &anthropicTypes.Request{
		Model: "model",
		Tools: []anthropicTypes.ToolUnion{
			{
				Custom: &anthropicTypes.Tool{
					Name:        "tool",
					Description: &desc,
					InputSchema: anthropicTypes.InputSchema{
						Type:       anthropicTypes.InputSchemaTypeObject,
						Properties: map[string]interface{}{"a": "b"},
					},
					CacheControl: cc,
				},
			},
			{
				Bash20250124: &anthropicTypes.ToolBash20250124{
					Name: anthropicTypes.ToolNameBash,
					Type: anthropicTypes.ToolTypeBash20250124,
				},
			},
		},
		ToolChoice: &anthropicTypes.ToolChoiceParam{
			Tool: &anthropicTypes.ToolChoiceTool{
				Type:                   anthropicTypes.ToolChoiceTypeTool,
				Name:                   "tool",
				DisableParallelToolUse: &disableParallel,
			},
		},
		Thinking: &anthropicTypes.ThinkingConfigParam{
			Enabled: &anthropicTypes.ThinkingConfigEnabled{
				Type:         anthropicTypes.ThinkingConfigTypeEnabled,
				BudgetTokens: 99,
			},
		},
	}

	contract, err := RequestToContract(req)
	if err != nil {
		t.Fatalf("ToContract 错误: %v", err)
	}
	if contract.Source != types.VendorSourceAnthropic {
		t.Fatalf("意外的合同源")
	}
	assert.Equal(t, "anthropic", GetVendorSource(contract), "供应商标识应为 anthropic")
	assert.True(t, IsAnthropicVendor(contract), "供应商判断应为 anthropic")
	if len(contract.Tools) != 1 || contract.Tools[0].Function == nil {
		t.Fatalf("意外的工具")
	}
	if contract.Tools[0].VendorExtras == nil {
		t.Fatalf("意外的工具供应商扩展：缺少 VendorExtras")
	}
	var cacheControl anthropicTypes.CacheControlEphemeral
	found, err := GetVendorExtra("cache_control", contract.Tools[0].VendorExtras, &cacheControl)
	if err != nil {
		t.Fatalf("读取工具供应商扩展失败: %v", err)
	}
	if !found || cacheControl.Type != cc.Type || (cacheControl.TTL != nil) != (cc.TTL != nil) {
		t.Fatalf("意外的工具供应商扩展：cache_control=%+v，期望：%+v", cacheControl, cc)
	}
	if contract.VendorExtras == nil {
		t.Fatalf("意外的供应商扩展")
	}
	extras, ok := contract.VendorExtras["tools_extras"].([]map[string]interface{})
	if !ok || len(extras) != 1 {
		t.Fatalf("意外的工具扩展")
	}
	if extras[0]["type"] != "bash_20250124" {
		t.Fatalf("意外的工具扩展类型")
	}
	if contract.ToolChoice == nil || contract.ToolChoice.Mode == nil || *contract.ToolChoice.Mode != "tool" {
		t.Fatalf("意外的工具选择")
	}
	if contract.ParallelToolCalls == nil || *contract.ParallelToolCalls != false {
		t.Fatalf("意外的并行工具调用")
	}
	if contract.Reasoning == nil || contract.Reasoning.Mode == nil || *contract.Reasoning.Mode != "enabled" {
		t.Fatalf("意外的推理模式")
	}
	if contract.Reasoning.Budget == nil || *contract.Reasoning.Budget != 99 {
		t.Fatalf("意外的推理预算")
	}
}
