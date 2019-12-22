package models

import (
	"github.com/danieldin95/openlan-go/libol"
)

type Point struct {
	Alias  string           `json:"alias"`
	Server string           `json:"server"`
	Uptime int64            `json:"uptime"`
	Status string           `json:"status"`
	IfName string           `json:"ifName"`
	Client *libol.TcpClient `json:"-"`
	Device *TapDevice       `json:"-"`
}

func NewPoint(c *libol.TcpClient, d *TapDevice) (w *Point) {
	w = &Point{
		Alias:  "",
		Server: c.LocalAddr(),
		Client: c,
		Device: d,
	}

	return
}

func (p *Point) Update() *Point{
	p.Uptime = p.Client.UpTime()
	p.Status = p.Client.GetState()
	p.IfName = p.Device.Name()
	return p
}
