package adapter

import (
	"testing"

	"github.com/MeowSalty/portal/request/adapter/anthropic/types"
	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
)

func TestProviderSignals_CompletionShouldNotImplyValidOutput(t *testing.T) {
	t.Run("openai_chat_finish_only", func(t *testing.T) {
		provider := NewOpenAIProvider()
		finishReason := openaiChat.FinishReasonStop
		event := &openaiChat.StreamEvent{
			Choices: []openaiChat.StreamChoice{{FinishReason: &finishReason}},
		}
		signal := provider.IdentifyStreamEventSignal("chat_completions", event)
		if !signal.IsCompletionSignal {
			t.Fatal("IsCompletionSignal should be true")
		}
		if signal.HasValidOutput {
			t.Fatal("HasValidOutput should be false for finish-only event")
		}
	})

	t.Run("openai_responses_completed_only", func(t *testing.T) {
		provider := NewOpenAIProvider()
		event := &openaiResponses.StreamEvent{Completed: &openaiResponses.ResponseCompletedEvent{}}
		signal := provider.IdentifyStreamEventSignal("responses", event)
		if !signal.IsCompletionSignal {
			t.Fatal("IsCompletionSignal should be true")
		}
		if signal.HasValidOutput {
			t.Fatal("HasValidOutput should be false for completed-only event")
		}
	})

	t.Run("gemini_finish_only", func(t *testing.T) {
		provider := NewGeminiProvider()
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{{FinishReason: geminiTypes.FinishReasonStop}},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Fatal("IsCompletionSignal should be true")
		}
		if signal.HasValidOutput {
			t.Fatal("HasValidOutput should be false for finish-only event")
		}
	})

	t.Run("anthropic_message_stop_only", func(t *testing.T) {
		provider := NewAnthropicProvider()
		event := &anthropicTypes.StreamEvent{MessageStop: &anthropicTypes.MessageStopEvent{}}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Fatal("IsCompletionSignal should be true")
		}
		if signal.HasValidOutput {
			t.Fatal("HasValidOutput should be false for message_stop-only event")
		}
	})
}

