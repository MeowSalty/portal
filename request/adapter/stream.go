package adapter

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
)

// SSE 解析常量，避免每次事件重复分配
var (
	sseDataPrefix = []byte("data:")
	sseDoneMarker = []byte("[DONE]")
)

// StreamPhase 表示流式处理的阶段。
// 用于断连发生时的状态记录。
type StreamPhase string

const (
	// StreamPhaseConnecting 连接建立阶段
	StreamPhaseConnecting StreamPhase = "connecting"
	// StreamPhaseReceiving 接收数据阶段
	StreamPhaseReceiving StreamPhase = "receiving"
	// StreamPhaseCompleted 完成阶段（收到明确完成信号）
	StreamPhaseCompleted StreamPhase = "completed"
	// StreamPhaseClosed 关闭阶段
	StreamPhaseClosed StreamPhase = "closed"
)

// StreamState 表示原生流式链路的内部状态。
//
// 该状态机用于跟踪流式处理的关键状态点，以支持：
//   - 区分"是否有输出""是否明确完成""是否完成后断连"
//   - 为重试语义提供精确的终止分类依据
//   - 避免仅依赖 [DONE] 标记或 EOF 进行猜测
//
// 状态机设计原则：
//   - 高内聚：状态仅在流式处理内部传递，不泄漏到无关层
//   - 最小化：仅记录必要状态，避免过度复杂化
type StreamState struct {
	// ReceivedFirstEvent 是否已收到首个有效事件。
	// 首个有效事件指成功解析的非空事件。
	ReceivedFirstEvent bool

	// HasValidOutput 是否已有有效输出内容。
	// 有效输出指文本增量、工具调用增量、音频增量等。
	HasValidOutput bool

	// ReceivedTerminalEvent 是否已收到明确终止/完成事件。
	// 终止事件指 provider 识别的 IsTerminalEvent。
	ReceivedTerminalEvent bool

	// HasCompletionSignal 是否已有协议级完成信号。
	// 完成信号指 provider 识别的 IsCompletionSignal。
	HasCompletionSignal bool

	// CompletionReason 完成原因（可选）。
	// 来自 provider 识别的 FinishReason。
	CompletionReason string

	// DisconnectedAfterCompletion 是否在完成后发生断连。
	// 用于区分"正常完成后的连接关闭"与"异常断连"。
	DisconnectedAfterCompletion bool

	// CurrentPhase 当前流处理阶段。
	CurrentPhase StreamPhase

	// DisconnectionPhase 断连发生时的阶段（可选）。
	// 仅在发生断连时记录。
	DisconnectionPhase StreamPhase
}

// NewStreamState 创建新的流状态实例。
func NewStreamState() *StreamState {
	return &StreamState{
		CurrentPhase: StreamPhaseConnecting,
	}
}

// UpdateFromSignal 根据事件信号更新状态。
func (s *StreamState) UpdateFromSignal(signal StreamEventSignal) {
	// 更新有效输出状态
	if signal.HasValidOutput {
		s.HasValidOutput = true
	}

	// 更新终止事件状态
	if signal.IsTerminalEvent {
		s.ReceivedTerminalEvent = true
	}

	// 更新完成信号状态
	if signal.IsCompletionSignal {
		s.HasCompletionSignal = true
		if signal.FinishReason != "" {
			s.CompletionReason = signal.FinishReason
		}
		// 收到完成信号时，进入完成阶段
		s.CurrentPhase = StreamPhaseCompleted
	}

	// 收到首个有效事件
	if signal.HasValidOutput || signal.IsTerminalEvent || signal.IsCompletionSignal {
		if !s.ReceivedFirstEvent {
			s.ReceivedFirstEvent = true
			s.CurrentPhase = StreamPhaseReceiving
		}
	}
}

// MarkDisconnected 标记断连状态。
func (s *StreamState) MarkDisconnected() {
	s.DisconnectionPhase = s.CurrentPhase
	// 如果在完成后断连，标记为完成后断连
	if s.HasCompletionSignal || s.ReceivedTerminalEvent {
		s.DisconnectedAfterCompletion = true
	}
}

