package responses

import (
	"encoding/json"
	"testing"

	"github.com/MeowSalty/portal/request/adapter/openai/types/shared"
)

// TestResponse_JSONRoundTrip 测试 Response 结构体的 JSON 序列化和反序列化
func TestResponse_JSONRoundTrip(t *testing.T) {
	// 构建测试数据
	status := "completed"
	completedAt := int64(1700000001)
	temperature := 0.7
	topP := 0.9
	maxOutputTokens := 1024
	toolChoiceAuto := "auto"

	original := Response{
		ID:                "resp_123",
		Object:            "response",
		CreatedAt:         1700000000,
		Model:             "gpt-4",
		ParallelToolCalls: true,
		Metadata:          map[string]string{"key": "value"},
		ToolChoice:        &shared.ToolChoiceUnion{Auto: &toolChoiceAuto},
		Tools: []shared.ToolUnion{
			{Function: &shared.ToolFunction{Type: "function", Name: strPtr("test_func")}},
		},
		Output: []OutputItem{
			{
				Message: &OutputMessage{
					Type:   OutputItemTypeMessage,
					ID:     "msg_001",
					Role:   "assistant",
					Status: "completed",
					Content: []OutputMessageContent{
						{
							OutputText: &OutputTextContent{
								Type: OutputMessageContentTypeOutputText,
								Text: "测试文本",
							},
						},
					},
				},
			},
		},
		Status:          &status,
		CompletedAt:     &completedAt,
		Temperature:     &temperature,
		TopP:            &topP,
		MaxOutputTokens: &maxOutputTokens,
		Usage: &Usage{
			InputTokens:  100,
			OutputTokens: 50,
			TotalTokens:  150,
			InputTokensDetails: InputTokensDetails{
				CachedTokens: 10,
			},
			OutputTokensDetails: OutputTokensDetails{
				ReasoningTokens: 5,
			},
		},
	}

	// 序列化为 JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("序列化失败：%v", err)
	}

	// 反序列化回结构体
	var decoded Response
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("反序列化失败：%v", err)
	}

	// 验证必需字段
	if decoded.ID != original.ID {
		t.Errorf("ID 不匹配：得到 %s, 期望 %s", decoded.ID, original.ID)
	}
	if decoded.Object != original.Object {
		t.Errorf("Object 不匹配：得到 %s, 期望 %s", decoded.Object, original.Object)
	}
	if decoded.CreatedAt != original.CreatedAt {
		t.Errorf("CreatedAt 不匹配：得到 %d, 期望 %d", decoded.CreatedAt, original.CreatedAt)
	}
	if decoded.Model != original.Model {
		t.Errorf("Model 不匹配：得到 %s, 期望 %s", decoded.Model, original.Model)
	}
	if decoded.ParallelToolCalls != original.ParallelToolCalls {
		t.Errorf("ParallelToolCalls 不匹配：得到 %v, 期望 %v", decoded.ParallelToolCalls, original.ParallelToolCalls)
	}

	// 验证可选字段
	if decoded.Status == nil || *decoded.Status != *original.Status {
		t.Errorf("Status 不匹配")
	}
	if decoded.CompletedAt == nil || *decoded.CompletedAt != *original.CompletedAt {
		t.Errorf("CompletedAt 不匹配")
	}
	if decoded.Temperature == nil || *decoded.Temperature != *original.Temperature {
		t.Errorf("Temperature 不匹配")
	}

	// 验证 Usage
	if decoded.Usage == nil {
		t.Fatal("Usage 为空")
	}
	if decoded.Usage.InputTokens != original.Usage.InputTokens {
		t.Errorf("InputTokens 不匹配：got %d, 期望 %d", decoded.Usage.InputTokens, original.Usage.InputTokens)
	}
	if decoded.Usage.TotalTokens != original.Usage.TotalTokens {
		t.Errorf("TotalTokens 不匹配：got %d, 期望 %d", decoded.Usage.TotalTokens, original.Usage.TotalTokens)
	}

	// 验证 Output
	if len(decoded.Output) != len(original.Output) {
		t.Fatalf("Output 长度不匹配：got %d, 期望 %d", len(decoded.Output), len(original.Output))
	}

	// 验证 Output 中的 Message 内容
	if decoded.Output[0].Message == nil {
		t.Fatal("Output[0].Message 为空")
	}
	decodedMsg := decoded.Output[0].Message
	if decodedMsg.Type != OutputItemTypeMessage {
		t.Errorf("Output[0].Message.Type 不匹配：得到 %s, 期望 %s", decodedMsg.Type, OutputItemTypeMessage)
	}
	if decodedMsg.ID != "msg_001" {
		t.Errorf("Output[0].Message.ID 不匹配：得到 %s, 期望 msg_001", decodedMsg.ID)
	}
	if decodedMsg.Role != "assistant" {
		t.Errorf("Output[0] Role 不匹配：得到 %s, 期望 assistant", decodedMsg.Role)
	}
	if decodedMsg.Status != "completed" {
		t.Errorf("Output[0] Status 不匹配：得到 %s, 期望 completed", decodedMsg.Status)
	}
	if len(decodedMsg.Content) != 1 {
		t.Fatalf("Output[0] Content 长度不匹配：got %d, 期望 1", len(decodedMsg.Content))
	}

	// 验证 Metadata
	if decoded.Metadata["key"] != "value" {
		t.Errorf("Metadata 不匹配")
	}
}

