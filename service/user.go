package service

import (
	"bufio"
	"github.com/lightstar-dev/openlan-go/models"
	"os"
	"strings"
	"sync"
)

type userService struct {
	lock sync.RWMutex
	users     map[string]*models.User
}

var UserService = userService {
	users: make(map[string]*models.User, 1024),
}

func (w *userService) LoadUsers(path string) error {
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
			user := models.NewUser(values[0], strings.TrimSpace(values[1]))
			w.AddUser(user)
		}
	}

	return nil
}

func (w *userService) AddUser(user *models.User) {
	w.lock.Lock()
	defer w.lock.Unlock()

	name := user.Name
	if name == "" {
		name = user.Token
	}
	w.users[name] = user
}

func (w *userService) DelUser(name string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if _, ok := w.users[name]; ok {
		delete(w.users, name)
	}
}

func (w *userService) GetUser(name string) *models.User {
	w.lock.RLock()
	defer w.lock.RUnlock()

	if u, ok := w.users[name]; ok {
		return u
	}

	return nil
}

func (w *userService) ListUser() <-chan *models.User {
	c := make(chan *models.User, 128)

	go func() {
		w.lock.RLock()
		defer w.lock.RUnlock()

		for _, u := range w.users {
			c <- u
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}
