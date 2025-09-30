// Package health 提供了资源健康状态管理功能
package health

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/MeowSalty/portal/types"
)

// Syncer 定义后台同步器接口
type Syncer interface {
	// Start 启动同步器
	Start()
	// Stop 停止同步器并等待所有操作完成
	Stop()
	// MarkDirty 标记资源为脏数据，需要同步到数据库
	MarkDirty(resourceType types.ResourceType, resourceID uint, status *types.Health)
	// Sync 立即执行同步操作
	Sync(ctx context.Context) error
}

// syncer 实现了后台同步器
type syncer struct {
	repo         types.DataRepository
	logger       *slog.Logger
	cache        Cache
	dirty        sync.Map      // 脏数据缓存，key: string, value: *types.Health
	syncInterval time.Duration // 同步间隔
	closeChan    chan struct{} // 关闭信号通道
	wg           sync.WaitGroup
}

// NewSyncer 创建一个新的同步器实例
func NewSyncer(
	repo types.DataRepository,
	logger *slog.Logger,
	cache Cache,
	syncInterval time.Duration,
) Syncer {
	return &syncer{
		repo:         repo,
		logger:       logger.WithGroup("syncer"),
		cache:        cache,
		syncInterval: syncInterval,
		closeChan:    make(chan struct{}),
	}
}

// Start 启动同步器
func (s *syncer) Start() {
	s.wg.Add(1)
	go s.run()
	s.logger.Info("健康状态同步器已启动", slog.Duration("interval", s.syncInterval))
}

// Stop 停止同步器并等待所有操作完成
func (s *syncer) Stop() {
	close(s.closeChan)
	s.wg.Wait()

	// 执行最终同步
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Sync(ctx); err != nil {
		s.logger.Error("关闭时健康同步失败", slog.Any("error", err))
	}

	s.logger.Info("健康状态同步器已停止")
}

// MarkDirty 标记资源为脏数据
func (s *syncer) MarkDirty(resourceType types.ResourceType, resourceID uint, status *types.Health) {
	key := generateKey(resourceType, resourceID)
	s.dirty.Store(key, status)
}

// run 是后台进程，定期将缓存与数据库同步
func (s *syncer) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 定期同步脏数据
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := s.Sync(ctx); err != nil {
				s.logger.Error("执行定期健康同步失败", slog.Any("error", err))
			}
			cancel()
		case <-s.closeChan:
			// 收到关闭信号，退出循环
			return
		}
	}
}

// Sync 立即执行同步操作
func (s *syncer) Sync(ctx context.Context) error {
	// 收集所有脏数据
	var statusesToSync []*types.Health
	var keysToDelete []string

	s.dirty.Range(func(key, value interface{}) bool {
		status := value.(*types.Health)
		statusesToSync = append(statusesToSync, status)
		keysToDelete = append(keysToDelete, key.(string))
		return true
	})

	if len(statusesToSync) == 0 {
		return nil
	}

	s.logger.Debug("开始同步健康状态", slog.Int("count", len(statusesToSync)))

	// 批量更新到数据库
	if err := s.repo.BatchUpdateHealthStatus(ctx, statusesToSync); err != nil {
		return fmt.Errorf("批量更新健康状态失败：%w", err)
	}

	// 清理已同步的脏数据
	for _, key := range keysToDelete {
		s.dirty.Delete(key)
	}

	s.logger.Info("健康状态同步完成", slog.Int("count", len(statusesToSync)))
	return nil
}
