package connectionmanager

import (
	"context"
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
	atemByIP      *xsync.MapOf[string, *ATEMInstance]      // host: instance
	atemByContext *xsync.MapOf[string, *ATEMInstance]      // context: binding
	contextsByIP  *xsync.MapOf[string, []ActionAndContext] // host: contexts
	usageCounts   *xsync.MapOf[string, int]                // host: usage count
	contextToIP   *xsync.MapOf[string, string]             // context: ip mapping
	logger        logger.Logger
}

func NewConnectionManager(logger logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		atemByIP:      xsync.NewMapOf[*ATEMInstance](),
		atemByContext: xsync.NewMapOf[*ATEMInstance](),
		contextsByIP:  xsync.NewMapOf[[]ActionAndContext](),
		usageCounts:   xsync.NewMapOf[int](),
		contextToIP:   xsync.NewMapOf[string](),
		logger:        logger,
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
	var contexts []ActionAndContext
	var ok bool

	a.contextsByIP.Range(func(key string, value []ActionAndContext) bool {
		if key == ip {
			contexts = append(contexts, value...)
			ok = true
			return false
		}

		return false
	})

	if !ok {
		a.logger.Error(ctx, "SolveContextsByIP ip:%s not found", ip)
	}

	return contexts, ok
}

func (a *ConnectionManager) Store(ctx context.Context, action, ip, context string, at *ATEMInstance) {
	a.logger.Debug(ctx, "Store action:%s ip:%s context:%s", action, ip, context)
	a.atemByIP.Store(ip, at)
	a.atemByContext.Store(context, at)
	
	// Check if this context was using a different IP before
	oldIP, wasUsingDifferentIP := a.contextToIP.Load(context)
	if wasUsingDifferentIP && oldIP != ip {
		a.logger.Debug(ctx, "Store context:%s changing from IP %s to %s", context, oldIP, ip)
		a.decrementUsageAndCleanup(ctx, oldIP)
		
		// Remove context from old IP's context list
		if contexts, exists := a.contextsByIP.Load(oldIP); exists {
			filteredContexts := make([]ActionAndContext, 0, len(contexts))
			for _, ac := range contexts {
				if ac.Context != context {
					filteredContexts = append(filteredContexts, ac)
				}
			}
			if len(filteredContexts) > 0 {
				a.contextsByIP.Store(oldIP, filteredContexts)
			} else {
				a.contextsByIP.Delete(oldIP)
			}
		}
	}
	
	a.contextToIP.Store(context, ip)
	
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
		a.contextsByIP.Store(ip, []ActionAndContext{{Action: action, Context: context}})
	} else {
		// Check if this context is already in the list to avoid duplicates
		found := false
		for _, ac := range contextIDs {
			if ac.Context == context {
				found = true
				break
			}
		}
		if !found {
			a.contextsByIP.Store(ip, append(contextIDs, ActionAndContext{Action: action, Context: context}))
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
	
	// Get the IP this context was using
	ip, ok := a.contextToIP.Load(contextID)
	if ok {
		a.decrementUsageAndCleanup(ctx, ip)
	}
	
	a.atemByContext.Delete(contextID)
	a.contextToIP.Delete(contextID)

	// Remove context from contextsByIP
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
	}
}

// UpdateContextIP handles when a context changes to a different IP
func (a *ConnectionManager) UpdateContextIP(ctx context.Context, contextID string, newIP string) {
	a.logger.Debug(ctx, "UpdateContextIP contextID:%s newIP:%s", contextID, newIP)
	
	// Get the old IP this context was using
	if oldIP, ok := a.contextToIP.Load(contextID); ok {
		if oldIP != newIP {
			a.logger.Debug(ctx, "UpdateContextIP context %s changing from %s to %s", contextID, oldIP, newIP)
			a.decrementUsageAndCleanup(ctx, oldIP)
		}
	}
}

// decrementUsageAndCleanup decrements usage count and schedules cleanup if needed
func (a *ConnectionManager) decrementUsageAndCleanup(ctx context.Context, ip string) {
	a.logger.Debug(ctx, "decrementUsageAndCleanup ip:%s", ip)
	
	if count, ok := a.usageCounts.Load(ip); ok {
		newCount := count - 1
		if newCount <= 0 {
			a.logger.Debug(ctx, "decrementUsageAndCleanup ip:%s usage count is zero, scheduling cleanup", ip)
			a.usageCounts.Delete(ip)
			
			// Use context.AfterFunc to schedule cleanup after a delay
			// This prevents race conditions when settings are changed rapidly
			context.AfterFunc(ctx, 5*time.Second, func() {
				a.cleanupUnusedATEM(ctx, ip)
			})
		} else {
			a.usageCounts.Store(ip, newCount)
			a.logger.Debug(ctx, "decrementUsageAndCleanup ip:%s usage count now %d", ip, newCount)
		}
	}
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
		}
	} else {
		a.logger.Debug(ctx, "cleanupUnusedATEM ip:%s is now in use again, skipping cleanup", ip)
	}
}
