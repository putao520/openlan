package storage

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
)

type _online struct {
	Lines *libol.SafeStrMap
}

func (p *_online) Init(size int) {
	p.Lines = libol.NewSafeStrMap(size)
}

func (p *_online) Add(m *models.Line) {
	_ = p.Lines.Set(m.String(), m)
}

func (p *_online) Update(m *models.Line) *models.Line {
	if v := p.Lines.Get(m.String()); v != nil {
		l := v.(*models.Line)
		l.HitTime = m.HitTime
	}
	return nil
}

func (p *_online) Get(key string) *models.Line {
	if v := p.Lines.Get(key); v != nil {
		return v.(*models.Line)
	}
	return nil
}

func (p *_online) Del(key string) {
	p.Lines.Del(key)
}

func (p *_online) List() <-chan *models.Line {
	c := make(chan *models.Line, 128)

	go func() {

		p.Lines.Iter(func(k string, v interface{}) {
			c <- v.(*models.Line)
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}

var Online = _online{
	Lines: libol.NewSafeStrMap(1024),
}
