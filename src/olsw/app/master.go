package app

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/network"
)

type Master interface {
	ReadTap(device network.Taper, readAt func(f *libol.FrameMessage) error)
	NewTap(tenant string) (network.Taper, error)
	UUID() string
	OffClient(client libol.SocketClient)
}
