package main

import (
	"context"
	_ "embed"
	"log"

	"github.com/FlowingSPDG/std-atem/Source/code/di"
	"github.com/FlowingSPDG/std-atem/Source/code/logger"
	"github.com/FlowingSPDG/std-atem/Source/code/stdatem"
)

const (
	// AppName Streamdeck plugin app name
	AppName = "dev.flowingspdg.atem.sdPlugin"
)

func main() {
	ctx := context.Background()
	sd, err := di.InitializeStreamDeckClient(ctx)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	logLevel := logger.DebugLevel | logger.InfoLevel | logger.WarnLevel | logger.ErrorLevel
	sdLogger, err := di.InitializeStreamDeckLogger(ctx, sd, logLevel)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	fileLogger := logger.NewFileLogger(ctx, logLevel)

	multiLogger := logger.NewMultiLogger(logLevel, fileLogger, sdLogger)

	if err := stdatem.Run(ctx, multiLogger, sd); err != nil {
		log.Fatalf("%v\n", err)
	}
}
