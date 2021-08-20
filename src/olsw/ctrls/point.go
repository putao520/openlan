package ctrls

import (
	"encoding/json"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/models"
	"github.com/danieldin95/openlan/src/olctl/libctrl"
	"github.com/danieldin95/openlan/src/olsw/store"
)

type Point struct {
	libctrl.Listen
	cc *CtrlC
}

func (p *Point) Add(key string, value interface{}) {
	libol.Cmd("Point.Add %s", key)
	if value == nil {
		return
	}
	if obj, ok := value.(*models.Point); ok {
		if d, e := json.Marshal(models.NewPointSchema(obj)); e == nil {
			p.cc.Send(libctrl.Message{
				Action:   "add",
				Resource: "point",
				Data:     string(d),
			})
		}
	}
}

func (p *Point) Del(key string) {
	libol.Cmd("Point.Del %s", key)
	p.cc.Send(libctrl.Message{
		Action:   "del",
		Resource: "point",
		Data:     key,
	})
}

func (p *Point) GetCtl(id string, m libctrl.Message) error {
	if m.Data == "" {
		for u := range store.Point.List() {
			if u == nil {
				break
			}
			p.Add(u.Client.String(), u)
		}
	} else {
		// TODO reply one POINT.
	}
	return nil
}
