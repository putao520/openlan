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
	tcpwroker *TcpWorker
	tapwroker *TapWorker
	brip      net.IP
	brnet     *net.IPNet
}

func NewPoint(config *Config) (p *Point) {
	ifce, err := water.New(water.Config{DeviceType: water.TAP})
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
}

func (p *Point) UpLink() error {
	//TODO
	return nil
}
