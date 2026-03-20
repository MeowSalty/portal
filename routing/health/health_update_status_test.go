package health

import "testing"

type testHealthStorageKey struct {
	resourceType ResourceType
	resourceID   uint
}

type testHealthStorage struct {
	data map[testHealthStorageKey]*Health
}

func newTestHealthStorage() *testHealthStorage {
	return &testHealthStorage{data: make(map[testHealthStorageKey]*Health)}
}

func (s *testHealthStorage) Get(resourceType ResourceType, resourceID uint) (*Health, error) {
	return s.data[testHealthStorageKey{resourceType: resourceType, resourceID: resourceID}], nil
}

func (s *testHealthStorage) Set(status *Health) error {
	s.data[testHealthStorageKey{resourceType: status.ResourceType, resourceID: status.ResourceID}] = status
	return nil
}

func (s *testHealthStorage) Delete(resourceType ResourceType, resourceID uint) error {
	delete(s.data, testHealthStorageKey{resourceType: resourceType, resourceID: resourceID})
	return nil
}

func TestUpdateStatus_FailureWritesStructuredAndLegacyFields_WithHTTPStatus(t *testing.T) {
	storage := newTestHealthStorage()
	svc, err := New(Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	httpStatus := 429
	snapshot := ErrorSnapshot{
		Message:      "请求失败",
		Code:         "REQUEST_FAILED",
		HTTPStatus:   &httpStatus,
		ErrorFrom:    "server",
		CauseMessage: "upstream 429",
	}

	if err := svc.UpdateStatus(ResourceTypePlatform, 1001, false, snapshot); err != nil {
		t.Fatalf("UpdateStatus 失败: %v", err)
	}

	status, err := svc.GetStatus(ResourceTypePlatform, 1001)
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}

	if status.LastErrorMessage != snapshot.Message {
		t.Fatalf("LastErrorMessage 不符合预期，actual=%q", status.LastErrorMessage)
	}
	if status.LastStructuredErrorCode != snapshot.Code {
		t.Fatalf("LastStructuredErrorCode 不符合预期，actual=%q", status.LastStructuredErrorCode)
	}
	if status.LastHTTPStatus == nil || *status.LastHTTPStatus != httpStatus {
		t.Fatalf("LastHTTPStatus 不符合预期，actual=%v", status.LastHTTPStatus)
	}
	if status.LastErrorFrom != snapshot.ErrorFrom {
		t.Fatalf("LastErrorFrom 不符合预期，actual=%q", status.LastErrorFrom)
	}
	if status.LastCauseMessage != snapshot.CauseMessage {
		t.Fatalf("LastCauseMessage 不符合预期，actual=%q", status.LastCauseMessage)
	}

	if status.LastError != snapshot.Message {
		t.Fatalf("LastError 兼容字段不符合预期，actual=%q", status.LastError)
	}
	if status.LastErrorCode != httpStatus {
		t.Fatalf("LastErrorCode 兼容字段不符合预期，actual=%d", status.LastErrorCode)
	}
}

func TestUpdateStatus_SuccessClearsStructuredAndLegacyFields(t *testing.T) {
	storage := newTestHealthStorage()
	svc, err := New(Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	if err := svc.UpdateStatus(ResourceTypeModel, 2002, false, ErrorSnapshot{
		Message:      "连接失败",
		Code:         "UNAVAILABLE",
		ErrorFrom:    "gateway",
		CauseMessage: "dial tcp 127.0.0.1:443: connectex: connection refused",
	}); err != nil {
		t.Fatalf("写入失败状态失败: %v", err)
	}

	if err := svc.UpdateStatus(ResourceTypeModel, 2002, true, ErrorSnapshot{}); err != nil {
		t.Fatalf("写入成功状态失败: %v", err)
	}

	status, err := svc.GetStatus(ResourceTypeModel, 2002)
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}

	if status.LastError != "" {
		t.Fatalf("LastError 期望清空，actual=%q", status.LastError)
	}
	if status.LastErrorCode != 0 {
		t.Fatalf("LastErrorCode 期望清空为 0，actual=%d", status.LastErrorCode)
	}
	if status.LastErrorMessage != "" {
		t.Fatalf("LastErrorMessage 期望清空，actual=%q", status.LastErrorMessage)
	}
	if status.LastStructuredErrorCode != "" {
		t.Fatalf("LastStructuredErrorCode 期望清空，actual=%q", status.LastStructuredErrorCode)
	}
	if status.LastHTTPStatus != nil {
		t.Fatalf("LastHTTPStatus 期望清空为 nil，actual=%v", status.LastHTTPStatus)
	}
	if status.LastErrorFrom != "" {
		t.Fatalf("LastErrorFrom 期望清空，actual=%q", status.LastErrorFrom)
	}
	if status.LastCauseMessage != "" {
		t.Fatalf("LastCauseMessage 期望清空，actual=%q", status.LastCauseMessage)
	}
}
