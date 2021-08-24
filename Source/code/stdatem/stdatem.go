package stdatem

import (
	"context"
	"sync"

	"github.com/FlowingSPDG/go-atem"
	"github.com/FlowingSPDG/streamdeck"
)

// App Main Engine
type App struct {
	state   state              // State management
	buttons buttons            // Setting per context(button)
	sd      *streamdeck.Client // StreamDeck Client
	atem    *atem.Atem         // ATEM Client
}

// NewApp Initiate App main engine
func NewApp(ctx context.Context, cfg Config) (*App, error) {
	app := &App{
		buttons: buttons{
			m: &sync.Map{},
		},
	}

	// Setup SD
	params, err := streamdeck.ParseRegistrationParams(cfg.Params)
	if err != nil {
		return nil, err
	}
	app.sd = streamdeck.NewClient(ctx, params)
	app.setupSD()

	// Setup ATEM
	app.atem = atem.Create(cfg.ATEMip, cfg.Debug)
	app.setupATEM()

	return app, nil
}

// Run Run Background Process.
func (a *App) Run() error {
	return a.sd.Run()
}

// ConnectATEM Connect to ATEM.
func (a *App) ConnectATEM() error {
	return a.atem.Connect()
}

// setup StreamDeck Client
func (a *App) setupSD() {
	prv := a.sd.Action(SetPreviewAction)
	prv.RegisterHandler(streamdeck.KeyDown, a.PRVKeyDownHandler)

	pgm := a.sd.Action(SetProgramAction)
	pgm.RegisterHandler(streamdeck.KeyDown, a.PGMKeyDownHandler)
}

func (a *App) setupATEM() {
	a.atem.On("connected", a.onAtemConnected)
	a.atem.On("closed", a.onAtemClosed)
}

func (a *App) onAtemConnected() {
	a.state = state{
		Preview:   a.atem.PreviewInput.Index,
		Program:   a.atem.ProgramInput.Index,
		Connected: a.atem.Connected(),
	}
}

func (a *App) onAtemClosed() {
	a.state.Connected = a.atem.Connected()
}

// PRVKeyDownHandler Set ATEM PRV
func (a *App) PRVKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	s, err := a.buttons.Load(event.Context)
	if err != nil {
		return err
	}

	a.atem.SetPreviewInput(atem.VideoInputType(s.input), uint8(s.input))
	return nil
}

// PGMKeyDownHandler Set ATEM PGM
func (a *App) PGMKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	s, err := a.buttons.Load(event.Context)
	if err != nil {
		return err
	}

	a.atem.SetProgramInput(atem.VideoInputType(s.input), uint8(s.input))
	return nil
}
