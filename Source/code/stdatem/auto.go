package stdatem

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	"golang.org/x/xerrors"
)

// AutoWillAppearHandler ATEM Autoを設定
func (a *App) AutoWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*AutoPropertyInspector](a)

	settingsParser := func(settings *AutoPropertyInspector) (interface{}, error) {
		return settings, nil
	}

	return handler.HandleWillAppear(ctx, client, event, autoAction, settingsParser)
}

// AutoWillDisappearHandler Autoのボタン非表示を処理
func (a *App) AutoWillDisappearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*AutoPropertyInspector](a)
	return handler.HandleWillDisappear(ctx, client, event)
}

// AutoKeyDownHandler ATEM Autoを実行
func (a *App) AutoKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*AutoPropertyInspector](a)

	settingsParser := func(settings *AutoPropertyInspector) (interface{}, error) {
		return settings, nil
	}

	actionHandler := func(parsed interface{}) error {
		instance, ok := a.connectionManager.SolveATEMByContext(ctx, event.Context)
		if !ok {
			return xerrors.New("ATEM instance not found")
		}

		a.logger.Debug(ctx, "AutoKeyDownHandler")
		instance.Client.PerformAutoTransition()
		return nil
	}

	return handler.HandleKeyDown(ctx, client, event, autoAction, settingsParser, actionHandler)
}

// AutoDidReceiveSettingsHandler Autoの設定を受け取る
func (a *App) AutoDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*AutoPropertyInspector](a)

	settingsParser := func(settings *AutoPropertyInspector) (interface{}, error) {
		return settings, nil
	}

	return handler.HandleDidReceiveSettings(ctx, client, event, autoAction, settingsParser)
}
