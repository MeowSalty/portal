package portal

import (
	"log/slog"

	"github.com/MeowSalty/portal/health"
	"github.com/MeowSalty/portal/types"
)

// Config 包含 GatewayManager 的所有依赖和配置
type Config struct {
	Repo          types.DataRepository
	HealthManager *health.Manager
	Selector      types.ChannelSelector
	AdapterTypes  []string // 指定启用的适配器类型，空则启用所有支持的适配器
	Logger        *slog.Logger
}
