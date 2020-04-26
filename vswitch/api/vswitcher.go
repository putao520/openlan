package api

import (
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/vswitch/schema"
)

type VSwitcher interface {
	UUID() string
	UpTime() int64
	Alias() string
	AddLink(tenant string, c *config.Point)
	DelLink(tenant, addr string)
	Config() *config.VSwitch
}

func NewWorkerSchema(sw VSwitcher) schema.Worker {
	return schema.Worker{
		UUID:   sw.UUID(),
		Uptime: sw.UpTime(),
		Alias:  sw.Alias(),
	}
}
