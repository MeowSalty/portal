package portal

import (
	"context"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/request"
	"github.com/MeowSalty/portal/routing"
	"github.com/MeowSalty/portal/routing/health"
	"github.com/MeowSalty/portal/routing/selector"
	"github.com/MeowSalty/portal/session"
	"github.com/MeowSalty/portal/types"
)

type Portal struct {
	session *session.Session
	routing *routing.Routing
	request *request.Request
}

type Config struct {
	PlatformRepo routing.PlatformRepository
	ModelRepo    routing.ModelRepository
	KeyRepo      routing.KeyRepository
	HealthRepo   health.HealthRepository
	LogRepo      request.RequestLogRepository
}

func New(cfg Config) (*Portal, error) {
	routing, err := routing.New(context.TODO(), routing.Config{
		PlatformRepo: cfg.PlatformRepo,
		ModelRepo:    cfg.ModelRepo,
		KeyRepo:      cfg.KeyRepo,
		HealthRepo:   cfg.HealthRepo,
		Selector:     selector.NewLRUSelector(),
	})
	if err != nil {
		return nil, err
	}
	portal := &Portal{
		session: session.New(),
		routing: routing,
		request: request.New(cfg.LogRepo),
	}
	return portal, nil
}

func (p *Portal) ChatCompletion(ctx context.Context, request *types.Request) (*types.Response, error) {
	var response *types.Response
	var channel *routing.Channel
	var err error
	for {
		channel, err = p.routing.GetChannel(ctx, request.Model)
		if err != nil {
			break
		}

		err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) (err error) {
			defer reqCancel()
			response, err = p.request.ChatCompletion(reqCtx, request, channel)
			return
		})

		// 检查错误是否可以重试
		if err != nil {
			// 网络错误
			if errors.IsCode(err, errors.ErrCodeUnavailable) {
				channel.MarkFailure(ctx, err)
				continue
			}
			// 请求失败
			if errors.IsCode(err, errors.ErrCodeRequestFailed) {
				channel.MarkFailure(ctx, err)
				continue
			}
			// 操作终止
			if errors.IsCode(err, errors.ErrCodeAborted) {
				channel.MarkSuccess(ctx)
			}
			break
		}
		channel.MarkSuccess(ctx)
		break
	}
	return response, err
}

func (p *Portal) ChatCompletionStream(ctx context.Context, request *types.Request) (<-chan *types.Response, error) {
	// 创建用于返回给函数调用方的流
	clientStream := make(chan *types.Response, 1024)

	var channel *routing.Channel
	var err error

	// 启动内部流处理协程
	go func() {
		for {
			channel, err = p.routing.GetChannel(ctx, request.Model)
			if err != nil {
				// 创建错误响应并发送到流中
				errorResponse := &types.Response{
					Choices: []types.Choice{
						{
							Error: &types.ErrorResponse{
								Code:    *err.(*errors.Error).HTTPStatus,
								Message: errors.GetMessage(err),
							},
						},
					},
				}
				select {
				case <-ctx.Done():
				default:
					select {
					case clientStream <- errorResponse:
					default:
					}
				}
				close(clientStream)
				break
			}

			err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) (err error) {
				defer reqCancel()
				return p.request.ChatCompletionStream(reqCtx, request, clientStream, channel)
			})

			// 检查错误是否可以重试
			if err != nil {
				// 网络错误
				if errors.IsCode(err, errors.ErrCodeUnavailable) {
					channel.MarkFailure(ctx, err)
					continue
				}
				// 请求失败
				if errors.IsCode(err, errors.ErrCodeRequestFailed) {
					channel.MarkFailure(ctx, err)
					continue
				}
				// 流处理失败
				if errors.IsCode(err, errors.ErrCodeStreamError) {
					channel.MarkFailure(ctx, err)
					continue
				}
				// 操作终止
				if errors.IsCode(err, errors.ErrCodeAborted) {
					channel.MarkSuccess(ctx)
				}
				channel.MarkFailure(ctx, err)
				close(clientStream)
				break
			}
			channel.MarkSuccess(ctx)
			break
		}
	}()

	return clientStream, err
}

func (p *Portal) Close(timeout time.Duration) error {
	return p.session.Shutdown(timeout, p.routing)
}
