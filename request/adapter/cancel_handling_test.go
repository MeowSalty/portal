package adapter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
)

type cancelTestProvider struct{}

func (p *cancelTestProvider) Name() string { return "cancel-test" }

func (p *cancelTestProvider) CreateRequest(request *types.RequestContract, channel *routing.Channel) (interface{}, error) {
	return map[string]any{"ok": true}, nil
}

func (p *cancelTestProvider) ParseResponse(variant string, responseData []byte) (*types.ResponseContract, error) {
	return &types.ResponseContract{}, nil
}

func (p *cancelTestProvider) ParseStreamResponse(variant string, ctx types.StreamIndexContext, responseData []byte) ([]*types.StreamEventContract, error) {
	return nil, nil
}

func (p *cancelTestProvider) APIEndpoint(variant string, model string, stream bool, config ...string) string {
	return "/stream"
}

func (p *cancelTestProvider) Headers(key string) map[string]string { return map[string]string{} }

func (p *cancelTestProvider) SupportsStreaming() bool { return true }

func (p *cancelTestProvider) SupportsNative() bool { return true }

func (p *cancelTestProvider) BuildNativeRequest(channel *routing.Channel, payload any) (body any, err error) {
	return payload, nil
}

func (p *cancelTestProvider) ParseNativeResponse(variant string, raw []byte) (any, error) {
	return map[string]any{"ok": true}, nil
}

func (p *cancelTestProvider) ParseNativeStreamEvent(variant string, raw []byte) (any, error) {
	return map[string]any{"raw": string(raw)}, nil
}

func (p *cancelTestProvider) ExtractUsageFromNativeStreamEvent(variant string, event any) *types.ResponseUsage {
	return nil
}

type hookSpy struct {
	firstChunkCount atomic.Int32
	completeCount   atomic.Int32
	errorCount      atomic.Int32
	errorCh         chan error
}

func (h *hookSpy) OnFirstChunk(t time.Time) {
	h.firstChunkCount.Add(1)
}

func (h *hookSpy) OnUsage(u types.Usage) {}

func (h *hookSpy) OnComplete(end time.Time) {
	h.completeCount.Add(1)
}

func (h *hookSpy) OnError(err error) {
	h.errorCount.Add(1)
	select {
	case h.errorCh <- err:
	default:
	}
}

func TestNormalizeCanceled_ContextCanceled(t *testing.T) {
	err := errors.NormalizeCanceled(context.Canceled)

	if !errors.IsCode(err, errors.ErrCodeAborted) {
		t.Fatalf("错误码期望 ABORTED，实际：%v", errors.GetCode(err))
	}

	if got := errors.GetHTTPStatus(err); got != errors.HTTPStatusClientClosedRequest {
		t.Fatalf("HTTP 状态码期望 499，实际：%d", got)
	}

	ctx := errors.GetContext(err)
	if ctx == nil || ctx["error_from"] != "client" {
		t.Fatalf("error_from 期望 client，实际：%v", ctx)
	}
}

