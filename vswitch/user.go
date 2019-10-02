package vswitch

import "fmt"

type User struct {
    Name string `json:"name"`
    Token string `json:"token"`
    Password string `json:"password"`
}

func NewUser(name string, password string) (this *User) {
    this = &User {
        Name: name,
        Password: password,
    }
    return
}

func (this *User) String() string {
    return fmt.Sprintf("%s, %s, %s", this.Name, this.Password, this.Token)
}