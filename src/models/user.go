package models

import (
	"fmt"
	"runtime"
	"strings"
)

type User struct {
	Alias    string `json:"alias"`
	Name     string `json:"name"`
	Network  string `json:"network"`
	Token    string `json:"token"`
	Password string `json:"password"`
	UUID     string `json:"uuid"`
	System   string `json:"system"`
}

func NewUser(name string, password string) (this *User) {
	this = &User{
		Name:     name,
		Password: password,
		System:   runtime.GOOS,
	}
	return
}

func (u *User) String() string {
	return fmt.Sprintf("%s, %s, %s, %s", u.UUID, u.Name, u.Password, u.Token)
}

func (u *User) Update() {
	// to support lower version
	if u.Network == "" {
		if strings.Contains(u.Name, "@") {
			u.Network = strings.SplitN(u.Name, "@", 2)[1]
		} else {
			u.Network = "default"
		}
	}
	if !strings.Contains(u.Name, "@") {
		u.Name += "@" + u.Network
	}
	u.Alias = strings.ToLower(u.Alias)
	if u.UUID == "" {
		u.UUID = u.Alias
	}
}
