package service

import (
	"github.com/danieldin95/openlan-go/libol"
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

func (p *_point) Get(c *libol.TcpClient) *models.Point {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if m, ok := p.clients[c.Addr]; ok {
		return m
	}
	return nil
}

func (p *_point) Del(c *libol.TcpClient) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if m, ok := p.clients[c.Addr]; ok {
		m.Device.Close()
		delete(p.clients, c.Addr)
	}
}

func (p *_point) List() <-chan *models.Point {
	c := make(chan *models.Point, 128)

	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()

		for _, m := range p.clients {
			c <- m
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}
