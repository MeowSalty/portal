package converter

import (
	"testing"

	"github.com/MeowSalty/portal/logger"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
)

// TestStreamEventToContract_NilEvent 测试 nil 事件的处理
func TestStreamEventToContract_NilEvent(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := adapterTypes.NewStreamIndexContext()

	contracts, err := StreamEventToContract(nil, ctx, log)
	if err != nil {
		t.Fatalf("期望返回 nil, 但得到错误: %v", err)
	}
	if contracts != nil {
		t.Fatal("nil 事件应返回 nil")
	}
}

// TestStreamEventToContract_NilLogger 测试 nil 日志记录器的处理
func TestStreamEventToContract_NilLogger(t *testing.T) {
	ctx := adapterTypes.NewStreamIndexContext()
	event := &geminiTypes.StreamEvent{
		ResponseID: "resp-123",
		Candidates: []geminiTypes.Candidate{
			{
				Index: 0,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{Text: strPtr("Hello")},
					},
				},
			},
		},
	}

	contracts, err := StreamEventToContract(event, ctx, nil)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if len(contracts) == 0 {
		t.Fatal("contracts 不应为空")
	}
}

// TestStreamEventToContract_NoCandidates 测试无候选响应的情况
func TestStreamEventToContract_NoCandidates(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := adapterTypes.NewStreamIndexContext()

	event := &geminiTypes.StreamEvent{
		ResponseID: "resp-123",
		Candidates: []geminiTypes.Candidate{},
	}

	contracts, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if len(contracts) != 1 {
		t.Fatalf("期望 1 个 contract，得到 %d", len(contracts))
	}

	// 验证索引字段补齐
	contract := contracts[0]
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID == "" {
		t.Error("ItemID 应被补齐")
	}
	// EnsureOutputIndex 从 0 开始，所以 0 是合法值
	// 这个断言已移除，因为 EnsureOutputIndex 返回的第一个索引就是 0
}

// TestStreamEventToContract_MessageDelta_WithText 测试包含文本的 message_delta 事件
func TestStreamEventToContract_MessageDelta_WithText(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := adapterTypes.NewStreamIndexContext()

	responseID := "resp-123"
	text := "Hello, world!"

	event := &geminiTypes.StreamEvent{
		ResponseID: responseID,
		Candidates: []geminiTypes.Candidate{
			{
				Index: 0,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{Text: &text},
					},
				},
			},
		},
	}

	contracts, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if len(contracts) != 1 {
		t.Fatalf("期望 1 个 contract，得到 %d", len(contracts))
	}

	contract := contracts[0]

	// 验证索引字段补齐
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID == "" {
		t.Error("ItemID 应被补齐")
	}
	if contract.OutputIndex != 0 {
		t.Errorf("OutputIndex = %d, 期望 0", contract.OutputIndex)
	}
	if contract.ContentIndex != 0 {
		t.Errorf("ContentIndex = %d, 期望 0", contract.ContentIndex)
	}

	// 验证基本字段
	if contract.Type != adapterTypes.StreamEventMessageDelta {
		t.Errorf("Type = %v, 期望 %v", contract.Type, adapterTypes.StreamEventMessageDelta)
	}
	if contract.Source != adapterTypes.StreamSourceGemini {
		t.Errorf("Source = %v, 期望 %v", contract.Source, adapterTypes.StreamSourceGemini)
	}
	if contract.ResponseID != responseID {
		t.Errorf("ResponseID = %v, 期望 %v", contract.ResponseID, responseID)
	}
	if contract.Message == nil {
		t.Fatal("Message 不应为 nil")
	}
	if contract.Message.ContentText == nil || *contract.Message.ContentText != text {
		t.Errorf("ContentText = %v, 期望 %v", contract.Message.ContentText, text)
	}
}

