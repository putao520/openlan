package models

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olsw/schema"
)

func NewPointSchema(p *Point) schema.Point {
	client, dev := p.Client, p.Device
	sts := client.Statistics()
	return schema.Point{
		Uptime:    p.Uptime,
		UUID:      p.UUID,
		Alias:     p.Alias,
		User:      p.User,
		Address:   client.String(),
		Device:    dev.Name(),
		RxBytes:   sts[libol.CsRecvOkay],
		TxBytes:   sts[libol.CsSendOkay],
		ErrPkt:    sts[libol.CsSendError],
		State:     client.Status().String(),
		Network:   p.Network,
		AliveTime: client.AliveTime(),
		System:    p.System,
	}
}

func NewLinkSchema(p *Point) schema.Link {
	client, dev := p.Client, p.Device
	sts := client.Statistics()
	return schema.Link{
		UUID:      p.UUID,
		User:      p.User,
		Uptime:    client.UpTime(),
		Device:    dev.Name(),
		Address:   client.String(),
		State:     client.Status().String(),
		RxBytes:   sts[libol.CsRecvOkay],
		TxBytes:   sts[libol.CsSendOkay],
		ErrPkt:    sts[libol.CsSendError],
		Network:   p.Network,
		AliveTime: client.AliveTime(),
	}
}

func NewNeighborSchema(n *Neighbor) schema.Neighbor {
	return schema.Neighbor{
		Uptime:  n.UpTime(),
		HwAddr:  n.HwAddr.String(),
		IpAddr:  n.IpAddr.String(),
		Client:  n.Client,
		Network: n.Network,
		Device:  n.Device,
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
