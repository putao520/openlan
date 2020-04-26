package point

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/network"
)

type Pointer interface {
	State() string
	Addr() string
	IfName() string
	IfAddr() string
	Client() *libol.TcpClient
	Device() network.Taper
	UpTime() int64
	UUID() string
}

type MixPoint struct {
	Tenant     string
	uuid       string
	worker     Worker
	config     *config.Point
	initialize bool
}

func NewMixPoint(config *config.Point) MixPoint {
	p := MixPoint{
		Tenant: config.Network,
		worker: Worker{
			IfAddr: config.If.Address,
			config: config,
		},
		config:     config,
		initialize: false,
	}
	return p
}

func (p *MixPoint) Initialize() {
	libol.Info("MixPoint.Initialize")
	p.initialize = true
	p.worker.SetUUID(p.UUID())
	p.worker.Initialize()
}

func (p *MixPoint) UUID() string {
	if p.uuid == "" {
		p.uuid = libol.GenToken(32)
	}
	return p.uuid
}

func (p *MixPoint) State() string {
	return p.worker.State()
}

func (p *MixPoint) Addr() string {
	return p.worker.Addr()
}

func (p *MixPoint) IfName() string {
	return p.worker.IfName()
}

func (p *MixPoint) Client() *libol.TcpClient {
	return p.worker.Client()
}

func (p *MixPoint) Device() network.Taper {
	return p.worker.Device()
}

func (p *MixPoint) UpTime() int64 {
	return p.worker.UpTime()
}

func (p *MixPoint) IfAddr() string {
	return p.worker.IfAddr
}
