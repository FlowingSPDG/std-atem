package stdatem

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
	"golang.org/x/xerrors"
)

// CutWillAppearHandler ATEM Cutを設定
func (a *App) CutWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.WillAppearPayload[*AutoPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	msg := fmt.Sprintf("Cut %#v でWillAppear", payload.Settings)
	a.logger.Debug(ctx, msg)

	// 新しいインスタンスを初期化
	if err := a.addATEMHost(ctx, cutAction, event.Context, payload.Settings.IP, false); err != nil {
		return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
	}

	return nil
}

// CutWillDisappearHandler Cutのボタン非表示を処理
func (a *App) CutWillDisappearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.WillDisappearPayload[AutoPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}
	a.handleDisappear(ctx, payload.Settings.IP)
	return nil
}

// CutKeyDownHandler ATEM Cutを実行
func (a *App) CutKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.KeyDownPayload[AutoPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	msg := fmt.Sprintf("Cut %v でKeyDown", payload.Settings)
	a.logger.Debug(ctx, msg)

	instance, ok := a.connectionManager.SolveATEMByContext(ctx, event.Context)
	if !ok {
		a.logger.Error(ctx, "CutKeyDownHandler ATEMが見つかりません")
		return xerrors.New("CutKeyDownHandler ATEMが見つかりません")
	}

	a.logger.Debug(ctx, "CutKeyDownHandler")

	instance.Client.PerformCut()
	a.logger.Debug(ctx, "CutKeyDownHandler 完了")
	return nil
}

// CutDidReceiveSettingsHandler Cutの設定を受け取る
func (a *App) CutDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.DidReceiveSettingsPayload[AutoPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	// 新しいインスタンスを初期化
	if err := a.addATEMHost(ctx, cutAction, event.Context, payload.Settings.IP, true); err != nil {
		return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
	}

	return nil
}
