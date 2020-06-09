package storage

import "github.com/danieldin95/openlan-go/libol"

type Listener interface {
	Add(key string, value interface{})
	Del(key string)
}

type Listen struct {
	listener *libol.SafeStrMap
}

func (l *Listen) Add(name string, listen Listener) error {
	return l.listener.Set(name, listen)
}

func (l *Listen) Del(name string) {
	l.listener.Del(name)
}

func (l *Listen) AddV(key string, m interface{}) error {
	l.listener.Iter(func(k string, v interface{}) {
		if f, ok := v.(Listener); ok {
			f.Add(key, m)
		}
	})
	return nil
}

func (l *Listen) DelV(key string) {
	l.listener.Iter(func(k string, v interface{}) {
		if f, ok := v.(Listener); ok {
			f.Del(key)
		}
	})
}
