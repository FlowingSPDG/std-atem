package stdatem

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
	"golang.org/x/xerrors"
)

// AutoWillAppearHandler ATEM Autoを設定
func (a *App) AutoWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.WillAppearPayload[*AutoPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	msg := fmt.Sprintf("Auto %#v でWillAppear", payload.Settings)
	a.logger.Debug(ctx, msg)

	// 新しいインスタンスを初期化
	if err := a.addATEMHost(ctx, autoAction, event.Context, payload.Settings.IP, false); err != nil {
		return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
	}

	return nil
}

// AutoWillDisappearHandler Autoのボタン非表示を処理
func (a *App) AutoWillDisappearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.WillDisappearPayload[AutoPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}
	a.handleDisappear(ctx, event.Context)
	return nil
}

// AutoKeyDownHandler ATEM Autoを実行
func (a *App) AutoKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.KeyDownPayload[AutoPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	msg := fmt.Sprintf("Auto %v でKeyDown", payload.Settings)
	a.logger.Debug(ctx, msg)

	instance, ok := a.connectionManager.SolveATEMByContext(ctx, event.Context)
	if !ok {
		a.logger.Error(ctx, "AutoKeyDownHandler ATEMが見つかりません")
		return xerrors.New("AutoKeyDownHandler ATEMが見つかりません")
	}

	a.logger.Debug(ctx, "AutoKeyDownHandler")

	instance.Client.PerformAutoTransition()
	a.logger.Debug(ctx, "AutoKeyDownHandler 完了")
	return nil
}

// AutoDidReceiveSettingsHandler Autoの設定を受け取る
func (a *App) AutoDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.DidReceiveSettingsPayload[AutoPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	// Handle IP change if this context was using a different IP
	a.connectionManager.UpdateContextIP(ctx, event.Context, payload.Settings.IP)

	// 新しいインスタンスを初期化
	if err := a.addATEMHost(ctx, autoAction, event.Context, payload.Settings.IP, true); err != nil {
		return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
	}

	return nil
}
