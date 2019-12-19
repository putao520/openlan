package service

import (
	"bufio"
	"github.com/lightstar-dev/openlan-go/models"
	"os"
	"strings"
	"sync"
)

type _user struct {
	lock sync.RWMutex
	_users     map[string]*models.User
}

var User = _user {
	_users: make(map[string]*models.User, 1024),
}

func (w *_user) LoadUsers(path string) error {
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
			w.AddUser(_user)
		}
	}

	return nil
}

func (w *_user) AddUser(_user *models.User) {
	w.lock.Lock()
	defer w.lock.Unlock()

	name := _user.Name
	if name == "" {
		name = _user.Token
	}
	w._users[name] = _user
}

func (w *_user) DelUser(name string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if _, ok := w._users[name]; ok {
		delete(w._users, name)
	}
}

func (w *_user) GetUser(name string) *models.User {
	w.lock.RLock()
	defer w.lock.RUnlock()

	if u, ok := w._users[name]; ok {
		return u
	}

	return nil
}

func (w *_user) ListUser() <-chan *models.User {
	c := make(chan *models.User, 128)

	go func() {
		w.lock.RLock()
		defer w.lock.RUnlock()

		for _, u := range w._users {
			c <- u
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}
