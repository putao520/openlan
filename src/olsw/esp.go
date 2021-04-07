package olsw

import (
	co "github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olsw/api"
	nl "github.com/vishvananda/netlink"
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
	cfg      *co.Network
	states   []*nl.XfrmState
	policies []*nl.XfrmPolicy
	inCfg    *co.ESPInterface
	out      *libol.SubLogger
}

func NewESPWorker(c *co.Network) *EspWorker {
	w := &EspWorker{
		cfg:      c,
		states:   make([]*nl.XfrmState, 0, 4),
		policies: make([]*nl.XfrmPolicy, 0, 32),
		out:      libol.NewSubLogger(c.Name),
	}
	w.inCfg, _ = c.Interface.(*co.ESPInterface)
	return w
}

func (w *EspWorker) newState(mem *co.ESPMember) *nl.XfrmState {
	spi := mem.Spi
	local := mem.State.LocalIp
	remote := mem.State.RemoteIp
	auth := mem.State.Auth
	crypt := mem.State.Crypt

	return &nl.XfrmState{
		Src:   local,
		Dst:   remote,
		Proto: nl.XFRM_PROTO_ESP,
		Mode:  nl.XFRM_MODE_TUNNEL,
		Spi:   int(spi),
		Auth: &nl.XfrmStateAlgo{
			Name: "hmac(sha256)",
			Key:  []byte(auth),
		},
		Crypt: &nl.XfrmStateAlgo{
			Name: "cbc(aes)",
			Key:  []byte(crypt),
		},
		Encap: &nl.XfrmStateEncap{
			Type:            nl.XFRM_ENCAP_ESPINUDP,
			SrcPort:         UDPPort,
			DstPort:         UDPPort,
			OriginalAddress: net.ParseIP("0.0.0.0"),
		},
	}
}

func (w *EspWorker) newPolicy(mem *co.ESPMember, src, dst *net.IPNet, dir nl.Dir) *nl.XfrmPolicy {
	spi := mem.Spi
	local := mem.State.LocalIp
	remote := mem.State.RemoteIp

	policy := &nl.XfrmPolicy{
		Src: src,
		Dst: dst,
		Dir: dir,
	}
	policy.Tmpls = append(policy.Tmpls, nl.XfrmPolicyTmpl{
		Src:   local,
		Dst:   remote,
		Proto: nl.XFRM_PROTO_ESP,
		Mode:  nl.XFRM_MODE_TUNNEL,
		Spi:   int(spi),
	})
	return policy
}

func (w *EspWorker) addState(mem *co.ESPMember) {
	if st := w.newState(mem); st != nil {
		w.states = append(w.states, st)
	}
	if st := w.newState(mem); st != nil {
		w.states = append(w.states, st)
	}
}

func (w *EspWorker) addPolicy(mem *co.ESPMember, pol *co.ESPPolicy) {
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
	if po := w.newPolicy(mem, src, dst, nl.XFRM_DIR_OUT); po != nil {
		w.policies = append(w.policies, po)
	}
	if po := w.newPolicy(mem, dst, src, nl.XFRM_DIR_IN); po != nil {
		w.policies = append(w.policies, po)
	}
	if po := w.newPolicy(mem, dst, src, nl.XFRM_DIR_FWD); po != nil {
		w.policies = append(w.policies, po)
	}
}

func (w *EspWorker) Initialize() {
	for _, mem := range w.inCfg.Members {
		if mem == nil {
			continue
		}
		local, _ := net.LookupIP(mem.State.Local)
		if local == nil {
			continue
		}
		remote, _ := net.LookupIP(mem.State.Remote)
		if remote == nil {
			continue
		}
		mem.State.LocalIp = local[0]
		mem.State.RemoteIp = remote[0]
		w.addState(mem)
		for _, pol := range mem.Policies {
			if pol == nil {
				continue
			}
			w.addPolicy(mem, pol)
		}
	}
}

func (w *EspWorker) UpDummy(name, addr, peer string) error {
	link, _ := nl.LinkByName(name)
	if link == nil {
		port := &nl.Dummy{
			LinkAttrs: nl.LinkAttrs{
				TxQLen: -1,
				Name:   name,
			},
		}
		if err := nl.LinkAdd(port); err != nil {
			return err
		}
		link, _ = nl.LinkByName(name)
	}
	if err := nl.LinkSetUp(link); err != nil {
		w.out.Error("EspWorker.UpDummy: %s", err)
	}
	w.out.Info("EspWorker.Open %s success", name)
	if addr != "" {
		ipAddr, err := nl.ParseAddr(addr)
		if err != nil {
			return err
		}
		if err := nl.AddrAdd(link, ipAddr); err != nil {
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
	rte := nl.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
		Gw:        next,
	}
	w.out.Debug("EspWorker.AddRoute: %s", rte)
	if err := nl.RouteAdd(&rte); err != nil {
		return err
	}
	return nil
}

func (w *EspWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	for _, state := range w.states {
		w.out.Debug("EspWorker.Start State %s", state)
		if err := nl.XfrmStateAdd(state); err != nil {
			w.out.Error("EspWorker.Start State %s", err)
		}
	}
	for _, policy := range w.policies {
		if err := nl.XfrmPolicyAdd(policy); err != nil {
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
	link, _ := nl.LinkByName(name)
	if link == nil {
		return nil
	}
	port := &nl.Dummy{
		LinkAttrs: nl.LinkAttrs{
			TxQLen: -1,
			Name:   name,
		},
	}
	if err := nl.LinkDel(port); err != nil {
		return err
	}
	return nil
}

func (w *EspWorker) Stop() {
	for _, mem := range w.inCfg.Members {
		if err := w.DownDummy(mem.Name); err != nil {
			w.out.Error("EspWorker.Stop %s %s", mem.Name, err)
		}
	}
	for _, state := range w.states {
		if err := nl.XfrmStateDel(state); err != nil {
			w.out.Error("EspWorker.Stop State %s", err)
		}
	}
	for _, policy := range w.policies {
		if err := nl.XfrmPolicyDel(policy); err != nil {
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

func (w *EspWorker) GetConfig() *co.Network {
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
