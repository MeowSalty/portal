package portal

import (
	"context"
	"sync"
	"sync/atomic"
)

// SessionManager 管理请求会话的生命周期
type SessionManager struct {
	isShuttingDown atomic.Bool
	activeSessions sync.WaitGroup
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// NewSessionManager 创建一个新的会话管理器
func NewSessionManager() *SessionManager {
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	return &SessionManager{
		shutdownCtx:    shutdownCtx,
		shutdownCancel: shutdownCancel,
	}
}

// withSession 管理单个请求会话的生命周期
//
// 该函数负责在处理请求时正确管理会话计数和上下文取消。
// 它会在网关关闭时拒绝新请求，并确保在处理过程中正确处理上下文取消。
//
// 参数：
//   - ctx: 父级上下文
//   - fn: 在会话上下文中执行的函数
//
// 返回值：
//   - error: 执行过程中发生的错误，如果服务正在关闭则返回 ErrServerShuttingDown
func (sm *SessionManager) withSession(ctx context.Context, fn func(reqCtx context.Context) error) error {
	// 检查服务是否正在关闭，如果是则拒绝新请求
	if sm.isShuttingDown.Load() {
		return ErrServerShuttingDown
	}
	sm.activeSessions.Add(1)
	defer sm.activeSessions.Done()

	// 创建可取消的请求上下文
	reqCtx, reqCancel := context.WithCancel(ctx)
	defer reqCancel()

	// 启动一个 goroutine 监听关闭信号和上下文完成信号
	done := make(chan struct{})
	go func() {
		select {
		case <-sm.shutdownCtx.Done():
			reqCancel()
		case <-reqCtx.Done():
		}
		close(done)
	}()

	// 执行会话函数
	err := fn(reqCtx)

	// 等待监听 goroutine 结束
	<-done

	return err
}

// withStreamSession 管理流式请求会话的生命周期
//
// 与 withSession 类似，但不会阻塞等待上下文完成，适用于需要立即返回响应流的场景。
// 监听 goroutine 会在后台运行，直到流结束或服务关闭。
//
// 参数：
//   - ctx: 父级上下文
//   - fn: 在会话上下文中执行的函数，该函数应该启动流处理并立即返回
//
// 返回值：
//   - error: 执行过程中发生的错误，如果服务正在关闭则返回 ErrServerShuttingDown
func (sm *SessionManager) withStreamSession(ctx context.Context, fn func(reqCtx context.Context) error) error {
	// 检查服务是否正在关闭，如果是则拒绝新请求
	if sm.isShuttingDown.Load() {
		return ErrServerShuttingDown
	}
	sm.activeSessions.Add(1)

	// 创建可取消的请求上下文
	reqCtx, reqCancel := context.WithCancel(ctx)

	// 启动一个 goroutine 监听关闭信号和上下文完成信号
	// 这个 goroutine 会在流结束时自动清理
	go func() {
		defer sm.activeSessions.Done()
		defer reqCancel()

		select {
		case <-sm.shutdownCtx.Done():
			// 服务关闭，取消请求上下文
		case <-reqCtx.Done():
			// 请求完成或被客户端取消
		}
	}()

	// 执行会话函数，立即返回
	return fn(reqCtx)
}

// markShuttingDown 标记服务正在关闭
func (sm *SessionManager) markShuttingDown() {
	sm.isShuttingDown.Store(true)
}

// waitForSessions 等待所有活动会话完成
func (sm *SessionManager) waitForSessions() {
	sm.activeSessions.Wait()
}

// cancelAllSessions 取消所有活动会话
func (sm *SessionManager) cancelAllSessions() {
	sm.shutdownCancel()
}
