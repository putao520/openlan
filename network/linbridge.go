package network

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/milosgajdos83/tenus"
	"net"
)

type LinBridge struct {
	ip     net.IP
	net    *net.IPNet
	mtu    int
	name   string
	device tenus.Bridger
}

func NewLinBridge(name string, mtu int) *LinBridge {
	b := &LinBridge{
		name: name,
		mtu:  mtu,
	}
	return b
}

func (b *LinBridge) Open(addr string) {
	var err error
	var br tenus.Bridger

	brName := b.name
	br, err = tenus.BridgeFromName(brName)
	if err != nil {
		br, err = tenus.NewBridgeWithName(brName)
		if err != nil {
			libol.Error("LinBridge.newBr: %s", err)
		}
	}

	brCtl := libol.NewBrCtl(brName)
	if err := brCtl.Stp(true); err != nil {
		libol.Error("LinBridge.newBr.Stp: %s", err)
	}

	if err = br.SetLinkUp(); err != nil {
		libol.Error("LinBridge.newBr: %s", err)
	}

	libol.Info("LinBridge.newBr %s", brName)

	if addr != "" {
		ip, net, err := net.ParseCIDR(addr)
		if err != nil {
			libol.Error("LinBridge.newBr.ParseCIDR %s : %s", addr, err)
		}
		if err := br.SetLinkIp(ip, net); err != nil {
			libol.Error("LinBridge.newBr.SetLinkIp %s : %s", brName, err)
		}

		b.ip = ip
		b.net = net
	}

	b.device = br
}

func (b *LinBridge) Close() {
	if b.device != nil && b.ip != nil {
		if err := b.device.UnsetLinkIp(b.ip, b.net); err != nil {
			libol.Error("LinBridge.Close.UnsetLinkIp %s : %s", b.name, err)
		}
	}
}

func (b *LinBridge) AddSlave(name string) error {
	link, err := tenus.NewLinkFrom(name)
	if err != nil {
		libol.Error("LinBridge.AddSlave: Get dev %s: %s", name, err)
		return err
	}

	if err := link.SetLinkUp(); err != nil {
		libol.Error("LinBridge.AddSlave.LinkUp: ", err)
	}

	if err := b.device.AddSlaveIfc(link.NetInterface()); err != nil {
		libol.Error("LinBridge.AddSlave: Switch dev %s: %s", name, err)
		return err
	}

	return nil
}

func (b *LinBridge) Name() string {
	return b.name
}

func (b *LinBridge) SetName(value string) {
	b.name = value
}
