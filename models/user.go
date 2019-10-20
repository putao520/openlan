package models

import "fmt"

type User struct {
	Name     string `json:"name"`
	Token    string `json:"token"`
	Password string `json:"password"`
}

func NewUser(name string, password string) (this *User) {
	this = &User{
		Name:     name,
		Password: password,
	}
	return
}

func (u *User) String() string {
	return fmt.Sprintf("%s, %s, %s", u.Name, u.Password, u.Token)
}
