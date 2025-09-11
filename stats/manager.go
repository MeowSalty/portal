package stats

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/MeowSalty/portal/types"
)

// Manager 负责管理请求统计
type Manager struct {
	repo   types.DataRepository
	logger *slog.Logger
}

// NewManager 创建一个新的统计管理器
func NewManager(repo types.DataRepository, logger *slog.Logger) *Manager {
	return &Manager{
		repo:   repo,
		logger: logger.WithGroup("stats_manager"),
	}
}

// RecordRequestStat 记录请求统计信息
func (m *Manager) RecordRequestStat(ctx context.Context, opts *types.RequestStat) error {
	stat := &types.RequestStat{
		ID:            generateID(),
		Timestamp:     opts.Timestamp,
		RequestType:   opts.RequestType,
		ModelName:     opts.ModelName,
		ChannelInfo:   opts.ChannelInfo,
		Duration:      opts.Duration,
		FirstByteTime: opts.FirstByteTime,
		Success:       opts.Success,
		ErrorMsg:      opts.ErrorMsg,
	}

	if err := m.repo.SaveRequestStat(ctx, stat); err != nil {
		m.logger.Error("保存请求统计信息失败", "error", err)
		return err
	}

	return nil
}

// generateID 生成唯一的 ID
func generateID() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%d%d", time.Now().UnixNano(), n)
}
