package stdatem

import (
	"context"
	"testing"

	"github.com/FlowingSPDG/std-atem/Source/code/connectionmanager"
	"github.com/FlowingSPDG/std-atem/Source/code/setting"
	"github.com/FlowingSPDG/streamdeck"
	"github.com/stretchr/testify/assert"
)

// mockLogger モックロガー
type mockLogger struct {
	debugCalls []string
	errorCalls []string
}

func (m *mockLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	return nil
}

func (m *mockLogger) Debug(ctx context.Context, format string, args ...any) error {
	m.debugCalls = append(m.debugCalls, format)
	return nil
}

func (m *mockLogger) Info(ctx context.Context, format string, args ...any) error {
	return nil
}

func (m *mockLogger) Warn(ctx context.Context, format string, args ...any) error {
	return nil
}

func (m *mockLogger) Error(ctx context.Context, format string, args ...any) error {
	m.errorCalls = append(m.errorCalls, format)
	return nil
}

// buildWillAppearPayload 型安全なWillAppearペイロードを作成
func buildWillAppearPayload[T any](settings T) streamdeck.WillAppearPayload[T] {
	return streamdeck.WillAppearPayload[T]{
		Settings: settings,
	}
}

// buildWillDisappearPayload 型安全なWillDisappearペイロードを作成
func buildWillDisappearPayload[T any](settings T) streamdeck.WillDisappearPayload[T] {
	return streamdeck.WillDisappearPayload[T]{
		Settings: settings,
	}
}

// buildKeyDownPayload 型安全なKeyDownペイロードを作成
func buildKeyDownPayload[T any](settings T) streamdeck.KeyDownPayload[T] {
	return streamdeck.KeyDownPayload[T]{
		Settings: settings,
	}
}

// buildDidReceiveSettingsPayload 型安全なDidReceiveSettingsペイロードを作成
func buildDidReceiveSettingsPayload[T any](settings T) streamdeck.DidReceiveSettingsPayload[T] {
	return streamdeck.DidReceiveSettingsPayload[T]{
		Settings: settings,
	}
}

func TestHandleWillAppear_TableDriven(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		handlerType string // preview | program | auto
		contextID   string
		settings    interface{}
		expectErr   bool
		validate    func(*testing.T, *App, string)
	}{
		{
			name:        "preview: normal appear",
			handlerType: "preview",
			contextID:   "ctx-prev-1",
			settings: &PreviewPropertyInspector{
				IP:      "192.168.20.10",
				Input:   "1",
				MeIndex: "0",
			},
			expectErr: false,
		},
		{
			name:        "program: normal appear",
			handlerType: "program",
			contextID:   "ctx-prog-1",
			settings: &ProgramPropertyInspector{
				IP:      "192.168.20.11",
				Input:   "2",
				MeIndex: "1",
			},
			expectErr: false,
		},
		{
			name:        "auto: normal appear",
			handlerType: "auto",
			contextID:   "ctx-auto-1",
			settings: &AutoPropertyInspector{
				IP: "192.168.20.12",
			},
			expectErr: false,
		},
		{
			name:        "preview: invalid input",
			handlerType: "preview",
			contextID:   "ctx-prev-invalid",
			settings: &PreviewPropertyInspector{
				IP:      "192.168.20.10",
				Input:   "invalid",
				MeIndex: "0",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog := &mockLogger{}
			app := &App{
				connectionManager:   connectionmanager.NewConnectionManager(mockLog),
				logger:              mockLog,
				previewSettingStore: setting.NewSettingStore[*previewPropertyInspector](),
				programSettingStore: setting.NewSettingStore[*programPropertyInspector](),
			}

			var err error
			switch tt.handlerType {
			case "preview":
				settings := tt.settings.(*PreviewPropertyInspector)
				payload := buildWillAppearPayload(settings)
				err = app.PRVWillAppearHandler(context.Background(), nil, payload)
			case "program":
				settings := tt.settings.(*ProgramPropertyInspector)
				payload := buildWillAppearPayload(settings)
				err = app.PGMWillAppearHandler(context.Background(), nil, payload)
			case "auto":
				settings := tt.settings.(*AutoPropertyInspector)
				payload := buildWillAppearPayload(settings)
				err = app.AutoWillAppearHandler(context.Background(), nil, payload)
			default:
				t.Fatalf("unknown handlerType: %s", tt.handlerType)
			}

			if tt.expectErr {
				asserts.Error(err)
			} else {
				asserts.NoError(err)
			}

			if tt.validate != nil {
				tt.validate(t, app, tt.contextID)
			}
		})
	}
}

