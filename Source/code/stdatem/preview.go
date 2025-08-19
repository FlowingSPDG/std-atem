package stdatem

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	"golang.org/x/xerrors"
)

// PRVWillAppearHandler ATEM PRVを設定
func (a *App) PRVWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*PreviewPropertyInspector](a)

	settingsParser := func(settings *PreviewPropertyInspector) (interface{}, error) {
		parsed, err := settings.Parse()
		if err != nil {
			return nil, err
		}
		a.previewSettingStore.Store(event.Context, parsed)
		return parsed, nil
	}

	return handler.HandleWillAppear(ctx, client, event, setPreviewAction, settingsParser)
}

// PRVWillDisappearHandler プレビューのボタン非表示を処理
func (a *App) PRVWillDisappearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*PreviewPropertyInspector](a)
	return handler.HandleWillDisappear(ctx, client, event)
}

// PRVKeyDownHandler ATEM PRVを設定
func (a *App) PRVKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*PreviewPropertyInspector](a)

	settingsParser := func(settings *PreviewPropertyInspector) (interface{}, error) {
		return settings.Parse()
	}

	actionHandler := func(parsed interface{}) error {
		previewSetting, ok := parsed.(*previewPropertyInspector)
		if !ok {
			return xerrors.New("invalid settings type")
		}

		instance, ok := a.connectionManager.SolveATEMByContext(ctx, event.Context)
		if !ok {
			return xerrors.New("ATEM instance not found")
		}

		a.logger.Debug(ctx, "PRVKeyDownHandler input:%d meIndex:%d", previewSetting.Input, previewSetting.MeIndex)
		instance.Client.SetPreviewInput(previewSetting.Input, previewSetting.MeIndex)
		return nil
	}

	return handler.HandleKeyDown(ctx, client, event, setPreviewAction, settingsParser, actionHandler)
}

// PRVDidReceiveSettingsHandler PRVの設定を受け取る
func (a *App) PRVDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*PreviewPropertyInspector](a)

	settingsParser := func(settings *PreviewPropertyInspector) (interface{}, error) {
		parsed, err := settings.Parse()
		if err != nil {
			return nil, err
		}
		a.previewSettingStore.Store(event.Context, parsed)
		return parsed, nil
	}

	return handler.HandleDidReceiveSettings(ctx, client, event, setPreviewAction, settingsParser)
}
