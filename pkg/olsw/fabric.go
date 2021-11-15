package olsw

import (
	"fmt"
	"github.com/danieldin95/go-openvswitch/ovs"
	co "github.com/danieldin95/openlan/pkg/config"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/danieldin95/openlan/pkg/network"
	"github.com/danieldin95/openlan/pkg/olsw/api"
	"github.com/vishvananda/netlink"
	"strings"
)

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

func (o *OvsBridge) delFlow(flow *ovs.MatchFlow) error {
	if err := o.cli.OpenFlow.DelFlows(o.name, flow); err != nil {
		o.out.Warn("OvsBridge.addFlow %s", err)
		return err
	}
	return nil
}

func (o *OvsBridge) addFlow(flow *ovs.Flow) error {
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

func (o *OvsBridge) delPort(name string) error {
	if err := o.cli.VSwitch.DeletePort(o.name, name); err != nil {
		o.out.Warn("OvsBridge.delPort %s %s", name, err)
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

const (
	MatchRegLv   = "reg5"
	MatchRegTun  = "reg6"
	NxmRegLv     = "NXM_NX_REG5[0..31]"
	NxmRegTun    = "NXM_NX_REG6[0..31]"
	NxmRegEthDst = "NXM_OF_ETH_DST[]"
	NxmRegEthSrc = "NXM_OF_ETH_SRC[]"
	NxmRegTunId  = "NXM_NX_TUN_ID[0..31]"
	NxmRegInPort = "NXM_OF_IN_PORT[]"
)

const (
	MatchUnicast = "00:00:00:00:00:00/01:00:00:00:00:00"
	MatchMulcast = "01:00:00:00:00:00/01:00:00:00:00:00"
)

type OvsPort struct {
	name    string
	portId  int
	options ovs.InterfaceOptions
}

type OvsNetwork struct {
	bridge string
	vni    uint32
	patch  *OvsPort
}

type FabricWorker struct {
	uuid     string
	cfg      *co.Network
	spec     *co.FabricSpecifies
	out      *libol.SubLogger
	ovs      *OvsBridge
	cookie   uint64
	tunnels  map[string]*OvsPort
	networks map[uint32]*OvsNetwork
}

func NewFabricWorker(c *co.Network) *FabricWorker {
	w := &FabricWorker{
		cfg:      c,
		out:      libol.NewSubLogger(c.Name),
		ovs:      NewOvsBridge(c.Bridge.Name),
		tunnels:  make(map[string]*OvsPort, 1024),
		networks: make(map[uint32]*OvsNetwork, 1024),
	}
	w.spec, _ = c.Specifies.(*co.FabricSpecifies)
	return w
}

func (w *FabricWorker) setupTable() {
	_ = w.ovs.addFlow(&ovs.Flow{
		Actions: []ovs.Action{ovs.Drop()},
	})
	// PatchLvToTun table will handle packets coming from patch_int
	// unicasts go to table UcastToTun where remote addresses are learnt
	_ = w.ovs.addFlow(&ovs.Flow{
		Table: PatchLvToTun,
		Matches: []ovs.Match{
			ovs.DataLinkDestination(MatchUnicast),
		},
		Actions: []ovs.Action{
			ovs.Resubmit(0, UcastToTun),
		},
	})
	// Broadcasts/multicasts go to table FloodToTun that handles flooding
	_ = w.ovs.addFlow(&ovs.Flow{
		Table: PatchLvToTun,
		Matches: []ovs.Match{
			ovs.DataLinkDestination(MatchMulcast),
		},
		Actions: []ovs.Action{
			ovs.Resubmit(0, FloodToTun),
		},
	})
	// Tables VxlanTunToLv will set REG6 depending on tun_id
	// for each tunnel type, and resubmit to table LearnFromTun where
	// remote mac addresses will be learnt
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:   VxlanTunToLv,
		Actions: []ovs.Action{ovs.Drop()},
	})
	// Egress unicast will be handled in table UcastToTun, where remote
	// mac addresses will be learned. For now, just add a default flow that
	// will resubmit unknown unicasts to table FloodToTun to treat them
	// as broadcasts/multicasts
	_ = w.ovs.addFlow(&ovs.Flow{
		Table: UcastToTun,
		Actions: []ovs.Action{
			ovs.Resubmit(0, FloodToTun),
		},
	})
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:   FloodToTun,
		Actions: []ovs.Action{ovs.Drop()},
	})
}

func (w *FabricWorker) Initialize() {
	if err := w.ovs.setUp(); err != nil {
		return
	}
	_ = w.ovs.setMode("secure")
	w.setupTable()
}

