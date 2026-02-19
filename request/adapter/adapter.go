package adapter

import (
	"context"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/request/adapter/types"
	"github.com/MeowSalty/portal/routing"
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
	request *types.RequestContract,
	channel *routing.Channel,
) (*types.ResponseContract, error) {
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
		err := a.handleHTTPError("API 返回错误状态码", httpResp.StatusCode, httpResp.ContentType, httpResp.Body)
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
	request *types.RequestContract,
	channel *routing.Channel,
	output chan<- *types.StreamEventContract,
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

// Native 执行原生 API 请求（非流式）
//
// 该方法允许直接使用提供商的原生请求/响应类型，不经过标准 contract 转换。
// 适用于需要访问提供商特定功能的场景。
//
// 参数：
//   - ctx: 上下文
//   - channel: 通道信息
//   - headers: 自定义 HTTP 头部
//   - payload: 原生请求 payload（根据 channel.APIVariant 决定具体类型）
//
// 返回：
//   - any: 原生响应对象
//   - error: 请求失败时返回错误
func (a *Adapter) Native(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	payload any,
) (any, error) {
	if !a.provider.SupportsNative() {
		return nil, errors.New(errors.ErrCodeUnimplemented, "提供商不支持原生 API 调用").
			WithContext("provider", a.provider.Name())
	}

	// 构建原生请求（不再返回 endpoint）
	body, err := a.provider.BuildNativeRequest(channel, payload)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "构建原生请求失败", err)
	}

	// 发送请求（使用 channel.APIEndpointConfig）
	httpResp, err := a.sendHTTPRequest(channel, headers, body, false)
	if err != nil {
		return nil, err
	}

	// 检查 HTTP 状态码
	if httpResp.StatusCode != fasthttp.StatusOK {
		err := a.handleHTTPError("API 返回错误状态码", httpResp.StatusCode, httpResp.ContentType, httpResp.Body)
		return nil, err
	}

	// 解析原生响应
	response, err := a.provider.ParseNativeResponse(channel.APIVariant, httpResp.Body)
	if err != nil {
		err := a.handleParseError("响应解析错误", err, httpResp.Body)
		return nil, err
	}

	return response, nil
}

// NativeStream 执行原生 API 流式请求
//
// 该方法允许直接使用提供商的原生请求类型，不经过标准 contract 转换。
// 响应为原生流事件。
//
// 参数：
//   - ctx: 上下文
//   - channel: 通道信息
//   - headers: 自定义 HTTP 头部
//   - payload: 原生请求 payload（根据 channel.APIVariant 决定具体类型）
//   - output: 原生流事件输出通道
//
// 返回：
//   - error: 请求失败时返回错误
func (a *Adapter) NativeStream(
	ctx context.Context,
	channel *routing.Channel,
	headers map[string]string,
	payload any,
	output chan<- any,
	hooks types.StreamHooks,
) error {
	if !a.provider.SupportsNative() {
		return errors.New(errors.ErrCodeUnimplemented, "提供商不支持原生 API 调用").
			WithContext("provider", a.provider.Name())
	}

	// 构建原生请求（不再返回 endpoint）
	body, err := a.provider.BuildNativeRequest(channel, payload)
	if err != nil {
		return errors.Wrap(errors.ErrCodeInvalidArgument, "构建原生请求失败", err)
	}

	// 启动原生流式处理协程
	return a.handleNativeStreaming(ctx, channel, headers, body, output, hooks)
}
