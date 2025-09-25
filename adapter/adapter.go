package adapter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	coreTypes "github.com/MeowSalty/portal/types"
	"github.com/valyala/fasthttp"
)

// Provider 定义 AI 提供商的接口
type Provider interface {
	// Name 返回提供商名称
	Name() string

	// CreateRequest 将核心请求转换为提供商特定请求
	CreateRequest(request *coreTypes.Request, channel *coreTypes.Channel) (interface{}, error)

	// ParseResponse 解析提供商响应并转换为核心响应
	ParseResponse(responseData []byte) (*coreTypes.Response, error)

	// ParseStreamResponse 解析提供商流式响应并转换为核心响应
	ParseStreamResponse(responseData []byte) (*coreTypes.Response, error)

	// APIEndpoint 返回 API 端点路径
	APIEndpoint() string

	// Headers 返回特定于提供商的 HTTP 头
	Headers() map[string]string

	// SupportsStreaming 返回是否支持流式传输
	SupportsStreaming() bool
}

// HTTPClient 统一的 HTTP 客户端
type HTTPClient struct {
	client *fasthttp.Client
	logger *slog.Logger
}

// HTTPResponse 统一的 HTTP 响应封装
type HTTPResponse struct {
	StatusCode int         // HTTP 状态码
	Body       []byte      // 非流式响应的完整响应体（已读取）
	BodyStream io.Reader   // 流式响应的响应体流（需持续读取）
	IsStream   bool        // 标记是否为流式响应
	userData   interface{} // 内部使用，用于存储原始响应对象（如*fasthttp.Response）以便正确释放资源
}

// NewHTTPClient 创建新的 HTTP 客户端
func NewHTTPClient(logger *slog.Logger) *HTTPClient {
	if logger == nil {
		logger = slog.Default()
	}

	client := &fasthttp.Client{
		StreamResponseBody:            true,
		DisableHeaderNamesNormalizing: true,
		// ReadTimeout:                   60 * time.Second,
		// WriteTimeout:                  30 * time.Second,
		// MaxIdleConnDuration:           60 * time.Second,
		// MaxConnDuration:               10 * time.Minute,
		// MaxConnsPerHost:               512,
		// Dial: (&fasthttp.TCPDialer{
		// 	Concurrency:      4096,
		// 	DNSCacheDuration: time.Hour,
		// }).Dial,
	}

	return &HTTPClient{
		client: client,
		logger: logger,
	}
}

// Adapter 统一适配器实现
type Adapter struct {
	client   *HTTPClient
	provider Provider
	logger   *slog.Logger
}

// NewAdapterFromProvider 从 Provider 创建适配器实例
func NewAdapterFromProvider(provider Provider, logger *slog.Logger) *Adapter {
	if logger == nil {
		logger = slog.Default()
	}

	return &Adapter{
		client:   NewHTTPClient(logger.WithGroup("http")),
		provider: provider,
		logger:   logger.WithGroup(provider.Name()),
	}
}

