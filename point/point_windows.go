package point

import (
	"net"

	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/songgao/water"
)

type Point struct {
	Client *libol.TcpClient
	Ifce   *water.Interface
	Brname string
	Ifaddr string
	Ifname string

	//
	tcpworker *TcpWorker
	tapworker *TapWorker
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
	return
}

func (p *Point) newIfce() {
	ifce, err := water.New(water.Config{DeviceType: water.TAP})
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
	libol.Debug("Point.Start Windows.")

	if p.Ifce == nil {
		p.newIfce()
	}
	if err := p.Client.Connect(); err != nil {
		libol.Error("Point.Start %s", err)
	}

	go p.tapworker.GoRecv(p.tcpworker.DoSend)
	go p.tapworker.GoLoop()

	go p.tcpworker.GoRecv(p.tapworker.DoSend)
	go p.tcpworker.GoLoop()
}

func (p *Point) Close() {
	p.tapworker.Close()
	p.tcpworker.Close()
	p.Ifce = nil
}

func (p *Point) UpLink() error {
	//TODO
	return nil
}
