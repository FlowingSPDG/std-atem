package stdatem

import (
	"encoding/json"

	"github.com/FlowingSPDG/go-atem"
	"golang.org/x/xerrors"
)

type PreviewPropertyInspector struct {
	IP      string      `json:"ip"`
	Input   json.Number `json:"input"`
	MeIndex json.Number `json:"meIndex"`
}

func (p *PreviewPropertyInspector) Parse() (*previewPropertyInspector, error) {
	ip := p.IP
	input, err := p.Input.Int64()
	if err != nil {
		return nil, xerrors.Errorf("inputの解析に失敗: %w", err)
	}
	meIndex, err := p.MeIndex.Int64()
	if err != nil {
		return nil, xerrors.Errorf("meIndexの解析に失敗: %w", err)
	}

	return &previewPropertyInspector{
		IP:      ip,
		Input:   solveATEMVideoInput(input),
		MeIndex: uint8(meIndex),
	}, nil
}

type previewPropertyInspector struct {
	IP      string
	Input   atem.VideoInputType
	MeIndex uint8
}

type ProgramPropertyInspector struct {
	IP      string      `json:"ip"`
	Input   json.Number `json:"input"`
	MeIndex json.Number `json:"meIndex"`
}

type programPropertyInspector struct {
	IP      string
	Input   atem.VideoInputType
	MeIndex uint8
}

func (p *ProgramPropertyInspector) Parse() (*programPropertyInspector, error) {
	ip := p.IP
	input, err := p.Input.Int64()
	if err != nil {
		return nil, xerrors.Errorf("inputの解析に失敗: %w", err)
	}
	meIndex, err := p.MeIndex.Int64()
	if err != nil {
		return nil, xerrors.Errorf("meIndexの解析に失敗: %w", err)
	}

	return &programPropertyInspector{
		IP:      ip,
		Input:   solveATEMVideoInput(input),
		MeIndex: uint8(meIndex),
	}, nil
}

type AutoPropertyInspector struct {
	IP string `json:"ip"`
}