// TestStreamEventToContract_MessageDelta_WithFunctionCall 测试包含函数调用的 message_delta 事件
func TestStreamEventToContract_MessageDelta_WithFunctionCall(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := adapterTypes.NewStreamIndexContext()

	responseID := "resp-123"
	toolID := "tool-123"
	toolName := "get_weather"

	event := &geminiTypes.StreamEvent{
		ResponseID: responseID,
		Candidates: []geminiTypes.Candidate{
			{
				Index: 0,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{
							FunctionCall: &geminiTypes.FunctionCall{
								ID:   &toolID,
								Name: toolName,
								Args: map[string]interface{}{"city": "Beijing"},
							},
						},
					},
				},
			},
		},
	}

	contracts, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if len(contracts) != 1 {
		t.Fatalf("期望 1 个 contract，得到 %d", len(contracts))
	}

	contract := contracts[0]

	// 验证索引字段补齐
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID != toolID {
		t.Errorf("ItemID = %v, 期望 %v（应使用 tool_call.id）", contract.ItemID, toolID)
	}

	// 验证工具调用
	if contract.Message == nil {
		t.Fatal("Message 不应为 nil")
	}
	if len(contract.Message.ToolCalls) != 1 {
		t.Fatalf("ToolCalls 长度 = %d, 期望 1", len(contract.Message.ToolCalls))
	}
	if contract.Message.ToolCalls[0].ID != toolID {
		t.Errorf("ToolCalls[0].ID = %v, 期望 %v", contract.Message.ToolCalls[0].ID, toolID)
	}
	if contract.Message.ToolCalls[0].Name != toolName {
		t.Errorf("ToolCalls[0].Name = %v, 期望 %v", contract.Message.ToolCalls[0].Name, toolName)
	}
}

// TestStreamEventToContract_MessageDelta_WithMultipleParts 测试包含多个 parts 的 message_delta 事件
func TestStreamEventToContract_MessageDelta_WithMultipleParts(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := adapterTypes.NewStreamIndexContext()

	responseID := "resp-123"
	text1 := "Hello"
	text2 := "World"

	event := &geminiTypes.StreamEvent{
		ResponseID: responseID,
		Candidates: []geminiTypes.Candidate{
			{
				Index: 0,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{Text: &text1},
						{Text: &text2},
					},
				},
			},
		},
	}

	contracts, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if len(contracts) != 1 {
		t.Fatalf("期望 1 个 contract，得到 %d", len(contracts))
	}

	contract := contracts[0]

	// 验证索引字段补齐
	if contract.SequenceNumber == 0 {
		t.Error("SequenceNumber 应被补齐")
	}
	if contract.ItemID == "" {
		t.Error("ItemID 应被补齐")
	}
	// content_index 应为最后一个 part 的索引 (1)
	if contract.ContentIndex != 1 {
		t.Errorf("ContentIndex = %d, 期望 1（最后一个 part 的索引）", contract.ContentIndex)
	}

	// 验证 parts
	if contract.Message == nil {
		t.Fatal("Message 不应为 nil")
	}
	if len(contract.Message.Parts) != 2 {
		t.Fatalf("Parts 长度 = %d, 期望 2", len(contract.Message.Parts))
	}
}

// TestStreamEventToContract_MessageStop 测试 message_stop 事件
func TestStreamEventToContract_MessageStop(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := adapterTypes.NewStreamIndexContext()

	responseID := "resp-123"
	finishReason := geminiTypes.FinishReasonStop

	event := &geminiTypes.StreamEvent{
		ResponseID: responseID,
		Candidates: []geminiTypes.Candidate{
			{
				Index:        0,
				FinishReason: finishReason,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{Text: strPtr("Hello")},
					},
				},
			},
		},
	}

	contracts, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 应该返回 2 个事件：message_delta 和 message_stop
	if len(contracts) != 2 {
		t.Fatalf("期望 2 个 contracts，得到 %d", len(contracts))
	}

	// 验证第一个事件（message_delta）
	deltaContract := contracts[0]
	if deltaContract.Type != adapterTypes.StreamEventMessageDelta {
		t.Errorf("第一个事件 Type = %v, 期望 %v", deltaContract.Type, adapterTypes.StreamEventMessageDelta)
	}
	if deltaContract.SequenceNumber == 0 {
		t.Error("第一个事件 SequenceNumber 应被补齐")
	}

	// 验证第二个事件（message_stop）
	stopContract := contracts[1]
	if stopContract.Type != adapterTypes.StreamEventMessageStop {
		t.Errorf("第二个事件 Type = %v, 期望 %v", stopContract.Type, adapterTypes.StreamEventMessageStop)
	}
	if stopContract.SequenceNumber == 0 {
		t.Error("第二个事件 SequenceNumber 应被补齐")
	}
	if stopContract.ItemID == "" {
		t.Error("第二个事件 ItemID 应被补齐")
	}
	if stopContract.OutputIndex != 0 {
		t.Errorf("第二个事件 OutputIndex = %d, 期望 0", stopContract.OutputIndex)
	}

	// 验证 finish_reason
	if stopContract.Content == nil {
		t.Fatal("第二个事件 Content 不应为 nil")
	}
	if stopContract.Content.Raw == nil {
		t.Fatal("第二个事件 Content.Raw 不应为 nil")
	}
	if stopContract.Content.Raw["finish_reason"] != finishReason {
		t.Errorf("finish_reason = %v, 期望 %v", stopContract.Content.Raw["finish_reason"], finishReason)
	}
}

