package libol

import (
	"encoding/binary"
)

const (
	ARP_REQUEST = 1
	ARP_REPLY   = 2
)

const (
	ARPHRD_NETROM = 0
	ARPHRD_ETHER  = 1
	ARPHRD_EETHER = 2
)

type Arp struct {
	HrdCode uint16 // format hardware address
	ProCode uint16 // format protocol address
	HrdLen  uint8  // length of hardware address
	ProLen  uint8  // length of protocol address
	OpCode  uint16 // ARP Op(command)

	SHwAddr []byte // sender hardware address.
	SIpAddr []byte // sender IP address.
	THwAddr []byte // target hardware address.
	TIpAddr []byte // target IP address.
	Len     int
}

func NewArp() (a *Arp) {
	a = &Arp{
		HrdCode: ARPHRD_ETHER,
		ProCode: ETH_P_IP4,
		HrdLen:  6,
		ProLen:  4,
		OpCode:  ARP_REQUEST,
		Len:     0,
	}

	return
}

func NewArpFromFrame(frame []byte) (a *Arp, err error) {
	a = &Arp{
		Len: 0,
	}
	err = a.Decode(frame)
	return
}

func (a *Arp) Decode(frame []byte) error {
	if len(frame) < 8 {
		return Errer("Arp.Decode: too small header: %d", len(frame))
	}

	a.HrdCode = binary.BigEndian.Uint16(frame[0:2])
	a.ProCode = binary.BigEndian.Uint16(frame[2:4])
	a.HrdLen = uint8(frame[4])
	a.ProLen = uint8(frame[5])
	a.OpCode = binary.BigEndian.Uint16(frame[6:8])

	p := uint8(8)
	if len(frame) < int(p+2*(a.HrdLen+a.ProLen)) {
		return Errer("Arp.Decode: too small frame: %d", len(frame))
	}

	a.SHwAddr = frame[p : p+a.HrdLen]
	p += a.HrdLen
	a.SIpAddr = frame[p : p+a.ProLen]
	p += a.ProLen

	a.THwAddr = frame[p : p+a.HrdLen]
	p += a.HrdLen
	a.TIpAddr = frame[p : p+a.ProLen]
	p += a.ProLen

	a.Len = int(p)

	return nil
}

func (a *Arp) Encode() []byte {
	buffer := make([]byte, 1024)

	binary.BigEndian.PutUint16(buffer[0:2], a.HrdCode)
	binary.BigEndian.PutUint16(buffer[2:4], a.ProCode)
	buffer[4] = byte(a.HrdLen)
	buffer[5] = byte(a.ProLen)
	binary.BigEndian.PutUint16(buffer[6:8], a.OpCode)

	p := uint8(8)
	copy(buffer[p:p+a.HrdLen], a.SHwAddr[0:a.HrdLen])
	p += a.HrdLen
	copy(buffer[p:p+a.ProLen], a.SIpAddr[0:a.ProLen])
	p += a.ProLen

	copy(buffer[p:p+a.HrdLen], a.THwAddr[0:a.HrdLen])
	p += a.HrdLen
	copy(buffer[p:p+a.ProLen], a.TIpAddr[0:a.ProLen])
	p += a.ProLen

	a.Len = int(p)

	return buffer[:p]
}

func (a *Arp) IsIP4() bool {
	return a.ProCode == ETH_P_IP4
}
