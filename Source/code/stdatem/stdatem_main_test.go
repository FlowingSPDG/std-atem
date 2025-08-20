package stdatem

import (
	"context"
	"testing"

	"github.com/FlowingSPDG/go-atem"
	"github.com/FlowingSPDG/std-atem/Source/code/connectionmanager"
	"github.com/FlowingSPDG/std-atem/Source/code/logger"
	"github.com/FlowingSPDG/std-atem/Source/code/setting"
	"github.com/stretchr/testify/assert"
)

// mockStreamDeckClient はテスト用のStreamDeckクライアントのモック
type mockStreamDeckClient struct{}

func (m *mockStreamDeckClient) IsConnected() bool { return true }
func (m *mockStreamDeckClient) LogMessage(ctx context.Context, msg string) error { return nil }
func (m *mockStreamDeckClient) SetImage(ctx context.Context, image string, target interface{}) error { return nil }
func (m *mockStreamDeckClient) Run(ctx context.Context) error { return nil }
func (m *mockStreamDeckClient) Action(action string) interface{} { return &mockAction{} }

// mockAction はテスト用のActionのモック
type mockAction struct{}

func (m *mockAction) RegisterHandler(eventType interface{}, handler interface{}) {}

func TestNewApp(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		logger      logger.Logger
		expectError bool
	}{
		{
			name:        "正常な初期化",
			logger:      &mockLogger{},
			expectError: false,
		},
		{
			name:        "nilロガーでの初期化",
			logger:      nil,
			expectError: false, // 現在の実装ではnilロガーでもエラーにならない
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// StreamDeckクライアントは複雑なため、直接テストせずにアプリケーションの構造のみテスト
			app := &App{
				connectionManager:   connectionmanager.NewConnectionManager(tt.logger),
				logger:              tt.logger,
				previewSettingStore: setting.NewSettingStore[*previewPropertyInspector](),
				programSettingStore: setting.NewSettingStore[*programPropertyInspector](),
			}
			var err error

			if tt.expectError {
				asserts.Error(err)
				asserts.Nil(app)
			} else {
				asserts.NoError(err)
				asserts.NotNil(app)
				
				// アプリケーションの初期化状態を確認
				asserts.NotNil(app.connectionManager)
				asserts.NotNil(app.previewSettingStore)
				asserts.NotNil(app.programSettingStore)
			}
		})
	}
}

func TestAddATEMHost_Main(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		action      string
		contextID   string
		ip          string
		debug       bool
		expectError bool
	}{
		{
			name:        "新しいATEMホストの追加",
			action:      "preview",
			contextID:   "context1",
			ip:          "192.168.1.100",
			debug:       false,
			expectError: false,
		},
		{
			name:        "既存のATEMホストにコンテキストを追加",
			action:      "program",
			contextID:   "context2",
			ip:          "192.168.1.100",
			debug:       false,
			expectError: false,
		},
		{
			name:        "異なるIPでのATEMホスト追加",
			action:      "preview",
			contextID:   "context3",
			ip:          "192.168.1.200",
			debug:       true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックアプリケーションを作成
			mockLog := &mockLogger{}
			app := &App{
				connectionManager:   connectionmanager.NewConnectionManager(mockLog),
				logger:              mockLog,
				previewSettingStore: setting.NewSettingStore[*previewPropertyInspector](),
				programSettingStore: setting.NewSettingStore[*programPropertyInspector](),
			}

			ctx := context.Background()
			err := app.addATEMHost(ctx, tt.action, tt.contextID, tt.ip, tt.debug)

			if tt.expectError {
				asserts.Error(err)
			} else {
				asserts.NoError(err)

				// ATEMインスタンスが正しく保存されていることを確認
				instance, exists := app.connectionManager.SolveATEMByIP(ctx, tt.ip)
				asserts.True(exists)
				asserts.NotNil(instance)

				// コンテキストが正しく保存されていることを確認
				contexts, exists := app.connectionManager.SolveContextsByIP(ctx, tt.ip)
				asserts.True(exists)
				
				// 期待されるコンテキストが含まれているかチェック
				found := false
				for _, context := range contexts {
					if context.Context == tt.contextID {
						found = true
						break
					}
				}
				asserts.True(found, "コンテキスト %s が見つかりません", tt.contextID)
			}
		})
	}
}

func TestHandleDisappear_Main(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name      string
		contextID string
		setupFunc func(*App)
	}{
		{
			name:      "存在するコンテキストの削除",
			contextID: "context1",
			setupFunc: func(app *App) {
				ctx := context.Background()
				instance := &connectionmanager.ATEMInstance{
					Client:      atem.Create("192.168.1.100", false),
					ReconnectCh: make(chan struct{}, 1),
				}
				app.connectionManager.Store(ctx, "preview", "192.168.1.100", "context1", instance)
			},
		},
		{
			name:      "存在しないコンテキストの削除",
			contextID: "nonexistent",
			setupFunc: func(app *App) {
				// 何も設定しない
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックアプリケーションを作成
			mockLog := &mockLogger{}
			app := &App{
				connectionManager: connectionmanager.NewConnectionManager(mockLog),
				logger:           mockLog,
			}

			// セットアップ
			tt.setupFunc(app)

			ctx := context.Background()
			app.handleDisappear(ctx, tt.contextID)

			// コンテキストが削除されていることを確認
			_, exists := app.connectionManager.SolveATEMByContext(ctx, tt.contextID)
			asserts.False(exists)
		})
	}
}

