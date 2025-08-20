package stdatem

import (
	"strconv"
	"strings"

	"github.com/FlowingSPDG/go-atem"
	"golang.org/x/xerrors"
)

type PreviewPropertyInspector struct {
	IP      string `json:"ip"`
	Input   string `json:"input"`
	MeIndex string `json:"meIndex"`
}

func (p *PreviewPropertyInspector) Parse() (*previewPropertyInspector, error) {
	ip := strings.TrimSpace(p.IP)

	inStr := strings.TrimSpace(p.Input)
	if inStr == "" {
		inStr = "1" // default to input 1
	}
	input64, err := strconv.ParseInt(inStr, 10, 64)
	if err != nil {
		return nil, xerrors.Errorf("inputの解析に失敗: %w", err)
	}

	meStr := strings.TrimSpace(p.MeIndex)
	if meStr == "" {
		meStr = "0" // default to ME 0
	}
	meIndex, err := strconv.ParseInt(meStr, 10, 64)
	if err != nil {
		return nil, xerrors.Errorf("meIndexの解析に失敗: %w", err)
	}

	return &previewPropertyInspector{
		IP:      ip,
		Input:   solveATEMVideoInput(input64),
		MeIndex: uint8(meIndex),
	}, nil
}

type previewPropertyInspector struct {
	IP      string
	Input   atem.VideoInputType
	MeIndex uint8
}

type ProgramPropertyInspector struct {
	IP      string `json:"ip"`
	Input   string `json:"input"`
	MeIndex string `json:"meIndex"`
}

type programPropertyInspector struct {
	IP      string
	Input   atem.VideoInputType
	MeIndex uint8
}

func (p *ProgramPropertyInspector) Parse() (*programPropertyInspector, error) {
	ip := strings.TrimSpace(p.IP)

	inStr := strings.TrimSpace(p.Input)
	if inStr == "" {
		inStr = "1" // default to input 1
	}
	input64, err := strconv.ParseInt(inStr, 10, 64)
	if err != nil {
		return nil, xerrors.Errorf("inputの解析に失敗: %w", err)
	}

	meStr := strings.TrimSpace(p.MeIndex)
	if meStr == "" {
		meStr = "0" // default to ME 0
	}
	meIndex, err := strconv.ParseInt(meStr, 10, 64)
	if err != nil {
		return nil, xerrors.Errorf("meIndexの解析に失敗: %w", err)
	}

	return &programPropertyInspector{
		IP:      ip,
		Input:   solveATEMVideoInput(input64),
		MeIndex: uint8(meIndex),
	}, nil
}

type AutoPropertyInspector struct {
	IP string `json:"ip"`
}
