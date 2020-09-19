package libol

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"os"
	"strconv"
)

type BrCtl struct {
	Name string
	Path string
}

func NewBrCtl(name string) (b *BrCtl) {
	return &BrCtl{
		Name: name,
	}
}

func (b *BrCtl) Has() bool {
	if _, err := netlink.LinkByName(b.Name); err == nil {
		return true
	}
	return false
}

func (b *BrCtl) SysPath(fun string) string {
	if b.Path == "" {
		b.Path = fmt.Sprintf("/sys/devices/virtual/net/%s/bridge", b.Name)
	}
	return fmt.Sprintf("%s/%s", b.Path, fun)
}

func (b *BrCtl) Stp(on bool) error {
	file := b.SysPath("stp_state")
	fp, err := os.OpenFile(file, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()
	if on {
		if _, err := fp.Write([]byte("1")); err != nil {
			return err
		}
	} else {
		if _, err := fp.Write([]byte("0")); err != nil {
			return err
		}
	}
	return nil
}

func (b *BrCtl) Delay(delay int) error { // by second
	file := b.SysPath("forward_delay")
	fp, err := os.OpenFile(file, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()
	if _, err := fp.Write([]byte(strconv.Itoa(delay * 100))); err != nil {
		return err
	}
	return nil
}

func (b *BrCtl) AddPort(port string) error {
	link, err := netlink.LinkByName(port)
	if err != nil {
		return err
	}
	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}
	la := netlink.LinkAttrs{TxQLen: -1, Name: b.Name}
	bridge := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkSetMaster(link, bridge); err != nil {
		return err
	}
	return nil
}

func (b *BrCtl) DelPort(port string) error {
	link, err := netlink.LinkByName(port)
	if err != nil {
		return err
	}
	if err := netlink.LinkSetNoMaster(link); err != nil {
		return err
	}
	return nil
}

type BrPort struct {
	Name string
	Path string
}

func NewBrPort(name string) (p *BrPort) {
	return &BrPort{
		Name: name,
	}
}

func (p *BrPort) SysPath(fun string) string {
	if p.Path == "" {
		p.Path = fmt.Sprintf("/sys/devices/virtual/net/%s/brport/", p.Name)
	}
	return fmt.Sprintf("%s/%s", p.Path, fun)
}

func (p *BrPort) Cost(cost int) error {
	file := p.SysPath("path_cost")
	fp, err := os.OpenFile(file, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()
	if _, err := fp.Write([]byte(strconv.Itoa(cost))); err != nil {
		return err
	}
	return nil
}
