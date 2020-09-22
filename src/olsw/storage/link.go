package storage

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/olap"
)

type link struct {
	Links *libol.SafeStrMap
}

var Link = link{
	Links: libol.NewSafeStrMap(1024),
}

func (p *link) Init(size int) {
	p.Links = libol.NewSafeStrMap(size)
}

func (p *link) Add(m *olap.Point) {
	link := &models.Point{
		Alias:   "",
		User:    m.User(),
		Network: m.Tenant(),
		Server:  m.Addr(),
		Uptime:  m.UpTime(),
		Status:  m.Status().String(),
		Client:  m.Client(),
		Device:  m.Device(),
		IfName:  m.IfName(),
		UUID:    m.UUID(),
	}
	_ = p.Links.Set(m.UUID(), link)
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
