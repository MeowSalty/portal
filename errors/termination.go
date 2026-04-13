package errors

import (
	"context"
	stdErrors "errors"
)

// TerminationKind 定义终止类型。
type TerminationKind string

const (
	// TerminationKindClientCancel 客户端主动取消。
	TerminationKindClientCancel TerminationKind = "client_cancel"
	// TerminationKindServerCancel 服务端主动取消。
	TerminationKindServerCancel TerminationKind = "server_cancel"
	// TerminationKindDeadline 超时。
	TerminationKindDeadline TerminationKind = "deadline"
	// TerminationKindUnknownCancel 未知取消类型。
	TerminationKindUnknownCancel TerminationKind = "unknown_cancel"
)

// TerminationClassification 定义终止分类结果。
type TerminationClassification struct {
	// Kind 终止类型。
	Kind TerminationKind
	// ErrorCode 映射的错误码。
	ErrorCode ErrorCode
	// HTTPStatus 映射的 HTTP 状态码。
	HTTPStatus int
	// ErrorFrom 映射的错误来源。
	ErrorFrom ErrorFromValue
}

// TerminationInput 定义终止分类输入。
type TerminationInput struct {
	// Err 原始错误。
	Err error
	// IsClientDisconnect 是否为客户端断连（用于区分 client/server cancel）。
	IsClientDisconnect bool
}

// terminationMapping 定义终止类型到错误属性的映射。
var terminationMapping = map[TerminationKind]TerminationClassification{
	TerminationKindClientCancel: {
		Kind:       TerminationKindClientCancel,
		ErrorCode:  ErrCodeAborted,
		HTTPStatus: HTTPStatusClientClosedRequest,
		ErrorFrom:  ErrorFromClient,
	},
	TerminationKindServerCancel: {
		Kind:       TerminationKindServerCancel,
		ErrorCode:  ErrCodeCanceled,
		HTTPStatus: HTTPStatusClientClosedRequest,
		ErrorFrom:  ErrorFromServer,
	},
	TerminationKindDeadline: {
		Kind:       TerminationKindDeadline,
		ErrorCode:  ErrCodeDeadlineExceeded,
		HTTPStatus: 504,
		ErrorFrom:  ErrorFromGateway,
	},
	TerminationKindUnknownCancel: {
		Kind:       TerminationKindUnknownCancel,
		ErrorCode:  ErrCodeAborted,
		HTTPStatus: HTTPStatusClientClosedRequest,
		ErrorFrom:  ErrorFromGateway,
	},
}

// ClassifyTermination 对终止错误进行分类。
//
// 分类规则：
//   - context.DeadlineExceeded -> deadline (DEADLINE_EXCEEDED/504/gateway)
//   - context.Canceled + IsClientDisconnect=true -> client_cancel (ABORTED/499/client)
//   - context.Canceled + IsClientDisconnect=false -> server_cancel (CANCELED/499/server)
//   - 其他 -> unknown_cancel
func ClassifyTermination(input TerminationInput) TerminationClassification {
	err := input.Err
	if err == nil {
		return terminationMapping[TerminationKindUnknownCancel]
	}

	// 优先判断 deadline
	if stdErrors.Is(err, context.DeadlineExceeded) {
		return terminationMapping[TerminationKindDeadline]
	}

	// 判断 context.Canceled
	if stdErrors.Is(err, context.Canceled) {
		if input.IsClientDisconnect {
			return terminationMapping[TerminationKindClientCancel]
		}
		return terminationMapping[TerminationKindServerCancel]
	}

	// 已有的 ABORTED 错误，保持 client 来源
	if IsCode(err, ErrCodeAborted) {
		return terminationMapping[TerminationKindClientCancel]
	}

	// 其他情况归为未知
	return terminationMapping[TerminationKindUnknownCancel]
}

// IsDeadlineExceeded 判断错误是否为超时类型。
func IsDeadlineExceeded(err error) bool {
	if err == nil {
		return false
	}
	return stdErrors.Is(err, context.DeadlineExceeded)
}

// IsContextCanceled 判断错误是否为 context.Canceled（不含 DeadlineExceeded）。
func IsContextCanceled(err error) bool {
	if err == nil {
		return false
	}
	return stdErrors.Is(err, context.Canceled)
}
