package ctrlc

import (
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/danieldin95/openlan/pkg/olctl/libctrl"
)

type Link struct {
	libctrl.Listen
	cc *CtrlC
}

func (h *Link) AddCtl(id string, m libctrl.Message) error {
	libol.Cmd("Link.AddCtl %s %s", id, m.Data)
	return nil
}

func (h *Link) DelCtl(id string, m libctrl.Message) error {
	libol.Cmd("Link.DelCtl %s %s", id, m.Data)
	return nil
}
