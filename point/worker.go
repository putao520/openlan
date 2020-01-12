package point

import (
	"crypto/tls"
	"fmt"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/network"
	"github.com/songgao/water"
)

type WorkerListener struct {
    AddAddr func(ipStr string) error
	DelAddr func(ipStr string) error
	OnTap func(w *TapWorker) error
	AddRoutes func(routes []*models.Route) error
	DelRoutes func(routes []*models.Route) error
}

type Worker struct {
	IfAddr   string
	Listener WorkerListener

	tcpWorker *TcpWorker
	tapWorker *TapWorker
	config    *config.Point
	uuid      string
}

func NewWorker(config *config.Point) (p *Worker) {
	p = &Worker{
		IfAddr: config.IfAddr,
		config: config,
	}

	return
}

func (p *Worker) Initialize() {
	if p.config == nil {
		return
	}

	var conf *water.Config
	var tlsConf *tls.Config

	if p.config.Tls {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}
	client := libol.NewTcpClient(p.config.Addr, tlsConf)
	p.tcpWorker = NewTcpWorker(client, p.config)

	if p.config.IfTun {
		conf = &water.Config{DeviceType: water.TUN}
	} else {
		conf = &water.Config{DeviceType: water.TAP}
	}
	p.tapWorker = NewTapWorker(conf, p.config)
}

func (p *Worker) Start() {
	libol.Debug("Worker.Start linux.")
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
		OnOpen: p.Listener.OnTap,
		ReadAt: p.tcpWorker.DoWrite,
	}
	p.tapWorker.Start()
	p.tcpWorker.Start()
}

func (p *Worker) Stop() {
	if p.tapWorker == nil || p.tcpWorker == nil {
		return
	}

	if p.Listener.DelRoutes != nil {
		p.Listener.DelRoutes(nil)
	}
	if p.Listener.DelAddr != nil {
		p.Listener.DelAddr("")
	}
	p.tcpWorker.Stop()
	p.tapWorker.Stop()
	p.tcpWorker = nil
	p.tapWorker = nil
}

func (p *Worker) Client() *libol.TcpClient {
	if p.tcpWorker != nil {
		return p.tcpWorker.Client
	}
	return nil
}

func (p *Worker) Device() network.Taper {
	if p.tapWorker != nil {
		return p.tapWorker.Device
	}
	return nil
}

func (p *Worker) UpTime() int64 {
	client := p.Client()
	if client != nil {
		return client.UpTime()
	}
	return 0
}

func (p *Worker) State() string {
	client := p.Client()
	if client != nil {
		return client.State()
	}
	return ""
}

func (p *Worker) Addr() string {
	client := p.Client()
	if client != nil {
		return client.Addr
	}
	return ""
}

func (p *Worker) IfName() string {
	dev := p.Device()
	if dev != nil {
		return dev.Name()
	}
	return ""
}

func (p *Worker) Worker() *TcpWorker {
	if p.tcpWorker != nil {
		return p.tcpWorker
	}
	return nil
}

func (p *Worker) OnIpAddr(w *TcpWorker, n *models.Network) error {
	libol.Info("Worker.OnIpAddr: %s/%s, %s", n.IfAddr, n.Netmask, n.Routes)

	prefix := libol.Netmask2Len(n.Netmask)
	ipStr := fmt.Sprintf("%s/%d", n.IfAddr, prefix)

	if p.Listener.AddAddr != nil {
		p.Listener.AddAddr(ipStr)
	}
	if p.Listener.AddRoutes != nil {
		p.Listener.AddRoutes(n.Routes)
	}

	return nil
}

func (p *Worker) OnClose(w *TcpWorker) error {
	libol.Info("Worker.OnClose")

	if p.Listener.DelRoutes != nil {
		p.Listener.DelRoutes(nil)
	}
	if p.Listener.DelAddr != nil {
		p.Listener.DelAddr("")
	}

	return nil
}

func (p *Worker) OnSuccess(w *TcpWorker) error {
	libol.Info("Worker.OnSuccess")

	if p.Listener.AddAddr != nil {
		p.Listener.AddAddr(p.IfAddr)
	}

	return nil
}

func (p *Worker) UUID() string {
	if p.uuid == "" {
		p.uuid = libol.GenToken(32)
	}
	return p.uuid
}

func (p *Worker) SetUUID(v string) {
	p.uuid = v
}