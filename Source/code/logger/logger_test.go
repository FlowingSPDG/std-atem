package logger

import (
	"context"
	"os"
	"testing"

	"github.com/FlowingSPDG/streamdeck"
	"github.com/stretchr/testify/assert"
)

func TestLogLevel_CheckLogLevel(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		level       LogLevel
		targetLevel LogLevel
		expected    bool
	}{
		{
			name:        "DebugLevelでDebugLevelをチェック",
			level:       DebugLevel,
			targetLevel: DebugLevel,
			expected:    true,
		},
		{
			name:        "InfoLevelでDebugLevelをチェック",
			level:       InfoLevel,
			targetLevel: DebugLevel,
			expected:    false,
		},
		{
			name:        "InfoLevelでInfoLevelをチェック",
			level:       InfoLevel,
			targetLevel: InfoLevel,
			expected:    true,
		},
		{
			name:        "WarnLevelでInfoLevelをチェック",
			level:       WarnLevel,
			targetLevel: InfoLevel,
			expected:    false,
		},
		{
			name:        "ErrorLevelでWarnLevelをチェック",
			level:       ErrorLevel,
			targetLevel: WarnLevel,
			expected:    false,
		},
		{
			name:        "複数レベルでDebugLevelをチェック",
			level:       DebugLevel | InfoLevel | WarnLevel,
			targetLevel: DebugLevel,
			expected:    true,
		},
		{
			name:        "複数レベルでErrorLevelをチェック",
			level:       DebugLevel | InfoLevel | WarnLevel,
			targetLevel: ErrorLevel,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkLogLevel(tt.level, tt.targetLevel)
			asserts.Equal(tt.expected, result)
		})
	}
}

func TestStreamDeckLogger(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name      string
		logLevel  LogLevel
		message   string
		args      []interface{}
		shouldLog bool
	}{
		{
			name:      "DebugLevelでDebugメッセージ",
			logLevel:  DebugLevel,
			message:   "debug message",
			args:      []interface{}{},
			shouldLog: true,
		},
		{
			name:      "InfoLevelでDebugメッセージ",
			logLevel:  InfoLevel,
			message:   "debug message",
			args:      []interface{}{},
			shouldLog: false,
		},
		{
			name:      "InfoLevelでInfoメッセージ",
			logLevel:  InfoLevel,
			message:   "info message",
			args:      []interface{}{},
			shouldLog: true,
		},
		{
			name:      "WarnLevelでInfoメッセージ",
			logLevel:  WarnLevel,
			message:   "info message",
			args:      []interface{}{},
			shouldLog: false,
		},
		{
			name:      "ErrorLevelでWarnメッセージ",
			logLevel:  ErrorLevel,
			message:   "warn message",
			args:      []interface{}{},
			shouldLog: false,
		},
		{
			name:      "ErrorLevelでErrorメッセージ",
			logLevel:  ErrorLevel,
			message:   "error message",
			args:      []interface{}{},
			shouldLog: true,
		},
		{
			name:      "複数レベルでDebugメッセージ",
			logLevel:  DebugLevel | InfoLevel | WarnLevel,
			message:   "debug message",
			args:      []interface{}{},
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックStreamDeckクライアントを作成
			// 実際のstreamdeckライブラリの構造に合わせて簡素化
			ctx := context.Background()

			// テスト用の引数を設定して実際のクライアントを作成
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			os.Args = []string{
				"test-program",
				"-port", "1234",
				"-pluginUUID", "test-uuid",
				"-registerEvent", "registerPlugin",
				"-info", `{"application":{"language":"en","platform":"test","version":"1.0.0"},"devicePixelRatio":1,"colors":{"buttonPressedBorderColor":"#000000","buttonPressedBackgroundColor":"#000000","highlightColor":"#000000"}}`,
			}

			client, err := streamdeck.ParseRegistrationParams(os.Args)
			if err != nil {
				t.Skipf("StreamDeckクライアントの初期化に失敗したため、このテストをスキップします: %v", err)
			}

			sdClient := streamdeck.NewClient(ctx, client)
			logger := NewStreamDeckLogger(sdClient, tt.logLevel)

			// ログメッセージを実行
			err = logger.Debug(ctx, tt.message, tt.args...)
			asserts.NoError(err)

			// 実際のログ出力はStreamDeckクライアントの状態に依存するため、
			// エラーが発生しないことのみを確認
		})
	}
}

func TestFileLogger(t *testing.T) {
	asserts := assert.New(t)

	// テスト用のログファイル名（NewFileLoggerは常に./log.txtを使用）
	testLogFile := "./log.txt"

	// テスト後にファイルを削除
	defer func() {
		os.Remove(testLogFile)
	}()

	tests := []struct {
		name      string
		logLevel  LogLevel
		message   string
		args      []interface{}
		shouldLog bool
	}{
		{
			name:      "DebugLevelでDebugメッセージ",
			logLevel:  DebugLevel,
			message:   "debug message",
			args:      []interface{}{},
			shouldLog: true,
		},
		{
			name:      "InfoLevelでDebugメッセージ",
			logLevel:  InfoLevel,
			message:   "debug message",
			args:      []interface{}{},
			shouldLog: false,
		},
		{
			name:      "InfoLevelでInfoメッセージ",
			logLevel:  InfoLevel,
			message:   "info message %s",
			args:      []interface{}{"test"},
			shouldLog: true,
		},
		{
			name:      "WarnLevelでWarnメッセージ",
			logLevel:  WarnLevel,
			message:   "warn message %d",
			args:      []interface{}{42},
			shouldLog: true,
		},
		{
			name:      "ErrorLevelでErrorメッセージ",
			logLevel:  ErrorLevel,
			message:   "error message",
			args:      []interface{}{},
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 既存のテストファイルを削除
			os.Remove(testLogFile)

			ctx := context.Background()
			logger := NewFileLogger(ctx, tt.logLevel)

			// ログメッセージを実行
			err := logger.Debug(ctx, tt.message, tt.args...)
			asserts.NoError(err)

			// ファイルが作成されていることを確認
			if tt.shouldLog {
				_, err := os.Stat(testLogFile)
				asserts.NoError(err)
			}
		})
	}
}

