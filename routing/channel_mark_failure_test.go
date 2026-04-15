package routing

import (
	"context"
	"testing"

	"github.com/MeowSalty/portal/errors"
	"github.com/MeowSalty/portal/routing/health"
)

type testChannelStorageKey struct {
	resourceType health.ResourceType
	resourceID   uint
}

type testChannelStorage struct {
	data map[testChannelStorageKey]*health.Health
}

func newTestChannelStorage() *testChannelStorage {
	return &testChannelStorage{data: make(map[testChannelStorageKey]*health.Health)}
}

func (s *testChannelStorage) Get(resourceType health.ResourceType, resourceID uint) (*health.Health, error) {
	return s.data[testChannelStorageKey{resourceType: resourceType, resourceID: resourceID}], nil
}

func (s *testChannelStorage) Set(status *health.Health) error {
	s.data[testChannelStorageKey{resourceType: status.ResourceType, resourceID: status.ResourceID}] = status
	return nil
}

func (s *testChannelStorage) Delete(resourceType health.ResourceType, resourceID uint) error {
	delete(s.data, testChannelStorageKey{resourceType: resourceType, resourceID: resourceID})
	return nil
}

// --- classifyHealthImpact 单元测试 ---

func TestClassifyHealthImpact_ClientCancel_无健康影响(t *testing.T) {
	// ABORTED + client 来源 → client_cancel → HealthImpactNone
	err := errors.New(errors.ErrCodeAborted, "请求已取消").
		WithContext("error_from", "client")

	impact := classifyHealthImpact(err)
	if impact != health.HealthImpactNone {
		t.Fatalf("client_cancel 期望 HealthImpactNone，actual=%v", impact)
	}
}

func TestClassifyHealthImpact_Deadline_可恢复失败(t *testing.T) {
	// DEADLINE_EXCEEDED + gateway 来源 → deadline → HealthImpactRecoverable
	err := errors.NewWithHTTPStatus(errors.ErrCodeDeadlineExceeded, "请求超时", 504).
		WithContext("error_from", "gateway")

	impact := classifyHealthImpact(err)
	if impact != health.HealthImpactRecoverable {
		t.Fatalf("deadline 期望 HealthImpactRecoverable，actual=%v", impact)
	}
}

func TestClassifyHealthImpact_ServerCancel_可恢复失败(t *testing.T) {
	// CANCELED + server 来源 → server_cancel → HealthImpactRecoverable
	err := errors.New(errors.ErrCodeCanceled, "服务端取消").
		WithContext("error_from", "server")

	impact := classifyHealthImpact(err)
	if impact != health.HealthImpactRecoverable {
		t.Fatalf("server_cancel 期望 HealthImpactRecoverable，actual=%v", impact)
	}
}

func TestClassifyHealthImpact_其他错误_完全降级(t *testing.T) {
	err := errors.NewWithHTTPStatus(errors.ErrCodeAuthenticationFailed, "认证失败", 401).
		WithContext("error_from", "server")

	impact := classifyHealthImpact(err)
	if impact != health.HealthImpactFull {
		t.Fatalf("其他错误期望 HealthImpactFull，actual=%v", impact)
	}
}

func TestClassifyHealthImpact_Nil错误_完全降级(t *testing.T) {
	impact := classifyHealthImpact(nil)
	if impact != health.HealthImpactFull {
		t.Fatalf("nil 错误期望 HealthImpactFull，actual=%v", impact)
	}
}

func TestClassifyHealthImpact_Aborted非Client_完全降级(t *testing.T) {
	// ABORTED 但来源不是 client → 不匹配 client_cancel 规则
	err := errors.New(errors.ErrCodeAborted, "操作中止").
		WithContext("error_from", "server")

	impact := classifyHealthImpact(err)
	if impact != health.HealthImpactFull {
		t.Fatalf("ABORTED+server 期望 HealthImpactFull，actual=%v", impact)
	}
}

func TestClassifyHealthImpact_DeadlineExceeded非Gateway_完全降级(t *testing.T) {
	// DEADLINE_EXCEEDED 但来源不是 gateway → 不匹配 deadline 规则
	err := errors.New(errors.ErrCodeDeadlineExceeded, "超时").
		WithContext("error_from", "server")

	impact := classifyHealthImpact(err)
	if impact != health.HealthImpactFull {
		t.Fatalf("DEADLINE_EXCEEDED+server 期望 HealthImpactFull，actual=%v", impact)
	}
}

// --- MarkFailure 健康影响联动集成测试 ---

