package logger

import (
	"context"
)

// Level 定义日志级别
type Level int

const (
	// LevelDebug 调试级别
	LevelDebug Level = iota - 1
	// LevelInfo 信息级别
	LevelInfo
	// LevelWarn 警告级别
	LevelWarn
	// LevelError 错误级别
	LevelError
)

// String 返回日志级别的字符串表示
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger 定义日志接口
//
// 该接口设计参考 log/slog，支持结构化日志记录。
// 用户可以实现此接口来集成自己的日志系统（如 zap、zerolog 等）。
type Logger interface {
	// Debug 记录调试级别日志
	Debug(msg string, args ...any)

	// DebugContext 记录带上下文的调试级别日志
	DebugContext(ctx context.Context, msg string, args ...any)

	// Info 记录信息级别日志
	Info(msg string, args ...any)

	// InfoContext 记录带上下文的信息级别日志
	InfoContext(ctx context.Context, msg string, args ...any)

	// Warn 记录警告级别日志
	Warn(msg string, args ...any)

	// WarnContext 记录带上下文的警告级别日志
	WarnContext(ctx context.Context, msg string, args ...any)

	// Error 记录错误级别日志
	Error(msg string, args ...any)

	// ErrorContext 记录带上下文的错误级别日志
	ErrorContext(ctx context.Context, msg string, args ...any)

	// With 返回一个新的 Logger，包含指定的属性
	//
	// args 参数应该是成对的键值对，例如：
	//   logger.With("key1", "value1", "key2", 123)
	With(args ...any) Logger

	// WithGroup 返回一个新的 Logger，后续日志将属于指定的组
	WithGroup(name string) Logger
}

// nopLogger 是一个空操作的日志实现
//
// 当用户未提供日志实现时使用此默认实现
type nopLogger struct{}

// NewNopLogger 创建一个空操作的日志记录器
func NewNopLogger() Logger {
	return &nopLogger{}
}

func (n *nopLogger) Debug(msg string, args ...any)                             {}
func (n *nopLogger) DebugContext(ctx context.Context, msg string, args ...any) {}
func (n *nopLogger) Info(msg string, args ...any)                              {}
func (n *nopLogger) InfoContext(ctx context.Context, msg string, args ...any)  {}
func (n *nopLogger) Warn(msg string, args ...any)                              {}
func (n *nopLogger) WarnContext(ctx context.Context, msg string, args ...any)  {}
func (n *nopLogger) Error(msg string, args ...any)                             {}
func (n *nopLogger) ErrorContext(ctx context.Context, msg string, args ...any) {}
func (n *nopLogger) With(args ...any) Logger                                   { return n }
func (n *nopLogger) WithGroup(name string) Logger                              { return n }

// defaultLogger 全局默认日志记录器
var defaultLogger Logger = NewNopLogger()

// SetDefault 设置全局默认日志记录器
func SetDefault(l Logger) {
	if l != nil {
		defaultLogger = l
	}
}

// Default 返回全局默认日志记录器
func Default() Logger {
	return defaultLogger
}
