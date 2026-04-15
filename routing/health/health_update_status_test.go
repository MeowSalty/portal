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

func TestUpdateStatus_HealthImpactNone_不更新健康状态(t *testing.T) {
	storage := newTestHealthStorage()
	svc, err := New(Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	// 先写入一次成功状态，确保资源存在
	if err := svc.UpdateStatus(ResourceTypeModel, 3001, true, ErrorSnapshot{}); err != nil {
		t.Fatalf("写入成功状态失败: %v", err)
	}

	// HealthImpactNone 的失败不应更新健康状态
	snapshot := ErrorSnapshot{
		Message:   "请求已取消",
		Code:      "ABORTED",
		ErrorFrom: "client",
		Impact:    HealthImpactNone,
	}
	if err := svc.UpdateStatus(ResourceTypeModel, 3001, false, snapshot); err != nil {
		t.Fatalf("UpdateStatus 失败: %v", err)
	}

	// 验证状态未被修改：SuccessCount 仍为 1，ErrorCount 仍为 0
	status, err := svc.GetStatus(ResourceTypeModel, 3001)
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}
	if status.SuccessCount != 1 {
		t.Fatalf("SuccessCount 期望 1，actual=%d", status.SuccessCount)
	}
	if status.ErrorCount != 0 {
		t.Fatalf("ErrorCount 期望 0，actual=%d", status.ErrorCount)
	}
	if status.Status != HealthStatusAvailable {
		t.Fatalf("Status 期望 Available，actual=%v", status.Status)
	}
}

func TestUpdateStatus_HealthImpactRecoverable_记录错误但不增加计数(t *testing.T) {
	storage := newTestHealthStorage()
	svc, err := New(Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	httpStatus := 504
	snapshot := ErrorSnapshot{
		Message:      "请求超时",
		Code:         "DEADLINE_EXCEEDED",
		HTTPStatus:   &httpStatus,
		ErrorFrom:    "gateway",
		CauseMessage: "context deadline exceeded",
		Impact:       HealthImpactRecoverable,
	}

	if err := svc.UpdateStatus(ResourceTypePlatform, 4001, false, snapshot); err != nil {
		t.Fatalf("UpdateStatus 失败: %v", err)
	}

	status, err := svc.GetStatus(ResourceTypePlatform, 4001)
	if err != nil {
		t.Fatalf("获取状态失败: %v", err)
	}

	// ErrorCount 不应增加
	if status.ErrorCount != 0 {
		t.Fatalf("可恢复失败 ErrorCount 期望 0，actual=%d", status.ErrorCount)
	}

	// 状态应为 Warning
	if status.Status != HealthStatusWarning {
		t.Fatalf("可恢复失败 Status 期望 Warning，actual=%v", status.Status)
	}

	// 错误信息应被记录
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

	// 历史兼容字段
	if status.LastError != snapshot.Message {
		t.Fatalf("LastError 兼容字段不符合预期，actual=%q", status.LastError)
	}
	if status.LastErrorCode != httpStatus {
		t.Fatalf("LastErrorCode 兼容字段不符合预期，actual=%d", status.LastErrorCode)
	}
}

func TestUpdateStatus_HealthImpactRecoverable_不覆盖Unavailable状态(t *testing.T) {
	storage := newTestHealthStorage()
	svc, err := New(Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	// 使用 DisableHealth 手动将资源设置为 Unavailable
	if err := svc.DisableHealth(ResourceTypeModel, 5001, "手动禁用"); err != nil {
		t.Fatalf("禁用健康状态失败: %v", err)
	}

	statusBefore, _ := svc.GetStatus(ResourceTypeModel, 5001)
	if statusBefore.Status != HealthStatusUnavailable {
		t.Fatalf("前置条件：状态应为 Unavailable，actual=%v", statusBefore.Status)
	}

	// 再写入可恢复失败，不应将 Unavailable 降级为 Warning
	if err := svc.UpdateStatus(ResourceTypeModel, 5001, false, ErrorSnapshot{
		Message:   "请求超时",
		Code:      "DEADLINE_EXCEEDED",
		ErrorFrom: "gateway",
		Impact:    HealthImpactRecoverable,
	}); err != nil {
		t.Fatalf("写入可恢复失败状态失败: %v", err)
	}

	statusAfter, _ := svc.GetStatus(ResourceTypeModel, 5001)
	if statusAfter.Status != HealthStatusUnavailable {
		t.Fatalf("可恢复失败不应将 Unavailable 降级为 Warning，actual=%v", statusAfter.Status)
	}
}

func TestUpdateStatus_HealthImpactNone_不创建新资源(t *testing.T) {
	storage := newTestHealthStorage()
	svc, err := New(Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	// 对不存在的资源写入 HealthImpactNone 失败
	snapshot := ErrorSnapshot{
		Message:   "请求已取消",
		Code:      "ABORTED",
		ErrorFrom: "client",
		Impact:    HealthImpactNone,
	}
	if err := svc.UpdateStatus(ResourceTypeModel, 6001, false, snapshot); err != nil {
		t.Fatalf("UpdateStatus 失败: %v", err)
	}

	// 不应创建任何资源记录
	if len(storage.data) != 0 {
		t.Fatalf("HealthImpactNone 不应创建新资源记录，actual=%d 条记录", len(storage.data))
	}
}