// IsNormalCompletion 判断是否为正常完成。
// 正常完成指：收到完成信号且完成原因为正常类型。
func (s *StreamState) IsNormalCompletion() bool {
	if !s.HasCompletionSignal {
		return false
	}
	// 根据完成原因判断是否为正常完成
	// 正常完成原因：stop, end_turn, tool_use, completed
	normalReasons := map[string]bool{
		"stop":      true,
		"end_turn":  true,
		"tool_use":  true,
		"completed": true,
		"STOP":      true, // Gemini
	}
	return normalReasons[s.CompletionReason]
}

// IsAbnormalTermination 判断是否为异常终止。
// 异常终止指：收到完成信号但完成原因为异常类型。
func (s *StreamState) IsAbnormalTermination() bool {
	if !s.HasCompletionSignal {
		return false
	}
	// 异常完成原因：length, max_tokens, safety, content_filter, failed, incomplete
	abnormalReasons := map[string]bool{
		"length":         true,
		"max_tokens":     true,
		"safety":         true,
		"content_filter": true,
		"failed":         true,
		"incomplete":     true,
		"MAX_TOKENS":     true, // Gemini
		"SAFETY":         true, // Gemini
	}
	return abnormalReasons[s.CompletionReason]
}

// HasOutputWithoutCompletion 判断是否有输出但未完成。
// 这种情况可能需要重试。
func (s *StreamState) HasOutputWithoutCompletion() bool {
	return s.HasValidOutput && !s.HasCompletionSignal && !s.ReceivedTerminalEvent
}

// handleStreaming 处理流式请求
func (a *Adapter) handleStreaming(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	apiReq any,
	stream chan<- *types.StreamEventContract,
) error {
	// 创建流索引上下文，用于在流式响应转换过程中生成和维护稳定的索引值
	indexCtx := types.NewStreamIndexContext()

	// 发送 HTTP 请求
	httpResp, err := a.sendHTTPRequest(ctx, channel, headers, apiReq, true)
	if err != nil {
		return err
	}

	if httpResp.StatusCode != http.StatusOK {
		// 读取响应体以获取详细错误信息
		var body []byte
		if httpResp.BodyStream != nil {
			body, err = io.ReadAll(httpResp.BodyStream)
			if err != nil {
				body = []byte{}
			}
		} else {
			body = []byte{}
		}
		return a.handleHTTPError("API 返回错误状态码", httpResp.StatusCode, body)
	}

	// 检查 BodyStream 是否为 nil
	if httpResp.BodyStream == nil {
		return errors.New(errors.ErrCodeStreamError, "流式响应体为空").
			WithContext("error_from", string(errors.ErrorFromGateway))
	}

	// 处理流式响应
	go func() {
		defer func() {
			close(stream)
			if httpResp.body != nil {
				httpResp.body.Close()
			}
		}()

		reader := bufio.NewReaderSize(httpResp.BodyStream, 4096)

		for {
			select {
			case <-ctx.Done():
				// 上下文已取消，停止流处理
				return
			default:
				lineBytes, err := reader.ReadBytes('\n')

				// 处理数据（零拷贝）
				lineBytes = bytes.TrimSpace(lineBytes)
				if len(lineBytes) > 0 && bytes.HasPrefix(lineBytes, sseDataPrefix) {
					data := bytes.TrimSpace(lineBytes[5:])
					if bytes.Equal(data, sseDoneMarker) {
						// 流式传输正常完成
						return
					}

					// 解析流式响应块，直接传 []byte 避免拷贝
					events, parseErr := a.provider.ParseStreamResponse(channel.APIVariant, indexCtx, data)
					if parseErr != nil {
						parseErr := errors.Wrap(errors.ErrCodeStreamError, "解析流块失败", stripErrorHTML(parseErr)).
							WithContext("data", string(data)).
							WithContext("error_from", string(errors.ErrorFromGateway))
						a.sendStreamError(ctx, stream, http.StatusInternalServerError, parseErr.Error())
						return
					}

					// 确保响应块有效后再发送
					if len(events) > 0 {
						for _, event := range events {
							if event == nil {
								continue
							}
							select {
							case <-ctx.Done():
								// 上下文已取消，停止发送响应块
								return
							case stream <- event:
							}
						}
					}
				}

				// 检查错误
				if err != nil {
					if err == io.EOF {
						// 流已结束
						return
					}
					// 检查取消错误，与 handleNativeStreaming 保持一致
					if errors.IsCanceled(err) || errors.IsCanceled(ctx.Err()) {
						return
					}
					streamErr := errors.Wrap(errors.ErrCodeStreamError, "读取流数据失败", stripErrorHTML(err)).
						WithContext("error_from", string(errors.ErrorFromGateway))
					a.sendStreamError(ctx, stream, http.StatusInternalServerError, streamErr.Error())
					return
				}
			}
		}
	}()

	return nil
}

