package ctl

import "github.com/danieldin95/openlan-go/libol"

type Link struct {
	cc *CtrlC
}

func (h *Link) GetCtl(id, data string) error {
	libol.Cmd("Link.GetCtl %s %s", id, data)
	return nil
}

func (h *Link) AddCtl(id, data string) error {
	libol.Cmd("Link.AddCtl %s %s", id, data)
	return nil
}

func (h *Link) DelCtl(id, data string) error {
	libol.Cmd("Link.DelCtl %s %s", id, data)
	return nil
}

func (h *Link) ModCtl(id, data string) error {
	libol.Cmd("Link.ModCtl %s %s", id, data)
	return nil
}
