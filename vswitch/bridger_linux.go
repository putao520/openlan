package vswitch

import (
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/milosgajdos83/tenus"
	"net"
)

type Bridger struct {
	Ip        net.IP
	Net       *net.IPNet
	Mtu       int
	Name      string
	Device    tenus.Bridger
}

func NewBridger(name string, mtu int) *Bridger {
	b := &Bridger{
		Name: name,
		Mtu:  mtu,
	}
	return b
}

func (b *Bridger) Open(addr string) {
	var err error
	var br tenus.Bridger

	brName := b.Name
	br, err = tenus.BridgeFromName(brName)
	if err != nil {
		br, err = tenus.NewBridgeWithName(brName)
		if err != nil {
			libol.Error("Worker.newBr: %s", err)
		}
	}

	brCtl := libol.NewBrCtl(brName)
	if err := brCtl.Stp(true); err != nil {
		libol.Error("Worker.newBr.Stp: %s", err)
	}

	if err = br.SetLinkUp(); err != nil {
		libol.Error("Worker.newBr: %s", err)
	}

	libol.Info("Worker.newBr %s", brName)

	if addr != "" {
		ip, net, err := net.ParseCIDR(addr)
		if err != nil {
			libol.Error("Worker.newBr.ParseCIDR %s : %s", addr, err)
		}
		if err := br.SetLinkIp(ip, net); err != nil {
			libol.Error("Worker.newBr.SetLinkIp %s : %s", brName, err)
		}

		b.Ip = ip
		b.Net = net
	}

	b.Device = br
}

func (b *Bridger) Close() {
	if b.Device != nil && b.Ip != nil {
		if err := b.Device.UnsetLinkIp(b.Ip, b.Net); err != nil {
			libol.Error("Bridger.Close.UnsetLinkIp %s : %s", b.Name, err)
		}
	}
}

func (b *Bridger) AddSlave(name string) error {
	link, err := tenus.NewLinkFrom(name)
	if err != nil {
		libol.Error("Bridger.AddSlave: Get dev %s: %s", name, err)
		return err
	}

	if err := link.SetLinkUp(); err != nil {
		libol.Error("Bridger.AddSlave.LinkUp: ", err)
	}

	if err := b.Device.AddSlaveIfc(link.NetInterface()); err != nil {
		libol.Error("Bridger.AddSlave: Switch dev %s: %s", name, err)
		return err
	}

	return nil
}


