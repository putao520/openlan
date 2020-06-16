package app

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/network"
)

type Master interface {
	ReadTap(dev network.Taper, readAt func(f *libol.FrameMessage) error)
	NewTap(tenant string) (network.Taper, error)
	UUID() string
	OffClient(client libol.SocketClient)
}
