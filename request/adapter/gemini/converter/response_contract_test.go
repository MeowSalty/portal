package converter

import (
	"testing"

	"github.com/MeowSalty/portal/logger"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

// TestResponseToContract_BasicConversion 测试基本响应的转换
func TestResponseToContract_BasicConversion(t *testing.T) {
	log := logger.NewNopLogger()

	// 准备测试数据
	model := "gemini-2.0-flash-exp"
	text := "Hello, world!"

	resp := &geminiTypes.Response{
		ResponseID:   "resp_123",
		ModelVersion: model,
		Candidates: []geminiTypes.Candidate{
			{
				Index:        0,
				FinishReason: geminiTypes.FinishReasonStop,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{
							Text: &text,
						},
					},
				},
			},
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
	if contract.Source != types.VendorSourceGemini {
		t.Errorf("Source = %v, 期望 %v", contract.Source, types.VendorSourceGemini)
	}
	if contract.ID != "resp_123" {
		t.Errorf("ID = %v, 期望 resp_123", contract.ID)
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
		finishReason   geminiTypes.FinishReason
		expectedReason types.ResponseFinishReason
	}{
		{"Stop", geminiTypes.FinishReasonStop, types.ResponseFinishReasonStop},
		{"MaxTokens", geminiTypes.FinishReasonMaxTokens, types.ResponseFinishReasonLength},
		{"Safety", geminiTypes.FinishReasonSafety, types.ResponseFinishReasonContentFilter},
		{"Blocklist", geminiTypes.FinishReasonBlocklist, types.ResponseFinishReasonContentFilter},
		{"Recitation", geminiTypes.FinishReasonRecitation, types.ResponseFinishReasonRecitation},
		{"Language", geminiTypes.FinishReasonLanguage, types.ResponseFinishReasonLanguage},
		{"MalformedFunction", geminiTypes.FinishReasonMalformedFunction, types.ResponseFinishReasonToolCallMalformed},
		{"Unknown", "unknown", types.ResponseFinishReasonUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := "test"
			resp := &geminiTypes.Response{
				ResponseID:   "resp_test",
				ModelVersion: "gemini-2.0-flash-exp",
				Candidates: []geminiTypes.Candidate{
					{
						Index:        0,
						FinishReason: tt.finishReason,
						Content: geminiTypes.Content{
							Role: "model",
							Parts: []geminiTypes.Part{
								{
									Text: &text,
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

	toolName := "get_weather"

	resp := &geminiTypes.Response{
		ResponseID:   "resp_123",
		ModelVersion: "gemini-2.0-flash-exp",
		Candidates: []geminiTypes.Candidate{
			{
				Index:        0,
				FinishReason: geminiTypes.FinishReasonStop,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{
							FunctionCall: &geminiTypes.FunctionCall{
								Name: toolName,
								Args: map[string]interface{}{"location": "Beijing"},
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
	if toolCall.Name == nil || *toolCall.Name != toolName {
		t.Errorf("ToolCall Name = %v, 期望 %v", toolCall.Name, toolName)
	}
}

// TestResponseToContract_Usage 测试使用统计的转换
func TestResponseToContract_Usage(t *testing.T) {
	log := logger.NewNopLogger()

	text := "test"

	resp := &geminiTypes.Response{
		ResponseID:   "resp_123",
		ModelVersion: "gemini-2.0-flash-exp",
		UsageMetadata: &geminiTypes.UsageMetadata{
			PromptTokenCount:        100,
			CandidatesTokenCount:    50,
			TotalTokenCount:         150,
			CachedContentTokenCount: 10,
			ToolUsePromptTokenCount: 5,
		},
		Candidates: []geminiTypes.Candidate{
			{
				Index:        0,
				FinishReason: geminiTypes.FinishReasonStop,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{
							Text: &text,
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

	// 验证 Extras 中的字段
	if val, ok := contract.Usage.Extras["gemini.cached_content_token_count"]; !ok || val != int32(10) {
		t.Errorf("cached_content_token_count = %v, 期望 10", val)
	}
	if val, ok := contract.Usage.Extras["gemini.tool_use_prompt_token_count"]; !ok || val != int32(5) {
		t.Errorf("tool_use_prompt_token_count = %v, 期望 5", val)
	}
}

// TestResponseToContract_RoundTrip 测试双向转换的可逆性
func TestResponseToContract_RoundTrip(t *testing.T) {
	log := logger.NewNopLogger()

	// 准备原始响应
	model := "gemini-2.0-flash-exp"
	text := "Hello, world!"

	original := &geminiTypes.Response{
		ResponseID:   "resp_123",
		ModelVersion: model,
		UsageMetadata: &geminiTypes.UsageMetadata{
			PromptTokenCount:     100,
			CandidatesTokenCount: 50,
			TotalTokenCount:      150,
		},
		Candidates: []geminiTypes.Candidate{
			{
				Index:        0,
				FinishReason: geminiTypes.FinishReasonStop,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{
							Text: &text,
						},
					},
				},
			},
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
	if restored.ResponseID != original.ResponseID {
		t.Errorf("ResponseID = %v, 期望 %v", restored.ResponseID, original.ResponseID)
	}
	if restored.ModelVersion != original.ModelVersion {
		t.Errorf("ModelVersion = %v, 期望 %v", restored.ModelVersion, original.ModelVersion)
	}
	if len(restored.Candidates) != len(original.Candidates) {
		t.Errorf("Candidates 长度 = %d, 期望 %d", len(restored.Candidates), len(original.Candidates))
	}
}

// TestResponseFromContract_InvalidSource 测试来源验证
func TestResponseFromContract_InvalidSource(t *testing.T) {
	log := logger.NewNopLogger()

	// 创建非 Gemini 来源的 contract
	contract := &types.ResponseContract{
		Source: types.VendorSourceOpenAIChat,
		ID:     "test_id",
	}

	_, err := ResponseFromContract(contract, log)
	if err == nil {
		t.Fatal("期望返回错误，但没有")
	}
}

// TestResponseToContract_NilResponse 测试 nil 响应
func TestResponseToContract_NilResponse(t *testing.T) {
	log := logger.NewNopLogger()

	_, err := ResponseToContract(nil, log)
	if err == nil {
		t.Fatal("期望返回错误，但没有")
	}
}
