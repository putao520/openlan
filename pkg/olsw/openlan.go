package olsw

import (
	co "github.com/danieldin95/openlan/pkg/config"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/danieldin95/openlan/pkg/models"
	"github.com/danieldin95/openlan/pkg/network"
	"github.com/danieldin95/openlan/pkg/olsw/api"
	"github.com/danieldin95/openlan/pkg/olsw/store"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
	"time"
)

func PeerName(name, prefix string) (string, string) {
	return name + prefix + "i", name + prefix + "o"
}

type OpenLANWorker struct {
	alias     string
	cfg       *co.Network
	newTime   int64
	startTime int64
	links     *Links
	uuid      string
	bridge    network.Bridger
	out       *libol.SubLogger
	openVPN   []*OpenVPN
}

func NewOpenLANWorker(c *co.Network) *OpenLANWorker {
	return &OpenLANWorker{
		alias:     c.Alias,
		cfg:       c,
		newTime:   time.Now().Unix(),
		startTime: 0,
		links:     NewLinks(),
		out:       libol.NewSubLogger(c.Name),
	}
}

func (w *OpenLANWorker) String() string {
	return w.cfg.Name
}

func (w *OpenLANWorker) Initialize() {
	brCfg := w.cfg.Bridge
	for _, pass := range w.cfg.Password {
		user := &models.User{
			Name:     pass.Username,
			Password: pass.Password,
			Network:  w.cfg.Name,
			Role:     "admin",
		}
		user.Update()
		store.User.Add(user)
	}
	n := models.Network{
		Name:    w.cfg.Name,
		IpStart: w.cfg.Subnet.Start,
		IpEnd:   w.cfg.Subnet.End,
		Netmask: w.cfg.Subnet.Netmask,
		IfAddr:  w.cfg.Bridge.Address,
		Routes:  make([]*models.Route, 0, 2),
	}
	for _, rt := range w.cfg.Routes {
		if rt.NextHop == "" {
			w.out.Warn("OpenLANWorker.Initialize: %s noNextHop", rt.Prefix)
			continue
		}
		rte := models.NewRoute(rt.Prefix, rt.NextHop, rt.Mode)
		if rt.Metric > 0 {
			rte.Metric = rt.Metric
		}
		n.Routes = append(n.Routes, rte)
	}
	store.Network.Add(&n)
	for _, ht := range w.cfg.Hosts {
		lease := store.Network.AddLease(ht.Hostname, ht.Address)
		if lease != nil {
			lease.Type = "static"
			lease.Network = w.cfg.Name
		}
	}
	w.bridge = network.NewBridger(brCfg.Provider, brCfg.Name, brCfg.IPMtu)
	vCfg := w.cfg.OpenVPN
	if vCfg != nil {
		obj := NewOpenVPN(vCfg)
		obj.Initialize()
		w.openVPN = append(w.openVPN, obj)
		for _, _vCfg := range vCfg.Breed {
			if _vCfg == nil {
				continue
			}
			obj := NewOpenVPN(_vCfg)
			obj.Initialize()
			w.openVPN = append(w.openVPN, obj)
		}
	}
}

func (w *OpenLANWorker) ID() string {
	return w.uuid
}

func (w *OpenLANWorker) LoadLinks() {
	if w.cfg.Links != nil {
		for _, lin := range w.cfg.Links {
			lin.Default()
			w.AddLink(&lin)
		}
	}
}

func (w *OpenLANWorker) UnLoadLinks() {
	w.links.lock.RLock()
	defer w.links.lock.RUnlock()
	for _, l := range w.links.links {
		l.Stop()
	}
}

func (w *OpenLANWorker) LoadRoutes() {
	// install routes
	cfg := w.cfg
	w.out.Debug("OpenLANWorker.LoadRoute: %v", cfg.Routes)
	ifAddr := strings.SplitN(cfg.Bridge.Address, "/", 2)[0]
	link, err := netlink.LinkByName(w.bridge.Name())
	if err != nil {
		return
	}
	for _, rt := range cfg.Routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		if ifAddr == rt.NextHop && rt.MultiPath == nil {
			// route's next-hop is local not install again.
			continue
		}
		nlrt := netlink.Route{Dst: dst}
		for _, hop := range rt.MultiPath {
			nxhe := &netlink.NexthopInfo{
				Hops: hop.Weight,
				Gw:   net.ParseIP(hop.NextHop),
			}
			nlrt.MultiPath = append(nlrt.MultiPath, nxhe)
		}
		if rt.MultiPath == nil {
			nlrt.LinkIndex = link.Attrs().Index
			nlrt.Gw = net.ParseIP(rt.NextHop)
			nlrt.Priority = rt.Metric
		}
		w.out.Debug("OpenLANWorker.LoadRoute: %s", nlrt)
		promise := &libol.Promise{
			First:  time.Second * 2,
			MaxInt: time.Minute,
			MinInt: time.Second * 10,
		}
		promise.Go(func() error {
			if err := netlink.RouteAdd(&nlrt); err != nil {
				w.out.Warn("OpenLANWorker.LoadRoute: %s", err)
				return err
			}
			w.out.Info("OpenLANWorker.LoadRoute: %v", rt)
			return nil
		})
	}
}

