package adapter

import (
	"encoding/json"
	"io"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/routing"
	"github.com/valyala/fasthttp"
)

// httpResponse 统一的 HTTP 响应封装
type httpResponse struct {
	StatusCode  int         // HTTP 状态码
	ContentType string      // 新增：响应的 Content-Type 头部
	Body        []byte      // 非流式响应的完整响应体（已读取）
	BodyStream  io.Reader   // 流式响应的响应体流（需持续读取）
	IsStream    bool        // 标记是否为流式响应
	userData    interface{} // 内部使用，用于存储原始响应对象（如*fasthttp.Response）以便正确释放资源
}

// newHTTPClient 创建新的 HTTP 客户端
func newHTTPClient() *fasthttp.Client {
	client := &fasthttp.Client{
		StreamResponseBody:            true,
		DisableHeaderNamesNormalizing: true,
	}

	return client
}

// sendHTTPRequest 发送 HTTP 请求
func (a *Adapter) sendHTTPRequest(
	channel *routing.Channel,
	headers map[string]string,
	payload interface{},
	isStream bool,
) (*httpResponse, error) {
	log := logger.Default().WithGroup("http")

	// 序列化请求体
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "序列化请求体失败", err)
	}

	// 构建 URL
	url := channel.BaseURL + a.provider.APIEndpoint(channel.ModelName, isStream, channel.APIEndpointConfig)

	// 记录调试日志：请求 URL 和请求体
	log.Debug("HTTP 请求准备完成",
		"url", url,
		"is_stream", isStream,
		"request_body", string(jsonData),
	)

	// 创建请求和响应对象
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()

	// 设置请求参数
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")
	req.Header.Set("Content-Type", "application/json")

	// 添加提供商特定头部（包括身份验证头部）
	providerHeaders := a.provider.Headers(channel.APIKey)
	for key, value := range providerHeaders {
		req.Header.Set(key, value)
	}

	// 流式请求的特殊头部
	if isStream {
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
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

	// 记录调试日志：完整的请求头部
	log.Debug("HTTP 请求头部信息",
		"url", url,
		"method", "POST",
		"is_stream", isStream,
		"provider_headers", providerHeaders,
		"custom_headers", headers,
		"channel_headers", channel.CustomHeaders,
	)

	// 发送请求
	err = a.client.Do(req, resp)
	if err != nil {
		// 发生错误时释放 response 对象
		fasthttp.ReleaseResponse(resp)
		log.Error("HTTP 请求失败",
			"url", url,
			"is_stream", isStream,
			"error", err,
		)
		return nil, errors.Wrap(errors.ErrCodeUnavailable, "HTTP 请求失败", stripErrorHTML(err))
	}

	// 记录调试日志：HTTP 响应状态
	log.Debug("HTTP 请求已发送",
		"url", url,
		"is_stream", isStream,
		"status_code", resp.StatusCode(),
		"content_type", string(resp.Header.ContentType()),
	)

	// 根据是否流式请求返回不同的响应体
	httpResp := &httpResponse{
		StatusCode:  resp.StatusCode(),
		ContentType: string(resp.Header.ContentType()), // 新增：提取 Content-Type 头部
		IsStream:    isStream,
		userData:    resp, // 总是存储 resp 对象以便后续释放
	}

	if isStream {
		// 流式请求返回 BodyStream
		bodyStream := resp.BodyStream()
		if bodyStream == nil {
			// 如果 BodyStream 为 nil，释放 response 并返回错误
			fasthttp.ReleaseResponse(resp)
			log.Error("流式响应体为空", "url", url)
			return nil, errors.New(errors.ErrCodeStreamError, "流式响应体为空")
		}
		httpResp.BodyStream = bodyStream
		log.Debug("流式响应已准备", "url", url)
	} else {
		// 非流式请求返回 Body，并释放 response 对象
		body := make([]byte, len(resp.Body()))
		copy(body, resp.Body())
		httpResp.Body = body
		// 确保非流式响应的 BodyStream 为空
		httpResp.BodyStream = nil
		// 非流式情况下立即释放
		fasthttp.ReleaseResponse(resp)
		httpResp.userData = nil
		log.Debug("非流式响应已准备", "url", url, "body_size", len(body))
	}

	return httpResp, nil
}
