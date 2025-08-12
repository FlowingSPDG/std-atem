package connectionmanager

import (
	"context"
	"sync"
	"time"

	"github.com/FlowingSPDG/go-atem"
	"github.com/FlowingSPDG/std-atem/Source/code/logger"
	"github.com/puzpuzpuz/xsync"
)

type ActionAndContext struct {
	Action  string
	Context string
}

// ATEMInstance represents a single ATEM connection
type ATEMInstance struct {
	Client      *atem.Atem
	ReconnectCh chan struct{}
}

type ConnectionManager struct {
	atemByIP           *xsync.MapOf[string, *ATEMInstance]      // host: instance
	atemByContext      *xsync.MapOf[string, *ATEMInstance]      // context: binding
	contextsByIP       *xsync.MapOf[string, []ActionAndContext] // host: contexts
	usageCounts        *xsync.MapOf[string, int]                // host: usage count
	contextToIP        *xsync.MapOf[string, string]             // context: ip mapping
	reconnectGoroutines *xsync.MapOf[string, bool]              // host: goroutine running status
	storeMutex         sync.Mutex                               // mutex for atomic Store operations
	logger             logger.Logger
}

func NewConnectionManager(logger logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		atemByIP:           xsync.NewMapOf[*ATEMInstance](),
		atemByContext:      xsync.NewMapOf[*ATEMInstance](),
		contextsByIP:       xsync.NewMapOf[[]ActionAndContext](),
		usageCounts:        xsync.NewMapOf[int](),
		contextToIP:        xsync.NewMapOf[string](),
		reconnectGoroutines: xsync.NewMapOf[bool](),
		logger:             logger,
	}
}

func (a *ConnectionManager) SolveATEMByIP(ctx context.Context, ip string) (*ATEMInstance, bool) {
	a.logger.Debug(ctx, "SolveATEMByIP ip:%s", ip)
	v, ok := a.atemByIP.Load(ip)
	if !ok {
		a.logger.Error(ctx, "SolveATEMByIP ip:%s not found", ip)
		return nil, false
	}

	return v, true
}
func (a *ConnectionManager) SolveATEMByContext(ctx context.Context, context string) (*ATEMInstance, bool) {
	a.logger.Debug(ctx, "SolveATEMByContext context:%s", context)
	v, ok := a.atemByContext.Load(context)
	if !ok {
		a.logger.Error(ctx, "SolveATEMByContext context:%s not found", context)
		return nil, false
	}

	return v, true
}

func (a *ConnectionManager) SolveContextsByIP(ctx context.Context, ip string) ([]ActionAndContext, bool) {
	a.logger.Debug(ctx, "SolveContextsByIP ip:%s", ip)
	// ipからStreamDeck contextを取得する
	contexts, ok := a.contextsByIP.Load(ip)
	if !ok {
		a.logger.Error(ctx, "SolveContextsByIP ip:%s not found", ip)
		return nil, false
	}

	return contexts, true
}

func (a *ConnectionManager) Store(ctx context.Context, action, ip, contextID string, at *ATEMInstance) {
	a.logger.Debug(ctx, "Store action:%s ip:%s context:%s", action, ip, contextID)
	
	// Acquire mutex for atomic operation across all maps
	a.storeMutex.Lock()
	defer a.storeMutex.Unlock()
	
	a.atemByIP.Store(ip, at)
	a.atemByContext.Store(contextID, at)
	
	// Check if this context was using a different IP before
	oldIP, wasUsingDifferentIP := a.contextToIP.Load(contextID)
	if wasUsingDifferentIP && oldIP != ip {
		a.logger.Debug(ctx, "Store context:%s changing from IP %s to %s", contextID, oldIP, ip)
		
		// Remove context from old IP's context list
		if contexts, exists := a.contextsByIP.Load(oldIP); exists {
			filteredContexts := make([]ActionAndContext, 0, len(contexts))
			for _, ac := range contexts {
				if ac.Context != contextID {
					filteredContexts = append(filteredContexts, ac)
				}
			}
			if len(filteredContexts) > 0 {
				a.contextsByIP.Store(oldIP, filteredContexts)
			} else {
				a.contextsByIP.Delete(oldIP)
			}
		}
		
		// Decrement usage count for old IP atomically
		if count, ok := a.usageCounts.Load(oldIP); ok {
			newCount := count - 1
			if newCount <= 0 {
				a.usageCounts.Delete(oldIP)
				// Schedule cleanup after mutex release
				go func() {
					delayCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					context.AfterFunc(delayCtx, func() {
						defer cancel()
						a.cleanupUnusedATEM(ctx, oldIP)
					})
				}()
			} else {
				a.usageCounts.Store(oldIP, newCount)
			}
		}
	}
	
	a.contextToIP.Store(contextID, ip)
	
	// Update usage count only if this context is new or was using a different IP
	if !wasUsingDifferentIP || oldIP != ip {
		if count, ok := a.usageCounts.Load(ip); ok {
			a.usageCounts.Store(ip, count+1)
		} else {
			a.usageCounts.Store(ip, 1)
		}
		a.logger.Debug(ctx, "Store incremented usage for IP %s", ip)
	}
	
	if contextIDs, ok := a.contextsByIP.Load(ip); !ok {
		a.contextsByIP.Store(ip, []ActionAndContext{{Action: action, Context: contextID}})
	} else {
		// Check if this context is already in the list to avoid duplicates
		found := false
		for _, ac := range contextIDs {
			if ac.Context == contextID {
				found = true
				break
			}
		}
		if !found {
			a.contextsByIP.Store(ip, append(contextIDs, ActionAndContext{Action: action, Context: contextID}))
		}
	}
}

