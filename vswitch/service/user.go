package service

import (
	"bufio"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"os"
	"strings"
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

func (w *_user) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		values := strings.Split(line, ":")
		if len(values) == 2 {
			_user := models.NewUser(values[0], strings.TrimSpace(values[1]))
			w.Add(_user)
		}
	}
	return nil
}

func (w *_user) Add(user *models.User) {
	name := user.Name
	if name == "" {
		name = user.Token
	}
	w.users.Set(name, user)
}

func (w *_user) Del(name string) {
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
