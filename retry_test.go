package portal

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/routing"
	"github.com/MeowSalty/portal/session"
)

type retryLogEntry struct {
	level string
	msg   string
	args  []any
}

type retryLogStore struct {
	mu      sync.Mutex
	entries []retryLogEntry
}

type retryCapturedLogger struct {
	store *retryLogStore
}

func (l *retryCapturedLogger) append(level, msg string, args ...any) {
	l.store.mu.Lock()
	defer l.store.mu.Unlock()
	l.store.entries = append(l.store.entries, retryLogEntry{level: level, msg: msg, args: append([]any(nil), args...)})
}

func (l *retryCapturedLogger) Debug(msg string, args ...any) { l.append("DEBUG", msg, args...) }

func (l *retryCapturedLogger) DebugContext(_ context.Context, msg string, args ...any) {
	l.Debug(msg, args...)
}

func (l *retryCapturedLogger) Info(msg string, args ...any) { l.append("INFO", msg, args...) }

func (l *retryCapturedLogger) InfoContext(_ context.Context, msg string, args ...any) {
	l.Info(msg, args...)
}

func (l *retryCapturedLogger) Warn(msg string, args ...any) { l.append("WARN", msg, args...) }

func (l *retryCapturedLogger) WarnContext(_ context.Context, msg string, args ...any) {
	l.Warn(msg, args...)
}

func (l *retryCapturedLogger) Error(msg string, args ...any) { l.append("ERROR", msg, args...) }

func (l *retryCapturedLogger) ErrorContext(_ context.Context, msg string, args ...any) {
	l.Error(msg, args...)
}

func (l *retryCapturedLogger) With(args ...any) logger.Logger { return l }

func (l *retryCapturedLogger) WithGroup(name string) logger.Logger { return l }

func countStreamFinishedByStatus(entries []retryLogEntry, status string) int {
	count := 0
	for _, entry := range entries {
		if entry.msg != "stream_finished" {
			continue
		}
		for i := 0; i+1 < len(entry.args); i += 2 {
			key, ok := entry.args[i].(string)
			if !ok || key != "status" {
				continue
			}
			if val, ok := entry.args[i+1].(string); ok && val == status {
				count++
			}
		}
	}
	return count
}

func countLogsByMessage(entries []retryLogEntry, msg string) int {
	count := 0
	for _, entry := range entries {
		if entry.msg == msg {
			count++
		}
	}
	return count
}

func hasLogKeyValue(entry retryLogEntry, key string, value any) bool {
	for i := 0; i+1 < len(entry.args); i += 2 {
		k, ok := entry.args[i].(string)
		if !ok || k != key {
			continue
		}
		if entry.args[i+1] == value {
			return true
		}
	}
	return false
}

func hasMessageWithKeyValue(entries []retryLogEntry, msg, key string, value any) bool {
	for _, entry := range entries {
		if entry.msg != msg {
			continue
		}
		if hasLogKeyValue(entry, key, value) {
			return true
		}
	}
	return false
}

func TestRetryNonStream_RetryableThenSuccess(t *testing.T) {
	p := &Portal{
		session: session.New(),
		logger:  logger.NewNopLogger(),
	}

	channelCalls := 0
	executeCalls := 0

	result, err := retryNonStream(context.Background(), p,
		func(ctx context.Context) (*routing.Channel, error) {
			channelCalls++
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel) (string, error) {
			executeCalls++
			if executeCalls == 1 {
				return "", errors.New(errors.ErrCodeUnavailable, "临时错误").
					WithContext("error_from", string(errors.ErrorFromGateway))
			}
			return "ok", nil
		},
		nil,
	)

	if err != nil {
		t.Fatalf("期望成功，实际错误: %v", err)
	}
	if result != "ok" {
		t.Fatalf("结果期望 ok，实际: %q", result)
	}
	if channelCalls != 2 {
		t.Fatalf("getChannel 调用次数期望 2，实际：%d", channelCalls)
	}
	if executeCalls != 2 {
		t.Fatalf("execute 调用次数期望 2，实际：%d", executeCalls)
	}
}

