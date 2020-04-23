package ctl

import "github.com/danieldin95/openlan-go/libol"

type Link struct {
	Listen
	cc *CtrlC
}

func (h *Link) AddCtl(id string, m Message) error {
	libol.Cmd("Link.AddCtl %s %s", id, m.Data)
	return nil
}

func (h *Link) DelCtl(id string, m Message) error {
	libol.Cmd("Link.DelCtl %s %s", id, m.Data)
	return nil
}
