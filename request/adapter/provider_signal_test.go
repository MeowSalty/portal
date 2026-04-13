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
