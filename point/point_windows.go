package point

import (
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"os/exec"
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
	p.MixPoint.Initialize()
	p.worker.Listener.AddAddr = p.AddAddr
	p.worker.Listener.DelAddr = p.DelAddr
	p.worker.Listener.AddRoutes = p.AddRoutes
	p.worker.Listener.DelRoutes = p.DelRoutes
}

func (p *Point) Start() {
	libol.Debug("Point.Start Windows.")
	p.Initialize()
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

	out, err := exec.Command("netsh", "interface",
		"ipv4", "add", "address", p.IfName(),
		ipStr, "store=active").Output()
	if err != nil {
		libol.Error("Point.AddAddr: %s, %s", err, out)
		return err
	}

	libol.Info("Point.AddAddr: %s", ipStr)
	p.addr = ipStr

	return nil
}

func (p *Point) DelAddr(ipStr string) error {
	if ipStr == "" {
		return nil
	}

	ipAddr := strings.Split(ipStr, "/")[0]
	out, err := exec.Command("netsh", "interface",
		"ipv4", "delete", "address", p.IfName(),
		ipAddr, "store=active").Output()
	if err != nil {
		libol.Error("Point.DelAddr: %s, %s", err, out)
		return err
	}
	libol.Info("Point.DelAddr: %s", ipStr)
	p.addr = ""

	return nil
}

func (p *Point) AddRoutes(routes []*models.Route) error {
	if routes == nil {
		return nil
	}

	for _, route := range routes {
		out, err := exec.Command("netsh", "interface",
			"ipv4", "add", "route", route.Prefix,
			p.IfName(), route.Nexthop, "store=active").Output()
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
	if routes == nil {
		return nil
	}

	for _, route := range routes {
		out, err := exec.Command("netsh", "interface",
			"ipv4", "del", "route", route.Prefix,
			p.IfName(), route.Nexthop, "store=active").Output()
		if err != nil {
			libol.Error("Point.DelRoutes: %s, %s", err, out)
			continue
		}
		libol.Info("Point.DelRoutes: %s via %s", route.Prefix, route.Nexthop)
	}

	p.routes = nil

	return nil
}
