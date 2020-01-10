package app

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/network"
)

type Worker interface {
	GetId() string
	GetServer() *libol.TcpServer
	ReadTap (dev network.Taper, readAt func(p []byte) error)
	NewTap() (network.Taper, error)
}