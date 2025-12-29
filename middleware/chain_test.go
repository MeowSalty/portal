package middleware

import (
	"context"
	"testing"

	"github.com/MeowSalty/portal/types"
)

// mockMiddleware 测试用的模拟中间件
type mockMiddleware struct {
	processFunc       func(ctx context.Context, req *types.Request, resp *types.Response) (*types.Response, error)
	processStreamFunc func(ctx context.Context, req *types.Request, resp *types.Response) (*types.Response, error)
}

func (m *mockMiddleware) Process(ctx context.Context, req *types.Request, resp *types.Response) (*types.Response, error) {
	if m.processFunc != nil {
		return m.processFunc(ctx, req, resp)
	}
	return resp, nil
}

func (m *mockMiddleware) ProcessStream(ctx context.Context, req *types.Request, resp *types.Response) (*types.Response, error) {
	if m.processStreamFunc != nil {
		return m.processStreamFunc(ctx, req, resp)
	}
	return resp, nil
}

// TestChain_Process_Empty 测试空中间件链
func TestChain_Process_Empty(t *testing.T) {
	chain := NewChain()
	content := "test content"
	resp := &types.Response{
		Choices: []types.Choice{
			{
				Message: &types.ResponseMessage{
					Content: &content,
				},
			},
		},
	}

	result, err := chain.Process(context.Background(), &types.Request{}, resp)
	if err != nil {
		t.Fatalf("处理失败: %v", err)
	}

	if *result.Choices[0].Message.Content != content {
		t.Errorf("期望内容为 '%s'，实际为 '%s'", content, *result.Choices[0].Message.Content)
	}
}

// TestChain_Process_SingleMiddleware 测试单个中间件
func TestChain_Process_SingleMiddleware(t *testing.T) {
	modified := "modified"
	middleware := &mockMiddleware{
		processFunc: func(ctx context.Context, req *types.Request, resp *types.Response) (*types.Response, error) {
			resp.Choices[0].Message.Content = &modified
			return resp, nil
		},
	}

	chain := NewChain(middleware)
	content := "original"
	resp := &types.Response{
		Choices: []types.Choice{
			{
				Message: &types.ResponseMessage{
					Content: &content,
				},
			},
		},
	}

	result, err := chain.Process(context.Background(), &types.Request{}, resp)
	if err != nil {
		t.Fatalf("处理失败: %v", err)
	}

	if *result.Choices[0].Message.Content != modified {
		t.Errorf("期望内容为 '%s'，实际为 '%s'", modified, *result.Choices[0].Message.Content)
	}
}

// TestChain_Process_MultipleMiddlewares 测试多个中间件的执行顺序
func TestChain_Process_MultipleMiddlewares(t *testing.T) {
	callOrder := []string{}

	m1 := &mockMiddleware{
		processFunc: func(ctx context.Context, req *types.Request, resp *types.Response) (*types.Response, error) {
			callOrder = append(callOrder, "m1")
			return resp, nil
		},
	}

	m2 := &mockMiddleware{
		processFunc: func(ctx context.Context, req *types.Request, resp *types.Response) (*types.Response, error) {
			callOrder = append(callOrder, "m2")
			return resp, nil
		},
	}

	chain := NewChain(m1, m2)
	content := "test"
	resp := &types.Response{
		Choices: []types.Choice{
			{
				Message: &types.ResponseMessage{
					Content: &content,
				},
			},
		},
	}

	_, err := chain.Process(context.Background(), &types.Request{}, resp)
	if err != nil {
		t.Fatalf("处理失败: %v", err)
	}

	if len(callOrder) != 2 {
		t.Fatalf("期望调用 2 次，实际调用 %d 次", len(callOrder))
	}

	if callOrder[0] != "m1" || callOrder[1] != "m2" {
		t.Errorf("调用顺序错误: %v", callOrder)
	}
}

// TestChain_ProcessStream_Empty 测试空中间件链的流式处理
func TestChain_ProcessStream_Empty(t *testing.T) {
	chain := NewChain()
	content := "stream content"
	req := &types.Request{}

	in := make(chan *types.Response, 1)
	in <- &types.Response{
		Choices: []types.Choice{
			{
				Delta: &types.Delta{
					Content: &content,
				},
			},
		},
	}
	close(in)

	out := chain.ProcessStream(context.Background(), req, in)

	result := <-out
	if *result.Choices[0].Delta.Content != content {
		t.Errorf("期望内容为 '%s'，实际为 '%s'", content, *result.Choices[0].Delta.Content)
	}
}

// TestChain_ProcessStream_SingleMiddleware 测试单个中间件的流式处理
func TestChain_ProcessStream_SingleMiddleware(t *testing.T) {
	modified := "modified stream"
	middleware := &mockMiddleware{
		processStreamFunc: func(ctx context.Context, req *types.Request, resp *types.Response) (*types.Response, error) {
			resp.Choices[0].Delta.Content = &modified
			return resp, nil
		},
	}

	chain := NewChain(middleware)
	content := "original stream"
	req := &types.Request{}

	in := make(chan *types.Response, 1)
	in <- &types.Response{
		Choices: []types.Choice{
			{
				Delta: &types.Delta{
					Content: &content,
				},
			},
		},
	}
	close(in)

	out := chain.ProcessStream(context.Background(), req, in)

	result := <-out
	if *result.Choices[0].Delta.Content != modified {
		t.Errorf("期望内容为 '%s'，实际为 '%s'", modified, *result.Choices[0].Delta.Content)
	}
}

// TestStatelessHandler_Handle 测试无状态处理器
func TestStatelessHandler_Handle(t *testing.T) {
	modified := "handled"
	middleware := &mockMiddleware{
		processStreamFunc: func(ctx context.Context, req *types.Request, resp *types.Response) (*types.Response, error) {
			resp.Choices[0].Delta.Content = &modified
			return resp, nil
		},
	}

	handler := &statelessHandler{
		middleware: middleware,
		req:        &types.Request{},
	}

	content := "original"
	resp := &types.Response{
		Choices: []types.Choice{
			{
				Delta: &types.Delta{
					Content: &content,
				},
			},
		},
	}

	results, err := handler.Handle(context.Background(), resp)
	if err != nil {
		t.Fatalf("处理失败: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("期望返回 1 个结果，实际返回 %d 个", len(results))
	}

	if *results[0].Choices[0].Delta.Content != modified {
		t.Errorf("期望内容为 '%s'，实际为 '%s'", modified, *results[0].Choices[0].Delta.Content)
	}
}

// TestStatelessHandler_Flush 测试无状态处理器的 Flush
func TestStatelessHandler_Flush(t *testing.T) {
	handler := &statelessHandler{
		middleware: &mockMiddleware{},
		req:        &types.Request{},
	}

	results, err := handler.Flush(context.Background())
	if err != nil {
		t.Fatalf("Flush 失败: %v", err)
	}

	if results != nil {
		t.Errorf("期望 Flush 返回 nil，实际返回 %v", results)
	}
}
