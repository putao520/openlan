package network

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/vishvananda/netlink"
)

type LinuxBridge struct {
	address *netlink.Addr
	ifMtu   int
	name    string
	device  netlink.Link
}

func NewLinuxBridge(name string, mtu int) *LinuxBridge {
	b := &LinuxBridge{
		name:  name,
		ifMtu: mtu,
	}
	return b
}

func (b *LinuxBridge) Open(addr string) {
	libol.Debug("LinuxBridge.Open: %s", b.name)

	la := netlink.LinkAttrs{TxQLen: -1, Name: b.name}
	br := &netlink.Bridge{LinkAttrs: la}
	link, _ := netlink.LinkByName(b.name)
	if link == nil {
		err := netlink.LinkAdd(br)
		if err != nil {
			libol.Error("LinuxBridge.Open: %s", err)
			return
		}
		link, err = netlink.LinkByName(b.name)
		if link == nil {
			libol.Error("LinuxBridge.Open: %s", err)
			return
		}
	}
	if err := netlink.LinkSetUp(link); err != nil {
		libol.Error("LinuxBridge.Open: %s", err)
	}
	libol.Info("LinuxBridge.Open %s", b.name)
	if addr != "" {
		ipAddr, err := netlink.ParseAddr(addr)
		if err != nil {
			libol.Error("LinuxBridge.Open: ParseCIDR %s : %s", addr, err)
		}
		if err := netlink.AddrAdd(link, ipAddr); err != nil {
			libol.Error("LinuxBridge.Open: SetLinkIp %s : %s", b.name, err)
		}
		b.address = ipAddr
	}
	b.device = link
}

func (b *LinuxBridge) Close() error {
	var err error

	if b.device != nil && b.address != nil {
		if err = netlink.AddrDel(b.device, b.address); err != nil {
			libol.Error("LinuxBridge.Close: UnsetLinkIp %s : %s", b.name, err)
		}
	}
	return err
}

func (b *LinuxBridge) AddSlave(dev Taper) error {
	name := dev.Name()
	link, err := netlink.LinkByName(name)
	if err != nil {
		libol.Error("LinuxBridge.AddSlave: Get %s: %s", name, err)
		return err
	}
	if err := netlink.LinkSetUp(link); err != nil {
		libol.Error("LinuxBridge.AddSlave: LinkUp: %s %s", name, err)
		return err
	}
	la := netlink.LinkAttrs{TxQLen: -1, Name: b.name}
	br := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkSetMaster(link, br); err != nil {
		libol.Error("LinuxBridge.AddSlave: master %s: %s", name, err)
		return err
	}
	dev.Slave(b)
	libol.Info("LinuxBridge.AddSlave: %s %s", name, b.name)
	return nil
}

func (b *LinuxBridge) DelSlave(dev Taper) error {
	name := dev.Name()
	link, err := netlink.LinkByName(name)
	if err != nil {
		libol.Error("LinuxBridge.DelSlave: Get %s: %s", name, err)
		return err
	}
	if err := netlink.LinkSetNoMaster(link); err != nil {
		libol.Error("LinuxBridge.DelSlave: noMaster %s: %s", name, err)
		return err
	}
	libol.Info("LinuxBridge.DelSlave: %s %s", name, b.name)
	return nil
}

func (b *LinuxBridge) Type() string {
	return "linux"
}

func (b *LinuxBridge) Name() string {
	return b.name
}

func (b *LinuxBridge) SetName(value string) {
	b.name = value
}

func (b *LinuxBridge) Input(m *Framer) error {
	return nil
}

func (b *LinuxBridge) SetTimeout(value int) {
	//TODO
}

func (b *LinuxBridge) Mtu() int {
	return b.ifMtu
}

func (b *LinuxBridge) Stp(enable bool) error {
	brCtl := libol.NewBrCtl(b.name)
	if err := brCtl.Stp(enable); err != nil {
		return err
	}
	return nil
}

func (b *LinuxBridge) Delay(value int) error {
	brCtl := libol.NewBrCtl(b.name)
	if err := brCtl.Delay(value); err != nil {
		return err
	}
	return nil
}
