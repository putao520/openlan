package olsw

import (
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olsw/api"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

const (
	UDPPort = 4500
	UDPBin  = "/usr/bin/openudp"
)

type EspWorker struct {
	uuid     string
	cfg      *config.Network
	states   []*netlink.XfrmState
	policies []*netlink.XfrmPolicy
	inCfg    *config.ESPInterface
	out      *libol.SubLogger
}

func NewESPWorker(c *config.Network) *EspWorker {
	w := &EspWorker{
		cfg:      c,
		states:   make([]*netlink.XfrmState, 0, 4),
		policies: make([]*netlink.XfrmPolicy, 0, 32),
		out:      libol.NewSubLogger(c.Name),
	}
	w.inCfg, _ = c.Interface.(*config.ESPInterface)
	return w
}

func (w *EspWorker) newState(spi uint32, src, dst, auth, crypt string) *netlink.XfrmState {
	return &netlink.XfrmState{
		Src:   net.ParseIP(dst).To4(),
		Dst:   net.ParseIP(src).To4(),
		Proto: netlink.XFRM_PROTO_ESP,
		Mode:  netlink.XFRM_MODE_TUNNEL,
		Spi:   int(spi),
		Auth: &netlink.XfrmStateAlgo{
			Name: "hmac(sha256)",
			Key:  []byte(auth),
		},
		Crypt: &netlink.XfrmStateAlgo{
			Name: "cbc(aes)",
			Key:  []byte(crypt),
		},
		Encap: &netlink.XfrmStateEncap{
			Type:            netlink.XFRM_ENCAP_ESPINUDP,
			SrcPort:         UDPPort,
			DstPort:         UDPPort,
			OriginalAddress: net.ParseIP("0.0.0.0").To4(),
		},
	}
}

func (w *EspWorker) newPolicy(spi uint32, local, remote string,
	src, dst *net.IPNet, dir netlink.Dir) *netlink.XfrmPolicy {
	policy := &netlink.XfrmPolicy{
		Src: src,
		Dst: dst,
		Dir: dir,
	}
	tmpl := netlink.XfrmPolicyTmpl{
		Src:   net.ParseIP(local),
		Dst:   net.ParseIP(remote),
		Proto: netlink.XFRM_PROTO_ESP,
		Mode:  netlink.XFRM_MODE_TUNNEL,
		Spi:   int(spi),
	}
	policy.Tmpls = append(policy.Tmpls, tmpl)
	return policy
}

func (w *EspWorker) addState(mem *config.ESPMember) {
	spi := mem.Spi
	local := mem.State.Local
	remote := mem.State.Remote
	auth := mem.State.Auth
	crypt := mem.State.Crypt
	if st := w.newState(spi, local, remote, auth, crypt); st != nil {
		w.states = append(w.states, st)
	}
	if st := w.newState(spi, remote, local, auth, crypt); st != nil {
		w.states = append(w.states, st)
	}
}

func (w *EspWorker) addPolicy(mem *config.ESPMember, pol *config.ESPPolicy) {
	spi := mem.Spi
	local := mem.State.Local
	remote := mem.State.Remote

	src, err := libol.ParseNet(pol.Source)
	if err != nil {
		w.out.Error("EspWorker.addPolicy %s", err)
		return
	}
	dst, err := libol.ParseNet(pol.Destination)
	if err != nil {
		w.out.Error("EspWorker.addPolicy %s", err)
		return
	}
	if po := w.newPolicy(spi, local, remote, src, dst, netlink.XFRM_DIR_OUT); po != nil {
		w.policies = append(w.policies, po)
	}
	if po := w.newPolicy(spi, remote, local, dst, src, netlink.XFRM_DIR_IN); po != nil {
		w.policies = append(w.policies, po)
	}
	if po := w.newPolicy(spi, remote, local, dst, src, netlink.XFRM_DIR_FWD); po != nil {
		w.policies = append(w.policies, po)
	}
}

func (w *EspWorker) Initialize() {
	for _, mem := range w.inCfg.Members {
		if mem == nil {
			continue
		}
		// add xfrm states
		w.addState(mem)
		// add xfrm policies
		for _, pol := range mem.Policies {
			if pol == nil {
				continue
			}
			w.addPolicy(mem, pol)
		}
	}
}

func (w *EspWorker) UpDummy(name, addr, peer string) error {
	link, _ := netlink.LinkByName(name)
	if link == nil {
		port := &netlink.Dummy{
			LinkAttrs: netlink.LinkAttrs{
				TxQLen: -1,
				Name:   name,
			},
		}
		if err := netlink.LinkAdd(port); err != nil {
			return err
		}
		link, _ = netlink.LinkByName(name)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		w.out.Error("EspWorker.UpDummy: %s", err)
	}
	w.out.Info("EspWorker.Open %s success", name)
	if addr != "" {
		ipAddr, err := netlink.ParseAddr(addr)
		if err != nil {
			return err
		}
		if err := netlink.AddrAdd(link, ipAddr); err != nil {
			return err
		}
	}
	// add peer routes.
	_, dst, err := net.ParseCIDR(peer)
	if err != nil {
		return err
	}
	ip := strings.SplitN(addr, "/", 2)[0]
	next := net.ParseIP(ip)
	rte := netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
		Gw:        next,
	}
	w.out.Debug("EspWorker.AddRoute: %s", rte)
	if err := netlink.RouteAdd(&rte); err != nil {
		return err
	}
	return nil
}

