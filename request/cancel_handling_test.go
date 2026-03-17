package request

import (
	"context"
	"testing"

	"github.com/MeowSalty/portal/errors"
)

func TestNormalizeCanceled_ContextCanceled(t *testing.T) {
	err := normalizeCanceled(context.Canceled)

	if !errors.IsCode(err, errors.ErrCodeAborted) {
		t.Fatalf("错误码期望 ABORTED，实际：%v", errors.GetCode(err))
	}

	if got := errors.GetHTTPStatus(err); got != httpStatusClientClosedRequest {
		t.Fatalf("HTTP 状态码期望 499，实际：%d", got)
	}

	ctx := errors.GetContext(err)
	if ctx == nil {
		t.Fatalf("错误上下文期望不为 nil")
	}

	if got, ok := ctx["error_from"]; !ok || got != "client" {
		t.Fatalf("error_from 期望 client，实际：%v", got)
	}
}

func TestNormalizeCanceled_ContextDeadlineExceeded(t *testing.T) {
	err := normalizeCanceled(context.DeadlineExceeded)

	if !errors.IsCode(err, errors.ErrCodeAborted) {
		t.Fatalf("错误码期望 ABORTED，实际：%v", errors.GetCode(err))
	}

	if got := errors.GetHTTPStatus(err); got != httpStatusClientClosedRequest {
		t.Fatalf("HTTP 状态码期望 499，实际：%d", got)
	}

	ctx := errors.GetContext(err)
	if ctx == nil {
		t.Fatalf("错误上下文期望不为 nil")
	}

	if got, ok := ctx["error_from"]; !ok || got != "client" {
		t.Fatalf("error_from 期望 client，实际：%v", got)
	}
}

func TestNormalizeCanceled_ExistingAbortedKeepsCodeAndSetsFields(t *testing.T) {
	sourceErr := errors.Wrap(errors.ErrCodeAborted, "连接被终止", context.Canceled)
	err := normalizeCanceled(sourceErr)

	if !errors.IsCode(err, errors.ErrCodeAborted) {
		t.Fatalf("错误码期望 ABORTED，实际：%v", errors.GetCode(err))
	}

	if got := errors.GetHTTPStatus(err); got != httpStatusClientClosedRequest {
		t.Fatalf("HTTP 状态码期望 499，实际：%d", got)
	}

	ctx := errors.GetContext(err)
	if ctx == nil {
		t.Fatalf("错误上下文期望不为 nil")
	}

	if got, ok := ctx["error_from"]; !ok || got != "client" {
		t.Fatalf("error_from 期望 client，实际：%v", got)
	}
}

func TestFillRequestLogErrorFields_CanceledError(t *testing.T) {
	log := &RequestLog{}
	err := normalizeCanceled(context.Canceled)

	fillRequestLogErrorFields(log, err)
	log.Success = false

	if log.ErrorCode == nil || *log.ErrorCode != string(errors.ErrCodeAborted) {
		t.Fatalf("ErrorCode 期望 ABORTED，实际：%+v", log.ErrorCode)
	}

	if log.HTTPStatus == nil || *log.HTTPStatus != httpStatusClientClosedRequest {
		t.Fatalf("HTTPStatus 期望 499，实际：%+v", log.HTTPStatus)
	}

	if log.ErrorFrom == nil || *log.ErrorFrom != "client" {
		t.Fatalf("ErrorFrom 期望 client，实际：%+v", log.ErrorFrom)
	}

	if log.Success {
		t.Fatalf("Success 期望 false，实际：true")
	}
}
