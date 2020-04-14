package service

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/point"
)

type _link struct {
	links *libol.SafeStrMap
}

var Link = _link{
	links: libol.NewSafeStrMap(1024),
}

func (p *_link) Init(size int) {
	p.links = libol.NewSafeStrMap(size)
}

func (p *_link) Add(m *point.Point) {
	link := &models.Point{
		Alias:   "",
		Network: m.Tenant,
		Server:  m.Addr(),
		Uptime:  m.UpTime(),
		Status:  m.State(),
		Client:  m.Client(),
		Device:  m.Device(),
		IfName:  m.IfName(),
		UUID:    m.UUID(),
	}
	_ = p.links.Set(m.Addr(), link)
}

func (p *_link) Get(key string) *models.Point {
	ret := p.links.Get(key)
	if ret != nil {
		v := ret.(*models.Point)
		v.Update()
		return v
	}
	return nil
}

func (p *_link) Del(key string) {
	p.links.Del(key)
}

func (p *_link) List() <-chan *models.Point {
	c := make(chan *models.Point, 128)
	go func() {
		p.links.Iter(func(k string, v interface{}) {
			m := v.(*models.Point)
			m.Update()
			c <- m
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}
