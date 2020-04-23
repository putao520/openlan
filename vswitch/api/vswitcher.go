package api

import (
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/vswitch/schema"
)

type VSwitcher interface {
	UUID() string
	UpTime() int64
	Alias() string
	AddLink(tenant string, c *config.Point)
	DelLink(tenant, addr string)
}

func NewWorkerSchema(sw VSwitcher) schema.Worker {
	return schema.Worker{
		UUID:   sw.UUID(),
		Uptime: sw.UpTime(),
		Alias:  sw.Alias(),
	}
}
