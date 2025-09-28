package converter_test

import (
	"encoding/json"
	"testing"

	"github.com/MeowSalty/portal/adapter/gemini/converter"
	"github.com/MeowSalty/portal/adapter/gemini/types"
	coryType "github.com/MeowSalty/portal/types"
)

func TestConvertCoreResponse_TextMessage(t *testing.T) {
	// 构造一个包含文本消息的 Gemini 响应 JSON
	geminiRespJSON := `{
		"candidates": [
			{
				"content": {
					"parts": [
						{
							"text": "Hello, world!"
						}
					],
					"role": "model"
				},
				"finishReason": "STOP",
				"index": 0
			}
		],
		"usageMetadata": {
			"promptTokenCount": 5,
			"candidatesTokenCount": 10,
			"totalTokenCount": 15
		},
		"modelVersion": "gemini-1.5-pro",
		"responseId": "test-response-id"
	}`

	// 解析 JSON 到 Gemini 响应结构体
	var geminiResp types.Response
	err := json.Unmarshal([]byte(geminiRespJSON), &geminiResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal Gemini response: %v", err)
	}

	// 调用转换函数
	coreResp := converter.ConvertCoreResponse(&geminiResp)

	// 验证转换结果
	if coreResp == nil {
		t.Fatal("Converted response should not be nil")
	}

	// 验证基本字段
	if coreResp.ID != "test-response-id" {
		t.Errorf("Expected ID to be 'test-response-id', got '%s'", coreResp.ID)
	}

	if coreResp.Model != "gemini-1.5-pro" {
		t.Errorf("Expected Model to be 'gemini-1.5-pro', got '%s'", coreResp.Model)
	}

	if coreResp.Object != "chat.completion" {
		t.Errorf("Expected Object to be 'chat.completion', got '%s'", coreResp.Object)
	}

	// 验证使用情况
	if coreResp.Usage == nil {
		t.Fatal("Usage should not be nil")
	}

	if coreResp.Usage.PromptTokens != 5 {
		t.Errorf("Expected PromptTokens to be 5, got %d", coreResp.Usage.PromptTokens)
	}

	if coreResp.Usage.CompletionTokens != 10 {
		t.Errorf("Expected CompletionTokens to be 10, got %d", coreResp.Usage.CompletionTokens)
	}

	if coreResp.Usage.TotalTokens != 15 {
		t.Errorf("Expected TotalTokens to be 15, got %d", coreResp.Usage.TotalTokens)
	}

	// 验证选项
	if len(coreResp.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(coreResp.Choices))
	}

	choice := coreResp.Choices[0]
	if choice.FinishReason == nil {
		t.Fatal("FinishReason should not be nil")
	}

	if *choice.FinishReason != "STOP" {
		t.Errorf("Expected FinishReason to be 'STOP', got '%s'", *choice.FinishReason)
	}

	// 验证消息内容
	if choice.Message == nil {
		t.Fatal("Message should not be nil")
	}

	if choice.Message.Role != "assistant" {
		t.Errorf("Expected Message.Role to be 'assistant', got '%s'", choice.Message.Role)
	}

	if choice.Message.Content == nil {
		t.Fatal("Message.Content should not be nil")
	}

	if *choice.Message.Content != "Hello, world!" {
		t.Errorf("Expected Message.Content to be 'Hello, world!', got '%s'", *choice.Message.Content)
	}
}

func TestConvertResponse_TextMessage(t *testing.T) {
	// 构造一个包含文本消息的核心响应 JSON
	coreRespJSON := `{
		"id": "test-response-id",
		"model": "gemini-1.5-pro",
		"object": "chat.completion",
		"choices": [
			{
				"finish_reason": "STOP",
				"message": {
					"role": "assistant",
					"content": "Hello, world!"
				}
			}
		],
		"usage": {
			"prompt_tokens": 5,
			"completion_tokens": 10,
			"total_tokens": 15
		}
	}`

	// 解析 JSON 到核心响应结构体
	var coreResp coryType.Response

	err := json.Unmarshal([]byte(coreRespJSON), &coreResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal core response: %v", err)
	}

	// 调用转换函数
	geminiResp := converter.ConvertResponse(&coreResp)

	// 验证转换结果
	if geminiResp == nil {
		t.Fatal("Converted response should not be nil")
	}

	// 验证基本字段
	if geminiResp.ResponseID != "test-response-id" {
		t.Errorf("Expected ResponseID to be 'test-response-id', got '%s'", geminiResp.ResponseID)
	}

	if geminiResp.ModelVersion != "gemini-1.5-pro" {
		t.Errorf("Expected ModelVersion to be 'gemini-1.5-pro', got '%s'", geminiResp.ModelVersion)
	}

	// 验证使用情况
	if geminiResp.UsageMetadata == nil {
		t.Fatal("UsageMetadata should not be nil")
	}

	if geminiResp.UsageMetadata.PromptTokenCount != 5 {
		t.Errorf("Expected PromptTokenCount to be 5, got %d", geminiResp.UsageMetadata.PromptTokenCount)
	}

	if geminiResp.UsageMetadata.CandidatesTokenCount != 10 {
		t.Errorf("Expected CandidatesTokenCount to be 10, got %d", geminiResp.UsageMetadata.CandidatesTokenCount)
	}

	if geminiResp.UsageMetadata.TotalTokenCount != 15 {
		t.Errorf("Expected TotalTokenCount to be 15, got %d", geminiResp.UsageMetadata.TotalTokenCount)
	}

	// 验证候选
	if len(geminiResp.Candidates) != 1 {
		t.Fatalf("Expected 1 candidate, got %d", len(geminiResp.Candidates))
	}

	candidate := geminiResp.Candidates[0]
	if candidate.FinishReason != "STOP" {
		t.Errorf("Expected FinishReason to be 'STOP', got '%s'", candidate.FinishReason)
	}

	// 验证消息内容
	if candidate.Content.Role != "model" {
		t.Errorf("Expected Content.Role to be 'model', got '%s'", candidate.Content.Role)
	}

	if len(candidate.Content.Parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(candidate.Content.Parts))
	}

	part := candidate.Content.Parts[0]
	if part.Text == nil {
		t.Fatal("Text part should not be nil")
	}

	if *part.Text != "Hello, world!" {
		t.Errorf("Expected Text to be 'Hello, world!', got '%s'", *part.Text)
	}
}
