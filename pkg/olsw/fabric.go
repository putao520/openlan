package olsw

import (
	"fmt"
	"github.com/danieldin95/go-openvswitch/ovs"
	co "github.com/danieldin95/openlan/pkg/config"
	"github.com/danieldin95/openlan/pkg/libol"
	cn "github.com/danieldin95/openlan/pkg/network"
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

func (o *OvsBridge) setDown() error {
	if err := o.cli.VSwitch.DeleteBridge(o.name); err != nil {
		o.out.Error("OvsBridge.DeleteBridge %s %s", o.name, err)
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
	TLsToTun     = 2  // From a switch include border to tunnels.
	TTunToLs     = 4  // From tunnels to a switch.
	TSourceLearn = 10 // Learning source mac.
	TUcastToTun  = 20 // Forwarding by fdb.
	TFloodToTun  = 30 // Flooding to tunnels or patch by flags.
	TFloodToBor  = 31 // Flooding to border in a switch.
	TFloodLoop   = 32 // Flooding to patch in a switch from border.
)

const (
	FFromLs  = 2 // In a logical switch.
	FFromTun = 4 // From peer tunnels.
)

const (
	MatchRegFlag = "reg10"
	NxmRegFlag   = "NXM_NX_REG10[0..31]"
	NxmRegEthDst = "NXM_OF_ETH_DST[]"
	NxmRegEthSrc = "NXM_OF_ETH_SRC[]"
	NxmRegTunId  = "NXM_NX_TUN_ID[0..31]"
	NxmRegInPort = "NXM_OF_IN_PORT[]"
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
	borders  map[string]*OvsPort
	networks map[uint32]*OvsNetwork
	bridges  map[string]*cn.LinuxBridge
}

func NewFabricWorker(c *co.Network) *FabricWorker {
	w := &FabricWorker{
		cfg:      c,
		out:      libol.NewSubLogger(c.Name),
		ovs:      NewOvsBridge(c.Bridge.Name),
		tunnels:  make(map[string]*OvsPort, 1024),
		borders:  make(map[string]*OvsPort, 1024),
		networks: make(map[uint32]*OvsNetwork, 1024),
		bridges:  make(map[string]*cn.LinuxBridge, 1024),
	}
	w.spec, _ = c.Specifies.(*co.FabricSpecifies)
	return w
}

func (w *FabricWorker) upTables() {
	_ = w.ovs.addFlow(&ovs.Flow{
		Actions: []ovs.Action{ovs.Drop()},
	})
	// Table 2: set flags from logical switch.
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:    TLsToTun,
		Priority: 1,
		Actions: []ovs.Action{
			ovs.Load(libol.Uint2S(FFromLs), NxmRegFlag),
			ovs.Resubmit(0, TSourceLearn),
		},
	})
	// Table 4: set flags from tunnels.
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:    TTunToLs,
		Priority: 1,
		Actions: []ovs.Action{
			ovs.Load(libol.Uint2S(FFromTun), NxmRegFlag),
			ovs.Resubmit(0, TSourceLearn),
		},
	})
	// Table 10: source learning
	w.addLearning()
	// Table 20: default to flood 30
	_ = w.ovs.addFlow(&ovs.Flow{
		Table: TUcastToTun,
		Actions: []ovs.Action{
			ovs.Resubmit(0, TFloodToTun),
		},
	})
	// Table 30: default drop.
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:   TFloodToTun,
		Actions: []ovs.Action{ovs.Drop()},
	})
	// Table 31: default drop.
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:   TFloodToBor,
		Actions: []ovs.Action{ovs.Drop()},
	})
}

func (w *FabricWorker) Initialize() {
	if err := w.ovs.setUp(); err != nil {
		return
	}
	_ = w.ovs.setMode("secure")
	w.upTables()
}

func (w *FabricWorker) vni2peer(vni uint32) (string, string) {
	tunPort := fmt.Sprintf("vb-%08d", vni)
	brPort := fmt.Sprintf("vt-%08d", vni)
	return brPort, tunPort
}

