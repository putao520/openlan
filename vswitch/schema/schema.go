package schema

import (
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"strings"
)

type Version struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Commit  string `json:"commit"`
}

func NewVersion() Version {
	return Version{
		Version: config.Version,
		Date:    config.Date,
		Commit:  config.Commit,
	}
}

type Worker struct {
	Uptime int64  `json:"uptime"`
	UUID   string `json:"uuid"`
	Alias  string `json:"alias"`
}

func NewWorker(sw VSwitcher) Worker {
	return Worker{
		UUID:   sw.UUID(),
		Uptime: sw.UpTime(),
		Alias:  sw.Alias(),
	}
}

type Point struct {
	Uptime  int64  `json:"uptime"`
	UUID    string `json:"uuid"`
	Network string `json:"network"`
	Alias   string `json:"alias"`
	Address string `json:"server"`
	IpAddr  string `json:"address"`
	Device  string `json:"device"`
	RxBytes uint64 `json:"rxBytes"`
	TxBytes uint64 `json:"txBytes"`
	ErrPkt  uint64 `json:"errors"`
	State   string `json:"state"`
}

func NewPoint(p *models.Point) Point {
	client, dev := p.Client, p.Device
	return Point{
		Uptime:  p.Uptime,
		UUID:    p.UUID,
		Alias:   p.Alias,
		Address: client.Addr,
		Device:  dev.Name(),
		RxBytes: client.Sts.RxOkay,
		TxBytes: client.Sts.TxOkay,
		ErrPkt:  client.Sts.TxError,
		State:   client.State(),
		Network: p.Network,
	}
}

type Link struct {
	Uptime  int64  `json:"uptime"`
	UUID    string `json:"uuid"`
	Alias   string `json:"alias"`
	Network string `json:"network"`
	Address string `json:"server"`
	IpAddr  string `json:"address"`
	Device  string `json:"device"`
	RxBytes uint64 `json:"rxBytes"`
	TxBytes uint64 `json:"txBytes"`
	ErrPkt  uint64 `json:"errors"`
	State   string `json:"state"`
}

func NewLink(p *models.Point) Link {
	client, dev := p.Client, p.Device
	return Link{
		UUID:    p.UUID,
		Uptime:  client.UpTime(),
		Device:  dev.Name(),
		Address: client.Addr,
		State:   client.State(),
		IpAddr:  strings.Split(client.Addr, ":")[0],
		Network: p.Network,
	}
}

type LinkConfig struct {
	config.Point
}

type Neighbor struct {
	Uptime int64  `json:"uptime"`
	UUID   string `json:"uuid"`
	HwAddr string `json:"ethernet"`
	IpAddr string `json:"address"`
	Client string `json:"client"`
}

func NewNeighbor(n *models.Neighbor) Neighbor {
	return Neighbor{
		Uptime: n.UpTime(),
		HwAddr: n.HwAddr.String(),
		IpAddr: n.IpAddr.String(),
		Client: n.Client.String(),
	}
}

type OnLine struct {
	Uptime     int64  `json:"uptime"`
	EthType    uint16 `json:"ethType"`
	IpSource   string `json:"ipSource"`
	IpDest     string `json:"ipDestination"`
	IpProto    string `json:"ipProtocol"`
	PortSource uint16 `json:"portSource"`
	PortDest   uint16 `json:"portDestination"`
}

func NewOnLine(l *models.Line) OnLine {
	return OnLine{
		Uptime:     l.UpTime(),
		EthType:    l.EthType,
		IpSource:   l.IpSource.String(),
		IpDest:     l.IpDest.String(),
		IpProto:    libol.IpProto2Str(l.IpProtocol),
		PortSource: l.PortSource,
		PortDest:   l.PortDest,
	}
}

type Index struct {
	Version   Version    `json:"version"`
	Worker    Worker     `json:"worker"`
	Points    []Point    `json:"points"`
	Links     []Link     `json:"links"`
	Neighbors []Neighbor `json:"neighbors"`
	OnLines   []OnLine   `json:"online"`
	Network   []Network  `json:"network"`
}

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Alias    string `json:"alias"`
}

func NewUser(u *models.User) User {
	return User{
		Name:     u.Name,
		Password: u.Password,
		Token:    u.Token,
		Alias:    u.Alias,
	}
}

func (user *User) ToModel() *models.User {
	return &models.User{
		Alias:    user.Alias,
		Token:    user.Token,
		Password: user.Password,
		Name:     user.Name,
	}
}

type Route struct {
	Prefix  string `json:"prefix"`
	Nexthop string `json:"nexthop"`
}

type Network struct {
	Name    string  `json:"name"`
	IfAddr  string  `json:"ifAddr"`
	IpAddr  string  `json:"ipAddr"`
	IpRange int     `json:"ipRange"`
	Netmask string  `json:"netmask"`
	Routes  []Route `json:"routes"`
}

func NewNetwork(n *models.Network) Network {
	routes := make([]Route, 0, 32)
	if len(n.Routes) != 0 {
		for _, route := range routes {
			routes = append(routes,
				Route{
					Nexthop: route.Nexthop,
					Prefix:  route.Prefix,
				})
		}
	}
	return Network{
		Name:    n.Name,
		IfAddr:  n.IfAddr,
		IpAddr:  n.IpAddr,
		IpRange: n.IpRange,
		Netmask: n.Netmask,
		Routes:  routes,
	}
}

type Ctrl struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}
