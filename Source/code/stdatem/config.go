package stdatem

import "github.com/FlowingSPDG/go-atem"

// ATEMInstance represents a single ATEM connection
type ATEMInstance struct {
	client      *atem.Atem
	state       state
	reconnectCh chan struct{}
}

type PreviewPropertyInspector struct {
	ATEMInstance
	Input uint16
}

type ProgramPropertyInspector struct {
	ATEMInstance
	Input uint16
}
