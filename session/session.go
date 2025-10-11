package session

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MeowSalty/portal/errors"
)

// Session 管理请求会话的生命周期和优雅停机
type Session struct {
	isShuttingDown atomic.Bool
	activeSessions sync.WaitGroup
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// New 创建一个新的会话管理器
func New() *Session {
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	return &Session{
		shutdownCtx:    shutdownCtx,
		shutdownCancel: shutdownCancel,
	}
}

// WithSession 管理单个请求会话的生命周期
//
// 该函数负责在处理请求时正确管理会话计数和上下文取消。
// 它会在网关关闭时拒绝新请求，并确保在处理过程中正确处理上下文取消。
//
// 参数：
//   - ctx: 父级上下文
//   - fn: 在会话上下文中执行的函数，接收 reqCtx 和 reqCancel 参数
//
// 返回值：
//   - error: 执行过程中发生的错误，如果服务正在关闭则返回 ErrServerShuttingDown
func (sm *Session) WithSession(ctx context.Context, fn func(reqCtx context.Context, reqCancel context.CancelFunc) error) error {
	// 检查服务是否正在关闭，如果是则拒绝新请求
	if sm.isShuttingDown.Load() {
		return errors.New(errors.ErrCodeUnavailable, "服务正在关闭")
	}
	sm.activeSessions.Add(1)

	// 创建可取消的请求上下文
	reqCtx, reqCancel := context.WithCancel(ctx)

	// 启动一个 goroutine 监听关闭信号和上下文完成信号
	// 这个 goroutine 会在流结束时自动清理
	go func() {
		defer sm.activeSessions.Done()

		select {
		case <-sm.shutdownCtx.Done():
			// 服务关闭，取消请求上下文
			reqCancel()
		case <-reqCtx.Done():
			// 请求完成或被客户端取消
		}
	}()

	// 执行会话函数
	return fn(reqCtx, reqCancel)
}

// Shutdown 优雅地关闭服务
//
// 它会等待所有正在进行的会话完成，然后关闭健康管理器。
// 可以通过 timeout 参数设置最长等待时间。
// 如果等待超时，所有正在进行的会话将被中断。
//
// 参数：
//   - timeout: 最长等待时间，0 表示无限等待
//   - routingShutdown: 用于同步关闭路由
//
// 返回值：
//   - error: 如果等待超时返回 ErrShutdownTimeout，否则返回 nil
func (s *Session) Shutdown(timeout time.Duration, routingShutdown interface{ Shutdown() }) error {
	// 1. 标记服务正在停机，拒绝新请求
	s.isShuttingDown.Store(true)

	// 2. 等待所有活动会话完成
	done := make(chan struct{})
	go func() {
		s.activeSessions.Wait()
		close(done)
	}()

	var err error
	if timeout > 0 {
		select {
		case <-done:
			// 所有活动会话已正常完成
		case <-time.After(timeout):
			// 停机等待超时，正在中断所有剩余会话...
			s.shutdownCancel()
			<-done // 等待被中断的会话完成清理
			// 所有被中断的会话已结束
			err = errors.ErrDeadlineExceeded
		}
	} else {
		// 无超时限制，无限等待
		<-done
		// 所有活动会话已正常完成
	}

	// 3. 关闭路由
	if routingShutdown != nil {
		// 正在关闭路由
		routingShutdown.Shutdown()
	}

	// 服务已成功停机
	return err
}
