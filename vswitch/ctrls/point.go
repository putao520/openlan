package ctrls

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/controller/ctl"
	"github.com/danieldin95/openlan-go/libol"
)

type Point struct {
	cc *CtrlC
}

func (p *Point) Add(key string, value interface{}) {
	libol.Cmd("Point.Add %s", key)
	data, _ := json.Marshal(value)
	if p.cc != nil {
		p.cc.Send(ctl.Message{
			Action:   "ADD",
			Resource: "POINT",
			Data:     string(data),
		})
	}
}

func (p *Point) Del(key string) {
	libol.Cmd("Point.Del %s", key)
	if p.cc != nil {
		p.cc.Send(ctl.Message{
			Action:   "DELETE",
			Resource: "POINT",
			Data:     key,
		})
	}
}
