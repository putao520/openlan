package models

import (
	"fmt"
	"net"
	"time"
)

type Line struct {
	EthType    uint16
	IpSource   net.IP
	IPDest     net.IP
	IpProtocol uint8
	PortDest   uint16
	PortSource uint16
	NewTime    int64
	HitTime    int64
}

func NewLine(t uint16) *Line {
	l := &Line{
		EthType:    t,
		IpSource:   nil,
		IpProtocol: 0,
		PortDest:   0,
		NewTime:    time.Now().Unix(),
		HitTime:    time.Now().Unix(),
	}
	return l
}

func (l *Line) String() string {
	return fmt.Sprintf("%d:%s:%s:%d:%d:%d",
		l.EthType, l.IpSource, l.IPDest, l.IpProtocol, l.PortSource, l.PortDest)
}

func (l *Line) Update() {
	l.HitTime = time.Now().Unix()
}