func (w *EspWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	for _, state := range w.states {
		w.out.Debug("EspWorker.Start State %s", state)
		if err := netlink.XfrmStateAdd(state); err != nil {
			w.out.Error("EspWorker.Start State %s", err)
		}
	}
	for _, policy := range w.policies {
		if err := netlink.XfrmPolicyAdd(policy); err != nil {
			w.out.Error("EspWorker.Start Policy %s", err)
		}
	}
	for _, mem := range w.inCfg.Members {
		if err := w.UpDummy(mem.Name, mem.Address, mem.Peer); err != nil {
			w.out.Error("EspWorker.Start %s %s", mem.Name, err)
		}
	}
}

func (w *EspWorker) DownDummy(name string) error {
	link, _ := netlink.LinkByName(name)
	if link == nil {
		return nil
	}
	port := &netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{
			TxQLen: -1,
			Name:   name,
		},
	}
	if err := netlink.LinkDel(port); err != nil {
		return err
	}
	return nil
}

func (w *EspWorker) Stop() {
	for _, mem := range w.inCfg.Members {
		if err := w.DownDummy(mem.Name); err != nil {
			w.out.Error("EspWorker.Stop %s dummy %s", mem.Name, err)
		}
	}
	for _, state := range w.states {
		if err := netlink.XfrmStateDel(state); err != nil {
			w.out.Error("EspWorker.Stop State %s", err)
		}
	}
	for _, policy := range w.policies {
		if err := netlink.XfrmPolicyDel(policy); err != nil {
			w.out.Error("EspWorker.Stop Policy %s", err)
		}
	}
}

func (w *EspWorker) String() string {
	return w.cfg.Name
}

func (w *EspWorker) ID() string {
	return w.uuid
}

func (w *EspWorker) GetBridge() network.Bridger {
	w.out.Warn("EspWorker.GetBridge operation notSupport")
	return nil
}

func (w *EspWorker) GetConfig() *config.Network {
	return w.cfg
}

func (w *EspWorker) GetSubnet() string {
	w.out.Warn("EspWorker.GetSubnet operation notSupport")
	return ""
}

func OpenUDP() {
	libol.Go(func() {
		args := []string{strconv.Itoa(UDPPort)}
		cmd := exec.Command(UDPBin, args...)
		if err := cmd.Run(); err != nil {
			libol.Error("esp.init %s", err)
		}
	})
}
