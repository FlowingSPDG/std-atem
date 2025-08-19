package stdatem

import (
	"context"
	"fmt"

	"github.com/FlowingSPDG/std-atem/Source/code/logger"
	"golang.org/x/xerrors"
)

// handlePayloadError 統一されたペイロードエラーハンドリング
func handlePayloadError(ctx context.Context, logger logger.Logger, operation string, err error) error {
	logger.Error(ctx, "%sの処理に失敗: %v", operation, err)
	return xerrors.Errorf("%sの処理に失敗: %w", operation, err)
}

// handleATEMError 統一されたATEMエラーハンドリング
func handleATEMError(ctx context.Context, logger logger.Logger, operation string, err error) error {
	logger.Error(ctx, "%sでATEMエラーが発生: %v", operation, err)
	return xerrors.Errorf("%sでATEMエラーが発生: %w", operation, err)
}

// handleConnectionError 統一された接続エラーハンドリング
func handleConnectionError(ctx context.Context, logger logger.Logger, operation string, err error) error {
	logger.Error(ctx, "%sで接続エラーが発生: %v", operation, err)
	return xerrors.Errorf("%sで接続エラーが発生: %w", operation, err)
}

// formatErrorMessage エラーメッセージの統一フォーマット
func formatErrorMessage(operation, details string) string {
	return fmt.Sprintf("%s: %s", operation, details)
}
