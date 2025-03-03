package stdatem

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/FlowingSPDG/go-atem"
	"github.com/FlowingSPDG/streamdeck"
)

// ATEMInstance represents a single ATEM connection
type ATEMInstance struct {
	client      *atem.Atem
	config      ATEMHost
	state       state
	reconnectCh chan struct{}
}

// App Main Engine
type App struct {
	buttons  buttons                  // Setting per context(button)
	sd       *streamdeck.Client       // StreamDeck Client
	atems    map[string]*ATEMInstance // ATEM Clients mapped by IP
	atemsMux sync.RWMutex             // Mutex for ATEM clients map
}

// NewApp Initiate App main engine
func NewApp(ctx context.Context, cfg Config) (*App, error) {
	app := &App{
		buttons: buttons{
			m: &sync.Map{},
		},
		atems: make(map[string]*ATEMInstance),
	}

	// Setup SD
	params, err := streamdeck.ParseRegistrationParams(cfg.Params)
	if err != nil {
		return nil, err
	}
	app.sd = streamdeck.NewClient(ctx, params)
	app.setupSD()

	// Setup ATEMs
	for _, host := range cfg.ATEMHosts {
		if err := app.addATEMHost(ctx, host, cfg.Debug); err != nil {
			return nil, err
		}
	}

	return app, nil
}

// addATEMHost adds a new ATEM host and sets up its connection
func (a *App) addATEMHost(ctx context.Context, host ATEMHost, debug bool) error {
	instance := &ATEMInstance{
		client:      atem.Create(host.IP, debug),
		config:      host,
		reconnectCh: make(chan struct{}, 1),
	}

	a.atemsMux.Lock()
	a.atems[host.IP] = instance
	a.atemsMux.Unlock()

	instance.client.On("connected", func() {
		a.atemsMux.RLock()
		instance, exists := a.atems[host.IP]
		a.atemsMux.RUnlock()

		if !exists {
			return
		}

		instance.state = state{
			Preview:   instance.client.PreviewInput.Index,
			Program:   instance.client.ProgramInput.Index,
			Connected: instance.client.Connected(),
		}
	})

	instance.client.On("closed", func() {
		a.atemsMux.RLock()
		instance, exists := a.atems[host.IP]
		a.atemsMux.RUnlock()

		if !exists {
			return
		}

		instance.state.Connected = instance.client.Connected()

		// Trigger reconnection if enabled
		if instance.config.AutoReconnect {
			select {
			case instance.reconnectCh <- struct{}{}:
			default:
			}
		}
	})

	// Start reconnection goroutine if auto-reconnect is enabled
	if host.AutoReconnect {
		go a.reconnectionLoop(ctx, host.IP)
	}

	return nil
}

// Run Run Background Process.
func (a *App) Run() error {
	return a.sd.Run()
}

// ConnectATEM Connect to all configured ATEM hosts.
func (a *App) ConnectATEM() error {
	a.atemsMux.RLock()
	defer a.atemsMux.RUnlock()

	for _, instance := range a.atems {
		if err := instance.client.Connect(); err != nil {
			return err
		}
	}
	return nil
}

// setup StreamDeck Client
func (a *App) setupSD() {
	prv := a.sd.Action(SetPreviewAction)
	prv.RegisterHandler(streamdeck.KeyDown, a.PRVKeyDownHandler)

	pgm := a.sd.Action(SetProgramAction)
	pgm.RegisterHandler(streamdeck.KeyDown, a.PGMKeyDownHandler)
}

// reconnectionLoop handles automatic reconnection for a specific ATEM host
func (a *App) reconnectionLoop(ctx context.Context, ip string) {
	a.atemsMux.RLock()
	instance, exists := a.atems[ip]
	a.atemsMux.RUnlock()

	if !exists {
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

// GetATEMByContext returns the ATEM instance associated with the button context
func (a *App) GetATEMByContext(ctx string) (*ATEMInstance, error) {
	s, err := a.buttons.Load(ctx)
	if err != nil {
		return nil, err
	}

	a.atemsMux.RLock()
	defer a.atemsMux.RUnlock()

	// Find the ATEM instance for this button
	for _, instance := range a.atems {
		if instance.config.Name == s.atemName {
			return instance, nil
		}
	}

	return nil, fmt.Errorf("no ATEM found for context %s", ctx)
}

// PRVKeyDownHandler Set ATEM PRV
func (a *App) PRVKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	instance, err := a.GetATEMByContext(event.Context)
	if err != nil {
		return err
	}

	s, err := a.buttons.Load(event.Context)
	if err != nil {
		return err
	}

	instance.client.SetPreviewInput(atem.VideoInputType(s.input), uint8(s.input))
	return nil
}

// PGMKeyDownHandler Set ATEM PGM
func (a *App) PGMKeyDownHandler(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
	instance, err := a.GetATEMByContext(event.Context)
	if err != nil {
		return err
	}

	s, err := a.buttons.Load(event.Context)
	if err != nil {
		return err
	}

	instance.client.SetProgramInput(atem.VideoInputType(s.input), uint8(s.input))
	return nil
}

// Run Initialize and run the application
func Run(ctx context.Context) error {
	// Load settings from environment variables
	cfg := Config{
		Params: os.Args,
		ATEMHosts: []ATEMHost{
			{
				IP:            os.Getenv("ATEM_IP"),
				Name:          os.Getenv("ATEM_NAME"),
				AutoReconnect: true,
			},
		},
		Debug: os.Getenv("DEBUG") == "true",
	}

	// Initialize the application
	app, err := NewApp(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}

	// Connect to all ATEM hosts
	if err := app.ConnectATEM(); err != nil {
		return fmt.Errorf("failed to connect to ATEM: %w", err)
	}

	// Run the application
	return app.Run()
}