func TestRetryNonStream_CanceledContextStopsImmediately(t *testing.T) {
	p := &Portal{
		session: session.New(),
		logger:  logger.NewNopLogger(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	channelCalls := 0

	_, err := retryNonStream(ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			channelCalls++
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel) (string, error) {
			return "", nil
		},
		nil,
	)

	if !errors.IsCode(err, errors.ErrCodeAborted) {
		t.Fatalf("错误码期望 ABORTED，实际: %v", errors.GetCode(err))
	}
	if channelCalls != 0 {
		t.Fatalf("已取消上下文不应拉取通道，实际调用次数：%d", channelCalls)
	}
}

func TestRetryNonStream_AbortedShouldNotRetry(t *testing.T) {
	p := &Portal{
		session: session.New(),
		logger:  logger.NewNopLogger(),
	}

	channelCalls := 0
	executeCalls := 0

	_, err := retryNonStream(context.Background(), p,
		func(ctx context.Context) (*routing.Channel, error) {
			channelCalls++
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel) (string, error) {
			executeCalls++
			return "", errors.NormalizeCanceled(context.Canceled)
		},
		nil,
	)

	if !errors.IsCode(err, errors.ErrCodeAborted) {
		t.Fatalf("错误码期望 ABORTED，实际: %v", errors.GetCode(err))
	}
	if channelCalls != 1 {
		t.Fatalf("getChannel 调用次数期望 1，实际：%d", channelCalls)
	}
	if executeCalls != 1 {
		t.Fatalf("execute 调用次数期望 1，实际：%d", executeCalls)
	}
}

func TestRetryNativeStream_CanceledAfterRetryableStopsRetry(t *testing.T) {
	p := &Portal{
		session: session.New(),
		logger:  logger.NewNopLogger(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	channelCalls := 0
	executeCalls := 0

	out := retryNativeStream[string](ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			channelCalls++
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			executeCalls++
			cancel()
			return errors.New(errors.ErrCodeUnavailable, "临时错误").
				WithContext("error_from", string(errors.ErrorFromGateway))
		},
		nil,
	)

	for range out {
	}

	if channelCalls != 1 {
		t.Fatalf("取消后不应继续重试，getChannel 调用次数期望 1，实际：%d", channelCalls)
	}
	if executeCalls != 1 {
		t.Fatalf("取消后不应继续重试，execute 调用次数期望 1，实际：%d", executeCalls)
	}
}

func TestRetryNativeStream_AsyncProducerNoPanicAndCompleteOutput(t *testing.T) {
	p := &Portal{
		session: session.New(),
		logger:  logger.NewNopLogger(),
	}

	out := retryNativeStream[string](context.Background(), p,
		func(ctx context.Context) (*routing.Channel, error) {
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			go func() {
				defer close(output)
				output <- "evt-1"
				time.Sleep(10 * time.Millisecond)
				output <- "evt-2"
			}()
			return nil
		},
		nil,
	)

	var got []string
	for evt := range out {
		got = append(got, evt)
	}

	if len(got) != 2 {
		t.Fatalf("事件数量期望 2，实际：%d", len(got))
	}
	if got[0] != "evt-1" || got[1] != "evt-2" {
		t.Fatalf("事件顺序不符合预期，实际：%v", got)
	}
}

func TestRetryNativeStream_CloseOutOnlyAfterProducerFinished(t *testing.T) {
	p := &Portal{
		session: session.New(),
		logger:  logger.NewNopLogger(),
	}

	releaseSecond := make(chan struct{})
	producerDone := make(chan struct{})

	out := retryNativeStream[string](context.Background(), p,
		func(ctx context.Context) (*routing.Channel, error) {
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			go func() {
				defer close(output)
				defer close(producerDone)
				output <- "first"
				<-releaseSecond
				output <- "second"
			}()
			return nil
		},
		nil,
	)

	select {
	case evt, ok := <-out:
		if !ok {
			t.Fatal("接收首个事件前输出通道已关闭")
		}
		if evt != "first" {
			t.Fatalf("首个事件期望 first，实际：%s", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("等待首个事件超时")
	}

	select {
	case _, ok := <-out:
		if !ok {
			t.Fatal("生产者未结束时输出通道不应关闭")
		}
		t.Fatal("生产者未结束时不应提前收到后续事件")
	default:
		// 符合预期：仍在等待生产者完成
	}

	close(releaseSecond)

	select {
	case evt, ok := <-out:
		if !ok {
			t.Fatal("期望收到 second，但输出通道已关闭")
		}
		if evt != "second" {
			t.Fatalf("第二个事件期望 second，实际：%s", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("等待 second 事件超时")
	}

	select {
	case <-producerDone:
	case <-time.After(time.Second):
		t.Fatal("生产者结束信号超时")
	}

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("生产者结束后输出通道应关闭")
		}
	case <-time.After(time.Second):
		t.Fatal("等待输出通道关闭超时")
	}
}

func TestRetryNativeStream_CompletedLogsSingleCompletedWithoutCanceled(t *testing.T) {
	logStore := &retryLogStore{}
	p := &Portal{
		session: session.New(),
		logger:  &retryCapturedLogger{store: logStore},
	}

	out := retryNativeStream[string](context.Background(), p,
		func(ctx context.Context) (*routing.Channel, error) {
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			go func() {
				defer close(output)
				output <- "evt"
			}()
			return nil
		},
		nil,
	)

	for range out {
	}

	logStore.mu.Lock()
	entries := append([]retryLogEntry(nil), logStore.entries...)
	logStore.mu.Unlock()

	completedCount := countStreamFinishedByStatus(entries, "completed")
	canceledCount := countStreamFinishedByStatus(entries, "canceled")
	if completedCount != 1 {
		t.Fatalf("completed 日志数量期望 1，实际：%d", completedCount)
	}
	if canceledCount != 0 {
		t.Fatalf("completed 场景不应出现 canceled 日志，实际：%d", canceledCount)
	}
}

func TestRetryNativeStream_TerminalThenCleanupCancelStillCompleted(t *testing.T) {
	logStore := &retryLogStore{}
	p := &Portal{
		session: session.New(),
		logger:  &retryCapturedLogger{store: logStore},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := retryNativeStream[string](ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			close(output)
			cancel()
			return nil
		},
		nil,
	)

	for range out {
	}

	logStore.mu.Lock()
	entries := append([]retryLogEntry(nil), logStore.entries...)
	logStore.mu.Unlock()

	if got := countLogsByMessage(entries, "stream_finished"); got != 1 {
		t.Fatalf("stream_finished 日志数量期望 1，实际：%d", got)
	}
}

func TestRetryNativeStream_StartupCanceledLogsSingleCanceled(t *testing.T) {
	logStore := &retryLogStore{}
	p := &Portal{
		session: session.New(),
		logger:  &retryCapturedLogger{store: logStore},
	}

	out := retryNativeStream[string](context.Background(), p,
		func(ctx context.Context) (*routing.Channel, error) {
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			return errors.NormalizeCanceled(context.Canceled)
		},
		nil,
	)

	for range out {
	}

	logStore.mu.Lock()
	entries := append([]retryLogEntry(nil), logStore.entries...)
	logStore.mu.Unlock()

	if got := countStreamFinishedByStatus(entries, "canceled"); got != 1 {
		t.Fatalf("启动阶段取消应记录 1 条 canceled，实际：%d", got)
	}
	if got := countStreamFinishedByStatus(entries, "completed"); got != 0 {
		t.Fatalf("启动阶段取消不应出现 completed 日志，实际：%d", got)
	}
	if got := countLogsByMessage(entries, "stream_finished"); got != 1 {
		t.Fatalf("stream_finished 日志数量期望 1，实际：%d", got)
	}
}

func TestRetryNativeStream_CanceledLogsSingleCanceledWithoutCompleted(t *testing.T) {
	logStore := &retryLogStore{}
	p := &Portal{
		session: session.New(),
		logger:  &retryCapturedLogger{store: logStore},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	releaseProducer := make(chan struct{})
	out := retryNativeStream[string](ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			go func() {
				defer close(output)
				output <- "evt"
				<-releaseProducer
			}()
			return nil
		},
		nil,
	)

	select {
	case <-time.After(time.Second):
		t.Fatal("等待首个事件超时")
	case _, ok := <-out:
		if !ok {
			t.Fatal("期望收到首个事件")
		}
	}

	cancel()
	close(releaseProducer)

	for range out {
	}

	logStore.mu.Lock()
	entries := append([]retryLogEntry(nil), logStore.entries...)
	logStore.mu.Unlock()

	if got := countLogsByMessage(entries, "stream_finished"); got != 1 {
		t.Fatalf("stream_finished 日志数量期望 1，实际：%d", got)
	}
	if !hasMessageWithKeyValue(entries, "stream_finished", "termination_phase", "forwarding") {
		t.Fatal("中途取消应记录 termination_phase=forwarding")
	}
	if !hasMessageWithKeyValue(entries, "stream_finished", "before_drain", true) {
		t.Fatal("中途取消应记录 before_drain=true")
	}
}
