package portal

import (
	"log/slog"
	"time"
)

// ShutdownManager 管理服务的优雅关闭
type ShutdownManager struct {
	sessionManager *SessionManager
	healthManager  interface{ Shutdown() }
	logger         *slog.Logger
}

// NewShutdownManager 创建一个新的关闭管理器
func NewShutdownManager(sessionManager *SessionManager, healthManager interface{ Shutdown() }, logger *slog.Logger) *ShutdownManager {
	return &ShutdownManager{
		sessionManager: sessionManager,
		healthManager:  healthManager,
		logger:         logger.WithGroup("shutdown"),
	}
}

// Shutdown 优雅地关闭服务
//
// 它会等待所有正在进行的会话完成，然后关闭健康管理器。
// 可以通过 timeout 参数设置最长等待时间。
// 如果等待超时，所有正在进行的会话将被中断。
//
// 参数：
//   - timeout: 最长等待时间，0 表示无限等待
//
// 返回值：
//   - error: 如果等待超时返回 ErrShutdownTimeout，否则返回 nil
func (sm *ShutdownManager) Shutdown(timeout time.Duration) error {
	// 1. 标记服务正在停机，拒绝新请求
	sm.sessionManager.markShuttingDown()
	sm.logger.Info("服务开始停机，不再接受新请求")

	// 2. 等待所有活动会话完成
	done := make(chan struct{})
	go func() {
		sm.sessionManager.waitForSessions()
		close(done)
	}()

	var err error
	if timeout > 0 {
		select {
		case <-done:
			sm.logger.Info("所有活动会话已正常完成")
		case <-time.After(timeout):
			sm.logger.Warn("停机等待超时，正在中断所有剩余会话...",
				slog.Duration("timeout", timeout))
			sm.sessionManager.cancelAllSessions()
			<-done // 等待被中断的会话完成清理
			sm.logger.Info("所有被中断的会话已结束")
			err = ErrShutdownTimeout
		}
	} else {
		// 无超时限制，无限等待
		<-done
		sm.logger.Info("所有活动会话已正常完成")
	}

	// 3. 关闭健康管理器
	if sm.healthManager != nil {
		sm.logger.Info("正在关闭健康管理器")
		sm.healthManager.Shutdown()
	}

	sm.logger.Info("服务已成功停机")
	return err
}
