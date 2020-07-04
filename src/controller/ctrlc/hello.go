package ctrlc

import (
	"github.com/danieldin95/openlan-go/src/controller/libctrl"
	"github.com/danieldin95/openlan-go/src/libol"
)

type Hello struct {
	libctrl.Listen
	cc *CtrlC
}

func (h *Hello) GetCtl(id string, m libctrl.Message) error {
	libol.Cmd("Hello.GetCtl %s %s", id, m.Data)
	return nil
}
