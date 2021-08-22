package app

import (
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/network"
)

type Master interface {
	UUID() string
	Protocol() string
	OffClient(client libol.SocketClient)
	ReadTap(device network.Taper, readAt func(f *libol.FrameMessage) error)
	NewTap(tenant string) (network.Taper, error)
}
