package di

import (
	"context"
	"os"
	"testing"

	"github.com/FlowingSPDG/std-atem/Source/code/logger"
	"github.com/stretchr/testify/assert"
)

func TestInitializeStreamDeckClient(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		setupArgs   func()
		expectError bool
	}{
		{
			name: "正常な引数での初期化",
			setupArgs: func() {
				// テスト用の引数を設定
				os.Args = []string{
					"test-program",
					"-port", "1234",
					"-pluginUUID", "test-uuid",
					"-registerEvent", "registerPlugin",
					"-info", `{"application":{"language":"en","platform":"test","version":"1.0.0"},"devicePixelRatio":1,"colors":{"buttonPressedBorderColor":"#000000","buttonPressedBackgroundColor":"#000000","buttonMouseOverBorderColor":"#000000","buttonMouseOverBackgroundColor":"#000000","highlightColor":"#000000"}}`,
				}
			},
			expectError: false,
		},
		{
			name: "引数が不足している場合",
			setupArgs: func() {
				os.Args = []string{"test-program"}
			},
			expectError: true,
		},
		{
			name: "無効なJSON形式のinfo",
			setupArgs: func() {
				os.Args = []string{
					"test-program",
					"-port", "1234",
					"-pluginUUID", "test-uuid",
					"-registerEvent", "registerPlugin",
					"-info", `invalid json`,
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト前の設定
			tt.setupArgs()

			ctx := context.Background()
			client, err := InitializeStreamDeckClient(ctx)

			if tt.expectError {
				asserts.Error(err)
				asserts.Nil(client)
			} else {
				asserts.NoError(err)
				asserts.NotNil(client)
			}
		})
	}
}

func TestInitializeStreamDeckLogger(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		logLevel    logger.LogLevel
		expectError bool
	}{
		{
			name:        "DebugLevelでの初期化",
			logLevel:    logger.DebugLevel,
			expectError: false,
		},
		{
			name:        "InfoLevelでの初期化",
			logLevel:    logger.InfoLevel,
			expectError: false,
		},
		{
			name:        "WarnLevelでの初期化",
			logLevel:    logger.WarnLevel,
			expectError: false,
		},
		{
			name:        "ErrorLevelでの初期化",
			logLevel:    logger.ErrorLevel,
			expectError: false,
		},
		{
			name:        "複数レベルでの初期化",
			logLevel:    logger.DebugLevel | logger.InfoLevel | logger.WarnLevel,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用のStreamDeckクライアントを作成
			ctx := context.Background()
			
			// テスト用の引数を設定
			os.Args = []string{
				"test-program",
				"-port", "1234",
				"-pluginUUID", "test-uuid",
				"-registerEvent", "registerPlugin",
				"-info", `{"application":{"language":"en","platform":"test","version":"1.0.0"},"devicePixelRatio":1,"colors":{"buttonPressedBorderColor":"#000000","buttonPressedBackgroundColor":"#000000","buttonMouseOverBorderColor":"#000000","buttonMouseOverBackgroundColor":"#000000","highlightColor":"#000000"}}`,
			}

			client, err := InitializeStreamDeckClient(ctx)
			if err != nil {
				t.Skipf("StreamDeckクライアントの初期化に失敗したため、このテストをスキップします: %v", err)
			}

			logger, err := InitializeStreamDeckLogger(ctx, client, tt.logLevel)

			if tt.expectError {
				asserts.Error(err)
				asserts.Nil(logger)
			} else {
				asserts.NoError(err)
				asserts.NotNil(logger)
			}
		})
	}
}

func TestInitializeStreamDeckLogger_WithNilClient(t *testing.T) {
	asserts := assert.New(t)

	ctx := context.Background()
	logger, err := InitializeStreamDeckLogger(ctx, nil, logger.DebugLevel)

	// nilクライアントでもエラーにならないことを確認
	asserts.NoError(err)
	asserts.NotNil(logger)
}

