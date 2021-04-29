package olsw

import (
	"github.com/danieldin95/openlan-go/src/config"
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
	cfg      *config.Network
	states   []*nl.XfrmState
	policies []*nl.XfrmPolicy
	inCfg    *config.ESPInterface
	out      *libol.SubLogger
}

func NewESPWorker(c *config.Network) *EspWorker {
	w := &EspWorker{
		cfg:      c,
		states:   make([]*nl.XfrmState, 0, 4),
		policies: make([]*nl.XfrmPolicy, 0, 32),
		out:      libol.NewSubLogger(c.Name),
	}
	w.inCfg, _ = c.Interface.(*config.ESPInterface)
	return w
}

func (w *EspWorker) newState(spi uint32, local, remote net.IP, auth, crypt string) *nl.XfrmState {
	return &nl.XfrmState{
		Src:   remote,
		Dst:   local,
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

func (w *EspWorker) newPolicy(spi uint32, local, remote net.IP, src, dst *net.IPNet, dir nl.Dir) *nl.XfrmPolicy {
	policy := &nl.XfrmPolicy{
		Src: src,
		Dst: dst,
		Dir: dir,
	}
	tmpl := nl.XfrmPolicyTmpl{
		Src:   local,
		Dst:   remote,
		Proto: nl.XFRM_PROTO_ESP,
		Mode:  nl.XFRM_MODE_TUNNEL,
		Spi:   int(spi),
	}
	policy.Tmpls = append(policy.Tmpls, tmpl)
	return policy
}

func (w *EspWorker) addState(mem *config.ESPMember) {
	spi := mem.Spi
	local := mem.State.LocalIp
	remote := mem.State.RemoteIp
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
	local := mem.State.LocalIp
	remote := mem.State.RemoteIp
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
	if po := w.newPolicy(spi, local, remote, src, dst, nl.XFRM_DIR_OUT); po != nil {
		w.policies = append(w.policies, po)
	}
	if po := w.newPolicy(spi, remote, local, dst, src, nl.XFRM_DIR_IN); po != nil {
		w.policies = append(w.policies, po)
	}
	if po := w.newPolicy(spi, remote, local, dst, src, nl.XFRM_DIR_FWD); po != nil {
		w.policies = append(w.policies, po)
	}
}

func (w *EspWorker) Initialize() {
	if w.inCfg == nil {
		w.out.Error("EspWorker.Initialize inCfg is nil")
		return
	}
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
	if w.inCfg == nil {
		w.out.Error("EspWorker.Start inCfg is nil")
		return
	}
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
	if w.inCfg == nil {
		w.out.Error("EspWorker.Stop inCfg is nil")
		return
	}
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