func TestChannelMarkFailure_资源归属优先分类器_APIKey(t *testing.T) {
	storage := newTestChannelStorage()
	svc, err := health.New(health.Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	ch := &Channel{
		PlatformID:    101,
		ModelID:       202,
		APIKeyID:      303,
		healthService: svc,
	}

	requestErr := errors.NewWithHTTPStatus(errors.ErrCodeAuthenticationFailed, "认证失败", 401).
		WithContext("error_from", "server")

	ch.MarkFailure(context.TODO(), requestErr)

	if len(storage.data) != 1 {
		t.Fatalf("写入资源数不符合预期，actual=%d", len(storage.data))
	}

	status, err := svc.GetStatus(health.ResourceTypeAPIKey, ch.APIKeyID)
	if err != nil {
		t.Fatalf("获取 APIKey 健康状态失败: %v", err)
	}
	if status.ErrorCount != 1 {
		t.Fatalf("APIKey 错误计数不符合预期，actual=%d", status.ErrorCount)
	}
}

func TestChannelMarkFailure_资源归属优先分类器_模型兜底覆盖旧平台推导(t *testing.T) {
	storage := newTestChannelStorage()
	svc, err := health.New(health.Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	ch := &Channel{
		PlatformID:    11,
		ModelID:       22,
		APIKeyID:      33,
		healthService: svc,
	}

	// 旧逻辑：ErrCodeInternal 会归属平台；新逻辑应优先使用分类器的 model 兜底资源。
	requestErr := errors.New(errors.ErrCodeInternal, "内部错误")

	ch.MarkFailure(context.TODO(), requestErr)

	if len(storage.data) != 1 {
		t.Fatalf("写入资源数不符合预期，actual=%d", len(storage.data))
	}

	status, err := svc.GetStatus(health.ResourceTypeModel, ch.ModelID)
	if err != nil {
		t.Fatalf("获取模型健康状态失败: %v", err)
	}
	if status.ErrorCount != 1 {
		t.Fatalf("模型错误计数不符合预期，actual=%d", status.ErrorCount)
	}
}

func TestChannelMarkFailure_资源归属优先分类器_平台关键词归属平台(t *testing.T) {
	storage := newTestChannelStorage()
	svc, err := health.New(health.Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	ch := &Channel{
		PlatformID:    7,
		ModelID:       8,
		APIKeyID:      9,
		healthService: svc,
	}

	requestErr := errors.New(errors.ErrCodeInternal, "route backend unavailable")

	ch.MarkFailure(context.TODO(), requestErr)

	if len(storage.data) != 1 {
		t.Fatalf("写入资源数不符合预期，actual=%d", len(storage.data))
	}

	status, err := svc.GetStatus(health.ResourceTypePlatform, ch.PlatformID)
	if err != nil {
		t.Fatalf("获取平台健康状态失败: %v", err)
	}
	if status.ErrorCount != 1 {
		t.Fatalf("平台错误计数不符合预期，actual=%d", status.ErrorCount)
	}
}

func TestChannelMarkFailure_ClientCancel_不记录失败(t *testing.T) {
	storage := newTestChannelStorage()
	svc, err := health.New(health.Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	ch := &Channel{
		PlatformID:    101,
		ModelID:       202,
		APIKeyID:      303,
		healthService: svc,
	}

	// client_cancel 错误
	requestErr := errors.New(errors.ErrCodeAborted, "请求已取消").
		WithContext("error_from", "client")

	ch.MarkFailure(context.TODO(), requestErr)

	// 不应写入任何健康状态
	if len(storage.data) != 0 {
		t.Fatalf("client_cancel 不应写入健康状态，actual=%d 条记录", len(storage.data))
	}
}

func TestChannelMarkFailure_Deadline_可恢复失败(t *testing.T) {
	storage := newTestChannelStorage()
	svc, err := health.New(health.Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	ch := &Channel{
		PlatformID:    101,
		ModelID:       202,
		APIKeyID:      303,
		healthService: svc,
	}

	// deadline 错误
	requestErr := errors.NewWithHTTPStatus(errors.ErrCodeDeadlineExceeded, "请求超时", 504).
		WithContext("error_from", "gateway")

	ch.MarkFailure(context.TODO(), requestErr)

	// 应写入健康状态，但 ErrorCount 不增加
	status, err := svc.GetStatus(health.ResourceTypeModel, ch.ModelID)
	if err != nil {
		t.Fatalf("获取模型健康状态失败: %v", err)
	}
	if status.ErrorCount != 0 {
		t.Fatalf("deadline 可恢复失败 ErrorCount 期望 0，actual=%d", status.ErrorCount)
	}
	if status.Status != health.HealthStatusWarning {
		t.Fatalf("deadline 可恢复失败 Status 期望 Warning，actual=%v", status.Status)
	}
	if status.LastStructuredErrorCode != "DEADLINE_EXCEEDED" {
		t.Fatalf("LastStructuredErrorCode 期望 DEADLINE_EXCEEDED，actual=%q", status.LastStructuredErrorCode)
	}
}

func TestChannelMarkFailure_ServerCancel_可恢复失败(t *testing.T) {
	storage := newTestChannelStorage()
	svc, err := health.New(health.Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	ch := &Channel{
		PlatformID:    101,
		ModelID:       202,
		APIKeyID:      303,
		healthService: svc,
	}

	// server_cancel 错误
	requestErr := errors.New(errors.ErrCodeCanceled, "服务端取消").
		WithContext("error_from", "server")

	ch.MarkFailure(context.TODO(), requestErr)

	// 应写入健康状态，但 ErrorCount 不增加
	status, err := svc.GetStatus(health.ResourceTypeModel, ch.ModelID)
	if err != nil {
		t.Fatalf("获取模型健康状态失败: %v", err)
	}
	if status.ErrorCount != 0 {
		t.Fatalf("server_cancel 可恢复失败 ErrorCount 期望 0，actual=%d", status.ErrorCount)
	}
	if status.Status != health.HealthStatusWarning {
		t.Fatalf("server_cancel 可恢复失败 Status 期望 Warning，actual=%v", status.Status)
	}
}

func TestChannelMarkFailure_CompletedThenDisconnected_不记录失败(t *testing.T) {
	storage := newTestChannelStorage()
	svc, err := health.New(health.Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	ch := &Channel{
		PlatformID:    101,
		ModelID:       202,
		APIKeyID:      303,
		healthService: svc,
	}

	// completed_then_disconnected 在系统中表现为 client_cancel（ABORTED + client）
	requestErr := errors.NewWithHTTPStatus(errors.ErrCodeAborted, "请求已取消", 499).
		WithContext("error_from", "client")

	ch.MarkFailure(context.TODO(), requestErr)

	// 不应写入任何健康状态
	if len(storage.data) != 0 {
		t.Fatalf("completed_then_disconnected 不应写入健康状态，actual=%d 条记录", len(storage.data))
	}
}