// ChatCompletion 执行聊天完成请求
func (a *Adapter) ChatCompletion(
	ctx context.Context,
	request *coreTypes.Request,
	channel *coreTypes.Channel,
) (*coreTypes.Response, error) {
	a.logger.Info("开始聊天完成请求",
		slog.String("model", request.Model),
		slog.String("provider", a.provider.Name()),
		slog.Int("messages_count", len(request.Messages)),
	)

	// 创建提供商特定请求
	apiReq, err := a.provider.CreateRequest(request, channel)
	if err != nil {
		a.logger.Error("创建请求失败",
			slog.String("model", request.Model),
			slog.String("provider", a.provider.Name()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("创建请求失败：%w", err)
	}

	// 发送请求
	httpResp, err := a.sendHTTPRequest(channel, apiReq, false)
	if err != nil {
		a.logger.Error("发送 HTTP 请求失败",
			slog.String("model", request.Model),
			slog.String("provider", a.provider.Name()),
			slog.String("error", err.Error()))
		return nil, err
	}

	// 检查 HTTP 状态码
	if httpResp.StatusCode != fasthttp.StatusOK {
		err := a.handleHTTPError("API 请求错误", httpResp.StatusCode, httpResp.Body)
		a.logger.Error("API 请求返回错误状态码",
			slog.String("model", request.Model),
			slog.String("provider", a.provider.Name()),
			slog.Int("status_code", httpResp.StatusCode),
			slog.String("response_body", string(httpResp.Body)))
		return nil, err
	}

	// 解析响应
	response, err := a.provider.ParseResponse(httpResp.Body)
	if err != nil {
		err := a.handleParseError("响应解析错误", err, httpResp.Body)
		a.logger.Error("解析响应失败",
			slog.String("model", request.Model),
			slog.String("provider", a.provider.Name()),
			slog.String("error", err.Error()),
			slog.String("response_body", string(httpResp.Body)))
		return nil, err
	}

	a.logger.Info("聊天完成请求成功",
		slog.String("model", request.Model),
		slog.String("provider", a.provider.Name()),
		slog.String("response_id", response.ID),
		slog.Int("response_choices", len(response.Choices)),
	)

	return response, nil
}

// ChatCompletionStream 执行流式聊天完成请求
func (a *Adapter) ChatCompletionStream(
	ctx context.Context,
	request *coreTypes.Request,
	channel *coreTypes.Channel,
) (<-chan *coreTypes.Response, error) {
	if !a.provider.SupportsStreaming() {
		return nil, fmt.Errorf("提供商 %s 不支持流式传输", a.provider.Name())
	}

	logger := a.logger.With(
		slog.Bool("streaming", true),
		slog.String("model", request.Model),
		slog.Int("messages_count", len(request.Messages)),
	)
	logger.Info("开始流式聊天完成请求")

	// 创建提供商特定请求
	apiReq, err := a.provider.CreateRequest(request, channel)
	if err != nil {
		logger.Error("创建流式请求失败",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("创建请求失败：%w", err)
	}

	// 创建响应通道
	stream := make(chan *coreTypes.Response, 1024)

	// 启动流式处理协程
	go a.handleStreaming(ctx, logger, channel, apiReq, stream)

	return stream, nil
}

// sendHTTPRequest 发送 HTTP 请求（通用实现）
func (a *Adapter) sendHTTPRequest(
	channel *coreTypes.Channel,
	payload interface{},
	isStream bool,
) (*HTTPResponse, error) {
	// 序列化请求体
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败：%w", err)
	}

	// 构建 URL
	url := channel.Platform.BaseURL + a.provider.APIEndpoint()

	// 创建请求和响应对象
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	// 注意：不能在这里释放 resp，因为流式响应需要保持有效
	// defer fasthttp.ReleaseResponse(resp) // 已移除

	// 设置请求参数
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+channel.APIKey.Value)

	// 添加提供商特定头部
	for key, value := range a.provider.Headers() {
		req.Header.Set(key, value)
	}

	// 流式请求的特殊头部
	if isStream {
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
	}

	req.SetBody(jsonData)

	// 发送请求
	err = a.client.client.Do(req, resp)
	if err != nil {
		// 发生错误时释放 response 对象
		fasthttp.ReleaseResponse(resp)
		return nil, fmt.Errorf("HTTP 请求失败：%w", err)
	}

	// 根据是否流式请求返回不同的响应体
	httpResp := &HTTPResponse{
		StatusCode: resp.StatusCode(),
		IsStream:   isStream,
		userData:   resp, // 总是存储 resp 对象以便后续释放
	}

	if isStream {
		// 流式请求返回 BodyStream
		bodyStream := resp.BodyStream()
		if bodyStream == nil {
			// 如果 BodyStream 为 nil，释放 response 并返回错误
			fasthttp.ReleaseResponse(resp)
			return nil, fmt.Errorf("流式响应体为空")
		}
		httpResp.BodyStream = bodyStream
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
	}

	return httpResp, nil
}

// handleStreaming 处理流式请求
func (a *Adapter) handleStreaming(
	ctx context.Context,
	logger *slog.Logger,
	channel *coreTypes.Channel,
	apiReq interface{},
	stream chan<- *coreTypes.Response,
) {
	defer close(stream)

	// 确保在函数退出时释放资源
	var respToRelease *fasthttp.Response
	defer func() {
		if respToRelease != nil {
			fasthttp.ReleaseResponse(respToRelease)
		}
	}()

	// 发送 HTTP 请求（流式版本）
	httpResp, err := a.sendHTTPRequest(channel, apiReq, true)
	if err != nil {
		a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError,
			fmt.Sprintf("发送HTTP请求失败: %s", err.Error()))
		return
	}

	// 获取需要释放的响应对象
	if resp, ok := httpResp.userData.(*fasthttp.Response); ok {
		respToRelease = resp
	}

	if httpResp.StatusCode != fasthttp.StatusOK {
		a.sendStreamError(ctx, stream, httpResp.StatusCode,
			fmt.Sprintf("API 返回错误状态码：%d，响应体：%s", httpResp.StatusCode, string(httpResp.Body)))
		return
	}

	// 检查 BodyStream 是否为 nil
	if httpResp.BodyStream == nil {
		a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError,
			"流式响应体为空")
		return
	}

	// 处理流式响应
	a.processStream(ctx, logger, httpResp.BodyStream, stream)
}

