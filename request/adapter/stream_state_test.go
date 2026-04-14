package adapter

import (
	"context"
	"testing"

	"github.com/MeowSalty/portal/errors"
)

// TestStreamState_UpdateFromSignal 测试状态机从信号更新状态
func TestStreamState_UpdateFromSignal(t *testing.T) {
	tests := []struct {
		name           string
		initialState   *StreamState
		signal         StreamEventSignal
		wantFirstEvent bool
		wantOutput     bool
		wantTerminal   bool
		wantCompletion bool
		wantReason     string
		wantPhase      StreamPhase
	}{
		{
			name:           "空信号不改变状态",
			initialState:   NewStreamState(),
			signal:         StreamEventSignal{},
			wantFirstEvent: false,
			wantOutput:     false,
			wantTerminal:   false,
			wantCompletion: false,
			wantPhase:      StreamPhaseConnecting,
		},
		{
			name:         "有效输出信号更新状态",
			initialState: NewStreamState(),
			signal: StreamEventSignal{
				HasValidOutput: true,
			},
			wantFirstEvent: true,
			wantOutput:     true,
			wantTerminal:   false,
			wantCompletion: false,
			wantPhase:      StreamPhaseReceiving,
		},
		{
			name:         "终止事件信号更新状态",
			initialState: NewStreamState(),
			signal: StreamEventSignal{
				IsTerminalEvent: true,
				FinishReason:    "stop",
			},
			wantFirstEvent: true,
			wantOutput:     false,
			wantTerminal:   true,
			wantCompletion: false,
			wantPhase:      StreamPhaseReceiving,
		},
		{
			name:         "完成信号更新状态",
			initialState: NewStreamState(),
			signal: StreamEventSignal{
				IsCompletionSignal: true,
				IsTerminalEvent:    true,
				FinishReason:       "end_turn",
			},
			wantFirstEvent: true,
			wantOutput:     false,
			wantTerminal:   true,
			wantCompletion: true,
			wantReason:     "end_turn",
			wantPhase:      StreamPhaseCompleted,
		},
		{
			name:         "首包即完成信号保持 completed 阶段",
			initialState: NewStreamState(),
			signal: StreamEventSignal{
				IsCompletionSignal: true,
				FinishReason:       "completed",
			},
			wantFirstEvent: true,
			wantOutput:     false,
			wantTerminal:   false,
			wantCompletion: true,
			wantReason:     "completed",
			wantPhase:      StreamPhaseCompleted,
		},
		{
			name: "累积状态更新",
			initialState: &StreamState{
				ReceivedFirstEvent: true,
				HasValidOutput:     true,
				CurrentPhase:       StreamPhaseReceiving,
			},
			signal: StreamEventSignal{
				IsCompletionSignal: true,
				IsTerminalEvent:    true,
				FinishReason:       "stop",
			},
			wantFirstEvent: true,
			wantOutput:     true,
			wantTerminal:   true,
			wantCompletion: true,
			wantReason:     "stop",
			wantPhase:      StreamPhaseCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initialState.UpdateFromSignal(tt.signal)
			if tt.initialState.ReceivedFirstEvent != tt.wantFirstEvent {
				t.Errorf("ReceivedFirstEvent = %v, want %v", tt.initialState.ReceivedFirstEvent, tt.wantFirstEvent)
			}
			if tt.initialState.HasValidOutput != tt.wantOutput {
				t.Errorf("HasValidOutput = %v, want %v", tt.initialState.HasValidOutput, tt.wantOutput)
			}
			if tt.initialState.ReceivedTerminalEvent != tt.wantTerminal {
				t.Errorf("ReceivedTerminalEvent = %v, want %v", tt.initialState.ReceivedTerminalEvent, tt.wantTerminal)
			}
			if tt.initialState.HasCompletionSignal != tt.wantCompletion {
				t.Errorf("HasCompletionSignal = %v, want %v", tt.initialState.HasCompletionSignal, tt.wantCompletion)
			}
			if tt.wantReason != "" && tt.initialState.CompletionReason != tt.wantReason {
				t.Errorf("CompletionReason = %v, want %v", tt.initialState.CompletionReason, tt.wantReason)
			}
			if tt.initialState.CurrentPhase != tt.wantPhase {
				t.Errorf("CurrentPhase = %v, want %v", tt.initialState.CurrentPhase, tt.wantPhase)
			}
		})
	}
}

