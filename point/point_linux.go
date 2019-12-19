package point

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os/exec"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/milosgajdos83/tenus"
	"github.com/songgao/water"
)

type Point struct {
	BrName string
	IfAddr string

	tcpWorker *TcpWorker
	tapWorker *TapWorker
	ipAddr    net.IP
	ipNet     *net.IPNet
	link      tenus.Linker
	config    *config.Point
}

func NewPoint(config *config.Point) (p *Point) {
	var tlsConf *tls.Config
	if config.Tls {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}

	p = &Point{
		BrName:    config.BrName,
		IfAddr:    config.IfAddr,
		config:    config,
	}

	client := libol.NewTcpClient(config.Addr, tlsConf)
	p.tcpWorker =  NewTcpWorker(client, config, p)
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
	go p.tapWorker.GoRecv(ctx, p.tcpWorker.DoSend)
	go p.tapWorker.GoLoop(ctx)

	go p.tcpWorker.GoRecv(ctx, p.tapWorker.DoSend)
	go p.tcpWorker.GoLoop(ctx)
}

func (p *Point) Stop() {
	defer libol.Catch("Point.Stop")

	p.tcpWorker.Stop()

	if p.link != nil && p.ipAddr != nil {
		if err := p.link.UnsetLinkIp(p.ipAddr, p.ipNet); err != nil {
			libol.Error("Point.Close.UnsetLinkIp: %s", err)
		}
	}
	p.tapWorker.Stop()
}

func (p *Point) OnTap(tap *TapWorker) error {
	name := tap.Device.Name()
	libol.Debug("Point.UpLink: %s", name)
	link, err := tenus.NewLinkFrom(name)
	if err != nil {
		libol.Error("Point.UpLink: Get dev %s: %s", name, err)
		return err
	}

	if err := link.SetLinkUp(); err != nil {
		libol.Error("Point.UpLink.SetLinkUp: %s: %s", name, err)
		return err
	}

	if p.BrName != "" {
		br, err := tenus.BridgeFromName(p.BrName)
		if err != nil {
			libol.Error("Point.UpLink.newBr: %s", err)
			br, err = tenus.NewBridgeWithName(p.BrName)
			if err != nil {
				libol.Error("Point.UpLink.newBr: %s", err)
			}
		}

		brCtl := libol.NewBrCtl(p.BrName)
		if err := brCtl.Stp(true); err != nil {
			libol.Error("Point.UpLink.Stp: %s", err)
		}

		if err := br.SetLinkUp(); err != nil {
			libol.Error("Point.UpLink.newBr.Up: %s", err)
		}

		if err := br.AddSlaveIfc(link.NetInterface()); err != nil {
			libol.Error("Point.UpLink.AddSlave: Switch dev %s: %s", name, err)
		}

		link, err = tenus.NewLinkFrom(p.BrName)
		if err != nil {
			libol.Error("Point.UpLink: Get dev %s: %s", p.BrName, err)
		}
	}

	if p.IfAddr != "" {
		ip, ipNet, err := net.ParseCIDR(p.IfAddr)
		if err != nil {
			libol.Error("Point.UpLink.ParseCIDR %s: %s", p.IfAddr, err)
			return err
		}
		if err := link.SetLinkIp(ip, ipNet); err != nil {
			libol.Error("Point.UpLink.SetLinkIp: %s", err)
			return err
		}

		p.ipAddr = ip
		p.ipNet = ipNet
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

func (p *Point) OnIpAddr(worker *TcpWorker, n *models.Network) error {
	libol.Info("Point.OnIpAddr: %s, %s, %s", n.IfAddr, n.Netmask, n.Routes)

	if n.IfAddr == "" || p.link == nil {
		return nil
	}

	prefix := libol.Netmask2Len(n.Netmask)
	ipStr := fmt.Sprintf("%s/%d", n.IfAddr, prefix)
	ip, ipNet, err := net.ParseCIDR(ipStr)
	if err != nil {
		libol.Error("Point.OnIpAddr.ParseCIDR %s: %s", ipStr, err)
		return err
	}

	if err := p.link.SetLinkIp(ip, ipNet); err != nil {
		libol.Error("Point.OnIpAddr.SetLinkIp: %s", err)
		return err
	}
	libol.Info("Point.OnIpAddr: %s", ipStr)

	p.ipAddr = ip
	p.ipNet = ipNet

	if n.Routes != nil {
		for _, route := range n.Routes {
			_, err := exec.Command("/usr/sbin/ip", "route",
				"add", route.Prefix, "dev", p.IfName(), "via", route.Nexthop).Output()
			if err != nil {
				libol.Error("Point.OnIpAddr.route: %s", err)
			}
			libol.Info("Point.OnIpAddr.route: %s via %s", route.Prefix, route.Nexthop)
		}
	}

	return nil
}