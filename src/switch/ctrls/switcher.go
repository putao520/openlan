package ctrls

import "github.com/danieldin95/openlan-go/src/cli/config"

type Switcher interface {
	UUID() string
	UpTime() int64
	Alias() string
	AddLink(tenant string, c *config.Point)
	DelLink(tenant, addr string)
}