func (w *OpenLANWorker) UnLoadRoutes() {
	cfg := w.cfg
	link, err := netlink.LinkByName(w.bridge.Name())
	if err != nil {
		return
	}
	for _, rt := range cfg.Routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		nlRt := netlink.Route{Dst: dst}
		if rt.MultiPath == nil {
			nlRt.LinkIndex = link.Attrs().Index
			nlRt.Gw = net.ParseIP(rt.NextHop)
			nlRt.Priority = rt.Metric
		}
		w.out.Debug("OpenLANWorker.UnLoadRoute: %s", nlRt)
		if err := netlink.RouteDel(&nlRt); err != nil {
			w.out.Warn("OpenLANWorker.UnLoadRoute: %s", err)
			continue
		}
		w.out.Info("OpenLANWorker.UnLoadRoute: %v", rt)
	}
}

func (w *OpenLANWorker) UpBridge(cfg *co.Bridge) {
	master := w.bridge
	// new it and configure address
	master.Open(cfg.Address)
	// configure stp
	if cfg.Stp == "on" {
		if err := master.Stp(true); err != nil {
			w.out.Warn("OpenLANWorker.UpBridge: Stp %s", err)
		}
	} else {
		_ = master.Stp(false)
	}
	// configure forward delay
	if err := master.Delay(cfg.Delay); err != nil {
		w.out.Warn("OpenLANWorker.UpBridge: Delay %s", err)
	}
	w.connectPeer(cfg)
	call := 1
	if w.cfg.Acl == "" {
		call = 0
	}
	if err := master.CallIptables(call); err != nil {
		w.out.Warn("OpenLANWorker.Start: CallIptables %s", err)
	}
}

func (w *OpenLANWorker) connectPeer(cfg *co.Bridge) {
	if cfg.Peer == "" {
		return
	}
	in, ex := PeerName(cfg.Network, "-e")
	link := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: in},
		PeerName:  ex,
	}
	br := network.NewBrCtl(cfg.Peer, cfg.IPMtu)
	promise := &libol.Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
	promise.Go(func() error {
		if !br.Has() {
			w.out.Warn("%s notFound", br.Name)
			return libol.NewErr("%s notFound", br.Name)
		}
		err := netlink.LinkAdd(link)
		if err != nil {
			w.out.Error("OpenLANWorker.connectPeer: %s", err)
			return nil
		}
		br0 := network.NewBrCtl(cfg.Name, cfg.IPMtu)
		if err := br0.AddPort(in); err != nil {
			w.out.Error("OpenLANWorker.connectPeer: %s", err)
		}
		br1 := network.NewBrCtl(cfg.Peer, cfg.IPMtu)
		if err := br1.AddPort(ex); err != nil {
			w.out.Error("OpenLANWorker.connectPeer: %s", err)
		}
		return nil
	})
}

func (w *OpenLANWorker) Start(v api.Switcher) {
	w.out.Info("OpenLANWorker.Start")
	w.UpBridge(w.cfg.Bridge)
	w.uuid = v.UUID()
	w.LoadLinks()
	w.LoadRoutes()
	for _, vpn := range w.openVPN {
		vpn.Start()
	}
	w.startTime = time.Now().Unix()
}

func (w *OpenLANWorker) downBridge(cfg *co.Bridge) {
	w.closePeer(cfg)
	_ = w.bridge.Close()
}

func (w *OpenLANWorker) closePeer(cfg *co.Bridge) {
	if cfg.Peer == "" {
		return
	}
	in, ex := PeerName(cfg.Network, "-e")
	link := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: in},
		PeerName:  ex,
	}
	err := netlink.LinkDel(link)
	if err != nil {
		w.out.Error("OpenLANWorker.closePeer: %s", err)
		return
	}
}

func (w *OpenLANWorker) Stop() {
	w.out.Info("OpenLANWorker.Close")
	for _, vpn := range w.openVPN {
		vpn.Stop()
	}
	w.UnLoadRoutes()
	w.UnLoadLinks()
	w.startTime = 0
	w.downBridge(w.cfg.Bridge)
}

func (w *OpenLANWorker) UpTime() int64 {
	if w.startTime != 0 {
		return time.Now().Unix() - w.startTime
	}
	return 0
}

func (w *OpenLANWorker) AddLink(c *co.Point) {
	br := w.cfg.Bridge
	uuid := libol.GenRandom(13)

	c.Alias = w.alias
	c.Network = w.cfg.Name
	c.Interface.Name = network.Taps.GenName()
	c.Interface.Bridge = br.Name
	c.Interface.Address = br.Address
	c.Interface.Provider = br.Provider
	c.Interface.IPMtu = br.IPMtu
	c.Log.File = "/dev/null"

	l := NewLink(uuid, c)
	l.Initialize()
	store.Link.Add(uuid, l.Model())
	w.links.Add(l)
	l.Start()
}

func (w *OpenLANWorker) DelLink(addr string) {
	if l := w.links.Remove(addr); l != nil {
		store.Link.Del(l.uuid)
	}
}

func (w *OpenLANWorker) GetSubnet() string {
	addr := ""
	cfg := w.cfg
	if cfg.Bridge.Address != "" {
		addr = cfg.Bridge.Address
	} else if cfg.Subnet.Start != "" && cfg.Subnet.Netmask != "" {
		addr = cfg.Subnet.Start + "/" + cfg.Subnet.Netmask
	}
	if addr != "" {
		if _, inet, err := net.ParseCIDR(addr); err == nil {
			return inet.String()
		}
	}
	return ""
}

func (w *OpenLANWorker) GetBridge() network.Bridger {
	return w.bridge
}

func (w *OpenLANWorker) GetConfig() *co.Network {
	return w.cfg
}

func (w *OpenLANWorker) Reload(c *co.Network) {

}
