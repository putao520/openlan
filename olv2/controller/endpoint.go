package controller

import (
	"net"
	"encoding/json"
	"strings"
)

type Endpoint struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
	Password string `json:"password"`

	Network string
	UdpAddr *net.UDPAddr
}

func GetNetwork(name string) (string) {
	values := strings.Split(name, "@")
	if len(values) == 2 {
		return values[1]
	}
	return ""
}

func NewEndpoint(name string, uuid string) (this *Endpoint) {
	this = &Endpoint {
		Name: name, 
		UUID: uuid,
	}

	this.Network = GetNetwork(this.Name)
	return
}

func NewEndpointFromJson(data string) (this *Endpoint, err error) {
	this = &Endpoint {}
	if err = json.Unmarshal([]byte(data), this); err != nil {
		return
	}

	this.Network = GetNetwork(this.Name)
	return
}

func (this *Endpoint) Equal(obj *Endpoint) bool {
	return this.Name == obj.Name && this.UUID == obj.UUID 
}

func (this *Endpoint) ToJson() (string, error) {
	data, err := json.Marshal(this)
	return string(data), err
}
