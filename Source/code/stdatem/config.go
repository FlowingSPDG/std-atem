package stdatem

import (
	"encoding/json"

	"github.com/FlowingSPDG/go-atem"
)

// ATEMInstance represents a single ATEM connection
type ATEMInstance struct {
	client      *atem.Atem
	state       state
	reconnectCh chan struct{}
}

type PreviewPropertyInspector struct {
	IP      string      `json:"ip"`
	Input   json.Number `json:"input"`
	MeIndex json.Number `json:"meIndex"`
}

type ProgramPropertyInspector struct {
	IP      string      `json:"ip"`
	Input   json.Number `json:"input"`
	MeIndex json.Number `json:"meIndex"`
}
