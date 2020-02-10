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
	p.neighbors.Set(m.IpAddr.String(), m)
}

func (p *_neighbor) Get(key string) *models.Neighbor {
	v := p.neighbors.Get(key)
	if v != nil {
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
