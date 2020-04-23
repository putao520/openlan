package ctrls

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/controller/ctl"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/vswitch/service"
)

type Neighbor struct {
	ctl.Listen
	cc *CtrlC
}

func (p *Neighbor) Add(key string, value interface{}) {
	libol.Cmd("Neighbor.Add %s %v", key, value)
	if value == nil {
		return
	}
	if n, ok := value.(*models.Neighbor); ok {
		data, _ := json.Marshal(models.NewNeighborSchema(n))
		if p.cc != nil {
			p.cc.Send(ctl.Message{
				Action:   "add",
				Resource: "neighbor",
				Data:     string(data),
			})
		}
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

func (p *Neighbor) GetCtl(id string, m ctl.Message) error {
	for u := range service.Neighbor.List() {
		if u == nil {
			break
		}
		p.Add(u.Client.Addr, u)
	}
	return nil
}
