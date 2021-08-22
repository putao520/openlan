package storage

import (
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/schema"
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
