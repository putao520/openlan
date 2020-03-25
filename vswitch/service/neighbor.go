package service

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
)

type _neighbor struct {
	neighbors *libol.SafeStrMap
}

var Neighbor = _neighbor{
	neighbors: libol.NewSafeStrMap(1024),
}

func (p *_neighbor) Init(size int) {
	p.neighbors = libol.NewSafeStrMap(size)
}

func (p *_neighbor) Add(m *models.Neighbor) {
	p.neighbors.Del(m.IpAddr.String())
	_ = p.neighbors.Set(m.IpAddr.String(), m)
}

func (p *_neighbor) Update(m *models.Neighbor) *models.Neighbor {
	if v := p.neighbors.Get(m.IpAddr.String()); v != nil {
		n := v.(*models.Neighbor)
		n.HwAddr = m.HwAddr
		n.HitTime = m.HitTime
	}
	return nil
}

func (p *_neighbor) Get(key string) *models.Neighbor {
	if v := p.neighbors.Get(key); v != nil {
		return v.(*models.Neighbor)
	}
	return nil
}

func (p *_neighbor) Del(key string) {
	p.neighbors.Del(key)
}

func (p *_neighbor) List() <-chan *models.Neighbor {
	c := make(chan *models.Neighbor, 128)

	go func() {
		p.neighbors.Iter(func(k string, v interface{}) {
			c <- v.(*models.Neighbor)
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}
