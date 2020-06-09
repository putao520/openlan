package storage

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
)

type _point struct {
	Clients  *libol.SafeStrMap
	UUIDAddr *libol.SafeStrStr
	AddrUUID *libol.SafeStrStr
	Listen   Listen
}

var Point = _point{
	Clients:  libol.NewSafeStrMap(1024),
	UUIDAddr: libol.NewSafeStrStr(1024),
	AddrUUID: libol.NewSafeStrStr(1024),
	Listen: Listen{
		listener: libol.NewSafeStrMap(32),
	},
}

func (p *_point) Init(size int) {
	p.Clients = libol.NewSafeStrMap(size)
}

func (p *_point) Add(m *models.Point) {
	_ = p.UUIDAddr.Reset(m.UUID, m.Client.Addr())
	_ = p.AddrUUID.Set(m.Client.Addr(), m.UUID)
	_ = p.Clients.Set(m.Client.Addr(), m)
	_ = p.Listen.AddV(m.Client.Addr(), m)
}

func (p *_point) Get(addr string) *models.Point {
	if v := p.Clients.Get(addr); v != nil {
		m := v.(*models.Point)
		m.Update()
		return m
	}
	return nil
}

func (p *_point) GetByUUID(uuid string) *models.Point {
	if addr := p.GetAddr(uuid); addr != "" {
		return p.Get(addr)
	}
	return nil
}

func (p *_point) GetUUID(addr string) string {
	return p.AddrUUID.Get(addr)
}

func (p *_point) GetAddr(uuid string) string {
	return p.UUIDAddr.Get(uuid)
}

func (p *_point) Del(addr string) {
	if v := p.Clients.Get(addr); v != nil {
		m := v.(*models.Point)
		if m.Device != nil {
			_ = m.Device.Close()
		}
		if p.UUIDAddr.Get(m.UUID) == addr { // not has newer
			p.UUIDAddr.Del(m.UUID)
		}
		p.AddrUUID.Del(m.Client.Addr())
		p.Clients.Del(addr)
	}
	p.Listen.DelV(addr)
}

func (p *_point) List() <-chan *models.Point {
	c := make(chan *models.Point, 128)

	go func() {
		p.Clients.Iter(func(k string, v interface{}) {
			if m, ok := v.(*models.Point); ok {
				m.Update()
				c <- m
			}
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}
