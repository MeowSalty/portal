package errors

import (
	"context"
	"testing"
)

func TestGetErrorFrom(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorFromValue
	}{
		{
			name: "nil 错误",
			err:  nil,
			want: "",
		},
		{
			name: "无上下文",
			err:  New(ErrCodeInternal, "内部错误"),
			want: "",
		},
		{
			name: "上下文无 error_from",
			err:  New(ErrCodeInternal, "内部错误").WithContext("k", "v"),
			want: "",
		},
		{
			name: "字符串 client",
			err:  New(ErrCodeInternal, "内部错误").WithContext("error_from", "client"),
			want: ErrorFromClient,
		},
		{
			name: "枚举 upstream",
			err:  New(ErrCodeInternal, "内部错误").WithContext("error_from", ErrorFromUpstream),
			want: ErrorFromUpstream,
		},
		{
			name: "非法字符串",
			err:  New(ErrCodeInternal, "内部错误").WithContext("error_from", "unknown_source"),
			want: "",
		},
		{
			name: "非法类型",
			err:  New(ErrCodeInternal, "内部错误").WithContext("error_from", 123),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetErrorFrom(tt.err); got != tt.want {
				t.Fatalf("GetErrorFrom() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsRetryable_ErrorFromMatrixAndFallback(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "client 不可重试",
			err:  New(ErrCodeUnavailable, "不可用").WithContext("error_from", string(ErrorFromClient)),
			want: false,
		},
		{
			name: "gateway 始终可重试（即使错误码不在白名单）",
			err:  New(ErrCodeInvalidArgument, "无效参数").WithContext("error_from", string(ErrorFromGateway)),
			want: true,
		},
		{
			name: "server 始终可重试",
			err:  New(ErrCodeInvalidArgument, "无效参数").WithContext("error_from", string(ErrorFromServer)),
			want: true,
		},
		{
			name: "upstream 始终可重试",
			err:  New(ErrCodeAuthenticationFailed, "认证失败").WithContext("error_from", string(ErrorFromUpstream)),
			want: true,
		},
		{
			name: "无 error_from 回退到白名单 - 可重试",
			err:  New(ErrCodeUnavailable, "不可用"),
			want: true,
		},
		{
			name: "无 error_from 回退到白名单 - 不可重试",
			err:  New(ErrCodeInvalidArgument, "无效参数"),
			want: false,
		},
		{
			name: "context.Canceled 不可重试",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "context.DeadlineExceeded 不可重试",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "NormalizeCanceled 后 ABORTED 不可重试",
			err:  NormalizeCanceled(context.Canceled),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.want {
				t.Fatalf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFromUpstreamAndGetErrorLevel_Upstream(t *testing.T) {
	err := New(ErrCodeAuthenticationFailed, "认证失败").WithContext("error_from", string(ErrorFromUpstream))

	if !isFromUpstream(err) {
		t.Fatalf("isFromUpstream() = false, want true")
	}

	if !IsUpstreamError(err) {
		t.Fatalf("IsUpstreamError() = false, want true")
	}

	if got := GetErrorLevel(err); got != ErrorLevelKey {
		t.Fatalf("GetErrorLevel() = %q, want %q", got, ErrorLevelKey)
	}
}

func TestGetErrorLevel_兼容行为_基于Resource映射(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorLevel
	}{
		{
			name: "认证失败仍为密钥层级",
			err:  New(ErrCodeAuthenticationFailed, "认证失败"),
			want: ErrorLevelKey,
		},
		{
			name: "资源未找到映射为模型层级",
			err:  New(ErrCodeNotFound, "资源未找到"),
			want: ErrorLevelModel,
		},
		{
			name: "命中平台关键词映射为平台层级",
			err:  New(ErrCodeInternal, "route backend unavailable"),
			want: ErrorLevelPlatform,
		},
		{
			name: "未知错误默认映射为模型层级",
			err:  New(ErrCodeInternal, "内部错误"),
			want: ErrorLevelModel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetErrorLevel(tt.err); got != tt.want {
				t.Fatalf("GetErrorLevel() = %d, want %d", got, tt.want)
			}
		})
	}
}