func (w *FabricWorker) UpLink(bridge string, vni uint32, addr string) *OvsNetwork {
	brPort, tunPort := w.vni2peer(vni)
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
	br := cn.NewLinuxBridge(bridge, 0)
	br.Open(addr)
	_ = br.AddSlave(brPort)
	if err := br.CallIptables(1); err != nil {
		w.out.Warn("FabricWorker.IpTables %s", err)
	}
	w.bridges[bridge] = br
	// Add port to OvS tunnel bridge
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

func (w *FabricWorker) addLearning() {
	// Table 10: source mac learning
	learnSpecs := []ovs.Match{
		ovs.FieldMatch(NxmRegTunId, NxmRegTunId),
		ovs.FieldMatch(NxmRegEthDst, NxmRegEthSrc),
	}
	learnActions := []ovs.Action{
		ovs.OutputField(NxmRegInPort),
	}
	_ = w.ovs.addFlow(&ovs.Flow{
		Table: TSourceLearn,
		Actions: []ovs.Action{
			ovs.Learn(&ovs.LearnedFlow{
				Table:       TUcastToTun,
				Matches:     learnSpecs,
				Priority:    1,
				HardTimeout: 300,
				Actions:     learnActions,
			}),
			ovs.Resubmit(0, TUcastToTun),
		},
	})
}

func (w *FabricWorker) AddNetwork(cfg *co.FabricNetwork) {
	libol.Info("Fabric.AddNetwork %d", cfg.Vni)
	net := w.UpLink(cfg.Bridge, cfg.Vni, cfg.Address)
	patchPort := net.patch.portId
	// Table 0: load tunnel id from patch port.
	_ = w.ovs.addFlow(&ovs.Flow{
		InPort:   patchPort,
		Priority: 1,
		Actions: []ovs.Action{
			ovs.Load(libol.Uint2S(cfg.Vni), NxmRegTunId),
			ovs.Resubmit(0, TLsToTun),
		},
	})
	// Table 30: flooding to patch from tunnels.
	w.networks[cfg.Vni] = net
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:    TFloodToTun,
		Priority: 2,
		Matches: []ovs.Match{
			ovs.FieldMatch(NxmRegTunId, libol.Uint2S(cfg.Vni)),
			ovs.FieldMatch(MatchRegFlag, libol.Uint2S(FFromTun)),
		},
		Actions: []ovs.Action{
			ovs.Output(patchPort),
			ovs.Resubmit(0, TFloodToBor),
		},
	})
	// Table 32: flooding to patch from border.
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:    TFloodLoop,
		Priority: 2,
		Matches: []ovs.Match{
			ovs.FieldMatch(NxmRegTunId, libol.Uint2S(cfg.Vni)),
			ovs.FieldMatch(MatchRegFlag, libol.Uint2S(FFromLs)),
		},
		Actions: []ovs.Action{
			ovs.Output(patchPort),
		},
	})
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
	br := cn.NewBrCtl(bridge, 0)
	if err := br.AddPort(subLink.Name); err != nil {
		w.out.Warn("FabricWorker.AddPort %s", err)
	}
}

func (w *FabricWorker) Addr2Port(addr, pre string) string {
	name := pre + strings.ReplaceAll(addr, ".", "")
	return libol.IfName(name)
}

func (w *FabricWorker) flood2Tunnel() {
	var actions []ovs.Action
	for _, tun := range w.tunnels {
		actions = append(actions, ovs.Output(tun.portId))
	}
	actions = append(actions, ovs.Resubmit(0, TFloodToBor))
	// Table 30: Flooding to tunnels from patch.
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:    TFloodToTun,
		Priority: 1,
		Matches: []ovs.Match{
			ovs.FieldMatch(MatchRegFlag, libol.Uint2S(FFromLs)),
		},
		Actions: actions,
	})
}

func (w *FabricWorker) flood2Border() {
	var actions []ovs.Action
	for _, port := range w.borders {
		actions = append(actions, ovs.Output(port.portId))
	}
	// Table 31: flooding to border from tunnels.
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:    TFloodToBor,
		Priority: 1,
		Matches: []ovs.Match{
			ovs.FieldMatch(MatchRegFlag, libol.Uint2S(FFromTun)),
		},
		Actions: actions,
	})
	// Table 32: flooding to border from a border.
	actions = append(actions, ovs.Resubmit(0, TFloodLoop))
	_ = w.ovs.addFlow(&ovs.Flow{
		Table:    TFloodToBor,
		Priority: 1,
		Matches: []ovs.Match{
			ovs.FieldMatch(MatchRegFlag, libol.Uint2S(FFromLs)),
		},
		Actions: actions,
	})
}

