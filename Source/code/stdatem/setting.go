package stdatem

import (
	"context"

	"github.com/FlowingSPDG/std-atem/Source/code/logger"
	"github.com/puzpuzpuz/xsync"
)

type actionAndContext struct {
	action  string
	context string
}

type atems struct {
	atemByIP      *xsync.MapOf[string, *ATEMInstance]      // host: instance
	atemByContext *xsync.MapOf[string, *ATEMInstance]      // context: binding
	contextsByIP  *xsync.MapOf[string, []actionAndContext] // host: contexts
	logger        logger.Logger
}

func newAtems(logger logger.Logger) *atems {
	return &atems{
		atemByIP:      xsync.NewMapOf[*ATEMInstance](),
		atemByContext: xsync.NewMapOf[*ATEMInstance](),
		contextsByIP:  xsync.NewMapOf[[]actionAndContext](),
		logger:        logger,
	}
}

func (a *atems) SolveATEMByIP(ctx context.Context, ip string) (*ATEMInstance, bool) {
	a.logger.Debug(ctx, "SolveATEMByIP ip:%s", ip)
	v, ok := a.atemByIP.Load(ip)
	if !ok {
		a.logger.Error(ctx, "SolveATEMByIP ip:%s not found", ip)
		return nil, false
	}

	return v, true
}
func (a *atems) SolveATEMByContext(ctx context.Context, context string) (*ATEMInstance, bool) {
	a.logger.Debug(ctx, "SolveATEMByContext context:%s", context)
	v, ok := a.atemByContext.Load(context)
	if !ok {
		a.logger.Error(ctx, "SolveATEMByContext context:%s not found", context)
		return nil, false
	}

	return v, true
}

func (a *atems) SolveContextsByIP(ctx context.Context, ip string) ([]actionAndContext, bool) {
	a.logger.Debug(ctx, "SolveContextsByIP ip:%s", ip)
	// ipからStreamDeck contextを取得する
	var contexts []actionAndContext
	var ok bool

	a.contextsByIP.Range(func(key string, value []actionAndContext) bool {
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

func (a *atems) Store(ctx context.Context, action, ip, context string, setting *ATEMInstance) {
	a.logger.Debug(ctx, "Store action:%s ip:%s context:%s", action, ip, context)
	a.atemByIP.Store(ip, setting)
	a.atemByContext.Store(context, setting)
	if contextIDs, ok := a.contextsByIP.Load(ip); !ok {
		a.contextsByIP.Store(ip, []actionAndContext{{action: action, context: context}})
	} else {
		a.contextsByIP.Store(ip, append(contextIDs, actionAndContext{action: action, context: context}))
	}
}

func (a *atems) DeleteATEMByIP(ctx context.Context, ip string) {
	a.logger.Debug(ctx, "DeleteATEMByIP ip:%s", ip)
	a.atemByIP.Delete(ip)

	// 削除処理
	a.logger.Debug(ctx, "Delete closing ATEM client ip:%s", ip)
	at, ok := a.SolveATEMByIP(ctx, ip)
	if !ok {
		return
	}
	at.client.Close()
}

func (a *atems) DeleteATEMByContext(ctx context.Context, context string) {
	a.logger.Debug(ctx, "DeleteATEMByContext context:%s", context)
	a.atemByContext.Delete(context)

	// 該当のATEMInstanceを利用するcontextが無くなったら、ATEMInstanceを削除する
	at, ok := a.SolveATEMByContext(ctx, context)
	if !ok {
		return
	}
	contexts, ok := a.SolveContextsByIP(ctx, at.client.Ip)
	if ok {
		if len(contexts) == 0 {
			a.logger.Debug(ctx, "Delete closing ATEM client ip:%s", at.client.Ip)
			at, ok := a.SolveATEMByIP(ctx, at.client.Ip)
			if !ok {
				return
			}
			at.client.Close()
		}
	}
}
