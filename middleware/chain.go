package middleware

import (
	"context"

	"github.com/MeowSalty/portal/request/adapter/types"
)

// Chain 中间件链
type Chain struct {
	middlewares []Middleware
}

// NewChain 创建中间件链
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{middlewares: middlewares}
}

// Process 处理非流式响应
func (c *Chain) Process(ctx context.Context, req *types.RequestContract, resp *types.ResponseContract) (*types.ResponseContract, error) {
	if len(c.middlewares) == 0 {
		return resp, nil
	}
	var err error
	for _, m := range c.middlewares {
		resp, err = m.Process(ctx, req, resp)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

// ProcessStream 处理流式响应（支持有状态的 StreamMiddleware）
func (c *Chain) ProcessStream(ctx context.Context, req *types.RequestContract, in <-chan *types.StreamEventContract) <-chan *types.StreamEventContract {
	if len(c.middlewares) == 0 {
		return in
	}

	// 为每个中间件创建 handler
	handlers := make([]StreamHandler, 0, len(c.middlewares))
	for _, m := range c.middlewares {
		if sm, ok := m.(StreamMiddleware); ok {
			handlers = append(handlers, sm.CreateHandler(ctx, req))
		} else {
			// 普通中间件包装为无状态 handler
			handlers = append(handlers, &statelessHandler{middleware: m, req: req})
		}
	}

	out := make(chan *types.StreamEventContract, cap(in))
	go func() {
		defer close(out)

		// 处理每个 chunk
		for resp := range in {
			results := []*types.StreamEventContract{resp}

			// 依次通过每个 handler
			for _, h := range handlers {
				var nextResults []*types.StreamEventContract
				for _, r := range results {
					processed, err := h.Handle(ctx, r)
					if err != nil {
						continue
					}
					nextResults = append(nextResults, processed...)
				}
				results = nextResults
				if len(results) == 0 {
					break // 所有数据都被缓冲了
				}
			}

			// 发送处理后的响应
			for _, r := range results {
				select {
				case <-ctx.Done():
					return
				case out <- r:
				}
			}
		}

		// 流结束，刷新所有 handler 的缓冲区
		for _, h := range handlers {
			flushed, err := h.Flush(ctx)
			if err != nil {
				continue
			}
			for _, r := range flushed {
				select {
				case <-ctx.Done():
					return
				case out <- r:
				}
			}
		}
	}()
	return out
}

// statelessHandler 将无状态中间件包装为 StreamHandler
type statelessHandler struct {
	middleware Middleware
	req        *types.RequestContract
}

func (h *statelessHandler) Handle(ctx context.Context, resp *types.StreamEventContract) ([]*types.StreamEventContract, error) {
	processed, err := h.middleware.ProcessStream(ctx, h.req, resp)
	if err != nil {
		return nil, err
	}
	return []*types.StreamEventContract{processed}, nil
}

func (h *statelessHandler) Flush(ctx context.Context) ([]*types.StreamEventContract, error) {
	return nil, nil // 无状态，无需刷新
}
