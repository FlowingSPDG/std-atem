package stdatem

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
	"golang.org/x/xerrors"
)

// BaseEventHandler 共通のイベントハンドラーベース
type BaseEventHandler[T any] struct {
	app *App
}

// NewBaseEventHandler 新しいベースイベントハンドラーを作成
func NewBaseEventHandler[T any](app *App) *BaseEventHandler[T] {
	return &BaseEventHandler[T]{app: app}
}

// HandleWillAppear 共通のWillAppear処理
func (h *BaseEventHandler[T]) HandleWillAppear(ctx context.Context, client *streamdeck.Client, event streamdeck.Event, action string, settingsParser func(T) (interface{}, error)) error {
	var payload streamdeck.WillAppearPayload[T]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		h.app.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	parsed, err := settingsParser(payload.Settings)
	if err != nil {
		h.app.logger.Error(ctx, fmt.Sprintf("payloadのパースに失敗: %v", err))
		return xerrors.Errorf("payloadのパースに失敗: %w", err)
	}

	// IPアドレスを取得（型アサーションで対応）
	var ip string
	switch s := parsed.(type) {
	case *previewPropertyInspector:
		ip = s.IP
	case *programPropertyInspector:
		ip = s.IP
	case *AutoPropertyInspector:
		ip = s.IP
	default:
		return xerrors.New("unsupported settings type")
	}

	msg := fmt.Sprintf("%s %#v でWillAppear", action, parsed)
	h.app.logger.Debug(ctx, msg)

	return h.app.addATEMHost(ctx, action, event.Context, ip, false)
}

// HandleWillDisappear 共通のWillDisappear処理
func (h *BaseEventHandler[T]) HandleWillDisappear(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.WillDisappearPayload[T]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		h.app.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}
	h.app.handleDisappear(ctx, event.Context)
	return nil
}

// HandleKeyDown 共通のKeyDown処理
func (h *BaseEventHandler[T]) HandleKeyDown(ctx context.Context, client *streamdeck.Client, event streamdeck.Event, action string, settingsParser func(T) (interface{}, error), actionHandler func(interface{}) error) error {
	var payload streamdeck.KeyDownPayload[T]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		h.app.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	parsed, err := settingsParser(payload.Settings)
	if err != nil {
		h.app.logger.Error(ctx, fmt.Sprintf("payloadのパースに失敗: %v", err))
		return xerrors.Errorf("payloadのパースに失敗: %w", err)
	}

	msg := fmt.Sprintf("%s %v でKeyDown", action, parsed)
	h.app.logger.Debug(ctx, msg)

	_, ok := h.app.connectionManager.SolveATEMByContext(ctx, event.Context)
	if !ok {
		h.app.logger.Error(ctx, "%s ATEMが見つかりません", action)
		return xerrors.Errorf("%s ATEMが見つかりません", action)
	}

	if err := actionHandler(parsed); err != nil {
		return xerrors.Errorf("%s アクション実行に失敗: %w", action, err)
	}

	h.app.logger.Debug(ctx, "%s 完了", action)
	return nil
}

// HandleDidReceiveSettings 共通のDidReceiveSettings処理
func (h *BaseEventHandler[T]) HandleDidReceiveSettings(ctx context.Context, client *streamdeck.Client, event streamdeck.Event, action string, settingsParser func(T) (interface{}, error)) error {
	var payload streamdeck.DidReceiveSettingsPayload[T]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		h.app.logger.Error(ctx, fmt.Sprintf("payloadのアンマーシャルに失敗: %v", err))
		return xerrors.Errorf("payloadのアンマーシャルに失敗: %w", err)
	}

	parsed, err := settingsParser(payload.Settings)
	if err != nil {
		h.app.logger.Error(ctx, fmt.Sprintf("payloadのパースに失敗: %v", err))
		return xerrors.Errorf("payloadのパースに失敗: %w", err)
	}

	// IPアドレスを取得
	var ip string
	switch s := parsed.(type) {
	case *previewPropertyInspector:
		ip = s.IP
	case *programPropertyInspector:
		ip = s.IP
	case *AutoPropertyInspector:
		ip = s.IP
	default:
		return xerrors.New("unsupported settings type")
	}

	// Handle IP change if this context was using a different IP
	h.app.connectionManager.UpdateContextIP(ctx, event.Context, ip)

	// 新しいインスタンスを初期化
	if err := h.app.addATEMHost(ctx, action, event.Context, ip, true); err != nil {
		return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
	}

	return nil
}