// TestOpenAI_IdentifyStreamEventSignal_ChatCompletions 测试 OpenAI Chat Completions 信号识别
func TestOpenAI_IdentifyStreamEventSignal_ChatCompletions(t *testing.T) {
	provider := NewOpenAIProvider()

	t.Run("文本增量事件", func(t *testing.T) {
		content := "Hello"
		event := &openaiChat.StreamEvent{
			Choices: []openaiChat.StreamChoice{
				{
					Delta: openaiChat.Delta{
						Content: &content,
					},
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("chat_completions", event)
		if !signal.HasValidOutput {
			t.Error("HasValidOutput should be true for text delta")
		}
		if signal.IsCompletionSignal {
			t.Error("IsCompletionSignal should be false for text delta")
		}
	})

	t.Run("工具调用增量事件", func(t *testing.T) {
		id := "call_123"
		event := &openaiChat.StreamEvent{
			Choices: []openaiChat.StreamChoice{
				{
					Delta: openaiChat.Delta{
						ToolCalls: []openaiChat.ToolCallChunk{
							{ID: &id},
						},
					},
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("chat_completions", event)
		if !signal.HasValidOutput {
			t.Error("HasValidOutput should be true for tool call delta")
		}
	})

	t.Run("完成信号-stop", func(t *testing.T) {
		finishReason := openaiChat.FinishReasonStop
		event := &openaiChat.StreamEvent{
			Choices: []openaiChat.StreamChoice{
				{
					FinishReason: &finishReason,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("chat_completions", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal should be true for finish_reason")
		}
		if !signal.IsTerminalEvent {
			t.Error("IsTerminalEvent should be true for finish_reason")
		}
		if signal.FinishReason != "stop" {
			t.Errorf("FinishReason = %v, want stop", signal.FinishReason)
		}
	})

	t.Run("完成信号-length", func(t *testing.T) {
		finishReason := openaiChat.FinishReasonLength
		event := &openaiChat.StreamEvent{
			Choices: []openaiChat.StreamChoice{
				{
					FinishReason: &finishReason,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("chat_completions", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal should be true for finish_reason")
		}
		if signal.FinishReason != "length" {
			t.Errorf("FinishReason = %v, want length", signal.FinishReason)
		}
	})

	t.Run("无效事件类型", func(t *testing.T) {
		signal := provider.IdentifyStreamEventSignal("chat_completions", "invalid")
		if signal.IsCompletionSignal || signal.HasValidOutput || signal.IsTerminalEvent {
			t.Error("Signal should be empty for invalid event type")
		}
	})
}

// TestOpenAI_IdentifyStreamEventSignal_Responses 测试 OpenAI Responses API 信号识别
func TestOpenAI_IdentifyStreamEventSignal_Responses(t *testing.T) {
	provider := NewOpenAIProvider()

	t.Run("文本增量事件", func(t *testing.T) {
		event := &openaiResponses.StreamEvent{
			OutputTextDelta: &openaiResponses.ResponseOutputTextDeltaEvent{
				Delta: "Hello",
			},
		}
		signal := provider.IdentifyStreamEventSignal("responses", event)
		if !signal.HasValidOutput {
			t.Error("HasValidOutput should be true for output_text_delta")
		}
	})

	t.Run("完成事件", func(t *testing.T) {
		event := &openaiResponses.StreamEvent{
			Completed: &openaiResponses.ResponseCompletedEvent{},
		}
		signal := provider.IdentifyStreamEventSignal("responses", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal should be true for response.completed")
		}
		if signal.FinishReason != "completed" {
			t.Errorf("FinishReason = %v, want completed", signal.FinishReason)
		}
	})

	t.Run("失败事件", func(t *testing.T) {
		event := &openaiResponses.StreamEvent{
			Failed: &openaiResponses.ResponseFailedEvent{},
		}
		signal := provider.IdentifyStreamEventSignal("responses", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal should be true for response.failed")
		}
		if signal.FinishReason != "failed" {
			t.Errorf("FinishReason = %v, want failed", signal.FinishReason)
		}
	})

	t.Run("推理增量事件", func(t *testing.T) {
		event := &openaiResponses.StreamEvent{
			ReasoningTextDelta: &openaiResponses.ResponseReasoningTextDeltaEvent{
				Delta: "thinking...",
			},
		}
		signal := provider.IdentifyStreamEventSignal("responses", event)
		if !signal.HasValidOutput {
			t.Error("HasValidOutput should be true for reasoning_text_delta")
		}
	})
}

// TestGemini_IdentifyStreamEventSignal 测试 Gemini 信号识别
func TestGemini_IdentifyStreamEventSignal(t *testing.T) {
	provider := NewGeminiProvider()

	t.Run("文本增量事件", func(t *testing.T) {
		text := "Hello"
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					Content: geminiTypes.Content{
						Parts: []geminiTypes.Part{
							{Text: &text},
						},
					},
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.HasValidOutput {
			t.Error("HasValidOutput should be true for text content")
		}
	})

	t.Run("完成信号-STOP", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					FinishReason: geminiTypes.FinishReasonStop,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal should be true for finish_reason STOP")
		}
		if signal.FinishReason != geminiTypes.FinishReasonStop {
			t.Errorf("FinishReason = %v, want %v", signal.FinishReason, geminiTypes.FinishReasonStop)
		}
	})

	t.Run("完成信号-MAX_TOKENS", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					FinishReason: geminiTypes.FinishReasonMaxTokens,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal should be true for finish_reason MAX_TOKENS")
		}
	})

	t.Run("阻止反馈", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			PromptFeedback: &geminiTypes.PromptFeedback{
				BlockReason: geminiTypes.BlockReasonSafety,
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal should be true for block_reason")
		}
	})

	t.Run("函数调用", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					Content: geminiTypes.Content{
						Parts: []geminiTypes.Part{
							{FunctionCall: &geminiTypes.FunctionCall{}},
						},
					},
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.HasValidOutput {
			t.Error("HasValidOutput should be true for function call")
		}
	})

	t.Run("无效事件类型", func(t *testing.T) {
		signal := provider.IdentifyStreamEventSignal("", "invalid")
		if signal.IsCompletionSignal || signal.HasValidOutput || signal.IsTerminalEvent {
			t.Error("Signal should be empty for invalid event type")
		}
	})
}

// TestAnthropic_IdentifyStreamEventSignal 测试 Anthropic 信号识别
func TestAnthropic_IdentifyStreamEventSignal(t *testing.T) {
	provider := NewAnthropicProvider()

	t.Run("message_stop 事件", func(t *testing.T) {
		event := &anthropicTypes.StreamEvent{
			MessageStop: &anthropicTypes.MessageStopEvent{},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal should be true for message_stop")
		}
		if !signal.IsTerminalEvent {
			t.Error("IsTerminalEvent should be true for message_stop")
		}
	})

	t.Run("message_delta 包含 stop_reason", func(t *testing.T) {
		stopReason := anthropicTypes.StopReasonEndTurn
		event := &anthropicTypes.StreamEvent{
			MessageDelta: &anthropicTypes.MessageDeltaEvent{
				Delta: anthropicTypes.MessageDelta{
					StopReason: &stopReason,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal should be true for message_delta with stop_reason")
		}
		if signal.FinishReason != "end_turn" {
			t.Errorf("FinishReason = %v, want end_turn", signal.FinishReason)
		}
	})

	t.Run("content_block_delta 文本增量", func(t *testing.T) {
		text := "Hello"
		event := &anthropicTypes.StreamEvent{
			ContentBlockDelta: &anthropicTypes.ContentBlockDeltaEvent{
				Delta: types.ContentBlockDelta{
					Text: &anthropicTypes.TextDelta{
						Text: text,
					},
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.HasValidOutput {
			t.Error("HasValidOutput should be true for text delta")
		}
	})

	t.Run("content_block_delta 思考增量", func(t *testing.T) {
		thinking := "thinking..."
		event := &anthropicTypes.StreamEvent{
			ContentBlockDelta: &anthropicTypes.ContentBlockDeltaEvent{
				Delta: types.ContentBlockDelta{
					Thinking: &anthropicTypes.ThinkingDelta{
						Thinking: thinking,
					},
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.HasValidOutput {
			t.Error("HasValidOutput should be true for thinking delta")
		}
	})

	t.Run("content_block_start 工具调用", func(t *testing.T) {
		event := &anthropicTypes.StreamEvent{
			ContentBlockStart: &anthropicTypes.ContentBlockStartEvent{
				ContentBlock: anthropicTypes.ResponseContentBlock{
					ToolUse: &anthropicTypes.ToolUseBlock{},
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.HasValidOutput {
			t.Error("HasValidOutput should be true for tool_use block start")
		}
	})

	t.Run("错误事件", func(t *testing.T) {
		event := &anthropicTypes.StreamEvent{
			Error: &anthropicTypes.ErrorEvent{},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal should be true for error event")
		}
		if signal.FinishReason != "error" {
			t.Errorf("FinishReason = %v, want error", signal.FinishReason)
		}
	})

	t.Run("无效事件类型", func(t *testing.T) {
		signal := provider.IdentifyStreamEventSignal("", "invalid")
		if signal.IsCompletionSignal || signal.HasValidOutput || signal.IsTerminalEvent {
			t.Error("Signal should be empty for invalid event type")
		}
	})
}

// TestStreamEventSignal_ZeroValue 测试零值信号
func TestStreamEventSignal_ZeroValue(t *testing.T) {
	var signal StreamEventSignal
	if signal.IsCompletionSignal {
		t.Error("IsCompletionSignal should be false by default")
	}
	if signal.IsTerminalEvent {
		t.Error("IsTerminalEvent should be false by default")
	}
	if signal.HasValidOutput {
		t.Error("HasValidOutput should be false by default")
	}
	if signal.FinishReason != "" {
		t.Error("FinishReason should be empty by default")
	}
}

// TestOpenAI_ChatCompletions_FinishReasonAndUsage 测试 OpenAI Chat Completions
// finish_reason 与 usage 联合判定完成语义。
func TestOpenAI_ChatCompletions_FinishReasonAndUsage(t *testing.T) {
	provider := NewOpenAIProvider()

	t.Run("finish_reason_stop_标记完成信号", func(t *testing.T) {
		finishReason := openaiChat.FinishReasonStop
		event := &openaiChat.StreamEvent{
			Choices: []openaiChat.StreamChoice{
				{
					FinishReason: &finishReason,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("chat_completions", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if !signal.IsTerminalEvent {
			t.Error("IsTerminalEvent 应为 true")
		}
		if signal.FinishReason != "stop" {
			t.Errorf("FinishReason = %v, want stop", signal.FinishReason)
		}
		if signal.HasValidOutput {
			t.Error("仅有 finish_reason 时 HasValidOutput 应为 false")
		}
	})

	t.Run("finish_reason_content_filter_标记异常终止", func(t *testing.T) {
		finishReason := openaiChat.FinishReasonContentFilter
		event := &openaiChat.StreamEvent{
			Choices: []openaiChat.StreamChoice{
				{
					FinishReason: &finishReason,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("chat_completions", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if signal.FinishReason != "content_filter" {
			t.Errorf("FinishReason = %v, want content_filter", signal.FinishReason)
		}
	})

	t.Run("finish_reason_tool_calls_标记正常完成", func(t *testing.T) {
		finishReason := openaiChat.FinishReasonToolCalls
		event := &openaiChat.StreamEvent{
			Choices: []openaiChat.StreamChoice{
				{
					FinishReason: &finishReason,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("chat_completions", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if signal.FinishReason != "tool_calls" {
			t.Errorf("FinishReason = %v, want tool_calls", signal.FinishReason)
		}
	})

	t.Run("usage_块不单独标记完成", func(t *testing.T) {
		// usage 块出现在最终 chunk 中，但不含 finish_reason 时不标记完成
		promptTokens := 10
		completionTokens := 20
		totalTokens := 30
		event := &openaiChat.StreamEvent{
			Usage: &openaiChat.Usage{
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				TotalTokens:      totalTokens,
			},
		}
		signal := provider.IdentifyStreamEventSignal("chat_completions", event)
		if signal.IsCompletionSignal {
			t.Error("仅有 usage 块时 IsCompletionSignal 应为 false")
		}
		if signal.HasValidOutput {
			t.Error("仅有 usage 块时 HasValidOutput 应为 false")
		}
	})

	t.Run("finish_reason 与 usage 同时出现", func(t *testing.T) {
		finishReason := openaiChat.FinishReasonStop
		promptTokens := 10
		completionTokens := 20
		totalTokens := 30
		event := &openaiChat.StreamEvent{
			Choices: []openaiChat.StreamChoice{
				{
					FinishReason: &finishReason,
				},
			},
			Usage: &openaiChat.Usage{
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				TotalTokens:      totalTokens,
			},
		}
		signal := provider.IdentifyStreamEventSignal("chat_completions", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if signal.FinishReason != "stop" {
			t.Errorf("FinishReason = %v, want stop", signal.FinishReason)
		}
	})
}

// TestOpenAI_Responses_CompletionSemantics 测试 OpenAI Responses API
// 完成语义精修：response.completed、response.incomplete、error 事件。
func TestOpenAI_Responses_CompletionSemantics(t *testing.T) {
	provider := NewOpenAIProvider()

	t.Run("response.completed_正常完成", func(t *testing.T) {
		event := &openaiResponses.StreamEvent{
			Completed: &openaiResponses.ResponseCompletedEvent{},
		}
		signal := provider.IdentifyStreamEventSignal("responses", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if !signal.IsTerminalEvent {
			t.Error("IsTerminalEvent 应为 true")
		}
		if signal.FinishReason != "completed" {
			t.Errorf("FinishReason = %v, want completed", signal.FinishReason)
		}
	})

	t.Run("response.incomplete_无详情", func(t *testing.T) {
		event := &openaiResponses.StreamEvent{
			Incomplete: &openaiResponses.ResponseIncompleteEvent{},
		}
		signal := provider.IdentifyStreamEventSignal("responses", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if signal.FinishReason != "incomplete" {
			t.Errorf("FinishReason = %v, want incomplete", signal.FinishReason)
		}
	})

	t.Run("response.incomplete_有 IncompleteDetails_max_output_tokens", func(t *testing.T) {
		reason := "max_output_tokens"
		event := &openaiResponses.StreamEvent{
			Incomplete: &openaiResponses.ResponseIncompleteEvent{
				Response: openaiResponses.Response{
					IncompleteDetails: &openaiResponses.IncompleteDetails{
						Reason: &reason,
					},
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("responses", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if signal.FinishReason != "max_output_tokens" {
			t.Errorf("FinishReason = %v, want max_output_tokens", signal.FinishReason)
		}
	})

	t.Run("response.incomplete_有 IncompleteDetails_content_filter", func(t *testing.T) {
		reason := "content_filter"
		event := &openaiResponses.StreamEvent{
			Incomplete: &openaiResponses.ResponseIncompleteEvent{
				Response: openaiResponses.Response{
					IncompleteDetails: &openaiResponses.IncompleteDetails{
						Reason: &reason,
					},
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("responses", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if signal.FinishReason != "content_filter" {
			t.Errorf("FinishReason = %v, want content_filter", signal.FinishReason)
		}
	})

	t.Run("error_事件标记异常终止", func(t *testing.T) {
		code := "server_error"
		msg := "内部错误"
		event := &openaiResponses.StreamEvent{
			Error: &openaiResponses.ResponseErrorEvent{
				Code:    &code,
				Message: msg,
			},
		}
		signal := provider.IdentifyStreamEventSignal("responses", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if !signal.IsTerminalEvent {
			t.Error("IsTerminalEvent 应为 true")
		}
		if signal.FinishReason != "error" {
			t.Errorf("FinishReason = %v, want error", signal.FinishReason)
		}
	})
}

// TestAnthropic_MessageStopVsContentBlockStop 测试 Anthropic
// message_stop 与 content_block_stop 的区分。
func TestAnthropic_MessageStopVsContentBlockStop(t *testing.T) {
	provider := NewAnthropicProvider()

	t.Run("content_block_stop_不标记完成", func(t *testing.T) {
		// content_block_stop 仅表示单个内容块结束，不是整体完成
		event := &anthropicTypes.StreamEvent{
			ContentBlockStop: &anthropicTypes.ContentBlockStopEvent{
				Index: 0,
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if signal.IsCompletionSignal {
			t.Error("content_block_stop 不应标记 IsCompletionSignal")
		}
		if signal.IsTerminalEvent {
			t.Error("content_block_stop 不应标记 IsTerminalEvent")
		}
		if signal.HasValidOutput {
			t.Error("content_block_stop 不应标记 HasValidOutput")
		}
	})

	t.Run("message_stop_标记完成信号", func(t *testing.T) {
		event := &anthropicTypes.StreamEvent{
			MessageStop: &anthropicTypes.MessageStopEvent{},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("message_stop 应标记 IsCompletionSignal")
		}
		if !signal.IsTerminalEvent {
			t.Error("message_stop 应标记 IsTerminalEvent")
		}
	})

	t.Run("message_delta_stop_reason 优先于 message_stop", func(t *testing.T) {
		// 模拟 message_delta 先到达（含 stop_reason），然后 message_stop 到达
		stopReason := anthropicTypes.StopReasonEndTurn
		deltaEvent := &anthropicTypes.StreamEvent{
			MessageDelta: &anthropicTypes.MessageDeltaEvent{
				Delta: anthropicTypes.MessageDelta{
					StopReason: &stopReason,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", deltaEvent)
		if !signal.IsCompletionSignal {
			t.Error("message_delta 含 stop_reason 应标记 IsCompletionSignal")
		}
		if signal.FinishReason != "end_turn" {
			t.Errorf("FinishReason = %v, want end_turn", signal.FinishReason)
		}
	})

	t.Run("message_stop_不覆盖已有 stop_reason", func(t *testing.T) {
		// 先通过 message_delta 设置 FinishReason
		state := NewStreamState()
		stopReason := anthropicTypes.StopReasonMaxTokens
		deltaEvent := &anthropicTypes.StreamEvent{
			MessageDelta: &anthropicTypes.MessageDeltaEvent{
				Delta: anthropicTypes.MessageDelta{
					StopReason: &stopReason,
				},
			},
		}
		deltaSignal := provider.IdentifyStreamEventSignal("", deltaEvent)
		state.UpdateFromSignal(deltaSignal)

		// 再收到 message_stop，不应覆盖已有的 max_tokens
		stopEvent := &anthropicTypes.StreamEvent{
			MessageStop: &anthropicTypes.MessageStopEvent{},
		}
		stopSignal := provider.IdentifyStreamEventSignal("", stopEvent)
		// message_stop 的 FinishReason 为 "stop"，但状态机中已有 "max_tokens"
		state.UpdateFromSignal(stopSignal)
		// CompletionReason 应保持 max_tokens（先到优先）
		if state.CompletionReason != "max_tokens" {
			t.Errorf("CompletionReason = %v, want max_tokens（不应被 message_stop 覆盖）", state.CompletionReason)
		}
	})

	t.Run("多内容块场景_content_block_stop 不误判", func(t *testing.T) {
		// 模拟多内容块场景：第一个 content_block_stop 后还有后续块
		blockStopEvent := &anthropicTypes.StreamEvent{
			ContentBlockStop: &anthropicTypes.ContentBlockStopEvent{
				Index: 0,
			},
		}
		signal := provider.IdentifyStreamEventSignal("", blockStopEvent)
		if signal.IsCompletionSignal || signal.IsTerminalEvent {
			t.Error("content_block_stop 不应标记为完成或终止事件")
		}

		// 后续的 message_delta 才是真正的完成信号
		stopReason := anthropicTypes.StopReasonEndTurn
		deltaEvent := &anthropicTypes.StreamEvent{
			MessageDelta: &anthropicTypes.MessageDeltaEvent{
				Delta: anthropicTypes.MessageDelta{
					StopReason: &stopReason,
				},
			},
		}
		signal = provider.IdentifyStreamEventSignal("", deltaEvent)
		if !signal.IsCompletionSignal {
			t.Error("message_delta 含 stop_reason 应标记 IsCompletionSignal")
		}
	})
}

// TestGemini_FinishReasonAndSafetyInterception 测试 Gemini
// finishReason / usageMetadata / 安全拦截判定。
func TestGemini_FinishReasonAndSafetyInterception(t *testing.T) {
	provider := NewGeminiProvider()

	t.Run("STOP_正常完成", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					FinishReason: geminiTypes.FinishReasonStop,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if signal.FinishReason != geminiTypes.FinishReasonStop {
			t.Errorf("FinishReason = %v, want STOP", signal.FinishReason)
		}
	})

	t.Run("MAX_TOKENS_输出截断", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					FinishReason: geminiTypes.FinishReasonMaxTokens,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if signal.FinishReason != geminiTypes.FinishReasonMaxTokens {
			t.Errorf("FinishReason = %v, want MAX_TOKENS", signal.FinishReason)
		}
	})

	t.Run("SAFETY_安全拦截", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					FinishReason: geminiTypes.FinishReasonSafety,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
		if signal.FinishReason != geminiTypes.FinishReasonSafety {
			t.Errorf("FinishReason = %v, want SAFETY", signal.FinishReason)
		}
	})

	t.Run("RECITATION_引用拦截", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					FinishReason: geminiTypes.FinishReasonRecitation,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
	})

	t.Run("BLOCKLIST_阻止列表拦截", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					FinishReason: geminiTypes.FinishReasonBlocklist,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
	})

	t.Run("PROHIBITED_CONTENT_禁止内容拦截", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					FinishReason: geminiTypes.FinishReasonProhibitedContent,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
	})

	t.Run("MALFORMED_FUNCTION_CALL_异常终止", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					FinishReason: geminiTypes.FinishReasonMalformedFunction,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("IsCompletionSignal 应为 true")
		}
	})

	t.Run("FINISH_REASON_UNSPECIFIED_不标记完成", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					FinishReason: geminiTypes.FinishReasonUnspecified,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if signal.IsCompletionSignal {
			t.Error("FINISH_REASON_UNSPECIFIED 不应标记 IsCompletionSignal")
		}
	})

	t.Run("PromptFeedback_SAFETY_提示级安全拦截", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			PromptFeedback: &geminiTypes.PromptFeedback{
				BlockReason: geminiTypes.BlockReasonSafety,
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("PromptFeedback 安全拦截应标记 IsCompletionSignal")
		}
		if !signal.IsTerminalEvent {
			t.Error("PromptFeedback 安全拦截应标记 IsTerminalEvent")
		}
		if signal.FinishReason != geminiTypes.BlockReasonSafety {
			t.Errorf("FinishReason = %v, want SAFETY", signal.FinishReason)
		}
	})

	t.Run("PromptFeedback_BLOCK_REASON_UNSPECIFIED_不标记完成", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			PromptFeedback: &geminiTypes.PromptFeedback{
				BlockReason: geminiTypes.BlockReasonUnspecified,
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if signal.IsCompletionSignal {
			t.Error("BLOCK_REASON_UNSPECIFIED 不应标记 IsCompletionSignal")
		}
	})

	t.Run("PromptFeedback_PROHIBITED_CONTENT_提示级禁止内容拦截", func(t *testing.T) {
		event := &geminiTypes.StreamEvent{
			PromptFeedback: &geminiTypes.PromptFeedback{
				BlockReason: geminiTypes.BlockReasonProhibited,
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if !signal.IsCompletionSignal {
			t.Error("PromptFeedback 禁止内容拦截应标记 IsCompletionSignal")
		}
	})

	t.Run("UsageMetadata_不单独标记完成", func(t *testing.T) {
		// UsageMetadata 单独出现不标记完成，需要与 FinishReason 联合
		event := &geminiTypes.StreamEvent{
			UsageMetadata: &geminiTypes.UsageMetadata{
				PromptTokenCount:     10,
				CandidatesTokenCount: 20,
				TotalTokenCount:      30,
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		if signal.IsCompletionSignal {
			t.Error("仅有 UsageMetadata 时 IsCompletionSignal 应为 false")
		}
	})

	t.Run("安全拦截与正常完成区分", func(t *testing.T) {
		// SAFETY 拦截应被 IsAbnormalTermination 识别
		state := NewStreamState()
		event := &geminiTypes.StreamEvent{
			Candidates: []geminiTypes.Candidate{
				{
					FinishReason: geminiTypes.FinishReasonSafety,
				},
			},
		}
		signal := provider.IdentifyStreamEventSignal("", event)
		state.UpdateFromSignal(signal)
		if !state.IsAbnormalTermination() {
			t.Error("SAFETY 应被识别为异常终止")
		}
		if state.IsNormalCompletion() {
			t.Error("SAFETY 不应被识别为正常完成")
		}
	})
}
