package main

import (
	"context"
	_ "embed"
	"io"
	"log"
	"os"

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
	if err := stdatem.Run(ctx); err != nil {
		log.Fatalf("%v\n", err)
	}
}
