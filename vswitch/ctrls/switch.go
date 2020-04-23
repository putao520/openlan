package ctrls

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/controller/libctrl"
	"github.com/danieldin95/openlan-go/vswitch/schema"
)

type Switch struct {
	libctrl.Listen
	cc *CtrlC
}

func (p *Switch) GetCtl(id string, m libctrl.Message) error {
	s := schema.Switch{
		Uptime: p.cc.Switcher.UpTime(),
		Alias:  p.cc.Switcher.Alias(),
		UUID:   p.cc.Switcher.UUID(),
	}
	if d, e := json.Marshal(s); e == nil {
		p.cc.Send(libctrl.Message{
			Action:   "add",
			Resource: "switch",
			Data:     string(d),
		})
	}
	return nil
}
