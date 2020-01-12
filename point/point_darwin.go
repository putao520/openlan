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
	uuid      string
}

func NewPoint(config *config.Point) (p *Point) {
	p = &Point{
		BrName: config.BrName,
		IfAddr: config.IfAddr,
		config: config,
	}

	return
}

func (p *Point) Initialize() {
	var tlsConf *tls.Config
	if p.config.Tls {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}

	client := libol.NewTcpClient(p.config.Addr, tlsConf)
	p.tcpWorker = NewTcpWorker(client, p.config)

	conf := &water.Config{DeviceType: water.TUN}
	p.tapWorker = NewTapWorker(conf, p.config)
}

func (p *Point) OnTap(w *TapWorker) error {
	return nil
}

func (p *Point) Start() {
	libol.Debug("Point.Start Darwin.")

	if p.tapWorker != nil || p.tcpWorker != nil {
		return
	}

	p.Initialize()

	p.tcpWorker.SetUUID(p.UUID())
	p.tcpWorker.Listener = TcpWorkerListener{
		OnClose:   p.OnClose,
		OnSuccess: p.OnSuccess,
		OnIpAddr:  p.OnIpAddr,
		ReadAt:    p.tapWorker.DoWrite,
	}
	p.tapWorker.Listener = TapWorkerListener{
		OnOpen: p.OnTap,
		ReadAt: p.tcpWorker.DoWrite,
	}

	p.tapWorker.Start(p)
	p.tcpWorker.Start(p)
}

func (p *Point) Stop() {
	if p.tapWorker == nil || p.tcpWorker == nil {
		return
	}

	defer libol.Catch("Point.Stop")

	p.tapWorker.Stop()
	p.tcpWorker.Stop()

	p.tapWorker = nil
	p.tcpWorker = nil
}

func (p *Point) Client() *libol.TcpClient {
	if p.tcpWorker != nil {
		return p.tcpWorker.Client
	}
	return nil
}

func (p *Point) Device() network.Taper {
	if p.tapWorker != nil {
		return p.tapWorker.Device
	}
	return nil
}

func (p *Point) UpTime() int64 {
	client := p.Client()
	if client != nil {
		return client.UpTime()
	}
	return 0
}

func (p *Point) State() string {
	client := p.Client()
	if client != nil {
		return client.State()
	}
	return ""
}

func (p *Point) Addr() string {
	client := p.Client()
	if client != nil {
		return client.Addr
	}
	return ""
}

func (p *Point) IfName() string {
	dev := p.Device()
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

func (p *Point) UUID() string {
	if p.uuid == "" {
		p.uuid = libol.GenToken(32)
	}
	return p.uuid
}
