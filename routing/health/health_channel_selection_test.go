package health

import (
	"testing"
	"time"
)

func TestGetChannelHealthAndLastTryTimes_AvailableMatchesExistingAPIs(t *testing.T) {
	storage := newTestHealthStorage()
	svc, err := New(Config{Storage: storage})
	if err != nil {
		t.Fatalf("创建健康服务失败: %v", err)
	}

	base := time.Now().Add(-2 * time.Minute)
	platformAt := base.Add(10 * time.Second)
	modelAt := base.Add(20 * time.Second)
	keyAt := base.Add(30 * time.Second)

	if err := storage.Set(&Health{
		ResourceType: ResourceTypePlatform,
		ResourceID:   11,
		Status:       HealthStatusAvailable,
		LastCheckAt:  platformAt,
		CreatedAt:    base,
		UpdatedAt:    base,
	}); err != nil {
		t.Fatalf("写入平台状态失败: %v", err)
	}
	if err := storage.Set(&Health{
		ResourceType: ResourceTypeModel,
		ResourceID:   22,
		Status:       HealthStatusAvailable,
		LastCheckAt:  modelAt,
		CreatedAt:    base,
		UpdatedAt:    base,
	}); err != nil {
		t.Fatalf("写入模型状态失败: %v", err)
	}
	if err := storage.Set(&Health{
		ResourceType: ResourceTypeAPIKey,
		ResourceID:   33,
		Status:       HealthStatusAvailable,
		LastCheckAt:  keyAt,
		CreatedAt:    base,
		UpdatedAt:    base,
	}); err != nil {
		t.Fatalf("写入密钥状态失败: %v", err)
	}

	gotResult, gotPlatform, gotModel, gotKey := svc.GetChannelHealthAndLastTryTimes(11, 22, 33)
	expectResult := svc.CheckChannelHealth(11, 22, 33)
	expectPlatform, expectModel, expectKey, err := svc.GetLastTryTimes(11, 22, 33)
	if err != nil {
		t.Fatalf("获取最近尝试时间失败: %v", err)
	}

	if gotResult.Status != expectResult.Status {
		t.Fatalf("通道状态不符合预期，actual=%v expected=%v", gotResult.Status, expectResult.Status)
	}
	if !gotResult.LastCheckAt.Equal(expectResult.LastCheckAt) {
		t.Fatalf("最后检查时间不符合预期，actual=%v expected=%v", gotResult.LastCheckAt, expectResult.LastCheckAt)
	}

	if !gotPlatform.Equal(expectPlatform) {
		t.Fatalf("平台最近尝试时间不符合预期，actual=%v expected=%v", gotPlatform, expectPlatform)
	}
	if !gotModel.Equal(expectModel) {
		t.Fatalf("模型最近尝试时间不符合预期，actual=%v expected=%v", gotModel, expectModel)
	}
	if !gotKey.Equal(expectKey) {
		t.Fatalf("密钥最近尝试时间不符合预期，actual=%v expected=%v", gotKey, expectKey)
	}
}