// processStream 处理流式响应数据
func (a *Adapter) processStream(
	ctx context.Context,
	logger *slog.Logger,
	bodyStream io.Reader,
	stream chan<- *coreTypes.Response,
) {
	reader := bufio.NewReaderSize(bodyStream, 4096) // 使用更大的缓冲区提高性能

	for {
		select {
		case <-ctx.Done():
			logger.Info("上下文已取消，停止流处理")
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					logger.Info("流已结束")
					return
				}
				a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError,
					fmt.Sprintf("读取流数据失败: %v", err))
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				logger.Info("流式传输正常完成")
				return
			}

			// 解析流式响应块
			chunk, err := a.provider.ParseStreamResponse([]byte(data))
			if err != nil {
				logger.Error("解析流块失败",
					slog.String("data", data),
					slog.String("error", err.Error()))
				a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError,
					fmt.Sprintf("解析流块失败: %v", err))
				return
			}

			// 确保响应块有效后再发送
			if chunk != nil {
				select {
				case stream <- chunk:
				case <-ctx.Done():
					logger.Info("上下文已取消，停止发送响应块")
					return
				}
			}
		}
	}
}

// sendStreamError 向流发送错误信息
func (a *Adapter) sendStreamError(
	ctx context.Context,
	stream chan<- *coreTypes.Response,
	code int,
	message string,
) {
	select {
	case stream <- &coreTypes.Response{
		Choices: []coreTypes.Choice{
			{
				Error: &coreTypes.ErrorResponse{
					Code:    code,
					Message: message,
				},
			},
		},
	}:
	case <-ctx.Done():
	}
}

// handleHTTPError 处理 HTTP 错误
func (a *Adapter) handleHTTPError(operation string, statusCode int, body []byte) error {
	a.logger.Error("HTTP 操作失败",
		slog.String("operation", operation),
		slog.Int("status_code", statusCode),
		slog.String("response_body", string(body)))

	// 尝试解析错误响应体以提供更多上下文
	var errResp coreTypes.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
		return fmt.Errorf("HTTP错误 %d: %s", statusCode, errResp.Message)
	}

	return fmt.Errorf("HTTP错误 %d: %s", statusCode, string(body))
}

// handleParseError 处理解析错误
func (a *Adapter) handleParseError(operation string, err error, body []byte) error {
	a.logger.Error("解析响应失败",
		slog.String("operation", operation),
		slog.Any("error", err),
		slog.String("response_body", string(body)))
	return fmt.Errorf("解析响应失败：%w", err)
}
