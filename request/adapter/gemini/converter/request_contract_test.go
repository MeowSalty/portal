package converter

import (
	"testing"

	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	"github.com/MeowSalty/portal/request/adapter/types"
)

func TestToContract_Basic(t *testing.T) {
	text := "Hello, world!"
	req := &geminiTypes.Request{
		Model: "gemini-pro",
		Contents: []geminiTypes.Content{
			{
				Role: "user",
				Parts: []geminiTypes.Part{
					{Text: &text},
				},
			},
		},
	}

	contract, err := RequestToContract(req)
	if err != nil {
		t.Fatalf("ToContract 失败: %v", err)
	}

	if contract.Model != "gemini-pro" {
		t.Errorf("Model 不匹配: 期望 %s, 实际 %s", "gemini-pro", contract.Model)
	}

	if contract.Source != types.VendorSourceGemini {
		t.Errorf("Source 不匹配: 期望 %s, 实际 %s", types.VendorSourceGemini, contract.Source)
	}

	if len(contract.Messages) != 1 {
		t.Fatalf("Messages 长度不匹配：期望 1, 实际 %d", len(contract.Messages))
	}

	if contract.Messages[0].Role != "user" {
		t.Errorf("Message Role 不匹配: 期望 user, 实际 %s", contract.Messages[0].Role)
	}
}

func TestToContract_WithGenerationConfig(t *testing.T) {
	text := "Test"
	temp := 0.7
	topP := 0.9
	topK := 40
	maxTokens := 1024

	req := &geminiTypes.Request{
		Model: "gemini-pro",
		Contents: []geminiTypes.Content{
			{
				Role:  "user",
				Parts: []geminiTypes.Part{{Text: &text}},
			},
		},
		GenerationConfig: &geminiTypes.GenerationConfig{
			Temperature:     &temp,
			TopP:            &topP,
			TopK:            &topK,
			MaxOutputTokens: &maxTokens,
			StopSequences:   []string{"END"},
		},
	}

	contract, err := RequestToContract(req)
	if err != nil {
		t.Fatalf("ToContract 失败: %v", err)
	}

	if contract.Temperature == nil || *contract.Temperature != temp {
		t.Errorf("Temperature 不匹配")
	}
	if contract.TopP == nil || *contract.TopP != topP {
		t.Errorf("TopP 不匹配")
	}
	if contract.TopK == nil || *contract.TopK != topK {
		t.Errorf("TopK 不匹配")
	}
	if contract.MaxOutputTokens == nil || *contract.MaxOutputTokens != maxTokens {
		t.Errorf("MaxOutputTokens 不匹配")
	}
	if contract.Stop == nil || len(contract.Stop.List) != 1 || contract.Stop.List[0] != "END" {
		t.Errorf("Stop 不匹配")
	}
}

func TestToContract_WithSystemInstruction(t *testing.T) {
	systemText := "You are a helpful assistant."
	userText := "Hello"

	req := &geminiTypes.Request{
		Model: "gemini-pro",
		SystemInstruction: &geminiTypes.Content{
			Parts: []geminiTypes.Part{{Text: &systemText}},
		},
		Contents: []geminiTypes.Content{
			{
				Role:  "user",
				Parts: []geminiTypes.Part{{Text: &userText}},
			},
		},
	}

	contract, err := RequestToContract(req)
	if err != nil {
		t.Fatalf("ToContract 失败: %v", err)
	}

	if contract.System == nil {
		t.Fatal("System 为空")
	}

	if len(contract.System.Parts) != 1 {
		t.Fatalf("System.Parts 长度不匹配：期望 1, 实际 %d", len(contract.System.Parts))
	}

	if contract.System.Parts[0].Text == nil || *contract.System.Parts[0].Text != systemText {
		t.Errorf("System 文本不匹配")
	}
}

func TestToContract_WithTools(t *testing.T) {
	req := &geminiTypes.Request{
		Model: "gemini-pro",
		Tools: []geminiTypes.Tool{
			{
				FunctionDeclarations: []geminiTypes.FunctionDeclaration{
					{
						Name:        "get_weather",
						Description: "获取天气信息",
					},
				},
			},
		},
	}

	contract, err := RequestToContract(req)
	if err != nil {
		t.Fatalf("ToContract 失败: %v", err)
	}

	if len(contract.Tools) != 1 {
		t.Fatalf("Tools 长度不匹配：期望 1, 实际 %d", len(contract.Tools))
	}

	if contract.Tools[0].Type != "function" {
		t.Errorf("Tool Type 不匹配: 期望 function, 实际 %s", contract.Tools[0].Type)
	}

	if contract.Tools[0].Function.Name != "get_weather" {
		t.Errorf("Tool Name 不匹配")
	}
}