// handleNativeStreaming 处理原生流式请求
//
// 该方法使用 Provider 的 ParseNativeStreamEvent 方法解析原生流事件。
// 与 handleStreaming 的区别在于直接输出原生类型，而不是 contract 类型。
func (a *Adapter) handleNativeStreaming(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	payload any,
	output chan<- any,
	hooks types.StreamHooks,
) error {
	log := logger.Default().WithGroup("native-stream").With(
		"api_provider", channel.Provider,
		"api_variant", channel.APIVariant,
		"api_baseurl", channel.BaseURL,
		"api_endpoint", channel.APIEndpointConfig,
	)

	// 发送 HTTP 请求
	log.Debug("request_started")
	httpResp, err := a.sendHTTPRequest(ctx, channel, headers, payload, true)
	if err != nil {
		return err
	}

	if httpResp.StatusCode != http.StatusOK {
		// 读取响应体以获取详细错误信息
		var body []byte
		if httpResp.BodyStream != nil {
			body, err = io.ReadAll(httpResp.BodyStream)
			if err != nil {
				body = []byte{}
			}
		} else {
			body = []byte{}
		}
		if httpResp.body != nil {
			httpResp.body.Close()
		}
		log.Warn("API 返回错误状态码",
			"status_code", httpResp.StatusCode,
			"response_body", string(body),
		)
		return a.handleHTTPError("API 返回错误状态码", httpResp.StatusCode, body)
	}

	// 检查 BodyStream 是否为 nil
	if httpResp.BodyStream == nil {
		if httpResp.body != nil {
			httpResp.body.Close()
		}
		return errors.New(errors.ErrCodeStreamError, "流式响应体为空").
			WithContext("error_from", string(errors.ErrorFromGateway))
	}

	// 首字节标记
	firstChunkReceived := false

	// 创建流状态跟踪器
	streamState := NewStreamState()

	// 处理流式响应
	go func() {
		var streamErr error
		defer func() {
			// 标记断连状态
			streamState.MarkDisconnected()
			finishInfo := buildNativeStreamFinishInfo(streamState, streamErr)

			if hooks != nil {
				hooks.OnStreamFinished(finishInfo)
				if isSuccessfulStreamFinish(finishInfo) {
					hooks.OnComplete(time.Now())
				} else {
					if streamErr == nil {
						streamErr = errors.New(errors.ErrCodeStreamError, "流式响应未完成").
							WithContext("completion_state", finishInfo.CompletionState).
							WithContext("connection_status", finishInfo.ConnectionStatus).
							WithContext("finish_status", finishInfo.FinishStatus).
							WithContext("error_from", string(errors.ErrorFromGateway))
					}
					hooks.OnError(streamErr)
				}
			}
			// stream_finished 语义统一由上层重试/入口层记录，
			// 适配层仅输出底层调试信息，避免 completed/canceled 重复记账。
			if isSuccessfulStreamFinish(finishInfo) {
				log.Debug("native_stream_worker_finished",
					"status", finishInfo.FinishStatus,
					"completion_state", finishInfo.CompletionState,
					"connection_status", finishInfo.ConnectionStatus,
					"stream_state", streamState,
				)
			} else if errors.IsCanceled(streamErr) {
				log.Debug("native_stream_worker_finished",
					"status", finishInfo.FinishStatus,
					"completion_state", finishInfo.CompletionState,
					"connection_status", finishInfo.ConnectionStatus,
					"cancel_source", finishInfo.CancelSource,
					"error", streamErr,
					"stream_state", streamState,
				)
			} else {
				log.Warn("native_stream_worker_finished",
					"status", finishInfo.FinishStatus,
					"completion_state", finishInfo.CompletionState,
					"connection_status", finishInfo.ConnectionStatus,
					"error", streamErr,
					"stream_state", streamState,
				)
			}
			close(output)
			if httpResp.body != nil {
				httpResp.body.Close()
			}
		}()

		reader := bufio.NewReaderSize(httpResp.BodyStream, 4096)

		for {
			select {
			case <-ctx.Done():
				// 上下文已取消，停止流处理
				streamErr = errors.NormalizeCanceled(ctx.Err())
				return
			default:
				lineBytes, err := reader.ReadBytes('\n')

				// 处理数据（零拷贝）
				lineBytes = bytes.TrimSpace(lineBytes)
				if len(lineBytes) > 0 && bytes.HasPrefix(lineBytes, sseDataPrefix) {
					data := bytes.TrimSpace(lineBytes[5:])
					if bytes.Equal(data, sseDoneMarker) {
						// 流式传输正常完成（[DONE] 标记）
						// 如果状态机未收到完成信号，这里作为兼容性兜底
						if !streamState.HasCompletionSignal {
							streamState.HasCompletionSignal = true
							streamState.CompletionReason = "done_marker"
							streamState.CurrentPhase = StreamPhaseCompleted
						}
						return
					}

					if errChunk, ok := a.tryBuildStreamChunkError("API 流中返回错误块", data); ok {
						streamErr = errChunk
						return
					}

					// 使用 Provider 解析原生流事件，直接传 []byte
					event, parseErr := a.provider.ParseNativeStreamEvent(channel.APIVariant, data)
					if parseErr != nil {
						streamErr = errors.Wrap(errors.ErrCodeStreamError, "解析原生流块失败", stripErrorHTML(parseErr)).
							WithContext("data", string(data)).
							WithContext("error_from", string(errors.ErrorFromGateway))
						return
					}

					// 识别事件信号并更新状态
					signal := a.provider.IdentifyStreamEventSignal(channel.APIVariant, event)
					streamState.UpdateFromSignal(signal)

					// 触发 OnFirstChunk Hook（首次收到有效事件）
					if !firstChunkReceived && (signal.HasValidOutput || signal.IsTerminalEvent || signal.IsCompletionSignal) {
						if hooks != nil {
							hooks.OnFirstChunk(time.Now())
							log.Debug("first_chunk_received")
						}
						firstChunkReceived = true
					}

					// 从原生流事件中提取 usage 信息
					responseUsage := a.provider.ExtractUsageFromNativeStreamEvent(channel.APIVariant, event)
					if responseUsage != nil && hooks != nil {
						usage := types.Usage{}
						if responseUsage.InputTokens != nil {
							usage.InputTokens = *responseUsage.InputTokens
						}
						if responseUsage.OutputTokens != nil {
							usage.OutputTokens = *responseUsage.OutputTokens
						}
						if responseUsage.TotalTokens != nil {
							usage.TotalTokens = *responseUsage.TotalTokens
						}
						hooks.OnUsage(usage)
					}

					// 发送原生事件
					select {
					case <-ctx.Done():
						// 上下文已取消，停止发送响应块
						streamErr = errors.NormalizeCanceled(ctx.Err())
						return
					case output <- event:
					}
				}

				// 检查错误
				if err != nil {
					if err == io.EOF {
						// 流已结束（EOF）
						// 如果状态机未收到完成信号，记录断连阶段
						if !streamState.HasCompletionSignal {
							streamState.MarkDisconnected()
							log.Debug("stream_ended_without_completion",
								"stream_state", streamState,
							)
						}
						return
					}

					if errors.IsCanceled(err) || errors.IsCanceled(ctx.Err()) {
						cancelErr := err
						if ctx.Err() != nil {
							cancelErr = ctx.Err()
						}
						streamErr = errors.NormalizeCanceled(cancelErr)
						return
					}

					streamErr = errors.Wrap(errors.ErrCodeStreamError, "读取流数据失败", stripErrorHTML(err)).
						WithContext("error_from", string(errors.ErrorFromGateway))
					return
				}
			}
		}
	}()

	return nil
}