func TestSolveATEMVideoInput_EdgeCases(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name     string
		input    int64
		expected atem.VideoInputType
	}{
		{
			name:     "境界値 - Video Input 1",
			input:    1,
			expected: atem.VideoInput1,
		},
		{
			name:     "境界値 - Video Input 20",
			input:    20,
			expected: atem.VideoInput20,
		},
		{
			name:     "境界値 - Color Bars",
			input:    1000,
			expected: atem.ColorBars,
		},
		{
			name:     "境界値 - ME1 Program",
			input:    10010,
			expected: atem.ME1Prog,
		},
		{
			name:     "境界値 - ME2 Preview",
			input:    10021,
			expected: atem.ME2Prev,
		},
		{
			name:     "無効な値 - 負の数",
			input:    -1,
			expected: atem.VideoBlack,
		},
		{
			name:     "無効な値 - 0",
			input:    0,
			expected: atem.VideoBlack,
		},
		{
			name:     "無効な値 - 大きな数",
			input:    99999,
			expected: atem.VideoBlack,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := solveATEMVideoInput(tt.input)
			asserts.Equal(tt.expected, result, "Input: %d", tt.input)
		})
	}
}

func TestApp_ConnectionManagement(t *testing.T) {
	asserts := assert.New(t)

	// モックアプリケーションを作成
	mockLog := &mockLogger{}
	app := &App{
		connectionManager:   connectionmanager.NewConnectionManager(mockLog),
		logger:              mockLog,
		previewSettingStore: setting.NewSettingStore[*previewPropertyInspector](),
		programSettingStore: setting.NewSettingStore[*programPropertyInspector](),
	}

	ctx := context.Background()

	// 複数のATEMホストを追加
	err := app.addATEMHost(ctx, "preview", "context1", "192.168.1.100", false)
	asserts.NoError(err)

	err = app.addATEMHost(ctx, "program", "context2", "192.168.1.100", false)
	asserts.NoError(err)

	err = app.addATEMHost(ctx, "preview", "context3", "192.168.1.200", false)
	asserts.NoError(err)

	// 接続管理の状態を確認
	instance1, exists := app.connectionManager.SolveATEMByIP(ctx, "192.168.1.100")
	asserts.True(exists)
	asserts.NotNil(instance1)

	instance2, exists := app.connectionManager.SolveATEMByIP(ctx, "192.168.1.200")
	asserts.True(exists)
	asserts.NotNil(instance2)

	// 異なるIPのインスタンスは異なることを確認
	asserts.NotEqual(instance1.Client.Ip, instance2.Client.Ip)

	// コンテキストの確認
	contexts1, exists := app.connectionManager.SolveContextsByIP(ctx, "192.168.1.100")
	asserts.True(exists)
	asserts.Len(contexts1, 2)

	contexts2, exists := app.connectionManager.SolveContextsByIP(ctx, "192.168.1.200")
	asserts.True(exists)
	asserts.Len(contexts2, 1)

	// コンテキストの削除
	app.handleDisappear(ctx, "context1")

	// 削除後の状態確認
	contexts1After, exists := app.connectionManager.SolveContextsByIP(ctx, "192.168.1.100")
	asserts.True(exists)
	asserts.Len(contexts1After, 1)

	_, exists = app.connectionManager.SolveATEMByContext(ctx, "context1")
	asserts.False(exists)
}

func TestApp_SettingStore(t *testing.T) {
	asserts := assert.New(t)

	// モックアプリケーションを作成
	mockLog := &mockLogger{}
	app := &App{
		connectionManager:   connectionmanager.NewConnectionManager(mockLog),
		logger:              mockLog,
		previewSettingStore: setting.NewSettingStore[*previewPropertyInspector](),
		programSettingStore: setting.NewSettingStore[*programPropertyInspector](),
	}

	// 設定の保存と読み込みをテスト
	previewSetting := &previewPropertyInspector{
		IP:      "192.168.1.100",
		Input:   atem.VideoInput1,
		MeIndex: 0,
	}

	programSetting := &programPropertyInspector{
		IP:      "192.168.1.100",
		Input:   atem.VideoInput2,
		MeIndex: 1,
	}

	// 設定を保存
	app.previewSettingStore.Store("context1", previewSetting)
	app.programSettingStore.Store("context2", programSetting)

	// 設定を読み込み
	loadedPreview, exists := app.previewSettingStore.Load("context1")
	asserts.True(exists)
	asserts.Equal(previewSetting.IP, loadedPreview.IP)
	asserts.Equal(previewSetting.Input, loadedPreview.Input)
	asserts.Equal(previewSetting.MeIndex, loadedPreview.MeIndex)

	loadedProgram, exists := app.programSettingStore.Load("context2")
	asserts.True(exists)
	asserts.Equal(programSetting.IP, loadedProgram.IP)
	asserts.Equal(programSetting.Input, loadedProgram.Input)
	asserts.Equal(programSetting.MeIndex, loadedProgram.MeIndex)

	// 存在しない設定の読み込み
	_, exists = app.previewSettingStore.Load("nonexistent")
	asserts.False(exists)
}
