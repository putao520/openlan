package ctl

import "github.com/danieldin95/openlan-go/libol"

type Hello struct {
	cc *CtrlC
}

func (h *Hello) GetCtl(id, data string) error {
	libol.Cmd("Hello.GetCtl %s %s", id, data)
	return nil
}

func (h *Hello) AddCtl(id, data string) error {
	libol.Cmd("Hello.AddCtl %s %s", id, data)
	return nil
}

func (h *Hello) DelCtl(id, data string) error {
	libol.Cmd("Hello.DelCtl %s %s", id, data)
	return nil
}

func (h *Hello) ModCtl(id, data string) error {
	libol.Cmd("Hello.ModCtl %s %s", id, data)
	return nil
}
