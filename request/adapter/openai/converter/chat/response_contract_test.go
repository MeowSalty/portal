package chat

import (
	"testing"

	"github.com/MeowSalty/portal/logger"
	chatTypes "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// TestResponseToContract_BasicConversion 测试基本响应的转换
func TestResponseToContract_BasicConversion(t *testing.T) {
	log := logger.NewNopLogger()

	// 准备测试数据
	model := "gpt-4"
	text := "Hello, world!"

	resp := &chatTypes.Response{
		ID:      "chatcmpl_123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   model,
		Choices: []chatTypes.Choice{
			{
				Index:        0,
				FinishReason: chatTypes.FinishReasonStop,
				Message: chatTypes.Message{
					Role:    chatTypes.ChatResponseMessageRoleAssistant,
					Content: &text,
				},
			},
		},
		Usage: &chatTypes.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	// 执行转换
	contract, err := ResponseToContract(resp, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 验证结果
	if contract == nil {
		t.Fatal("contract 不应为 nil")
	}
	if contract.Source != types.VendorSourceOpenAIChat {
		t.Errorf("Source = %v, 期望 %v", contract.Source, types.VendorSourceOpenAIChat)
	}
	if contract.ID != "chatcmpl_123" {
		t.Errorf("ID = %v, 期望 chatcmpl_123", contract.ID)
	}
	if contract.Model == nil || *contract.Model != model {
		t.Errorf("Model = %v, 期望 %v", contract.Model, model)
	}
	if len(contract.Choices) != 1 {
		t.Fatalf("Choices 长度 = %d, 期望 1", len(contract.Choices))
	}
	if contract.Choices[0].Message == nil {
		t.Fatal("Message 不应为 nil")
	}
	if contract.Choices[0].Message.Content == nil || *contract.Choices[0].Message.Content != text {
		t.Errorf("Content = %v, 期望 %v", contract.Choices[0].Message.Content, text)
	}
}

// TestResponseToContract_FinishReasonMapping 测试完成原因的映射
func TestResponseToContract_FinishReasonMapping(t *testing.T) {
	log := logger.NewNopLogger()

	tests := []struct {
		name           string
		finishReason   chatTypes.FinishReason
		expectedReason types.ResponseFinishReason
	}{
		{"Stop", chatTypes.FinishReasonStop, types.ResponseFinishReasonStop},
		{"Length", chatTypes.FinishReasonLength, types.ResponseFinishReasonLength},
		{"ToolCalls", chatTypes.FinishReasonToolCalls, types.ResponseFinishReasonToolCalls},
		{"ContentFilter", chatTypes.FinishReasonContentFilter, types.ResponseFinishReasonContentFilter},
		{"FunctionCall", chatTypes.FinishReasonFunctionCall, types.ResponseFinishReasonToolCalls},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := "test"
			resp := &chatTypes.Response{
				ID:      "chatcmpl_test",
				Object:  "chat.completion",
				Created: 1234567890,
				Model:   "gpt-4",
				Choices: []chatTypes.Choice{
					{
						Index:        0,
						FinishReason: tt.finishReason,
						Message: chatTypes.Message{
							Role:    chatTypes.ChatResponseMessageRoleAssistant,
							Content: &text,
						},
					},
				},
			}

			contract, err := ResponseToContract(resp, log)
			if err != nil {
				t.Fatalf("转换失败: %v", err)
			}

			if len(contract.Choices) == 0 {
				t.Fatal("Choices 不应为空")
			}
			if contract.Choices[0].FinishReason == nil {
				t.Fatal("FinishReason 不应为 nil")
			}
			if *contract.Choices[0].FinishReason != tt.expectedReason {
				t.Errorf("FinishReason = %v, 期望 %v", *contract.Choices[0].FinishReason, tt.expectedReason)
			}
		})
	}
}

// TestResponseToContract_ToolCalls 测试工具调用的转换
func TestResponseToContract_ToolCalls(t *testing.T) {
	log := logger.NewNopLogger()

	toolID := "call_123"
	toolName := "get_weather"
	toolArgs := `{"location":"Beijing"}`

	resp := &chatTypes.Response{
		ID:      "chatcmpl_123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []chatTypes.Choice{
			{
				Index:        0,
				FinishReason: chatTypes.FinishReasonToolCalls,
				Message: chatTypes.Message{
					Role: chatTypes.ChatResponseMessageRoleAssistant,
					ToolCalls: []chatTypes.MessageToolCall{
						{
							ID:   toolID,
							Type: chatTypes.ToolCallTypeFunction,
							Function: &chatTypes.ToolCallFunction{
								Name:      toolName,
								Arguments: toolArgs,
							},
						},
					},
				},
			},
		},
	}

	contract, err := ResponseToContract(resp, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if len(contract.Choices) == 0 {
		t.Fatal("Choices 不应为空")
	}
	if contract.Choices[0].Message == nil {
		t.Fatal("Message 不应为 nil")
	}
	if len(contract.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("ToolCalls 长度 = %d, 期望 1", len(contract.Choices[0].Message.ToolCalls))
	}

	toolCall := contract.Choices[0].Message.ToolCalls[0]
	if toolCall.ID == nil || *toolCall.ID != toolID {
		t.Errorf("ToolCall ID = %v, 期望 %v", toolCall.ID, toolID)
	}
	if toolCall.Name == nil || *toolCall.Name != toolName {
		t.Errorf("ToolCall Name = %v, 期望 %v", toolCall.Name, toolName)
	}
	if toolCall.Arguments == nil || *toolCall.Arguments != toolArgs {
		t.Errorf("ToolCall Arguments = %v, 期望 %v", toolCall.Arguments, toolArgs)
	}
}

// TestResponseToContract_Usage 测试使用统计的转换
func TestResponseToContract_Usage(t *testing.T) {
	log := logger.NewNopLogger()

	text := "test"

	resp := &chatTypes.Response{
		ID:      "chatcmpl_123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []chatTypes.Choice{
			{
				Index:        0,
				FinishReason: chatTypes.FinishReasonStop,
				Message: chatTypes.Message{
					Role:    chatTypes.ChatResponseMessageRoleAssistant,
					Content: &text,
				},
			},
		},
		Usage: &chatTypes.Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}

	contract, err := ResponseToContract(resp, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if contract.Usage == nil {
		t.Fatal("Usage 不应为 nil")
	}
	if contract.Usage.InputTokens == nil || *contract.Usage.InputTokens != 100 {
		t.Errorf("InputTokens = %v, 期望 100", contract.Usage.InputTokens)
	}
	if contract.Usage.OutputTokens == nil || *contract.Usage.OutputTokens != 50 {
		t.Errorf("OutputTokens = %v, 期望 50", contract.Usage.OutputTokens)
	}
	if contract.Usage.TotalTokens == nil || *contract.Usage.TotalTokens != 150 {
		t.Errorf("TotalTokens = %v, 期望 150", contract.Usage.TotalTokens)
	}
}

// TestResponseToContract_RoundTrip 测试双向转换的可逆性
func TestResponseToContract_RoundTrip(t *testing.T) {
	log := logger.NewNopLogger()

	// 准备原始响应
	model := "gpt-4"
	text := "Hello, world!"

	original := &chatTypes.Response{
		ID:      "chatcmpl_123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   model,
		Choices: []chatTypes.Choice{
			{
				Index:        0,
				FinishReason: chatTypes.FinishReasonStop,
				Message: chatTypes.Message{
					Role:    chatTypes.ChatResponseMessageRoleAssistant,
					Content: &text,
				},
			},
		},
		Usage: &chatTypes.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	// RequestToContract
	contract, err := ResponseToContract(original, log)
	if err != nil {
		t.Fatalf("ToContract 失败: %v", err)
	}

	// FromContract
	restored, err := ResponseFromContract(contract, log)
	if err != nil {
		t.Fatalf("FromContract 失败: %v", err)
	}

	// 验证关键字段
	if restored.ID != original.ID {
		t.Errorf("ID = %v, 期望 %v", restored.ID, original.ID)
	}
	if restored.Model != original.Model {
		t.Errorf("Model = %v, 期望 %v", restored.Model, original.Model)
	}
	if len(restored.Choices) != len(original.Choices) {
		t.Errorf("Choices 长度 = %d, 期望 %d", len(restored.Choices), len(original.Choices))
	}
	if restored.Usage.PromptTokens != original.Usage.PromptTokens {
		t.Errorf("PromptTokens = %v, 期望 %v", restored.Usage.PromptTokens, original.Usage.PromptTokens)
	}
}

// TestResponseToContract_NilResponse 测试 nil 响应
func TestResponseToContract_NilResponse(t *testing.T) {
	log := logger.NewNopLogger()

	contract, err := ResponseToContract(nil, log)
	if err != nil {
		t.Fatalf("不应返回错误: %v", err)
	}
	if contract != nil {
		t.Error("contract 应为 nil")
	}
}