func TestHandleNativeStreaming_CancelTriggersOnErrorOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatalf("响应写入器不支持 Flusher")
		}

		_, _ = fmt.Fprintln(w, "data: {\"chunk\":1}")
		flusher.Flush()

		// 保持连接，等待客户端取消
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	a := NewAdapterFromProvider(&cancelTestProvider{})
	channel := &routing.Channel{
		Provider:   "cancel-test",
		BaseURL:    server.URL,
		ModelName:  "m",
		APIKey:     "k",
		APIVariant: "v",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	output := make(chan any, 8)
	hooks := &hookSpy{errorCh: make(chan error, 1)}

	if err := a.handleNativeStreaming(ctx, channel, nil, map[string]any{"x": 1}, output, hooks); err != nil {
		t.Fatalf("handleNativeStreaming 启动失败：%v", err)
	}

	// 等待至少一个事件，确保读取协程已经工作
	select {
	case <-time.After(1 * time.Second):
		t.Fatalf("未在预期时间内收到流事件")
	case _, ok := <-output:
		if !ok {
			t.Fatalf("输出通道被提前关闭")
		}
	}

	cancel()

	select {
	case <-time.After(2 * time.Second):
		t.Fatalf("取消后未收到 OnError 回调")
	case err := <-hooks.errorCh:
		if !errors.IsCode(err, errors.ErrCodeAborted) {
			t.Fatalf("OnError 错误码期望 ABORTED，实际：%v", errors.GetCode(err))
		}
		if got := errors.GetHTTPStatus(err); got != errors.HTTPStatusClientClosedRequest {
			t.Fatalf("OnError HTTP 状态码期望 499，实际：%d", got)
		}
	}

	if hooks.completeCount.Load() != 0 {
		t.Fatalf("取消场景不应触发 OnComplete，实际次数：%d", hooks.completeCount.Load())
	}
}

type streamErrorChunkProvider struct {
	cancelTestProvider
	parseCalls atomic.Int32
}

func (p *streamErrorChunkProvider) ParseNativeStreamEvent(variant string, raw []byte) (any, error) {
	p.parseCalls.Add(1)
	return map[string]any{"raw": string(raw)}, nil
}

func TestHandleNativeStreaming_StreamErrorChunk_BypassNativeEventParser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatalf("响应写入器不支持 Flusher")
		}

		_, _ = fmt.Fprintln(w, `data: {"error":{"type":"rate_limit_error","message":"Concurrency limit exceeded for user, please retry later"}}`)
		flusher.Flush()
	}))
	defer server.Close()

	provider := &streamErrorChunkProvider{}
	a := NewAdapterFromProvider(provider)
	channel := &routing.Channel{
		Provider:   "cancel-test",
		BaseURL:    server.URL,
		ModelName:  "m",
		APIKey:     "k",
		APIVariant: "responses",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	output := make(chan any, 8)
	hooks := &hookSpy{errorCh: make(chan error, 1)}

	if err := a.handleNativeStreaming(ctx, channel, nil, map[string]any{"x": 1}, output, hooks); err != nil {
		t.Fatalf("handleNativeStreaming 启动失败：%v", err)
	}

	select {
	case <-time.After(2 * time.Second):
		t.Fatalf("流错误块场景未收到 OnError 回调")
	case err := <-hooks.errorCh:
		if from := errors.GetErrorFrom(err); from != errors.ErrorFromServer {
			t.Fatalf("GetErrorFrom() = %q, want %q", from, errors.ErrorFromServer)
		}
		if got := errors.GetCode(err); got != errors.ErrCodeRateLimitExceeded {
			t.Fatalf("GetCode() = %s, want %s", got, errors.ErrCodeRateLimitExceeded)
		}
		if errors.HasHTTPStatus(err) {
			t.Fatalf("流错误块不应携带 HTTP 状态码")
		}
		ctxMap := errors.GetContext(err)
		if got, ok := ctxMap["stream_error_chunk"].(bool); !ok || !got {
			t.Fatalf("stream_error_chunk 上下文不符合预期：%+v", ctxMap["stream_error_chunk"])
		}
		if got, ok := ctxMap["http_status_available"].(bool); !ok || got {
			t.Fatalf("http_status_available 上下文不符合预期：%+v", ctxMap["http_status_available"])
		}
	}

	if got := provider.parseCalls.Load(); got != 0 {
		t.Fatalf("ParseNativeStreamEvent 不应被调用，实际调用次数：%d", got)
	}

	if got := hooks.firstChunkCount.Load(); got != 0 {
		t.Fatalf("错误块场景不应触发 OnFirstChunk，实际次数：%d", got)
	}

	if hooks.completeCount.Load() != 0 {
		t.Fatalf("错误块场景不应触发 OnComplete，实际次数：%d", hooks.completeCount.Load())
	}
}