// TestStreamEventToContract_MessageStop_WithFinishMessage 测试带完成消息的 message_stop 事件
func TestStreamEventToContract_MessageStop_WithFinishMessage(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := adapterTypes.NewStreamIndexContext()

	responseID := "resp-123"
	finishReason := geminiTypes.FinishReasonMaxTokens
	finishMessage := "Token limit reached"

	event := &geminiTypes.StreamEvent{
		ResponseID: responseID,
		Candidates: []geminiTypes.Candidate{
			{
				Index:         0,
				FinishReason:  finishReason,
				FinishMessage: finishMessage,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{Text: strPtr("Hello")},
					},
				},
			},
		},
	}

	contracts, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if len(contracts) != 2 {
		t.Fatalf("期望 2 个 contracts，得到 %d", len(contracts))
	}

	// 验证第二个事件（message_stop）的 finish_message
	stopContract := contracts[1]
	if stopContract.Extensions == nil {
		t.Fatal("Extensions 不应为 nil")
	}
	geminiExt, ok := stopContract.Extensions["gemini"].(map[string]interface{})
	if !ok {
		t.Fatal("Extensions[gemini] 不存在或类型错误")
	}
	if geminiExt["finish_message"] != finishMessage {
		t.Errorf("finish_message = %v, 期望 %v", geminiExt["finish_message"], finishMessage)
	}
}

// TestStreamEventToContract_MultipleCandidates 测试多个候选响应
func TestStreamEventToContract_MultipleCandidates(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := adapterTypes.NewStreamIndexContext()

	responseID := "resp-123"

	event := &geminiTypes.StreamEvent{
		ResponseID: responseID,
		Candidates: []geminiTypes.Candidate{
			{
				Index: 0,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{Text: strPtr("Response 1")},
					},
				},
			},
			{
				Index: 1,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{Text: strPtr("Response 2")},
					},
				},
			},
		},
	}

	contracts, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 应该返回 2 个 message_delta 事件
	if len(contracts) != 2 {
		t.Fatalf("期望 2 个 contracts，得到 %d", len(contracts))
	}

	// 验证第一个候选
	if contracts[0].OutputIndex != 0 {
		t.Errorf("第一个候选 OutputIndex = %d, 期望 0", contracts[0].OutputIndex)
	}
	if contracts[0].SequenceNumber == 0 {
		t.Error("第一个候选 SequenceNumber 应被补齐")
	}

	// 验证第二个候选
	if contracts[1].OutputIndex != 1 {
		t.Errorf("第二个候选 OutputIndex = %d, 期望 1", contracts[1].OutputIndex)
	}
	if contracts[1].SequenceNumber == 0 {
		t.Error("第二个候选 SequenceNumber 应被补齐")
	}

	// 验证 sequence_number 递增
	if contracts[1].SequenceNumber <= contracts[0].SequenceNumber {
		t.Errorf("第二个候选 SequenceNumber (%d) 应大于第一个 (%d)",
			contracts[1].SequenceNumber, contracts[0].SequenceNumber)
	}
}

// TestStreamEventToContract_WithUsageMetadata 测试带使用统计的事件
func TestStreamEventToContract_WithUsageMetadata(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := adapterTypes.NewStreamIndexContext()

	responseID := "resp-123"
	inputTokens := int32(100)
	outputTokens := int32(50)

	event := &geminiTypes.StreamEvent{
		ResponseID: responseID,
		Candidates: []geminiTypes.Candidate{
			{
				Index: 0,
				Content: geminiTypes.Content{
					Role: "model",
					Parts: []geminiTypes.Part{
						{Text: strPtr("Hello")},
					},
				},
			},
		},
		UsageMetadata: &geminiTypes.UsageMetadata{
			PromptTokenCount:     inputTokens,
			CandidatesTokenCount: outputTokens,
			TotalTokenCount:      150,
		},
	}

	contracts, err := StreamEventToContract(event, ctx, log)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if len(contracts) != 1 {
		t.Fatalf("期望 1 个 contract，得到 %d", len(contracts))
	}

	contract := contracts[0]

	// 验证使用统计
	if contract.Usage == nil {
		t.Fatal("Usage 不应为 nil")
	}
	if contract.Usage.InputTokens == nil || *contract.Usage.InputTokens != int(inputTokens) {
		t.Errorf("InputTokens = %v, 期望 %d", contract.Usage.InputTokens, inputTokens)
	}
	if contract.Usage.OutputTokens == nil || *contract.Usage.OutputTokens != int(outputTokens) {
		t.Errorf("OutputTokens = %v, 期望 %d", contract.Usage.OutputTokens, outputTokens)
	}
	if contract.Usage.TotalTokens == nil || *contract.Usage.TotalTokens != 150 {
		t.Errorf("TotalTokens = %v, 期望 150", contract.Usage.TotalTokens)
	}
}

