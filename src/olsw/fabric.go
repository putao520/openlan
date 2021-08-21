package olsw

import (
	"github.com/danieldin95/openlan/src/config"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/network"
	"github.com/danieldin95/openlan/src/olsw/api"
	"github.com/digitalocean/go-openvswitch/ovs"
	"strings"
)

func ifName(name string) string {
	size := len(name)
	if size < 15 {
		return name
	}
	return name[size-15 : size]
}

type OvsBridge struct {
	name string
	cli  *ovs.Client
	out  *libol.SubLogger
}

func NewOvsBridge(name string) *OvsBridge {
	return &OvsBridge{
		name: name,
		cli:  ovs.New(),
		out:  libol.NewSubLogger(name),
	}
}

func (o *OvsBridge) addFlow(flow *ovs.Flow) error {
	if data, err := flow.MarshalText(); err == nil {
		o.out.Info("OvsBridge.addFlow %s", data)
	}
	if err := o.cli.OpenFlow.AddFlow(o.name, flow); err != nil {
		o.out.Warn("OvsBridge.addFlow %s", err)
		return err
	}
	return nil
}

func (o *OvsBridge) setUp() error {
	if err := o.cli.VSwitch.AddBridge(o.name); err != nil {
		o.out.Error("OvsBridge.AddBridge %s %s", o.name, err)
		return err
	}
	return nil
}

func (o *OvsBridge) setMode(mode ovs.FailMode) error {
	if err := o.cli.VSwitch.SetFailMode(o.name, mode); err != nil {
		o.out.Warn("OvsBridge.setMode %s %s", mode, err)
		return err
	}
	return nil
}

func (o *OvsBridge) addPort(name string, options *ovs.InterfaceOptions) error {
	if err := o.cli.VSwitch.AddPort(o.name, name); err != nil {
		o.out.Warn("OvsBridge.addPort %s %s", name, err)
		return err
	}
	if options == nil {
		return nil
	}
	if err := o.cli.VSwitch.Set.Interface(name, *options); err != nil {
		o.out.Warn("OvsBridge.addPort %s %s", name, err)
		return err
	}
	return nil
}

func (o *OvsBridge) setPort(name string, options ovs.InterfaceOptions) error {
	if err := o.cli.VSwitch.Set.Interface(name, options); err != nil {
		o.out.Warn("OvsBridge.setPort %s %s", name, err)
		return err
	}
	return nil
}

func (o *OvsBridge) dumpPort(name string) *ovs.PortStats {
	if port, err := o.cli.OpenFlow.DumpPort(o.name, name); err != nil {
		o.out.Warn("OvsBridge.dumpPort %s %s", name, err)
		return nil
	} else {
		return port
	}
}

const (
	PatchLvToTun = 2
	VxlanTunToLv = 4
	LearnFromTun = 10
	UcastToTun   = 20
	FloodToTun   = 22
)

type FabricWorker struct {
	uuid   string
	cfg    *config.Network
	out    *libol.SubLogger
	br     *OvsBridge
	cookie uint64
}

func NewFabricWorker(c *config.Network) *FabricWorker {
	w := &FabricWorker{
		cfg:    c,
		out:    libol.NewSubLogger(c.Name),
		br:     NewOvsBridge(c.Bridge.Name),
		cookie: 0x0102,
	}
	return w
}

