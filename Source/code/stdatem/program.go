package stdatem

import (
	"context"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	"golang.org/x/xerrors"
)

// PGMWillAppearHandler ATEM PGMを設定
func (a *App) PGMWillAppearHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.WillAppearPayload[*ProgramPropertyInspector]) error {
	handler := NewBaseEventHandler[*ProgramPropertyInspector](a)

	settingsParser := func(settings *ProgramPropertyInspector) (interface{}, error) {
		parsed, err := settings.Parse()
		if err != nil {
			return nil, err
		}
		contextID := sdcontext.Context(ctx)
		a.programSettingStore.Store(contextID, parsed)
		return parsed, nil
	}

	return handler.HandleWillAppear(ctx, client, p, setProgramAction, settingsParser)
}

// PGMWillDisappearHandler プログラムのボタン非表示を処理
func (a *App) PGMWillDisappearHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.WillDisappearPayload[*ProgramPropertyInspector]) error {
	handler := NewBaseEventHandler[*ProgramPropertyInspector](a)
	return handler.HandleWillDisappear(ctx, client, p)
}

// PGMKeyDownHandler ATEM PGMを設定
func (a *App) PGMKeyDownHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.KeyDownPayload[*ProgramPropertyInspector]) error {
	handler := NewBaseEventHandler[*ProgramPropertyInspector](a)

	settingsParser := func(settings *ProgramPropertyInspector) (interface{}, error) {
		return settings.Parse()
	}

	actionHandler := func(parsed interface{}) error {
		programSetting, ok := parsed.(*programPropertyInspector)
		if !ok {
			return xerrors.New("invalid settings type")
		}

		contextID := sdcontext.Context(ctx)
		instance, ok := a.connectionManager.SolveATEMByContext(ctx, contextID)
		if !ok {
			return xerrors.New("ATEM instance not found")
		}

		a.logger.Debug(ctx, "PGMKeyDownHandler input:%d meIndex:%d", programSetting.Input, programSetting.MeIndex)
		instance.Client.SetProgramInput(programSetting.Input, programSetting.MeIndex)
		return nil
	}

	return handler.HandleKeyDown(ctx, client, p, setProgramAction, settingsParser, actionHandler)
}

// PGMDidReceiveSettingsHandler PGMの設定を受け取る
func (a *App) PGMDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, p streamdeck.DidReceiveSettingsPayload[*ProgramPropertyInspector]) error {
	handler := NewBaseEventHandler[*ProgramPropertyInspector](a)

	settingsParser := func(settings *ProgramPropertyInspector) (interface{}, error) {
		parsed, err := settings.Parse()
		if err != nil {
			return nil, err
		}
		contextID := sdcontext.Context(ctx)
		a.programSettingStore.Store(contextID, parsed)
		return parsed, nil
	}

	return handler.HandleDidReceiveSettings(ctx, client, p, setProgramAction, settingsParser)
}
