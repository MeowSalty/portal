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
func (m *Manager) RecordRequestStat(ctx context.Context, opts *RecordOptions) error {
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

// QueryStats 查询请求统计列表
func (m *Manager) QueryStats(ctx context.Context, params *types.StatsQueryParams) ([]*types.RequestStat, error) {
	stats, err := m.repo.QueryRequestStats(ctx, params)
	if err != nil {
		m.logger.Error("查询请求统计信息失败", "error", err)
		return nil, err
	}

	return stats, nil
}

// CountStats 统计请求计数
func (m *Manager) CountStats(ctx context.Context, params *types.StatsQueryParams) (*types.StatsSummary, error) {
	summary, err := m.repo.CountRequestStats(ctx, params)
	if err != nil {
		m.logger.Error("统计请求计数失败", "error", err)
		return nil, err
	}

	return summary, nil
}

// RecordOptions 包含记录统计信息所需的选项
type RecordOptions struct {
	// 请求基本信息
	Timestamp   time.Time         `json:"timestamp"`    // 请求时间
	RequestType string            `json:"request_type"` // 请求类型：stream 或 non-stream
	ModelName   string            `json:"model_name"`   // 模型名称
	ChannelInfo types.ChannelInfo `json:"channel_info"` // 通道信息

	// 耗时信息
	Duration      time.Duration  `json:"duration"`                  // 总用时
	FirstByteTime *time.Duration `json:"first_byte_time,omitempty"` // 首字用时（仅流式）

	// 结果状态
	Success   bool    `json:"success"`              // 是否成功
	ErrorCode *int    `json:"error_code,omitempty"` // 错误码（失败时）
	ErrorMsg  *string `json:"error_msg,omitempty"`  // 错误信息（失败时）
}

// generateID 生成唯一的 ID
func generateID() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%d%d", time.Now().UnixNano(), n)
}
