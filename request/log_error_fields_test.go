package request

import (
	"strings"
	"testing"

	"github.com/MeowSalty/portal/errors"
)

func TestFillRequestLogErrorFields_ResponseBodyJSON_CodeString(t *testing.T) {
	log := &RequestLog{}
	err := errors.NewWithHTTPStatus(errors.ErrCodeRequestFailed, "请求失败", 429).
		WithContext("error_from", "upstream").
		WithContext("response_body", `{"error":{"type":"invalid_request_error","code":"rate_limit_exceeded","message":"Too many requests","param":"model"}}`)

	fillRequestLogErrorFields(log, err)

	if log.ErrorCode == nil || *log.ErrorCode != "REQUEST_FAILED" {
		t.Fatalf("ErrorCode 期望 REQUEST_FAILED，实际：%+v", log.ErrorCode)
	}
	if log.ErrorLevel == nil || *log.ErrorLevel != "model" {
		t.Fatalf("ErrorLevel 期望 model，实际：%+v", log.ErrorLevel)
	}
	if log.HTTPStatus == nil || *log.HTTPStatus != 429 {
		t.Fatalf("HTTPStatus 期望 429，实际：%+v", log.HTTPStatus)
	}
	if log.ErrorFrom == nil || *log.ErrorFrom != "upstream" {
		t.Fatalf("ErrorFrom 期望 upstream，实际：%+v", log.ErrorFrom)
	}
	if log.UpstreamErrorType == nil || *log.UpstreamErrorType != "invalid_request_error" {
		t.Fatalf("UpstreamErrorType 不符合预期：%+v", log.UpstreamErrorType)
	}
	if log.UpstreamErrorCode == nil || *log.UpstreamErrorCode != "rate_limit_exceeded" {
		t.Fatalf("UpstreamErrorCode 不符合预期：%+v", log.UpstreamErrorCode)
	}
	if log.UpstreamErrorParam == nil || *log.UpstreamErrorParam != "model" {
		t.Fatalf("UpstreamErrorParam 不符合预期：%+v", log.UpstreamErrorParam)
	}
	if log.UpstreamErrorMessage == nil || *log.UpstreamErrorMessage != "Too many requests" {
		t.Fatalf("UpstreamErrorMessage 不符合预期：%+v", log.UpstreamErrorMessage)
	}
	if log.ResponseBodyIsJSON == nil || !*log.ResponseBodyIsJSON {
		t.Fatalf("ResponseBodyIsJSON 期望 true，实际：%+v", log.ResponseBodyIsJSON)
	}
	if log.ResponseBodyRaw != nil {
		t.Fatalf("ResponseBodyRaw 期望为 nil，实际：%+v", log.ResponseBodyRaw)
	}
}

func TestFillRequestLogErrorFields_ResponseBodyJSON_CodeNumber(t *testing.T) {
	log := &RequestLog{}
	err := errors.NewWithHTTPStatus(errors.ErrCodeRequestFailed, "请求失败", 400).
		WithContext("error_from", "upstream").
		WithContext("response_body", `{"error":{"type":"invalid_request_error","code":400,"message":"Bad request"}}`)

	fillRequestLogErrorFields(log, err)

	if log.UpstreamErrorCode == nil || *log.UpstreamErrorCode != "400" {
		t.Fatalf("UpstreamErrorCode 期望字符串 400，实际：%+v", log.UpstreamErrorCode)
	}
	if log.UpstreamErrorParam != nil {
		t.Fatalf("UpstreamErrorParam 期望为 nil，实际：%+v", log.UpstreamErrorParam)
	}
}

func TestFillRequestLogErrorFields_ResponseBodyJSON_MissingFields(t *testing.T) {
	log := &RequestLog{}
	err := errors.NewWithHTTPStatus(errors.ErrCodeRequestFailed, "请求失败", 400).
		WithContext("response_body", `{"error":{"message":"仅消息"}}`)

	fillRequestLogErrorFields(log, err)

	if log.UpstreamErrorMessage == nil || *log.UpstreamErrorMessage != "仅消息" {
		t.Fatalf("UpstreamErrorMessage 不符合预期：%+v", log.UpstreamErrorMessage)
	}
	if log.UpstreamErrorType != nil {
		t.Fatalf("UpstreamErrorType 期望为 nil，实际：%+v", log.UpstreamErrorType)
	}
	if log.UpstreamErrorCode != nil {
		t.Fatalf("UpstreamErrorCode 期望为 nil，实际：%+v", log.UpstreamErrorCode)
	}
	if log.UpstreamErrorParam != nil {
		t.Fatalf("UpstreamErrorParam 期望为 nil，实际：%+v", log.UpstreamErrorParam)
	}
}

func TestFillRequestLogErrorFields_ResponseBodyNonJSON_504(t *testing.T) {
	log := &RequestLog{}
	err := errors.NewWithHTTPStatus(errors.ErrCodeUnavailable, "请求失败", 504).
		WithContext("response_body", "error code: 504")

	fillRequestLogErrorFields(log, err)

	if log.ResponseBodyIsJSON == nil || *log.ResponseBodyIsJSON {
		t.Fatalf("ResponseBodyIsJSON 期望 false，实际：%+v", log.ResponseBodyIsJSON)
	}
	if log.ResponseBodyRaw == nil || *log.ResponseBodyRaw != "error code: 504" {
		t.Fatalf("ResponseBodyRaw 不符合预期：%+v", log.ResponseBodyRaw)
	}
}

