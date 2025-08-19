package connectionmanager

import (
	"context"
	"testing"

	"github.com/FlowingSPDG/go-atem"
)

type mockLogger struct{}

func (m *mockLogger) LogMessage(ctx context.Context, format string, args ...any) error { return nil }
func (m *mockLogger) Debug(ctx context.Context, msg string, args ...any) error         { return nil }
func (m *mockLogger) Info(ctx context.Context, msg string, args ...any) error          { return nil }
func (m *mockLogger) Warn(ctx context.Context, msg string, args ...any) error          { return nil }
func (m *mockLogger) Error(ctx context.Context, msg string, args ...any) error         { return nil }

func TestNewConnectionManager(t *testing.T) {
	cm := NewConnectionManager(&mockLogger{})
	if cm == nil {
		t.Fatal("NewConnectionManager returned nil")
	}

	// Test that all maps are initialized
	if cm.connections == nil {
		t.Error("connections not initialized")
	}
	if cm.contexts == nil {
		t.Error("contexts not initialized")
	}
}

func TestSolveATEMByIP(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(*ConnectionManager)
		ip             string
		expectedExists bool
	}{
		{
			name: "existing IP returns instance",
			setupFunc: func(cm *ConnectionManager) {
				instance := &ATEMInstance{
					Client:      atem.Create("192.168.1.100", false),
					ReconnectCh: make(chan struct{}, 1),
				}
				info := &ConnectionInfo{
					Instance: instance,
					Contexts: []string{"context1"},
				}
				cm.connections.Store("192.168.1.100", info)
			},
			ip:             "192.168.1.100",
			expectedExists: true,
		},
		{
			name:           "non-existing IP returns nil",
			setupFunc:      func(cm *ConnectionManager) {},
			ip:             "192.168.1.200",
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewConnectionManager(&mockLogger{})
			tt.setupFunc(cm)

			ctx := context.Background()
			instance, exists := cm.SolveATEMByIP(ctx, tt.ip)

			if exists != tt.expectedExists {
				t.Errorf("SolveATEMByIP() exists = %v, want %v", exists, tt.expectedExists)
			}

			if tt.expectedExists && instance == nil {
				t.Error("Expected instance but got nil")
			}

			if !tt.expectedExists && instance != nil {
				t.Error("Expected nil but got instance")
			}
		})
	}
}

func TestSolveATEMByContext(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(*ConnectionManager)
		contextID      string
		expectedExists bool
	}{
		{
			name: "existing context returns instance",
			setupFunc: func(cm *ConnectionManager) {
				instance := &ATEMInstance{
					Client:      atem.Create("192.168.1.100", false),
					ReconnectCh: make(chan struct{}, 1),
				}
				info := &ConnectionInfo{
					Instance: instance,
					Contexts: []string{"context1"},
				}
				cm.connections.Store("192.168.1.100", info)
				cm.contexts.Store("context1", "192.168.1.100")
			},
			contextID:      "context1",
			expectedExists: true,
		},
		{
			name:           "non-existing context returns nil",
			setupFunc:      func(cm *ConnectionManager) {},
			contextID:      "context2",
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewConnectionManager(&mockLogger{})
			tt.setupFunc(cm)

			ctx := context.Background()
			instance, exists := cm.SolveATEMByContext(ctx, tt.contextID)

			if exists != tt.expectedExists {
				t.Errorf("SolveATEMByContext() exists = %v, want %v", exists, tt.expectedExists)
			}

			if tt.expectedExists && instance == nil {
				t.Error("Expected instance but got nil")
			}

			if !tt.expectedExists && instance != nil {
				t.Error("Expected nil but got instance")
			}
		})
	}
}

func TestSolveContextsByIP(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(*ConnectionManager)
		ip             string
		expectedExists bool
		expectedCount  int
	}{
		{
			name: "existing IP returns contexts",
			setupFunc: func(cm *ConnectionManager) {
				instance := &ATEMInstance{
					Client:      atem.Create("192.168.1.100", false),
					ReconnectCh: make(chan struct{}, 1),
				}
				info := &ConnectionInfo{
					Instance: instance,
					Contexts: []string{"context1", "context2"},
				}
				cm.connections.Store("192.168.1.100", info)
			},
			ip:             "192.168.1.100",
			expectedExists: true,
			expectedCount:  2,
		},
		{
			name:           "non-existing IP returns nil",
			setupFunc:      func(cm *ConnectionManager) {},
			ip:             "192.168.1.200",
			expectedExists: false,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewConnectionManager(&mockLogger{})
			tt.setupFunc(cm)

			ctx := context.Background()
			contexts, exists := cm.SolveContextsByIP(ctx, tt.ip)

			if exists != tt.expectedExists {
				t.Errorf("SolveContextsByIP() exists = %v, want %v", exists, tt.expectedExists)
			}

			if tt.expectedExists {
				if len(contexts) != tt.expectedCount {
					t.Errorf("Expected %d contexts, got %d", tt.expectedCount, len(contexts))
				}
			} else {
				if contexts != nil {
					t.Error("Expected nil contexts but got non-nil")
				}
			}
		})
	}
}