func TestFromContract_Basic(t *testing.T) {
	text := "Hello, world!"
	contract := &types.RequestContract{
		Source: types.VendorSourceGemini,
		Model:  "gemini-pro",
		Messages: []types.Message{
			{
				Role: "user",
				Content: types.Content{
					Text: &text,
				},
			},
		},
	}

	req, err := FromContract(contract)
	if err != nil {
		t.Fatalf("FromContract 失败: %v", err)
	}

	if req.Model != "gemini-pro" {
		t.Errorf("Model 不匹配: 期望 %s, 实际 %s", "gemini-pro", req.Model)
	}

	if len(req.Contents) != 1 {
		t.Fatalf("Contents 长度不匹配：期望 1, 实际 %d", len(req.Contents))
	}

	if req.Contents[0].Role != "user" {
		t.Errorf("Content Role 不匹配: 期望 user, 实际 %s", req.Contents[0].Role)
	}
}

func TestFromContract_WithPrompt(t *testing.T) {
	prompt := "Hello, world!"
	contract := &types.RequestContract{
		Source: types.VendorSourceGemini,
		Model:  "gemini-pro",
		Prompt: &prompt,
	}

	req, err := FromContract(contract)
	if err != nil {
		t.Fatalf("FromContract 失败: %v", err)
	}

	if len(req.Contents) != 1 {
		t.Fatalf("Contents 长度不匹配：期望 1, 实际 %d", len(req.Contents))
	}

	if req.Contents[0].Role != "user" {
		t.Errorf("Content Role 不匹配: 期望 user, 实际 %s", req.Contents[0].Role)
	}

	if len(req.Contents[0].Parts) != 1 || req.Contents[0].Parts[0].Text == nil {
		t.Error("Content Parts 不正确")
	}
}

func TestFromContract_WithGenerationConfig(t *testing.T) {
	temp := 0.7
	topP := 0.9
	topK := 40
	maxTokens := 1024

	contract := &types.RequestContract{
		Source:          types.VendorSourceGemini,
		Model:           "gemini-pro",
		Temperature:     &temp,
		TopP:            &topP,
		TopK:            &topK,
		MaxOutputTokens: &maxTokens,
		Stop: &types.Stop{
			List: []string{"END"},
		},
	}

	req, err := FromContract(contract)
	if err != nil {
		t.Fatalf("FromContract 失败: %v", err)
	}

	gc := req.GenerationConfig
	if gc == nil {
		t.Fatal("GenerationConfig 为空")
	}

	if gc.Temperature == nil || *gc.Temperature != temp {
		t.Errorf("Temperature 不匹配")
	}
	if gc.TopP == nil || *gc.TopP != topP {
		t.Errorf("TopP 不匹配")
	}
	if gc.TopK == nil || *gc.TopK != topK {
		t.Errorf("TopK 不匹配")
	}
	if gc.MaxOutputTokens == nil || *gc.MaxOutputTokens != maxTokens {
		t.Errorf("MaxOutputTokens 不匹配")
	}
	if len(gc.StopSequences) != 1 || gc.StopSequences[0] != "END" {
		t.Errorf("StopSequences 不匹配")
	}
}

func TestRoundTrip(t *testing.T) {
	// 测试双向转换的一致性
	text := "Hello, world!"
	systemText := "You are a helpful assistant."
	temp := 0.7
	topK := 40

	original := &geminiTypes.Request{
		Model: "gemini-pro",
		SystemInstruction: &geminiTypes.Content{
			Parts: []geminiTypes.Part{{Text: &systemText}},
		},
		Contents: []geminiTypes.Content{
			{
				Role:  "user",
				Parts: []geminiTypes.Part{{Text: &text}},
			},
		},
		GenerationConfig: &geminiTypes.GenerationConfig{
			Temperature: &temp,
			TopK:        &topK,
		},
	}

	// Gemini -> Contract
	contract, err := RequestToContract(original)
	if err != nil {
		t.Fatalf("ToContract 失败: %v", err)
	}

	// Contract -> Gemini
	restored, err := FromContract(contract)
	if err != nil {
		t.Fatalf("FromContract 失败: %v", err)
	}

	// 验证关键字段
	if restored.Model != original.Model {
		t.Errorf("Model 不匹配")
	}

	if len(restored.Contents) != len(original.Contents) {
		t.Errorf("Contents 长度不匹配")
	}

	if restored.GenerationConfig.Temperature == nil ||
		*restored.GenerationConfig.Temperature != *original.GenerationConfig.Temperature {
		t.Errorf("Temperature 不匹配")
	}

	if restored.GenerationConfig.TopK == nil ||
		*restored.GenerationConfig.TopK != *original.GenerationConfig.TopK {
		t.Errorf("TopK 不匹配")
	}
}

func TestToContract_Nil(t *testing.T) {
	contract, err := RequestToContract(nil)
	if err != nil {
		t.Fatalf("ToContract(nil) 失败: %v", err)
	}
	if contract != nil {
		t.Error("ToContract(nil) 应返回 nil")
	}
}

func TestFromContract_Nil(t *testing.T) {
	req, err := FromContract(nil)
	if err != nil {
		t.Fatalf("FromContract(nil) 失败: %v", err)
	}
	if req != nil {
		t.Error("FromContract(nil) 应返回 nil")
	}
}