// TestStreamEventToContract_IndexConsistency 测试同一流内索引一致性
func TestStreamEventToContract_IndexConsistency(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := adapterTypes.NewStreamIndexContext()

	responseID := "resp-123"

	// 模拟多个事件
	events := []*geminiTypes.StreamEvent{
		{
			ResponseID: responseID,
			Candidates: []geminiTypes.Candidate{
				{
					Index: 0,
					Content: geminiTypes.Content{
						Role: "model",
						Parts: []geminiTypes.Part{
							{Text: strPtr("Hello")},
						},
					},
				},
			},
		},
		{
			ResponseID: responseID,
			Candidates: []geminiTypes.Candidate{
				{
					Index: 0,
					Content: geminiTypes.Content{
						Role: "model",
						Parts: []geminiTypes.Part{
							{Text: strPtr("World")},
						},
					},
				},
			},
		},
	}

	var allContracts []*adapterTypes.StreamEventContract
	for _, event := range events {
		contracts, err := StreamEventToContract(event, ctx, log)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}
		allContracts = append(allContracts, contracts...)
	}

	if len(allContracts) != 2 {
		t.Fatalf("期望 2 个 contracts，得到 %d", len(allContracts))
	}

	// 验证 sequence_number 递增
	if allContracts[1].SequenceNumber <= allContracts[0].SequenceNumber {
		t.Errorf("第二个事件 SequenceNumber (%d) 应大于第一个 (%d)",
			allContracts[1].SequenceNumber, allContracts[0].SequenceNumber)
	}

	// 验证 item_id 在同一流内一致
	if allContracts[0].ItemID != allContracts[1].ItemID {
		t.Errorf("item_id 不一致: %v vs %v", allContracts[0].ItemID, allContracts[1].ItemID)
	}

	// 验证 output_index 一致
	if allContracts[0].OutputIndex != allContracts[1].OutputIndex {
		t.Errorf("output_index 不一致：%d vs %d", allContracts[0].OutputIndex, allContracts[1].OutputIndex)
	}
}

// TestStreamEventToContract_ContentIndexNoRegression 测试 content_index 不发生回退
func TestStreamEventToContract_ContentIndexNoRegression(t *testing.T) {
	log := logger.NewNopLogger()
	ctx := adapterTypes.NewStreamIndexContext()

	responseID := "resp-123"

	// 模拟多个事件，每个事件有不同数量的 parts
	events := []*geminiTypes.StreamEvent{
		{
			ResponseID: responseID,
			Candidates: []geminiTypes.Candidate{
				{
					Index: 0,
					Content: geminiTypes.Content{
						Role: "model",
						Parts: []geminiTypes.Part{
							{Text: strPtr("Part 1")},
							{Text: strPtr("Part 2")},
						},
					},
				},
			},
		},
		{
			ResponseID: responseID,
			Candidates: []geminiTypes.Candidate{
				{
					Index: 0,
					Content: geminiTypes.Content{
						Role: "model",
						Parts: []geminiTypes.Part{
							{Text: strPtr("Part 3")},
						},
					},
				},
			},
		},
	}

	var allContracts []*adapterTypes.StreamEventContract
	for _, event := range events {
		contracts, err := StreamEventToContract(event, ctx, log)
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}
		allContracts = append(allContracts, contracts...)
	}

	if len(allContracts) != 2 {
		t.Fatalf("期望 2 个 contracts，得到 %d", len(allContracts))
	}

	// 第一个事件有 2 个 parts，content_index 应为 1
	if allContracts[0].ContentIndex != 1 {
		t.Errorf("第一个事件 ContentIndex = %d, 期望 1", allContracts[0].ContentIndex)
	}

	// 第二个事件有 1 个 part，但由于 EnsureContentIndex 对同一 item_id 使用缓存值，
	// content_index 应保持为 1（首次缓存值），而不是 0
	// 这符合 "no regression" 设计原则
	if allContracts[1].ContentIndex != 1 {
		t.Errorf("第二个事件 ContentIndex = %d, 期望 1（保持缓存值，不回退）", allContracts[1].ContentIndex)
	}

	// 验证 content_index 在合理范围内（非负）
	for i, contract := range allContracts {
		if contract.ContentIndex < 0 {
			t.Errorf("事件 %d 的 ContentIndex = %d, 应 >= 0", i, contract.ContentIndex)
		}
	}
}

// 辅助函数：返回字符串指针
func strPtr(s string) *string {
	return &s
}
