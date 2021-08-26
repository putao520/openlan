package ctrls

import "github.com/danieldin95/openlan/pkg/config"

type Switcher interface {
	UUID() string
	UpTime() int64
	Alias() string
	AddLink(tenant string, c *config.Point)
	DelLink(tenant, addr string)
}
