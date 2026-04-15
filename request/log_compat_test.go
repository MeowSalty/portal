package request

import (
	"context"
	"testing"

	portalErrors "github.com/MeowSalty/portal/errors"
)

// TestIsConnectionAnomaly_ConnectionStatusSet 验证 connection_status 设置时的连接异常判定。
func TestIsConnectionAnomaly_ConnectionStatusSet(t *testing.T) {
	tests := []struct {
		name             string
		connectionStatus string
		success          bool
		want             bool
	}{
		{
			name:             "completed 不算连接异常",
			connectionStatus: "completed",
			success:          true,
			want:             false,
		},
		{
			name:             "disconnected 算连接异常",
			connectionStatus: "disconnected",
			success:          false,
			want:             true,
		},
		{
			name:             "completed_then_disconnected 算连接异常",
			connectionStatus: "completed_then_disconnected",
			success:          true,
			want:             true,
		},
		{
			name:             "timed_out 算连接异常",
			connectionStatus: "timed_out",
			success:          false,
			want:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &RequestLog{
				Success:          tt.success,
				ConnectionStatus: &tt.connectionStatus,
			}
			if got := log.IsConnectionAnomaly(); got != tt.want {
				t.Fatalf("IsConnectionAnomaly() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsConnectionAnomaly_NoConnectionStatus_FallbackToSuccess 验证旧数据兼容：
// connection_status 未设置时退化为 success 判定。
func TestIsConnectionAnomaly_NoConnectionStatus_FallbackToSuccess(t *testing.T) {
	log := &RequestLog{Success: true}
	if log.IsConnectionAnomaly() {
		t.Fatal("旧数据兼容：success=true 时 IsConnectionAnomaly 应返回 false")
	}

	log2 := &RequestLog{Success: false}
	if !log2.IsConnectionAnomaly() {
		t.Fatal("旧数据兼容：success=false 时 IsConnectionAnomaly 应返回 true")
	}
}

// TestIsBusinessSuccess 验证 IsBusinessSuccess 与 Success 字段等价。
func TestIsBusinessSuccess(t *testing.T) {
	log := &RequestLog{Success: true}
	if !log.IsBusinessSuccess() {
		t.Fatal("IsBusinessSuccess() 应返回 true")
	}

	log2 := &RequestLog{Success: false}
	if log2.IsBusinessSuccess() {
		t.Fatal("IsBusinessSuccess() 应返回 false")
	}
}

// TestEnsureNonStreamDefaults_SuccessPath 验证非流式成功路径的默认值填充。
func TestEnsureNonStreamDefaults_SuccessPath(t *testing.T) {
	log := &RequestLog{}
	ensureNonStreamDefaults(log, true)

	if log.CompletionState == nil || *log.CompletionState != "completed" {
		t.Fatalf("CompletionState 期望 completed，实际：%+v", log.CompletionState)
	}
	if log.ConnectionStatus == nil || *log.ConnectionStatus != "completed" {
		t.Fatalf("ConnectionStatus 期望 completed，实际：%+v", log.ConnectionStatus)
	}
	if log.FinishStatus == nil || *log.FinishStatus != "completed" {
		t.Fatalf("FinishStatus 期望 completed，实际：%+v", log.FinishStatus)
	}
}

// TestEnsureNonStreamDefaults_FailurePath_NoHTTPStatus 验证非流式失败路径
// 且无 HTTP 响应时的默认值填充。
func TestEnsureNonStreamDefaults_FailurePath_NoHTTPStatus(t *testing.T) {
	log := &RequestLog{}
	ensureNonStreamDefaults(log, false)

	if log.CompletionState == nil || *log.CompletionState != "not_completed" {
		t.Fatalf("CompletionState 期望 not_completed，实际：%+v", log.CompletionState)
	}
	if log.ConnectionStatus == nil || *log.ConnectionStatus != "disconnected" {
		t.Fatalf("ConnectionStatus 期望 disconnected，实际：%+v", log.ConnectionStatus)
	}
	if log.FinishStatus == nil || *log.FinishStatus != "failed" {
		t.Fatalf("FinishStatus 期望 failed，实际：%+v", log.FinishStatus)
	}
}

// TestEnsureNonStreamDefaults_FailurePath_WithHTTPStatus 验证非流式失败路径
// 但有 HTTP 响应时连接状态为 completed（连接本身正常完成）。
func TestEnsureNonStreamDefaults_FailurePath_WithHTTPStatus(t *testing.T) {
	httpStatus := 429
	log := &RequestLog{HTTPStatus: &httpStatus}
	ensureNonStreamDefaults(log, false)

	if log.CompletionState == nil || *log.CompletionState != "not_completed" {
		t.Fatalf("CompletionState 期望 not_completed，实际：%+v", log.CompletionState)
	}
	if log.ConnectionStatus == nil || *log.ConnectionStatus != "completed" {
		t.Fatalf("ConnectionStatus 期望 completed（有 HTTP 响应），实际：%+v", log.ConnectionStatus)
	}
	if log.FinishStatus == nil || *log.FinishStatus != "failed" {
		t.Fatalf("FinishStatus 期望 failed，实际：%+v", log.FinishStatus)
	}
}

// TestEnsureNonStreamDefaults_DoesNotOverwriteExisting 验证不覆盖已设置的值。
func TestEnsureNonStreamDefaults_DoesNotOverwriteExisting(t *testing.T) {
	existingState := "completed_then_disconnected"
	existingConn := "completed_then_disconnected"
	existingFinish := "completed_then_disconnected"
	log := &RequestLog{
		CompletionState:  &existingState,
		ConnectionStatus: &existingConn,
		FinishStatus:     &existingFinish,
	}
	ensureNonStreamDefaults(log, true)

	if *log.CompletionState != "completed_then_disconnected" {
		t.Fatal("ensureNonStreamDefaults 不应覆盖已设置的 CompletionState")
	}
	if *log.ConnectionStatus != "completed_then_disconnected" {
		t.Fatal("ensureNonStreamDefaults 不应覆盖已设置的 ConnectionStatus")
	}
	if *log.FinishStatus != "completed_then_disconnected" {
		t.Fatal("ensureNonStreamDefaults 不应覆盖已设置的 FinishStatus")
	}
}

// TestEnsureNonStreamDefaults_NilLog 验证 nil 安全。
func TestEnsureNonStreamDefaults_NilLog(t *testing.T) {
	ensureNonStreamDefaults(nil, true)
	// 不应 panic
}

// TestFillRequestLogCancelSource_DeadlineExceeded 验证超时类取消的 cancel_source 填充。
func TestFillRequestLogCancelSource_DeadlineExceeded(t *testing.T) {
	log := &RequestLog{}
	err := portalErrors.NormalizeCanceled(context.DeadlineExceeded)
	fillRequestLogCancelSource(log, err)

	if log.CancelSource == nil || *log.CancelSource != "deadline" {
		t.Fatalf("CancelSource 期望 deadline，实际：%+v", log.CancelSource)
	}
	if log.FinishStatus == nil || *log.FinishStatus != "timed_out" {
		t.Fatalf("FinishStatus 期望 timed_out，实际：%+v", log.FinishStatus)
	}
}

// TestFillRequestLogCancelSource_ClientCancel 验证客户端取消的 cancel_source 填充。
func TestFillRequestLogCancelSource_ClientCancel(t *testing.T) {
	log := &RequestLog{}
	err := portalErrors.New(portalErrors.ErrCodeAborted, "客户端取消").
		WithContext("error_from", string(portalErrors.ErrorFromClient))
	fillRequestLogCancelSource(log, err)

	if log.CancelSource == nil || *log.CancelSource != "client" {
		t.Fatalf("CancelSource 期望 client，实际：%+v", log.CancelSource)
	}
	if log.FinishStatus == nil || *log.FinishStatus != "canceled" {
		t.Fatalf("FinishStatus 期望 canceled，实际：%+v", log.FinishStatus)
	}
}

// TestFillRequestLogCancelSource_ServerCancel 验证服务端取消的 cancel_source 填充。
func TestFillRequestLogCancelSource_ServerCancel(t *testing.T) {
	log := &RequestLog{}
	err := portalErrors.New(portalErrors.ErrCodeCanceled, "服务端取消").
		WithContext("error_from", string(portalErrors.ErrorFromServer))
	fillRequestLogCancelSource(log, err)

	if log.CancelSource == nil || *log.CancelSource != "server" {
		t.Fatalf("CancelSource 期望 server，实际：%+v", log.CancelSource)
	}
	if log.FinishStatus == nil || *log.FinishStatus != "canceled" {
		t.Fatalf("FinishStatus 期望 canceled，实际：%+v", log.FinishStatus)
	}
}

// TestFillRequestLogCancelSource_NonCanceledError 验证非取消类错误不受影响。
func TestFillRequestLogCancelSource_NonCanceledError(t *testing.T) {
	log := &RequestLog{}
	err := portalErrors.New(portalErrors.ErrCodeUnavailable, "服务不可用")
	fillRequestLogCancelSource(log, err)

	if log.CancelSource != nil {
		t.Fatalf("非取消类错误不应设置 CancelSource，实际：%+v", log.CancelSource)
	}
	if log.FinishStatus != nil {
		t.Fatalf("非取消类错误不应设置 FinishStatus，实际：%+v", log.FinishStatus)
	}
}

// TestFillRequestLogCancelSource_NilInputs 验证 nil 安全。
func TestFillRequestLogCancelSource_NilInputs(t *testing.T) {
	fillRequestLogCancelSource(nil, portalErrors.New(portalErrors.ErrCodeInternal, "错误"))
	fillRequestLogCancelSource(&RequestLog{}, nil)
	// 不应 panic
}

// TestFillRequestLogCancelSource_DoesNotOverwriteExistingCancelSource
// 验证不覆盖已设置的 cancel_source。
func TestFillRequestLogCancelSource_DoesNotOverwriteExistingCancelSource(t *testing.T) {
	existing := "client"
	log := &RequestLog{CancelSource: &existing}
	err := portalErrors.NormalizeCanceled(context.DeadlineExceeded)
	fillRequestLogCancelSource(log, err)

	if *log.CancelSource != "client" {
		t.Fatal("fillRequestLogCancelSource 不应覆盖已设置的 CancelSource")
	}
}

// TestFillRequestLogCancelSource_OverwritesFailedFinishStatus
// 验证取消类错误会覆盖 "failed" 的 finish_status。
func TestFillRequestLogCancelSource_OverwritesFailedFinishStatus(t *testing.T) {
	failed := "failed"
	log := &RequestLog{FinishStatus: &failed}
	err := portalErrors.NormalizeCanceled(context.DeadlineExceeded)
	fillRequestLogCancelSource(log, err)

	if *log.FinishStatus != "timed_out" {
		t.Fatalf("FinishStatus 期望被覆盖为 timed_out，实际：%s", *log.FinishStatus)
	}
}

// TestNewOldFieldsCoexistence 验证新旧字段并存输出。
//
// 旧字段（error_from/error_code/http_status）与新字段
// （cancel_source/connection_status/completion_state/finish_status）
// 应能同时存在于 RequestLog 中且互不干扰。
func TestNewOldFieldsCoexistence(t *testing.T) {
	errorFrom := "server"
	errorCode := "CANCELED"
	httpStatus := 499
	cancelSource := "server"
	connectionStatus := "disconnected"
	completionState := "not_completed"
	finishStatus := "canceled"

	log := &RequestLog{
		Success:          false,
		ErrorFrom:        &errorFrom,
		ErrorCode:        &errorCode,
		HTTPStatus:       &httpStatus,
		CancelSource:     &cancelSource,
		ConnectionStatus: &connectionStatus,
		CompletionState:  &completionState,
		FinishStatus:     &finishStatus,
	}

	// 旧字段查询不受破坏
	if log.ErrorFrom == nil || *log.ErrorFrom != "server" {
		t.Fatalf("旧字段 ErrorFrom 应可用，实际：%+v", log.ErrorFrom)
	}
	if log.ErrorCode == nil || *log.ErrorCode != "CANCELED" {
		t.Fatalf("旧字段 ErrorCode 应可用，实际：%+v", log.ErrorCode)
	}
	if log.HTTPStatus == nil || *log.HTTPStatus != 499 {
		t.Fatalf("旧字段 HTTPStatus 应可用，实际：%+v", log.HTTPStatus)
	}

	// 新字段并行输出
	if log.CancelSource == nil || *log.CancelSource != "server" {
		t.Fatalf("新字段 CancelSource 应可用，实际：%+v", log.CancelSource)
	}
	if log.ConnectionStatus == nil || *log.ConnectionStatus != "disconnected" {
		t.Fatalf("新字段 ConnectionStatus 应可用，实际：%+v", log.ConnectionStatus)
	}
	if log.CompletionState == nil || *log.CompletionState != "not_completed" {
		t.Fatalf("新字段 CompletionState 应可用，实际：%+v", log.CompletionState)
	}
	if log.FinishStatus == nil || *log.FinishStatus != "canceled" {
		t.Fatalf("新字段 FinishStatus 应可用，实际：%+v", log.FinishStatus)
	}

	// 报表迁移辅助方法
	if !log.IsConnectionAnomaly() {
		t.Fatal("connection_status=disconnected 应判定为连接异常")
	}
	if log.IsBusinessSuccess() {
		t.Fatal("success=false 时 IsBusinessSuccess 应返回 false")
	}
}

// TestOldFieldsOnly_LegacyQueryNotBroken 验证仅有旧字段时查询不受破坏。
//
// 模拟旧数据场景：只有 error_from/error_code/http_status/success，
// 新字段全部为 nil。
func TestOldFieldsOnly_LegacyQueryNotBroken(t *testing.T) {
	errorFrom := "upstream"
	errorCode := "REQUEST_FAILED"
	httpStatus := 500

	log := &RequestLog{
		Success:    false,
		ErrorFrom:  &errorFrom,
		ErrorCode:  &errorCode,
		HTTPStatus: &httpStatus,
	}

	// 旧字段查询正常
	if log.ErrorFrom == nil || *log.ErrorFrom != "upstream" {
		t.Fatalf("旧字段 ErrorFrom 应可用")
	}
	if log.ErrorCode == nil || *log.ErrorCode != "REQUEST_FAILED" {
		t.Fatalf("旧字段 ErrorCode 应可用")
	}
	if log.HTTPStatus == nil || *log.HTTPStatus != 500 {
		t.Fatalf("旧字段 HTTPStatus 应可用")
	}

	// 新字段为 nil 不影响旧查询
	if log.CancelSource != nil {
		t.Fatal("CancelSource 应为 nil")
	}
	if log.ConnectionStatus != nil {
		t.Fatal("ConnectionStatus 应为 nil")
	}
	if log.CompletionState != nil {
		t.Fatal("CompletionState 应为 nil")
	}
	if log.FinishStatus != nil {
		t.Fatal("FinishStatus 应为 nil")
	}

	// 旧数据兼容：IsConnectionAnomaly 退化为 success 判定
	if !log.IsConnectionAnomaly() {
		t.Fatal("旧数据兼容：success=false 时 IsConnectionAnomaly 应返回 true")
	}
}

// TestNonStreamSuccess_NewFieldsParallelOutput 验证非流式成功路径
// 新字段并行输出（ensureNonStreamDefaults 填充）。
func TestNonStreamSuccess_NewFieldsParallelOutput(t *testing.T) {
	log := &RequestLog{Success: true}
	ensureNonStreamDefaults(log, true)

	// 旧字段不受影响
	if log.ErrorFrom != nil {
		t.Fatal("成功路径不应设置 ErrorFrom")
	}

	// 新字段并行输出
	if log.CompletionState == nil || *log.CompletionState != "completed" {
		t.Fatalf("CompletionState 期望 completed，实际：%+v", log.CompletionState)
	}
	if log.ConnectionStatus == nil || *log.ConnectionStatus != "completed" {
		t.Fatalf("ConnectionStatus 期望 completed，实际：%+v", log.ConnectionStatus)
	}
	if log.FinishStatus == nil || *log.FinishStatus != "completed" {
		t.Fatalf("FinishStatus 期望 completed，实际：%+v", log.FinishStatus)
	}

	// 报表辅助
	if log.IsConnectionAnomaly() {
		t.Fatal("connection_status=completed 不应判定为连接异常")
	}
	if !log.IsBusinessSuccess() {
		t.Fatal("success=true 时 IsBusinessSuccess 应返回 true")
	}
}

// TestNonStreamFailure_WithCancel_NewFieldsParallelOutput 验证非流式失败路径
// 取消类错误的新字段并行输出。
func TestNonStreamFailure_WithCancel_NewFieldsParallelOutput(t *testing.T) {
	log := &RequestLog{Success: false}
	err := portalErrors.New(portalErrors.ErrCodeAborted, "客户端取消").
		WithContext("error_from", string(portalErrors.ErrorFromClient))

	fillRequestLogErrorFields(log, err)
	fillRequestLogCancelSource(log, err)
	ensureNonStreamDefaults(log, false)

	// 旧字段保留
	if log.ErrorFrom == nil || *log.ErrorFrom != "client" {
		t.Fatalf("旧字段 ErrorFrom 期望 client，实际：%+v", log.ErrorFrom)
	}
	if log.ErrorCode == nil || *log.ErrorCode != "ABORTED" {
		t.Fatalf("旧字段 ErrorCode 期望 ABORTED，实际：%+v", log.ErrorCode)
	}

	// 新字段并行输出
	if log.CancelSource == nil || *log.CancelSource != "client" {
		t.Fatalf("新字段 CancelSource 期望 client，实际：%+v", log.CancelSource)
	}
	if log.CompletionState == nil || *log.CompletionState != "not_completed" {
		t.Fatalf("新字段 CompletionState 期望 not_completed，实际：%+v", log.CompletionState)
	}
	if log.FinishStatus == nil || *log.FinishStatus != "canceled" {
		t.Fatalf("新字段 FinishStatus 期望 canceled，实际：%+v", log.FinishStatus)
	}

	// 报表辅助
	if !log.IsConnectionAnomaly() {
		t.Fatal("取消类错误应判定为连接异常")
	}
}

// TestReportMigration_OldFailureRateVsNewConnectionAnomalyRate
// 验证报表迁移场景：旧失败率基于 success=false，新连接异常率基于 IsConnectionAnomaly()。
//
// 场景：completed_then_disconnected
//   - 旧报表：success=true，不计入失败率
//   - 新报表：IsConnectionAnomaly()=true，计入连接异常率
func TestReportMigration_OldFailureRateVsNewConnectionAnomalyRate(t *testing.T) {
	connStatus := "completed_then_disconnected"
	compState := "completed_then_disconnected"
	finish := "completed_then_disconnected"

	log := &RequestLog{
		Success:          true,
		ConnectionStatus: &connStatus,
		CompletionState:  &compState,
		FinishStatus:     &finish,
	}

	// 旧报表：success=true，不计入失败率
	if !log.Success {
		t.Fatal("completed_then_disconnected 场景 success 应为 true，旧失败率不受影响")
	}

	// 新报表：IsConnectionAnomaly()=true，计入连接异常率
	if !log.IsConnectionAnomaly() {
		t.Fatal("completed_then_disconnected 场景应判定为连接异常")
	}
}

// TestReportMigration_SuccessNotAnomaly 验证正常成功场景：
// 旧报表不计入失败率，新报表也不计入连接异常率。
func TestReportMigration_SuccessNotAnomaly(t *testing.T) {
	connStatus := "completed"
	compState := "completed"
	finish := "completed"

	log := &RequestLog{
		Success:          true,
		ConnectionStatus: &connStatus,
		CompletionState:  &compState,
		FinishStatus:     &finish,
	}

	if !log.Success {
		t.Fatal("正常成功场景 success 应为 true")
	}
	if log.IsConnectionAnomaly() {
		t.Fatal("正常成功场景不应判定为连接异常")
	}
}
