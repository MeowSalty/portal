package adapter

import (
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
)

// Provider 定义 AI 提供商的接口
//
// Provider 接口是适配器模式中的"目标接口"，定义了统一的 AI 服务提供商操作规范。
// 所有具体的提供商实现（如 OpenAI、Gemini）都必须实现此接口。
type Provider interface {
	// Name 返回提供商名称
	//
	// 该名称用于日志记录和错误信息中标识提供商。
	Name() string

	// CreateRequest 将核心请求转换为提供商特定请求
	//
	// 这是适配器模式的核心方法之一，将统一的请求格式转换为特定提供商的 API 格式。
	//
	// 参数：
	//   - request: 核心请求对象
	//   - channel: 渠道信息，包含 API 密钥等配置
	//
	// 返回：
	//   - interface{}: 提供商特定的请求对象
	//   - error: 转换失败时返回错误
	CreateRequest(request *types.RequestContract, channel *routing.Channel) (interface{}, error)

	// ParseResponse 解析提供商响应并转换为核心响应
	//
	// 这是适配器模式的另一个核心方法，将提供商特定的响应格式转换为统一的核心格式。
	//
	// 参数：
	//   - responseData: 原始响应数据（JSON 字节数组）
	//
	// 返回：
	//   - *types.ResponseContract: 统一的核心响应对象
	//   - error: 解析失败时返回错误
	ParseResponse(responseData []byte) (*types.ResponseContract, error)

	// ParseStreamResponse 解析提供商流式响应并转换为核心响应
	//
	// 用于解析 Server-Sent Events (SSE) 格式的流式响应。
	//
	// 参数：
	//   - ctx: 流索引上下文，用于生成和维护稳定的索引值
	//   - responseData: 单个流式响应块的数据（JSON 字节数组）
	//
	// 返回：
	//   - []*types.StreamEventContract: 统一的核心流式事件列表
	//   - error: 解析失败时返回错误
	ParseStreamResponse(ctx types.StreamIndexContext, responseData []byte) ([]*types.StreamEventContract, error)

	// APIEndpoint 返回 API 端点路径
	//
	// 不同提供商的 API 端点可能不同，此方法返回相对于 BaseURL 的路径。
	//
	// EndpointConfig 解析规则：
	//   - config == "" → 使用默认端点
	//   - config 以 "/" 结尾 → 视为"前缀"；拼接默认端点
	//   - 其他情况 → 视为完整路径
	//
	// 参数：
	//   - model: 模型名称
	//   - stream: 是否为流式请求
	//   - config: 可选端点配置（前缀或完整路径）
	//
	// 返回：
	//   - string: API 端点路径（如 "/v1/chat/completions"）
	APIEndpoint(model string, stream bool, config ...string) string

	// Headers 返回特定于提供商的 HTTP 头，包括身份验证头部
	//
	// 不同提供商的认证方式可能不同（如 Bearer Token、API Key 等）。
	//
	// 参数：
	//   - key: API 密钥
	//
	// 返回：
	//   - map[string]string: HTTP 头部键值对
	Headers(key string) map[string]string

	// SupportsStreaming 返回是否支持流式传输
	//
	// 某些提供商可能不支持流式传输，此方法用于能力检查。
	SupportsStreaming() bool

	// SupportsNative 返回是否支持原生 API 调用
	//
	// 原生 API 调用允许直接使用提供商的原生请求/响应类型，
	// 不经过标准 contract 转换。适用于需要访问提供商特定功能的场景。
	//
	// 返回：
	//   - bool: 是否支持原生 API 调用
	SupportsNative() bool

	// BuildNativeRequest 构建原生请求
	//
	// 将原生 payload 转换为提供商特定的请求格式。
	//
	// 参数：
	//   - channel: 渠道信息，包含 API 密钥和配置（包括 APIEndpointConfig）
	//   - payload: 原生请求 payload（根据 APIVariant 决定具体类型）
	//
	// 返回：
	//   - body: 转换后的请求体
	//   - error: 转换失败时返回错误
	BuildNativeRequest(channel *routing.Channel, payload any) (body any, err error)

	// ParseNativeResponse 解析原生响应
	//
	// 将提供商的原生响应转换为统一的 any 类型返回。
	//
	// 参数：
	//   - variant: API 变体（如 "chat_completions", "responses"）
	//   - raw: 原始响应数据（JSON 字节数组）
	//
	// 返回：
	//   - any: 原生响应对象
	//   - error: 解析失败时返回错误
	ParseNativeResponse(variant string, raw []byte) (any, error)

	// ParseNativeStreamEvent 解析原生流事件
	//
	// 将提供商的原生流事件转换为统一的 any 类型返回。
	//
	// 参数：
	//   - variant: API 变体（如 "chat_completions", "responses"）
	//   - raw: 单个流事件的数据（JSON 字节数组）
	//
	// 返回：
	//   - any: 原生流事件对象
	//   - error: 解析失败时返回错误
	ParseNativeStreamEvent(variant string, raw []byte) (any, error)

	// ExtractUsageFromNativeStreamEvent 从原生流事件中提取使用统计信息
	//
	// 从 ParseNativeStreamEvent 返回的原生流事件对象中提取 usage 信息。
	//
	// 参数：
	//   - variant: API 变体（如 "chat_completions", "responses"）
	//   - event: 原生流事件对象
	//
	// 返回：
	//   - *types.ResponseUsage: 提取的使用统计信息，如果事件中不包含 usage 则返回 nil
	ExtractUsageFromNativeStreamEvent(variant string, event any) *types.ResponseUsage
}
