package point

import (
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
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
		BrName:   config.BrName,
		MixPoint: NewMixPoint(config),
	}
	return &p
}

func (p *Point) Initialize() {
	p.worker.Listener.AddAddr = p.AddAddr
	p.worker.Listener.DelAddr = p.DelAddr
	p.worker.Listener.AddRoutes = p.AddRoutes
	p.worker.Listener.DelRoutes = p.DelRoutes
	p.MixPoint.Initialize()
}

func (p *Point) Start() {
	libol.Info("Point.Start Windows.")
	if !p.initialize {
		p.Initialize()
	}
	p.worker.Start()
}

func (p *Point) Stop() {
	defer libol.Catch("Point.Stop")
	p.worker.Stop()
}

func (p *Point) AddAddr(ipStr string) error {
	if ipStr == "" {
		return nil
	}

	addrExisted := libol.IpAddrShow(p.IfName())
	if len(addrExisted) > 0 {
		for _, addr := range addrExisted {
			libol.IpAddrDel(p.IfName(), addr)
		}
	}
	out, err := libol.IpAddrAdd(p.IfName(), ipStr)
	if err != nil {
		libol.Error("Point.AddAddr: %s, %s", err, out)
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
		libol.Error("Point.DelAddr: %s, %s", err, out)
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

	for _, route := range routes {
		out, err := libol.IpRouteAdd(p.IfName(), route.Prefix, route.Nexthop)
		if err != nil {
			libol.Error("Point.AddRoutes: %s, %s", err, out)
			continue
		}
		libol.Info("Point.AddRoutes: %s via %s", route.Prefix, route.Nexthop)
	}

	p.routes = routes
	return nil
}

func (p *Point) DelRoutes(routes []*models.Route) error {
	for _, route := range routes {
		out, err := libol.IpRouteDel(p.IfName(), route.Prefix, route.Nexthop)
		if err != nil {
			libol.Error("Point.DelRoutes: %s, %s", err, out)
			continue
		}
		libol.Info("Point.DelRoutes: %s via %s", route.Prefix, route.Nexthop)
	}

	p.routes = nil

	return nil
}
