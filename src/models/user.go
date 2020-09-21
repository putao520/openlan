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

func NewUser(name, network, password string) *User {
	return &User{
		Name:     name,
		Password: password,
		Network:  network,
		System:   runtime.GOOS,
	}
}

func (u *User) String() string {
	return fmt.Sprintf("%s, %s, %s, %s", u.UUID, u.Name, u.Password, u.Token)
}

func (u *User) Update() {
	// to support lower version
	if u.Network == "" {
		if strings.Contains(u.Name, "@") {
			u.Network = strings.SplitN(u.Name, "@", 2)[1]
		}
	}
	if u.Network == "" {
		u.Network = "default"
	}
	if strings.Contains(u.Name, "@") {
		u.Name = strings.SplitN(u.Name, "@", 2)[0]
	}
	u.Alias = strings.ToLower(u.Alias)
	if u.UUID == "" {
		u.UUID = u.Alias
	}
}

func (u *User) Id() string {
	if u.Name == "" {
		return u.Token
	}
	return u.Name + "@" + u.Network
}
