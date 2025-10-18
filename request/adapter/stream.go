package adapter

import (
	"bufio"
	"context"
	"io"
	"strings"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
	"github.com/valyala/fasthttp"
)

// handleStreaming 处理流式请求
func (a *Adapter) handleStreaming(
	ctx context.Context,
	channel *routing.Channel,
	apiReq interface{},
	stream chan<- *coreTypes.Response,
) error {
	// 发送 HTTP 请求
	httpResp, err := a.sendHTTPRequest(channel, apiReq, true)
	if err != nil {
		return err
	}

	// 获取需要释放的响应对象
	var respToRelease *fasthttp.Response
	if resp, ok := httpResp.userData.(*fasthttp.Response); ok {
		respToRelease = resp
	}

	if httpResp.StatusCode != fasthttp.StatusOK {
		return errors.NewWithHTTPStatus(errors.ErrCodeRequestFailed, "API 返回错误状态码", httpResp.StatusCode)
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
				if err != nil {
					if err == io.EOF {
						// 流已结束
						return
					}
					streamErr := errors.Wrap(errors.ErrCodeStreamError, "读取流数据失败", stripErrorHTML(err))
					a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError, streamErr.Error())
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
					// 流式传输正常完成
					return
				}

				// 解析流式响应块
				chunk, err := a.provider.ParseStreamResponse([]byte(data))
				if err != nil {
					parseErr := errors.Wrap(errors.ErrCodeStreamError, "解析流块失败", stripErrorHTML(err)).
						WithContext("data", data)
					a.sendStreamError(ctx, stream, fasthttp.StatusInternalServerError, parseErr.Error())
					return
				}

				// 确保响应块有效后再发送
				if chunk != nil {
					select {
					case <-ctx.Done():
						// 上下文已取消，停止发送响应块
						return
					default:
						stream <- chunk
					}
				}
			}
		}
	}()

	return nil
}

// sendStreamError 向流发送错误信息
func (a *Adapter) sendStreamError(
	ctx context.Context,
	stream chan<- *coreTypes.Response,
	code int,
	message string,
) {
	select {
	case <-ctx.Done():
	default:
		stream <- &coreTypes.Response{
			Choices: []coreTypes.Choice{
				{
					Error: &coreTypes.ErrorResponse{
						Code:    code,
						Message: message,
					},
				},
			},
		}
	}
}
