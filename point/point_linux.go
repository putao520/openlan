package point

import (
	"net"

	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/milosgajdos83/tenus"
	"github.com/songgao/water"
)

type Point struct {
	Client *libol.TcpClient
	Ifce   *water.Interface
	Brname string
	Ifaddr string
	Ifname string

	tcpworker *TcpWorker
	tapworker *TapWorker
	br        tenus.Bridger
	brip      net.IP
	brnet     *net.IPNet
	config    *Config
}

func NewPoint(config *Config) (p *Point) {
	client := libol.NewTcpClient(config.Addr)
	p = &Point{
		Client:    client,
		Brname:    config.Brname,
		Ifaddr:    config.Ifaddr,
		tcpworker: NewTcpWorker(client, config),
		config:    config,
	}
	p.newIfce()
	return
}

func (p *Point) newIfce() {
	var err error
	var ifce *water.Interface

	if p.config.Iftun {
		ifce, err = water.New(water.Config{DeviceType: water.TUN})
	} else {
		ifce, err = water.New(water.Config{DeviceType: water.TAP})
	}
	if err != nil {
		libol.Fatal("NewPoint: %s", err)
		return
	}

	libol.Info("NewPoint.device %s", ifce.Name())
	p.Ifce   = ifce
	p.Ifname = ifce.Name()
	p.tapworker = NewTapWorker(ifce, p.config)
}

func (p *Point) Start() {
	libol.Debug("Point.Start linux.")

	p.UpLink()
	if err := p.Client.Connect(); err != nil {
		libol.Error("Point.Start %s", err)
	}

	go p.tapworker.GoRecv(p.tcpworker.DoSend)
	go p.tapworker.GoLoop()

	go p.tcpworker.GoRecv(p.tapworker.DoSend)
	go p.tcpworker.GoLoop()
}

func (p *Point) Close() {
	p.tcpworker.Close()

	if p.br != nil && p.brip != nil {
		if err := p.br.UnsetLinkIp(p.brip, p.brnet); err != nil {
			libol.Error("Point.Close.UnsetLinkIp %s: %s", p.br.NetInterface().Name, err)
		}
	}
	p.tapworker.Close()
	p.Ifce = nil
}

func (p *Point) UpLink() error {
	if p.Ifce == nil {
		p.newIfce()
	}
	if p.Ifce == nil {
		return libol.Errer("create device.")
	}

	name := p.Ifce.Name()
	libol.Debug("Point.UpLink: %s", name)
	link, err := tenus.NewLinkFrom(name)
	if err != nil {
		libol.Error("Point.UpLink: Get ifce %s: %s", name, err)
		return err
	}

	if err := link.SetLinkUp(); err != nil {
		libol.Error("Point.UpLink.SetLinkUp: %s: %s", name, err)
		return err
	}

	if p.Brname != "" {
		br, err := tenus.BridgeFromName(p.Brname)
		if err != nil {
			libol.Error("Point.UpLink.newBr: %s", err)
			br, err = tenus.NewBridgeWithName(p.Brname)
			if err != nil {
				libol.Error("Point.UpLink.newBr: %s", err)
			}
		}

		brctl := libol.NewBrCtl(p.Brname)
		if err := brctl.Stp(true); err != nil {
			libol.Error("Point.UpLink.Stp: %s", err)
		}

		if err := br.SetLinkUp(); err != nil {
			libol.Error("Point.UpLink.newBr.Up: %s", err)
		}

		if err := br.AddSlaveIfc(link.NetInterface()); err != nil {
			libol.Error("Point.UpLink.AddSlave: Switch ifce %s: %s", name, err)
		}

		link, err = tenus.NewLinkFrom(p.Brname)
		if err != nil {
			libol.Error("Point.UpLink: Get ifce %s: %s", p.Brname, err)
		}

		p.br = br
	}

	if p.Ifaddr != "" {
		ip, ipnet, err := net.ParseCIDR(p.Ifaddr)
		if err != nil {
			libol.Error("Point.UpLink.ParseCIDR %s: %s", p.Ifaddr, err)
			return err
		}
		if err := link.SetLinkIp(ip, ipnet); err != nil {
			libol.Error("Point.UpLink.SetLinkIp: %s", err)
			return err
		}

		p.brip = ip
		p.brnet = ipnet
	}

	return nil
}
