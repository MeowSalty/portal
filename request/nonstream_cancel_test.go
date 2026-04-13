package request

import (
	"context"
	"testing"

	"github.com/MeowSalty/portal/errors"
)

func TestNormalizeNonStreamCanceledError_Deadline(t *testing.T) {
	err := normalizeNonStreamCanceledError(context.DeadlineExceeded)

	if !errors.IsCode(err, errors.ErrCodeDeadlineExceeded) {
		t.Fatalf("错误码期望 DEADLINE_EXCEEDED，实际：%v", errors.GetCode(err))
	}
	if got := errors.GetHTTPStatus(err); got != 504 {
		t.Fatalf("HTTP 状态码期望 504，实际：%d", got)
	}
	if got := errors.GetErrorFrom(err); got != errors.ErrorFromGateway {
		t.Fatalf("error_from 期望 gateway，实际：%v", got)
	}
}

func TestNormalizeNonStreamCanceledError_ServerCancel(t *testing.T) {
	err := normalizeNonStreamCanceledError(errors.Wrap(errors.ErrCodeCanceled, "服务端主动取消", context.Canceled).
		WithContext("error_from", string(errors.ErrorFromServer)))

	if !errors.IsCode(err, errors.ErrCodeCanceled) {
		t.Fatalf("错误码期望 CANCELED，实际：%v", errors.GetCode(err))
	}
	if got := errors.GetHTTPStatus(err); got != errors.HTTPStatusClientClosedRequest {
		t.Fatalf("HTTP 状态码期望 499，实际：%d", got)
	}
	if got := errors.GetErrorFrom(err); got != errors.ErrorFromServer {
		t.Fatalf("error_from 期望 server，实际：%v", got)
	}
}

func TestNormalizeNonStreamCanceledError_ClientCancel(t *testing.T) {
	err := normalizeNonStreamCanceledError(context.Canceled)

	if !errors.IsCode(err, errors.ErrCodeAborted) {
		t.Fatalf("错误码期望 ABORTED，实际：%v", errors.GetCode(err))
	}
	if got := errors.GetHTTPStatus(err); got != errors.HTTPStatusClientClosedRequest {
		t.Fatalf("HTTP 状态码期望 499，实际：%d", got)
	}
	if got := errors.GetErrorFrom(err); got != errors.ErrorFromClient {
		t.Fatalf("error_from 期望 client，实际：%v", got)
	}
}