func (w *FabricWorker) tunnelType() ovs.InterfaceType {
	if w.spec.Driver == "stt" {
		return ovs.InterfaceTypeSTT
	}
	return ovs.InterfaceTypeVXLAN
}

func (w *FabricWorker) AddTunnel(cfg *co.FabricTunnel) {
	name := w.Addr2Port(cfg.Remote, "vx-")
	options := ovs.InterfaceOptions{
		Type:      w.tunnelType(),
		BfdEnable: true,
		RemoteIP:  cfg.Remote,
		Key:       "flow",
		DstPort:   cfg.DstPort,
	}
	if w.spec.Fragment {
		options.DfDefault = "false"
	} else {
		options.DfDefault = "true"
	}
	if err := w.ovs.addPort(name, &options); err != nil {
		return
	}
	port := w.ovs.dumpPort(name)
	if port == nil {
		return
	}
	if cfg.Mode == "border" {
		_ = w.ovs.addFlow(&ovs.Flow{
			InPort:   port.PortID,
			Priority: 1,
			Actions: []ovs.Action{
				ovs.Resubmit(0, TLsToTun),
			},
		})
		w.borders[name] = &OvsPort{
			name:    name,
			portId:  port.PortID,
			options: options,
		}
		// Update flow for flooding to border.
		w.flood2Border()
	} else {
		_ = w.ovs.addFlow(&ovs.Flow{
			InPort:   port.PortID,
			Priority: 1,
			Actions: []ovs.Action{
				ovs.Resubmit(0, TTunToLs),
			},
		})
		w.tunnels[name] = &OvsPort{
			name:    name,
			portId:  port.PortID,
			options: options,
		}
		// Update flow for flooding to tunnels.
		w.flood2Tunnel()
	}
}

func (w *FabricWorker) Start(v api.Switcher) {
	w.out.Info("FabricWorker.Start")
	firewall := v.Firewall()
	mss := w.spec.Mss
	for _, tunnel := range w.spec.Tunnels {
		w.AddTunnel(tunnel)
	}
	for _, net := range w.spec.Networks {
		w.AddNetwork(net)
		if mss > 0 {
			firewall.AddRule(cn.IpRule{
				Table:   cn.TMangle,
				Chain:   cn.OLCPost,
				Output:  net.Bridge,
				Proto:   "tcp",
				Match:   "tcp",
				TcpFlag: []string{"SYN,RST", "SYN"},
				Jump:    "TCPMSS",
				SetMss:  mss,
			})
		}
		firewall.AddRule(cn.IpRule{
			Table:  cn.TFilter,
			Chain:  cn.OLCForward,
			Input:  net.Bridge,
			Output: net.Bridge,
		})
		for _, port := range net.Outputs {
			w.AddOutput(net.Bridge, port.Vlan, port.Interface)
		}
	}
}

func (w *FabricWorker) downTables() {
	_ = w.ovs.delFlow(nil)
}

func (w *FabricWorker) downNetwork(bridge string, vni uint32) {
	brPort, tunPort := w.vni2peer(vni)
	if err := w.ovs.delPort(tunPort); err != nil {
		libol.Warn("FabricWorker.downNetwork %s", err)
	}
	link := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: tunPort},
		PeerName:  brPort,
	}
	_ = netlink.LinkDel(link)
}

func (w *FabricWorker) DelNetwork(bridge string, vni uint32) {
	w.downNetwork(bridge, vni)
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
	w.downTables()
	for _, tunnel := range w.spec.Tunnels {
		w.DelTunnel(tunnel.Remote)
	}
	for _, net := range w.spec.Networks {
		for _, port := range net.Outputs {
			w.DelOutput(net.Bridge, port.Vlan, port.Interface)
		}
		w.DelNetwork(net.Bridge, net.Vni)
	}
	for _, br := range w.bridges {
		_ = br.Close()
	}
}

func (w *FabricWorker) String() string {
	return w.cfg.Name
}

func (w *FabricWorker) ID() string {
	return w.uuid
}

func (w *FabricWorker) GetBridge() cn.Bridger {
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