func (w *FabricWorker) setupTable() {
	_ = w.br.addFlow(&ovs.Flow{
		Actions: []ovs.Action{ovs.Drop()},
	})
	// PatchLvToTun table will handle packets coming from patch_int
	// unicasts go to table UcastToTun where remote addresses are learnt
	_ = w.br.addFlow(&ovs.Flow{
		Table: PatchLvToTun,
		Matches: []ovs.Match{
			ovs.DataLinkDestination("00:00:00:00:00:00/01:00:00:00:00:00"),
		},
		Actions: []ovs.Action{
			ovs.Resubmit(0, UcastToTun),
		},
	})
	// Broadcasts/multicasts go to table FloodToTun that handles flooding
	_ = w.br.addFlow(&ovs.Flow{
		Table: PatchLvToTun,
		Matches: []ovs.Match{
			ovs.DataLinkDestination("01:00:00:00:00:00/01:00:00:00:00:00"),
		},
		Actions: []ovs.Action{
			ovs.Resubmit(0, FloodToTun),
		},
	})
	// Tables VxlanTunToLv will set lvid depending on tun_id
	// for each tunnel type, and resubmit to table LearnFromTun where
	// remote mac addresses will be learnt
	_ = w.br.addFlow(&ovs.Flow{
		Table:   VxlanTunToLv,
		Actions: []ovs.Action{ovs.Drop()},
	})
	// Egress unicast will be handled in table UcastToTun, where remote
	// mac addresses will be learned. For now, just add a default flow that
	// will resubmit unknown unicasts to table FloodToTun to treat them
	// as broadcasts/multicasts
	_ = w.br.addFlow(&ovs.Flow{
		Table: UcastToTun,
		Actions: []ovs.Action{
			ovs.Resubmit(0, FloodToTun),
		},
	})
	_ = w.br.addFlow(&ovs.Flow{
		Table:   FloodToTun,
		Actions: []ovs.Action{ovs.Drop()},
	})
}

func (w *FabricWorker) Initialize() {
	if err := w.br.setUp(); err != nil {
		return
	}
	_ = w.br.setMode("secure")
	w.setupTable()
}

func (w *FabricWorker) AddNetwork(vni uint64) {
	// Flow from tunnel resubmit to LearningFromTun for learning src mac.
	_ = w.br.addFlow(&ovs.Flow{
		Table:    VxlanTunToLv,
		Priority: 1,
		Matches: []ovs.Match{
			ovs.TunnelID(vni),
		},
		Actions: []ovs.Action{
			ovs.Resubmit(0, LearnFromTun),
		},
	})
	// TODO add flood flow from outputs to vxlan tunnel.
}

func (w *FabricWorker) AddOutput(output string) {
	//
	patchPort := 0xff
	// Table 0 (default) will sort incoming traffic depending on in_port
	_ = w.br.addFlow(&ovs.Flow{
		InPort:   patchPort,
		Priority: 1,
		Actions: []ovs.Action{
			ovs.Resubmit(0, PatchLvToTun),
		},
	})
	// LearnFromTun table will have a single flow using a learn action to
	// dynamically set-up flows in UcastToTun corresponding to remote mac
	// Once remote mac addresses are learnt, output packet to patch_int
	learnSpecs := []ovs.Match{
		ovs.FieldMatch("NXM_OF_ETH_DST[]", "NXM_OF_ETH_SRC[]"),
	}
	learnActions := []ovs.Action{
		ovs.Load("NXM_NX_TUN_ID[]", "NXM_NX_TUN_ID[]"),
		ovs.OutputField("NXM_OF_IN_PORT[]"),
	}
	_ = w.br.addFlow(&ovs.Flow{
		Table:    LearnFromTun,
		Priority: 1,
		Actions: []ovs.Action{
			ovs.Learn(&ovs.LearnedFlow{
				Table:       UcastToTun,
				Cookie:      w.cookie,
				Matches:     learnSpecs,
				Priority:    1,
				HardTimeout: 300,
				Actions:     learnActions,
			}),
			ovs.Output(patchPort),
		},
	})
}

func (w *FabricWorker) Ip2Port(addr, pre string) string {
	name := pre + strings.ReplaceAll(addr, ".", "")
	return ifName(name)
}

func (w *FabricWorker) AddTunnel(remote string) {
	name := w.Ip2Port(remote, "vx-")
	options := ovs.InterfaceOptions{
		Type:     ovs.InterfaceTypeVXLAN,
		RemoteIP: remote,
		Key:      "flow",
	}
	if err := w.br.addPort(name, &options); err != nil {
		return
	}
	port := w.br.dumpPort(name)
	if port == nil {
		return
	}
	_ = w.br.addFlow(&ovs.Flow{
		InPort:   port.PortID,
		Priority: 1,
		Actions: []ovs.Action{
			ovs.Resubmit(0, VxlanTunToLv),
		},
	})
}

func (w *FabricWorker) Start(v api.Switcher) {
	w.out.Info("FabricWorker.Start")
	w.AddOutput("br-output")
	w.AddNetwork(0x1024)
	w.AddTunnel("192.168.7.119")
	w.AddTunnel("192.168.111.119")
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
