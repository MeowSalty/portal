package responses_test

import (
	"encoding/json"
	"testing"

	openaiResponsesConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/responses"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	coreTypes "github.com/MeowSalty/portal/types"
)

func TestConvertResponsesStreamEvent_Delta(t *testing.T) {
	jsonData := `{"type":"response.output_text.delta","delta":"Hi","response_id":"resp_1","item_id":"msg_1","output_index":0,"content_index":0}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertStreamEvent(&event)
	if resp == nil {
		t.Fatal("期望返回非空响应")
	}

	if len(resp.Choices) != 1 || resp.Choices[0].Delta == nil || resp.Choices[0].Delta.Content == nil {
		t.Fatal("期望返回 Delta 内容")
	}

	if *resp.Choices[0].Delta.Content != "Hi" {
		t.Errorf("期望 Delta 内容为 Hi，实际为 %s", *resp.Choices[0].Delta.Content)
	}

	if resp.Choices[0].Delta.ExtraFieldsFormat != "openai" {
		t.Fatalf("期望 ExtraFieldsFormat 为 openai")
	}

	if resp.Choices[0].Delta.ExtraFields == nil {
		t.Fatal("期望 ExtraFields 不为空")
	}

	if resp.Choices[0].Delta.ExtraFields["item_id"] != "msg_1" {
		t.Fatalf("期望 item_id 为 msg_1")
	}
}

func TestConvertResponsesStreamEvent_CompletedUsageOnly(t *testing.T) {
	jsonData := `{"type":"response.completed","response":{"id":"resp_999","object":"response","created_at":1700000200,"model":"gpt-4o-mini","usage":{"input_tokens":1,"output_tokens":2,"total_tokens":3}}}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertStreamEvent(&event)
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

func TestConvertResponsesStreamEvent_OutputItemAdded(t *testing.T) {
	jsonData := `{"type":"response.output_item.added","output_index":0,"item":{"id":"msg_1","type":"message","role":"assistant","content":[]}}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertStreamEvent(&event)
	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
		t.Fatal("期望返回 Message")
	}

	if resp.Choices[0].Message.ID == nil || *resp.Choices[0].Message.ID != "msg_1" {
		t.Fatalf("期望 Message ID 为 msg_1")
	}
}

func TestConvertResponsesStreamEvent_OutputTextDone(t *testing.T) {
	jsonData := `{"type":"response.output_text.done","text":"Hello","item_id":"msg_1","output_index":0,"content_index":0}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertStreamEvent(&event)
	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
		t.Fatal("期望返回 Message")
	}

	if resp.Choices[0].Message.Content == nil || *resp.Choices[0].Message.Content != "Hello" {
		t.Fatalf("期望内容为 Hello")
	}

	if resp.Choices[0].Message.ExtraFields["output_index"] != 0 {
		t.Fatalf("期望 output_index 为 0")
	}
}

func TestConvertResponsesStreamEvent_OutputTextAnnotationAdded(t *testing.T) {
	jsonData := `{"type":"response.output_text.annotation.added","annotation":{"type":"note"},"annotation_index":0,"item_id":"msg_1","output_index":0,"content_index":0}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertStreamEvent(&event)
	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
		t.Fatal("期望返回 Message")
	}

	if len(resp.Choices[0].Message.ContentParts) != 1 || len(resp.Choices[0].Message.ContentParts[0].Annotations) != 1 {
		t.Fatalf("期望 annotations 被写入")
	}
}

func TestConvertResponsesStreamEvent_Error(t *testing.T) {
	jsonData := `{"type":"error","message":"bad request","code":"invalid_request_error","param":"input"}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertStreamEvent(&event)
	if resp == nil {
		t.Fatal("期望返回非空响应")
	}

	if len(resp.Choices) != 1 || resp.Choices[0].Error == nil {
		t.Fatal("期望返回 Error")
	}

	if resp.Choices[0].Error.Message != "bad request" {
		t.Errorf("期望错误消息为 bad request，实际为 %s", resp.Choices[0].Error.Message)
	}

	if resp.Choices[0].Error.Metadata["code"] != "invalid_request_error" {
		t.Fatalf("期望错误 code 为 invalid_request_error")
	}

	if resp.Choices[0].Error.Metadata["param"] != "input" {
		t.Fatalf("期望错误 param 为 input")
	}
}

func TestConvertStreamResponse_OutputItemAdded(t *testing.T) {
	msgID := "msg_1"
	resp := &coreTypes.Response{
		ID: "resp_1",
		Choices: []coreTypes.Choice{
			{
				Message: &coreTypes.ResponseMessage{
					ID:   &msgID,
					Role: "assistant",
				},
			},
		},
	}

	event := openaiResponsesConverter.ConvertStreamResponse(resp, &openaiResponsesConverter.StreamEventMeta{
		SequenceNumber: 1,
		OutputIndex:    0,
		ContentIndex:   0,
	})
	if event == nil {
		t.Fatal("期望返回事件")
	}
	if event.Type != "response.output_item.added" {
		t.Fatalf("期望事件类型为 response.output_item.added，实际为 %s", event.Type)
	}
	if event.Item == nil || event.Item.ID != "msg_1" {
		t.Fatalf("期望 item_id 为 msg_1")
	}
}

