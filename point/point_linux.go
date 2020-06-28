package point

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/models"
	"github.com/vishvananda/netlink"
	"net"
)

type Point struct {
	MixPoint
	// private
	brName string
	addr   string
	routes []*models.Route
	link   netlink.Link
	uuid   string
}

func NewPoint(config *config.Point) *Point {
	p := Point{
		brName:   config.Interface.Bridge,
		MixPoint: NewMixPoint(config),
	}
	return &p
}

func (p *Point) Initialize() {
	p.worker.listener.AddAddr = p.AddAddr
	p.worker.listener.DelAddr = p.DelAddr
	p.worker.listener.AddRoutes = p.AddRoutes
	p.worker.listener.DelRoutes = p.DelRoutes
	p.worker.listener.OnTap = p.OnTap
	p.MixPoint.Initialize()
}

func (p *Point) Start() {
	libol.Info("Point.Start: linux.")
	p.worker.Start()
}

func (p *Point) Stop() {
	defer libol.Catch("Point.Stop")
	p.worker.Stop()
}

func (p *Point) DelAddr(ipStr string) error {
	if p.link == nil || ipStr == "" {
		return nil
	}
	ipAddr, err := netlink.ParseAddr(ipStr)
	if err != nil {
		libol.Error("Point.AddAddr.ParseCIDR %s: %s", ipStr, err)
		return err
	}
	if err := netlink.AddrDel(p.link, ipAddr); err != nil {
		libol.Warn("Point.DelAddr.UnsetLinkIp: %s", err)
	}
	libol.Info("Point.DelAddr: %s", ipStr)
	p.addr = ""
	return nil
}

func (p *Point) AddAddr(ipStr string) error {
	if ipStr == "" || p.link == nil {
		return nil
	}
	ipAddr, err := netlink.ParseAddr(ipStr)
	if err != nil {
		libol.Error("Point.AddAddr.ParseCIDR %s: %s", ipStr, err)
		return err
	}
	if err := netlink.AddrAdd(p.link, ipAddr); err != nil {
		libol.Warn("Point.AddAddr.SetLinkIp: %s", err)
		return err
	}
	libol.Info("Point.AddAddr: %s", ipStr)
	p.addr = ipStr
	return nil
}

func (p *Point) UpBr(name string) *netlink.Bridge {
	if name == "" {
		return nil
	}
	la := netlink.LinkAttrs{TxQLen: -1, Name: name}
	br := &netlink.Bridge{LinkAttrs: la}
	if link, err := netlink.LinkByName(name); link == nil {
		libol.Warn("Point.UpBr: %s %s", name, err)
		err := netlink.LinkAdd(br)
		if err != nil {
			libol.Warn("Point.UpBr.newBr: %s %s", name, err)
		}
	}
	link, err := netlink.LinkByName(name)
	if link == nil {
		libol.Error("Point.UpBr: %s %s", name, err)
		return nil
	}
	if err := netlink.LinkSetUp(link); err != nil {
		libol.Error("Point.UpBr.LinkUp: %s", err)
	}
	brCtl := libol.NewBrCtl(name)
	if err := brCtl.Stp(true); err != nil {
		libol.Error("Point.UpBr.Stp: %s", err)
	}
	if err := brCtl.Delay(2); err != nil {
		libol.Error("Point.UpBr.Delay: %s", err)
	}
	return br
}

func (p *Point) OnTap(w *TapWorker) error {
	libol.Info("Point.OnTap")

	name := w.device.Name()
	link, err := netlink.LinkByName(name)
	if err != nil {
		libol.Error("Point.OnTap: Get dev %s: %s", name, err)
		return err
	}
	if err := netlink.LinkSetUp(link); err != nil {
		libol.Error("Point.OnTap.SetLinkUp: %s: %s", name, err)
		return err
	}
	if br := p.UpBr(p.brName); br != nil {
		if err := netlink.LinkSetMaster(link, br); err != nil {
			libol.Error("Point.OnTap.AddSlave: Switch dev %s: %s", name, err)
		}
		link, err = netlink.LinkByName(p.brName)
		if err != nil {
			libol.Error("Point.OnTap: Get dev %s: %s", p.brName, err)
		}
	}
	p.link = link
	return nil
}

func (p *Point) AddRoutes(routes []*models.Route) error {
	if routes == nil || p.link == nil {
		return nil
	}

	for _, route := range routes {
		_, dst, err := net.ParseCIDR(route.Prefix)
		if err != nil {
			continue
		}
		nxt := net.ParseIP(route.NextHop)
		rte := netlink.Route{LinkIndex: p.link.Attrs().Index, Dst: dst, Gw: nxt}
		libol.Debug("Point.AddRoute: %s", rte)
		if err := netlink.RouteAdd(&rte); err != nil {
			libol.Warn("Point.AddRoute: %s", err)
			continue
		}
		libol.Info("Point.AddRoutes: route %s via %s", route.Prefix, route.NextHop)
	}
	p.routes = routes
	return nil
}

func (p *Point) DelRoutes(routes []*models.Route) error {
	if routes == nil || p.link == nil {
		return nil
	}
	for _, route := range routes {
		_, dst, err := net.ParseCIDR(route.Prefix)
		if err != nil {
			continue
		}
		nxt := net.ParseIP(route.NextHop)
		rte := netlink.Route{LinkIndex: p.link.Attrs().Index, Dst: dst, Gw: nxt}
		if err := netlink.RouteDel(&rte); err != nil {
			libol.Warn("Point.DelRoute: %s", err)
			continue
		}
		libol.Info("Point.DelRoutes: route %s via %s", route.Prefix, route.NextHop)
	}
	p.routes = nil
	return nil
}
