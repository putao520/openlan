package libctrl

import "github.com/danieldin95/openlan/pkg/libol"

type Listener interface {
	GetCtl(id string, m Message) error
	AddCtl(id string, m Message) error
	DelCtl(id string, m Message) error
	ModCtl(id string, m Message) error
}

type Listen struct {
}

func (l *Listen) GetCtl(id string, m Message) error {
	libol.Cmd("Listen %s %s %s", id, m.Action, m.Resource)
	return nil
}

func (l *Listen) AddCtl(id string, m Message) error {
	libol.Cmd("Listen %s %s %s", id, m.Action, m.Resource)
	return nil
}

func (l *Listen) DelCtl(id string, m Message) error {
	libol.Cmd("Listen %s %s %s", id, m.Action, m.Resource)
	return nil
}

func (l *Listen) ModCtl(id string, m Message) error {
	libol.Cmd("Listen %s %s %s", id, m.Action, m.Resource)
	return nil
}