func TestConvertStreamResponse_OutputTextDone(t *testing.T) {
	msgID := "msg_1"
	text := "Hello"
	resp := &coreTypes.Response{
		ID: "resp_1",
		Choices: []coreTypes.Choice{
			{
				Message: &coreTypes.ResponseMessage{
					ID:      &msgID,
					Role:    "assistant",
					Content: &text,
				},
			},
		},
	}

	event := openaiResponsesConverter.ConvertStreamResponse(resp, &openaiResponsesConverter.StreamEventMeta{
		SequenceNumber: 1,
		OutputIndex:    0,
		ContentIndex:   0,
	})
	if event == nil {
		t.Fatal("期望返回事件")
	}
	if event.Type != "response.output_text.done" {
		t.Fatalf("期望事件类型为 response.output_text.done，实际为 %s", event.Type)
	}
	if event.Text != "Hello" {
		t.Fatalf("期望 text 为 Hello")
	}
	if event.ItemID != "msg_1" {
		t.Fatalf("期望 item_id 为 msg_1")
	}
}

func TestConvertStreamResponse_CompletedWithUsage(t *testing.T) {
	resp := &coreTypes.Response{
		ID: "resp_1",
		Usage: &coreTypes.ResponseUsage{
			PromptTokens:     1,
			CompletionTokens: 2,
			TotalTokens:      3,
		},
	}

	event := openaiResponsesConverter.ConvertStreamResponse(resp, &openaiResponsesConverter.StreamEventMeta{
		SequenceNumber: 1,
	})
	if event == nil {
		t.Fatal("期望返回事件")
	}
	if event.Type != "response.completed" {
		t.Fatalf("期望事件类型为 response.completed，实际为 %s", event.Type)
	}
	if event.Response == nil || event.Response.Usage == nil {
		t.Fatalf("期望包含 usage")
	}
}

func TestConvertResponsesStreamEvent_InProgress(t *testing.T) {
	jsonData := `{"type":"response.in_progress","response":{"id":"resp_1","object":"response","created_at":1700000100,"model":"gpt-4o","status":"in_progress"}}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertStreamEvent(&event)
	if resp == nil {
		t.Fatal("期望返回非空响应")
	}

	if resp.ID != "resp_1" {
		t.Fatalf("期望 ID 为 resp_1，实际为 %s", resp.ID)
	}

	if len(resp.Choices) != 1 || resp.Choices[0].Message == nil {
		t.Fatal("期望返回 Message")
	}

	if resp.Choices[0].Message.ExtraFields["status"] != "in_progress" {
		t.Fatalf("期望 status 为 in_progress")
	}
}

func TestConvertResponsesStreamEvent_OutputItemDone(t *testing.T) {
	jsonData := `{"type":"response.output_item.done","output_index":0,"item":{"id":"msg_1","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"Hello"}]}}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertStreamEvent(&event)
	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
		t.Fatal("期望返回 Message")
	}

	if resp.Choices[0].Message.ID == nil || *resp.Choices[0].Message.ID != "msg_1" {
		t.Fatalf("期望 Message ID 为 msg_1")
	}

	if resp.Choices[0].Message.ExtraFields["status"] != "completed" {
		t.Fatalf("期望 status 为 completed")
	}

	if resp.Choices[0].Message.Content == nil || *resp.Choices[0].Message.Content != "Hello" {
		t.Fatalf("期望内容为 Hello")
	}
}

func TestConvertResponsesStreamEvent_ContentPartAdded(t *testing.T) {
	jsonData := `{"type":"response.content_part.added","item_id":"msg_1","output_index":0,"content_index":0,"part":{"type":"output_text","text":""}}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertStreamEvent(&event)
	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
		t.Fatal("期望返回 Message")
	}

	if len(resp.Choices[0].Message.ContentParts) != 1 {
		t.Fatalf("期望 ContentParts 长度为 1")
	}

	if resp.Choices[0].Message.ContentParts[0].Type != "output_text" {
		t.Fatalf("期望 ContentPart 类型为 output_text")
	}
}

func TestConvertResponsesStreamEvent_ContentPartDone(t *testing.T) {
	jsonData := `{"type":"response.content_part.done","item_id":"msg_1","output_index":0,"content_index":0,"part":{"type":"output_text","text":"Hello world","annotations":[]}}`
	var event openaiResponses.ResponsesStreamEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	resp := openaiResponsesConverter.ConvertStreamEvent(&event)
	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
		t.Fatal("期望返回 Message")
	}

	if resp.Choices[0].Message.ExtraFields["status"] != "completed" {
		t.Fatalf("期望 status 为 completed")
	}

	if resp.Choices[0].Message.Content == nil || *resp.Choices[0].Message.Content != "Hello world" {
		t.Fatalf("期望内容为 Hello world")
	}
}
