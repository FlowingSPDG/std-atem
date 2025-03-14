package connectionmanager

import (
	"context"

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
	logger        logger.Logger
}

func NewConnectionManager(logger logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		atemByIP:      xsync.NewMapOf[*ATEMInstance](),
		atemByContext: xsync.NewMapOf[*ATEMInstance](),
		contextsByIP:  xsync.NewMapOf[[]ActionAndContext](),
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
	if contextIDs, ok := a.contextsByIP.Load(ip); !ok {
		a.contextsByIP.Store(ip, []ActionAndContext{{Action: action, Context: context}})
	} else {
		a.contextsByIP.Store(ip, append(contextIDs, ActionAndContext{Action: action, Context: context}))
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

func (a *ConnectionManager) DeleteATEMByContext(ctx context.Context, context string) {
	a.logger.Debug(ctx, "DeleteATEMByContext context:%s", context)
	a.atemByContext.Delete(context)

	// 該当のATEMInstanceを利用するcontextが無くなったら、ATEMInstanceを削除する
	at, ok := a.SolveATEMByContext(ctx, context)
	if !ok {
		return
	}
	contexts, ok := a.SolveContextsByIP(ctx, at.Client.Ip)
	if ok {
		if len(contexts) == 0 {
			a.logger.Debug(ctx, "Delete closing ATEM client ip:%s", at.Client.Ip)
			at, ok := a.SolveATEMByIP(ctx, at.Client.Ip)
			if !ok {
				return
			}
			at.Client.Close()
		}
	}
}
