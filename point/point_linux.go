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

	tcpwroker *TcpWorker
	tapwroker *TapWorker
	br        tenus.Bridger
	brip      net.IP
	brnet     *net.IPNet
}

func NewPoint(config *Config) (p *Point) {
	var err error
	var ifce *water.Interface

	if config.Iftun {
		ifce, err = water.New(water.Config{DeviceType: water.TUN})
	} else {
		ifce, err = water.New(water.Config{DeviceType: water.TAP})
	}
	if err != nil {
		libol.Fatal("NewPoint: %s", err)
	}

	libol.Info("NewPoint.device %s", ifce.Name())
	client := libol.NewTcpClient(config.Addr)
	p = &Point{
		Client:    client,
		Ifce:      ifce,
		Brname:    config.Brname,
		Ifaddr:    config.Ifaddr,
		Ifname:    ifce.Name(),
		tapwroker: NewTapWorker(ifce, config),
		tcpwroker: NewTcpWorker(client, config),
	}
	return
}

func (p *Point) Start() {
	libol.Debug("Point.Start linux.")

	if err := p.Client.Connect(); err != nil {
		libol.Error("Point.Start %s", err)
	}

	go p.tapwroker.GoRecv(p.tcpwroker.DoSend)
	go p.tapwroker.GoLoop()

	go p.tcpwroker.GoRecv(p.tapwroker.DoSend)
	go p.tcpwroker.GoLoop()
}

func (p *Point) Close() {
	p.Client.Close()
	p.Ifce.Close()

	if p.br != nil && p.brip != nil {
		if err := p.br.UnsetLinkIp(p.brip, p.brnet); err != nil {
			libol.Error("Point.Close.UnsetLinkIp %s: %s", p.br.NetInterface().Name, err)
		}
	}
}

func (p *Point) UpLink() error {
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
