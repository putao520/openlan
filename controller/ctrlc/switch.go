package ctrlc

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/controller/libctrl"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/vswitch/schema"
)

type Switch struct {
	libctrl.Listen
	cc *CtrlC
}

func (h *Switch) AddCtl(id string, m libctrl.Message) error {
	libol.Cmd("Switch.AddCtl %s %s", id, m.Data)
	p := schema.Switch{}
	if err := json.Unmarshal([]byte(m.Data), &p); err != nil {
		return err
	}
	_ = Storager.Switch.Set(p.Alias, &p)
	return nil
}

func (h *Switch) DelCtl(id string, m libctrl.Message) error {
	libol.Cmd("Switch.DelCtl %s %s", id, m.Data)
	Storager.Switch.Del(m.Data)
	return nil
}
