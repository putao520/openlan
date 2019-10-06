package point

import (
	"net"

	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/songgao/water"
)

type Point struct {
	Brname string
	Ifaddr string
	Ifname string

	tcpworker *TcpWorker
	tapworker *TapWorker
	brip      net.IP
	brnet     *net.IPNet
	config    *Config
}

func NewPoint(config *Config) (p *Point) {
	client := libol.NewTcpClient(config.Addr)
	p = &Point{
		Brname:    config.Brname,
		Ifaddr:    config.Ifaddr,
		tcpworker: NewTcpWorker(client, config),
		config:    config,
	}
	p.newIfce()
	return
}

func (p *Point) newIfce() {
	ifce, err := water.New(water.Config{DeviceType: water.TUN})
	if err != nil {
		libol.Fatal("NewPoint: %s", err)
		return
	}

	libol.Info("NewPoint.device %s", ifce.Name())
	p.Ifname = ifce.Name()
	p.tapworker = NewTapWorker(ifce, p.config)
}

func (p *Point) UpLink() error {
	if p.GetIfce() == nil {
		p.newIfce()
	}
	if p.GetIfce() == nil {
		return libol.Errer("create device.")
	}
	return nil
}

func (p *Point) Start() {
	libol.Debug("Point.Start Darwin.")

	p.UpLink()
	if err := p.tcpworker.Connect(); err != nil {
		libol.Error("Point.Start %s", err)
	}

	go p.tapworker.GoRecv(p.tcpworker.DoSend)
	go p.tapworker.GoLoop()

	go p.tcpworker.GoRecv(p.tapworker.DoSend)
	go p.tcpworker.GoLoop()
}

func (p *Point) Stop() {
	p.tapworker.Stop()
	p.tcpworker.Stop()
}

func (p *Point) GetClient() *libol.TcpClient {
	if p.tcpworker != nil {
		return p.tcpworker.Client
	}
	return nil
}

func (p *Point) GetIfce() *water.Interface {
	if p.tapworker != nil {
		return p.tapworker.Ifce
	}
	return nil
}
