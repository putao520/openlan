package ctrls

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/controller/ctl"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/vswitch/schema"
	"github.com/danieldin95/openlan-go/vswitch/service"
)

type Point struct {
	cc *CtrlC
}

func (p *Point) Add(key string, value interface{}) {
	libol.Cmd("Point.Add %s", key)
	if value == nil {
		return
	}
	if obj, ok := value.(*models.Point); ok {
		data, _ := json.Marshal(schema.NewPoint(obj))
		if p.cc != nil {
			p.cc.Send(ctl.Message{
				Action:   "add",
				Resource: "point",
				Data:     string(data),
			})
		}
	}
}

func (p *Point) Del(key string) {
	libol.Cmd("Point.Del %s", key)
	if p.cc != nil {
		p.cc.Send(ctl.Message{
			Action:   "del",
			Resource: "point",
			Data:     key,
		})
	}
}

func (p *Point) GetCtl(id, data string) error {
	if data == "" {
		for u := range service.Point.List() {
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

func (p *Point) AddCtl(id, data string) error {
	panic("implement me")
}

func (p *Point) DelCtl(id, data string) error {
	panic("implement me")
}

func (p *Point) ModCtl(id, data string) error {
	panic("implement me")
}
