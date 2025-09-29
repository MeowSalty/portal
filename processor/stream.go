package processor

import (
	"context"
	"log/slog"
	"time"

	"github.com/MeowSalty/portal/types"
	"github.com/valyala/fasthttp"
)

// ProcessChatCompletionStream 处理流式聊天完成请求
func (p *RequestProcessor) ProcessChatCompletionStream(ctx context.Context, request *types.Request, doneFunc func()) (<-chan *types.Response, error) {
	// 准备通道
	allChannels, err := p.prepareChannels(ctx, request.Model)
	if err != nil {
		return nil, err
	}

	// 创建响应流
	wrappedStream := make(chan *types.Response)

	// 创建重试上下文
	retryCtx := p.createRetryContext(allChannels, time.Now())

	// 启动流处理协程
	go p.processStreamWithRetry(ctx, request, wrappedStream, retryCtx, doneFunc)

	return wrappedStream, nil
}

// processStreamWithRetry 处理带重试的流请求
func (p *RequestProcessor) processStreamWithRetry(
	ctx context.Context,
	request *types.Request,
	output chan<- *types.Response,
	retryCtx *retryContext,
	doneFunc func(),
) {
	defer close(output)
	defer doneFunc()

	for len(retryCtx.allChannels) > 0 {
		// 选择通道
		selectedChannel, err := p.selectChannel(ctx, retryCtx.allChannels, time.Now())
		if err != nil {
			p.sendErrorResponse(output, fasthttp.StatusInternalServerError, sanitizeError(err).Error())
			return
		}

		// 尝试处理流
		if p.tryProcessStream(ctx, request, output, retryCtx, selectedChannel) {
			return // 成功完成
		}

		// 失败，移除通道并继续
		retryCtx.allChannels = p.removeChannel(retryCtx, selectedChannel)
	}

	// 所有通道都失败
	p.sendErrorResponse(output, fasthttp.StatusServiceUnavailable, "所有通道都已尝试且失败")
}

// tryProcessStream 尝试处理单个流
func (p *RequestProcessor) tryProcessStream(
	ctx context.Context,
	request *types.Request,
	output chan<- *types.Response,
	retryCtx *retryContext,
	selectedChannel *types.Channel,
) bool {
	// 获取适配器
	adapter, err := p.getAdapter(selectedChannel.Platform.Format)
	if err != nil {
		p.Logger.Error("适配器未找到",
			slog.String("格式", selectedChannel.Platform.Format),
			slog.Any("error", err))
		return false
	}

	// 创建流上下文
	streamCtx := &streamContext{
		request:         request,
		selectedChannel: selectedChannel,
		requestStart:    time.Now(),
		statOptions: &types.RequestStat{
			Timestamp:   retryCtx.startTime,
			RequestType: "stream",
			ModelName:   request.Model,
			ChannelInfo: types.ChannelInfo{
				PlatformID: selectedChannel.Platform.ID,
				APIKeyID:   selectedChannel.APIKey.ID,
				ModelID:    selectedChannel.Model.ID,
			},
		},
	}

	// 创建流
	stream, err := adapter.ChatCompletionStream(ctx, request, selectedChannel)
	if err != nil {
		p.handleStreamInitError(streamCtx, err)
		return false
	}

	// 处理流数据
	return p.handleStreamData(ctx, stream, output, streamCtx)
}

// handleStreamData 处理流数据
func (p *RequestProcessor) handleStreamData(
	ctx context.Context,
	input <-chan *types.Response,
	output chan<- *types.Response,
	streamCtx *streamContext,
) bool {
	firstByteRecorded := false

	for response := range input {
		// 检查错误
		if err := p.checkResponseError(response); err != nil {
			streamCtx.streamErr = err
			p.logStreamError(streamCtx)
			p.recordFailedStats(streamCtx)
			return false
		}

		// 记录首字节时间
		if !firstByteRecorded {
			now := time.Now()
			streamCtx.firstByteTime = &now
			firstByteRecorded = true
		}

		// 发送响应
		if !p.sendResponse(ctx, output, response, streamCtx) {
			return false
		}
	}

	// 流成功完成
	if streamCtx.streamErr == nil {
		p.recordSuccessStats(streamCtx)
		return true
	}

	return false
}

// sendResponse 发送响应到输出通道
func (p *RequestProcessor) sendResponse(
	ctx context.Context,
	output chan<- *types.Response,
	response *types.Response,
	streamCtx *streamContext,
) bool {
	select {
	case output <- response:
		return true
	case <-ctx.Done():
		streamCtx.streamErr = ctx.Err()
		p.logClientDisconnect(streamCtx)
		p.recordFailedStats(streamCtx)
		return false
	}
}

// handleStreamInitError 处理流初始化错误
func (p *RequestProcessor) handleStreamInitError(streamCtx *streamContext, err error) {
	err = sanitizeError(err)
	p.Logger.Error("流式请求初始化失败",
		slog.String("model", streamCtx.request.Model),
		slog.String("platform", streamCtx.selectedChannel.Platform.Name),
		slog.Any("error", err),
		slog.String("request_type", "stream"))

	errorMsg := err.Error()
	streamCtx.streamErr = err
	p.recordStats(streamCtx.statOptions, streamCtx.requestStart, nil, false, &errorMsg, streamCtx.selectedChannel)
}

// logStreamError 记录流错误
func (p *RequestProcessor) logStreamError(streamCtx *streamContext) {
	p.Logger.Error("流式传输过程中发生错误",
		slog.String("model", streamCtx.request.Model),
		slog.String("platform", streamCtx.selectedChannel.Platform.Name),
		slog.Any("error", streamCtx.streamErr),
		slog.String("request_type", "stream"))
}

// logClientDisconnect 记录客户端断开连接
func (p *RequestProcessor) logClientDisconnect(streamCtx *streamContext) {
	p.Logger.Error("客户端断开连接",
		slog.String("model", streamCtx.request.Model),
		slog.String("platform", streamCtx.selectedChannel.Platform.Name),
		slog.Any("error", streamCtx.streamErr))
}

// recordFailedStats 记录失败的统计信息
func (p *RequestProcessor) recordFailedStats(streamCtx *streamContext) {
	errorMsg := streamCtx.streamErr.Error()
	p.recordStats(streamCtx.statOptions, streamCtx.requestStart, streamCtx.firstByteTime, false, &errorMsg, streamCtx.selectedChannel)
}

// recordSuccessStats 记录成功的统计信息
func (p *RequestProcessor) recordSuccessStats(streamCtx *streamContext) {
	p.recordStats(streamCtx.statOptions, streamCtx.requestStart, streamCtx.firstByteTime, true, nil, streamCtx.selectedChannel)
}

// sendErrorResponse 发送错误响应
func (p *RequestProcessor) sendErrorResponse(output chan<- *types.Response, code int, message string) {
	resp := p.errorResponsePool.Get().(*types.Response)
	defer p.errorResponsePool.Put(resp)

	resp.Choices[0].Error.Code = code
	resp.Choices[0].Error.Message = message

	select {
	case output <- resp:
	default:
		// 通道已关闭或满
	}
}
