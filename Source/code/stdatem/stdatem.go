package stdatem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/FlowingSPDG/go-atem"
	"github.com/FlowingSPDG/streamdeck"
	"golang.org/x/xerrors"
)

// App Main Engine
type App struct {
	atems *atems             // Setting per context(button)
	sd    *streamdeck.Client // StreamDeck Client
}

// NewApp Initiate App main engine
func NewApp(ctx context.Context) (*App, error) {
	app := &App{
		atems: newAtems(),
	}

	// Setup SD
	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse registration params: %w", err)
	}
	app.sd = streamdeck.NewClient(ctx, params)
	app.setupSD()

	return app, nil
}

// addATEMHost adds a new ATEM host and sets up its connection
func (a *App) addATEMHost(ctx context.Context, sdcontext string, host *ATEMInstance, debug bool) error {
	msg := fmt.Sprintf("Adding ATEM host %s...", host.client.Ip)
	a.sd.LogMessage(ctx, msg)

	instance := &ATEMInstance{
		client:      atem.Create(host.client.Ip, debug),
		reconnectCh: make(chan struct{}, 1),
	}

	a.atems.Store(host.client.Ip, sdcontext, instance)

	instance.client.On("connected", func() {
		if instance, ok := a.atems.SolveATEMByIP(host.client.Ip); ok {
			instance.state = state{
				Preview:   instance.client.PreviewInput.Index,
				Program:   instance.client.ProgramInput.Index,
				Connected: instance.client.Connected(),
			}
		}
	})

	instance.client.On("closed", func() {
		if instance, ok := a.atems.SolveATEMByIP(host.client.Ip); ok {
			instance.state.Connected = instance.client.Connected()

			// Trigger reconnection
			select {
			case instance.reconnectCh <- struct{}{}:
			case <-ctx.Done():
				return
			default:
			}
		}
	})

	// Start reconnection goroutine
	go a.reconnectionLoop(ctx, host.client.Ip)

	return nil
}

// Run Run Background Process.
func (a *App) Run(ctx context.Context) error {
	return a.sd.Run(ctx)
}

// setup StreamDeck Client
func (a *App) setupSD() {
	prv := a.sd.Action(setPreviewAction)
	prv.RegisterHandler(streamdeck.KeyDown, a.PRVKeyDownHandler)
	prv.RegisterHandler(streamdeck.WillAppear, a.PRVWillAppearHandler)
	prv.RegisterHandler(streamdeck.WillDisappear, nil)

	pgm := a.sd.Action(setProgramAction)
	pgm.RegisterHandler(streamdeck.KeyDown, a.PGMKeyDownHandler)
	pgm.RegisterHandler(streamdeck.WillAppear, a.PGMWillAppearHandler)
	pgm.RegisterHandler(streamdeck.WillDisappear, nil)
}

// reconnectionLoop handles automatic reconnection for a specific ATEM host
func (a *App) reconnectionLoop(ctx context.Context, ip string) {
	instance, ok := a.atems.SolveATEMByIP(ip)
	if !ok {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-instance.reconnectCh:
			if err := instance.client.Connect(); err != nil {
				// Wait before retrying
				time.Sleep(5 * time.Second)
				// Try again
				select {
				case instance.reconnectCh <- struct{}{}:
				default:
				}
			}
		}
	}
}

// PRVKeyDownHandler Set ATEM PRV
func (a *App) PRVKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.KeyDownPayload[PreviewPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return xerrors.Errorf("failed to unmarshal payload: %w", err)
	}

	msg := fmt.Sprintf("KeyDown on PRV %v", payload.Settings)
	a.sd.LogMessage(ctx, msg)

	instance, ok := a.atems.SolveATEMByContext(event.Context)
	if !ok {
		return xerrors.New("ATEM not found")
	}

	instance.client.SetPreviewInput(atem.VideoInputType(payload.Settings.Input), uint8(payload.Settings.Input))
	return nil
}

// PRVWillAppearHandler Set ATEM PRV
func (a *App) PRVWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.WillAppearPayload[PreviewPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return xerrors.Errorf("failed to unmarshal payload: %w", err)
	}

	msg := fmt.Sprintf("WillAppear on PRV %v", payload.Settings)
	a.sd.LogMessage(ctx, msg)

	if _, ok := a.atems.SolveATEMByContext(event.Context); !ok {
		// initialize new instance
		if err := a.addATEMHost(ctx, event.Context, &ATEMInstance{
			client: atem.Create(payload.Settings.IP, true),
		}, true); err != nil {
			return xerrors.Errorf("failed to add ATEM host: %w", err)
		}
	}

	return nil
}

// PGMKeyDownHandler Set ATEM PGM
func (a *App) PGMKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.KeyDownPayload[ProgramPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return xerrors.Errorf("failed to unmarshal payload: %w", err)
	}

	msg := fmt.Sprintf("KeyDown on PGM %v", payload.Settings)
	a.sd.LogMessage(ctx, msg)

	instance, ok := a.atems.SolveATEMByContext(event.Context)
	if !ok {
		return xerrors.New("ATEM not found")
	}

	instance.client.SetProgramInput(atem.VideoInputType(payload.Settings.Input), uint8(payload.Settings.Input))
	return nil
}

// PGMWillAppearHandler Set ATEM PGM
func (a *App) PGMWillAppearHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	var payload streamdeck.WillAppearPayload[ProgramPropertyInspector]
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return xerrors.Errorf("failed to unmarshal payload: %w", err)
	}

	msg := fmt.Sprintf("WillAppear on PGM %v", payload.Settings)
	a.sd.LogMessage(ctx, msg)

	if _, ok := a.atems.SolveATEMByContext(event.Context); !ok {
		// initialize new instance
		if err := a.addATEMHost(ctx, event.Context, &ATEMInstance{
			client: atem.Create(payload.Settings.IP, true),
		}, true); err != nil {
			return xerrors.Errorf("failed to add ATEM host: %w", err)
		}
	}

	return nil
}

// Run Initialize and run the application
func Run(ctx context.Context) error {
	// Initialize the application
	app, err := NewApp(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}

	// Run the application
	return app.Run(ctx)
}
