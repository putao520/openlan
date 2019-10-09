package point

import (
	"net"

	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/songgao/water"
)

type Point struct {
	BrName string
	IfAddr string

	tcpWorker *TcpWorker
	tapWorker *TapWorker
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
	return
}

func (p *Point) newIfce() {
	ifce, err := water.New(water.Config{DeviceType: water.TAP})
	if err != nil {
		libol.Fatal("NewPoint: %s", err)
		return
	}

	libol.Info("NewPoint.device %s", ifce.Name())
	p.tapWorker = NewTapWorker(ifce, p.config)
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
	libol.Debug("Point.Start Windows.")

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
	p.tapWorker.Stop()
	p.tcpWorker.Stop()
}

func (p *Point) GetClient() *libol.TcpClient {
	if p.tcpWorker != nil {
		return p.tcpWorker.Client
	}
	return nil
}

func (p *Point) GetIfce() *water.Interface {
	if p.tapWorker != nil {
		return p.tapWorker.Ifce
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
	return "-"
}

func (p *Point) Addr() string {
	client := p.GetClient()
	if client != nil {
		return client.Addr
	}
	return "-"
}

func (p *Point) IfName() string {
	ifce := p.GetIfce()
	if ifce != nil {
		return ifce.Name()
	}
	return "-"
}