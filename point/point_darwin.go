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

	tcpwroker *TcpWorker
	tapwroker *TapWorker
	brip      net.IP
	brnet     *net.IPNet
}

func NewPoint(config *Config) (this *Point) {
	ifce, err := water.New(water.Config{DeviceType: water.TUN})
	if err != nil {
		libol.Fatal("NewPoint: %s", err)
	}
	libol.Info("NewPoint.device %s", ifce.Name())

	client := libol.NewTcpClient(config.Addr)
	this = &Point{
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

func (this *Point) Start() {
	if err := this.Client.Connect(); err != nil {
		libol.Error("Point.Start %s", err)
	}

	go this.tapwroker.GoRecv(this.tcpwroker.DoSend)
	go this.tapwroker.GoLoop()

	go this.tcpwroker.GoRecv(this.tapwroker.DoSend)
	go this.tcpwroker.GoLoop()
}

func (this *Point) Close() {
	this.Client.Close()
	this.Ifce.Close()
}