// TestResponse_FromJSON 测试从 JSON 字符串反序列化
func TestResponse_FromJSON(t *testing.T) {
	jsonStr := `{
		"id": "resp_456",
		"object": "response",
		"created_at": 1700000000,
		"model": "gpt-4o",
		"output": [
			{
				"type": "message",
				"id": "msg_002",
				"role": "assistant",
				"content": [
					{
						"type": "output_text",
						"text": "你好"
					}
				],
				"status": "completed"
			}
		],
		"parallel_tool_calls": false,
		"metadata": {},
		"tool_choice": "auto",
		"tools": [],
		"status": "completed",
		"usage": {
			"input_tokens": 10,
			"input_tokens_details": {"cached_tokens": 0},
			"output_tokens": 20,
			"output_tokens_details": {"reasoning_tokens": 0},
			"total_tokens": 30
		}
	}`

	var resp Response
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("反序列化失败：%v", err)
	}

	if resp.ID != "resp_456" {
		t.Errorf("ID 不匹配：得到 %s, 期望 resp_456", resp.ID)
	}
	if resp.Model != "gpt-4o" {
		t.Errorf("Model 不匹配：得到 %s, 期望 gpt-4o", resp.Model)
	}
	if len(resp.Output) != 1 {
		t.Fatalf("Output 长度不匹配：got %d, 期望 1", len(resp.Output))
	}

	// 验证 Output 中的 Message 内容
	if resp.Output[0].Message == nil {
		t.Fatal("Output[0].Message 为空")
	}
	decodedMsg := resp.Output[0].Message
	if decodedMsg.Type != OutputItemTypeMessage {
		t.Errorf("Output[0].Message.Type 不匹配：得到 %s, 期望 %s", decodedMsg.Type, OutputItemTypeMessage)
	}
	if decodedMsg.ID != "msg_002" {
		t.Errorf("Output[0].Message.ID 不匹配：得到 %s, 期望 msg_002", decodedMsg.ID)
	}
	if decodedMsg.Role != "assistant" {
		t.Errorf("Output[0] Role 不匹配：得到 %s, 期望 assistant", decodedMsg.Role)
	}
	if decodedMsg.Status != "completed" {
		t.Errorf("Output[0] Status 不匹配：得到 %s, 期望 completed", decodedMsg.Status)
	}
	if len(decodedMsg.Content) != 1 {
		t.Fatalf("Output[0] Content 长度不匹配：got %d, 期望 1", len(decodedMsg.Content))
	}

	if resp.Usage == nil || resp.Usage.TotalTokens != 30 {
		t.Errorf("Usage.TotalTokens 不匹配")
	}
}

// TestResponseError_JSONRoundTrip 测试 ResponseError 的 JSON 转换
func TestResponseError_JSONRoundTrip(t *testing.T) {
	errType := "invalid_request_error"
	param := "model"

	original := ResponseError{
		Code:    "invalid_model",
		Message: "模型不存在",
		Type:    &errType,
		Param:   &param,
	}

	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("序列化失败：%v", err)
	}

	var decoded ResponseError
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("反序列化失败：%v", err)
	}

	if decoded.Code != original.Code {
		t.Errorf("Code 不匹配：得到 %s, 期望 %s", decoded.Code, original.Code)
	}
	if decoded.Message != original.Message {
		t.Errorf("Message 不匹配：得到 %s, 期望 %s", decoded.Message, original.Message)
	}
	if decoded.Type == nil || *decoded.Type != *original.Type {
		t.Errorf("Type 不匹配")
	}
	if decoded.Param == nil || *decoded.Param != *original.Param {
		t.Errorf("Param 不匹配")
	}
}

// TestUsage_JSONRoundTrip 测试 Usage 的 JSON 转换
func TestUsage_JSONRoundTrip(t *testing.T) {
	original := Usage{
		InputTokens:  100,
		OutputTokens: 200,
		TotalTokens:  300,
		InputTokensDetails: InputTokensDetails{
			CachedTokens: 50,
		},
		OutputTokensDetails: OutputTokensDetails{
			ReasoningTokens: 25,
		},
	}

	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("序列化失败：%v", err)
	}

	var decoded Usage
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("反序列化失败：%v", err)
	}

	if decoded.InputTokens != original.InputTokens {
		t.Errorf("InputTokens 不匹配")
	}
	if decoded.OutputTokens != original.OutputTokens {
		t.Errorf("OutputTokens 不匹配")
	}
	if decoded.TotalTokens != original.TotalTokens {
		t.Errorf("TotalTokens 不匹配")
	}
	if decoded.InputTokensDetails.CachedTokens != original.InputTokensDetails.CachedTokens {
		t.Errorf("CachedTokens 不匹配")
	}
	if decoded.OutputTokensDetails.ReasoningTokens != original.OutputTokensDetails.ReasoningTokens {
		t.Errorf("ReasoningTokens 不匹配")
	}
}

func strPtr(s string) *string {
	return &s
}
