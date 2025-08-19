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

// ConnectionInfo 接続情報を統合管理
type ConnectionInfo struct {
	Instance         *ATEMInstance
	Contexts         []string
	UsageCount       int
	ReconnectRunning bool
}

type ConnectionManager struct {
	connections *xsync.MapOf[string, *ConnectionInfo] // IP -> ConnectionInfo
	contexts    *xsync.MapOf[string, string]          // Context -> IP
	logger      logger.Logger
	storeMutex  sync.Mutex // mutex for atomic Store operations
}

func NewConnectionManager(logger logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		connections: xsync.NewMapOf[*ConnectionInfo](),
		contexts:    xsync.NewMapOf[string](),
		logger:      logger,
	}
}

func (a *ConnectionManager) SolveATEMByIP(ctx context.Context, ip string) (*ATEMInstance, bool) {
	a.logger.Debug(ctx, "SolveATEMByIP ip:%s", ip)
	info, ok := a.connections.Load(ip)
	if !ok {
		a.logger.Error(ctx, "SolveATEMByIP ip:%s not found", ip)
		return nil, false
	}
	return info.Instance, true
}

func (a *ConnectionManager) SolveATEMByContext(ctx context.Context, context string) (*ATEMInstance, bool) {
	a.logger.Debug(ctx, "SolveATEMByContext context:%s", context)
	ip, ok := a.contexts.Load(context)
	if !ok {
		a.logger.Error(ctx, "SolveATEMByContext context:%s not found", context)
		return nil, false
	}
	return a.SolveATEMByIP(ctx, ip)
}

func (a *ConnectionManager) SolveContextsByIP(ctx context.Context, ip string) ([]ActionAndContext, bool) {
	a.logger.Debug(ctx, "SolveContextsByIP ip:%s", ip)
	info, ok := a.connections.Load(ip)
	if !ok {
		a.logger.Error(ctx, "SolveContextsByIP ip:%s not found", ip)
		return nil, false
	}

	// ActionAndContextの配列に変換
	contexts := make([]ActionAndContext, 0, len(info.Contexts))
	for _, contextID := range info.Contexts {
		// ここではActionを特定できないため、空文字列を設定
		// 必要に応じて、context -> action のマッピングを追加
		contexts = append(contexts, ActionAndContext{
			Action:  "", // TODO: action情報を追加する必要がある場合
			Context: contextID,
		})
	}
	return contexts, true
}

func (a *ConnectionManager) Store(ctx context.Context, action, ip, contextID string, at *ATEMInstance) {
	a.logger.Debug(ctx, "Store action:%s ip:%s context:%s", action, ip, contextID)

	// Acquire mutex for atomic operation across all maps
	a.storeMutex.Lock()
	defer a.storeMutex.Unlock()

	// Check if this context was using a different IP before
	oldIP, wasUsingDifferentIP := a.contexts.Load(contextID)
	if wasUsingDifferentIP && oldIP != ip {
		a.logger.Debug(ctx, "Store context:%s changing from IP %s to %s", contextID, oldIP, ip)
		a.removeContextFromIP(ctx, contextID, oldIP)
	}

	// Update or create connection info
	info, exists := a.connections.Load(ip)
	if !exists {
		info = &ConnectionInfo{
			Instance:   at,
			Contexts:   []string{contextID},
			UsageCount: 1,
		}
	} else {
		// Check if context is already in the list to avoid duplicates
		found := false
		for _, existingContext := range info.Contexts {
			if existingContext == contextID {
				found = true
				break
			}
		}
		if !found {
			info.Contexts = append(info.Contexts, contextID)
			info.UsageCount++
		}
	}

	a.connections.Store(ip, info)
	a.contexts.Store(contextID, ip)
}

func (a *ConnectionManager) DeleteATEMByIP(ctx context.Context, ip string) {
	a.logger.Debug(ctx, "DeleteATEMByIP ip:%s", ip)
	info, ok := a.connections.Load(ip)
	if !ok {
		return
	}

	// Close ATEM client
	if info.Instance != nil {
		a.logger.Debug(ctx, "Delete closing ATEM client ip:%s", ip)
		info.Instance.Client.Close()
	}

	a.connections.Delete(ip)
}

func (a *ConnectionManager) DeleteATEMByContext(ctx context.Context, contextID string) {
	a.logger.Debug(ctx, "DeleteATEMByContext contextID:%s", contextID)

	// Acquire mutex for atomic operation
	a.storeMutex.Lock()
	defer a.storeMutex.Unlock()

	// Get the IP this context was using
	ip, ok := a.contexts.Load(contextID)
	if !ok {
		return
	}

	a.contexts.Delete(contextID)
	a.removeContextFromIP(ctx, contextID, ip)
}

// UpdateContextIP handles when a context changes to a different IP
func (a *ConnectionManager) UpdateContextIP(ctx context.Context, contextID string, newIP string) {
	a.logger.Debug(ctx, "UpdateContextIP contextID:%s newIP:%s", contextID, newIP)

	// This functionality is now handled directly in Store method
	// This method is kept for backward compatibility but does nothing
}

// removeContextFromIP removes a context from an IP's context list and handles cleanup
func (a *ConnectionManager) removeContextFromIP(ctx context.Context, contextID, ip string) {
	info, exists := a.connections.Load(ip)
	if !exists {
		return
	}

	// Remove context from the list
	filteredContexts := make([]string, 0, len(info.Contexts))
	for _, existingContext := range info.Contexts {
		if existingContext != contextID {
			filteredContexts = append(filteredContexts, existingContext)
		}
	}

	info.Contexts = filteredContexts
	info.UsageCount--

	if info.UsageCount <= 0 {
		// Schedule cleanup after mutex release
		go func() {
			delayCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			time.AfterFunc(5*time.Second, func() {
				a.cleanupUnusedATEM(delayCtx, ip)
			})
		}()
	} else {
		a.connections.Store(ip, info)
	}
}

// cleanupUnusedATEM removes ATEM instance if it's still unused after delay
func (a *ConnectionManager) cleanupUnusedATEM(ctx context.Context, ip string) {
	a.logger.Debug(ctx, "cleanupUnusedATEM ip:%s checking if still unused", ip)

	// Double-check that the instance is still not in use
	info, exists := a.connections.Load(ip)
	if !exists || info.UsageCount <= 0 {
		if exists && info.Instance != nil {
			a.logger.Debug(ctx, "cleanupUnusedATEM ip:%s closing unused ATEM connection", ip)
			info.Instance.Client.Close()
		}
		a.connections.Delete(ip)
		// Mark goroutine as stopped
		if exists {
			info.ReconnectRunning = false
		}
	} else {
		a.logger.Debug(ctx, "cleanupUnusedATEM ip:%s is now in use again, skipping cleanup", ip)
	}
}

// IsReconnectGoroutineRunning checks if a reconnect goroutine is already running for the IP
func (a *ConnectionManager) IsReconnectGoroutineRunning(ip string) bool {
	info, exists := a.connections.Load(ip)
	return exists && info.ReconnectRunning
}

// SetReconnectGoroutineRunning marks a reconnect goroutine as running for the IP
func (a *ConnectionManager) SetReconnectGoroutineRunning(ip string, running bool) {
	info, exists := a.connections.Load(ip)
	if exists {
		info.ReconnectRunning = running
		a.connections.Store(ip, info)
	}
}
