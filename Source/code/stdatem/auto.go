package stdatem

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	"golang.org/x/xerrors"
)

// AutoWillAppearHandler ATEM Autoを設定
func (a *App) AutoWillAppearHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.WillAppearPayload[*AutoPropertyInspector]) error {
	handler := NewBaseEventHandler[*AutoPropertyInspector](a)

	settingsParser := func(settings *AutoPropertyInspector) (interface{}, error) {
		return settings, nil
	}

	return handler.HandleWillAppear(ctx, client, p, autoAction, settingsParser)
}

// AutoWillDisappearHandler Autoのボタン非表示を処理
func (a *App) AutoWillDisappearHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.WillDisappearPayload[*AutoPropertyInspector]) error {
	handler := NewBaseEventHandler[*AutoPropertyInspector](a)
	return handler.HandleWillDisappear(ctx, client, p)
}

// AutoKeyDownHandler ATEM Autoを実行
func (a *App) AutoKeyDownHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.KeyDownPayload[*AutoPropertyInspector]) error {
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

		a.logger.Debug(ctx, "AutoKeyDownHandler")
		instance.Client.PerformAutoTransition()
		return nil
	}

	return handler.HandleKeyDown(ctx, client, p, autoAction, settingsParser, actionHandler)
}

// AutoDidReceiveSettingsHandler Autoの設定を受け取る
func (a *App) AutoDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.DidReceiveSettingsPayload[*AutoPropertyInspector]) error {
	handler := NewBaseEventHandler[*AutoPropertyInspector](a)

	settingsParser := func(settings *AutoPropertyInspector) (interface{}, error) {
		return settings, nil
	}

	return handler.HandleDidReceiveSettings(ctx, client, p, autoAction, settingsParser)
}
