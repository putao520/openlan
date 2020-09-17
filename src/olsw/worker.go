package olsw

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olap"
	"github.com/danieldin95/openlan-go/src/olsw/api"
	"github.com/danieldin95/openlan-go/src/olsw/storage"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
	"sync"
	"time"
)

type NetworkWorker struct {
	alias     string
	cfg       config.Network
	newTime   int64
	startTime int64
	linksLock sync.RWMutex
	links     map[string]*olap.Point
	uuid      string
	crypt     *config.Crypt
	bridge    network.Bridger
	out       *libol.SubLogger
}

func NewNetworkWorker(c config.Network, crypt *config.Crypt) *NetworkWorker {
	return &NetworkWorker{
		alias:     c.Alias,
		cfg:       c,
		newTime:   time.Now().Unix(),
		startTime: 0,
		links:     make(map[string]*olap.Point),
		crypt:     crypt,
		out:       libol.NewSubLogger(c.Name),
	}
}

func (w *NetworkWorker) String() string {
	return w.cfg.Name
}

func (w *NetworkWorker) Initialize() {
	brCfg := w.cfg.Bridge
	for _, pass := range w.cfg.Password {
		user := models.User{
			Name:     pass.Username + "@" + w.cfg.Name,
			Password: pass.Password,
		}
		storage.User.Add(&user)
	}
	n := models.Network{
		Name:    w.cfg.Name,
		IpStart: w.cfg.Subnet.Start,
		IpEnd:   w.cfg.Subnet.End,
		Netmask: w.cfg.Subnet.Netmask,
		Routes:  make([]*models.Route, 0, 2),
	}
	for _, rt := range w.cfg.Routes {
		if rt.NextHop == "" {
			w.out.Warn("NetworkWorker.Initialize: %s noNextHop", rt.Prefix)
			continue
		}
		n.Routes = append(n.Routes, &models.Route{
			Prefix:  rt.Prefix,
			NextHop: rt.NextHop,
		})
	}
	storage.Network.Add(&n)
	for _, ht := range w.cfg.Hosts {
		lease := storage.Network.AddLease(ht.Hostname, ht.Address)
		if lease != nil {
			lease.Type = "static"
			lease.Network = w.cfg.Name
		}
	}
	w.bridge = network.NewBridger(brCfg.Provider, brCfg.Name, brCfg.IfMtu)
}

func (w *NetworkWorker) ID() string {
	return w.uuid
}

func (w *NetworkWorker) LoadLinks() {
	if w.cfg.Links != nil {
		for _, lin := range w.cfg.Links {
			lin.Default()
			w.AddLink(lin)
		}
	}
}

func (w *NetworkWorker) UnLoadLinks() {
	for _, p := range w.links {
		p.Stop()
	}
}

func (w *NetworkWorker) LoadRoutes() {
	// install routes
	w.out.Debug("NetworkWorker.LoadRoute: %v", w.cfg.Routes)
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
			Dst:       dst, Gw: next,
			Priority: rt.Metric,
		}
		w.out.Debug("NetworkWorker.LoadRoute: %s", rte)
		if err := netlink.RouteAdd(&rte); err != nil {
			w.out.Warn("NetworkWorker.LoadRoute: %s", err)
			continue
		}
		w.out.Info("NetworkWorker.LoadRoute: %v", rt)
	}
}

func (w *NetworkWorker) UnLoadRoutes() {
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
		w.out.Debug("NetworkWorker.UnLoadRoute: %s", rte)
		if err := netlink.RouteDel(&rte); err != nil {
			w.out.Warn("NetworkWorker.UnLoadRoute: %s", err)
			continue
		}
		w.out.Info("NetworkWorker.UnLoadRoute: %v", rt)
	}
}

func (w *NetworkWorker) UpBridge(cfg config.Bridge) {
	w.bridge.Open(cfg.Address)
	if cfg.Stp == "on" {
		if err := w.bridge.Stp(true); err != nil {
			w.out.Warn("NetworkWorker.Start: Stp %s", err)
		}
	} else {
		_ = w.bridge.Stp(false)
	}
	if err := w.bridge.Delay(cfg.Delay); err != nil {
		w.out.Warn("NetworkWorker.Start: Delay %s", err)
	}
}

func (w *NetworkWorker) UpPeer(cfg config.Bridge) {
	if cfg.Peer == "" {
		return
	}
	in := cfg.Name + "-in"
	ex := cfg.Name + "-ex"
	link := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: in},
		PeerName:  ex,
	}
	br := libol.NewBrCtl(cfg.Peer)
	promise := &libol.Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
	promise.Go(func() error {
		if !br.Has() {
			w.out.Warn("%s notFound", br.Name)
			return libol.NewErr("notFound")
		}
		err := netlink.LinkAdd(link)
		if err != nil {
			w.out.Error("NetworkWorker.UpPeer: %s", err)
			return nil
		}
		bridge := libol.NewBrCtl(cfg.Name)
		if err := bridge.AddPort(in); err != nil {
			w.out.Error("NetworkWorker.UpPeer: %s", err)
		}
		peer := libol.NewBrCtl(cfg.Peer)
		if err := peer.AddPort(ex); err != nil {
			w.out.Error("NetworkWorker.UpPeer: %s", err)
		}
		return nil
	})
}

func (w *NetworkWorker) Start(v api.Switcher) {
	w.out.Info("NetworkWorker.Start")
	brCfg := w.cfg.Bridge
	w.UpBridge(brCfg)
	w.UpPeer(brCfg)
	w.uuid = v.UUID()
	w.startTime = time.Now().Unix()
	w.LoadLinks()
	w.LoadRoutes()
}

func (w *NetworkWorker) DownPeer(cfg config.Bridge) {
	if cfg.Peer == "" {
		return
	}
	in := cfg.Name + "-in"
	ex := cfg.Name + "-ex"
	link := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: in},
		PeerName:  ex,
	}
	err := netlink.LinkDel(link)
	if err != nil {
		w.out.Error("NetworkWorker.DownPeer: %s", err)
		return
	}
}

func (w *NetworkWorker) Stop() {
	w.out.Info("NetworkWorker.Close")
	w.UnLoadRoutes()
	w.UnLoadLinks()
	w.startTime = 0
	_ = w.bridge.Close()
	brCfg := w.cfg.Bridge
	w.DownPeer(brCfg)
}

func (w *NetworkWorker) UpTime() int64 {
	if w.startTime != 0 {
		return time.Now().Unix() - w.startTime
	}
	return 0
}

func (w *NetworkWorker) AddLink(c *config.Point) {
	c.Alias = w.alias
	c.RequestAddr = false
	c.Network = w.cfg.Name
	c.Interface.Bridge = w.cfg.Bridge.Name // reset bridge name.
	c.Interface.Address = w.cfg.Bridge.Address
	c.Interface.Provider = w.cfg.Bridge.Provider
	libol.Go(func() {
		p := olap.NewPoint(c)
		p.Initialize()
		w.linksLock.Lock()
		w.links[c.Connection] = p
		w.linksLock.Unlock()
		storage.Link.Add(p)
		p.Start()
	})
}

func (w *NetworkWorker) DelLink(addr string) {
	w.linksLock.Lock()
	defer w.linksLock.Unlock()
	if p, ok := w.links[addr]; ok {
		p.Stop()
		storage.Link.Del(p.UUID())
		delete(w.links, addr)
	}
}
