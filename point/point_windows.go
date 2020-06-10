package point

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/models"
	"strings"
)

type Point struct {
	MixPoint
	BrName string

	addr   string
	routes []*models.Route
	config *config.Point
}

func NewPoint(config *config.Point) *Point {
	p := Point{
		BrName:   config.If.Bridge,
		MixPoint: NewMixPoint(config),
	}
	return &p
}

func (p *Point) Initialize() {
	p.worker.Listener.AddAddr = p.AddAddr
	p.worker.Listener.DelAddr = p.DelAddr
	p.worker.Listener.AddRoutes = p.AddRoutes
	p.worker.Listener.DelRoutes = p.DelRoutes
	p.worker.Listener.OnTap = p.OnTap
	p.MixPoint.Initialize()
}

func (p *Point) Start() {
	libol.Info("Point.Start: Windows.")
	if !p.initialize {
		p.Initialize()
	}
	p.worker.Start()
}

func (p *Point) Stop() {
	defer libol.Catch("Point.Stop")
	p.worker.Stop()
}

func (p *Point) OnTap(w *TapWorker) error {
	// clean routes previous
	routes := make([]*models.Route, 0, 32)
	if err := libol.UnmarshalLoad(&routes, ".routes.json"); err == nil {
		for _, route := range routes {
			_, _ = libol.IpRouteDel(p.IfName(), route.Prefix, route.Nexthop)
			libol.Info("Point.OnTap: clear %s via %s", route.Prefix, route.Nexthop)
		}
	}
	return nil
}

func (p *Point) AddAddr(ipStr string) error {
	if ipStr == "" {
		return nil
	}

	addrExisted := libol.IpAddrShow(p.IfName())
	if len(addrExisted) > 0 {
		for _, addr := range addrExisted {
			_, _ = libol.IpAddrDel(p.IfName(), addr)
		}
	}
	out, err := libol.IpAddrAdd(p.IfName(), ipStr)
	if err != nil {
		libol.Warn("Point.AddAddr: %s, %s", err, out)
		return err
	}
	libol.Info("Point.AddAddr: %s", ipStr)
	p.addr = ipStr

	return nil
}

func (p *Point) DelAddr(ipStr string) error {
	ipv4 := strings.Split(ipStr, "/")[0]
	out, err := libol.IpAddrDel(p.IfName(), ipv4)
	if err != nil {
		libol.Warn("Point.DelAddr: %s, %s", err, out)
		return err
	}
	libol.Info("Point.DelAddr: %s", ipv4)
	p.addr = ""

	return nil
}

func (p *Point) AddRoutes(routes []*models.Route) error {
	if routes == nil {
		return nil
	}

	_ = libol.MarshalSave(routes, ".routes.json", true)
	for _, route := range routes {
		out, err := libol.IpRouteAdd(p.IfName(), route.Prefix, route.Nexthop)
		if err != nil {
			libol.Warn("Point.AddRoutes: %s, %s", err, out)
			continue
		}
		libol.Info("Point.AddRoutes: route %s via %s", route.Prefix, route.Nexthop)
	}
	p.routes = routes
	return nil
}

func (p *Point) DelRoutes(routes []*models.Route) error {
	for _, route := range routes {
		out, err := libol.IpRouteDel(p.IfName(), route.Prefix, route.Nexthop)
		if err != nil {
			libol.Warn("Point.DelRoutes: %s, %s", err, out)
			continue
		}
		libol.Info("Point.DelRoutes: route %s via %s", route.Prefix, route.Nexthop)
	}
	p.routes = nil
	return nil
}
