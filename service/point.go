package service

import (
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/models"
	"sync"
)

type _point struct {
	lock    sync.RWMutex
	clients map[string]*models.Point
}

var Point = _point {
	clients: make(map[string]*models.Point, 1024),
}

func (p *_point) AddPoint(m *models.Point) {
	p.lock.Lock()
	defer p.lock.Unlock()

	Storage.SavePoint(m.Server, m, true)
	p.clients[m.Client.Addr] = m
}

func (p *_point) GetPoint(c *libol.TcpClient) *models.Point {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if m, ok := p.clients[c.Addr]; ok {
		return m
	}
	return nil
}

func (p *_point) DelPoint(c *libol.TcpClient) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if m, ok := p.clients[c.Addr]; ok {
		m.Device.Close()
		Storage.SavePoint(m.Server, m, false)
		delete(p.clients, c.Addr)
	}
}

func (p *_point) ListPoint() <-chan *models.Point {
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



