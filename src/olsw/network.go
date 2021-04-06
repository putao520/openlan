package olsw

import (
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olsw/api"
)

type Networker interface {
	String() string
	ID() string
	Initialize()
	Start(v api.Switcher)
	Stop()
	GetBridge() network.Bridger
	GetConfig() *config.Network
	GetSubnet() string
}

func NewNetworker(c *config.Network) Networker {
	switch c.Provider {
	case "esp":
		return NewESPWorker(c)
	case "vxlan":
		return NewVxLANWorker(c)
	default:
		return NewOpenLANWorker(c)
	}
}
