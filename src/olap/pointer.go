package olap

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/network"
)

type Pointer interface {
	Addr() string
	IfName() string
	IfAddr() string
	Client() libol.SocketClient
	Device() network.Taper
	Status() libol.SocketStatus
	UpTime() int64
	UUID() string
	User() string
	Record() map[string]int64
	Tenant() string
	Alias() string
	Config() *config.Point
}

type MixPoint struct {
	uuid   string
	worker *Worker
	config *config.Point
	out    *libol.SubLogger
}

func NewMixPoint(config *config.Point) MixPoint {
	return MixPoint{
		worker: NewWorker(config),
		config: config,
		out:    libol.NewSubLogger(config.Id()),
	}
}

func (p *MixPoint) Initialize() {
	libol.Info("MixPoint.Initialize")
	p.worker.SetUUID(p.UUID())
	p.worker.Initialize()
}

func (p *MixPoint) UUID() string {
	if p.uuid == "" {
		p.uuid = libol.GenToken(32)
	}
	return p.uuid
}

func (p *MixPoint) Status() libol.SocketStatus {
	return p.worker.Status()
}

func (p *MixPoint) Addr() string {
	return p.worker.Addr()
}

func (p *MixPoint) IfName() string {
	return p.worker.IfName()
}

func (p *MixPoint) Client() libol.SocketClient {
	return p.worker.Client()
}

func (p *MixPoint) Device() network.Taper {
	return p.worker.Device()
}

func (p *MixPoint) UpTime() int64 {
	return p.worker.UpTime()
}

func (p *MixPoint) IfAddr() string {
	return p.worker.ifAddr
}

func (p *MixPoint) Tenant() string {
	return p.config.Network
}

func (p *MixPoint) User() string {
	return p.config.Username
}

func (p *MixPoint) Alias() string {
	return p.config.Alias
}

func (p *MixPoint) Record() map[string]int64 {
	rt := p.worker.conWorker.record
	// TODO padding data from tapWorker
	return rt.Data()
}

func (p *MixPoint) Config() *config.Point {
	return p.config
}
