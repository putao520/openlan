package storage

import (
	"github.com/danieldin95/openlan-go/src/config"
)

func Init(cfg *config.Perf) {
	Point.Init(cfg.Point)
	Link.Init(cfg.Link)
	Neighbor.Init(cfg.Neighbor)
	Online.Init(cfg.OnLine)
	User.Init(cfg.User)
}
