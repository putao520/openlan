package models

import (
	"fmt"
)

type User struct {
	Alias    string `json:"alias"`
	Name     string `json:"name"`
	Tenant   string `json:"tenant"`
	Token    string `json:"token"`
	Password string `json:"password"`
	UUID     string `json:"uuid"`
}

func NewUser(name string, password string) (this *User) {
	this = &User{
		Name:     name,
		Password: password,
	}
	return
}

func (u *User) String() string {
	return fmt.Sprintf("%s, %s, %s, %s", u.UUID, u.Name, u.Password, u.Token)
}
