package point

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/network"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
	"net"
)

type Point struct {
	BrName string
	IfAddr string

	tcpWorker *TcpWorker
	tapWorker *TapWorker
	addr      string
	routes    []*models.Route
	link      netlink.Link
	config    *config.Point
}

func NewPoint(config *config.Point) (p *Point) {
	var tlsConf *tls.Config
	if config.Tls {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}

	p = &Point{
		BrName: config.BrName,
		IfAddr: config.IfAddr,
		config: config,
	}

	client := libol.NewTcpClient(config.Addr, tlsConf)
	p.tcpWorker = NewTcpWorker(client, config, p)
	p.newDevice()

	return
}

func (p *Point) newDevice() {
	var conf *water.Config

	if p.config.IfTun {
		conf = &water.Config{DeviceType: water.TUN}
	} else {
		conf = &water.Config{DeviceType: water.TAP}
	}

	p.tapWorker = NewTapWorker(conf, p.config, p)
}

func (p *Point) Start() {
	libol.Debug("Point.Start linux.")

	if err := p.tcpWorker.Connect(); err != nil {
		libol.Error("Point.Start %s", err)
	}

	ctx := context.Background()
	go p.tapWorker.Read(ctx, p.tcpWorker.DoWrite)
	go p.tapWorker.Loop(ctx)

	go p.tcpWorker.Read(ctx, p.tapWorker.DoWrite)
	go p.tcpWorker.Loop(ctx)
}

func (p *Point) Stop() {
	defer libol.Catch("Point.Stop")

	p.DelRoutes(p.routes)
	p.DelAddr(p.addr)

	p.tcpWorker.Stop()
	p.tapWorker.Stop()
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
		libol.Error("Point.DelAddr.UnsetLinkIp: %s", err)
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
		libol.Error("Point.AddAddr.SetLinkIp: %s", err)
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
	if link, _ := netlink.LinkByName(name); link == nil {
		err := netlink.LinkAdd(br)
		if err != nil {
			libol.Error("Point.UpBr.newBr: %s", err)
			return nil
		}
	}

	link, err := netlink.LinkByName(name)
	if link == nil {
		libol.Error("Point.UpBr.newBr: %s", err)
		return nil
	}

	brCtl := libol.NewBrCtl(name)
	if err := brCtl.Stp(true); err != nil {
		libol.Error("Point.UpBr.Stp: %s", err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		libol.Error("Point.UpBr.newBr.Up: %s", err)
	}

	return br
}

func (p *Point) OnTap(w *TapWorker) error {
	libol.Info("Point.OnTap")

	name := w.Device.Name()
	link, err := netlink.LinkByName(name)
	if err != nil {
		libol.Error("Point.OnTap: Get dev %s: %s", name, err)
		return err
	}
	if err := netlink.LinkSetUp(link); err != nil {
		libol.Error("Point.OnTap.SetLinkUp: %s: %s", name, err)
		return err
	}

	if br := p.UpBr(p.BrName); br != nil {
		if err := netlink.LinkSetMaster(link, br); err != nil {
			libol.Error("Point.OnTap.AddSlave: Switch dev %s: %s", name, err)
		}

		link, err = netlink.LinkByName(p.BrName)
		if err != nil {
			libol.Error("Point.OnTap: Get dev %s: %s", p.BrName, err)
		}
	}

	p.link = link

	return nil
}

func (p *Point) GetClient() *libol.TcpClient {
	if p.tcpWorker != nil {
		return p.tcpWorker.Client
	}
	return nil
}

func (p *Point) GetDevice() network.Taper {
	if p.tapWorker != nil {
		return p.tapWorker.Device
	}
	return nil
}

func (p *Point) UpTime() int64 {
	client := p.GetClient()
	if client != nil {
		return client.UpTime()
	}
	return 0
}

func (p *Point) State() string {
	client := p.GetClient()
	if client != nil {
		return client.GetState()
	}
	return ""
}

func (p *Point) Addr() string {
	client := p.GetClient()
	if client != nil {
		return client.Addr
	}
	return ""
}

func (p *Point) IfName() string {
	dev := p.GetDevice()
	if dev != nil {
		return dev.Name()
	}
	return ""
}

func (p *Point) GetWorker() *TcpWorker {
	if p.tcpWorker != nil {
		return p.tcpWorker
	}
	return nil
}

func (p *Point) OnIpAddr(w *TcpWorker, n *models.Network) error {
	libol.Info("Point.OnIpAddr: %s, %s, %s", n.IfAddr, n.Netmask, n.Routes)

	if n.IfAddr == "" || p.link == nil {
		return nil
	}

	prefix := libol.Netmask2Len(n.Netmask)
	ipStr := fmt.Sprintf("%s/%d", n.IfAddr, prefix)

	p.AddAddr(ipStr)
	p.AddRoutes(n.Routes)

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

		nxt := net.ParseIP(route.Nexthop)
		rte := netlink.Route{LinkIndex: p.link.Attrs().Index, Dst: dst, Gw: nxt}
		libol.Debug("Point.AddRoute: %s", rte)
		if err := netlink.RouteAdd(&rte); err != nil {
			libol.Error("Point.AddRoute: %s", err)
			continue
		}
		libol.Info("Point.OnIpAddr.route: %s via %s", route.Prefix, route.Nexthop)
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

		nxt := net.ParseIP(route.Nexthop)
		rte := netlink.Route{LinkIndex: p.link.Attrs().Index, Dst: dst, Gw: nxt}
		if err := netlink.RouteDel(&rte); err != nil {
			libol.Error("Point.DelRoute: %s", err)
			continue
		}
		libol.Info("Point.DelRoutes.route: %s via %s", route.Prefix, route.Nexthop)
	}

	p.routes = nil

	return nil
}

func (p *Point) OnClose(w *TcpWorker) error {
	libol.Info("Point.OnClose")

	p.DelRoutes(p.routes)
	p.DelAddr(p.addr)

	return nil
}

func (p *Point) OnSuccess(w *TcpWorker) error {
	libol.Info("Point.OnSuccess")

	p.AddAddr(p.IfAddr)

	return nil
}
