package models

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/vswitch/schema"
	"strings"
)

func NewPointSchema(p *Point) schema.Point {
	client, dev := p.Client, p.Device
	return schema.Point{
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

func NewLinkSchema(p *Point) schema.Link {
	client, dev := p.Client, p.Device
	return schema.Link{
		UUID:    p.UUID,
		Uptime:  client.UpTime(),
		Device:  dev.Name(),
		Address: client.Addr,
		State:   client.State(),
		IpAddr:  strings.Split(client.Addr, ":")[0],
		Network: p.Network,
	}
}

func NewNeighborSchema(n *Neighbor) schema.Neighbor {
	return schema.Neighbor{
		Uptime: n.UpTime(),
		HwAddr: n.HwAddr.String(),
		IpAddr: n.IpAddr.String(),
		Client: n.Client.String(),
	}
}

func NewOnLineSchema(l *Line) schema.OnLine {
	return schema.OnLine{
		Uptime:     l.UpTime(),
		EthType:    l.EthType,
		IpSource:   l.IpSource.String(),
		IpDest:     l.IpDest.String(),
		IpProto:    libol.IpProto2Str(l.IpProtocol),
		PortSource: l.PortSource,
		PortDest:   l.PortDest,
	}
}

func NewUserSchema(u *User) schema.User {
	return schema.User{
		Name:     u.Name,
		Password: u.Password,
		Token:    u.Token,
		Alias:    u.Alias,
	}
}

func SchemaToUserModel(user *schema.User) *User {
	return &User{
		Alias:    user.Alias,
		Token:    user.Token,
		Password: user.Password,
		Name:     user.Name,
	}
}

func NewNetworkSchema(n *Network) schema.Network {
	sn := schema.Network{
		Name:    n.Name,
		IpStart: n.IpStart,
		IpEnd:   n.IpEnd,
		Netmask: n.Netmask,
		Routes:  make([]schema.Route, 0, 32),
	}
	for _, route := range n.Routes {
		sn.Routes = append(sn.Routes,
			schema.Route{
				Nexthop: route.Nexthop,
				Prefix:  route.Prefix,
			})
	}
	return sn
}
