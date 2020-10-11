package models

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/network"
)

type Point struct {
	UUID    string             `json:"uuid"`
	Alias   string             `json:"alias"`
	Network string             `json:"network"`
	User    string             `json:"user"`
	Server  string             `json:"server"`
	Uptime  int64              `json:"uptime"`
	Status  string             `json:"status"`
	IfName  string             `json:"device"`
	Client  libol.SocketClient `json:"-"`
	Device  network.Taper      `json:"-"`
	System  string             `json:"system"`
}

func NewPoint(c libol.SocketClient, d network.Taper) (w *Point) {
	return &Point{
		Alias:  "",
		Server: c.LocalAddr(),
		Client: c,
		Device: d,
	}
}

func (p *Point) Update() *Point {
	client := p.Client
	if client != nil {
		p.Uptime = client.UpTime()
		p.Status = client.Status().String()
	}
	device := p.Device
	if device != nil {
		p.IfName = device.Name()
	}
	return p
}

func (p *Point) SetUser(user *User) {
	p.User = user.Name
	p.UUID = user.UUID
	if len(p.UUID) > 13 {
		// too long and using short uuid.
		p.UUID = p.UUID[:13]
	}
	p.Network = user.Network
	p.System = user.System
	p.Alias = user.Alias
}
