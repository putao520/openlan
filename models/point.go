package models

import (
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/songgao/water"
)

type Point struct {
	Client *libol.TcpClient
	Device *water.Interface
}

func NewPoint(c *libol.TcpClient, d *water.Interface) (w *Point) {
	w = &Point{
		Client: c,
		Device: d,
	}

	return
}
