package di

import (
	"context"
	"os"

	"github.com/FlowingSPDG/std-atem/Source/code/logger"

	"github.com/FlowingSPDG/streamdeck"
	"golang.org/x/xerrors"
)

func InitializeStreamDeckClient(ctx context.Context) (*streamdeck.Client, error) {
	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		return nil, xerrors.Errorf("registration paramsの解析に失敗: %w", err)
	}
	sd := streamdeck.NewClient(ctx, params)
	return sd, nil
}

func InitializeStreamDeckLogger(ctx context.Context, sd *streamdeck.Client) (logger.Logger, error) {
	logger := logger.NewStreamDeckLogger(sd, logger.InfoLevel)
	return logger, nil
}
