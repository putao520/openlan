package olsw

import (
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olap"
	"github.com/danieldin95/openlan-go/src/olsw/api"
	"github.com/danieldin95/openlan-go/src/olsw/store"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
	"sync"
	"time"
)

func PeerName(name, prefix string) (string, string) {
	return name + prefix + "i", name + prefix + "o"
}

type OLANWorker struct {
	alias     string
	cfg       *config.Network
	newTime   int64
	startTime int64
	linksLock sync.RWMutex
	links     map[string]*olap.Point
	uuid      string
	crypt     *config.Crypt
	bridge    network.Bridger
	out       *libol.SubLogger
	openVPN   *OpenVPN
}

func NewOLANWorker(c *config.Network) *OLANWorker {
	return &OLANWorker{
		alias:     c.Alias,
		cfg:       c,
		newTime:   time.Now().Unix(),
		startTime: 0,
		links:     make(map[string]*olap.Point),
		crypt:     c.Crypt,
		out:       libol.NewSubLogger(c.Name),
	}
}

func (w *OLANWorker) String() string {
	return w.cfg.Name
}

func (w *OLANWorker) Initialize() {
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
			w.out.Warn("OLANWorker.Initialize: %s noNextHop", rt.Prefix)
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
	w.bridge = network.NewBridger(brCfg.Provider, brCfg.Name, brCfg.IfMtu)
	if w.cfg.OpenVPN != nil {
		w.openVPN = NewOpenVPN(w.cfg.OpenVPN)
		w.openVPN.Initialize()
	}
}

func (w *OLANWorker) ID() string {
	return w.uuid
}

func (w *OLANWorker) LoadLinks() {
	if w.cfg.Links != nil {
		for _, lin := range w.cfg.Links {
			lin.Default()
			w.AddLink(lin)
		}
	}
}

func (w *OLANWorker) UnLoadLinks() {
	for _, p := range w.links {
		p.Stop()
	}
}

func (w *OLANWorker) LoadRoutes() {
	// install routes
	w.out.Debug("OLANWorker.LoadRoute: %v", w.cfg.Routes)
	ifAddr := strings.SplitN(w.cfg.Bridge.Address, "/", 2)[0]
	link, err := netlink.LinkByName(w.bridge.Name())
	if ifAddr == "" || err != nil {
		return
	}
	for _, rt := range w.cfg.Routes {
		if ifAddr == rt.NextHop { // route's next-hop is local not install again.
			continue
		}
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		next := net.ParseIP(rt.NextHop)
		rte := netlink.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       dst,
			Gw:        next,
			Priority:  rt.Metric,
		}
		w.out.Debug("OLANWorker.LoadRoute: %s", rte)
		if err := netlink.RouteAdd(&rte); err != nil {
			w.out.Warn("OLANWorker.LoadRoute: %s", err)
			continue
		}
		w.out.Info("OLANWorker.LoadRoute: %v", rt)
	}
}

func (w *OLANWorker) UnLoadRoutes() {
	link, err := netlink.LinkByName(w.bridge.Name())
	if w.cfg.Bridge.Address == "" || err != nil {
		return
	}
	for _, rt := range w.cfg.Routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		next := net.ParseIP(rt.NextHop)
		rte := netlink.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       dst,
			Gw:        next,
		}
		w.out.Debug("OLANWorker.UnLoadRoute: %s", rte)
		if err := netlink.RouteDel(&rte); err != nil {
			w.out.Warn("OLANWorker.UnLoadRoute: %s", err)
			continue
		}
		w.out.Info("OLANWorker.UnLoadRoute: %v", rt)
	}
}

func (w *OLANWorker) UpBridge(cfg config.Bridge) {
	master := w.bridge
	// new it and configure address
	master.Open(cfg.Address)
	// configure stp
	if cfg.Stp == "on" {
		if err := master.Stp(true); err != nil {
			w.out.Warn("OLANWorker.UpBridge: Stp %s", err)
		}
	} else {
		_ = master.Stp(false)
	}
	// configure forward delay
	if err := master.Delay(cfg.Delay); err != nil {
		w.out.Warn("OLANWorker.UpBridge: Delay %s", err)
	}
	w.ConnectPeer(cfg)
}

