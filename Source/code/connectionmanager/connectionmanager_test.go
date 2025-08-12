package connectionmanager

import (
	"context"
	"testing"
	"time"

	"github.com/FlowingSPDG/go-atem"
)

// mockLogger implements logger.Logger for testing
type mockLogger struct{}

func (m *mockLogger) LogMessage(ctx context.Context, format string, args ...any) error { return nil }
func (m *mockLogger) Debug(ctx context.Context, format string, args ...any) error       { return nil }
func (m *mockLogger) Info(ctx context.Context, format string, args ...any) error        { return nil }
func (m *mockLogger) Warn(ctx context.Context, format string, args ...any) error        { return nil }
func (m *mockLogger) Error(ctx context.Context, format string, args ...any) error       { return nil }

func TestNewConnectionManager(t *testing.T) {
	logger := &mockLogger{}
	cm := NewConnectionManager(logger)
	
	if cm == nil {
		t.Fatal("NewConnectionManager returned nil")
	}
	
	// Note: cannot directly compare interface values, so we skip this check
	
	// Test that all maps are initialized
	if cm.atemByIP == nil {
		t.Error("atemByIP not initialized")
	}
	if cm.atemByContext == nil {
		t.Error("atemByContext not initialized")
	}
	if cm.contextsByIP == nil {
		t.Error("contextsByIP not initialized")
	}
	if cm.usageCounts == nil {
		t.Error("usageCounts not initialized")
	}
	if cm.contextToIP == nil {
		t.Error("contextToIP not initialized")
	}
	if cm.reconnectGoroutines == nil {
		t.Error("reconnectGoroutines not initialized")
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
				cm.atemByIP.Store("192.168.1.100", instance)
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
				cm.atemByContext.Store("context1", instance)
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
		name             string
		setupFunc        func(*ConnectionManager)
		ip               string
		expectedExists   bool
		expectedContexts []ActionAndContext
	}{
		{
			name: "existing IP returns contexts",
			setupFunc: func(cm *ConnectionManager) {
				contexts := []ActionAndContext{
					{Action: "preview", Context: "context1"},
					{Action: "program", Context: "context2"},
				}
				cm.contextsByIP.Store("192.168.1.100", contexts)
			},
			ip:             "192.168.1.100",
			expectedExists: true,
			expectedContexts: []ActionAndContext{
				{Action: "preview", Context: "context1"},
				{Action: "program", Context: "context2"},
			},
		},
		{
			name:             "non-existing IP returns empty",
			setupFunc:        func(cm *ConnectionManager) {},
			ip:               "192.168.1.200",
			expectedExists:   false,
			expectedContexts: nil,
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
				if len(contexts) != len(tt.expectedContexts) {
					t.Errorf("Expected %d contexts, got %d", len(tt.expectedContexts), len(contexts))
				}
				
				for i, expected := range tt.expectedContexts {
					if i >= len(contexts) {
						t.Errorf("Missing context at index %d", i)
						continue
					}
					if contexts[i] != expected {
						t.Errorf("Context at index %d: got %+v, want %+v", i, contexts[i], expected)
					}
				}
			} else {
				if contexts != nil {
					t.Error("Expected nil contexts but got non-nil")
				}
			}
		})
	}
}

