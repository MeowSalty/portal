package adapter

import (
	"bufio"
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
	"github.com/valyala/fasthttp"
)

// handleStreaming 处理流式请求
func (a *Adapter) handleStreaming(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	apiReq interface{},
	stream chan<- *types.StreamEventContract,
) error {
	// 创建流索引上下文，用于在流式响应转换过程中生成和维护稳定的索引值
	indexCtx := types.NewStreamIndexContext()

	// 发送 HTTP 请求
	httpResp, err := a.sendHTTPRequest(channel, headers, apiReq, true)
	if err != nil {
		return err
	}

	// 获取需要释放的响应对象
	var respToRelease *fasthttp.Response
	if resp, ok := httpResp.userData.(*fasthttp.Response); ok {
		respToRelease = resp
	}

	if httpResp.StatusCode != fasthttp.StatusOK {
		// 读取响应体以获取详细错误信息
		var body []byte
		if httpResp.BodyStream != nil {
			// 读取 BodyStream 的内容
			body, err = io.ReadAll(httpResp.BodyStream)
			if err != nil {
				body = []byte{}
			}
		} else {
			body = []byte{}
		}
		return a.handleHTTPError("API 返回错误状态码", httpResp.StatusCode, httpResp.ContentType, body)
	}

	// 检查 BodyStream 是否为 nil
	if httpResp.BodyStream == nil {
		return errors.New(errors.ErrCodeStreamError, "流式响应体为空")
	}

	// 处理流式响应
	go func() {
		defer func() {
			close(stream)
			if respToRelease != nil {
				fasthttp.ReleaseResponse(respToRelease)
			}
		}()

		reader := bufio.NewReaderSize(httpResp.BodyStream, 4096) // 使用更大的缓冲区提高性能

		for {
			select {
			case <-ctx.Done():
				// 上下文已取消，停止流处理
				return
			default:
				line, err := reader.ReadString('\n')

				// 处理数据
				line = strings.TrimSpace(line)
				if line != "" && strings.HasPrefix(line, "data:") {
					// 提取数据部分
					data := strings.TrimSpace(line[5:])
					if data == "[DONE]" {
						// 流式传输正常完成
						return
					}

					// 解析流式响应块，传入流索引上下文
					events, parseErr := a.provider.ParseStreamResponse(indexCtx, []byte(data))
					if parseErr != nil {
						parseErr := errors.Wrap(errors.ErrCodeStreamError, "解析流块失败", stripErrorHTML(parseErr)).
							WithContext("data", data)
						a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError, parseErr.Error())
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
							default:
								stream <- event
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
					streamErr := errors.Wrap(errors.ErrCodeStreamError, "读取流数据失败", stripErrorHTML(err))
					a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError, streamErr.Error())
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
	payload interface{},
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
	log.Debug("开始发送流式请求")
	httpResp, err := a.sendHTTPRequest(channel, headers, payload, true)
	if err != nil {
		log.Error("发送 HTTP 请求失败", "error", err)
		return err
	}

	// 获取需要释放的响应对象
	var respToRelease *fasthttp.Response
	if resp, ok := httpResp.userData.(*fasthttp.Response); ok {
		respToRelease = resp
	}

	log.Debug("收到 HTTP 响应",
		"status_code", httpResp.StatusCode,
		"content_type", string(httpResp.ContentType),
	)

	if httpResp.StatusCode != fasthttp.StatusOK {
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
		if respToRelease != nil {
			fasthttp.ReleaseResponse(respToRelease)
		}
		log.Warn("API 返回错误状态码",
			"status_code", httpResp.StatusCode,
			"response_body", string(body),
		)
		return a.handleHTTPError("API 返回错误状态码", httpResp.StatusCode, httpResp.ContentType, body)
	}

	// 检查 BodyStream 是否为 nil
	if httpResp.BodyStream == nil {
		log.Error("流式响应体为空")
		if respToRelease != nil {
			fasthttp.ReleaseResponse(respToRelease)
		}
		return errors.New(errors.ErrCodeStreamError, "流式响应体为空")
	}

	log.Debug("开始处理流式响应")

	// 首字节标记
	firstChunkReceived := false

	// 处理流式响应
	go func() {
		var streamErr error
		defer func() {
			// 触发 OnComplete Hook
			if hooks != nil {
				hooks.OnComplete(time.Now())
			}
			log.Debug("流式响应处理完成")
			close(output)
			if respToRelease != nil {
				fasthttp.ReleaseResponse(respToRelease)
			}
		}()

		reader := bufio.NewReaderSize(httpResp.BodyStream, 4096)

		for {
			select {
			case <-ctx.Done():
				// 上下文已取消，停止流处理
				log.Debug("上下文已取消，停止流处理", "context_err", ctx.Err())
				streamErr = ctx.Err()
				return
			default:
				line, err := reader.ReadString('\n')

				// 处理数据
				line = strings.TrimSpace(line)
				if line != "" && strings.HasPrefix(line, "data:") {
					// 提取数据部分
					data := strings.TrimSpace(line[5:])
					log.Debug("读取到流数据行", "raw_line", line, "data", data[:min(len(data), 100)]) // 只记录前 100 个字符避免日志过长
					if data == "[DONE]" {
						log.Debug("流式传输正常完成")
						// 流式传输正常完成
						return
					}

					// 使用 Provider 解析原生流事件
					event, parseErr := a.provider.ParseNativeStreamEvent(channel.APIVariant, []byte(data))
					if parseErr != nil {
						// 触发 OnError Hook
						if hooks != nil {
							hooks.OnError(parseErr)
						}
						streamErr = errors.Wrap(errors.ErrCodeStreamError, "解析原生流块失败", stripErrorHTML(parseErr)).
							WithContext("data", data)
						log.Error("解析原生流块失败",
							"data", data,
							"error", streamErr,
						)
						return
					}

					// 触发 OnFirstChunk Hook（首次收到有效事件）
					if !firstChunkReceived && hooks != nil {
						hooks.OnFirstChunk(time.Now())
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
						log.Debug("上下文已取消，停止发送响应块", "context_err", ctx.Err())
						streamErr = ctx.Err()
						return
					default:
						output <- event
						log.Debug("成功发送事件到输出通道")
					}
				}

				// 检查错误
				if err != nil {
					if err == io.EOF {
						log.Debug("流已结束 (EOF)")
						// 流已结束
						return
					}
					// 触发 OnError Hook
					if hooks != nil {
						hooks.OnError(err)
					}
					streamErr = err
					log.Error("读取流数据失败",
						"error", err,
					)
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
		Extensions: map[string]interface{}{
			"status_code": code,
		},
	}

	select {
	case <-ctx.Done():
	case stream <- errEvent:
	}
}
