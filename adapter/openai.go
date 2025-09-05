package adapter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/MeowSalty/portal/adapter/openai"
	"github.com/MeowSalty/portal/adapter/openai/types"
	coreTypes "github.com/MeowSalty/portal/types"
	"github.com/valyala/fasthttp"
)

// init 函数在包被导入时自动注册适配器
func init() {
	RegisterAdapterFactory("OpenAI", func(logger *slog.Logger) coreTypes.Adapter {
		return NewOpenAIAdapter(logger)
	})
}

// OpenAIAdapter 实现了适用于 OpenAI 兼容 API 的适配器接口。
type OpenAIAdapter struct {
	logger *slog.Logger
}

// NewOpenAIAdapter 创建一个新的 OpenAI 适配器。
func NewOpenAIAdapter(logger *slog.Logger) *OpenAIAdapter {
	if logger == nil {
		logger = slog.Default()
	}
	return &OpenAIAdapter{
		logger: logger.WithGroup("openai"),
	}
}

// ChatCompletion 执行非流式聊天完成请求。
func (a *OpenAIAdapter) ChatCompletion(
	ctx context.Context,
	request *coreTypes.Request,
	channel *coreTypes.Channel,
) (*coreTypes.Response, error) {
	a.logger.Info(
		"调用 OpenAIAdapter 的 ChatCompletion (非流式)",
		slog.String("model", request.Model),
		slog.String("platform_base_url", channel.Platform.BaseURL),
	)

	openAIReq, err := a.buildChatRequest(request, channel)
	if err != nil {
		return nil, err
	}
	openAIReq.Stream = false

	code, body, errs := a.sendRequest(ctx, channel, "/v1/chat/completions", openAIReq)
	if len(errs) > 0 {
		a.logger.Error("发送 HTTP 请求失败", slog.Any("error", errs[0]))
		return nil, fmt.Errorf("发送 HTTP 请求失败：%w", errs[0])
	}
	if code != fasthttp.StatusOK {
		a.logger.Error("API 返回错误状态码", slog.Int("status_code", code), slog.String("response_body", string(body)))
		return nil, fmt.Errorf("API 返回错误状态码: %d, 响应内容: %s", code, string(body))
	}

	var openAIResp types.ChatCompletionResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		a.logger.Error("解析响应体失败", slog.Any("error", err), slog.String("response_body", string(body)))
		return nil, fmt.Errorf("解析响应体失败：%w", err)
	}

	// 转换 OpenAI 响应为核心响应格式
	coreResp := openai.ChatCompletionResponseToResponse(&openAIResp)

	a.logger.Info("ChatCompletion 请求成功完成",
		slog.String("model", request.Model),
		slog.String("response_id", coreResp.ID))

	return coreResp, nil
}

// ChatCompletionStream 执行流式聊天完成请求。
func (a *OpenAIAdapter) ChatCompletionStream(
	ctx context.Context,
	request *coreTypes.Request,
	channel *coreTypes.Channel,
) (<-chan *coreTypes.Response, error) {
	logger := a.logger.With(
		"Adapter", "OpenAI",
		slog.Bool("Streaming", true),
		slog.String("model", fmt.Sprintf("%s(%s)", channel.Model.Name, request.Model)),
		slog.String("platform_base_url", channel.Platform.BaseURL),
	)
	logger.Info("开始调用")

	openAIReq, err := a.buildChatRequest(request, channel)
	if err != nil {
		return nil, err
	}
	openAIReq.Stream = true

	// 创建用于返回的流
	stream := make(chan *coreTypes.Response, 1024)

	// 启动一个 goroutine 来处理整个流式请求
	go a.handleStreamingRequest(ctx, logger, channel, openAIReq, stream)

	return stream, nil
}

// handleStreamingRequest 处理流式请求的具体逻辑
func (a *OpenAIAdapter) handleStreamingRequest(
	ctx context.Context,
	logger *slog.Logger,
	channel *coreTypes.Channel,
	openAIReq *types.ChatCompletionRequest,
	stream chan<- *coreTypes.Response,
) {
	defer close(stream)

	// 创建请求
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(channel.Platform.BaseURL + "/v1/chat/completions")
	req.Header.SetMethod("POST")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+channel.APIKey.Value)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// 设置请求体
	jsonData, err := json.Marshal(openAIReq)
	if err != nil {
		a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError, fmt.Sprintf("序列化请求体失败：%s", err.Error()))
		return
	}
	req.SetBody(jsonData)

	// 创建响应
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// 创建 HTTP 客户端
	client := &fasthttp.Client{
		StreamResponseBody:            true,
		DisableHeaderNamesNormalizing: true,
	}

	// 设置超时
	if deadline, ok := ctx.Deadline(); ok {
		client.ReadTimeout = time.Until(deadline)
		client.WriteTimeout = time.Until(deadline)
	}

	// 发起请求
	err = client.Do(req, resp)
	if err != nil {
		logger.Error("发送 HTTP 请求失败", slog.Any("error", err))
		return
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		body := resp.Body()
		logger.Error("API 返回错误状态码",
			slog.Int("status_code", resp.StatusCode()),
			slog.String("response_body", string(body)))
		return
	}

	// 获取响应体的 Reader
	bodyReader := resp.BodyStream()
	if bodyReader == nil {
		logger.Error("响应体流为空")
		return
	}

	defer func() {
		if closer, ok := bodyReader.(io.Closer); ok {
			closer.Close()
		}
	}()

	reader := bufio.NewReader(bodyReader)
	a.processStreamingResponse(ctx, logger, reader, stream)
}

