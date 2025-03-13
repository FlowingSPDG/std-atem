package main

import (
	"context"
	_ "embed"
	"io"
	"log"
	"os"

	"github.com/FlowingSPDG/std-atem/Source/code/di"
	"github.com/FlowingSPDG/std-atem/Source/code/stdatem"
)

const (
	// AppName Streamdeck plugin app name
	AppName = "dev.flowingspdg.atem.sdPlugin"
)

func main() {
	logfile, err := os.OpenFile("./streamdeck-atem-plugin.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open log:" + err.Error())
	}
	defer logfile.Close()
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime)

	ctx := context.Background()
	log.Println("Starting...")
	sd, err := di.InitializeStreamDeckClient(ctx)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	logger, err := di.InitializeStreamDeckLogger(ctx, sd)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	if err := stdatem.Run(ctx, logger, sd); err != nil {
		log.Fatalf("%v\n", err)
	}
}
