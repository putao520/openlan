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
)

const (
	UDPPort = 4500
	UDPBin  = "/usr/bin/udp-4500"
)

type EspWorker struct {
	netCfg   *config.Network
	states   []*netlink.XfrmState
	policies []*netlink.XfrmPolicy
	cfg      *config.ESPInterface
	out      *libol.SubLogger
}

func NewESPWorker(c *config.Network) *EspWorker {
	w := &EspWorker{
		netCfg:   c,
		states:   make([]*netlink.XfrmState, 0, 4),
		policies: make([]*netlink.XfrmPolicy, 0, 32),
		out:      libol.NewSubLogger(c.Name),
	}
	w.cfg, _ = c.Interface.(*config.ESPInterface)
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
	local := mem.State.Private
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
	local := mem.State.Private
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
	if po := w.newPolicy(spi, remote, local, src, dst, netlink.XFRM_DIR_IN); po != nil {
		w.policies = append(w.policies, po)
	}
	if po := w.newPolicy(spi, remote, local, src, dst, netlink.XFRM_DIR_FWD); po != nil {
		w.policies = append(w.policies, po)
	}
}

func (w *EspWorker) Initialize() {
	for _, mem := range w.cfg.Members {
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

func (w *EspWorker) Start(v api.Switcher) {
	for _, state := range w.states {
		libol.Info("EspWorker.Start State %s", state)
		if err := netlink.XfrmStateAdd(state); err != nil {
			libol.Error("EspWorker.Start State %s", err)
		}
	}
	for _, policy := range w.policies {
		if err := netlink.XfrmPolicyAdd(policy); err != nil {
			libol.Error("EspWorker.Start Policy %s", err)
		}
	}
}

func (w *EspWorker) Stop() {
	for _, state := range w.states {
		if err := netlink.XfrmStateDel(state); err != nil {
			libol.Error("EspWorker.Stop State %s", err)
		}
	}
	for _, policy := range w.policies {
		if err := netlink.XfrmPolicyDel(policy); err != nil {
			libol.Error("EspWorker.Stop Policy %s", err)
		}
	}
}

func (w *EspWorker) String() string {
	return ""
}

func (w *EspWorker) ID() string {
	return ""
}

func (w *EspWorker) GetBridge() network.Bridger {
	return nil
}

func (w *EspWorker) GetConfig() *config.Network {
	return w.netCfg
}

func (w *EspWorker) GetSubnet() string {
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
