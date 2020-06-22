package storage

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	pp "github.com/danieldin95/openlan-go/point"
)

type link struct {
	Links  *libol.SafeStrMap
	Listen Listen
}

var Link = link{
	Links: libol.NewSafeStrMap(1024),
	Listen: Listen{
		listener: libol.NewSafeStrMap(32),
	},
}

func (p *link) Init(size int) {
	p.Links = libol.NewSafeStrMap(size)
}

func (p *link) Add(m *pp.Point) {
	link := &models.Point{
		Alias:   "",
		Network: m.Tenant(),
		Server:  m.Addr(),
		Uptime:  m.UpTime(),
		Status:  m.State(),
		Client:  m.Client(),
		Device:  m.Device(),
		IfName:  m.IfName(),
		UUID:    m.UUID(),
	}
	_ = p.Links.Set(m.Addr(), link)
	_ = p.Listen.AddV(m.Addr(), link)
}

func (p *link) Get(key string) *models.Point {
	ret := p.Links.Get(key)
	if ret != nil {
		v := ret.(*models.Point)
		v.Update()
		return v
	}
	return nil
}

func (p *link) Del(key string) {
	p.Links.Del(key)
	p.Listen.DelV(key)
}

func (p *link) List() <-chan *models.Point {
	c := make(chan *models.Point, 128)
	go func() {
		p.Links.Iter(func(k string, v interface{}) {
			m := v.(*models.Point)
			m.Update()
			c <- m
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}
