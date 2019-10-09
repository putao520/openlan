package point

import (
	"net"

	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/milosgajdos83/tenus"
	"github.com/songgao/water"
)

type Point struct {
	BrName string
	IfAddr string

	tcpWorker *TcpWorker
	tapWorker *TapWorker
	br        tenus.Bridger
	brIp      net.IP
	brNet     *net.IPNet
	config    *Config
}

func NewPoint(config *Config) (p *Point) {
	client := libol.NewTcpClient(config.Addr)
	p = &Point{
		BrName:    config.BrName,
		IfAddr:    config.IfAddr,
		tcpWorker: NewTcpWorker(client, config),
		config:    config,
	}
	p.newDevice()
	return
}

func (p *Point) newDevice() {
	var err error
	var dev *water.Interface

	if p.config.IfTun {
		dev, err = water.New(water.Config{DeviceType: water.TUN})
	} else {
		dev, err = water.New(water.Config{DeviceType: water.TAP})
	}
	if err != nil {
		libol.Fatal("NewPoint: %s", err)
		return
	}

	libol.Info("NewPoint.device %s", dev.Name())
	p.tapWorker = NewTapWorker(dev, p.config)
}

func (p *Point) Start() {
	libol.Debug("Point.Start linux.")

	p.UpLink()
	if err := p.tcpWorker.Connect(); err != nil {
		libol.Error("Point.Start %s", err)
	}

	go p.tapWorker.GoRecv(p.tcpWorker.DoSend)
	go p.tapWorker.GoLoop()

	go p.tcpWorker.GoRecv(p.tapWorker.DoSend)
	go p.tcpWorker.GoLoop()
}

func (p *Point) Stop() {
	p.tcpWorker.Stop()

	if p.br != nil && p.brIp != nil {
		if err := p.br.UnsetLinkIp(p.brIp, p.brNet); err != nil {
			libol.Error("Point.Close.UnsetLinkIp %s: %s", p.br.NetInterface().Name, err)
		}
	}
	p.tapWorker.Stop()
}

func (p *Point) UpLink() error {
	if p.GetDevice() == nil {
		p.newDevice()
	}
	if p.GetDevice() == nil {
		return libol.Errer("create device.")
	}

	name := p.GetDevice().Name()
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

		p.br = br
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

		p.brIp = ip
		p.brNet = ipNet
	}

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
		return p.tapWorker.device
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
		return client.State()
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
