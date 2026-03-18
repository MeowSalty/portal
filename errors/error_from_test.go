package errors

import "testing"

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
			name: "枚举 upstream_dependency",
			err:  New(ErrCodeInternal, "内部错误").WithContext("error_from", ErrorFromUpstreamDependency),
			want: ErrorFromUpstreamDependency,
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
			name: "server 按白名单可重试",
			err:  New(ErrCodeUnavailable, "不可用").WithContext("error_from", string(ErrorFromServer)),
			want: true,
		},
		{
			name: "upstream 始终可重试",
			err:  New(ErrCodeInvalidArgument, "无效参数").WithContext("error_from", string(ErrorFromUpstream)),
			want: true,
		},
		{
			name: "upstream_dependency 始终可重试",
			err:  New(ErrCodeAuthenticationFailed, "认证失败").WithContext("error_from", string(ErrorFromUpstreamDependency)),
			want: true,
		},
		{
			name: "无 error_from 回退到白名单-可重试",
			err:  New(ErrCodeUnavailable, "不可用"),
			want: true,
		},
		{
			name: "无 error_from 回退到白名单-不可重试",
			err:  New(ErrCodeInvalidArgument, "无效参数"),
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

func TestIsFromUpstreamAndGetErrorLevel_UpstreamDependency(t *testing.T) {
	err := New(ErrCodeAuthenticationFailed, "认证失败").WithContext("error_from", string(ErrorFromUpstreamDependency))

	if !isFromUpstream(err) {
		t.Fatalf("isFromUpstream() = false, want true")
	}

	if !IsUpstreamError(err) {
		t.Fatalf("IsUpstreamError() = false, want true")
	}

	if got := GetErrorLevel(err); got != ErrorLevelModel {
		t.Fatalf("GetErrorLevel() = %q, want %q", got, ErrorLevelModel)
	}
}