func buildNativeStreamFinishInfo(state *StreamState, err error) types.StreamFinishInfo {
	info := types.StreamFinishInfo{
		ConnectionStatus: "disconnected",
	}

	if state == nil {
		info.CompletionState = "not_completed"
		info.FinishStatus = "failed"
		if err != nil {
			info.FinishStatus, info.CancelSource = classifyStreamCancel(err)
		}
		return info
	}

	if state.HasCompletionSignal || state.ReceivedTerminalEvent {
		info.CompletionState = "completed"
	} else if state.HasValidOutput {
		info.CompletionState = "partial"
	} else {
		info.CompletionState = "not_completed"
	}

	if state.DisconnectedAfterCompletion {
		info.ConnectionStatus = "completed_then_disconnected"
	}

	if state.DisconnectionPhase != "" {
		info.TerminationPhase = string(state.DisconnectionPhase)
	}

	switch info.CompletionState {
	case "completed":
		if info.ConnectionStatus == "completed_then_disconnected" {
			info.FinishStatus = "completed_then_disconnected"
		} else {
			info.FinishStatus = "completed"
		}
	case "partial":
		info.FinishStatus = "partial"
	default:
		info.FinishStatus = "failed"
	}

	if err != nil {
		if finishStatus, cancelSource := classifyStreamCancel(err); finishStatus != "" {
			if info.CompletionState == "completed" {
				if info.ConnectionStatus == "completed_then_disconnected" {
					info.FinishStatus = "completed_then_disconnected"
				} else {
					info.FinishStatus = "completed"
				}
			} else if info.CompletionState == "partial" {
				info.FinishStatus = "partial"
				info.CancelSource = cancelSource
			} else {
				info.FinishStatus = finishStatus
				info.CancelSource = cancelSource
			}
		}
	}

	return info
}

