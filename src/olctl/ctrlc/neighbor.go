package ctrlc

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olctl/libctrl"
	"github.com/danieldin95/openlan-go/src/schema"
)

type Neighbor struct {
	libctrl.Listen
	cc *CtrlC
}

func (h *Neighbor) AddCtl(id string, m libctrl.Message) error {
	libol.Cmd("Neighbor.AddCtl %s %s", id, m.Data)
	p := schema.Neighbor{}
	if err := json.Unmarshal([]byte(m.Data), &p); err != nil {
		return err
	}
	if p.Switch == "" {
		p.Switch = id
	}
	_ = Storager.Neighbor.Mod(p.IpAddr, &p)
	return nil
}

func (h *Neighbor) DelCtl(id string, m libctrl.Message) error {
	libol.Cmd("Neighbor.DelCtl %s %s", id, m.Data)
	Storager.Neighbor.Del(m.Data)
	return nil
}
