package olsw

import (
	co "github.com/danieldin95/openlan/pkg/config"
	"github.com/danieldin95/openlan/pkg/network"
	"github.com/danieldin95/openlan/pkg/olsw/api"
)

type Networker interface {
	String() string
	ID() string
	Initialize()
	Start(v api.Switcher)
	Stop()
	GetBridge() network.Bridger
	GetConfig() *co.Network
	GetSubnet() string
	Reload(c *co.Network)
}

func NewNetworker(c *co.Network) Networker {
	switch c.Provider {
	case "esp":
		return NewESPWorker(c)
	case "vxlan":
		return NewVxLANWorker(c)
	case "fabric":
		return NewFabricWorker(c)
	default:
		return NewOpenLANWorker(c)
	}
}
