package point

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"strings"
)

type Point struct {
	MixPoint
	// private
	brName string
	addr   string
	routes []*models.Route
	config *config.Point
}

func NewPoint(config *config.Point) *Point {
	p := Point{
		brName:   config.Interface.Bridge,
		MixPoint: NewMixPoint(config),
	}
	return &p
}

func (p *Point) Initialize() {
	p.worker.listener.AddAddr = p.AddAddr
	p.worker.listener.DelAddr = p.DelAddr
	p.worker.listener.AddRoutes = p.AddRoutes
	p.worker.listener.DelRoutes = p.DelRoutes
	p.worker.listener.OnTap = p.OnTap
	p.MixPoint.Initialize()
}

func (p *Point) Start() {
	libol.Info("Point.Start: Windows.")
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
			_, _ = libol.IpRouteDel(p.IfName(), route.Prefix, route.NextHop)
			libol.Info("Point.OnTap: clear %s via %s", route.Prefix, route.NextHop)
		}
	}
	if out, err := libol.IpMetricSet(p.IfName(), "235"); err != nil {
		libol.Warn("Point.OnTap: metricSet %s: %s", err, out)
	}
	return nil
}

func (p *Point) Trim(out []byte) string {
	return strings.TrimSpace(string(out))
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
		libol.Warn("Point.AddAddr: %s, %s", err, p.Trim(out))
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
		libol.Warn("Point.DelAddr: %s, %s", err, p.Trim(out))
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
		out, err := libol.IpRouteAdd(p.IfName(), route.Prefix, route.NextHop)
		if err != nil {
			libol.Warn("Point.AddRoutes: %s, %s", err, p.Trim(out))
			continue
		}
		libol.Info("Point.AddRoutes: route %s via %s", route.Prefix, route.NextHop)
	}
	p.routes = routes
	return nil
}

func (p *Point) DelRoutes(routes []*models.Route) error {
	for _, route := range routes {
		out, err := libol.IpRouteDel(p.IfName(), route.Prefix, route.NextHop)
		if err != nil {
			libol.Warn("Point.DelRoutes: %s, %s", err, p.Trim(out))
			continue
		}
		libol.Info("Point.DelRoutes: route %s via %s", route.Prefix, route.NextHop)
	}
	p.routes = nil
	return nil
}
