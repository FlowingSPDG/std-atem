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
	a.handleDisappear(ctx, payload.Settings.IP)
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

	instance, ok := a.atems.SolveATEMByContext(ctx, event.Context)
	if !ok {
		a.logger.Error(ctx, "PRVKeyDownHandler ATEMが見つかりません")
		return xerrors.New("PRVKeyDownHandler ATEMが見つかりません")
	}

	a.logger.Debug(ctx, "PRVKeyDownHandler input:%d meIndex:%d", parsed.Input, parsed.MeIndex)

	instance.client.SetPreviewInput(parsed.Input, parsed.MeIndex)
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

	// 新しいインスタンスを初期化
	if err := a.addATEMHost(ctx, setPreviewAction, event.Context, parsed.IP, true); err != nil {
		return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
	}

	a.previewSettingStore.Store(event.Context, parsed)

	return nil
}
