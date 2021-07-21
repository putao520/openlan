package store

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
)

type _link struct {
	Links *libol.SafeStrMap
}

func (p *_link) Init(size int) {
	p.Links = libol.NewSafeStrMap(size)
}

func (p *_link) Add(uuid string, link *models.Link) {
	_ = p.Links.Set(uuid, link)
}

func (p *_link) Get(key string) *models.Link {
	ret := p.Links.Get(key)
	if ret != nil {
		return ret.(*models.Link)
	}
	return nil
}

func (p *_link) Del(key string) {
	p.Links.Del(key)
}

func (p *_link) List() <-chan *models.Link {
	c := make(chan *models.Link, 128)
	go func() {
		p.Links.Iter(func(k string, v interface{}) {
			m := v.(*models.Link)
			c <- m
		})
		c <- nil //Finish channel by nil.
	}()
	return c
}

var Link = _link{
	Links: libol.NewSafeStrMap(1024),
}