// processStreamingResponse 处理流式响应数据
func (a *OpenAIAdapter) processStreamingResponse(
	ctx context.Context,
	logger *slog.Logger,
	reader *bufio.Reader,
	stream chan<- *coreTypes.Response,
) {
	for {
		select {
		case <-ctx.Done():
			logger.Info("上下文已取消，停止处理流")
			a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError, "上下文已取消")
			return
		default:
			// 读取一行数据
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					logger.Info("流已结束")
					a.sendStreamError(ctx, stream, fasthttp.StatusOK, "流已结束")
					return
				}
				errMsg := fmt.Sprintf("读取行时发生错误: %v", err)
				logger.Error(errMsg)
				a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError, errMsg)
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// 处理 SSE 格式数据
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")

			// 检查是否是结束标记
			if data == "[DONE]" {
				logger.Info("流式传输正常结束")
				return // 正常结束
			}

			var chunk types.ChatCompletionResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				errMsg := fmt.Sprintf("解析流数据块失败：%v。数据: %s", err, data)
				logger.Error(errMsg)
				a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError, errMsg)
				return
			}

			// 转换数据块为核心格式
			coreChunk := openai.ChatCompletionResponseToResponse(&chunk)

			// 尝试发送数据块到流
			select {
			case stream <- coreChunk:
			case <-ctx.Done():
				logger.Info("上下文已取消，停止处理流")
				return
			}
		}
	}
}

// sendStreamError 向流中发送错误信息
func (a *OpenAIAdapter) sendStreamError(ctx context.Context, stream chan<- *coreTypes.Response, code int, message string) {
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

// --- 辅助函数 ---

func (a *OpenAIAdapter) buildChatRequest(request *coreTypes.Request, channel *coreTypes.Channel) (*types.ChatCompletionRequest, error) {
	openAIReq := &types.ChatCompletionRequest{
		Model:    channel.Model.Name,
		Messages: make([]types.RequestMessage, len(request.Messages)),
	}

	// 处理流参数
	if request.Stream != nil {
		openAIReq.Stream = *request.Stream
	}

	// 处理 Temperature 参数
	if request.Temperature != nil {
		openAIReq.Temperature = *request.Temperature
	}

	// 处理 TopP 参数
	if request.TopP != nil {
		openAIReq.TopP = *request.TopP
	}

	for i, msg := range request.Messages {
		// 直接使用 msg.Content，现在它可以是字符串或数组
		openAIReq.Messages[i] = types.RequestMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return openAIReq, nil
}

func (a *OpenAIAdapter) sendRequest(ctx context.Context, channel *coreTypes.Channel, path string, payload interface{}) (int, []byte, []error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, []error{fmt.Errorf("序列化请求体失败：%w", err)}
	}

	url := channel.Platform.BaseURL + path

	// 创建 fasthttp 请求
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// 设置请求参数
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+channel.APIKey.Value)
	req.SetBody(jsonData)

	// 创建客户端
	client := &fasthttp.Client{}

	// 设置超时
	if deadline, ok := ctx.Deadline(); ok {
		client.ReadTimeout = time.Until(deadline)
		client.WriteTimeout = time.Until(deadline)
	}

	// 发送请求
	err = client.Do(req, resp)
	if err != nil {
		return 0, nil, []error{fmt.Errorf("发送 HTTP 请求失败：%w", err)}
	}

	// 获取响应状态码和内容
	code := resp.StatusCode()
	body := resp.Body()

	// 记录请求结果
	if code != fasthttp.StatusOK {
		a.logger.Warn("HTTP 请求返回非成功状态码",
			slog.Int("status_code", code),
			slog.String("url", url))
	} else {
		a.logger.Debug("HTTP 请求成功",
			slog.Int("status_code", code),
			slog.String("url", url))
	}

	return code, body, nil
}
