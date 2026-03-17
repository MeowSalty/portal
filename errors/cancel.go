package errors

import (
	"context"
	stdErrors "errors"
)

// HTTPStatusClientClosedRequest 表示客户端主动断连的 HTTP 状态码 (499)。
const HTTPStatusClientClosedRequest = 499

// IsCanceled 判断错误是否由客户端取消/断连引起。
func IsCanceled(err error) bool {
	if err == nil {
		return false
	}

	if stdErrors.Is(err, context.Canceled) || stdErrors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return IsCode(err, ErrCodeAborted)
}

// NormalizeCanceled 将取消类错误归一为 ABORTED/499/error_from=client。
func NormalizeCanceled(err error) error {
	if !IsCanceled(err) {
		return err
	}

	var portalErr *Error
	if As(err, &portalErr) && portalErr.Code == ErrCodeAborted {
		if portalErr.HTTPStatus == nil {
			portalErr.WithHTTPStatus(HTTPStatusClientClosedRequest)
		}

		if portalErr.Context == nil || portalErr.Context["error_from"] == nil {
			portalErr.WithContext("error_from", "client")
		}

		return portalErr
	}

	return WrapWithHTTPStatus(ErrCodeAborted, "请求已取消", err, HTTPStatusClientClosedRequest).
		WithContext("error_from", "client")
}
