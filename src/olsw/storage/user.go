package storage

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
)

type user struct {
	Users *libol.SafeStrMap
}

var User = user{
	Users: libol.NewSafeStrMap(1024),
}

func (w *user) Init(size int) {
	w.Users = libol.NewSafeStrMap(size)
}

func (w *user) Add(user *models.User) {
	libol.Debug("user.Add %v", user)
	key := user.Id()
	w.Users.Del(key)
	_ = w.Users.Set(key, user)
}

func (w *user) Del(key string) {
	libol.Debug("user.Add %s", key)
	w.Users.Del(key)
}

func (w *user) Get(key string) *models.User {
	if v := w.Users.Get(key); v != nil {
		return v.(*models.User)
	}
	return nil
}

func (w *user) List() <-chan *models.User {
	c := make(chan *models.User, 128)

	go func() {
		w.Users.Iter(func(k string, v interface{}) {
			c <- v.(*models.User)
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}
