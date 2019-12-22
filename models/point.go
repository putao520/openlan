package models

import (
	"github.com/danieldin95/openlan-go/libol"
)

type Point struct {
	Alias  string           `json:"alias"`
	Server string           `json:"server"`
	Uptime uint64           `json:"uptime"`
	Status uint64           `json:"status"`
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