func (w *OLANWorker) ConnectPeer(cfg config.Bridge) {
	if cfg.Peer == "" {
		return
	}
	in, ex := PeerName(cfg.Network, "-e")
	link := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: in},
		PeerName:  ex,
	}
	br := network.NewBrCtl(cfg.Peer)
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
			w.out.Error("OLANWorker.ConnectPeer: %s", err)
			return nil
		}
		br0 := network.NewBrCtl(cfg.Name)
		if err := br0.AddPort(in); err != nil {
			w.out.Error("OLANWorker.ConnectPeer: %s", err)
		}
		br1 := network.NewBrCtl(cfg.Peer)
		if err := br1.AddPort(ex); err != nil {
			w.out.Error("OLANWorker.ConnectPeer: %s", err)
		}
		return nil
	})
}

func (w *OLANWorker) Start(v api.Switcher) {
	w.out.Info("OLANWorker.Start")
	brCfg := w.cfg.Bridge
	w.UpBridge(brCfg)
	call := 1
	if w.cfg.Acl == "" {
		call = 0
	}
	if err := w.bridge.CallIptables(call); err != nil {
		w.out.Warn("OLANWorker.Start: CallIptables %s", err)
	}
	w.uuid = v.UUID()
	w.LoadLinks()
	w.LoadRoutes()
	if w.openVPN != nil {
		w.openVPN.Start()
	}
	w.startTime = time.Now().Unix()
}

func (w *OLANWorker) DownBridge(cfg config.Bridge) {
	w.ClosePeer(cfg)
	_ = w.bridge.Close()
}

func (w *OLANWorker) ClosePeer(cfg config.Bridge) {
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
		w.out.Error("OLANWorker.ClosePeer: %s", err)
		return
	}
}

func (w *OLANWorker) Stop() {
	w.out.Info("OLANWorker.Close")
	if w.openVPN != nil {
		w.openVPN.Stop()
	}
	w.UnLoadRoutes()
	w.UnLoadLinks()
	w.startTime = 0
	w.DownBridge(w.cfg.Bridge)
}

func (w *OLANWorker) UpTime() int64 {
	if w.startTime != 0 {
		return time.Now().Unix() - w.startTime
	}
	return 0
}

func (w *OLANWorker) AddLink(c *config.Point) {
	brName := w.cfg.Bridge.Name
	c.Alias = w.alias
	c.RequestAddr = false
	c.Network = w.cfg.Name
	c.Interface.Name = "auto"
	c.Interface.Bridge = brName // reset bridge name.
	c.Interface.Address = w.cfg.Bridge.Address
	c.Interface.Provider = w.cfg.Bridge.Provider
	libol.Go(func() {
		p := olap.NewPoint(c)
		p.Initialize()
		w.linksLock.Lock()
		w.links[c.Connection] = p
		w.linksLock.Unlock()
		store.Link.Add(p)
		p.Start()
	})
}

func (w *OLANWorker) DelLink(addr string) {
	w.linksLock.Lock()
	defer w.linksLock.Unlock()
	if p, ok := w.links[addr]; ok {
		p.Stop()
		store.Link.Del(p.UUID())
		delete(w.links, addr)
	}
}

func (w *OLANWorker) GetSubnet() string {
	addr := ""
	if w.cfg.Bridge.Address != "" {
		addr = w.cfg.Bridge.Address
	} else if w.cfg.Subnet.Start != "" && w.cfg.Subnet.Netmask != "" {
		addr = w.cfg.Subnet.Start + "/" + w.cfg.Subnet.Netmask
	}
	if addr != "" {
		if _, inet, err := net.ParseCIDR(addr); err == nil {
			return inet.String()
		}
	}
	return ""
}

func (w *OLANWorker) GetBridge() network.Bridger {
	return w.bridge
}

func (w *OLANWorker) GetConfig() *config.Network {
	return w.cfg
}
