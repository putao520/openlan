package service

import (
	"github.com/danieldin95/openlan-go/point"
	"sync"
)

type _link struct {
	lock  sync.RWMutex
	links map[string]*point.Point
}

var Link = _link{
	links: make(map[string]*point.Point, 1024),
}

func (p *_link) Add(m *point.Point) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.links[m.Addr()] = m
}

func (p *_link) Get(key string) *point.Point {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if m, ok := p.links[key]; ok {
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

func (p *_link) List() <-chan *point.Point {
	c := make(chan *point.Point, 128)

	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()

		for _, m := range p.links {
			c <- m
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}