func TestBuildNativeStreamFinishInfo_CancelAndCompletionMatrix(t *testing.T) {
	tests := []struct {
		name                 string
		state                *StreamState
		err                  error
		wantCompletionState  string
		wantConnectionStatus string
		wantFinishStatus     string
		wantCancelSource     string
	}{
		{
			name: "无内容且客户端取消",
			state: &StreamState{
				CurrentPhase:       StreamPhaseReceiving,
				DisconnectionPhase: StreamPhaseReceiving,
			},
			err:                  errors.NormalizeCanceled(context.Canceled),
			wantCompletionState:  "not_completed",
			wantConnectionStatus: "disconnected",
			wantFinishStatus:     "canceled",
			wantCancelSource:     "client",
		},
		{
			name: "无内容且超时",
			state: &StreamState{
				CurrentPhase:       StreamPhaseReceiving,
				DisconnectionPhase: StreamPhaseReceiving,
			},
			err:                  errors.NormalizeCanceled(context.DeadlineExceeded),
			wantCompletionState:  "not_completed",
			wantConnectionStatus: "disconnected",
			wantFinishStatus:     "timed_out",
			wantCancelSource:     "deadline",
		},
		{
			name: "部分输出中断保持 partial",
			state: &StreamState{
				HasValidOutput:     true,
				CurrentPhase:       StreamPhaseReceiving,
				DisconnectionPhase: StreamPhaseReceiving,
			},
			err:                  errors.NormalizeCanceledWithSource(context.Canceled, false),
			wantCompletionState:  "partial",
			wantConnectionStatus: "disconnected",
			wantFinishStatus:     "partial",
			wantCancelSource:     "server",
		},
		{
			name: "明确完成后断连",
			state: &StreamState{
				HasCompletionSignal:         true,
				DisconnectedAfterCompletion: true,
				CurrentPhase:                StreamPhaseCompleted,
				DisconnectionPhase:          StreamPhaseCompleted,
			},
			err:                  nil,
			wantCompletionState:  "completed",
			wantConnectionStatus: "completed_then_disconnected",
			wantFinishStatus:     "completed_then_disconnected",
			wantCancelSource:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := buildNativeStreamFinishInfo(tt.state, tt.err)
			if info.CompletionState != tt.wantCompletionState {
				t.Errorf("CompletionState = %v, want %v", info.CompletionState, tt.wantCompletionState)
			}
			if info.ConnectionStatus != tt.wantConnectionStatus {
				t.Errorf("ConnectionStatus = %v, want %v", info.ConnectionStatus, tt.wantConnectionStatus)
			}
			if info.FinishStatus != tt.wantFinishStatus {
				t.Errorf("FinishStatus = %v, want %v", info.FinishStatus, tt.wantFinishStatus)
			}
			if info.CancelSource != tt.wantCancelSource {
				t.Errorf("CancelSource = %v, want %v", info.CancelSource, tt.wantCancelSource)
			}
		})
	}
}

// TestStreamState_MarkDisconnected 测试断连状态标记
func TestStreamState_MarkDisconnected(t *testing.T) {
	tests := []struct {
		name                   string
		state                  *StreamState
		wantAfterCompletion    bool
		wantDisconnectionPhase StreamPhase
	}{
		{
			name:                   "接收阶段断连",
			state:                  &StreamState{CurrentPhase: StreamPhaseReceiving},
			wantAfterCompletion:    false,
			wantDisconnectionPhase: StreamPhaseReceiving,
		},
		{
			name: "完成后断连",
			state: &StreamState{
				CurrentPhase:        StreamPhaseCompleted,
				HasCompletionSignal: true,
			},
			wantAfterCompletion:    true,
			wantDisconnectionPhase: StreamPhaseCompleted,
		},
		{
			name: "终止事件后断连",
			state: &StreamState{
				CurrentPhase:          StreamPhaseReceiving,
				ReceivedTerminalEvent: true,
			},
			wantAfterCompletion:    true,
			wantDisconnectionPhase: StreamPhaseReceiving,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.MarkDisconnected()
			if tt.state.DisconnectedAfterCompletion != tt.wantAfterCompletion {
				t.Errorf("DisconnectedAfterCompletion = %v, want %v", tt.state.DisconnectedAfterCompletion, tt.wantAfterCompletion)
			}
			if tt.state.DisconnectionPhase != tt.wantDisconnectionPhase {
				t.Errorf("DisconnectionPhase = %v, want %v", tt.state.DisconnectionPhase, tt.wantDisconnectionPhase)
			}
		})
	}
}

