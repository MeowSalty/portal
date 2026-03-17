package adapter

import (
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
)

// Provider 定义 AI 提供商的接口
//
// Provider 接口是适配器模式中的"目标接口"，定义了统一的 AI 服务提供商操作规范。
// 所有具体的提供商实现（如 OpenAI、Gemini）都必须实现此接口。
//
// Provider 实现必须是无状态的（不持有可变的实例级字段），
// 以便安全地在多个请求之间共享和缓存。
type Provider interface {
	// Name 返回提供商名称
	Name() string

	// CreateRequest 将核心请求转换为提供商特定请求
	CreateRequest(request *types.RequestContract, channel *routing.Channel) (interface{}, error)

	// ParseResponse 解析提供商响应并转换为核心响应
	//
	// 参数：
	//   - variant: API 变体（如 "chat_completions", "responses"），由 channel.APIVariant 传入
	//   - responseData: 原始响应数据（JSON 字节数组）
	ParseResponse(variant string, responseData []byte) (*types.ResponseContract, error)

	// ParseStreamResponse 解析提供商流式响应并转换为核心响应
	//
	// 参数：
	//   - variant: API 变体
	//   - ctx: 流索引上下文，用于生成和维护稳定的索引值
	//   - responseData: 单个流式响应块的数据（JSON 字节数组）
	ParseStreamResponse(variant string, ctx types.StreamIndexContext, responseData []byte) ([]*types.StreamEventContract, error)

	// APIEndpoint 返回 API 端点路径
	//
	// EndpointConfig 解析规则：
	//   - config == "" → 使用默认端点
	//   - config 以 "/" 结尾 → 视为"前缀"；拼接默认端点
	//   - 其他情况 → 视为完整路径
	//
	// 参数：
	//   - variant: API 变体
	//   - model: 模型名称
	//   - stream: 是否为流式请求
	//   - config: 可选端点配置（前缀或完整路径）
	APIEndpoint(variant string, model string, stream bool, config ...string) string

	// Headers 返回特定于提供商的 HTTP 头，包括身份验证头部
	Headers(key string) map[string]string

	// SupportsStreaming 返回是否支持流式传输
	SupportsStreaming() bool

	// SupportsNative 返回是否支持原生 API 调用
	SupportsNative() bool

	// BuildNativeRequest 构建原生请求
	BuildNativeRequest(channel *routing.Channel, payload any) (body any, err error)

	// ParseNativeResponse 解析原生响应
	ParseNativeResponse(variant string, raw []byte) (any, error)

	// ParseNativeStreamEvent 解析原生流事件
	ParseNativeStreamEvent(variant string, raw []byte) (any, error)

	// ExtractUsageFromNativeStreamEvent 从原生流事件中提取使用统计信息
	ExtractUsageFromNativeStreamEvent(variant string, event any) *types.ResponseUsage
}
