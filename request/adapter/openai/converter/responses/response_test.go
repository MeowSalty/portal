package responses_test

import (
	"testing"

	openaiResponsesConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/responses"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
)

func TestConvertResponsesCoreResponse_TextOutput(t *testing.T) {
	resp := &openaiResponses.Response{
		ID:        "resp_123",
		Object:    "response",
		CreatedAt: 1700000100,
		Model:     "gpt-4o-mini",
		Output: []openaiResponses.OutputItem{
			{
				ID:   "msg_1",
				Type: "message",
				Role: "assistant",
				Content: []openaiResponses.OutputPart{
					{
						Type: "output_text",
						Text: "Hello",
						Annotations: []interface{}{
							map[string]interface{}{"type": "source"},
						},
					},
					{
						Type: "output_text",
						Text: " world",
					},
				},
			},
		},
		Usage: &openaiResponses.Usage{
			InputTokens:  3,
			OutputTokens: 4,
			TotalTokens:  7,
		},
	}

	result := openaiResponsesConverter.ConvertCoreResponse(resp)
	if result == nil {
		t.Fatal("期望返回非空响应")
	}

	if result.ID != "resp_123" {
		t.Errorf("期望 ID 为 resp_123，实际为 %s", result.ID)
	}

	if result.Usage == nil {
		t.Fatal("期望 Usage 不为空")
	}

	if result.Usage.PromptTokens != 3 || result.Usage.CompletionTokens != 4 || result.Usage.TotalTokens != 7 {
		t.Errorf("Usage 映射不正确")
	}

	if len(result.Choices) != 1 {
		t.Fatalf("期望 Choices 长度为 1，实际为 %d", len(result.Choices))
	}

	if result.Choices[0].Message == nil || result.Choices[0].Message.Content == nil {
		t.Fatal("期望 Message 与 Content 不为空")
	}

	if *result.Choices[0].Message.Content != "Hello world" {
		t.Errorf("期望内容为 Hello world，实际为 %s", *result.Choices[0].Message.Content)
	}

	if result.Choices[0].Message.ID == nil || *result.Choices[0].Message.ID != "msg_1" {
		t.Fatalf("期望 Message ID 为 msg_1")
	}

	if len(result.Choices[0].Message.ContentParts) != 2 {
		t.Fatalf("期望 ContentParts 长度为 2，实际为 %d", len(result.Choices[0].Message.ContentParts))
	}

	if len(result.Choices[0].Message.ContentParts[0].Annotations) != 1 {
		t.Fatalf("期望第一个 ContentPart 的 annotations 不为空")
	}
}

func TestConvertResponsesResponse_WithContentParts(t *testing.T) {
	coreResponse := openaiResponsesConverter.ConvertCoreResponse(&openaiResponses.Response{
		ID:        "resp_456",
		Object:    "response",
		CreatedAt: 1700000300,
		Model:     "gpt-4o-mini",
		Output: []openaiResponses.OutputItem{
			{
				ID:   "msg_9",
				Type: "message",
				Role: "assistant",
				Content: []openaiResponses.OutputPart{
					{
						Type: "output_text",
						Text: "Hi",
						Annotations: []interface{}{
							map[string]interface{}{"type": "note"},
						},
					},
				},
			},
		},
	})

	if coreResponse == nil || len(coreResponse.Choices) == 0 || coreResponse.Choices[0].Message == nil {
		t.Fatal("期望转换出 Message")
	}

	converted := openaiResponsesConverter.ConvertResponse(coreResponse)
	if converted == nil || len(converted.Output) == 0 {
		t.Fatal("期望转换出 Output")
	}

	if converted.Output[0].ID != "msg_9" {
		t.Fatalf("期望 Output ID 为 msg_9，实际为 %s", converted.Output[0].ID)
	}

	if len(converted.Output[0].Content) != 1 {
		t.Fatalf("期望 Output Content 长度为 1，实际为 %d", len(converted.Output[0].Content))
	}

	if len(converted.Output[0].Content[0].Annotations) != 1 {
		t.Fatalf("期望 Output Content annotations 不为空")
	}
}
