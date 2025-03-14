package stdatem

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
	"golang.org/x/xerrors"
)

// PGMWillAppearHandler ATEM PGMを設定
func (a *App) PGMWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.WillAppearPayload[*ProgramPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}
	parsed, err := payload.Settings.Parse()
	if err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのパースに失敗: %v", err))
		return xerrors.Errorf("payloadのパースに失敗: %w", err)
	}

	msg := fmt.Sprintf("PGM %#v でWillAppear", parsed)
	a.logger.Debug(ctx, msg)

	a.programSettingStore.Store(event.Context, parsed)

	// 新しいインスタンスを初期化
	if err := a.addATEMHost(ctx, setProgramAction, event.Context, parsed.IP, false); err != nil {
		return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
	}

	return nil
}

// PGMWillDisappearHandler プログラムのボタン非表示を処理
func (a *App) PGMWillDisappearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.WillDisappearPayload[*ProgramPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}
	a.handleDisappear(ctx, payload.Settings.IP)
	return nil
}

// PGMKeyDownHandler ATEM PGMを設定
func (a *App) PGMKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.KeyDownPayload[*ProgramPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	parsed, err := payload.Settings.Parse()
	if err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのパースに失敗: %v", err))
		return xerrors.Errorf("payloadのパースに失敗: %w", err)
	}

	msg := fmt.Sprintf("PGM %v でKeyDown", parsed)
	a.logger.Debug(ctx, msg)

	instance, ok := a.connectionManager.SolveATEMByContext(ctx, event.Context)
	if !ok {
		a.logger.Error(ctx, "PGMKeyDownHandler ATEMが見つかりません")
		return xerrors.New("PGMKeyDownHandler ATEMが見つかりません")
	}

	a.logger.Debug(ctx, "PGMKeyDownHandler input:%d meIndex:%d", parsed.Input, parsed.MeIndex)

	instance.Client.SetProgramInput(parsed.Input, parsed.MeIndex)
	a.logger.Debug(ctx, "PGMKeyDownHandler 完了")
	return nil
}

// PGMDidReceiveSettingsHandler PGMの設定を受け取る
func (a *App) PGMDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.DidReceiveSettingsPayload[ProgramPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	parsed, err := payload.Settings.Parse()
	if err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのパースに失敗: %v", err))
		return xerrors.Errorf("payloadのパースに失敗: %w", err)
	}

	// 新しいインスタンスを初期化
	if err := a.addATEMHost(ctx, setProgramAction, event.Context, parsed.IP, true); err != nil {
		return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
	}

	a.programSettingStore.Store(event.Context, parsed)

	return nil
}
