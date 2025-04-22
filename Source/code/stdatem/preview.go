package stdatem

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
	"golang.org/x/xerrors"
)

// PRVWillAppearHandler ATEM PRVを設定
func (a *App) PRVWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.WillAppearPayload[*PreviewPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}
	parsed, err := payload.Settings.Parse()
	if err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのパースに失敗: %v", err))
		return xerrors.Errorf("payloadのパースに失敗: %w", err)
	}

	msg := fmt.Sprintf("PRV %#v でWillAppear", parsed)
	a.logger.Debug(ctx, msg)

	a.previewSettingStore.Store(event.Context, parsed)

	// 新しいインスタンスを初期化
	if err := a.addATEMHost(ctx, setPreviewAction, event.Context, parsed.IP, false); err != nil {
		return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
	}

	return nil
}

// PRVWillDisappearHandler プレビューのボタン非表示を処理
func (a *App) PRVWillDisappearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.WillDisappearPayload[PreviewPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}
	a.handleDisappear(ctx, event.Context)
	return nil
}

// PRVKeyDownHandler ATEM PRVを設定
func (a *App) PRVKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.KeyDownPayload[PreviewPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	parsed, err := payload.Settings.Parse()
	if err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのパースに失敗: %v", err))
		return xerrors.Errorf("payloadのパースに失敗: %w", err)
	}

	msg := fmt.Sprintf("PRV %v でKeyDown", parsed)
	a.logger.Debug(ctx, msg)

	instance, ok := a.connectionManager.SolveATEMByContext(ctx, event.Context)
	if !ok {
		a.logger.Error(ctx, "PRVKeyDownHandler ATEMが見つかりません")
		return xerrors.New("PRVKeyDownHandler ATEMが見つかりません")
	}

	a.logger.Debug(ctx, "PRVKeyDownHandler input:%d meIndex:%d", parsed.Input, parsed.MeIndex)

	instance.Client.SetPreviewInput(parsed.Input, parsed.MeIndex)
	a.logger.Debug(ctx, "PRVKeyDownHandler 完了")
	return nil
}

// PRVDidReceiveSettingsHandler PRVの設定を受け取る
func (a *App) PRVDidReceiveSettingsHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.DidReceiveSettingsPayload[PreviewPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	parsed, err := payload.Settings.Parse()
	if err != nil {
		a.logger.Error(ctx, fmt.Sprintf("payloadのパースに失敗: %v", err))
		return xerrors.Errorf("payloadのパースに失敗: %w", err)
	}

	oldSettings, ok := a.previewSettingStore.Load(event.Context)
	if !ok {
		a.logger.Error(ctx, "PRVDidReceiveSettingsHandler 設定が見つかりません")
	}

	if oldSettings.IP != parsed.IP {
		a.logger.Debug(ctx, "PRVDidReceiveSettingsHandler IPが変更されました")
		// 新しいインスタンスを初期化
		if err := a.addATEMHost(ctx, setPreviewAction, event.Context, parsed.IP, true); err != nil {
			return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
		}

		// 既存のコンテキストを削除し、新しいコンテキストを保存する
		// 既存のATEMインスタンスが他のコンテキストで使用されていない場合は削除する
		a.connectionManager.DeleteATEMByContext(ctx, event.Context)

		// 新しいATEMインスタンスを取得
		instance, ok := a.connectionManager.SolveATEMByIP(ctx, parsed.IP)
		if !ok {
			a.logger.Error(ctx, "PRVDidReceiveSettingsHandler 新しいATEMインスタンスが見つかりません")
			return xerrors.New("PRVDidReceiveSettingsHandler 新しいATEMインスタンスが見つかりません")
		}

		// 新しいコンテキストを保存
		a.connectionManager.Store(ctx, setPreviewAction, parsed.IP, event.Context, instance)
		a.logger.Debug(ctx, "PRVDidReceiveSettingsHandler コンテキストを更新しました")
	}

	a.previewSettingStore.Store(event.Context, parsed)

	return nil
}
