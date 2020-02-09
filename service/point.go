package service

import (
	"github.com/danieldin95/openlan-go/models"
	"sync"
)

type _point struct {
	lock    sync.RWMutex
	clients map[string]*models.Point
}

var Point = _point{
	clients: make(map[string]*models.Point, 1024),
}

func (p *_point) Add(m *models.Point) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.clients[m.Client.Addr] = m
}

func (p *_point) Get(addr string) *models.Point {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if m, ok := p.clients[addr]; ok {
		m.Update()
		return m
	}
	return nil
}

func (p *_point) Del(addr string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if m, ok := p.clients[addr]; ok {
		m.Device.Close()
		delete(p.clients, addr)
	}
}

func (p *_point) List() <-chan *models.Point {
	c := make(chan *models.Point, 128)

	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()

		for _, m := range p.clients {
			m.Update()
			c <- m
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}
