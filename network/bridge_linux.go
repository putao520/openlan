package network

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/vishvananda/netlink"
)

type LinBridge struct {
	addr   *netlink.Addr
	mtu    int
	name   string
	device netlink.Link
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
	var br netlink.Link

	brName := b.name
	br, err = netlink.LinkByName(brName)
	if br == nil {
		la := netlink.LinkAttrs{TxQLen:-1, Name: brName}
		err := netlink.LinkAdd(&netlink.Bridge{LinkAttrs: la})
		if err != nil {
			libol.Error("LinBridge.newBr: %s", err)
			return
		}
		br, err = netlink.LinkByName(brName)
		if br == nil {
			return
		}
	}

	brCtl := libol.NewBrCtl(brName)
	if err := brCtl.Stp(true); err != nil {
		libol.Error("LinBridge.newBr.Stp: %s", err)
	}
	if err = netlink.LinkSetUp(br); err != nil {
		libol.Error("LinBridge.newBr: %s", err)
	}

	libol.Info("LinBridge.newBr %s", brName)
	if addr != "" {
		ipAddr, err := netlink.ParseAddr(addr)
		if err != nil {
			libol.Error("LinBridge.newBr.ParseCIDR %s : %s", addr, err)
		}
		if err := netlink.AddrAdd(br, ipAddr); err != nil {
			libol.Error("LinBridge.newBr.SetLinkIp %s : %s", brName, err)
		}
		b.addr = ipAddr
	}

	b.device = br
}

func (b *LinBridge) Close() error {
	var err error

	if b.device != nil && b.addr != nil {
		if err = netlink.AddrDel(b.device, b.addr); err != nil {
			libol.Error("LinBridge.Close.UnsetLinkIp %s : %s", b.name, err)
		}
	}
	return err
}

func (b *LinBridge) AddSlave(dev Taper) error {
	name := dev.Name()

	link, err := netlink.LinkByName(name)
	if err != nil {
		libol.Error("LinBridge.AddSlave: Get dev %s: %s", name, err)
		return err
	}
	if err := netlink.LinkSetUp(link); err != nil {
		libol.Error("LinBridge.AddSlave.LinkUp: %s %s", name, err)
		return err
	}
	if err := netlink.LinkSetMaster(link, b.device); err != nil {
		libol.Error("LinBridge.AddSlave: Switch dev %s: %s", name, err)
		return err
	}

	dev.Slave(b)
	libol.Info("LinBridge.AddSlave: %s %s", name, b.name)

	return nil
}

func (b *LinBridge) DelSlave(dev Taper) error {
	name := dev.Name()

	link, err := netlink.LinkByName(name)
	if err != nil {
		libol.Error("LinBridge.DelSlave: Get dev %s: %s", name, err)
		return err
	}
	if err := netlink.LinkSetMaster(link, b.device); err != nil {
		libol.Error("LinBridge.DelSlave: Switch dev %s: %s", name, err)
		return err
	}

	libol.Info("LinBridge.DelSlave: %s %s", name, b.name)

	return nil
}

func (b *LinBridge) Name() string {
	return b.name
}

func (b *LinBridge) SetName(value string) {
	b.name = value
}

func (b *LinBridge) Input(m *Framer) error {
	return nil
}

func (b *LinBridge) SetTimeout(value int) {
	//TODO
}