func TestInitializeStreamDeckLogger_LogLevelValidation(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		logLevel    logger.LogLevel
		description string
	}{
		{
			name:        "DebugLevel",
			logLevel:    logger.DebugLevel,
			description: "デバッグレベルのログ",
		},
		{
			name:        "InfoLevel",
			logLevel:    logger.InfoLevel,
			description: "情報レベルのログ",
		},
		{
			name:        "WarnLevel",
			logLevel:    logger.WarnLevel,
			description: "警告レベルのログ",
		},
		{
			name:        "ErrorLevel",
			logLevel:    logger.ErrorLevel,
			description: "エラーレベルのログ",
		},
		{
			name:        "CriticalLevel",
			logLevel:    logger.CriticalLevel,
			description: "重大エラーレベルのログ",
		},
		{
			name:        "複数レベル",
			logLevel:    logger.DebugLevel | logger.InfoLevel | logger.WarnLevel,
			description: "複数のログレベル",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			
			// テスト用の引数を設定
			os.Args = []string{
				"test-program",
				"-port", "1234",
				"-pluginUUID", "test-uuid",
				"-registerEvent", "registerPlugin",
				"-info", `{"application":{"language":"en","platform":"test","version":"1.0.0"},"devicePixelRatio":1,"colors":{"buttonPressedBorderColor":"#000000","buttonPressedBackgroundColor":"#000000","buttonMouseOverBorderColor":"#000000","buttonMouseOverBackgroundColor":"#000000","highlightColor":"#000000"}}`,
			}

			client, err := InitializeStreamDeckClient(ctx)
			if err != nil {
				t.Skipf("StreamDeckクライアントの初期化に失敗したため、このテストをスキップします: %v", err)
			}

			logger, err := InitializeStreamDeckLogger(ctx, client, tt.logLevel)

			asserts.NoError(err)
			asserts.NotNil(logger)

			// ロガーが正しく動作することを確認
			err = logger.Debug(ctx, "test debug message")
			asserts.NoError(err)

			err = logger.Info(ctx, "test info message")
			asserts.NoError(err)

			err = logger.Warn(ctx, "test warn message")
			asserts.NoError(err)

			err = logger.Error(ctx, "test error message")
			asserts.NoError(err)
		})
	}
}

func TestInitializeStreamDeckClient_InvalidArguments(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		args        []string
		description string
	}{
		{
			name:        "空の引数",
			args:        []string{"test-program"},
			description: "引数が不足している場合",
		},
		{
			name:        "portが不足",
			args:        []string{"test-program", "-pluginUUID", "test-uuid"},
			description: "port引数が不足している場合",
		},
		{
			name:        "pluginUUIDが不足",
			args:        []string{"test-program", "-port", "1234"},
			description: "pluginUUID引数が不足している場合",
		},
		{
			name:        "registerEventが不足",
			args:        []string{"test-program", "-port", "1234", "-pluginUUID", "test-uuid"},
			description: "registerEvent引数が不足している場合",
		},
		{
			name:        "infoが不足",
			args:        []string{"test-program", "-port", "1234", "-pluginUUID", "test-uuid", "-registerEvent", "registerPlugin"},
			description: "info引数が不足している場合",
		},
		{
			name:        "無効なport番号",
			args:        []string{"test-program", "-port", "invalid", "-pluginUUID", "test-uuid", "-registerEvent", "registerPlugin", "-info", `{"application":{"language":"en","platform":"test","version":"1.0.0"},"devicePixelRatio":1,"colors":{"buttonPressedBorderColor":"#000000","buttonPressedBackgroundColor":"#000000","buttonMouseOverBorderColor":"#000000","buttonMouseOverBackgroundColor":"#000000","highlightColor":"#000000"}}`},
			description: "無効なport番号の場合",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用の引数を設定
			os.Args = tt.args

			ctx := context.Background()
			client, err := InitializeStreamDeckClient(ctx)

			asserts.Error(err)
			asserts.Nil(client)
		})
	}
}

func TestInitializeStreamDeckClient_ContextHandling(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		ctx         context.Context
		description string
	}{
		{
			name:        "通常のコンテキスト",
			ctx:         context.Background(),
			description: "通常のコンテキストでの初期化",
		},
		{
			name:        "キャンセルされたコンテキスト",
			ctx:         func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			description: "キャンセルされたコンテキストでの初期化",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用の引数を設定
			os.Args = []string{
				"test-program",
				"-port", "1234",
				"-pluginUUID", "test-uuid",
				"-registerEvent", "registerPlugin",
				"-info", `{"application":{"language":"en","platform":"test","version":"1.0.0"},"devicePixelRatio":1,"colors":{"buttonPressedBorderColor":"#000000","buttonPressedBackgroundColor":"#000000","buttonMouseOverBorderColor":"#000000","buttonMouseOverBackgroundColor":"#000000","highlightColor":"#000000"}}`,
			}

			client, err := InitializeStreamDeckClient(tt.ctx)

			// コンテキストの状態に関係なく、引数が正しければ初期化は成功する
			asserts.NoError(err)
			asserts.NotNil(client)
		})
	}
}

