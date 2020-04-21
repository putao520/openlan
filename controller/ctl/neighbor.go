package ctl

import "github.com/danieldin95/openlan-go/libol"

type Neighbor struct {
	cc *CtrlC
}

func (h *Neighbor) GetCtl(id, data string) error {
	libol.Cmd("Neighbor.GetCtl %s %s", id, data)
	return nil
}

func (h *Neighbor) AddCtl(id, data string) error {
	libol.Cmd("Neighbor.AddCtl %s %s", id, data)
	return nil
}

func (h *Neighbor) DelCtl(id, data string) error {
	libol.Cmd("Neighbor.DelCtl %s %s", id, data)
	return nil
}

func (h *Neighbor) ModCtl(id, data string) error {
	libol.Cmd("Neighbor.ModCtl %s %s", id, data)
	return nil
}
