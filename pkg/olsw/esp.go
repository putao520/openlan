package olsw

import (
	co "github.com/danieldin95/openlan/pkg/config"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/danieldin95/openlan/pkg/models"
	"github.com/danieldin95/openlan/pkg/network"
	"github.com/danieldin95/openlan/pkg/olsw/api"
	"github.com/danieldin95/openlan/pkg/olsw/store"
	"github.com/danieldin95/openlan/pkg/schema"
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

func GetStateEncap(mode string) *nl.XfrmStateEncap {
	if mode == "udp" {
		return &nl.XfrmStateEncap{
			Type:            nl.XFRM_ENCAP_ESPINUDP,
			SrcPort:         UDPPort,
			DstPort:         UDPPort,
			OriginalAddress: net.ParseIP("0.0.0.0"),
		}
	}
	return nil
}

type EspWorker struct {
	uuid     string
	cfg      *co.Network
	states   []*nl.XfrmState
	policies []*nl.XfrmPolicy
	spec     *co.ESPSpecifies
	out      *libol.SubLogger
	proto    nl.Proto
	mode     nl.Mode
}

func NewESPWorker(c *co.Network) *EspWorker {
	w := &EspWorker{
		cfg:   c,
		out:   libol.NewSubLogger(c.Name),
		proto: nl.XFRM_PROTO_ESP,
		mode:  nl.XFRM_MODE_TUNNEL,
	}
	w.spec, _ = c.Specifies.(*co.ESPSpecifies)
	return w
}

func (w *EspWorker) newState(spi int, local, remote net.IP, auth, crypt string) *nl.XfrmState {
	state := &nl.XfrmState{
		Src:   remote,
		Dst:   local,
		Proto: w.proto,
		Mode:  w.mode,
		Spi:   int(spi),
		Auth: &nl.XfrmStateAlgo{
			Name: "hmac(sha256)",
			Key:  []byte(auth),
		},
		Crypt: &nl.XfrmStateAlgo{
			Name: "cbc(aes)",
			Key:  []byte(crypt),
		},
	}
	return state
}

func (w *EspWorker) newPolicy(spi int, local, remote net.IP, src, dst *net.IPNet, dir nl.Dir) *nl.XfrmPolicy {
	policy := &nl.XfrmPolicy{
		Src: src,
		Dst: dst,
		Dir: dir,
	}
	tmpl := nl.XfrmPolicyTmpl{
		Src:   local,
		Dst:   remote,
		Proto: w.proto,
		Mode:  w.mode,
		Spi:   int(spi),
	}
	policy.Tmpls = append(policy.Tmpls, tmpl)
	return policy
}

func (w *EspWorker) addState(mem *co.ESPMember) {
	spi := mem.Spi
	local := mem.State.LocalIp
	remote := mem.State.RemoteIp
	auth := mem.State.Auth
	crypt := mem.State.Crypt
	encap := GetStateEncap(mem.State.Encap)

	w.out.Info("EspWorker.addState %s %s", local, remote)
	if st := w.newState(spi, local, remote, auth, crypt); st != nil {
		st.Encap = encap
		w.states = append(w.states, st)
	}
	if st := w.newState(spi, remote, local, auth, crypt); st != nil {
		st.Encap = encap
		w.states = append(w.states, st)
	}
	store.EspState.Add(&models.EspState{
		EspState: &schema.EspState{
			Name:   w.spec.Name,
			Spi:    spi,
			Source: local.String(),
			Dest:   remote.String(),
			Proto:  uint8(w.proto),
			Mode:   uint8(w.mode),
		},
	})
}

func (w *EspWorker) delState(mem *co.ESPMember) {
	spi := mem.Spi
	local := mem.State.LocalIp
	remote := mem.State.RemoteIp

	w.out.Info("EspWorker.delState %s %s", local, remote)
	model := models.EspState{
		EspState: &schema.EspState{
			Name:   w.spec.Name,
			Spi:    spi,
			Source: local.String(),
			Dest:   remote.String(),
			Proto:  uint8(w.proto),
			Mode:   uint8(w.mode),
		},
	}
	store.EspState.Del(model.ID())
}

func (w *EspWorker) addPolicy(mem *co.ESPMember, pol *co.ESPPolicy) {
	spi := mem.Spi
	local := mem.State.LocalIp
	remote := mem.State.RemoteIp
	w.out.Info("EspWorker.addPolicy %s %s %s %s", local, remote, pol.Source, pol.Dest)
	src, err := libol.ParseNet(pol.Source)
	if err != nil {
		w.out.Error("EspWorker.addPolicy %s %s", pol.Source, err)
		return
	}
	dst, err := libol.ParseNet(pol.Dest)
	if err != nil {
		w.out.Error("EspWorker.addPolicy %s %s", pol.Dest, err)
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
	store.EspPolicy.Add(&models.EspPolicy{
		EspPolicy: &schema.EspPolicy{
			Spi:    spi,
			Name:   w.spec.Name,
			Source: pol.Source,
			Dest:   pol.Dest,
		},
	})
}

func (w *EspWorker) delPolicy(mem *co.ESPMember, pol *co.ESPPolicy) {
	spi := mem.Spi
	local := mem.State.LocalIp
	remote := mem.State.RemoteIp
	w.out.Info("EspWorker.delPolicy %s %s %s %s", local, remote, pol.Source, pol.Dest)
	obj := models.EspPolicy{
		EspPolicy: &schema.EspPolicy{
			Spi:    spi,
			Name:   w.spec.Name,
			Source: pol.Source,
			Dest:   pol.Dest,
		},
	}
	store.EspPolicy.Del(obj.ID())
}

func (w *EspWorker) updateXfrm() {
	w.states = nil
	w.policies = nil

	store.EspState.Clear()
	store.EspState.Clear()
	for _, mem := range w.spec.Members {
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
			if pol.Dest == "" {
				pol.Dest = mem.Peer
			}
			if pol.Source == "" {
				pol.Source = mem.Address
			}
			w.addPolicy(mem, pol)
		}
	}
}

func (w *EspWorker) Initialize() {
	if w.spec == nil {
		w.out.Error("EspWorker.Initialize spec is nil")
		return
	}
	w.updateXfrm()
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

func (w *EspWorker) addXfrm() {
	for _, state := range w.states {
		w.out.Debug("EspWorker.AddXfrm State %s", state)
		if err := nl.XfrmStateAdd(state); err != nil {
			w.out.Error("EspWorker.Start State %s", err)
		}
	}
	for _, policy := range w.policies {
		if err := nl.XfrmPolicyAdd(policy); err != nil {
			w.out.Error("EspWorker.addXfrm Policy %s", err)
		}
	}
}

func (w *EspWorker) Start(v api.Switcher) {
	if w.spec == nil {
		w.out.Error("EspWorker.Start spec is nil")
		return
	}
	w.uuid = v.UUID()
	w.addXfrm()
	for _, mem := range w.spec.Members {
		if err := w.UpDummy(mem.Name, mem.Address, mem.Peer); err != nil {
			w.out.Error("EspWorker.Start %s %s", mem.Name, err)
		}
	}
	store.Esp.Add(&models.Esp{
		Name:    w.cfg.Name,
		Address: w.spec.Address,
	})
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

func (w *EspWorker) delXfrm() {
	for _, state := range w.states {
		if err := nl.XfrmStateDel(state); err != nil {
			w.out.Warn("EspWorker.delXfrm State %s", err)
		}
	}
	for _, policy := range w.policies {
		if err := nl.XfrmPolicyDel(policy); err != nil {
			w.out.Warn("EspWorker.delXfrm Policy %s", err)
		}
	}
}

func (w *EspWorker) Stop() {
	if w.spec == nil {
		w.out.Error("EspWorker.Stop spec is nil")
		return
	}
	for _, mem := range w.spec.Members {
		if err := w.DownDummy(mem.Name); err != nil {
			w.out.Error("EspWorker.Stop %s %s", mem.Name, err)
		}
	}
	w.delXfrm()
}

func (w *EspWorker) String() string {
	return w.cfg.Name
}

func (w *EspWorker) ID() string {
	return w.uuid
}

func (w *EspWorker) GetBridge() network.Bridger {
	w.out.Warn("EspWorker.GetBridge notSupport")
	return nil
}

func (w *EspWorker) GetConfig() *co.Network {
	return w.cfg
}

func (w *EspWorker) GetSubnet() string {
	w.out.Warn("EspWorker.GetSubnet notSupport")
	return ""
}

func (w *EspWorker) Reload(c *co.Network) {
	w.delXfrm()
	w.updateXfrm()
	w.addXfrm()
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
