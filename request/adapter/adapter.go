package adapter

import (
	"context"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/routing"
	coreTypes "github.com/MeowSalty/portal/types"
	"github.com/valyala/fasthttp"
)

// Adapter 统一适配器实现
//
// Adapter 是适配器模式的核心类，它封装了：
//   - HTTP 客户端：处理底层网络通信
//   - Provider 接口：委托给具体提供商处理请求/响应转换
//   - 日志记录器：统一的日志记录
//
// Adapter 通过组合模式将不同提供商的 API 统一为一致的调用接口。
type Adapter struct {
	client   *fasthttp.Client
	provider Provider
}

// NewAdapterFromProvider 从 Provider 创建适配器实例
//
// 这是创建适配器的标准方法，通常不直接调用，而是通过 GetAdapter() 函数创建。
//
// 参数：
//   - provider: 提供商实例
//   - logger: 日志记录器，如果为 nil 则使用默认记录器
//
// 返回：
//   - *Adapter: 适配器实例
func NewAdapterFromProvider(provider Provider) *Adapter {
	if provider == nil {
		panic("provider 不能为空")
	}

	return &Adapter{
		client:   newHTTPClient(),
		provider: provider,
	}
}

// Name 返回适配器所使用的提供商名称
func (a *Adapter) Name() string {
	return a.provider.Name()
}

// ChatCompletion 执行聊天完成请求
func (a *Adapter) ChatCompletion(
	ctx context.Context,
	request *coreTypes.Request,
	channel *routing.Channel,
) (*coreTypes.Response, error) {
	// 创建提供商特定请求
	apiReq, err := a.provider.CreateRequest(request, channel)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "创建请求失败", err)
	}

	// 发送请求
	httpResp, err := a.sendHTTPRequest(channel, request.Headers, apiReq, false)
	if err != nil {
		return nil, err
	}

	// 检查 HTTP 状态码
	if httpResp.StatusCode != fasthttp.StatusOK {
		err := a.handleHTTPError("API 返回错误状态码", httpResp.StatusCode, httpResp.Body)
		return nil, err
	}

	// 解析响应
	response, err := a.provider.ParseResponse(httpResp.Body)
	if err != nil {
		err := a.handleParseError("响应解析错误", err, httpResp.Body)
		return nil, err
	}

	return response, nil
}

// ChatCompletionStream 执行流式聊天完成请求
func (a *Adapter) ChatCompletionStream(
	ctx context.Context,
	request *coreTypes.Request,
	channel *routing.Channel,
	output chan<- *coreTypes.Response,
) error {
	if !a.provider.SupportsStreaming() {
		return errors.New(errors.ErrCodeUnimplemented, "提供商不支持流式传输").
			WithContext("provider", a.provider.Name())
	}

	// 创建提供商特定请求
	apiReq, err := a.provider.CreateRequest(request, channel)
	if err != nil {
		return errors.Wrap(errors.ErrCodeInvalidArgument, "创建请求失败", err)
	}

	// 启动流式处理协程
	return a.handleStreaming(ctx, channel, request.Headers, apiReq, output)
}
