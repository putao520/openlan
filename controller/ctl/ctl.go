package ctl

import (
	"github.com/danieldin95/openlan-go/controller/storage"
	"sync"
)

type Ctl struct {
	Lock    sync.RWMutex
	VSwitch map[string]*VSwitch
}

func (c *Ctl) Load(srv *storage.Storage) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	for v := range srv.VSwitch.List() {
		if v == nil {
			break
		}
		vs := &VSwitch{
			Service: *v,
		}
		vs.Init()
		vs.Start()
		v.Ctl = vs
		c.VSwitch[v.Name] = vs
	}
}

var CTL = Ctl{
	VSwitch: make(map[string]*VSwitch, 32),
}
