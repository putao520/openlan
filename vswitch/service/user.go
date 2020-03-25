package service

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
)

type _user struct {
	users *libol.SafeStrMap
}

var User = _user{
	users: libol.NewSafeStrMap(1024),
}

func (w *_user) Init(size int) {
	w.users = libol.NewSafeStrMap(size)
}

func (w *_user) Load(tenant, path string) error {
	users := make([]*models.User, 32)
	if err := libol.UnmarshalLoad(&users, path); err != nil {
		libol.Error("_user.load: %s", err)
		return err
	}
	for _, user := range users {
		user.Name = user.Name + "@" + tenant
		w.Add(user)
	}
	return nil
}

func (w *_user) Add(user *models.User) {
	libol.Debug("_user.Add %v", *user)
	name := user.Name
	if name == "" {
		name = user.Token
	}
	w.users.Del(name)
	_ = w.users.Set(name, user)
}

func (w *_user) Del(name string) {
	libol.Debug("_user.Add %s", name)
	w.users.Del(name)
}

func (w *_user) Get(name string) *models.User {
	if v := w.users.Get(name); v != nil {
		return v.(*models.User)
	}
	return nil
}

func (w *_user) List() <-chan *models.User {
	c := make(chan *models.User, 128)

	go func() {
		w.users.Iter(func(k string, v interface{}) {
			c <- v.(*models.User)
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}
