package point

import (
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
)

type Point struct {
	MixPoint
	BrName string
}

func NewPoint(config *config.Point) *Point {
	p := Point{
		BrName:   config.BrName,
		MixPoint: NewMixPoint(config),
	}
	return &p
}

func (p *Point) Start() {
	libol.Debug("Point.Start Darwin.")
	p.Initialize()
	p.worker.Start()
}

func (p *Point) Stop() {
	defer libol.Catch("Point.Stop")
	p.worker.Stop()
}
