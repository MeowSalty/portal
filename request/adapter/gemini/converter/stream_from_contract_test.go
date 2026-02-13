package converter

import (
	"testing"

	"github.com/MeowSalty/portal/errors"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
)

func TestConvertFromIntermediateStreamEvents(t *testing.T) {

	t.Run("空输入返回 nil", func(t *testing.T) {
		result, err := StreamEventFromContract(nil)
		if err != nil {
			t.Fatalf("期望无错误，得到: %v", err)
		}
		if result != nil {
			t.Fatalf("期望 nil，得到: %v", result)
		}
	})

	t.Run("错误来源返回错误", func(t *testing.T) {
		contract := &adapterTypes.StreamEventContract{
			Source: adapterTypes.StreamSourceAnthropic,
			Type:   adapterTypes.StreamEventMessageDelta,
		}

		_, err := StreamEventFromContract(contract)
		if err == nil {
			t.Fatal("期望返回错误")
		}
	})
}

func TestMessageDeltaToGeminiResponse(t *testing.T) {

	t.Run("基础 message_delta 转换", func(t *testing.T) {
		contentText := "Hello"
		contract := &adapterTypes.StreamEventContract{
			Source:     adapterTypes.StreamSourceGemini,
			Type:       adapterTypes.StreamEventMessageDelta,
			ResponseID: "resp-123",
			Message: &adapterTypes.StreamMessagePayload{
				Role:        "model",
				ContentText: &contentText,
			},
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}

		if event == nil {
			t.Fatal("期望事件不为空")
		}

		if event.ResponseID != "resp-123" {
			t.Errorf("期望 ResponseID 'resp-123'，得到 '%s'", event.ResponseID)
		}

		if len(event.Candidates) != 1 {
			t.Fatalf("期望 1 个候选，得到 %d", len(event.Candidates))
		}

		candidate := event.Candidates[0]
		if candidate.Content.Role != "model" {
			t.Errorf("期望 Role 'model'，得到 '%s'", candidate.Content.Role)
		}

		if len(candidate.Content.Parts) != 1 {
			t.Fatalf("期望 1 个 part，得到 %d", len(candidate.Content.Parts))
		}

		part := candidate.Content.Parts[0]
		if part.Text == nil || *part.Text != "Hello" {
			t.Errorf("期望 Text 'Hello'，得到 '%v'", part.Text)
		}
	})

	t.Run("带 usageMetadata 的 message_delta", func(t *testing.T) {
		inputTokens := 10
		outputTokens := 5
		totalTokens := 15
		contract := &adapterTypes.StreamEventContract{
			Source:     adapterTypes.StreamSourceGemini,
			Type:       adapterTypes.StreamEventMessageDelta,
			ResponseID: "resp-123",
			Usage: &adapterTypes.StreamUsagePayload{
				InputTokens:  &inputTokens,
				OutputTokens: &outputTokens,
				TotalTokens:  &totalTokens,
			},
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}

		if event.UsageMetadata == nil {
			t.Fatal("期望 usageMetadata 不为空")
		}

		if event.UsageMetadata.PromptTokenCount != 10 {
			t.Errorf("期望 PromptTokenCount 10，得到 %d", event.UsageMetadata.PromptTokenCount)
		}

		if event.UsageMetadata.CandidatesTokenCount != 5 {
			t.Errorf("期望 CandidatesTokenCount 5，得到 %d", event.UsageMetadata.CandidatesTokenCount)
		}

		if event.UsageMetadata.TotalTokenCount != 15 {
			t.Errorf("期望 TotalTokenCount 15，得到 %d", event.UsageMetadata.TotalTokenCount)
		}
	})
}

