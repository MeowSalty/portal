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
	// 参数：
	//   - model: 模型名称
	//   - stream: 是否为流式请求
	//
	// 返回：
	//   - string: API 端点路径（如 "/v1/chat/completions"）
	APIEndpoint(model string, stream bool) string

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
}