// TestStreamState_IsNormalCompletion 测试正常完成判断
func TestStreamState_IsNormalCompletion(t *testing.T) {
	tests := []struct {
		name       string
		state      *StreamState
		wantNormal bool
	}{
		{
			name:       "无完成信号",
			state:      &StreamState{HasCompletionSignal: false},
			wantNormal: false,
		},
		// OpenAI Chat Completions
		{
			name: "stop 正常完成 (OpenAI Chat)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "stop",
			},
			wantNormal: true,
		},
		{
			name: "tool_calls 正常完成 (OpenAI Chat)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "tool_calls",
			},
			wantNormal: true,
		},
		// OpenAI Responses
		{
			name: "completed 正常完成 (OpenAI Responses)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "completed",
			},
			wantNormal: true,
		},
		// Anthropic
		{
			name: "end_turn 正常完成 (Anthropic)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "end_turn",
			},
			wantNormal: true,
		},
		{
			name: "tool_use 正常完成 (Anthropic)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "tool_use",
			},
			wantNormal: true,
		},
		{
			name: "stop_sequence 正常完成 (Anthropic)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "stop_sequence",
			},
			wantNormal: true,
		},
		{
			name: "pause_turn 正常完成 (Anthropic)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "pause_turn",
			},
			wantNormal: true,
		},
		// Gemini
		{
			name: "STOP 正常完成 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "STOP",
			},
			wantNormal: true,
		},
		// 异常终止
		{
			name: "length 异常终止 (OpenAI Chat)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "length",
			},
			wantNormal: false,
		},
		{
			name: "max_tokens 异常终止 (Anthropic)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "max_tokens",
			},
			wantNormal: false,
		},
		{
			name: "content_filter 异常终止 (OpenAI)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "content_filter",
			},
			wantNormal: false,
		},
		{
			name: "failed 异常终止 (OpenAI Responses)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "failed",
			},
			wantNormal: false,
		},
		{
			name: "SAFETY 异常终止 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "SAFETY",
			},
			wantNormal: false,
		},
		{
			name: "MAX_TOKENS 异常终止 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "MAX_TOKENS",
			},
			wantNormal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsNormalCompletion(); got != tt.wantNormal {
				t.Errorf("IsNormalCompletion() = %v, want %v", got, tt.wantNormal)
			}
		})
	}
}

