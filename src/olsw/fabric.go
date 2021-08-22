package olsw

import (
	"github.com/danieldin95/openlan/src/config"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/olsw/api"
	"github.com/digitalocean/go-openvswitch/ovs"
)

type FabricWorker struct {
	uuid string
	cfg  *config.Network
	out  *libol.SubLogger
	cli  *ovs.Client
}

func NewFabricWorker(c *config.Network) *FabricWorker {
	w := &FabricWorker{
		cfg: c,
		out: libol.NewSubLogger(c.Name),
		cli: ovs.New(),
	}
	return w
}

func (w *FabricWorker) Initialize() {
	br := w.cfg.Bridge.Name
	if err := w.cli.VSwitch.AddBridge(br); err != nil {
		w.out.Error("FabricWorker.Initialize add bridge: %v", err)
	}
}

func (w *FabricWorker) Start(v api.Switcher) {
	w.out.Info("FabricWorker.Start")
}

func (w *FabricWorker) Stop() {
	w.out.Info("FabricWorker.Stop")
}

func (w *FabricWorker) String() string {
	return w.cfg.Name
}

func (w *FabricWorker) ID() string {
	return w.uuid
}
