package errors

import (
	"context"
	stdErrors "errors"
)

// HTTPStatusClientClosedRequest 表示客户端主动断连的 HTTP 状态码 (499)。
const HTTPStatusClientClosedRequest = 499

// IsCanceled 判断错误是否为取消或超时类型。
//
// 说明：
//   - 返回 true 表示错误属于取消类（context.Canceled）或超时类（context.DeadlineExceeded）
//   - 如需区分具体类型，请使用 ClassifyTermination
func IsCanceled(err error) bool {
	if err == nil {
		return false
	}

	if stdErrors.Is(err, context.Canceled) || stdErrors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return IsCode(err, ErrCodeAborted) || IsCode(err, ErrCodeCanceled)
}

// NormalizeCanceled 将取消类错误归一为统一格式。
//
// 映射规则：
//   - context.DeadlineExceeded -> DEADLINE_EXCEEDED/504/gateway
//   - context.Canceled（默认）-> ABORTED/499/client
//   - 已有 ABORTED/CANCELED 错误保持原错误码，补充 HTTP 状态码和 error_from
//
// 如需更精细控制（区分 client/server cancel），请使用 ClassifyTermination。
func NormalizeCanceled(err error) error {
	if !IsCanceled(err) {
		return err
	}

	// 使用终止分类器进行分类
	classification := ClassifyTermination(TerminationInput{
		Err:                err,
		IsClientDisconnect: true, // 默认假设为客户端取消，保持向后兼容
	})

	// 检查是否已有 Portal 错误
	var portalErr *Error
	if As(err, &portalErr) {
		// 已有错误，补充缺失字段
		if portalErr.HTTPStatus == nil {
			portalErr.WithHTTPStatus(classification.HTTPStatus)
		}

		if portalErr.Context == nil || portalErr.Context["error_from"] == nil {
			portalErr.WithContext("error_from", string(classification.ErrorFrom))
		}

		return portalErr
	}

	// 创建新的错误
	return WrapWithHTTPStatus(classification.ErrorCode, "请求已取消", err, classification.HTTPStatus).
		WithContext("error_from", string(classification.ErrorFrom))
}

// NormalizeCanceledWithSource 将取消类错误归一为统一格式，支持指定取消来源。
//
// 映射规则：
//   - context.DeadlineExceeded -> DEADLINE_EXCEEDED/504/gateway
//   - context.Canceled + isClientDisconnect=true -> ABORTED/499/client
//   - context.Canceled + isClientDisconnect=false -> CANCELED/499/server
func NormalizeCanceledWithSource(err error, isClientDisconnect bool) error {
	if !IsCanceled(err) {
		return err
	}

	// 使用终止分类器进行分类
	classification := ClassifyTermination(TerminationInput{
		Err:                err,
		IsClientDisconnect: isClientDisconnect,
	})

	// 检查是否已有 Portal 错误
	var portalErr *Error
	if As(err, &portalErr) {
		// 已有错误，补充缺失字段
		if portalErr.HTTPStatus == nil {
			portalErr.WithHTTPStatus(classification.HTTPStatus)
		}

		if portalErr.Context == nil || portalErr.Context["error_from"] == nil {
			portalErr.WithContext("error_from", string(classification.ErrorFrom))
		}

		return portalErr
	}

	// 创建新的错误
	message := "请求已取消"
	if classification.Kind == TerminationKindDeadline {
		message = "请求超时"
	}

	return WrapWithHTTPStatus(classification.ErrorCode, message, err, classification.HTTPStatus).
		WithContext("error_from", string(classification.ErrorFrom))
}
