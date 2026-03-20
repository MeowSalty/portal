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

	// 处理流式响应
	go func() {
		var streamErr error
		defer func() {
			if hooks != nil {
				if streamErr != nil {
					hooks.OnError(streamErr)
				} else {
					hooks.OnComplete(time.Now())
				}
			}
			if streamErr == nil {
				log.Info("stream_finished", "status", "completed")
			} else if errors.IsCanceled(streamErr) {
				log.Info("stream_finished", "status", "canceled", "error", streamErr)
			} else {
				log.Warn("stream_finished", "status", "aborted", "error", streamErr)
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
						// 流式传输正常完成
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

					// 触发 OnFirstChunk Hook（首次收到有效事件）
					if !firstChunkReceived && hooks != nil {
						hooks.OnFirstChunk(time.Now())
						log.Debug("first_chunk_received")
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
						// 流已结束
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
