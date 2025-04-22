module github.com/FlowingSPDG/std-atem/Source/code

go 1.23.0

toolchain go1.23.1

require (
	github.com/FlowingSPDG/go-atem v0.0.0-20210521024700-964b2bac8248
	github.com/FlowingSPDG/streamdeck v0.0.0-20250312080211-6e0c0c0223d6
	github.com/puzpuzpuz/xsync v1.5.2
	github.com/puzpuzpuz/xsync/v3 v3.4.0
	github.com/samber/lo v1.49.1
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da
)

require (
	golang.org/x/text v0.21.0 // indirect
	nhooyr.io/websocket v1.8.17 // indirect
)

replace github.com/FlowingSPDG/streamdeck => ../../../streamdeck

replace github.com/FlowingSPDG/go-atem => ../../../go-atem
