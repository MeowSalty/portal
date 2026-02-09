// Package middleware 提供响应处理中间件功能
package middleware

import (
	"context"

	"github.com/MeowSalty/portal/request/adapter/types"
)

// Middleware 响应中间件接口（无状态，适用于简单处理）
type Middleware interface {
	// Process 处理非流式响应
	Process(ctx context.Context, req *types.RequestContract, resp *types.ResponseContract) (*types.ResponseContract, error)

	// ProcessStream 处理流式响应的单个 chunk（无状态）
	ProcessStream(ctx context.Context, req *types.RequestContract, resp *types.StreamEventContract) (*types.StreamEventContract, error)
}

// StreamMiddleware 流式中间件接口（有状态，支持缓冲）
//
// 适用于需要跨 chunk 处理的场景，如 XML 工具调用解析
type StreamMiddleware interface {
	Middleware

	// CreateHandler 为每个流式请求创建独立的处理器
	//
	// 返回的 StreamHandler 维护该请求的状态
	CreateHandler(ctx context.Context, req *types.RequestContract) StreamHandler
}

// StreamHandler 流式处理器接口（有状态）
type StreamHandler interface {
	// Handle 处理单个 chunk，可能返回：
	//   - 零个响应（缓冲中）
	//   - 一个响应（直接传递或处理后发送）
	//   - 多个响应（缓冲区释放）
	Handle(ctx context.Context, resp *types.StreamEventContract) ([]*types.StreamEventContract, error)

	// Flush 流结束时调用，返回缓冲区中剩余的数据
	Flush(ctx context.Context) ([]*types.StreamEventContract, error)
}
