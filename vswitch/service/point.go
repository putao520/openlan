package service

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
)

type _point struct {
	Clients *libol.SafeStrMap
	Listen  Listen
}

var Point = _point{
	Clients: libol.NewSafeStrMap(1024),
	Listen: Listen{
		listener: libol.NewSafeStrMap(32),
	},
}

func (p *_point) Init(size int) {
	p.Clients = libol.NewSafeStrMap(size)
}

func (p *_point) Add(m *models.Point) {
	_ = p.Clients.Set(m.Client.Addr, m)
	_ = p.Listen.AddV(m.Client.Addr, m)
}

func (p *_point) Get(addr string) *models.Point {
	if v := p.Clients.Get(addr); v != nil {
		m := v.(*models.Point)
		m.Update()
		return m
	}
	return nil
}

func (p *_point) Del(addr string) {
	if v := p.Clients.Get(addr); v != nil {
		m := v.(*models.Point)
		if m.Device != nil {
			_ = m.Device.Close()
		}
		p.Clients.Del(addr)
	}
	p.Listen.DelV(addr)
}

func (p *_point) List() <-chan *models.Point {
	c := make(chan *models.Point, 128)

	go func() {
		p.Clients.Iter(func(k string, v interface{}) {
			m := v.(*models.Point)
			m.Update()
			c <- m
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}