func (a *ConnectionManager) DeleteATEMByIP(ctx context.Context, ip string) {
	a.logger.Debug(ctx, "DeleteATEMByIP ip:%s", ip)
	a.atemByIP.Delete(ip)

	// 削除処理
	a.logger.Debug(ctx, "Delete closing ATEM client ip:%s", ip)
	at, ok := a.SolveATEMByIP(ctx, ip)
	if !ok {
		return
	}
	at.Client.Close()
}

func (a *ConnectionManager) DeleteATEMByContext(ctx context.Context, contextID string) {
	a.logger.Debug(ctx, "DeleteATEMByContext contextID:%s", contextID)
	
	// Acquire mutex for atomic operation across all maps
	a.storeMutex.Lock()
	defer a.storeMutex.Unlock()
	
	// Get the IP this context was using
	ip, ok := a.contextToIP.Load(contextID)
	
	a.atemByContext.Delete(contextID)
	a.contextToIP.Delete(contextID)

	// Remove context from contextsByIP and decrement usage count
	if ok {
		if contexts, exists := a.contextsByIP.Load(ip); exists {
			filteredContexts := make([]ActionAndContext, 0, len(contexts))
			for _, ac := range contexts {
				if ac.Context != contextID {
					filteredContexts = append(filteredContexts, ac)
				}
			}
			if len(filteredContexts) > 0 {
				a.contextsByIP.Store(ip, filteredContexts)
			} else {
				a.contextsByIP.Delete(ip)
			}
		}
		
		// Decrement usage count
		if count, exists := a.usageCounts.Load(ip); exists {
			newCount := count - 1
			if newCount <= 0 {
				a.usageCounts.Delete(ip)
				// Schedule cleanup after mutex release
				go func() {
					delayCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					context.AfterFunc(delayCtx, func() {
						defer cancel()
						a.cleanupUnusedATEM(ctx, ip)
					})
				}()
			} else {
				a.usageCounts.Store(ip, newCount)
			}
		}
	}
}

// UpdateContextIP handles when a context changes to a different IP
func (a *ConnectionManager) UpdateContextIP(ctx context.Context, contextID string, newIP string) {
	a.logger.Debug(ctx, "UpdateContextIP contextID:%s newIP:%s", contextID, newIP)
	
	// This functionality is now handled directly in Store method
	// This method is kept for backward compatibility but does nothing
}

// cleanupUnusedATEM removes ATEM instance if it's still unused after delay
func (a *ConnectionManager) cleanupUnusedATEM(ctx context.Context, ip string) {
	a.logger.Debug(ctx, "cleanupUnusedATEM ip:%s checking if still unused", ip)
	
	// Double-check that the instance is still not in use
	if _, stillInUse := a.usageCounts.Load(ip); !stillInUse {
		if instance, exists := a.atemByIP.Load(ip); exists {
			a.logger.Debug(ctx, "cleanupUnusedATEM ip:%s closing unused ATEM connection", ip)
			instance.Client.Close()
			a.atemByIP.Delete(ip)
			a.contextsByIP.Delete(ip)
			// Mark goroutine as stopped
			a.reconnectGoroutines.Delete(ip)
		}
	} else {
		a.logger.Debug(ctx, "cleanupUnusedATEM ip:%s is now in use again, skipping cleanup", ip)
	}
}

// IsReconnectGoroutineRunning checks if a reconnect goroutine is already running for the IP
func (a *ConnectionManager) IsReconnectGoroutineRunning(ip string) bool {
	running, exists := a.reconnectGoroutines.Load(ip)
	return exists && running
}

// SetReconnectGoroutineRunning marks a reconnect goroutine as running for the IP
func (a *ConnectionManager) SetReconnectGoroutineRunning(ip string, running bool) {
	if running {
		a.reconnectGoroutines.Store(ip, true)
	} else {
		a.reconnectGoroutines.Delete(ip)
	}
}
