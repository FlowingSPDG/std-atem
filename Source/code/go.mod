module github.com/FlowingSPDG/std-atem/Source/code

go 1.23.0

toolchain go1.23.1

require (
	github.com/FlowingSPDG/go-atem v0.0.0-20210521024700-964b2bac8248
	github.com/FlowingSPDG/streamdeck v0.0.0-20250312080211-6e0c0c0223d6
	github.com/puzpuzpuz/xsync v1.5.2
	github.com/puzpuzpuz/xsync/v3 v3.5.1
	github.com/stretchr/testify v1.10.0
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da
)

require (
	github.com/coder/websocket v1.8.13 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/FlowingSPDG/streamdeck => ../../../streamdeck

replace github.com/FlowingSPDG/go-atem => ../../../go-atem
