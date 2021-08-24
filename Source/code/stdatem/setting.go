package stdatem

import (
	"fmt"
	"sync"
)

// SD button setting
type buttonSetting struct {
	input   uint16
	meIndex uint8
}

type buttons struct {
	m *sync.Map
}

func (b *buttons) Load(context string) (*buttonSetting, error) {
	v, ok := b.m.Load(context)
	if !ok {
		return nil, fmt.Errorf("Setting not found for this context")
	}

	return (v).(*buttonSetting), nil
}

func (b *buttons) Store(context string, setting *buttonSetting) {
	b.m.Store(context, setting)
}