func TestMessageStopToGeminiResponse(t *testing.T) {

	t.Run("message_stop 带 finishReason", func(t *testing.T) {
		finishReason := "STOP"
		contract := &adapterTypes.StreamEventContract{
			Source:     adapterTypes.StreamSourceGemini,
			Type:       adapterTypes.StreamEventMessageStop,
			ResponseID: "resp-123",
			Content: &adapterTypes.StreamContentPayload{
				Raw: map[string]interface{}{
					"finish_reason": finishReason,
				},
			},
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}

		if event == nil {
			t.Fatal("期望事件不为空")
		}

		candidate := event.Candidates[0]
		if string(candidate.FinishReason) != finishReason {
			t.Errorf("期望 FinishReason '%s'，得到 '%s'", finishReason, candidate.FinishReason)
		}
	})

	t.Run("message_stop 带 finishMessage", func(t *testing.T) {
		finishMessage := "Completed successfully"
		contract := &adapterTypes.StreamEventContract{
			Source:     adapterTypes.StreamSourceGemini,
			Type:       adapterTypes.StreamEventMessageStop,
			ResponseID: "resp-123",
			Extensions: map[string]interface{}{
				"gemini": map[string]interface{}{
					"finish_message": finishMessage,
				},
			},
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}

		if event == nil {
			t.Fatal("期望事件不为空")
		}

		candidate := event.Candidates[0]
		if candidate.FinishMessage != finishMessage {
			t.Errorf("期望 FinishMessage '%s'，得到 '%s'", finishMessage, candidate.FinishMessage)
		}
	})

	t.Run("message_stop 带 usage", func(t *testing.T) {
		inputTokens := 10
		contract := &adapterTypes.StreamEventContract{
			Source:     adapterTypes.StreamSourceGemini,
			Type:       adapterTypes.StreamEventMessageStop,
			ResponseID: "resp-123",
			Usage: &adapterTypes.StreamUsagePayload{
				InputTokens: &inputTokens,
			},
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}

		if event.UsageMetadata == nil {
			t.Fatal("期望 usageMetadata 不为空")
		}

		if event.UsageMetadata.PromptTokenCount != 10 {
			t.Errorf("期望 PromptTokenCount 10，得到 %d", event.UsageMetadata.PromptTokenCount)
		}
	})
}

func TestToolCallConversion(t *testing.T) {

	t.Run("functionCall part 转换", func(t *testing.T) {
		toolCallID := "call-123"
		toolName := "get_weather"
		toolArgs := `{"city":"Beijing"}`
		contract := &adapterTypes.StreamEventContract{
			Source:     adapterTypes.StreamSourceGemini,
			Type:       adapterTypes.StreamEventMessageDelta,
			ResponseID: "resp-123",
			Message: &adapterTypes.StreamMessagePayload{
				Role: "model",
				ToolCalls: []adapterTypes.StreamToolCall{
					{
						ID:        toolCallID,
						Name:      toolName,
						Arguments: toolArgs,
					},
				},
			},
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}

		if event == nil {
			t.Fatal("期望事件不为空")
		}

		candidate := event.Candidates[0]
		if len(candidate.Content.Parts) != 1 {
			t.Fatalf("期望 1 个 part，得到 %d", len(candidate.Content.Parts))
		}

		part := candidate.Content.Parts[0]
		if part.FunctionCall == nil {
			t.Fatal("期望 FunctionCall 不为空")
		}

		if part.FunctionCall.ID == nil || *part.FunctionCall.ID != toolCallID {
			t.Errorf("期望 ID '%s'，得到 '%v'", toolCallID, part.FunctionCall.ID)
		}

		if part.FunctionCall.Name != toolName {
			t.Errorf("期望 Name '%s'，得到 '%s'", toolName, part.FunctionCall.Name)
		}

		if part.FunctionCall.Args == nil || part.FunctionCall.Args["city"] != "Beijing" {
			t.Errorf("期望 Args 包含 city=Beijing，得到 '%v'", part.FunctionCall.Args)
		}
	})
}

func TestExtensionsExtraction(t *testing.T) {

	t.Run("提取 promptFeedback", func(t *testing.T) {
		contract := &adapterTypes.StreamEventContract{
			Source:     adapterTypes.StreamSourceGemini,
			Type:       adapterTypes.StreamEventMessageDelta,
			ResponseID: "resp-123",
			Extensions: map[string]interface{}{
				"gemini": map[string]interface{}{
					"prompt_feedback": map[string]interface{}{
						"blockReason": "SAFETY",
					},
				},
			},
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}

		if event.PromptFeedback == nil {
			t.Fatal("期望 PromptFeedback 不为空")
		}

		if event.PromptFeedback.BlockReason != geminiTypes.BlockReasonSafety {
			t.Errorf("期望 BlockReason SAFETY，得到 '%s'", event.PromptFeedback.BlockReason)
		}
	})

	t.Run("提取 safetyRatings", func(t *testing.T) {
		contract := &adapterTypes.StreamEventContract{
			Source:     adapterTypes.StreamSourceGemini,
			Type:       adapterTypes.StreamEventMessageDelta,
			ResponseID: "resp-123",
			Extensions: map[string]interface{}{
				"gemini": map[string]interface{}{
					"safety_ratings": []interface{}{
						map[string]interface{}{
							"category":    "HARM_CATEGORY_SEXUALLY_EXPLICIT",
							"probability": "NEGLIGIBLE",
						},
					},
				},
			},
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}

		if len(event.Candidates) == 0 {
			t.Fatal("期望至少 1 个候选")
		}

		if event.Candidates[0].SafetyRatings == nil {
			t.Fatal("期望 SafetyRatings 不为空")
		}

		if len(event.Candidates[0].SafetyRatings) != 1 {
			t.Fatalf("期望 1 个 safety rating，得到 %d", len(event.Candidates[0].SafetyRatings))
		}
	})
}

