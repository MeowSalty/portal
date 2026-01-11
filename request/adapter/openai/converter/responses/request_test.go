package responses_test

import (
	"testing"

	openaiResponsesConverter "github.com/MeowSalty/portal/request/adapter/openai/converter/responses"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
)

func TestConvertResponsesRequest_Prompt(t *testing.T) {
	prompt := "Say hello"
	coreReq := &coreTypes.Request{Prompt: &prompt}
	channel := &routing.Channel{ModelName: "gpt-4o-mini"}

	result := openaiResponsesConverter.ConvertRequest(coreReq, channel)
	respReq, ok := result.(*openaiResponses.Request)
	if !ok {
		t.Fatalf("期望返回 ResponsesRequest，实际为 %T", result)
	}

	if respReq.Model != "gpt-4o-mini" {
		t.Errorf("期望模型为 gpt-4o-mini，实际为 %s", respReq.Model)
	}

	input, ok := respReq.Input.(string)
	if !ok {
		t.Fatalf("期望 input 为 string，实际为 %T", respReq.Input)
	}

	if input != "Say hello" {
		t.Errorf("期望 input 为 Say hello，实际为 %s", input)
	}
}

func TestConvertResponsesRequest_Messages(t *testing.T) {
	coreReq := &coreTypes.Request{
		Messages: []coreTypes.Message{
			{
				Role: "user",
				Content: coreTypes.MessageContent{
					StringValue: func() *string { s := "Hello"; return &s }(),
				},
			},
		},
		Stream:    func() *bool { b := true; return &b }(),
		MaxTokens: func() *int { v := 42; return &v }(),
	}
	channel := &routing.Channel{ModelName: "gpt-4o-mini"}

	result := openaiResponsesConverter.ConvertRequest(coreReq, channel)
	respReq, ok := result.(*openaiResponses.Request)
	if !ok {
		t.Fatalf("期望返回 ResponsesRequest，实际为 %T", result)
	}

	if respReq.Stream == nil || !*respReq.Stream {
		t.Fatal("期望 stream 为 true")
	}

	if respReq.MaxOutputTokens == nil || *respReq.MaxOutputTokens != 42 {
		t.Fatalf("期望 max_output_tokens 为 42")
	}

	items, ok := respReq.Input.([]openaiResponses.InputItem)
	if !ok {
		t.Fatalf("期望 input 为 []ResponsesInputItem，实际为 %T", respReq.Input)
	}

	if len(items) != 1 || len(items[0].Content) != 1 {
		t.Fatalf("期望 1 条 input message")
	}

	if items[0].Content[0].Type != "input_text" || items[0].Content[0].Text != "Hello" {
		t.Fatalf("input_text 映射不正确")
	}
}
