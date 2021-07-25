package olap

import (
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/vishvananda/netlink"
	"net"
)

type Point struct {
	MixPoint
	// private
	brName string
	ipMtu  int
	addr   string
	routes []*models.Route
	link   netlink.Link
	uuid   string
}

func NewPoint(config *config.Point) *Point {
	ipMtu := config.Interface.IPMtu
	if ipMtu == 0 {
		ipMtu = 1500
	}
	p := Point{
		ipMtu:    ipMtu,
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

func (p *Point) DelAddr(ipStr string) error {
	if p.link == nil || ipStr == "" {
		return nil
	}
	ipAddr, err := netlink.ParseAddr(ipStr)
	if err != nil {
		p.out.Error("Point.AddAddr.ParseCIDR %s: %s", ipStr, err)
		return err
	}
	if err := netlink.AddrDel(p.link, ipAddr); err != nil {
		p.out.Warn("Point.DelAddr.UnsetLinkIp: %s", err)
	}
	p.out.Info("Point.DelAddr: %s", ipStr)
	p.addr = ""
	return nil
}

func (p *Point) AddAddr(ipStr string) error {
	if ipStr == "" || p.link == nil {
		return nil
	}
	ipAddr, err := netlink.ParseAddr(ipStr)
	if err != nil {
		p.out.Error("Point.AddAddr.ParseCIDR %s: %s", ipStr, err)
		return err
	}
	if err := netlink.AddrAdd(p.link, ipAddr); err != nil {
		p.out.Warn("Point.AddAddr.SetLinkIp: %s", err)
		return err
	}
	p.out.Info("Point.AddAddr: %s", ipStr)
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
		p.out.Warn("Point.UpBr: %s %s", name, err)
		err := netlink.LinkAdd(br)
		if err != nil {
			p.out.Warn("Point.UpBr.newBr: %s %s", name, err)
		}
	}
	link, err := netlink.LinkByName(name)
	if link == nil {
		p.out.Error("Point.UpBr: %s %s", name, err)
		return nil
	}
	if err := netlink.LinkSetUp(link); err != nil {
		p.out.Error("Point.UpBr.LinkUp: %s", err)
	}
	return br
}

func (p *Point) OnTap(w *TapWorker) error {
	p.out.Info("Point.OnTap")
	tap := w.device
	name := tap.Name()
	if tap.Type() == network.ProviderVir { // virtual device
		p.out.Error("Point.OnTap: not support %s", tap.Type())
		return nil
	}
	// kernel device
	link, err := netlink.LinkByName(name)
	if err != nil {
		p.out.Error("Point.OnTap: Get %s: %s", name, err)
		return err
	}
	if err := netlink.LinkSetMTU(link, p.ipMtu); err != nil {
		p.out.Error("Point.OnTap.SetMTU: %s", err)
	}
	if br := p.UpBr(p.brName); br != nil {
		if err := netlink.LinkSetMaster(link, br); err != nil {
			p.out.Error("Point.OnTap.AddSlave: Switch dev %s: %s", name, err)
		}
		link, err = netlink.LinkByName(p.brName)
		if err != nil {
			p.out.Error("Point.OnTap: Get %s: %s", p.brName, err)
		}
	}
	if p.config.Interface.Cost > 0 {
		port := network.NewBrPort(name)
		if err := port.Cost(p.config.Interface.Cost); err != nil {
			p.out.Error("Point.OnTap: Cost %s: %s", err)
		}
	}
	p.link = link
	return nil
}

func (p *Point) AddRoutes(routes []*models.Route) error {
	if routes == nil || p.link == nil {
		return nil
	}
	for _, rt := range routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		nxt := net.ParseIP(rt.NextHop)
		rte := netlink.Route{
			LinkIndex: p.link.Attrs().Index,
			Dst:       dst,
			Gw:        nxt,
			Priority:  rt.Metric,
		}
		p.out.Debug("Point.AddRoute: %s", rte)
		if err := netlink.RouteAdd(&rte); err != nil {
			p.out.Warn("Point.AddRoute: %s %s", rt.Prefix, err)
			continue
		}
		p.out.Info("Point.AddRoutes: route %s via %s", rt.Prefix, rt.NextHop)
	}
	p.routes = routes
	return nil
}

func (p *Point) DelRoutes(routes []*models.Route) error {
	if routes == nil || p.link == nil {
		return nil
	}
	for _, rt := range routes {
		_, dst, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		nxt := net.ParseIP(rt.NextHop)
		rte := netlink.Route{
			LinkIndex: p.link.Attrs().Index,
			Dst:       dst,
			Gw:        nxt,
			Priority:  rt.Metric,
		}
		if err := netlink.RouteDel(&rte); err != nil {
			p.out.Warn("Point.DelRoute: %s %s", rt.Prefix, err)
			continue
		}
		p.out.Info("Point.DelRoutes: route %s via %s", rt.Prefix, rt.NextHop)
	}
	p.routes = nil
	return nil
}
