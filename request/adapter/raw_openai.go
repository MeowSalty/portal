package adapter

import (
	"context"
	"encoding/json"

	openaiChat "github.com/MeowSalty/portal/request/adapter/openai/types/chat"
	openaiResponses "github.com/MeowSalty/portal/request/adapter/openai/types/responses"
	"github.com/MeowSalty/portal/routing"
	"github.com/valyala/fasthttp"
)

// RawOpenAIChatCompletion 执行 OpenAI Chat 原生请求（非流式）
//
// 该方法直接发送原生请求到 OpenAI Chat Completions API，不经过 Provider 接口的转换。
// 请求体和响应体均为 OpenAI 原生类型。
//
// 参数：
//   - ctx: 上下文
//   - channel: 通道信息
//   - headers: 自定义 HTTP 头部
//   - req: OpenAI Chat 原生请求对象
//
// 返回：
//   - *openaiChat.Response: OpenAI Chat 原生响应对象
//   - error: 请求失败时返回错误
func (a *Adapter) RawOpenAIChatCompletion(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	req *openaiChat.Request,
) (*openaiChat.Response, error) {
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
	var response openaiChat.Response
	if err := json.Unmarshal(httpResp.Body, &response); err != nil {
		err := a.handleParseError("响应解析错误", err, httpResp.Body)
		return nil, err
	}

	return &response, nil
}

// RawOpenAIChatCompletionStream 执行 OpenAI Chat 原生流式请求
//
// 该方法直接发送原生请求到 OpenAI Chat Completions API，不经过 Provider 接口的转换。
// 请求体为 OpenAI 原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - channel: 通道信息
//   - headers: 自定义 HTTP 头部
//   - req: OpenAI Chat 原生请求对象
//   - output: 原生流事件输出通道
//
// 返回：
//   - error: 请求失败时返回错误
func (a *Adapter) RawOpenAIChatCompletionStream(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	req *openaiChat.Request,
	output chan<- *openaiChat.StreamEvent,
) error {
	return a.handleStreamingRaw(
		ctx,
		channel,
		headers,
		req,
		"/v1/chat/completions",
		func(data []byte) (interface{}, error) {
			var event openaiChat.StreamEvent
			err := json.Unmarshal(data, &event)
			return &event, err
		},
		func(ctx context.Context, event interface{}) {
			if evt, ok := event.(*openaiChat.StreamEvent); ok && evt != nil {
				select {
				case <-ctx.Done():
				case output <- evt:
				}
			}
		},
	)
}

// RawOpenAIResponses 执行 OpenAI Responses 原生请求（非流式）
//
// 该方法直接发送原生请求到 OpenAI Responses API，不经过 Provider 接口的转换。
// 请求体和响应体均为 OpenAI Responses 原生类型。
//
// 参数：
//   - ctx: 上下文
//   - channel: 通道信息
//   - headers: 自定义 HTTP 头部
//   - req: OpenAI Responses 原生请求对象
//
// 返回：
//   - *openaiResponses.Response: OpenAI Responses 原生响应对象
//   - error: 请求失败时返回错误
func (a *Adapter) RawOpenAIResponses(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	req *openaiResponses.Request,
) (*openaiResponses.Response, error) {
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
	var response openaiResponses.Response
	if err := json.Unmarshal(httpResp.Body, &response); err != nil {
		err := a.handleParseError("响应解析错误", err, httpResp.Body)
		return nil, err
	}

	return &response, nil
}

// RawOpenAIResponsesStream 执行 OpenAI Responses 原生流式请求
//
// 该方法直接发送原生请求到 OpenAI Responses API，不经过 Provider 接口的转换。
// 请求体为 OpenAI Responses 原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - channel: 通道信息
//   - headers: 自定义 HTTP 头部
//   - req: OpenAI Responses 原生请求对象
//   - output: 原生流事件输出通道
//
// 返回：
//   - error: 请求失败时返回错误
func (a *Adapter) RawOpenAIResponsesStream(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	req *openaiResponses.Request,
	output chan<- *openaiResponses.StreamEvent,
) error {
	return a.handleStreamingRaw(
		ctx,
		channel,
		headers,
		req,
		"/v1/responses",
		func(data []byte) (interface{}, error) {
			var event openaiResponses.StreamEvent
			err := json.Unmarshal(data, &event)
			return &event, err
		},
		func(ctx context.Context, event interface{}) {
			if evt, ok := event.(*openaiResponses.StreamEvent); ok && evt != nil {
				select {
				case <-ctx.Done():
				case output <- evt:
				}
			}
		},
	)
}
