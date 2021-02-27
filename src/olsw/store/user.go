package store

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
)

type _user struct {
	File  string
	Users *libol.SafeStrMap
}

func (w *_user) Save() error {
	if w.File == "" {
		return nil
	}
	fp, err := libol.OpenTrunk(w.File)
	if err != nil {
		return err
	}
	for obj := range w.List() {
		if obj == nil {
			break
		}
		line := obj.Id() + ":" + obj.Password + ":" + obj.Role
		_, _ = fp.WriteString(line + "\n")
	}
	return nil
}

func (w *_user) SetFile(value string) {
	w.File = value
}

func (w *_user) Init(size int) {
	w.Users = libol.NewSafeStrMap(size)
}

func (w *_user) Add(user *models.User) {
	libol.Debug("_user.Add %v", user)
	key := user.Id()
	older := w.Get(key)
	if older == nil {
		_ = w.Users.Set(key, user)
	} else { // Update pass and role.
		older.Role = user.Role
		older.Password = user.Password
	}
}

func (w *_user) Del(key string) {
	libol.Debug("_user.Add %s", key)
	w.Users.Del(key)
}

func (w *_user) Get(key string) *models.User {
	if v := w.Users.Get(key); v != nil {
		return v.(*models.User)
	}
	return nil
}

func (w *_user) List() <-chan *models.User {
	c := make(chan *models.User, 128)

	go func() {
		w.Users.Iter(func(k string, v interface{}) {
			c <- v.(*models.User)
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}

var User = _user{
	Users: libol.NewSafeStrMap(1024),
}
