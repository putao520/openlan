package ctrls

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/controller/ctl"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/vswitch/service"
)

type Neighbor struct {
	cc *CtrlC
}

func (p *Neighbor) Add(key string, value interface{}) {
	libol.Cmd("Neighbor.Add %s", key)
	data, _ := json.Marshal(value)
	if p.cc != nil {
		p.cc.Send(ctl.Message{
			Action:   "add",
			Resource: "neighbor",
			Data:     string(data),
		})
	}
}

func (p *Neighbor) Del(key string) {
	libol.Cmd("Neighbor.Del %s", key)
	if p.cc != nil {
		p.cc.Send(ctl.Message{
			Action:   "delete",
			Resource: "neighbor",
			Data:     key,
		})
	}
}

func (p *Neighbor) GetCtl(id, data string) error {
	if data == "" {
		for u := range service.Neighbor.List() {
			if u == nil {
				break
			}
			p.Add(u.Client.Addr, u)
		}
	} else {
		// TODO reply one POINT.
	}
	return nil
}

func (p *Neighbor) AddCtl(id, data string) error {
	panic("implement me")
}

func (p *Neighbor) DelCtl(id, data string) error {
	panic("implement me")
}

func (p *Neighbor) ModCtl(id, data string) error {
	panic("implement me")
}
