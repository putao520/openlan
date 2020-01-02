package point

import (
	"context"
	"crypto/tls"
	"fmt"
	"os/exec"

	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/milosgajdos83/tenus"
	"github.com/songgao/water"
	"net"
)

type Point struct {
	BrName string
	IfAddr string

	tcpWorker *TcpWorker
	tapWorker *TapWorker
	addr      string
	routes    []*models.Route
	link      tenus.Linker
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

	p.DelAddr(p.addr)
	p.tcpWorker.Stop()
	p.tapWorker.Stop()
}

func (p *Point) DelAddr(ipStr string) error {
	if p.link == nil || ipStr == "" {
		return nil
	}

	ip, ipNet, err := net.ParseCIDR(ipStr)
	if err != nil {
		libol.Error("Point.AddAddr.ParseCIDR %s: %s", ipStr, err)
		return err
	}

	if err := p.link.UnsetLinkIp(ip, ipNet); err != nil {
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

	ip, ipNet, err := net.ParseCIDR(ipStr)
	if err != nil {
		libol.Error("Point.AddAddr.ParseCIDR %s: %s", ipStr, err)
		return err
	}
	if err := p.link.SetLinkIp(ip, ipNet); err != nil {
		libol.Error("Point.AddAddr.SetLinkIp: %s", err)
		return err
	}

	libol.Info("Point.AddAddr: %s", ipStr)

	p.addr = ipStr

	return nil
}

func (p *Point) UpBr(name string) tenus.Bridger {
	if name == "" {
		return nil
	}

	br, err := tenus.BridgeFromName(name)
	if err != nil {
		libol.Error("Point.UpBr.newBr: %s", err)
		br, err = tenus.NewBridgeWithName(name)
		if err != nil {
			libol.Error("Point.UpBr.newBr: %s", err)
			return nil
		}
	}

	brCtl := libol.NewBrCtl(name)
	if err := brCtl.Stp(true); err != nil {
		libol.Error("Point.UpBr.Stp: %s", err)
	}

	if err := br.SetLinkUp(); err != nil {
		libol.Error("Point.UpBr.newBr.Up: %s", err)
	}

	return br
}

func (p *Point) OnTap(w *TapWorker) error {
	libol.Info("Point.OnTap")

	name := w.Device.Name()
	link, err := tenus.NewLinkFrom(name)
	if err != nil {
		libol.Error("Point.OnTap: Get dev %s: %s", name, err)
		return err
	}

	if err := link.SetLinkUp(); err != nil {
		libol.Error("Point.OnTap.SetLinkUp: %s: %s", name, err)
		return err
	}

	if br := p.UpBr(p.BrName); br != nil {
		if err := br.AddSlaveIfc(link.NetInterface()); err != nil {
			libol.Error("Point.OnTap.AddSlave: Switch dev %s: %s", name, err)
		}

		link, err = tenus.NewLinkFrom(p.BrName)
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

func (p *Point) GetDevice() *water.Interface {
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
	if routes == nil {
		return nil
	}

	for _, route := range routes {
		out, err := exec.Command("/usr/sbin/ip", "route",
			"add", route.Prefix, "dev", p.IfName(), "via", route.Nexthop).Output()
		if err != nil {
			libol.Error("Point.OnIpAddr.route: %s, %s", err, out)
		}
		libol.Info("Point.OnIpAddr.route: %s via %s", route.Prefix, route.Nexthop)
	}

	p.routes = routes

	return nil
}

func (p *Point) DelRoutes(routes []*models.Route) error {
	if routes == nil {
		return nil
	}

	for _, route := range routes {
		out, err := exec.Command("/usr/sbin/ip", "route",
			"del", route.Prefix, "dev", p.IfName(), "via", route.Nexthop).Output()
		if err != nil {
			libol.Error("Point.DelRoutes.route: %s, %s", err, out)
		}
		libol.Info("Point.DelRoutes.route: %s via %s", route.Prefix, route.Nexthop)
	}

	p.routes = nil

	return nil
}

func (p *Point) OnClose(w *TcpWorker) error {
	libol.Info("Point.OnClose")

	p.DelAddr(p.addr)
	p.DelRoutes(p.routes)

	return nil
}

func (p *Point) OnSuccess(w *TcpWorker) error {
	libol.Info("Point.OnSuccess")

	p.AddAddr(p.IfAddr)

	return nil
}
