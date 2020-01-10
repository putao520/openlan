package point

import (
	"crypto/tls"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/network"
	"github.com/songgao/water"
)

type Point struct {
	BrName string
	IfAddr string

	tcpWorker *TcpWorker
	tapWorker *TapWorker
	config    *config.Point
}

func NewPoint(config *config.Point) (p *Point) {
	var tlsConf *tls.Config
	if config.Tls {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}
	client := libol.NewTcpClient(config.Addr, tlsConf)
	p = &Point{
		BrName: config.BrName,
		IfAddr: config.IfAddr,
		config: config,
	}
	p.tcpWorker = NewTcpWorker(client, config)
	p.newDevice()
	return
}

func (p *Point) newDevice() {
	conf := &water.Config{DeviceType: water.TUN}
	p.tapWorker = NewTapWorker(conf, p.config)
}

func (p *Point) OnTap(w *TapWorker) error {
	return nil
}

func (p *Point) Start() {
	ctx := context.Background()
	libol.Debug("Point.Start Darwin.")

	if err := p.tcpWorker.Connect(); err != nil {
		libol.Error("Point.Start %s", err)
	}

	p.tapWorker.Listener = TapWorkerListener{
		OnOpen: p.OnTap,
		ReadAt: p.tcpWorker.DoWrite,
	}
	go p.tapWorker.Read(ctx)
	go p.tapWorker.Loop(ctx)
	p.tcpWorker.Listener = TcpWorkerListener{
		OnClose:   p.OnClose,
		OnSuccess: p.OnSuccess,
		OnIpAddr:  p.OnIpAddr,
		ReadAt:    p.tapWorker.DoWrite,
	}
	go p.tcpWorker.Read(ctx)
	go p.tcpWorker.Loop(ctx)
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

func (p *Point) GetDevice() network.Taper {
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

func (p *Point) OnIpAddr(w *TcpWorker, n *models.Network) error {
	return nil
}

func (p *Point) OnClose(w *TcpWorker) error {
	return nil
}

func (p *Point) OnSuccess(w *TcpWorker) error {
	return nil
}
