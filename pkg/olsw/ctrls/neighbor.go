package ctrls

import (
	"encoding/json"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/danieldin95/openlan/pkg/models"
	"github.com/danieldin95/openlan/pkg/olctl/libctrl"
	"github.com/danieldin95/openlan/pkg/olsw/store"
)

type Neighbor struct {
	libctrl.Listen
	cc *CtrlC
}

func (p *Neighbor) Add(key string, value interface{}) {
	libol.Cmd("Neighbor.Add %s %v", key, value)
	if value == nil {
		return
	}
	if n, ok := value.(*models.Neighbor); ok {
		if d, e := json.Marshal(models.NewNeighborSchema(n)); e == nil {
			p.cc.Send(libctrl.Message{
				Action:   "add",
				Resource: "neighbor",
				Data:     string(d),
			})
		}
	}
}

func (p *Neighbor) Del(key string) {
	libol.Cmd("Neighbor.Del %s", key)
	p.cc.Send(libctrl.Message{
		Action:   "delete",
		Resource: "neighbor",
		Data:     key,
	})
}

func (p *Neighbor) GetCtl(id string, m libctrl.Message) error {
	for u := range store.Neighbor.List() {
		if u == nil {
			break
		}
		p.Add(u.Client, u)
	}
	return nil
}
