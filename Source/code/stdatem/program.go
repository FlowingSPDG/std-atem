package stdatem

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

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

	// IPの整合性チェック
	if net.ParseIP(payload.Settings.IP) == nil {
		a.logger.Error(ctx, "IPが不正です")
		return xerrors.New("IPが不正です")
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
	a.handleDisappear(ctx, event.Context)
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

	oldSettings, ok := a.programSettingStore.Load(event.Context)
	if !ok {
		a.logger.Error(ctx, "PGMDidReceiveSettingsHandler 設定が見つかりません")
	}

	if oldSettings.IP != parsed.IP {
		a.logger.Debug(ctx, "PGMDidReceiveSettingsHandler IPが変更されました")

		// IPの整合性チェック
		if net.ParseIP(parsed.IP) == nil {
			a.logger.Error(ctx, "IPが不正です")
			return xerrors.New("IPが不正です")
		}

		// 新しいインスタンスを初期化
		if err := a.addATEMHost(ctx, setProgramAction, event.Context, parsed.IP, true); err != nil {
			return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
		}

		// 既存のコンテキストを削除し、新しいコンテキストを保存する
		// 既存のATEMインスタンスが他のコンテキストで使用されていない場合は削除する
		a.connectionManager.DeleteATEMByContext(ctx, event.Context)

		// 新しいATEMインスタンスを取得
		instance, ok := a.connectionManager.SolveATEMByIP(ctx, parsed.IP)
		if !ok {
			a.logger.Error(ctx, "PGMDidReceiveSettingsHandler 新しいATEMインスタンスが見つかりません")
			return xerrors.New("PGMDidReceiveSettingsHandler 新しいATEMインスタンスが見つかりません")
		}

		// 新しいコンテキストを保存
		a.connectionManager.Store(ctx, setProgramAction, parsed.IP, event.Context, instance)
		a.logger.Debug(ctx, "PGMDidReceiveSettingsHandler コンテキストを更新しました")
	}

	a.programSettingStore.Store(event.Context, parsed)

	return nil
}