func TestHandleWillDisappear_TableDriven(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		handlerType string // preview | program | auto
		contextID   string
		setup       func(*App, string)
		settings    interface{}
		expectErr   bool
	}{
		{
			name:        "preview: normal disappear after appear",
			handlerType: "preview",
			contextID:   "ctx-prev-disp-1",
			setup: func(app *App, ctxID string) {
				settings := &PreviewPropertyInspector{
					IP:      "192.168.20.10",
					Input:   "1",
					MeIndex: "0",
				}
				payload := buildWillAppearPayload(settings)
				_ = app.PRVWillAppearHandler(context.Background(), nil, payload)
			},
			settings:  &PreviewPropertyInspector{},
			expectErr: false,
		},
		{
			name:        "program: normal disappear after appear",
			handlerType: "program",
			contextID:   "ctx-prog-disp-1",
			setup: func(app *App, ctxID string) {
				settings := &ProgramPropertyInspector{
					IP:      "192.168.20.11",
					Input:   "2",
					MeIndex: "1",
				}
				payload := buildWillAppearPayload(settings)
				_ = app.PGMWillAppearHandler(context.Background(), nil, payload)
			},
			settings:  &ProgramPropertyInspector{},
			expectErr: false,
		},
		{
			name:        "auto: normal disappear after appear",
			handlerType: "auto",
			contextID:   "ctx-auto-disp-1",
			setup: func(app *App, ctxID string) {
				settings := &AutoPropertyInspector{
					IP: "192.168.20.12",
				}
				payload := buildWillAppearPayload(settings)
				_ = app.AutoWillAppearHandler(context.Background(), nil, payload)
			},
			settings:  &AutoPropertyInspector{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog := &mockLogger{}
			app := &App{
				connectionManager:   connectionmanager.NewConnectionManager(mockLog),
				logger:              mockLog,
				previewSettingStore: setting.NewSettingStore[*previewPropertyInspector](),
				programSettingStore: setting.NewSettingStore[*programPropertyInspector](),
			}

			tt.setup(app, tt.contextID)

			var err error
			switch tt.handlerType {
			case "preview":
				settings := tt.settings.(*PreviewPropertyInspector)
				payload := buildWillDisappearPayload(settings)
				err = app.PRVWillDisappearHandler(context.Background(), nil, payload)
			case "program":
				settings := tt.settings.(*ProgramPropertyInspector)
				payload := buildWillDisappearPayload(settings)
				err = app.PGMWillDisappearHandler(context.Background(), nil, payload)
			case "auto":
				settings := tt.settings.(*AutoPropertyInspector)
				payload := buildWillDisappearPayload(settings)
				err = app.AutoWillDisappearHandler(context.Background(), nil, payload)
			default:
				t.Fatalf("unknown handlerType: %s", tt.handlerType)
			}

			if tt.expectErr {
				asserts.Error(err)
			} else {
				asserts.NoError(err)
			}
		})
	}
}

