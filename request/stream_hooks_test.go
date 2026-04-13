package request

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	portalErrors "github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	adapterTypes "github.com/MeowSalty/portal/request/adapter/types"
)

type countingRequestLogRepo struct {
	count int32
}

func (r *countingRequestLogRepo) CreateRequestLog(_ context.Context, _ *RequestLog) error {
	atomic.AddInt32(&r.count, 1)
	return nil
}

func TestRequestLogHooks_FinalizeOnlyOnce(t *testing.T) {
	repo := &countingRequestLogRepo{}
	req := New(repo, logger.NewNopLogger())

	requestLog := &RequestLog{Timestamp: time.Now().Add(-100 * time.Millisecond)}
	hooks := &RequestLogHooks{log: requestLog, request: req}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				hooks.OnComplete(time.Now())
				return
			}
			hooks.OnError(portalErrors.New(portalErrors.ErrCodeInternal, "并发错误"))
		}(i)
	}
	wg.Wait()

	if got := atomic.LoadInt32(&repo.count); got != 1 {
		t.Fatalf("请求日志持久化次数期望 1，实际：%d", got)
	}
}

func TestRequestLogHooks_FirstChunkOnlyOnce(t *testing.T) {
	hooks := &RequestLogHooks{log: &RequestLog{Timestamp: time.Now().Add(-50 * time.Millisecond)}}

	for i := 0; i < 5; i++ {
		hooks.OnFirstChunk(time.Now().Add(time.Duration(i) * time.Millisecond))
	}

	if hooks.log.FirstByteTime == nil {
		t.Fatalf("FirstByteTime 期望被设置")
	}

	first := *hooks.log.FirstByteTime
	hooks.OnFirstChunk(time.Now().Add(100 * time.Millisecond))
	if hooks.log.FirstByteTime == nil || *hooks.log.FirstByteTime != first {
		t.Fatalf("FirstByteTime 不应被重复覆盖")
	}
}

func TestRequestLogHooks_OnStreamFinishedAndOnComplete_CompletedThenDisconnectedSuccess(t *testing.T) {
	repo := &countingRequestLogRepo{}
	req := New(repo, logger.NewNopLogger())
	requestLog := &RequestLog{Timestamp: time.Now().Add(-80 * time.Millisecond), IsStream: true}
	hooks := &RequestLogHooks{log: requestLog, request: req}

	hooks.OnStreamFinished(adapterTypes.StreamFinishInfo{
		CompletionState:  "completed",
		ConnectionStatus: "completed_then_disconnected",
		FinishStatus:     "completed_then_disconnected",
	})
	hooks.OnComplete(time.Now())

	if !hooks.log.Success {
		t.Fatal("completed_then_disconnected 场景 Success 期望为 true")
	}
	if hooks.log.ConnectionStatus == nil || *hooks.log.ConnectionStatus != "completed_then_disconnected" {
		t.Fatalf("ConnectionStatus 期望 completed_then_disconnected，实际：%+v", hooks.log.ConnectionStatus)
	}
	if hooks.log.FinishStatus == nil || *hooks.log.FinishStatus != "completed_then_disconnected" {
		t.Fatalf("FinishStatus 期望 completed_then_disconnected，实际：%+v", hooks.log.FinishStatus)
	}
	if got := atomic.LoadInt32(&repo.count); got != 1 {
		t.Fatalf("请求日志持久化次数期望 1，实际：%d", got)
	}
}

func TestRequestLogHooks_OnError_DeadlineWithoutOutput(t *testing.T) {
	repo := &countingRequestLogRepo{}
	req := New(repo, logger.NewNopLogger())
	requestLog := &RequestLog{Timestamp: time.Now().Add(-120 * time.Millisecond), IsStream: true}
	hooks := &RequestLogHooks{log: requestLog, request: req}

	hooks.OnError(portalErrors.NormalizeCanceled(context.DeadlineExceeded))

	if hooks.log.Success {
		t.Fatal("超时场景 Success 期望为 false")
	}
	if hooks.log.CompletionState == nil || *hooks.log.CompletionState != "not_completed" {
		t.Fatalf("CompletionState 期望 not_completed，实际：%+v", hooks.log.CompletionState)
	}
	if hooks.log.FinishStatus == nil || *hooks.log.FinishStatus != "timed_out" {
		t.Fatalf("FinishStatus 期望 timed_out，实际：%+v", hooks.log.FinishStatus)
	}
	if hooks.log.CancelSource == nil || *hooks.log.CancelSource != "deadline" {
		t.Fatalf("CancelSource 期望 deadline，实际：%+v", hooks.log.CancelSource)
	}
	if got := atomic.LoadInt32(&repo.count); got != 1 {
		t.Fatalf("请求日志持久化次数期望 1，实际：%d", got)
	}
}
