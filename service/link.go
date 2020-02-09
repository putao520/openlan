package service

import (
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/point"
	"sync"
)

type _link struct {
	lock  sync.RWMutex
	links map[string]*models.Point
}

var Link = _link{
	links: make(map[string]*models.Point, 1024),
}

func (p *_link) Add(m *point.Point) {
	p.lock.Lock()
	defer p.lock.Unlock()

	link := &models.Point{
		Alias:  "",
		Server: m.Addr(),
		Uptime: m.UpTime(),
		Status: m.State(),
		Client: m.Client(),
		Device: m.Device(),
		IfName: m.IfName(),
		UUID:   m.UUID(),
	}
	p.links[m.Addr()] = link
}

func (p *_link) Get(key string) *models.Point {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if m, ok := p.links[key]; ok {
		m.Update()
		return m
	}
	return nil
}

func (p *_link) Del(key string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.links[key]; ok {
		delete(p.links, key)
	}
}

func (p *_link) List() <-chan *models.Point {
	c := make(chan *models.Point, 128)

	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()

		for _, m := range p.links {
			m.Update()
			c <- m
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}
