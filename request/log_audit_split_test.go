package request

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MeowSalty/portal/logger"
)

type capturedLogEntry struct {
	level string
	msg   string
	args  []any
}

type capturedLogStore struct {
	entries []capturedLogEntry
}

type capturedLogger struct {
	store *capturedLogStore
	attrs []any
}

func (l *capturedLogger) Debug(msg string, args ...any) {
	l.store.entries = append(l.store.entries, capturedLogEntry{level: "DEBUG", msg: msg, args: append([]any(nil), args...)})
}

func (l *capturedLogger) DebugContext(_ context.Context, msg string, args ...any) {
	l.Debug(msg, args...)
}

func (l *capturedLogger) Info(msg string, args ...any) {
	l.store.entries = append(l.store.entries, capturedLogEntry{level: "INFO", msg: msg, args: append([]any(nil), args...)})
}

func (l *capturedLogger) InfoContext(_ context.Context, msg string, args ...any) {
	l.Info(msg, args...)
}

func (l *capturedLogger) Warn(msg string, args ...any) {
	l.store.entries = append(l.store.entries, capturedLogEntry{level: "WARN", msg: msg, args: append([]any(nil), args...)})
}

func (l *capturedLogger) WarnContext(_ context.Context, msg string, args ...any) {
	l.Warn(msg, args...)
}

func (l *capturedLogger) Error(msg string, args ...any) {
	l.store.entries = append(l.store.entries, capturedLogEntry{level: "ERROR", msg: msg, args: append([]any(nil), args...)})
}

func (l *capturedLogger) ErrorContext(_ context.Context, msg string, args ...any) {
	l.Error(msg, args...)
}

func (l *capturedLogger) With(args ...any) logger.Logger {
	next := &capturedLogger{
		store: l.store,
		attrs: append(append([]any(nil), l.attrs...), args...),
	}
	return next
}

func (l *capturedLogger) WithGroup(_ string) logger.Logger {
	return l
}

type stubRequestLogRepo struct {
	err error
}

func (s *stubRequestLogRepo) CreateRequestLog(_ context.Context, _ *RequestLog) error {
	return s.err
}

func TestRecordRequestLog_BusinessFailureDoesNotEmitAuditError(t *testing.T) {
	log := &capturedLogger{store: &capturedLogStore{}}
	req := &Request{repo: &stubRequestLogRepo{}, logger: log}

	errMsg := "业务请求失败"
	start := time.Now().Add(-150 * time.Millisecond)
	requestLog := &RequestLog{
		Timestamp: start,
		ErrorMsg:  &errMsg,
	}

	req.recordRequestLog(requestLog, nil, false)

	if !containsMessage(log.store.entries, "DEBUG", "请求结束摘要") {
		t.Fatalf("缺少请求结束摘要日志")
	}

	if containsMessage(log.store.entries, "ERROR", "请求失败") {
		t.Fatalf("审计层不应重复输出业务失败日志")
	}

	if containsMessage(log.store.entries, "ERROR", "audit_log_persist_failed") {
		t.Fatalf("持久化成功时不应输出审计持久化失败日志")
	}
}

func TestRecordRequestLog_AuditPersistFailureOnly(t *testing.T) {
	persistErr := errors.New("db down")
	log := &capturedLogger{store: &capturedLogStore{}}
	req := &Request{repo: &stubRequestLogRepo{err: persistErr}, logger: log}

	requestLog := &RequestLog{Timestamp: time.Now().Add(-200 * time.Millisecond)}
	req.recordRequestLog(requestLog, nil, true)

	if !containsMessage(log.store.entries, "ERROR", "audit_log_persist_failed") {
		t.Fatalf("缺少审计持久化失败日志")
	}

	if containsMessage(log.store.entries, "ERROR", "请求失败") {
		t.Fatalf("不应输出业务失败日志")
	}
}

func containsMessage(entries []capturedLogEntry, level, msg string) bool {
	for _, entry := range entries {
		if entry.level == level && entry.msg == msg {
			return true
		}
	}
	return false
}
