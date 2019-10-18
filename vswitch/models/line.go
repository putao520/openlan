package models

import (
	"fmt"
	"net"
)

type Line struct {
	EthType    uint16
	IpSource   net.IP
	IPDest     net.IP
	IpProtocol uint8
	PortDest   uint16
	PortSource uint16
}

func NewLine(t uint16) *Line {
	l := &Line{
		EthType:    t,
		IpSource:   nil,
		IpProtocol: 0,
		PortDest:   0,
	}
	return l
}

func (l *Line) String() string {
	return fmt.Sprintf("%d:%s:%s:%d:%d",
		l.EthType, l.IpSource, l.IPDest, l.IpProtocol, l.PortDest)
}
