package adapter

import (
	"testing"
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
	}{
		{
			name:           "空信号不改变状态",
			initialState:   NewStreamState(),
			signal:         StreamEventSignal{},
			wantFirstEvent: false,
			wantOutput:     false,
			wantTerminal:   false,
			wantCompletion: false,
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
		{
			name: "stop正常完成",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "stop",
			},
			wantNormal: true,
		},
		{
			name: "end_turn正常完成",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "end_turn",
			},
			wantNormal: true,
		},
		{
			name: "tool_use正常完成",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "tool_use",
			},
			wantNormal: true,
		},
		{
			name: "STOP正常完成(Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "STOP",
			},
			wantNormal: true,
		},
		{
			name: "length异常终止",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "length",
			},
			wantNormal: false,
		},
		{
			name: "max_tokens异常终止",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "max_tokens",
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
		{
			name: "stop正常完成",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "stop",
			},
			wantAbnormal: false,
		},
		{
			name: "length异常终止",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "length",
			},
			wantAbnormal: true,
		},
		{
			name: "max_tokens异常终止",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "max_tokens",
			},
			wantAbnormal: true,
		},
		{
			name: "SAFETY异常终止(Gemini)",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "SAFETY",
			},
			wantAbnormal: true,
		},
		{
			name: "content_filter异常终止",
			state: &StreamState{
				HasCompletionSignal: true,
				CompletionReason:    "content_filter",
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