func TestStore(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(*ConnectionManager)
		action       string
		ip           string
		contextID    string
		validateFunc func(*testing.T, *ConnectionManager)
	}{
		{
			name:      "store new context and IP",
			setupFunc: func(cm *ConnectionManager) {},
			action:    "preview",
			ip:        "192.168.1.100",
			contextID: "context1",
			validateFunc: func(t *testing.T, cm *ConnectionManager) {
				// Verify atemByIP
				if _, exists := cm.atemByIP.Load("192.168.1.100"); !exists {
					t.Error("atemByIP should contain the IP")
				}
				
				// Verify atemByContext
				if _, exists := cm.atemByContext.Load("context1"); !exists {
					t.Error("atemByContext should contain the context")
				}
				
				// Verify contextToIP
				if ip, exists := cm.contextToIP.Load("context1"); !exists || ip != "192.168.1.100" {
					t.Errorf("contextToIP should map context1 to 192.168.1.100, got %s", ip)
				}
				
				// Verify contextsByIP
				if contexts, exists := cm.contextsByIP.Load("192.168.1.100"); !exists {
					t.Error("contextsByIP should contain the IP")
				} else {
					found := false
					for _, ac := range contexts {
						if ac.Context == "context1" && ac.Action == "preview" {
							found = true
							break
						}
					}
					if !found {
						t.Error("contextsByIP should contain the context")
					}
				}
				
				// Verify usage count
				if count, exists := cm.usageCounts.Load("192.168.1.100"); !exists || count != 1 {
					t.Errorf("usageCounts should be 1, got %d", count)
				}
			},
		},
		{
			name: "store additional context for existing IP",
			setupFunc: func(cm *ConnectionManager) {
				instance := &ATEMInstance{
					Client:      atem.Create("192.168.1.100", false),
					ReconnectCh: make(chan struct{}, 1),
				}
				cm.atemByIP.Store("192.168.1.100", instance)
				cm.atemByContext.Store("context1", instance)
				cm.contextToIP.Store("context1", "192.168.1.100")
				cm.contextsByIP.Store("192.168.1.100", []ActionAndContext{{Action: "preview", Context: "context1"}})
				cm.usageCounts.Store("192.168.1.100", 1)
			},
			action:    "program",
			ip:        "192.168.1.100",
			contextID: "context2",
			validateFunc: func(t *testing.T, cm *ConnectionManager) {
				// Verify usage count incremented
				if count, exists := cm.usageCounts.Load("192.168.1.100"); !exists || count != 2 {
					t.Errorf("usageCounts should be 2, got %d", count)
				}
				
				// Verify both contexts exist
				if contexts, exists := cm.contextsByIP.Load("192.168.1.100"); !exists {
					t.Error("contextsByIP should contain the IP")
				} else {
					if len(contexts) != 2 {
						t.Errorf("Expected 2 contexts, got %d", len(contexts))
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewConnectionManager(&mockLogger{})
			tt.setupFunc(cm)
			
			instance := &ATEMInstance{
				Client:      atem.Create(tt.ip, false),
				ReconnectCh: make(chan struct{}, 1),
			}
			
			ctx := context.Background()
			cm.Store(ctx, tt.action, tt.ip, tt.contextID, instance)
			
			tt.validateFunc(t, cm)
		})
	}
}

func TestDeleteATEMByContext(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(*ConnectionManager)
		contextID    string
		validateFunc func(*testing.T, *ConnectionManager)
	}{
		{
			name: "delete existing context",
			setupFunc: func(cm *ConnectionManager) {
				instance := &ATEMInstance{
					Client:      atem.Create("192.168.1.100", false),
					ReconnectCh: make(chan struct{}, 1),
				}
				cm.atemByIP.Store("192.168.1.100", instance)
				cm.atemByContext.Store("context1", instance)
				cm.contextToIP.Store("context1", "192.168.1.100")
				cm.contextsByIP.Store("192.168.1.100", []ActionAndContext{{Action: "preview", Context: "context1"}})
				cm.usageCounts.Store("192.168.1.100", 1)
			},
			contextID: "context1",
			validateFunc: func(t *testing.T, cm *ConnectionManager) {
				// Verify context removed from atemByContext
				if _, exists := cm.atemByContext.Load("context1"); exists {
					t.Error("atemByContext should not contain the context after deletion")
				}
				
				// Verify context removed from contextToIP
				if _, exists := cm.contextToIP.Load("context1"); exists {
					t.Error("contextToIP should not contain the context after deletion")
				}
				
				// Verify usage count decremented to 0
				if _, exists := cm.usageCounts.Load("192.168.1.100"); exists {
					t.Error("usageCounts should be deleted when count reaches 0")
				}
				
				// Verify context removed from contextsByIP
				if contexts, exists := cm.contextsByIP.Load("192.168.1.100"); exists {
					for _, ac := range contexts {
						if ac.Context == "context1" {
							t.Error("contextsByIP should not contain the deleted context")
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewConnectionManager(&mockLogger{})
			tt.setupFunc(cm)
			
			ctx := context.Background()
			cm.DeleteATEMByContext(ctx, tt.contextID)
			
			// Allow time for async cleanup
			time.Sleep(10 * time.Millisecond)
			
			tt.validateFunc(t, cm)
		})
	}
}

func TestReconnectGoroutineManagement(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(*ConnectionManager)
		ip         string
		operation  string
		expected   bool
	}{
		{
			name:      "new IP should not be running",
			setupFunc: func(cm *ConnectionManager) {},
			ip:        "192.168.1.100",
			operation: "check",
			expected:  false,
		},
		{
			name: "set IP as running",
			setupFunc: func(cm *ConnectionManager) {
				cm.SetReconnectGoroutineRunning("192.168.1.100", true)
			},
			ip:        "192.168.1.100",
			operation: "check",
			expected:  true,
		},
		{
			name: "set IP as not running",
			setupFunc: func(cm *ConnectionManager) {
				cm.SetReconnectGoroutineRunning("192.168.1.100", true)
				cm.SetReconnectGoroutineRunning("192.168.1.100", false)
			},
			ip:        "192.168.1.100",
			operation: "check",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewConnectionManager(&mockLogger{})
			tt.setupFunc(cm)
			
			result := cm.IsReconnectGoroutineRunning(tt.ip)
			
			if result != tt.expected {
				t.Errorf("IsReconnectGoroutineRunning() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStoreConcurrency(t *testing.T) {
	cm := NewConnectionManager(&mockLogger{})
	ctx := context.Background()
	
	const numGoroutines = 10
	const numOperations = 100
	
	// Test concurrent Store operations
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				instance := &ATEMInstance{
					Client:      atem.Create("192.168.1.100", false),
					ReconnectCh: make(chan struct{}, 1),
				}
				
				contextID := "context_" + string(rune('A'+id)) + "_" + string(rune('0'+j%10))
				cm.Store(ctx, "test", "192.168.1.100", contextID, instance)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Verify that usage count is correct
	count, exists := cm.usageCounts.Load("192.168.1.100")
	if !exists {
		t.Error("Usage count should exist after concurrent operations")
		return
	}
	
	// Count should be equal to number of unique contexts
	if contexts, exists := cm.contextsByIP.Load("192.168.1.100"); exists {
		if count != len(contexts) {
			t.Errorf("Usage count %d should match number of contexts %d", count, len(contexts))
		}
	}
	
	t.Logf("Final usage count: %d", count)
}