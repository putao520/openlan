package storage

import (
	"github.com/danieldin95/openlan-go/controller/schema"
	"github.com/danieldin95/openlan-go/libol"
)

type Storage struct {
	Users Users
}

var Storager = Storage{
	Users: Users{
		Users: make(map[string]*schema.User, 32),
	},
}

func (s *Storage) Load(path string) {
	if err := s.Users.Load(path + "/auth.json"); err != nil {
		libol.Error("Storage.Load.Users %s", err)
	}
	libol.Debug("Storage.Load %s", s.Users)
}