func (w *FabricWorker) getPatchTun(vni uint32) (string, string) {
	brPort := fmt.Sprintf("vnb-%x", vni)
	tunPort := fmt.Sprintf("vnt-%x", vni)
	return brPort, tunPort
}

func (w *FabricWorker) setupNetwork(bridge string, vni uint32) *OvsNetwork {
	brPort, tunPort := w.getPatchTun(vni)
	link := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: tunPort},
		PeerName:  brPort,
	}
	if err := netlink.LinkAdd(link); err != nil {
		w.out.Warn("FabricWorker.addLink %s", err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		w.out.Warn("FabricWorker.setLinkUp %s", err)
	}
	// Setup linux bridge for outputs
	br := network.NewLinuxBridge(bridge, 0)
	br.Open("")
	_ = br.AddSlave(brPort)
	if err := br.CallIptables(1); err != nil {
		w.out.Warn("FabricWorker.CallIptables %s", err)
	}
	// Add port to Ovs tunnel bridge
	_ = w.ovs.addPort(tunPort, nil)
	net := OvsNetwork{
		bridge: bridge,
		vni:    vni,
		patch: &OvsPort{
			name: tunPort,
		},
	}
	if port := w.ovs.dumpPort(tunPort); port != nil {
		net.patch.portId = port.PortID
	}
	return &net
}

func (w *FabricWorker) AddNetwork(bridge string, vni uint32) {
	net := w.setupNetwork(bridge, vni)
	// Flow from tunnel resubmit to LearningFromTun for learning src mac.
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:    VxlanTunToLv,
		Priority: 1,
		Matches: []ovs.Match{
			ovs.TunnelID(uint64(vni)),
		},
		Actions: []ovs.Action{
			ovs.Move(NxmRegTunId, NxmRegTun),
			ovs.Resubmit(0, LearnFromTun),
		},
	})
	patchPort := net.patch.portId
	// Table 0 (default) will sort incoming traffic depending on in_port
	_ = w.ovs.addFlow(&ovs.Flow{
		InPort:   patchPort,
		Priority: 1,
		Actions: []ovs.Action{
			ovs.Load(libol.Uint2S(vni), NxmRegLv),
			ovs.Resubmit(0, PatchLvToTun),
		},
	})
	// LearnFromTun table will have a single flow using a learn action to
	// dynamically set-up flows in UcastToTun corresponding to remote mac
	// Once remote mac addresses are learnt, output packet to patch_int
	learnSpecs := []ovs.Match{
		ovs.FieldMatch(MatchRegLv, NxmRegTun),
		ovs.FieldMatch(NxmRegEthDst, NxmRegEthSrc),
	}
	learnActions := []ovs.Action{
		ovs.Load(NxmRegTunId, NxmRegTunId),
		ovs.OutputField(NxmRegInPort),
	}
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:    LearnFromTun,
		Priority: 1,
		Matches: []ovs.Match{
			ovs.FieldMatch(MatchRegTun, libol.Uint2S(vni)),
		},
		Actions: []ovs.Action{
			ovs.Learn(&ovs.LearnedFlow{
				Table:       UcastToTun,
				Matches:     learnSpecs,
				Priority:    1,
				HardTimeout: 300,
				Actions:     learnActions,
			}),
			ovs.Output(patchPort),
		},
	})
	w.networks[vni] = net
	// Install flow for flooding to tunnels.
	w.flood2Tunnel(vni)
}

func (w *FabricWorker) AddOutput(bridge string, vlan int, output string) {
	link, err := netlink.LinkByName(output)
	if err != nil {
		w.out.Error("FabricWorker.LinkByName %s", err)
		return
	}
	if err := netlink.LinkSetUp(link); err != nil {
		w.out.Warn("FabricWorker.LinkSetUp %s", err)
	}
	subLink := &netlink.Vlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:        fmt.Sprintf("%s.%d", output, vlan),
			ParentIndex: link.Attrs().Index,
		},
		VlanId: vlan,
	}
	if err := netlink.LinkAdd(subLink); err != nil {
		w.out.Error("FabricWorker.LinkAdd %s", err)
		return
	}
	br := network.NewBrCtl(bridge, 0)
	if err := br.AddPort(subLink.Name); err != nil {
		w.out.Warn("FabricWorker.AddPort %s", err)
	}
}

func (w *FabricWorker) Addr2Port(addr, pre string) string {
	name := pre + strings.ReplaceAll(addr, ".", "")
	return libol.IfName(name)
}

