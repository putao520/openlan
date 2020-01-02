package api

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/network"
)

type Worker interface {
	GetId() string
	GetServer() *libol.TcpServer
	NewTap() (network.Taper, error)
	Write(network.Taper, []byte)
}