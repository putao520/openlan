package storage

import (
	"github.com/danieldin95/lightstar/libstar"
	"github.com/danieldin95/openlan-go/controller/schema"
	"github.com/danieldin95/openlan-go/libol"
	"sync"
)

type VSwitch struct {
	Lock    sync.RWMutex
	File    string
	VSwitch map[string]*schema.VSwitch `json:"vswitch"`
}

func (v *VSwitch) Save() error {
	v.Lock.RLock()
	defer v.Lock.RUnlock()

	if err := libstar.JSON.MarshalSave(&v.VSwitch, v.File, true); err != nil {
		return err
	}
	return nil
}

func (v *VSwitch) Load(file string) error {
	v.Lock.Lock()
	defer v.Lock.Unlock()

	v.File = file
	if err := libstar.JSON.UnmarshalLoad(&v.VSwitch, file); err != nil {
		return err
	}
	for name, value := range v.VSwitch {
		if value == nil {
			continue
		}
		if value.Name == "" {
			value.Name = name
		}
		if value.Token == "" {
			value.Token = libol.GenToken(64)
		}
		Storager.Users.Add(&schema.User{
			Name:     value.Name,
			Password: value.Token,
		})
		value.Init()
	}
	return nil
}

func (v *VSwitch) Get(name string) (schema.VSwitch, bool) {
	v.Lock.RLock()
	defer v.Lock.RUnlock()

	data, ok := v.VSwitch[name]
	if data == nil {
		return schema.VSwitch{}, false
	}
	return *data, ok
}

func (v *VSwitch) List() <-chan *schema.VSwitch {
	c := make(chan *schema.VSwitch, 128)
	go func() {
		v.Lock.RLock()
		defer v.Lock.RUnlock()

		for _, d := range v.VSwitch {
			c <- d
		}
		c <- nil // Finish channel by nil.
	}()
	return c
}
