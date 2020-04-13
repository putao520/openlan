package models

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/network"
)

type Point struct {
	UUID   string           `json:"uuid"`
	Alias  string           `json:"alias"`
	Tenant string           `json:"Network"`
	Server string           `json:"server"`
	Uptime int64            `json:"uptime"`
	Status string           `json:"status"`
	IfName string           `json:"ifName"`
	Client *libol.TcpClient `json:"-"`
	Device network.Taper    `json:"-"`
}

func NewPoint(c *libol.TcpClient, d network.Taper) (w *Point) {
	w = &Point{
		Alias:  "",
		Server: c.LocalAddr(),
		Client: c,
		Device: d,
	}

	return
}

func (p *Point) Update() *Point {
	p.Uptime = p.Client.UpTime()
	p.Status = p.Client.State()
	p.IfName = p.Device.Name()
	return p
}
