package service

import (
	"github.com/danieldin95/openlan-go/models"
	"sync"
)

type _neighbor struct {
	lock      sync.RWMutex
	neighbors map[string]*models.Neighbor
}

var Neighbor = _neighbor{
	neighbors: make(map[string]*models.Neighbor, 1024),
}

func (p *_neighbor) Add(m *models.Neighbor) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.neighbors[m.IpAddr.String()] = m
}

func (p *_neighbor) Get(key string) *models.Neighbor {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if m, ok := p.neighbors[key]; ok {
		return m
	}
	return nil
}

func (p *_neighbor) Del(key string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.neighbors[key]; ok {
		delete(p.neighbors, key)
	}
}

func (p *_neighbor) List() <-chan *models.Neighbor {
	c := make(chan *models.Neighbor, 128)

	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()

		for _, m := range p.neighbors {
			c <- m
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}
