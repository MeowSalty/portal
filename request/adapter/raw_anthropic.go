package adapter

import (
	"context"
	"encoding/json"

	anthropicTypes "github.com/MeowSalty/portal/request/adapter/anthropic/types"
	"github.com/MeowSalty/portal/routing"
	"github.com/valyala/fasthttp"
)

// RawAnthropicMessages 执行 Anthropic Messages 原生请求（非流式）
//
// 该方法直接发送原生请求到 Anthropic Messages API，不经过 Provider 接口的转换。
// 请求体和响应体均为 Anthropic 原生类型。
//
// 参数：
//   - ctx: 上下文
//   - channel: 通道信息
//   - headers: 自定义 HTTP 头部
//   - req: Anthropic Messages 原生请求对象
//
// 返回：
//   - *anthropicTypes.Response: Anthropic Messages 原生响应对象
//   - error: 请求失败时返回错误
func (a *Adapter) RawAnthropicMessages(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	req *anthropicTypes.Request,
) (*anthropicTypes.Response, error) {
	// 发送请求
	httpResp, err := a.sendHTTPRequest(channel, headers, req, false)
	if err != nil {
		return nil, err
	}

	// 检查 HTTP 状态码
	if httpResp.StatusCode != fasthttp.StatusOK {
		err := a.handleHTTPError("API 返回错误状态码", httpResp.StatusCode, httpResp.ContentType, httpResp.Body)
		return nil, err
	}

	// 直接解析为原生响应类型
	var response anthropicTypes.Response
	if err := json.Unmarshal(httpResp.Body, &response); err != nil {
		err := a.handleParseError("响应解析错误", err, httpResp.Body)
		return nil, err
	}

	return &response, nil
}

// RawAnthropicMessagesStream 执行 Anthropic Messages 原生流式请求
//
// 该方法直接发送原生请求到 Anthropic Messages API，不经过 Provider 接口的转换。
// 请求体为 Anthropic Messages 原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - channel: 通道信息
//   - headers: 自定义 HTTP 头部
//   - req: Anthropic Messages 原生请求对象
//   - output: 原生流事件输出通道
//
// 返回：
//   - error: 请求失败时返回错误
func (a *Adapter) RawAnthropicMessagesStream(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	req *anthropicTypes.Request,
	output chan<- *anthropicTypes.StreamEvent,
) error {
	return a.handleStreamingRaw(
		ctx,
		channel,
		headers,
		req,
		"/v1/messages",
		func(data []byte) (interface{}, error) {
			var event anthropicTypes.StreamEvent
			err := json.Unmarshal(data, &event)
			return &event, err
		},
		func(ctx context.Context, event interface{}) {
			if evt, ok := event.(*anthropicTypes.StreamEvent); ok && evt != nil {
				select {
				case <-ctx.Done():
				case output <- evt:
				}
			}
		},
	)
}
