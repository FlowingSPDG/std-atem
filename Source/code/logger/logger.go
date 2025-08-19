package logger

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/FlowingSPDG/streamdeck"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DebugLevel    LogLevel = 1 << 0 // 1
	InfoLevel     LogLevel = 1 << 1 // 2
	WarnLevel     LogLevel = 1 << 2 // 4
	ErrorLevel    LogLevel = 1 << 3 // 8
	CriticalLevel LogLevel = 1 << 4 // 16
)

// 汎用的なログレベルチェック関数
func checkLogLevel(level LogLevel, targetLevel LogLevel) bool {
	return level&targetLevel != 0
}

type Logger interface {
	LogMessage(ctx context.Context, format string, args ...any) error
	Debug(ctx context.Context, format string, args ...any) error
	Info(ctx context.Context, format string, args ...any) error
	Warn(ctx context.Context, format string, args ...any) error
	Error(ctx context.Context, format string, args ...any) error
}

// 共通のベースロガー構造体
type baseLogger struct {
	level LogLevel
}

func (l *baseLogger) shouldLog(targetLevel LogLevel) bool {
	return checkLogLevel(l.level, targetLevel)
}

type streamDeckLogger struct {
	baseLogger
	client *streamdeck.Client
}

func NewStreamDeckLogger(client *streamdeck.Client, level LogLevel) Logger {
	return &streamDeckLogger{
		baseLogger: baseLogger{level: level},
		client:     client,
	}
}

func (l *streamDeckLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	if l.client == nil {
		return nil
	}
	if !l.client.IsConnected() {
		return nil
	}
	return l.client.LogMessage(ctx, msg)
}

func (l *streamDeckLogger) Debug(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(DebugLevel) {
		return nil
	}
	return l.LogMessage(ctx, "[DEBUG] "+format, args...)
}

func (l *streamDeckLogger) Info(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(InfoLevel) {
		return nil
	}
	return l.LogMessage(ctx, "[INFO] "+format, args...)
}

func (l *streamDeckLogger) Warn(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(WarnLevel) {
		return nil
	}
	return l.LogMessage(ctx, "[WARN] "+format, args...)
}

func (l *streamDeckLogger) Error(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(ErrorLevel) {
		return nil
	}
	return l.LogMessage(ctx, "[ERROR] "+format, args...)
}

type fileLogger struct {
	baseLogger
	file *os.File
}

func NewFileLogger(ctx context.Context, level LogLevel) Logger {
	file, err := os.OpenFile("./log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return &fileLogger{
		baseLogger: baseLogger{level: level},
		file:       file,
	}
}

func (l *fileLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	_, err := l.file.WriteString(msg + "\n")
	if err != nil {
		return err
	}
	return nil
}

func (l *fileLogger) Debug(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(DebugLevel) {
		return nil
	}
	return l.LogMessage(ctx, "[DEBUG] "+format, args...)
}

func (l *fileLogger) Info(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(InfoLevel) {
		return nil
	}
	return l.LogMessage(ctx, "[INFO] "+format, args...)
}

func (l *fileLogger) Warn(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(WarnLevel) {
		return nil
	}
	return l.LogMessage(ctx, "[WARN] "+format, args...)
}

func (l *fileLogger) Error(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(ErrorLevel) {
		return nil
	}
	return l.LogMessage(ctx, "[ERROR] "+format, args...)
}

type multiLogger struct {
	baseLogger
	loggers []Logger
}

func NewMultiLogger(level LogLevel, loggers ...Logger) Logger {
	return &multiLogger{
		baseLogger: baseLogger{level: level},
		loggers:    loggers,
	}
}

func (l *multiLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	for _, logger := range l.loggers {
		logger.LogMessage(ctx, format, args...)
	}
	return nil
}

func (l *multiLogger) Debug(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(DebugLevel) {
		return nil
	}
	for _, logger := range l.loggers {
		logger.Debug(ctx, format, args...)
	}
	return nil
}

func (l *multiLogger) Info(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(InfoLevel) {
		return nil
	}
	for _, logger := range l.loggers {
		logger.Info(ctx, format, args...)
	}
	return nil
}

func (l *multiLogger) Warn(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(WarnLevel) {
		return nil
	}
	for _, logger := range l.loggers {
		logger.Warn(ctx, format, args...)
	}
	return nil
}

func (l *multiLogger) Error(ctx context.Context, format string, args ...any) error {
	if !l.shouldLog(ErrorLevel) {
		return nil
	}
	for _, logger := range l.loggers {
		logger.Error(ctx, format, args...)
	}
	return nil
}

type testLogger struct {
	t *testing.T
}

func NewTestLogger(t *testing.T) Logger {
	return &testLogger{t: t}
}

func (l *testLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	l.t.Logf(format, args...)
	return nil
}

func (l *testLogger) Debug(ctx context.Context, format string, args ...any) error {
	l.t.Logf(format, args...)
	return nil
}

func (l *testLogger) Info(ctx context.Context, format string, args ...any) error {
	l.t.Logf(format, args...)
	return nil
}

func (l *testLogger) Warn(ctx context.Context, format string, args ...any) error {
	l.t.Logf(format, args...)
	return nil
}

func (l *testLogger) Error(ctx context.Context, format string, args ...any) error {
	l.t.Logf(format, args...)
	return nil
}
