package portal

import (
	"context"
	"time"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
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
	logger  logger.Logger
}

type Config struct {
	PlatformRepo routing.PlatformRepository
	ModelRepo    routing.ModelRepository
	KeyRepo      routing.KeyRepository
	HealthRepo   health.HealthRepository
	LogRepo      request.RequestLogRepository
	Logger       logger.Logger // 可选的日志记录器，如果为 nil 则使用默认的空操作日志记录器
}

func New(cfg Config) (*Portal, error) {
	// 如果未提供日志记录器，使用默认的空操作日志记录器
	log := cfg.Logger
	if log == nil {
		log = logger.NewNopLogger()
	}

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
		request: request.New(cfg.LogRepo, log.WithGroup("request")),
		logger:  log,
	}
	return portal, nil
}

func (p *Portal) ChatCompletion(ctx context.Context, request *types.Request) (*types.Response, error) {
	p.logger.DebugContext(ctx, "开始处理聊天完成请求", "model", request.Model)

	var response *types.Response
	var channel *routing.Channel
	var err error
	for {
		channel, err = p.routing.GetChannel(ctx, request.Model)
		if err != nil {
			p.logger.ErrorContext(ctx, "获取通道失败", "model", request.Model, "error", err)
			break
		}

		// 使用 With 创建带有通道上下文的日志记录器
		channelLogger := p.logger.With(
			"platform_id", channel.PlatformID,
			"model_id", channel.ModelID,
			"api_key_id", channel.APIKeyID)

		channelLogger.DebugContext(ctx, "获取到通道")

		err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) (err error) {
			defer reqCancel()
			response, err = p.request.ChatCompletion(reqCtx, request, channel)
			return
		})

		// 检查错误是否可以重试
		if err != nil {
			if errors.IsRetryable(err) {
				channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
				channel.MarkFailure(ctx, err)
				continue
			}
			// 特殊处理：操作终止时标记成功
			if errors.IsCode(err, errors.ErrCodeAborted) {
				channelLogger.InfoContext(ctx, "操作终止")
				channel.MarkSuccess(ctx)
			}
			channelLogger.ErrorContext(ctx, "请求处理失败", "error", err)
			channel.MarkFailure(ctx, err)
			break
		}
		channel.MarkSuccess(ctx)
		channelLogger.InfoContext(ctx, "请求处理成功")
		break
	}
	return response, err
}

func (p *Portal) ChatCompletionStream(ctx context.Context, request *types.Request) (<-chan *types.Response, error) {
	p.logger.DebugContext(ctx, "开始处理流式聊天完成请求", "model", request.Model)

	// 创建用于返回给函数调用方的流
	clientStream := make(chan *types.Response, 1024)

	var channel *routing.Channel
	var err error

	// 启动内部流处理协程
	go func() {
		for {
			channel, err = p.routing.GetChannel(ctx, request.Model)
			if err != nil {
				p.logger.ErrorContext(ctx, "获取通道失败", "model", request.Model, "error", err)
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

			// 使用 With 创建带有通道上下文的日志记录器
			channelLogger := p.logger.With(
				"platform_id", channel.PlatformID,
				"model_id", channel.ModelID,
				"api_key_id", channel.APIKeyID)

			channelLogger.DebugContext(ctx, "获取到通道")

			err = p.session.WithSession(ctx, func(reqCtx context.Context, reqCancel context.CancelFunc) (err error) {
				defer reqCancel()
				return p.request.ChatCompletionStream(reqCtx, request, clientStream, channel)
			})

			// 检查错误是否可以重试
			if err != nil {
				if errors.IsRetryable(err) {
					channelLogger.WarnContext(ctx, "请求失败，尝试重试", "error", err)
					channel.MarkFailure(ctx, err)
					continue
				}
				// 特殊处理：操作终止时标记成功
				if errors.IsCode(err, errors.ErrCodeAborted) {
					channelLogger.InfoContext(ctx, "操作终止")
					channel.MarkSuccess(ctx)
				}
				channelLogger.ErrorContext(ctx, "流处理失败", "error", err)
				channel.MarkFailure(ctx, err)
				close(clientStream)
				break
			}
			channel.MarkSuccess(ctx)
			channelLogger.InfoContext(ctx, "流处理成功")
			break
		}
	}()

	return clientStream, err
}

func (p *Portal) Close(timeout time.Duration) error {
	return p.session.Shutdown(timeout, p.routing)
}
