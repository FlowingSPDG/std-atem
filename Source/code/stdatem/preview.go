package stdatem

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	"golang.org/x/xerrors"
)

// PRVWillAppearHandler ATEM PRVを設定
func (a *App) PRVWillAppearHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.WillAppearPayload[*PreviewPropertyInspector]) error {
	handler := NewBaseEventHandler[*PreviewPropertyInspector](a)

	settingsParser := func(settings *PreviewPropertyInspector) (interface{}, error) {
		parsed, err := settings.Parse()
		if err != nil {
			return nil, err
		}
		contextID := sdcontext.Context(ctx)
		a.previewSettingStore.Store(contextID, parsed)
		return parsed, nil
	}

	return handler.HandleWillAppear(ctx, client, p, setPreviewAction, settingsParser)
}

// PRVWillDisappearHandler プレビューのボタン非表示を処理
func (a *App) PRVWillDisappearHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.WillDisappearPayload[*PreviewPropertyInspector]) error {
	handler := NewBaseEventHandler[*PreviewPropertyInspector](a)
	return handler.HandleWillDisappear(ctx, client, p)
}

// PRVKeyDownHandler ATEM PRVを設定
func (a *App) PRVKeyDownHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.KeyDownPayload[*PreviewPropertyInspector]) error {
	handler := NewBaseEventHandler[*PreviewPropertyInspector](a)

	settingsParser := func(settings *PreviewPropertyInspector) (interface{}, error) {
		return settings.Parse()
	}

	actionHandler := func(parsed interface{}) error {
		previewSetting, ok := parsed.(*previewPropertyInspector)
		if !ok {
			return xerrors.New("invalid settings type")
		}

		contextID := sdcontext.Context(ctx)
		instance, ok := a.connectionManager.SolveATEMByContext(ctx, contextID)
		if !ok {
			return xerrors.New("ATEM instance not found")
		}

		a.logger.Debug(ctx, "PRVKeyDownHandler input:%d meIndex:%d", previewSetting.Input, previewSetting.MeIndex)
		instance.Client.SetPreviewInput(previewSetting.Input, previewSetting.MeIndex)
		return nil
	}

	return handler.HandleKeyDown(ctx, client, p, setPreviewAction, settingsParser, actionHandler)
}

// PRVDidReceiveSettingsHandler PRVの設定を受け取る
func (a *App) PRVDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.DidReceiveSettingsPayload[*PreviewPropertyInspector]) error {
	handler := NewBaseEventHandler[*PreviewPropertyInspector](a)

	settingsParser := func(settings *PreviewPropertyInspector) (interface{}, error) {
		parsed, err := settings.Parse()
		if err != nil {
			return nil, err
		}
		contextID := sdcontext.Context(ctx)
		a.previewSettingStore.Store(contextID, parsed)
		return parsed, nil
	}

	return handler.HandleDidReceiveSettings(ctx, client, p, setPreviewAction, settingsParser)
}
