package responses_test

import (
	"encoding/json"
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

	result := openaiResponsesConverter.ConvertResponsesCoreResponse(resp)
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
}

func TestConvertResponsesStreamEvent_Delta(t *testing.T) {
	jsonData := `{"type":"response.output_text.delta","delta":"Hi"}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertResponsesStreamEvent(&event)
	if resp == nil {
		t.Fatal("期望返回非空响应")
	}

	if len(resp.Choices) != 1 || resp.Choices[0].Delta == nil || resp.Choices[0].Delta.Content == nil {
		t.Fatal("期望返回 Delta 内容")
	}

	if *resp.Choices[0].Delta.Content != "Hi" {
		t.Errorf("期望 Delta 内容为 Hi，实际为 %s", *resp.Choices[0].Delta.Content)
	}
}

func TestConvertResponsesStreamEvent_CompletedUsageOnly(t *testing.T) {
	jsonData := `{"type":"response.completed","response":{"id":"resp_999","object":"response","created_at":1700000200,"model":"gpt-4o-mini","usage":{"input_tokens":1,"output_tokens":2,"total_tokens":3}}}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertResponsesStreamEvent(&event)
	if resp == nil {
		t.Fatal("期望返回非空响应")
	}

	if resp.Usage == nil {
		t.Fatal("期望 Usage 不为空")
	}

	if resp.Usage.PromptTokens != 1 || resp.Usage.CompletionTokens != 2 || resp.Usage.TotalTokens != 3 {
		t.Errorf("Usage 映射不正确")
	}

	if len(resp.Choices) != 0 {
		t.Errorf("completed 事件不应包含 choices，实际为 %d", len(resp.Choices))
	}
}

func TestConvertResponsesStreamEvent_Error(t *testing.T) {
	jsonData := `{"type":"error","error":{"message":"bad request","type":"invalid_request_error"}}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertResponsesStreamEvent(&event)
	if resp == nil {
		t.Fatal("期望返回非空响应")
	}

	if len(resp.Choices) != 1 || resp.Choices[0].Error == nil {
		t.Fatal("期望返回 Error")
	}

	if resp.Choices[0].Error.Message != "bad request" {
		t.Errorf("期望错误消息为 bad request，实际为 %s", resp.Choices[0].Error.Message)
	}
}
