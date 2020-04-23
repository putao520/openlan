package ctl

import "github.com/danieldin95/openlan-go/libol"

type Neighbor struct {
	Listen
	cc *CtrlC
}


func (h *Neighbor) AddCtl(id string, m Message) error {
	libol.Cmd("Neighbor.AddCtl %s %s", id, m.Data)
	return nil
}

func (h *Neighbor) DelCtl(id string, m Message) error {
	libol.Cmd("Neighbor.DelCtl %s %s", id, m.Data)
	return nil
}