func TestFillRequestLogErrorFields_ResponseBodyNonJSON_Nginx(t *testing.T) {
	log := &RequestLog{}
	raw := "403 Forbidden nginx"
	err := errors.NewWithHTTPStatus(errors.ErrCodePermissionDenied, "请求失败", 403).
		WithContext("response_body", raw)

	fillRequestLogErrorFields(log, err)

	if log.ResponseBodyRaw == nil || *log.ResponseBodyRaw != raw {
		t.Fatalf("ResponseBodyRaw 不符合预期：%+v", log.ResponseBodyRaw)
	}
}

func TestFillRequestLogErrorFields_ErrorLevelModel_WhenErrorFromUpstream(t *testing.T) {
	log := &RequestLog{}
	err := errors.New(errors.ErrCodeUnavailable, "服务不可用").
		WithContext("error_from", "upstream")

	fillRequestLogErrorFields(log, err)

	if log.ErrorLevel == nil || *log.ErrorLevel != "model" {
		t.Fatalf("ErrorLevel 期望 model，实际：%+v", log.ErrorLevel)
	}
}

func TestFillRequestLogErrorFields_ErrorLevelKey_WhenAuthenticationFailed(t *testing.T) {
	log := &RequestLog{}
	err := errors.New(errors.ErrCodeAuthenticationFailed, "认证失败")

	fillRequestLogErrorFields(log, err)

	if log.ErrorLevel == nil || *log.ErrorLevel != "key" {
		t.Fatalf("ErrorLevel 期望 key，实际：%+v", log.ErrorLevel)
	}
}

func TestFillRequestLogErrorFields_ExtractRequestID(t *testing.T) {
	log := &RequestLog{}
	err := errors.NewWithHTTPStatus(errors.ErrCodeRequestFailed, "请求失败", 500).
		WithContext("response_body", `{"error":{"message":"upstream failed (request id: req_12345)"}}`)

	fillRequestLogErrorFields(log, err)

	if log.UpstreamRequestID == nil || *log.UpstreamRequestID != "req_12345" {
		t.Fatalf("UpstreamRequestID 不符合预期：%+v", log.UpstreamRequestID)
	}
}

func TestFillRequestLogErrorFields_JSONSuccessWithoutRaw(t *testing.T) {
	log := &RequestLog{}
	err := errors.NewWithHTTPStatus(errors.ErrCodeRequestFailed, "请求失败", 500).
		WithContext("response_body", `{"error":{"message":"json body"}}`)

	fillRequestLogErrorFields(log, err)

	if log.ResponseBodyIsJSON == nil || !*log.ResponseBodyIsJSON {
		t.Fatalf("ResponseBodyIsJSON 期望 true，实际：%+v", log.ResponseBodyIsJSON)
	}
	if log.ResponseBodyRaw != nil {
		t.Fatalf("ResponseBodyRaw 期望为 nil，实际：%+v", log.ResponseBodyRaw)
	}
}

func TestFillRequestLogErrorFields_JSONFailWithRaw(t *testing.T) {
	log := &RequestLog{}
	raw := "not a json body"
	err := errors.NewWithHTTPStatus(errors.ErrCodeRequestFailed, "请求失败", 500).
		WithContext("response_body", raw)

	fillRequestLogErrorFields(log, err)

	if log.ResponseBodyIsJSON == nil || *log.ResponseBodyIsJSON {
		t.Fatalf("ResponseBodyIsJSON 期望 false，实际：%+v", log.ResponseBodyIsJSON)
	}
	if log.ResponseBodyRaw == nil || *log.ResponseBodyRaw != raw {
		t.Fatalf("ResponseBodyRaw 不符合预期：%+v", log.ResponseBodyRaw)
	}
}

func TestFillRequestLogErrorFields_ClipLongFields(t *testing.T) {
	log := &RequestLog{}
	longRaw := strings.Repeat("a", requestLogLongFieldMaxLength+100)
	longMsg := strings.Repeat("b", requestLogLongFieldMaxLength+100)
	err := errors.NewWithHTTPStatus(errors.ErrCodeRequestFailed, "请求失败", 500).
		WithContext("response_body", longRaw)

	fillRequestLogErrorFields(log, err)
	if log.ResponseBodyRaw == nil {
		t.Fatalf("ResponseBodyRaw 期望不为 nil")
	}
	if got := len([]rune(*log.ResponseBodyRaw)); got != requestLogLongFieldMaxLength {
		t.Fatalf("ResponseBodyRaw 长度期望 %d，实际 %d", requestLogLongFieldMaxLength, got)
	}

	log2 := &RequestLog{}
	err2 := errors.NewWithHTTPStatus(errors.ErrCodeRequestFailed, "请求失败", 500).
		WithContext("response_body", `{"error":{"message":"`+longMsg+`"}}`)

	fillRequestLogErrorFields(log2, err2)
	if log2.UpstreamErrorMessage == nil {
		t.Fatalf("UpstreamErrorMessage 期望不为 nil")
	}
	if got := len([]rune(*log2.UpstreamErrorMessage)); got != requestLogLongFieldMaxLength {
		t.Fatalf("UpstreamErrorMessage 长度期望 %d，实际 %d", requestLogLongFieldMaxLength, got)
	}
}
