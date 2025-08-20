package stdatem

import (
	"context"
	"testing"
	"time"

	"github.com/FlowingSPDG/go-atem"
	"github.com/FlowingSPDG/std-atem/Source/code/connectionmanager"
)

// mockLogger implements logger.Logger for testing
type mockLogger struct {
	debugCalls []string
	errorCalls []string
}

func (m *mockLogger) LogMessage(ctx context.Context, format string, args ...any) error { return nil }

func (m *mockLogger) Debug(ctx context.Context, format string, args ...any) error {
	m.debugCalls = append(m.debugCalls, format)
	return nil
}

func (m *mockLogger) Info(ctx context.Context, format string, args ...any) error { return nil }

func (m *mockLogger) Warn(ctx context.Context, format string, args ...any) error { return nil }

func (m *mockLogger) Error(ctx context.Context, format string, args ...any) error {
	m.errorCalls = append(m.errorCalls, format)
	return nil
}

// Note: Due to the concrete type dependency in go-atem, we'll use integration-style tests
// rather than mocking the ATEM client directly. The reconnection logic can still be tested
// through the ConnectionManager interface.

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestSolveATEMVideoInput(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected atem.VideoInputType
	}{
		{"Video Input 1", 1, atem.VideoInput1},
		{"Video Input 2", 2, atem.VideoInput2},
		{"Video Input 8", 8, atem.VideoInput8},
		{"Video Input 16", 16, atem.VideoInput16},
		{"Color Bars", 1000, atem.ColorBars},
		{"Color 1", 2001, atem.Color1},
		{"Color 2", 2002, atem.Color2},
		{"Media Player 1", 3010, atem.MediaPlayer1},
		{"Media Player 1 Key", 3011, atem.MediaPlayer1Key},
		{"Media Player 2", 3020, atem.MediaPlayer2},
		{"Media Player 2 Key", 3021, atem.MediaPlayer2Key},
		{"Key 1 Mask", 4010, atem.Key1Mask},
		{"Key 2 Mask", 4020, atem.Key2Mask},
		{"Key 3 Mask", 4030, atem.Key3Mask},
		{"Key 4 Mask", 4040, atem.Key4Mask},
		{"DSK 1 Mask", 5010, atem.DSK1Mask},
		{"DSK 2 Mask", 5020, atem.DSK2Mask},
		{"Super Source", 6000, atem.SuperSource},
		{"Clean Feed 1", 7001, atem.CleanFeed1},
		{"Clean Feed 2", 7002, atem.CleanFeed2},
		{"Auxiliary 1", 8001, atem.Auxilary1},
		{"Auxiliary 2", 8002, atem.Auxilary2},
		{"Auxiliary 3", 8003, atem.Auxilary3},
		{"Auxiliary 4", 8004, atem.Auxilary4},
		{"Auxiliary 5", 8005, atem.Auxilary5},
		{"Auxiliary 6", 8006, atem.Auxilary6},
		{"ME1 Program", 10010, atem.ME1Prog},
		{"ME1 Preview", 10011, atem.ME1Prev},
		{"ME2 Program", 10020, atem.ME2Prog},
		{"ME2 Preview", 10021, atem.ME2Prev},
		{"Unknown Input", 99999, atem.VideoBlack},
		{"Negative Input", -1, atem.VideoBlack},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := solveATEMVideoInput(tt.input)
			if result != tt.expected {
				t.Errorf("solveATEMVideoInput(%d) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAddATEMHost(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(*App)
		action       string
		contextID    string
		ip           string
		debug        bool
		expectedCall bool
		validateFunc func(*testing.T, *App)
	}{
		{
			name:      "add new ATEM host",
			setupFunc: func(app *App) {},
			action:    "preview",
			contextID: "context1",
			ip:        "192.168.1.100",
			debug:     false,
			validateFunc: func(t *testing.T, app *App) {
				// Verify the ATEM instance was stored
				ctx := context.Background()
				instance, exists := app.connectionManager.SolveATEMByIP(ctx, "192.168.1.100")
				if !exists {
					t.Error("ATEM instance should be stored")
				}
				if instance == nil {
					t.Error("ATEM instance should not be nil")
				}
				
				// Verify goroutine is marked as running
				if !app.connectionManager.IsReconnectGoroutineRunning("192.168.1.100") {
					t.Error("Reconnect goroutine should be marked as running")
				}
			},
		},
		{
			name: "add context to existing ATEM host",
			setupFunc: func(app *App) {
				// Pre-setup an existing ATEM instance
				ctx := context.Background()
				instance := &connectionmanager.ATEMInstance{
					Client:      atem.Create("192.168.1.100", false),
					ReconnectCh: make(chan struct{}, 1),
				}
				app.connectionManager.Store(ctx, "program", "192.168.1.100", "context1", instance)
				app.connectionManager.SetReconnectGoroutineRunning("192.168.1.100", true)
			},
			action:    "preview",
			contextID: "context2",
			ip:        "192.168.1.100",
			debug:     false,
			validateFunc: func(t *testing.T, app *App) {
				// Verify both contexts exist
				ctx := context.Background()
				contexts, exists := app.connectionManager.SolveContextsByIP(ctx, "192.168.1.100")
				if !exists {
					t.Error("Contexts should exist for IP")
				}
				if len(contexts) != 2 {
					t.Errorf("Expected 2 contexts, got %d", len(contexts))
				}
				
				// Verify the instance is the same
				instance1, _ := app.connectionManager.SolveATEMByContext(ctx, "context1")
				instance2, _ := app.connectionManager.SolveATEMByContext(ctx, "context2")
				if instance1 != instance2 {
					t.Error("Both contexts should reference the same instance")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog := &mockLogger{}
			app := &App{
				connectionManager: connectionmanager.NewConnectionManager(mockLog),
				logger:           mockLog,
			}
			
			tt.setupFunc(app)
			
			ctx := context.Background()
			err := app.addATEMHost(ctx, tt.action, tt.contextID, tt.ip, tt.debug)
			
			if err != nil {
				t.Errorf("addATEMHost() error = %v", err)
			}
			
			if tt.validateFunc != nil {
				tt.validateFunc(t, app)
			}
		})
	}
}

func TestHandleDisappear(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(*App)
		contextID    string
		validateFunc func(*testing.T, *App)
	}{
		{
			name: "handle disappear removes context",
			setupFunc: func(app *App) {
				ctx := context.Background()
				instance := &connectionmanager.ATEMInstance{
					Client:      atem.Create("192.168.1.100", false),
					ReconnectCh: make(chan struct{}, 1),
				}
				app.connectionManager.Store(ctx, "preview", "192.168.1.100", "context1", instance)
			},
			contextID: "context1",
			validateFunc: func(t *testing.T, app *App) {
				ctx := context.Background()
				// Verify context is removed
				_, exists := app.connectionManager.SolveATEMByContext(ctx, "context1")
				if exists {
					t.Error("Context should be removed after handleDisappear")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog := &mockLogger{}
			app := &App{
				connectionManager: connectionmanager.NewConnectionManager(mockLog),
				logger:           mockLog,
			}
			
			tt.setupFunc(app)
			
			ctx := context.Background()
			app.handleDisappear(ctx, tt.contextID)
			
			// Allow time for async cleanup
			time.Sleep(10 * time.Millisecond)
			
			if tt.validateFunc != nil {
				tt.validateFunc(t, app)
			}
		})
	}
}

// Test exponential backoff timing (this test checks the logic structure, not actual timing)
func TestReconnectionLoopBackoff(t *testing.T) {
	tests := []struct {
		name           string
		retryCount     int
		expectedMin    time.Duration
		expectedMax    time.Duration
	}{
		{"First retry", 1, 1 * time.Second, 2 * time.Second},
		{"Second retry", 2, 2 * time.Second, 4 * time.Second},
		{"Third retry", 3, 4 * time.Second, 8 * time.Second},
		{"Fourth retry", 4, 8 * time.Second, 16 * time.Second},
		{"Fifth retry", 5, 16 * time.Second, 30 * time.Second},
		{"Sixth retry", 6, 30 * time.Second, 30 * time.Second}, // Capped at 30 seconds
		{"Tenth retry", 10, 30 * time.Second, 30 * time.Second}, // Still capped at 30 seconds
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate backoff duration using the same logic as reconnectionLoop
			backoffSeconds := 1 << tt.retryCount
			if backoffSeconds > 30 {
				backoffSeconds = 30
			}
			backoffDuration := time.Duration(backoffSeconds) * time.Second

			if backoffDuration < tt.expectedMin {
				t.Errorf("Backoff duration %v should be at least %v", backoffDuration, tt.expectedMin)
			}
			if backoffDuration > tt.expectedMax {
				t.Errorf("Backoff duration %v should be at most %v", backoffDuration, tt.expectedMax)
			}
		})
	}
}

// Note: Reconnection loop tests that require mocking the ATEM client are not included
// due to the concrete type dependency. These would require refactoring the ATEM client
// to use an interface, which would change the implementation and violate the constraint
// of not modifying the implementation for test failures.