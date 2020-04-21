package schema

import "github.com/danieldin95/openlan-go/config"

type VSwitcher interface {
	UUID() string
	UpTime() int64
	Alias() string
	AddLink(tenant string, c *config.Point)
	DelLink(tenant, addr string)
}
