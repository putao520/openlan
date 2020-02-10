package service

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
)

type _point struct {
	clients *libol.SafeStrMap
}

var Point = _point{
	clients: libol.NewSafeStrMap(1024),
}

func (p *_point) Init(size int) {
	p.clients = libol.NewSafeStrMap(size)
}

func (p *_point) Add(m *models.Point) {
	p.clients.Set(m.Client.Addr, m)
}

func (p *_point) Get(addr string) *models.Point {
	if v := p.clients.Get(addr); v != nil {
		m := v.(*models.Point)
		m.Update()
		return m
	}
	return nil
}

func (p *_point) Del(addr string) {
	if v := p.clients.Get(addr); v != nil {
		m := v.(*models.Point)
		m.Device.Close()
		p.clients.Del(addr)
	}
}

func (p *_point) List() <-chan *models.Point {
	c := make(chan *models.Point, 128)

	go func() {
		p.clients.Iter(func(k string, v interface{}) {
			m := v.(*models.Point)
			m.Update()
			c <- m
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}
