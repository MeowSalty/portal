package session

import (
	"context"
	"testing"
	"time"

	"github.com/MeowSalty/portal/errors"
)

func TestWithSession_ActiveLifecycleBoundToFunctionExecution(t *testing.T) {
	sm := New()

	started := make(chan struct{})
	finished := make(chan struct{})

	go func() {
		_ = sm.WithSession(context.Background(), func(reqCtx context.Context, reqCancel context.CancelFunc) error {
			close(started)
			// 模拟业务层提前取消请求上下文，但函数本体仍在执行。
			reqCancel()
			time.Sleep(80 * time.Millisecond)
			close(finished)
			return nil
		})
	}()

	<-started
	begin := time.Now()
	err := sm.Shutdown(20 * time.Millisecond)
	elapsed := time.Since(begin)

	if !errors.IsCode(err, errors.ErrCodeDeadlineExceeded) {
		t.Fatalf("Shutdown 错误码期望 DEADLINE_EXCEEDED，实际：%v", errors.GetCode(err))
	}

	// 如果会话计数被 reqCtx.Done() 提前释放，这里会非常快返回。
	if elapsed < 60*time.Millisecond {
		t.Fatalf("Shutdown 返回过快，说明会话生命周期可能未绑定到函数执行时长，elapsed=%v", elapsed)
	}

	select {
	case <-finished:
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("WithSession 函数未按预期完成")
	}
}

func TestWithSessionStream_ErrorShouldReleaseActiveSessionImmediately(t *testing.T) {
	sm := New()
	done := make(chan struct{})

	expectErr := errors.New(errors.ErrCodeInternal, "测试错误")
	err := sm.WithSessionStream(context.Background(), done, func(reqCtx context.Context) error {
		return expectErr
	})
	if err != expectErr {
		t.Fatalf("WithSessionStream 返回错误不符合预期")
	}

	shutdownErr := sm.Shutdown(50 * time.Millisecond)
	if shutdownErr != nil {
		t.Fatalf("WithSessionStream 失败后会话应立即释放，Shutdown 不应超时，实际：%v", shutdownErr)
	}
}
