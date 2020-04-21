package ctl

import "github.com/danieldin95/openlan-go/libol"

type Point struct {
	cc *CtrlC
}

func (h *Point) GetCtl(id, data string) error {
	libol.Cmd("Point.GetCtl %s %s", id, data)
	return nil
}

func (h *Point) AddCtl(id, data string) error {
	libol.Cmd("Point.AddCtl %s %s", id, data)
	return nil
}

func (h *Point) DelCtl(id, data string) error {
	libol.Cmd("Point.DelCtl %s %s", id, data)
	return nil
}

func (h *Point) ModCtl(id, data string) error {
	libol.Cmd("Point.ModCtl %s %s", id, data)
	return nil
}
