package olsw

import (
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olsw/api"
)

type VxLANWorker struct {
	alias string
	cfg   *config.Network
}

func NewVxLANWorker(c *config.Network) *VxLANWorker {
	w := &VxLANWorker{
		cfg: c,
	}
	return w
}

func (w *VxLANWorker) Initialize() {

}

func (w *VxLANWorker) Start(v api.Switcher) {

}

func (w *VxLANWorker) Stop() {

}

func (w *VxLANWorker) String() string {
	return ""
}

func (w *VxLANWorker) ID() string {
	return ""
}

func (w *VxLANWorker) GetBridge() network.Bridger {
	return nil
}

func (w *VxLANWorker) GetConfig() *config.Network {
	return w.cfg
}

func (w *VxLANWorker) GetSubnet() string {
	return ""
}