func classifyStreamCancel(err error) (finishStatus string, cancelSource string) {
	if err == nil || !errors.IsCanceled(err) {
		return "", ""
	}

	if errors.IsCode(err, errors.ErrCodeDeadlineExceeded) || errors.IsDeadlineExceeded(err) {
		return "timed_out", "deadline"
	}
	if errors.IsCode(err, errors.ErrCodeCanceled) || errors.GetErrorFrom(err) == errors.ErrorFromServer {
		return "canceled", "server"
	}
	if errors.IsCode(err, errors.ErrCodeAborted) || errors.GetErrorFrom(err) == errors.ErrorFromClient {
		return "canceled", "client"
	}

	return "canceled", "unknown"
}

func isSuccessfulStreamFinish(info types.StreamFinishInfo) bool {
	return info.CompletionState == "completed"
}

// sendStreamError 向流发送错误信息
func (a *Adapter) sendStreamError(
	ctx context.Context,
	stream chan<- *types.StreamEventContract,
	code int,
	message string,
) {
	errEvent := &types.StreamEventContract{
		Type: types.StreamEventError,
		Error: &types.StreamErrorPayload{
			Message: message,
			Type:    "stream_error",
			Code:    strconv.Itoa(code),
		},
		Extensions: map[string]any{
			"status_code": code,
		},
	}

	select {
	case <-ctx.Done():
	case stream <- errEvent:
	}
}