// TestStreamState_IsAbnormalTermination 测试异常终止判断
func TestStreamState_IsAbnormalTermination(t *testing.T) {
	tests := []struct {
		name         string
		state        *StreamState
		wantAbnormal bool
	}{
		{
			name:         "无完成信号",
			state:        &StreamState{HasCompletionSignal: false},
			wantAbnormal: false,
		},
		// 正常完成不应被判为异常
		{
			name: "stop 正常完成",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "stop",
			},
			wantAbnormal: false,
		},
		{
			name: "completed 正常完成 (OpenAI Responses)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "completed",
			},
			wantAbnormal: false,
		},
		{
			name: "STOP 正常完成 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "STOP",
			},
			wantAbnormal: false,
		},
		{
			name: "tool_calls 正常完成 (OpenAI Chat)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "tool_calls",
			},
			wantAbnormal: false,
		},
		{
			name: "end_turn 正常完成 (Anthropic)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "end_turn",
			},
			wantAbnormal: false,
		},
		// OpenAI Chat 异常终止
		{
			name: "length 异常终止 (OpenAI Chat)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "length",
			},
			wantAbnormal: true,
		},
		{
			name: "content_filter 异常终止 (OpenAI Chat)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "content_filter",
			},
			wantAbnormal: true,
		},
		// OpenAI Responses 异常终止
		{
			name: "failed 异常终止 (OpenAI Responses)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "failed",
			},
			wantAbnormal: true,
		},
		{
			name: "incomplete 异常终止 (OpenAI Responses)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "incomplete",
			},
			wantAbnormal: true,
		},
		{
			name: "max_output_tokens 异常终止 (OpenAI Responses IncompleteDetails)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "max_output_tokens",
			},
			wantAbnormal: true,
		},
		{
			name: "error 异常终止 (OpenAI Responses 流级错误)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "error",
			},
			wantAbnormal: true,
		},
		// Anthropic 异常终止
		{
			name: "max_tokens 异常终止 (Anthropic)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "max_tokens",
			},
			wantAbnormal: true,
		},
		{
			name: "refusal 异常终止 (Anthropic)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "refusal",
			},
			wantAbnormal: true,
		},
		// Gemini 异常终止
		{
			name: "SAFETY 异常终止 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "SAFETY",
			},
			wantAbnormal: true,
		},
		{
			name: "MAX_TOKENS 异常终止 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "MAX_TOKENS",
			},
			wantAbnormal: true,
		},
		{
			name: "RECITATION 异常终止 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "RECITATION",
			},
			wantAbnormal: true,
		},
		{
			name: "BLOCKLIST 异常终止 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "BLOCKLIST",
			},
			wantAbnormal: true,
		},
		{
			name: "PROHIBITED_CONTENT 异常终止 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "PROHIBITED_CONTENT",
			},
			wantAbnormal: true,
		},
		{
			name: "MALFORMED_FUNCTION_CALL 异常终止 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "MALFORMED_FUNCTION_CALL",
			},
			wantAbnormal: true,
		},
		{
			name: "SPII 异常终止 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "SPII",
			},
			wantAbnormal: true,
		},
		{
			name: "LANGUAGE 异常终止 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "LANGUAGE",
			},
			wantAbnormal: true,
		},
		{
			name: "OTHER 异常终止 (Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "OTHER",
			},
			wantAbnormal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsAbnormalTermination(); got != tt.wantAbnormal {
				t.Errorf("IsAbnormalTermination() = %v, want %v", got, tt.wantAbnormal)
			}
		})
	}
}

// TestStreamState_HasOutputWithoutCompletion 测试有输出但未完成判断
func TestStreamState_HasOutputWithoutCompletion(t *testing.T) {
	tests := []struct {
		name    string
		state   *StreamState
		wantHas bool
	}{
		{
			name:    "无输出无完成",
			state:   &StreamState{},
			wantHas: false,
		},
		{
			name: "有输出无完成",
			state: &StreamState{
				HasValidOutput: true,
			},
			wantHas: true,
		},
		{
			name: "有输出有完成",
			state: &StreamState{
				HasValidOutput:      true,
				HasCompletionSignal: true,
			},
			wantHas: false,
		},
		{
			name: "有输出有终止事件",
			state: &StreamState{
				HasValidOutput:        true,
				ReceivedTerminalEvent: true,
			},
			wantHas: false,
		},
		{
			name: "有输出有完成信号",
			state: &StreamState{
				HasValidOutput:      true,
				HasCompletionSignal: true,
			},
			wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.HasOutputWithoutCompletion(); got != tt.wantHas {
				t.Errorf("HasOutputWithoutCompletion() = %v, want %v", got, tt.wantHas)
			}
		})
	}
}

// TestNewStreamState 测试状态机初始化
func TestNewStreamState(t *testing.T) {
	state := NewStreamState()
	if state == nil {
		t.Fatal("NewStreamState() returned nil")
	}
	if state.CurrentPhase != StreamPhaseConnecting {
		t.Errorf("CurrentPhase = %v, want %v", state.CurrentPhase, StreamPhaseConnecting)
	}
	if state.ReceivedFirstEvent {
		t.Error("ReceivedFirstEvent should be false initially")
	}
	if state.HasValidOutput {
		t.Error("HasValidOutput should be false initially")
	}
	if state.HasCompletionSignal {
		t.Error("HasCompletionSignal should be false initially")
	}
}
