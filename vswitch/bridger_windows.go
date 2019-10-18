package vswitch

import (
	"net"
)

type Bridger struct {
	Ip        net.IP
	Net       *net.IPNet
	Mtu       int
	Name      string
}

func NewBridger(name string, mtu int) *Bridger {
	b := &Bridger{
		Name: name,
		Mtu:  mtu,
	}
	return b
}

func (b *Bridger) Open(addr string) {
	//TODO
}

func (b *Bridger) Close() {
	//TODO
}

func (b *Bridger) AddSlave(name string) error {
	//TODO
	return nil
}


