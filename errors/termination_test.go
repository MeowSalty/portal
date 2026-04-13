package errors

import (
	"context"
	"testing"
)

func TestClassifyTermination_DeadlineExceeded(t *testing.T) {
	result := ClassifyTermination(TerminationInput{
		Err:                context.DeadlineExceeded,
		IsClientDisconnect: false,
	})

	if result.Kind != TerminationKindDeadline {
		t.Fatalf("Kind = %q, want %q", result.Kind, TerminationKindDeadline)
	}
	if result.ErrorCode != ErrCodeDeadlineExceeded {
		t.Fatalf("ErrorCode = %q, want %q", result.ErrorCode, ErrCodeDeadlineExceeded)
	}
	if result.HTTPStatus != 504 {
		t.Fatalf("HTTPStatus = %d, want 504", result.HTTPStatus)
	}
	if result.ErrorFrom != ErrorFromGateway {
		t.Fatalf("ErrorFrom = %q, want %q", result.ErrorFrom, ErrorFromGateway)
	}
}

func TestClassifyTermination_ContextCanceled_ClientDisconnect(t *testing.T) {
	result := ClassifyTermination(TerminationInput{
		Err:                context.Canceled,
		IsClientDisconnect: true,
	})

	if result.Kind != TerminationKindClientCancel {
		t.Fatalf("Kind = %q, want %q", result.Kind, TerminationKindClientCancel)
	}
	if result.ErrorCode != ErrCodeAborted {
		t.Fatalf("ErrorCode = %q, want %q", result.ErrorCode, ErrCodeAborted)
	}
	if result.HTTPStatus != HTTPStatusClientClosedRequest {
		t.Fatalf("HTTPStatus = %d, want %d", result.HTTPStatus, HTTPStatusClientClosedRequest)
	}
	if result.ErrorFrom != ErrorFromClient {
		t.Fatalf("ErrorFrom = %q, want %q", result.ErrorFrom, ErrorFromClient)
	}
}

func TestClassifyTermination_ContextCanceled_ServerCancel(t *testing.T) {
	result := ClassifyTermination(TerminationInput{
		Err:                context.Canceled,
		IsClientDisconnect: false,
	})

	if result.Kind != TerminationKindServerCancel {
		t.Fatalf("Kind = %q, want %q", result.Kind, TerminationKindServerCancel)
	}
	if result.ErrorCode != ErrCodeCanceled {
		t.Fatalf("ErrorCode = %q, want %q", result.ErrorCode, ErrCodeCanceled)
	}
	if result.HTTPStatus != HTTPStatusClientClosedRequest {
		t.Fatalf("HTTPStatus = %d, want %d", result.HTTPStatus, HTTPStatusClientClosedRequest)
	}
	if result.ErrorFrom != ErrorFromServer {
		t.Fatalf("ErrorFrom = %q, want %q", result.ErrorFrom, ErrorFromServer)
	}
}

func TestClassifyTermination_AbortedError(t *testing.T) {
	err := New(ErrCodeAborted, "操作中止")
	result := ClassifyTermination(TerminationInput{
		Err:                err,
		IsClientDisconnect: false,
	})

	if result.Kind != TerminationKindClientCancel {
		t.Fatalf("Kind = %q, want %q", result.Kind, TerminationKindClientCancel)
	}
}

func TestClassifyTermination_NilError(t *testing.T) {
	result := ClassifyTermination(TerminationInput{
		Err:                nil,
		IsClientDisconnect: true,
	})

	if result.Kind != TerminationKindUnknownCancel {
		t.Fatalf("Kind = %q, want %q", result.Kind, TerminationKindUnknownCancel)
	}
}

func TestClassifyTermination_UnknownError(t *testing.T) {
	err := New(ErrCodeInternal, "内部错误")
	result := ClassifyTermination(TerminationInput{
		Err:                err,
		IsClientDisconnect: true,
	})

	if result.Kind != TerminationKindUnknownCancel {
		t.Fatalf("Kind = %q, want %q", result.Kind, TerminationKindUnknownCancel)
	}
}

func TestIsDeadlineExceeded(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "DeadlineExceeded 返回 true",
			err:  context.DeadlineExceeded,
			want: true,
		},
		{
			name: "Canceled 返回 false",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "nil 返回 false",
			err:  nil,
			want: false,
		},
		{
			name: "其他错误返回 false",
			err:  New(ErrCodeInternal, "内部错误"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDeadlineExceeded(tt.err); got != tt.want {
				t.Fatalf("IsDeadlineExceeded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsContextCanceled(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "Canceled 返回 true",
			err:  context.Canceled,
			want: true,
		},
		{
			name: "DeadlineExceeded 返回 false",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "nil 返回 false",
			err:  nil,
			want: false,
		},
		{
			name: "其他错误返回 false",
			err:  New(ErrCodeInternal, "内部错误"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContextCanceled(tt.err); got != tt.want {
				t.Fatalf("IsContextCanceled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeCanceledWithSource_DeadlineExceeded(t *testing.T) {
	err := NormalizeCanceledWithSource(context.DeadlineExceeded, true)

	if !IsCode(err, ErrCodeDeadlineExceeded) {
		t.Fatalf("ErrorCode = %q, want %q", GetCode(err), ErrCodeDeadlineExceeded)
	}
	if got := GetHTTPStatus(err); got != 504 {
		t.Fatalf("HTTPStatus = %d, want 504", got)
	}
	if GetErrorFrom(err) != ErrorFromGateway {
		t.Fatalf("ErrorFrom = %q, want %q", GetErrorFrom(err), ErrorFromGateway)
	}
}

func TestNormalizeCanceledWithSource_ClientCancel(t *testing.T) {
	err := NormalizeCanceledWithSource(context.Canceled, true)

	if !IsCode(err, ErrCodeAborted) {
		t.Fatalf("ErrorCode = %q, want %q", GetCode(err), ErrCodeAborted)
	}
	if got := GetHTTPStatus(err); got != HTTPStatusClientClosedRequest {
		t.Fatalf("HTTPStatus = %d, want %d", got, HTTPStatusClientClosedRequest)
	}
	if GetErrorFrom(err) != ErrorFromClient {
		t.Fatalf("ErrorFrom = %q, want %q", GetErrorFrom(err), ErrorFromClient)
	}
}

func TestNormalizeCanceledWithSource_ServerCancel(t *testing.T) {
	err := NormalizeCanceledWithSource(context.Canceled, false)

	if !IsCode(err, ErrCodeCanceled) {
		t.Fatalf("ErrorCode = %q, want %q", GetCode(err), ErrCodeCanceled)
	}
	if got := GetHTTPStatus(err); got != HTTPStatusClientClosedRequest {
		t.Fatalf("HTTPStatus = %d, want %d", got, HTTPStatusClientClosedRequest)
	}
	if GetErrorFrom(err) != ErrorFromServer {
		t.Fatalf("ErrorFrom = %q, want %q", GetErrorFrom(err), ErrorFromServer)
	}
}
