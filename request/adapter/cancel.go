package adapter

import (
	"context"
	stdErrors "errors"

	"github.com/MeowSalty/portal/errors"
)

const httpStatusClientClosedRequest = 499

// isCanceled 判断错误是否由客户端取消/断连引起。
func isCanceled(err error) bool {
	if err == nil {
		return false
	}

	if stdErrors.Is(err, context.Canceled) || stdErrors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return errors.IsCode(err, errors.ErrCodeAborted)
}

// normalizeCanceled 将取消类错误归一为 ABORTED/499/error_from=client。
func normalizeCanceled(err error) error {
	if !isCanceled(err) {
		return err
	}

	var portalErr *errors.Error
	if errors.As(err, &portalErr) && portalErr.Code == errors.ErrCodeAborted {
		if portalErr.HTTPStatus == nil {
			portalErr.WithHTTPStatus(httpStatusClientClosedRequest)
		}

		if portalErr.Context == nil || portalErr.Context["error_from"] == nil {
			portalErr.WithContext("error_from", "client")
		}

		return portalErr
	}

	return errors.WrapWithHTTPStatus(errors.ErrCodeAborted, "请求已取消", err, httpStatusClientClosedRequest).
		WithContext("error_from", "client")
}
