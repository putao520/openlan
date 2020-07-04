package api

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/switch/schema"
)

type Switcher interface {
	UUID() string
	UpTime() int64
	Alias() string
	AddLink(tenant string, c *config.Point)
	DelLink(tenant, addr string)
	Config() *config.Switch
	Server() libol.SocketServer
}

func NewWorkerSchema(s Switcher) schema.Worker {
	return schema.Worker{
		UUID:   s.UUID(),
		Uptime: s.UpTime(),
		Alias:  s.Alias(),
	}
}
