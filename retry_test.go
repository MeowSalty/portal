package portal

import (
	"context"
	"testing"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/logger"
	"github.com/MeowSalty/portal/routing"
	"github.com/MeowSalty/portal/session"
)

func TestRetryNonStream_RetryableThenSuccess(t *testing.T) {
	p := &Portal{
		session: session.New(),
		logger:  logger.NewNopLogger(),
	}

	channelCalls := 0
	executeCalls := 0

	result, err := retryNonStream(context.Background(), p,
		func(ctx context.Context) (*routing.Channel, error) {
			channelCalls++
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel) (string, error) {
			executeCalls++
			if executeCalls == 1 {
				return "", errors.New(errors.ErrCodeUnavailable, "临时错误").
					WithContext("error_from", string(errors.ErrorFromGateway))
			}
			return "ok", nil
		},
		nil,
	)

	if err != nil {
		t.Fatalf("期望成功，实际错误: %v", err)
	}
	if result != "ok" {
		t.Fatalf("结果期望 ok，实际: %q", result)
	}
	if channelCalls != 2 {
		t.Fatalf("getChannel 调用次数期望 2，实际: %d", channelCalls)
	}
	if executeCalls != 2 {
		t.Fatalf("execute 调用次数期望 2，实际: %d", executeCalls)
	}
}

func TestRetryNonStream_CanceledContextStopsImmediately(t *testing.T) {
	p := &Portal{
		session: session.New(),
		logger:  logger.NewNopLogger(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	channelCalls := 0

	_, err := retryNonStream[string](ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			channelCalls++
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel) (string, error) {
			return "", nil
		},
		nil,
	)

	if !errors.IsCode(err, errors.ErrCodeAborted) {
		t.Fatalf("错误码期望 ABORTED，实际: %v", errors.GetCode(err))
	}
	if channelCalls != 0 {
		t.Fatalf("已取消上下文不应拉取通道，实际调用次数: %d", channelCalls)
	}
}

func TestRetryNonStream_AbortedShouldNotRetry(t *testing.T) {
	p := &Portal{
		session: session.New(),
		logger:  logger.NewNopLogger(),
	}

	channelCalls := 0
	executeCalls := 0

	_, err := retryNonStream[string](context.Background(), p,
		func(ctx context.Context) (*routing.Channel, error) {
			channelCalls++
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel) (string, error) {
			executeCalls++
			return "", errors.NormalizeCanceled(context.Canceled)
		},
		nil,
	)

	if !errors.IsCode(err, errors.ErrCodeAborted) {
		t.Fatalf("错误码期望 ABORTED，实际: %v", errors.GetCode(err))
	}
	if channelCalls != 1 {
		t.Fatalf("getChannel 调用次数期望 1，实际: %d", channelCalls)
	}
	if executeCalls != 1 {
		t.Fatalf("execute 调用次数期望 1，实际: %d", executeCalls)
	}
}

func TestRetryNativeStream_CanceledAfterRetryableStopsRetry(t *testing.T) {
	p := &Portal{
		session: session.New(),
		logger:  logger.NewNopLogger(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	channelCalls := 0
	executeCalls := 0

	out := retryNativeStream[string](ctx, p,
		func(ctx context.Context) (*routing.Channel, error) {
			channelCalls++
			return &routing.Channel{}, nil
		},
		func(reqCtx context.Context, ch *routing.Channel, output chan<- any) error {
			executeCalls++
			cancel()
			return errors.New(errors.ErrCodeUnavailable, "临时错误").
				WithContext("error_from", string(errors.ErrorFromGateway))
		},
		nil,
	)

	for range out {
	}

	if channelCalls != 1 {
		t.Fatalf("取消后不应继续重试，getChannel 调用次数期望 1，实际: %d", channelCalls)
	}
	if executeCalls != 1 {
		t.Fatalf("取消后不应继续重试，execute 调用次数期望 1，实际: %d", executeCalls)
	}
}
