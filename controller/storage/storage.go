package storage

import (
	"github.com/danieldin95/lightstar/libstar"
	"github.com/danieldin95/openlan-go/controller/schema"
)

type Storage struct {
	Users   Users
	VSwitch VSwitch
}

var Storager = Storage{
	Users: Users{
		Users: make(map[string]*schema.User, 32),
	},
	VSwitch: VSwitch{
		VSwitch: make(map[string]*schema.VSwitch, 32),
	},
}

func (s *Storage) Load(path string) {
	if err := s.Users.Load(path + "/auth.json"); err != nil {
		libstar.Error("Storage.Load.Users %s", err)
	}
	libstar.Debug("Storage.Load %s", s.Users)
	if err := s.VSwitch.Load(path + "/vswitch.json"); err != nil {
		libstar.Error("Storage.Load.VSwitch %s", err)
	}
	libstar.Debug("Storage.Load %s", s.VSwitch)
}