func (w *FabricWorker) flood2Tunnel(vni uint32) {
	actions := []ovs.Action{
		ovs.Move(NxmRegLv, NxmRegTunId),
	}
	for _, tun := range w.tunnels {
		actions = append(actions, ovs.Output(tun.portId))
	}
	if vni == 0 {
		for vni := range w.networks {
			_ = w.ovs.addFlow(&ovs.Flow{
				Table:    FloodToTun,
				Priority: 1,
				Matches: []ovs.Match{
					ovs.FieldMatch(MatchRegLv, libol.Uint2S(vni)),
				},
				Actions: actions,
			})
		}
	} else {
		_ = w.ovs.addFlow(&ovs.Flow{
			Table:    FloodToTun,
			Priority: 1,
			Matches: []ovs.Match{
				ovs.FieldMatch(MatchRegLv, libol.Uint2S(vni)),
			},
			Actions: actions,
		})
	}
}

func (w *FabricWorker) AddTunnel(remote string, dport uint32) {
	name := w.Addr2Port(remote, "vx-")
	options := ovs.InterfaceOptions{
		Type:      ovs.InterfaceTypeVXLAN,
		BfdEnable: true,
		RemoteIP:  remote,
		Key:       "flow",
		DfDefault: "false",
		DstPort:   dport,
	}
	if err := w.ovs.addPort(name, &options); err != nil {
		return
	}
	port := w.ovs.dumpPort(name)
	if port == nil {
		return
	}
	_ = w.ovs.addFlow(&ovs.Flow{
		InPort:   port.PortID,
		Priority: 1,
		Actions: []ovs.Action{
			ovs.Resubmit(0, VxlanTunToLv),
		},
	})
	w.tunnels[name] = &OvsPort{
		name:    name,
		portId:  port.PortID,
		options: options,
	}
	// Update flow for flooding to tunnels.
	w.flood2Tunnel(0)
}

func (w *FabricWorker) Start(v api.Switcher) {
	w.out.Info("FabricWorker.Start")
	firewall := v.Firewall()
	mss := w.spec.Mss
	for _, tunnel := range w.spec.Tunnels {
		w.AddTunnel(tunnel.Remote, tunnel.DstPort)
	}
	for _, net := range w.spec.Networks {
		w.AddNetwork(net.Bridge, net.Vni)
		if mss > 0 {
			firewall.AddRule(network.IpRule{
				Table:   network.TMangle,
				Chain:   network.CPostRoute,
				Output:  net.Bridge,
				Proto:   "tcp",
				Match:   "tcp",
				TcpFlag: []string{"SYN,RST", "SYN"},
				Jump:    "TCPMSS",
				SetMss:  mss,
			})
		}
		firewall.AddRule(network.IpRule{
			Table:  network.TFilter,
			Chain:  network.OLCForward,
			Input:  net.Bridge,
			Output: net.Bridge,
		})
		for _, port := range net.Outputs {
			w.AddOutput(net.Bridge, port.Vlan, port.Interface)
		}
	}
}

func (w *FabricWorker) clear() {
	_ = w.ovs.delFlow(nil)
}

func (w *FabricWorker) cleanNetwork(bridge string, vni uint32) {
	brPort, tunPort := w.getPatchTun(vni)
	_ = w.ovs.delPort(tunPort)
	link := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: tunPort},
		PeerName:  brPort,
	}
	_ = netlink.LinkDel(link)
}

func (w *FabricWorker) DelNetwork(bridge string, vni uint32) {
	w.cleanNetwork(bridge, vni)
}

func (w *FabricWorker) DelTunnel(remote string) {
	name := w.Addr2Port(remote, "vx-")
	_ = w.ovs.delPort(name)
}

func (w *FabricWorker) DelOutput(bridge string, vlan int, output string) {
	subLink := &netlink.Vlan{
		LinkAttrs: netlink.LinkAttrs{
			Name: fmt.Sprintf("%s.%d", output, vlan),
		},
	}
	if err := netlink.LinkDel(subLink); err != nil {
		w.out.Error("FabricWorker.LinkDel %s", err)
		return
	}
}

func (w *FabricWorker) Stop() {
	w.out.Info("FabricWorker.Stop")
	for _, net := range w.spec.Networks {
		for _, port := range net.Outputs {
			w.DelOutput(net.Bridge, port.Vlan, port.Interface)
		}
		w.DelNetwork(net.Bridge, net.Vni)
	}
	for _, tunnel := range w.spec.Tunnels {
		w.DelTunnel(tunnel.Remote)
	}
	w.clear()
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

func (w *FabricWorker) GetConfig() *co.Network {
	return w.cfg
}

func (w *FabricWorker) GetSubnet() string {
	w.out.Warn("FabricWorker.GetSubnet notSupport")
	return ""
}

func (w *FabricWorker) Reload(c *co.Network) {

}
