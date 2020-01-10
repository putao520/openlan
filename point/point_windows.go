package point

import (
	"crypto/tls"
	"fmt"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/network"
	"github.com/songgao/water"
	"os/exec"
	"strings"
)

type Point struct {
	BrName string
	IfAddr string

	tcpWorker *TcpWorker
	tapWorker *TapWorker
	addr      string
	routes    []*models.Route
	config    *config.Point
}

func NewPoint(config *config.Point) (p *Point) {
	var tlsConf *tls.Config
	if config.Tls {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}
	client := libol.NewTcpClient(config.Addr, tlsConf)
	p = &Point{
		BrName: config.BrName,
		IfAddr: config.IfAddr,
		config: config,
	}
	p.tcpWorker = NewTcpWorker(client, config)
	p.newDevice()
	return
}

func (p *Point) newDevice() {
	conf := &water.Config{DeviceType: water.TAP}
	p.tapWorker = NewTapWorker(conf, p.config)
}

func (p *Point) OnTap(w *TapWorker) error {
	libol.Info("Point.OnTap")
	return nil
}

func (p *Point) Start() {
	ctx := context.Background()
	libol.Debug("Point.Start Windows.")

	if err := p.tcpWorker.Connect(); err != nil {
		libol.Error("Point.Start %s", err)
	}

	p.tapWorker.Listener = TapWorkerListener{
		OnOpen: p.OnTap,
		ReadAt: p.tcpWorker.DoWrite,
	}
	go p.tapWorker.Read(ctx)
	go p.tapWorker.Loop(ctx)
	p.tcpWorker.Listener = TcpWorkerListener{
		OnClose:   p.OnClose,
		OnSuccess: p.OnSuccess,
		OnIpAddr:  p.OnIpAddr,
		ReadAt:    p.tapWorker.DoWrite,
	}
	go p.tcpWorker.Read(ctx)
	go p.tcpWorker.Loop(ctx)
}

func (p *Point) Stop() {
	defer libol.Catch("Point.Stop")

	p.DelAddr(p.addr)
	p.tapWorker.Stop()
	p.tcpWorker.Stop()
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

func (p *Point) AddAddr(ipStr string) error {
	if ipStr == "" {
		return nil
	}

	out, err := exec.Command("netsh", "interface",
		"ipv4", "add", "address", p.IfName(),
		ipStr, "store=active").Output()
	if err != nil {
		libol.Error("Point.AddAddr: %s, %s", err, out)
		return err
	}

	libol.Info("Point.AddAddr: %s", ipStr)
	p.addr = ipStr

	return nil
}

func (p *Point) DelAddr(ipStr string) error {
	if ipStr == "" {
		return nil
	}

	ipAddr := strings.Split(ipStr, "/")[0]
	out, err := exec.Command("netsh", "interface",
		"ipv4", "delete", "address", p.IfName(),
		ipAddr, "store=active").Output()
	if err != nil {
		libol.Error("Point.DelAddr: %s, %s", err, out)
		return err
	}
	libol.Info("Point.DelAddr: %s", ipStr)
	p.addr = ""

	return nil
}

func (p *Point) AddRoutes(routes []*models.Route) error {
	if routes == nil {
		return nil
	}

	for _, route := range routes {
		out, err := exec.Command("netsh", "interface",
			"ipv4", "add", "route", route.Prefix,
			p.IfName(), route.Nexthop, "store=active").Output()
		if err != nil {
			libol.Error("Point.AddRoutes: %s, %s", err, out)
			continue
		}
		libol.Info("Point.AddRoutes: %s via %s", route.Prefix, route.Nexthop)
	}

	p.routes = routes
	return nil
}

func (p *Point) DelRoutes(routes []*models.Route) error {
	if routes == nil {
		return nil
	}

	for _, route := range routes {
		out, err := exec.Command("netsh", "interface",
			"ipv4", "del", "route", route.Prefix,
			p.IfName(), route.Nexthop, "store=active").Output()
		if err != nil {
			libol.Error("Point.DelRoutes: %s, %s", err, out)
			continue
		}
		libol.Info("Point.DelRoutes: %s via %s", route.Prefix, route.Nexthop)
	}

	p.routes = nil

	return nil
}

func (p *Point) OnIpAddr(w *TcpWorker, n *models.Network) error {
	libol.Info("Point.OnIpAddr: %s, %s, %s", n.IfAddr, n.Netmask, n.Routes)

	if n.IfAddr == "" {
		return nil
	}

	prefix := libol.Netmask2Len(n.Netmask)
	ipStr := fmt.Sprintf("%s/%d", n.IfAddr, prefix)
	p.AddAddr(ipStr)
	p.AddRoutes(n.Routes)

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
