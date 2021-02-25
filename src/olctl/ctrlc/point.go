package ctrlc

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olctl/libctrl"
	"github.com/danieldin95/openlan-go/src/schema"
)

type Point struct {
	libctrl.Listen
	cc *CtrlC
}

func (h *Point) AddCtl(id string, m libctrl.Message) error {
	libol.Cmd("Point.AddCtl %s %s", id, m.Data)
	p := schema.Point{}
	if err := json.Unmarshal([]byte(m.Data), &p); err != nil {
		return err
	}
	if p.Switch == "" {
		p.Switch = id
	}
	_ = Storager.Point.Mod(p.Address, &p)
	return nil
}

func (h *Point) DelCtl(id string, m libctrl.Message) error {
	libol.Cmd("Point.DelCtl %s %s", id, m.Data)
	Storager.Point.Del(m.Data)
	return nil
}
