package olsw

import (
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olsw/api"
)

type ESPWorker struct {
	alias string
	cfg   *config.Network
}

func NewESPWorker(c *config.Network) *ESPWorker {
	w := &ESPWorker{}
	return w
}

func (w *ESPWorker) Initialize() {

}

func (w *ESPWorker) Start(v api.Switcher) {

}

func (w *ESPWorker) Stop() {

}

func (w *ESPWorker) String() string {
	return ""
}

func (w *ESPWorker) ID() string {
	return ""
}

func (w *ESPWorker) GetBridge() network.Bridger {
	return nil
}

func (w *ESPWorker) GetConfig() *config.Network {
	return w.cfg
}

func (w *ESPWorker) GetSubnet() string {
	return ""
}
