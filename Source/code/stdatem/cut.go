package stdatem

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	"golang.org/x/xerrors"
)

// CutWillAppearHandler ATEM Cutを設定
func (a *App) CutWillAppearHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.WillAppearPayload[*AutoPropertyInspector]) error {
	handler := NewBaseEventHandler[*AutoPropertyInspector](a)

	settingsParser := func(settings *AutoPropertyInspector) (interface{}, error) {
		return settings, nil
	}

	return handler.HandleWillAppear(ctx, client, p, cutAction, settingsParser)
}

// CutWillDisappearHandler Cutのボタン非表示を処理
func (a *App) CutWillDisappearHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.WillDisappearPayload[*AutoPropertyInspector]) error {
	handler := NewBaseEventHandler[*AutoPropertyInspector](a)
	return handler.HandleWillDisappear(ctx, client, p)
}

// CutKeyDownHandler ATEM Cutを実行
func (a *App) CutKeyDownHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.KeyDownPayload[*AutoPropertyInspector]) error {
	handler := NewBaseEventHandler[*AutoPropertyInspector](a)

	settingsParser := func(settings *AutoPropertyInspector) (interface{}, error) {
		return settings, nil
	}

	actionHandler := func(parsed interface{}) error {
		contextID := sdcontext.Context(ctx)
		instance, ok := a.connectionManager.SolveATEMByContext(ctx, contextID)
		if !ok {
			return xerrors.New("ATEM instance not found")
		}

		a.logger.Debug(ctx, "CutKeyDownHandler")
		instance.Client.PerformCut()
		return nil
	}

	return handler.HandleKeyDown(ctx, client, p, cutAction, settingsParser, actionHandler)
}

// CutDidReceiveSettingsHandler Cutの設定を受け取る
func (a *App) CutDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.DidReceiveSettingsPayload[*AutoPropertyInspector]) error {
	handler := NewBaseEventHandler[*AutoPropertyInspector](a)

	settingsParser := func(settings *AutoPropertyInspector) (interface{}, error) {
		return settings, nil
	}

	return handler.HandleDidReceiveSettings(ctx, client, p, cutAction, settingsParser)
}
