package olsw

import (
	"github.com/danieldin95/openlan/src/config"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/network"
	"github.com/danieldin95/openlan/src/olsw/api"
	"github.com/digitalocean/go-openvswitch/ovs"
)

const (
	FloodToTun = 22
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

func (w *FabricWorker) setupTable() {
	br := w.cfg.Bridge.Name
	_ = w.cli.OpenFlow.AddFlow(br, &ovs.Flow{
		Table:    0,
		Priority: 0,
		Actions:  []ovs.Action{ovs.Drop()},
	})
	_ = w.cli.OpenFlow.AddFlow(br, &ovs.Flow{
		Table:    FloodToTun,
		Priority: 0,
		Actions:  []ovs.Action{ovs.Drop()},
	})
}

func (w *FabricWorker) Initialize() {
	br := w.cfg.Bridge.Name
	if err := w.cli.VSwitch.AddBridge(br); err != nil {
		w.out.Error("FabricWorker.AddBridge %v", err)
	}
	if err := w.cli.VSwitch.SetFailMode(br, "secure"); err != nil {
		w.out.Error("FabricWorker.SetFailMode %v", err)
	}
	w.setupTable()
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

func (w *FabricWorker) GetBridge() network.Bridger {
	w.out.Warn("FabricWorker.GetBridge notSupport")
	return nil
}

func (w *FabricWorker) GetConfig() *config.Network {
	return w.cfg
}

func (w *FabricWorker) GetSubnet() string {
	w.out.Warn("FabricWorker.GetSubnet notSupport")
	return ""
}
