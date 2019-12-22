package service

import (
	"github.com/danieldin95/openlan-go/models"
	"sync"
)

type _online struct {
	lock  sync.RWMutex
	lines map[string]*models.Line
}

var Online = _online{
	lines: make(map[string]*models.Line, 1024),
}

func (p *_online) Add(m *models.Line) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.lines[m.String()] = m
}

func (p *_online) Get(key string) *models.Line {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if m, ok := p.lines[key]; ok {
		return m
	}
	return nil
}

func (p *_online) Del(key string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.lines[key]; ok {
		delete(p.lines, key)
	}
}

func (p *_online) List() <-chan *models.Line {
	c := make(chan *models.Line, 128)

	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()

		for _, m := range p.lines {
			c <- m
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}
