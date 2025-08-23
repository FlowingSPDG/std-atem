package stdatem

import (
	"context"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
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
func (h *BaseEventHandler[T]) HandleWillAppear(ctx context.Context, client *streamdeck.Client, p streamdeck.WillAppearPayload[T], action string, settingsParser func(T) (interface{}, error)) error {
	parsed, err := settingsParser(p.Settings)
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

	contextID := sdcontext.Context(ctx)
	return h.app.addATEMHost(ctx, action, contextID, ip, false)
}

// HandleWillDisappear 共通のWillDisappear処理
func (h *BaseEventHandler[T]) HandleWillDisappear(ctx context.Context, client *streamdeck.Client, p streamdeck.WillDisappearPayload[T]) error {
	contextID := sdcontext.Context(ctx)
	h.app.handleDisappear(ctx, contextID)
	return nil
}

// HandleKeyDown 共通のKeyDown処理
func (h *BaseEventHandler[T]) HandleKeyDown(ctx context.Context, client *streamdeck.Client, p streamdeck.KeyDownPayload[T], action string, settingsParser func(T) (interface{}, error), actionHandler func(interface{}) error) error {
	parsed, err := settingsParser(p.Settings)
	if err != nil {
		h.app.logger.Error(ctx, fmt.Sprintf("payloadのパースに失敗: %v", err))
		return xerrors.Errorf("payloadのパースに失敗: %w", err)
	}

	msg := fmt.Sprintf("%s %v でKeyDown", action, parsed)
	h.app.logger.Debug(ctx, msg)

	contextID := sdcontext.Context(ctx)
	_, ok := h.app.connectionManager.SolveATEMByContext(ctx, contextID)
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
func (h *BaseEventHandler[T]) HandleDidReceiveSettings(ctx context.Context, client *streamdeck.Client, p streamdeck.DidReceiveSettingsPayload[T], action string, settingsParser func(T) (interface{}, error)) error {
	parsed, err := settingsParser(p.Settings)
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

	contextID := sdcontext.Context(ctx)
	// Handle IP change if this context was using a different IP
	h.app.connectionManager.UpdateContextIP(ctx, contextID, ip)

	// 新しいインスタンスを初期化
	if err := h.app.addATEMHost(ctx, action, contextID, ip, true); err != nil {
		return xerrors.Errorf("ATEMホストの追加に失敗: %w", err)
	}

	return nil
}
