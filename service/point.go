package service

import (
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/models"
	"sync"
)

type pointService struct {
	lock    sync.RWMutex
	clients map[string]*models.Point
}

var PointService = pointService {
	clients: make(map[string]*models.Point, 1024),
}

func (p *pointService) AddPoint(m *models.Point) {
	p.lock.Lock()
	defer p.lock.Unlock()

	StorageService.SavePoint("", m, true)
	p.clients[m.Client.Addr] = m
}

func (p *pointService) GetPoint(c *libol.TcpClient) *models.Point {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if m, ok := p.clients[c.Addr]; ok {
		return m
	}
	return nil
}

func (p *pointService) DelPoint(c *libol.TcpClient) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if m, ok := p.clients[c.Addr]; ok {
		m.Device.Close()
		StorageService.SavePoint("", m, false)
		delete(p.clients, c.Addr)
	}
}

func (p *pointService) ListPoint() <-chan *models.Point {
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



