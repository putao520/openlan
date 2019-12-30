package vswitch

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/milosgajdos83/tenus"
	"net"
)

type LinuxBridger struct {
	ip     net.IP
	net    *net.IPNet
	mtu    int
	name   string
	device tenus.Bridger
}

func NewLinuxBridger(name string, mtu int) *LinuxBridger {
	b := &LinuxBridger{
		name: name,
		mtu:  mtu,
	}
	return b
}

func (b *LinuxBridger) Open(addr string) {
	var err error
	var br tenus.Bridger

	brName := b.name
	br, err = tenus.BridgeFromName(brName)
	if err != nil {
		br, err = tenus.NewBridgeWithName(brName)
		if err != nil {
			libol.Error("LinuxBridger.newBr: %s", err)
		}
	}

	brCtl := libol.NewBrCtl(brName)
	if err := brCtl.Stp(true); err != nil {
		libol.Error("LinuxBridger.newBr.Stp: %s", err)
	}

	if err = br.SetLinkUp(); err != nil {
		libol.Error("LinuxBridger.newBr: %s", err)
	}

	libol.Info("LinuxBridger.newBr %s", brName)

	if addr != "" {
		ip, net, err := net.ParseCIDR(addr)
		if err != nil {
			libol.Error("LinuxBridger.newBr.ParseCIDR %s : %s", addr, err)
		}
		if err := br.SetLinkIp(ip, net); err != nil {
			libol.Error("LinuxBridger.newBr.SetLinkIp %s : %s", brName, err)
		}

		b.ip = ip
		b.net = net
	}

	b.device = br
}

func (b *LinuxBridger) Close() {
	if b.device != nil && b.ip != nil {
		if err := b.device.UnsetLinkIp(b.ip, b.net); err != nil {
			libol.Error("LinuxBridger.Close.UnsetLinkIp %s : %s", b.name, err)
		}
	}
}

func (b *LinuxBridger) AddSlave(name string) error {
	link, err := tenus.NewLinkFrom(name)
	if err != nil {
		libol.Error("LinuxBridger.AddSlave: Get dev %s: %s", name, err)
		return err
	}

	if err := link.SetLinkUp(); err != nil {
		libol.Error("LinuxBridger.AddSlave.LinkUp: ", err)
	}

	if err := b.device.AddSlaveIfc(link.NetInterface()); err != nil {
		libol.Error("LinuxBridger.AddSlave: Switch dev %s: %s", name, err)
		return err
	}

	return nil
}


func (b *LinuxBridger) Name() string {
	return b.name
}

func (b *LinuxBridger) SetName(value string)  {
	b.name = value
}
