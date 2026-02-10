package adapter

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/MeowSalty/portal/errors"
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

// handleStreamingRaw 处理原生流式请求（通用 SSE 处理器）
//
// 该方法提供通用的 SSE 流处理能力，将数据解析委托给外部回调函数。
// 用于原生入口的流式请求处理，不经过 Provider 接口，直接解析原生类型。
//
// 参数：
//   - ctx: 上下文
//   - channel: 通道信息
//   - headers: 自定义 HTTP 头部
//   - payload: 请求体（原生类型，将直接序列化为 JSON）
//   - apiEndpoint: API 端点路径
//   - parseData: 解析 data 块的回调函数，将原始数据解析为原生事件类型
//   - sendEvent: 发送事件的回调函数，将解析后的事件发送到输出通道
func (a *Adapter) handleStreamingRaw(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	payload interface{},
	apiEndpoint string,
	parseData func([]byte) (interface{}, error),
	sendEvent func(context.Context, interface{}),
) error {
	// 构建 URL（使用传入的 API 端点）
	url := channel.BaseURL + apiEndpoint

	// 序列化请求体
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(errors.ErrCodeInternal, "序列化请求体失败", err)
	}

	// 创建请求和响应对象
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()

	// 设置请求参数
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// 添加提供商特定头部（包括身份验证头部）
	for key, value := range a.provider.Headers(channel.APIKey) {
		req.Header.Set(key, value)
	}

	// 应用请求级别的自定义 HTTP 头部
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 应用通道级别的自定义 HTTP 头部（优先级最高，会覆盖请求级别的同名头部）
	if channel.CustomHeaders != nil {
		for key, value := range channel.CustomHeaders {
			req.Header.Set(key, value)
		}
	}

	req.SetBody(jsonData)

	// 发送请求
	err = a.client.Do(req, resp)
	if err != nil {
		fasthttp.ReleaseResponse(resp)
		return errors.Wrap(errors.ErrCodeUnavailable, "HTTP 请求失败", stripErrorHTML(err))
	}

	// 检查 HTTP 状态码
	if resp.StatusCode() != fasthttp.StatusOK {
		body := resp.Body()
		defer fasthttp.ReleaseResponse(resp)
		return a.handleHTTPError("API 返回错误状态码", resp.StatusCode(), string(resp.Header.ContentType()), body)
	}

	// 获取响应体流
	bodyStream := resp.BodyStream()
	if bodyStream == nil {
		fasthttp.ReleaseResponse(resp)
		return errors.New(errors.ErrCodeStreamError, "流式响应体为空")
	}

	// 处理流式响应
	go func() {
		defer func() {
			fasthttp.ReleaseResponse(resp)
		}()

		reader := bufio.NewReaderSize(bodyStream, 4096)

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

					// 使用回调解析数据块
					event, parseErr := parseData([]byte(data))
					if parseErr != nil {
						// 解析失败，停止流处理
						return
					}

					// 发送事件
					sendEvent(ctx, event)
				}

				// 检查错误
				if err != nil {
					if err == io.EOF {
						// 流已结束
						return
					}
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
