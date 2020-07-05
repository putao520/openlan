package models

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/switch/schema"
)

func NewPointSchema(p *Point) schema.Point {
	client, dev := p.Client, p.Device
	return schema.Point{
		Uptime:    p.Uptime,
		UUID:      p.UUID,
		Alias:     p.Alias,
		User:      p.User,
		Address:   client.Addr(),
		Device:    dev.Name(),
		RxBytes:   client.Sts().RecvOkay,
		TxBytes:   client.Sts().SendOkay,
		ErrPkt:    client.Sts().SendError,
		State:     client.State(),
		Network:   p.Network,
		AliveTime: client.AliveTime(),
	}
}

func NewLinkSchema(p *Point) schema.Link {
	client, dev := p.Client, p.Device
	return schema.Link{
		UUID:      p.UUID,
		User:      p.User,
		Uptime:    client.UpTime(),
		Device:    dev.Name(),
		Address:   client.Addr(),
		State:     client.State(),
		Network:   p.Network,
		AliveTime: client.AliveTime(),
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
		HitTime:    l.LastTime(),
		UpTime:     l.UpTime(),
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
		Routes:  make([]schema.PrefixRoute, 0, 32),
	}
	for _, route := range n.Routes {
		sn.Routes = append(sn.Routes,
			schema.PrefixRoute{
				NextHop: route.NextHop,
				Prefix:  route.Prefix,
			})
	}
	return sn
}
