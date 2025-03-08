package stdatem

import "github.com/FlowingSPDG/go-atem"

// ATEMInstance represents a single ATEM connection
type ATEMInstance struct {
	client      *atem.Atem
	state       state
	reconnectCh chan struct{}
}

type PreviewPropertyInspector struct {
	IP    string
	Input uint16
}

type ProgramPropertyInspector struct {
	IP    string
	Input uint16
}
