package ctl

import "github.com/danieldin95/openlan-go/libol"

type Hello struct {
	Listen
	cc *CtrlC
}

func (h *Hello) GetCtl(id string, m Message) error {
	libol.Cmd("Hello.GetCtl %s %s", id, m.Data)
	return nil
}
