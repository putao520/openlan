package api

import (
	"github.com/danieldin95/openlan/src/config"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/network"
	"github.com/danieldin95/openlan/src/schema"
)

type Switcher interface {
	UUID() string
	UpTime() int64
	Alias() string
	Config() *config.Switch
	Server() libol.SocketServer
	AddLink(tenant string, c *config.Point)
	DelLink(tenant, addr string)
	AddEsp(tenant string, c *config.ESPSpecifies)
	DelEsp(tenant, c *config.ESPSpecifies)
	AddVxLAN(tenant string, c *config.VxLANSpecifies)
	DelVxLAN(tenant, c *config.VxLANSpecifies)
	Firewall() *network.FireWall
}

func NewWorkerSchema(s Switcher) schema.Worker {
	protocol := ""
	if cfg := s.Config(); cfg != nil {
		protocol = cfg.Protocol
	}
	return schema.Worker{
		UUID:     s.UUID(),
		Uptime:   s.UpTime(),
		Alias:    s.Alias(),
		Protocol: protocol,
	}
}
