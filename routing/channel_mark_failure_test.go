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
