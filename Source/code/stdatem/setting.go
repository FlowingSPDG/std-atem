package stdatem

import (
	"context"

	"github.com/FlowingSPDG/std-atem/Source/code/logger"
	"github.com/puzpuzpuz/xsync"
)

type atems struct {
	byIP      *xsync.MapOf[string, *ATEMInstance] // host: instance
	byContext *xsync.MapOf[string, *ATEMInstance] // context: binding
	logger    logger.Logger
}

func newAtems(logger logger.Logger) *atems {
	return &atems{
		byIP:      xsync.NewMapOf[*ATEMInstance](),
		byContext: xsync.NewMapOf[*ATEMInstance](),
		logger:    logger,
	}
}

func (a *atems) SolveATEMByIP(ctx context.Context, ip string) (*ATEMInstance, bool) {
	a.logger.Debug(ctx, "SolveATEMByIP ip:%s", ip)
	v, ok := a.byIP.Load(ip)
	if !ok {
		a.logger.Error(ctx, "SolveATEMByIP ip:%s not found", ip)
		return nil, false
	}

	return v, true
}
func (a *atems) SolveATEMByContext(ctx context.Context, context string) (*ATEMInstance, bool) {
	a.logger.Debug(ctx, "SolveATEMByContext context:%s", context)
	v, ok := a.byContext.Load(context)
	if !ok {
		a.logger.Error(ctx, "SolveATEMByContext context:%s not found", context)
		return nil, false
	}

	return v, true
}

func (a *atems) SolveContextByIP(ctx context.Context, ip string) (string, bool) {
	a.logger.Debug(ctx, "SolveContextByIP ip:%s", ip)
	// ipからStreamDeck contextを取得する
	var context string
	var ok bool

	a.byIP.Range(func(key string, value *ATEMInstance) bool {
		if value.client.Ip == ip {
			context = key
			ok = true
			return false
		}

		return false
	})

	if !ok {
		a.logger.Error(ctx, "SolveContextByIP ip:%s not found", ip)
	}

	return context, ok
}

func (a *atems) Store(ctx context.Context, ip, context string, setting *ATEMInstance) {
	a.logger.Debug(ctx, "Store ip:%s context:%s", ip, context)
	a.byIP.Store(ip, setting)
	a.byContext.Store(context, setting)
}

func (a *atems) Delete(ctx context.Context, ip, context string) {
	a.logger.Debug(ctx, "Delete ip:%s context:%s", ip, context)
	a.byIP.Delete(ip)
	a.byContext.Delete(context)

	// 該当のATEMInstanceを利用するcontextが無くなったら、ATEMInstanceを削除する
	context, ok := a.SolveContextByIP(ctx, ip)
	if ok {
		return
	}
	at, ok := a.SolveATEMByIP(ctx, ip)
	if !ok {
		return
	}

	// 削除処理
	a.logger.Debug(ctx, "Delete closing ATEM client ip:%s", ip)
	at.client.Close()
	a.byIP.Delete(ip)
	a.byContext.Delete(context)
}
