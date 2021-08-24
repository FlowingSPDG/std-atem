package stdatem

import (
	"context"

	"github.com/FlowingSPDG/go-atem"
	"github.com/FlowingSPDG/streamdeck"
)

// App Main Engine
type App struct {
	state state              // State management
	sd    *streamdeck.Client // StreamDeck Client
	atem  *atem.Atem         // ATEM Client
}

// Run Run Background Process.
func Run(ctx context.Context) error {
	return nil
}

// setup StreamDeck Client
func setupSD(client *streamdeck.Client) {

}

func (a *App) onAtemConnected() {

}

func (a *App) onAtemClosed() {

}
