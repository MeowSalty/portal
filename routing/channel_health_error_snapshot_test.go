package routing

import (
	stdErrors "errors"
	"testing"

	"github.com/MeowSalty/portal/errors"
)

func TestBuildHealthErrorSnapshot_HTTPError(t *testing.T) {
	root := stdErrors.New("upstream 429")
	err := errors.WrapWithHTTPStatus(errors.ErrCodeRequestFailed, "请求失败", root, 429).
		WithContext("error_from", "server")

	snapshot := buildHealthErrorSnapshot(err)

	if snapshot.Message != "请求失败" {
		t.Fatalf("Message 不符合预期，actual=%q", snapshot.Message)
	}
	if snapshot.Code != "REQUEST_FAILED" {
		t.Fatalf("Code 不符合预期，actual=%q", snapshot.Code)
	}
	if snapshot.HTTPStatus == nil || *snapshot.HTTPStatus != 429 {
		t.Fatalf("HTTPStatus 不符合预期，actual=%v", snapshot.HTTPStatus)
	}
	if snapshot.ErrorFrom != "server" {
		t.Fatalf("ErrorFrom 不符合预期，actual=%q", snapshot.ErrorFrom)
	}
	if snapshot.CauseMessage != root.Error() {
		t.Fatalf("CauseMessage 不符合预期，actual=%q", snapshot.CauseMessage)
	}
}

func TestBuildHealthErrorSnapshot_ConnectionError(t *testing.T) {
	root := stdErrors.New("dial tcp 127.0.0.1:443: connectex: connection refused")
	err := errors.Wrap(errors.ErrCodeUnavailable, "连接上游失败", root).
		WithContext("error_from", "gateway")

	snapshot := buildHealthErrorSnapshot(err)

	if snapshot.Message != "连接上游失败" {
		t.Fatalf("Message 不符合预期，actual=%q", snapshot.Message)
	}
	if snapshot.Code != "UNAVAILABLE" {
		t.Fatalf("Code 不符合预期，actual=%q", snapshot.Code)
	}
	if snapshot.HTTPStatus != nil {
		t.Fatalf("HTTPStatus 期望为 nil，actual=%v", snapshot.HTTPStatus)
	}
	if snapshot.ErrorFrom != "gateway" {
		t.Fatalf("ErrorFrom 不符合预期，actual=%q", snapshot.ErrorFrom)
	}
	if snapshot.CauseMessage != root.Error() {
		t.Fatalf("CauseMessage 不符合预期，actual=%q", snapshot.CauseMessage)
	}
}
