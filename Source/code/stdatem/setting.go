package stdatem

import (
	"github.com/puzpuzpuz/xsync"
)

type atems struct {
	byIP      *xsync.MapOf[string, *ATEMInstance] // host: instance
	byContext *xsync.MapOf[string, *ATEMInstance] // context: binding
}

func newAtems() *atems {
	return &atems{
		byIP:      xsync.NewMapOf[*ATEMInstance](),
		byContext: xsync.NewMapOf[*ATEMInstance](),
	}
}

func (a *atems) SolveATEMByIP(ip string) (*ATEMInstance, bool) {
	v, ok := a.byIP.Load(ip)
	if !ok {
		return nil, false
	}

	return v, true
}

func (a *atems) SolveATEMByContext(context string) (*ATEMInstance, bool) {
	v, ok := a.byContext.Load(context)
	if !ok {
		return nil, false
	}

	return v, true
}

func (a *atems) SolveContextByIP(ip string) (string, bool) {
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

	return context, ok
}

func (a *atems) Store(ip, context string, setting *ATEMInstance) {
	a.byIP.Store(ip, setting)
	a.byContext.Store(context, setting)
}

func (a *atems) Delete(ip, context string) {
	a.byIP.Delete(ip)
	a.byContext.Delete(context)

	// 該当のATEMInstanceを利用するcontextが無くなったら、ATEMInstanceを削除する
	context, ok := a.SolveContextByIP(ip)
	if ok {
		return
	}
	at, ok := a.SolveATEMByIP(ip)
	if !ok {
		return
	}

	// 削除処理
	at.client.Close()
	a.byIP.Delete(ip)
	a.byContext.Delete(context)
}
