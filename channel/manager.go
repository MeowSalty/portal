package channel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/MeowSalty/portal/types"
)

// Manager 负责通道的构建和管理
type Manager struct {
	repo   types.DataRepository
	logger *slog.Logger
}

// NewManager 创建一个新的通道管理器
func NewManager(repo types.DataRepository, logger *slog.Logger) *Manager {
	return &Manager{
		repo:   repo,
		logger: logger.WithGroup("channel"),
	}
}

// BuildChannels 从模型列表创建所有可能的通道列表
//
// 该方法会为每个模型获取对应的平台和 API 密钥，并构建通道对象
func (m *Manager) BuildChannels(ctx context.Context, models []*types.Model) ([]*types.Channel, error) {
	var channels []*types.Channel
	var errs []error

	for _, model := range models {
		platform, err := m.repo.GetPlatformByID(ctx, model.PlatformID)
		if err != nil {
			m.logger.Error("获取模型平台失败",
				slog.Uint64("模型 ID", uint64(model.ID)),
				slog.Uint64("平台 ID", uint64(model.PlatformID)),
				slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("模型 ID %d: 获取平台失败：%w", model.ID, err))
			continue
		}

		apiKeys, err := m.repo.GetAllAPIKeys(ctx, platform.ID)
		if err != nil {
			m.logger.Error("获取平台 API 密钥失败",
				slog.Uint64("平台 ID", uint64(platform.ID)),
				slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("平台 ID %d: 获取 API 密钥失败：%w", platform.ID, err))
			continue
		}

		if len(apiKeys) == 0 {
			m.logger.Warn("平台没有配置 API 密钥",
				slog.Uint64("平台 ID", uint64(platform.ID)),
				slog.String("平台名称", platform.Name))
			continue
		}

		for _, key := range apiKeys {
			channels = append(channels, &types.Channel{
				Platform: platform,
				Model:    model,
				APIKey:   key,
			})
		}
	}

	// 如果没有成功构建任何通道但有错误，则返回错误
	if len(channels) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("构建通道失败: %v", errs)
	}

	return channels, nil
}
