package olp

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/network"
)

type Pointer interface {
	State() string
	Addr() string
	IfName() string
	IfAddr() string
	Client() libol.SocketClient
	Device() network.Taper
	UpTime() int64
	UUID() string
	User() string
}

type MixPoint struct {
	// private
	uuid   string
	worker *Worker
	config *config.Point
	out    *libol.SubLogger
}

func NewMixPoint(config *config.Point) MixPoint {
	p := MixPoint{
		worker: NewWorker(config),
		config: config,
		out:    libol.NewSubLogger(config.Id()),
	}
	return p
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

func (p *MixPoint) State() string {
	return p.worker.State()
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
	if p.config != nil {
		return p.config.Network
	}
	return ""
}

func (p *MixPoint) User() string {
	if p.config != nil {
		return p.config.Username
	}
	return ""
}
