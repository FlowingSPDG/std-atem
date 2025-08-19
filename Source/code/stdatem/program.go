package stdatem

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	"golang.org/x/xerrors"
)

// PGMWillAppearHandler ATEM PGMを設定
func (a *App) PGMWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*ProgramPropertyInspector](a)

	settingsParser := func(settings *ProgramPropertyInspector) (interface{}, error) {
		parsed, err := settings.Parse()
		if err != nil {
			return nil, err
		}
		a.programSettingStore.Store(event.Context, parsed)
		return parsed, nil
	}

	return handler.HandleWillAppear(ctx, client, event, setProgramAction, settingsParser)
}

// PGMWillDisappearHandler プログラムのボタン非表示を処理
func (a *App) PGMWillDisappearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*ProgramPropertyInspector](a)
	return handler.HandleWillDisappear(ctx, client, event)
}

// PGMKeyDownHandler ATEM PGMを設定
func (a *App) PGMKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*ProgramPropertyInspector](a)

	settingsParser := func(settings *ProgramPropertyInspector) (interface{}, error) {
		return settings.Parse()
	}

	actionHandler := func(parsed interface{}) error {
		programSetting, ok := parsed.(*programPropertyInspector)
		if !ok {
			return xerrors.New("invalid settings type")
		}

		instance, ok := a.connectionManager.SolveATEMByContext(ctx, event.Context)
		if !ok {
			return xerrors.New("ATEM instance not found")
		}

		a.logger.Debug(ctx, "PGMKeyDownHandler input:%d meIndex:%d", programSetting.Input, programSetting.MeIndex)
		instance.Client.SetProgramInput(programSetting.Input, programSetting.MeIndex)
		return nil
	}

	return handler.HandleKeyDown(ctx, client, event, setProgramAction, settingsParser, actionHandler)
}

// PGMDidReceiveSettingsHandler PGMの設定を受け取る
func (a *App) PGMDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	handler := NewBaseEventHandler[*ProgramPropertyInspector](a)

	settingsParser := func(settings *ProgramPropertyInspector) (interface{}, error) {
		parsed, err := settings.Parse()
		if err != nil {
			return nil, err
		}
		a.programSettingStore.Store(event.Context, parsed)
		return parsed, nil
	}

	return handler.HandleDidReceiveSettings(ctx, client, event, setProgramAction, settingsParser)
}
