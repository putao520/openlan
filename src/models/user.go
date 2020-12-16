package models

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"runtime"
	"strings"
)

type User struct {
	Alias    string             `json:"alias"`
	Name     string             `json:"name"`
	Network  string             `json:"network"`
	Token    string             `json:"token"`
	Password string             `json:"password"`
	UUID     string             `json:"uuid"`
	System   string             `json:"system"`
	Role     string             `json:"type"` // admin or guest
	Last     libol.SocketClient `json:"last"` // lastly accessed by this.
}

func NewUser(name, network, password string) *User {
	return &User{
		Name:     name,
		Password: password,
		Network:  network,
		System:   runtime.GOOS,
		Role:     "guest",
	}
}

func (u *User) String() string {
	return fmt.Sprintf("%s, %s, %s", u.Name, u.Password, u.Role)
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
