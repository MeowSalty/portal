package adapter

import (
	"context"
	"encoding/json"

	geminiTypes "github.com/MeowSalty/portal/request/adapter/gemini/types"
	"github.com/MeowSalty/portal/routing"
	"github.com/valyala/fasthttp"
)

// RawGeminiGenerateContent 执行 Gemini GenerateContent 原生请求（非流式）
//
// 该方法直接发送原生请求到 Gemini GenerateContent API，不经过 Provider 接口的转换。
// 请求体和响应体均为 Gemini 原生类型。
//
// 参数：
//   - ctx: 上下文
//   - channel: 通道信息
//   - headers: 自定义 HTTP 头部
//   - req: Gemini GenerateContent 原生请求对象
//
// 返回：
//   - *geminiTypes.Response: Gemini GenerateContent 原生响应对象
//   - error: 请求失败时返回错误
func (a *Adapter) RawGeminiGenerateContent(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	req *geminiTypes.Request,
) (*geminiTypes.Response, error) {
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
	var response geminiTypes.Response
	if err := json.Unmarshal(httpResp.Body, &response); err != nil {
		err := a.handleParseError("响应解析错误", err, httpResp.Body)
		return nil, err
	}

	return &response, nil
}

// RawGeminiStreamGenerateContent 执行 Gemini StreamGenerateContent 原生流式请求
//
// 该方法直接发送原生请求到 Gemini StreamGenerateContent API，不经过 Provider 接口的转换。
// 请求体为 Gemini GenerateContent 原生类型，响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - channel: 通道信息
//   - headers: 自定义 HTTP 头部
//   - req: Gemini GenerateContent 原生请求对象
//   - output: 原生流事件输出通道
//
// 返回：
//   - error: 请求失败时返回错误
func (a *Adapter) RawGeminiStreamGenerateContent(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	req *geminiTypes.Request,
	output chan<- *geminiTypes.StreamEvent,
) error {
	return a.handleStreamingRaw(
		ctx,
		channel,
		headers,
		req,
		"/v1beta/models/"+req.Model+":streamGenerateContent",
		func(data []byte) (interface{}, error) {
			var event geminiTypes.StreamEvent
			err := json.Unmarshal(data, &event)
			return &event, err
		},
		func(ctx context.Context, event interface{}) {
			if evt, ok := event.(*geminiTypes.StreamEvent); ok && evt != nil {
				select {
				case <-ctx.Done():
				case output <- evt:
				}
			}
		},
	)
}
