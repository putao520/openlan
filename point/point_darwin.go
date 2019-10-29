package point

import (
	"crypto/tls"
	"github.com/lightstar-dev/openlan-go/config"
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
	config    *config.Point
}

func NewPoint(config *config.Point) (p *Point) {
	var tlsConf *tls.Config
	if config.Tls {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}
	client := libol.NewTcpClient(config.Addr, tlsConf)
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
	conf := &water.Config{DeviceType: water.TUN}
	p.tapWorker = NewTapWorker(conf, p.config)
}

func (p *Point) UpLink(tap *TapWorker) error {
	return nil
}

func (p *Point) Start() {
	libol.Debug("Point.Start Darwin.")
	
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

	p.tapWorker.Stop()
	p.tcpWorker.Stop()
}

func (p *Point) GetClient() *libol.TcpClient {
	if p.tcpWorker != nil {
		return p.tcpWorker.Client
	}
	return nil
}

func (p *Point) GetDevice() *water.Interface {
	if p.tapWorker != nil {
		return p.tapWorker.Dev
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