func TestHandleKeyDown_TableDriven(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		handlerType string // preview | program | auto
		contextID   string
		setup       func(*App, string)
		settings    interface{}
		expectErr   bool
	}{
		{
			name:        "preview: normal keydown",
			handlerType: "preview",
			contextID:   "ctx-prev-key-1",
			setup: func(app *App, ctxID string) {
				settings := &PreviewPropertyInspector{
					IP:      "192.168.20.10",
					Input:   "1",
					MeIndex: "0",
				}
				payload := buildWillAppearPayload(settings)
				_ = app.PRVWillAppearHandler(context.Background(), nil, payload)
			},
			settings: &PreviewPropertyInspector{
				IP:      "192.168.20.10",
				Input:   "1",
				MeIndex: "0",
			},
			expectErr: false,
		},
		{
			name:        "program: normal keydown",
			handlerType: "program",
			contextID:   "ctx-prog-key-1",
			setup: func(app *App, ctxID string) {
				settings := &ProgramPropertyInspector{
					IP:      "192.168.20.11",
					Input:   "2",
					MeIndex: "1",
				}
				payload := buildWillAppearPayload(settings)
				_ = app.PGMWillAppearHandler(context.Background(), nil, payload)
			},
			settings: &ProgramPropertyInspector{
				IP:      "192.168.20.11",
				Input:   "2",
				MeIndex: "1",
			},
			expectErr: false,
		},
		{
			name:        "auto: normal keydown",
			handlerType: "auto",
			contextID:   "ctx-auto-key-1",
			setup: func(app *App, ctxID string) {
				settings := &AutoPropertyInspector{
					IP: "192.168.20.12",
				}
				payload := buildWillAppearPayload(settings)
				_ = app.AutoWillAppearHandler(context.Background(), nil, payload)
			},
			settings: &AutoPropertyInspector{
				IP: "192.168.20.12",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog := &mockLogger{}
			app := &App{
				connectionManager:   connectionmanager.NewConnectionManager(mockLog),
				logger:              mockLog,
				previewSettingStore: setting.NewSettingStore[*previewPropertyInspector](),
				programSettingStore: setting.NewSettingStore[*programPropertyInspector](),
			}

			tt.setup(app, tt.contextID)

			var err error
			switch tt.handlerType {
			case "preview":
				settings := tt.settings.(*PreviewPropertyInspector)
				payload := buildKeyDownPayload(settings)
				err = app.PRVKeyDownHandler(context.Background(), nil, payload)
			case "program":
				settings := tt.settings.(*ProgramPropertyInspector)
				payload := buildKeyDownPayload(settings)
				err = app.PGMKeyDownHandler(context.Background(), nil, payload)
			case "auto":
				settings := tt.settings.(*AutoPropertyInspector)
				payload := buildKeyDownPayload(settings)
				err = app.AutoKeyDownHandler(context.Background(), nil, payload)
			default:
				t.Fatalf("unknown handlerType: %s", tt.handlerType)
			}

			if tt.expectErr {
				asserts.Error(err)
			} else {
				asserts.NoError(err)
			}
		})
	}
}

func TestHandleDidReceiveSettings_TableDriven(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		handlerType string // preview | program | auto
		contextID   string
		settings    interface{}
		expectErr   bool
	}{
		{
			name:        "preview: normal settings",
			handlerType: "preview",
			contextID:   "ctx-prev-settings-1",
			settings: &PreviewPropertyInspector{
				IP:      "192.168.20.10",
				Input:   "1",
				MeIndex: "0",
			},
			expectErr: false,
		},
		{
			name:        "program: normal settings",
			handlerType: "program",
			contextID:   "ctx-prog-settings-1",
			settings: &ProgramPropertyInspector{
				IP:      "192.168.20.11",
				Input:   "2",
				MeIndex: "1",
			},
			expectErr: false,
		},
		{
			name:        "auto: normal settings",
			handlerType: "auto",
			contextID:   "ctx-auto-settings-1",
			settings: &AutoPropertyInspector{
				IP: "192.168.20.12",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog := &mockLogger{}
			app := &App{
				connectionManager:   connectionmanager.NewConnectionManager(mockLog),
				logger:              mockLog,
				previewSettingStore: setting.NewSettingStore[*previewPropertyInspector](),
				programSettingStore: setting.NewSettingStore[*programPropertyInspector](),
			}

			var err error
			switch tt.handlerType {
			case "preview":
				settings := tt.settings.(*PreviewPropertyInspector)
				payload := buildDidReceiveSettingsPayload(settings)
				err = app.PRVDidReceiveSettingsHandler(context.Background(), nil, payload)
			case "program":
				settings := tt.settings.(*ProgramPropertyInspector)
				payload := buildDidReceiveSettingsPayload(settings)
				err = app.PGMDidReceiveSettingsHandler(context.Background(), nil, payload)
			case "auto":
				settings := tt.settings.(*AutoPropertyInspector)
				payload := buildDidReceiveSettingsPayload(settings)
				err = app.AutoDidReceiveSettingsHandler(context.Background(), nil, payload)
			default:
				t.Fatalf("unknown handlerType: %s", tt.handlerType)
			}

			if tt.expectErr {
				asserts.Error(err)
			} else {
				asserts.NoError(err)
			}
		})
	}
}