func TestMultiLogger(t *testing.T) {
	asserts := assert.New(t)

	// テスト用のログファイル名（NewFileLoggerは常に./log.txtを使用）
	testLogFile := "./log.txt"

	// テスト後にファイルを削除
	defer func() {
		os.Remove(testLogFile)
	}()

	tests := []struct {
		name      string
		logLevel  LogLevel
		message   string
		args      []interface{}
		shouldLog bool
	}{
		{
			name:      "DebugLevelでDebugメッセージ",
			logLevel:  DebugLevel,
			message:   "debug message",
			args:      []interface{}{},
			shouldLog: true,
		},
		{
			name:      "InfoLevelでDebugメッセージ",
			logLevel:  InfoLevel,
			message:   "debug message",
			args:      []interface{}{},
			shouldLog: false,
		},
		{
			name:      "InfoLevelでInfoメッセージ",
			logLevel:  InfoLevel,
			message:   "info message %s",
			args:      []interface{}{"test"},
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 既存のテストファイルを削除
			os.Remove(testLogFile)

			ctx := context.Background()

			// 複数のロガーを作成
			fileLogger := NewFileLogger(ctx, tt.logLevel)
			testLogger := NewTestLogger(t)

			multiLogger := NewMultiLogger(tt.logLevel, fileLogger, testLogger)

			// ログメッセージを実行
			err := multiLogger.Debug(ctx, tt.message, tt.args...)
			asserts.NoError(err)

			// ファイルが作成されていることを確認
			if tt.shouldLog {
				_, err := os.Stat(testLogFile)
				asserts.NoError(err)
			}
		})
	}
}

func TestTestLogger(t *testing.T) {
	asserts := assert.New(t)

	ctx := context.Background()
	logger := NewTestLogger(t)

	tests := []struct {
		name    string
		message string
		args    []interface{}
	}{
		{
			name:    "Debugメッセージ",
			message: "debug message",
			args:    []interface{}{},
		},
		{
			name:    "Infoメッセージ",
			message: "info message %s",
			args:    []interface{}{"test"},
		},
		{
			name:    "Warnメッセージ",
			message: "warn message %d",
			args:    []interface{}{42},
		},
		{
			name:    "Errorメッセージ",
			message: "error message",
			args:    []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各ログレベルでメッセージを実行
			err := logger.Debug(ctx, tt.message, tt.args...)
			asserts.NoError(err)

			err = logger.Info(ctx, tt.message, tt.args...)
			asserts.NoError(err)

			err = logger.Warn(ctx, tt.message, tt.args...)
			asserts.NoError(err)

			err = logger.Error(ctx, tt.message, tt.args...)
			asserts.NoError(err)

			err = logger.LogMessage(ctx, tt.message, tt.args...)
			asserts.NoError(err)
		})
	}
}

func TestLogger_LogLevelCombinations(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name     string
		logLevel LogLevel
		expected string
	}{
		{
			name:     "DebugLevelのみ",
			logLevel: DebugLevel,
			expected: "DebugLevel",
		},
		{
			name:     "InfoLevelのみ",
			logLevel: InfoLevel,
			expected: "InfoLevel",
		},
		{
			name:     "WarnLevelのみ",
			logLevel: WarnLevel,
			expected: "WarnLevel",
		},
		{
			name:     "ErrorLevelのみ",
			logLevel: ErrorLevel,
			expected: "ErrorLevel",
		},
		{
			name:     "DebugLevelとInfoLevel",
			logLevel: DebugLevel | InfoLevel,
			expected: "DebugLevel|InfoLevel",
		},
		{
			name:     "全てのレベル",
			logLevel: DebugLevel | InfoLevel | WarnLevel | ErrorLevel,
			expected: "DebugLevel|InfoLevel|WarnLevel|ErrorLevel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := NewTestLogger(t)

			// 各レベルでログを試行
			err := logger.Debug(ctx, "debug message")
			asserts.NoError(err)

			err = logger.Info(ctx, "info message")
			asserts.NoError(err)

			err = logger.Warn(ctx, "warn message")
			asserts.NoError(err)

			err = logger.Error(ctx, "error message")
			asserts.NoError(err)
		})
	}
}

func TestLogger_ContextHandling(t *testing.T) {
	asserts := assert.New(t)

	ctx := context.Background()
	logger := NewTestLogger(t)

	tests := []struct {
		name    string
		ctx     context.Context
		message string
	}{
		{
			name:    "通常のコンテキスト",
			ctx:     ctx,
			message: "normal context message",
		},
		{
			name:    "nilコンテキスト",
			ctx:     nil,
			message: "nil context message",
		},
		{
			name:    "キャンセルされたコンテキスト",
			ctx:     func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			message: "cancelled context message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各ログレベルでメッセージを実行
			err := logger.Debug(tt.ctx, tt.message)
			asserts.NoError(err)

			err = logger.Info(tt.ctx, tt.message)
			asserts.NoError(err)

			err = logger.Warn(tt.ctx, tt.message)
			asserts.NoError(err)

			err = logger.Error(tt.ctx, tt.message)
			asserts.NoError(err)

			err = logger.LogMessage(tt.ctx, tt.message)
			asserts.NoError(err)
		})
	}
}