func TestUnsupportedEventType(t *testing.T) {

	t.Run("不支持的事件类型被忽略", func(t *testing.T) {
		contract := &adapterTypes.StreamEventContract{
			Source: adapterTypes.StreamSourceGemini,
			Type:   adapterTypes.StreamEventContentBlockStart,
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("期望无错误，得到: %v", err)
		}

		if event != nil {
			t.Fatal("期望 nil，因为不支持该事件类型")
		}
	})
}

func TestErrorEvent(t *testing.T) {

	t.Run("error 事件返回错误", func(t *testing.T) {
		errorMessage := "Invalid request"
		contract := &adapterTypes.StreamEventContract{
			Source: adapterTypes.StreamSourceGemini,
			Type:   adapterTypes.StreamEventError,
			Error: &adapterTypes.StreamErrorPayload{
				Message: errorMessage,
				Type:    "invalid_request_error",
				Code:    "400",
			},
		}

		_, err := StreamEventFromContract(contract)
		if err == nil {
			t.Fatal("期望返回错误")
		}

		if errors.GetMessage(err) != errorMessage {
			t.Errorf("期望错误消息 '%s'，得到 '%s'", errorMessage, err.Error())
		}
	})
}

func TestContentPartsConversion(t *testing.T) {

	t.Run("text part 转换", func(t *testing.T) {
		contract := &adapterTypes.StreamEventContract{
			Source:     adapterTypes.StreamSourceGemini,
			Type:       adapterTypes.StreamEventMessageDelta,
			ResponseID: "resp-123",
			Message: &adapterTypes.StreamMessagePayload{
				Role: "model",
				Parts: []adapterTypes.StreamContentPart{
					{
						Type: "text",
						Text: "Sample text",
					},
				},
			},
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}

		if event == nil {
			t.Fatal("期望事件不为空")
		}

		part := event.Candidates[0].Content.Parts[0]
		if part.Text == nil || *part.Text != "Sample text" {
			t.Errorf("期望文本 'Sample text'，得到 '%v'", part.Text)
		}
	})

	t.Run("inlineData part 转换", func(t *testing.T) {
		contract := &adapterTypes.StreamEventContract{
			Source:     adapterTypes.StreamSourceGemini,
			Type:       adapterTypes.StreamEventMessageDelta,
			ResponseID: "resp-123",
			Message: &adapterTypes.StreamMessagePayload{
				Role: "model",
				Parts: []adapterTypes.StreamContentPart{
					{
						Type: "image",
						Raw: map[string]interface{}{
							"inline_data": map[string]interface{}{
								"mimeType": "image/jpeg",
								"data":     "base64data",
							},
						},
					},
				},
			},
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}

		if event == nil {
			t.Fatal("期望事件不为空")
		}

		part := event.Candidates[0].Content.Parts[0]
		if part.InlineData == nil {
			t.Fatal("期望 InlineData 不为空")
		}

		if part.InlineData.MimeType != "image/jpeg" {
			t.Errorf("期望 MimeType 'image/jpeg'，得到 '%s'", part.InlineData.MimeType)
		}

		if part.InlineData.Data != "base64data" {
			t.Errorf("期望 Data 'base64data'，得到 '%s'", part.InlineData.Data)
		}
	})
}

func TestEmptyContentHandling(t *testing.T) {

	t.Run("空 content 也生成事件", func(t *testing.T) {
		contract := &adapterTypes.StreamEventContract{
			Source:     adapterTypes.StreamSourceGemini,
			Type:       adapterTypes.StreamEventMessageDelta,
			ResponseID: "resp-123",
		}

		event, err := StreamEventFromContract(contract)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}

		if event == nil {
			t.Fatal("期望事件不为空")
		}

		if event.ResponseID != "resp-123" {
			t.Errorf("期望 ResponseID 'resp-123'，得到 '%s'", event.ResponseID)
		}

		// 空 content 时 candidates 可能为空
		// Gemini 官方允许空块
	})
}
