package stdatem

import (
	"github.com/puzpuzpuz/xsync"
)

type atems struct {
	m *xsync.MapOf[string, *ATEMInstance]
}

func (a *atems) Load(context string) (*ATEMInstance, bool) {
	v, ok := a.m.Load(context)
	if !ok {
		return nil, false
	}

	return v, true
}

func (a *atems) Store(context string, setting *ATEMInstance) {
	a.m.Store(context, setting)
}
