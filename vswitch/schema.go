package vswitch

import (
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"strings"
)

type VersionSchema struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Commit  string `json:"commit"`
}

func NewVersionSchema() VersionSchema {
	return VersionSchema{
		Version: config.Version,
		Date:    config.Date,
		Commit:  config.Commit,
	}
}

type WorkerSchema struct {
	Uptime int64  `json:"uptime"`
	UUID   string `json:"uuid"`
	Alias  string `json:"alias"`
}

func NewWorkerSchema(w *Worker) WorkerSchema {
	return WorkerSchema{
		UUID:   w.UUID(),
		Uptime: w.UpTime(),
		Alias:  w.Alias,
	}
}

type PointSchema struct {
	Uptime  int64  `json:"uptime"`
	UUID    string `json:"uuid"`
	Alias   string `json:"alias"`
	Address string `json:"server"`
	IpAddr  string `json:"address"`
	Device  string `json:"device"`
	RxBytes uint64 `json:"rxBytes"`
	TxBytes uint64 `json:"txBytes"`
	ErrPkt  uint64 `json:"errors"`
	State   string `json:"state"`
}

func NewPointSchema(p *models.Point) PointSchema {
	client, dev := p.Client, p.Device
	return PointSchema{
		Uptime:  p.Uptime,
		UUID:    p.UUID,
		Alias:   p.Alias,
		Address: client.Addr,
		Device:  dev.Name(),
		RxBytes: client.Sts.RxOkay,
		TxBytes: client.Sts.TxOkay,
		ErrPkt:  client.Sts.TxError,
		State:   client.State(),
	}
}

type LinkSchema struct {
	Uptime  int64  `json:"uptime"`
	UUID    string `json:"uuid"`
	Alias   string `json:"alias"`
	Address string `json:"server"`
	IpAddr  string `json:"address"`
	Device  string `json:"device"`
	RxBytes uint64 `json:"rxBytes"`
	TxBytes uint64 `json:"txBytes"`
	ErrPkt  uint64 `json:"errors"`
	State   string `json:"state"`
}

func NewLinkSchema(p *models.Point) LinkSchema {
	client, dev := p.Client, p.Device
	return LinkSchema{
		UUID:    p.UUID,
		Uptime:  client.UpTime(),
		Device:  dev.Name(),
		Address: client.Addr,
		State:   client.State(),
		IpAddr:  strings.Split(client.Addr, ":")[0],
	}
}

type LinkConfigSchema struct {
	config.Point
}

type NeighborSchema struct {
	Uptime int64  `json:"uptime"`
	UUID   string `json:"uuid"`
	HwAddr string `json:"ethernet"`
	IpAddr string `json:"address"`
	Client string `json:"client"`
}

func NewNeighborSchema(n *models.Neighbor) NeighborSchema {
	return NeighborSchema{
		Uptime: n.UpTime(),
		HwAddr: n.HwAddr.String(),
		IpAddr: n.IpAddr.String(),
		Client: n.Client.String(),
	}
}

type OnLineSchema struct {
	Uptime     int64  `json:"uptime"`
	EthType    uint16 `json:"ethType"`
	IpSource   string `json:"ipSource"`
	IpDest     string `json:"ipDestination"`
	IpProto    string `json:"ipProtocol"`
	PortSource uint16 `json:"portSource"`
	PortDest   uint16 `json:"portDestination"`
}

func NewOnLineSchema(l *models.Line) OnLineSchema {
	return OnLineSchema{
		Uptime:     l.UpTime(),
		EthType:    l.EthType,
		IpSource:   l.IpSource.String(),
		IpDest:     l.IpDest.String(),
		IpProto:    libol.IpProto2Str(l.IpProtocol),
		PortSource: l.PortSource,
		PortDest:   l.PortDest,
	}
}

type IndexSchema struct {
	Version   VersionSchema    `json:"version"`
	Worker    WorkerSchema     `json:"worker"`
	Points    []PointSchema    `json:"points"`
	Links     []LinkSchema     `json:"links"`
	Neighbors []NeighborSchema `json:"neighbors"`
	OnLines   []OnLineSchema   `json:"online"`
	Network   []NetworkSchema  `json:"network"`
}

type UserSchema struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Alias    string `json:"alias"`
}

func NewUserSchema(u *models.User) UserSchema {
	return UserSchema{
		Name:     u.Name,
		Password: u.Password,
		Token:    u.Token,
		Alias:    u.Alias,
	}
}

func (user *UserSchema) ToModel() *models.User {
	return &models.User{
		Alias:    user.Alias,
		Token:    user.Token,
		Password: user.Password,
		Name:     user.Name,
	}
}

type RouteSchema struct {
	Prefix  string `json:"prefix"`
	Nexthop string `json:"nexthop"`
}

type NetworkSchema struct {
	Tenant  string        `json:"tenant"`
	IfAddr  string        `json:"ifAddr"`
	IpAddr  string        `json:"ipAddr"`
	IpRange int           `json:"ipRange"`
	Netmask string        `json:"netmask"`
	Routes  []RouteSchema `json:"routes"`
}

func NewNetworkSchema(n *models.Network) NetworkSchema {
	routes := make([]RouteSchema, 0, 32)
	if len(n.Routes) != 0 {
		for _, route := range routes {
			routes = append(routes,
				RouteSchema{
					Nexthop: route.Nexthop,
					Prefix:  route.Prefix,
				})
		}
	}
	return NetworkSchema{
		Tenant:  n.Tenant,
		IfAddr:  n.IfAddr,
		IpAddr:  n.IpAddr,
		IpRange: n.IpRange,
		Netmask: n.Netmask,
		Routes:  routes,
	}
}
