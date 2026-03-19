package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/routing"
)

var (
	sharedClient *http.Client
	clientOnce   sync.Once
)

// httpResponse 统一的 HTTP 响应封装
type httpResponse struct {
	StatusCode  int           // HTTP 状态码
	ContentType string        // 响应的 Content-Type 头部
	Body        []byte        // 非流式响应的完整响应体（已读取）
	BodyStream  io.Reader     // 流式响应的响应体流（需持续读取）
	IsStream    bool          // 标记是否为流式响应
	body        io.ReadCloser // 存储 resp.Body 用于流式场景的延迟关闭
}

var responseBodyPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 8*1024)
		return &b
	},
}

// readResponseBody 使用 pool 和 Content-Length 预分配读取响应体
func readResponseBody(resp *http.Response) ([]byte, error) {
	bufp := responseBodyPool.Get().(*[]byte)
	buf := (*bufp)[:0]

	if resp.ContentLength > 0 && resp.ContentLength <= 10*1024*1024 {
		if cap(buf) < int(resp.ContentLength) {
			buf = make([]byte, 0, resp.ContentLength)
		}
	}

	buf, err := appendReader(buf, resp.Body)
	if err != nil {
		*bufp = buf
		responseBodyPool.Put(bufp)
		return nil, err
	}

	result := make([]byte, len(buf))
	copy(result, buf)

	*bufp = buf
	responseBodyPool.Put(bufp)
	return result, nil
}

func appendReader(buf []byte, r io.Reader) ([]byte, error) {
	for {
		if len(buf) == cap(buf) {
			buf = append(buf, 0)[:len(buf)]
		}
		n, err := r.Read(buf[len(buf):cap(buf)])
		buf = buf[:len(buf)+n]
		if err != nil {
			if err == io.EOF {
				return buf, nil
			}
			return buf, err
		}
	}
}

// getSharedHTTPClient 返回共享的 HTTP 客户端单例
//
// 使用 sync.Once 确保只初始化一次。http.Client 是线程安全的，
// 共享单例可以复用底层 TCP/TLS 连接池，避免每次请求创建新连接。
func getSharedHTTPClient() *http.Client {
	clientOnce.Do(func() {
		sharedClient = &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          1000,
				MaxIdleConnsPerHost:   100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				ReadBufferSize:        32 * 1024,
				WriteBufferSize:       32 * 1024,
				DisableCompression:    true,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	})
	return sharedClient
}

// sendHTTPRequest 发送 HTTP 请求
func (a *Adapter) sendHTTPRequest(
	ctx context.Context,
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
	url := joinBaseURL(
		channel.BaseURL,
		a.provider.APIEndpoint(channel.APIVariant, channel.ModelName, isStream, channel.APIEndpointConfig),
	)

	// 记录调试日志：请求 URL 和请求体
	log.Debug("HTTP 请求准备完成",
		"url", url,
		"is_stream", isStream,
		"request_body", string(jsonData),
	)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "创建 HTTP 请求失败", err)
	}

	// 设置请求头
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
	resp, err := a.client.Do(req)
	if err != nil {
		log.Error("HTTP 请求失败",
			"url", url,
			"is_stream", isStream,
			"error", err,
		)
		return nil, errors.Wrap(errors.ErrCodeUnavailable, "HTTP 请求失败", stripErrorHTML(err)).
			WithContext("error_from", string(errors.ErrorFromGateway))
	}

	// 记录调试日志：HTTP 响应状态
	log.Debug("HTTP 请求已发送",
		"url", url,
		"is_stream", isStream,
		"status_code", resp.StatusCode,
		"content_type", resp.Header.Get("Content-Type"),
	)

	// 根据是否流式请求返回不同的响应体
	httpResp := &httpResponse{
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		IsStream:    isStream,
	}

	if isStream {
		// 流式请求：保留 resp.Body 供后续读取，由调用者负责关闭
		if resp.Body == nil {
			log.Error("流式响应体为空", "url", url)
			return nil, errors.New(errors.ErrCodeStreamError, "流式响应体为空")
		}
		httpResp.BodyStream = resp.Body
		httpResp.body = resp.Body
		log.Debug("流式响应已准备", "url", url)
	} else {
		// 非流式请求：读取完整响应体后关闭
		defer resp.Body.Close()
		body, err := readResponseBody(resp)
		if err != nil {
			log.Error("读取响应体失败", "url", url, "error", err)
			return nil, errors.Wrap(errors.ErrCodeInternal, "读取响应体失败", err)
		}
		httpResp.Body = body
		log.Debug("非流式响应已准备", "url", url, "body_size", len(body))
	}

	return httpResp, nil
}
