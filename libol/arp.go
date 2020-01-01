package libol

import (
	"bytes"
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
		ProCode: ETHPIP4,
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
	var err error

	if len(frame) < 8 {
		return Errer("Arp.Decode: too small header: %d", len(frame))
	}

	reader := bytes.NewReader(frame)

	err = binary.Read(reader, binary.BigEndian, &a.HrdCode)
	err = binary.Read(reader, binary.BigEndian, &a.ProCode)
	err = binary.Read(reader, binary.BigEndian, &a.HrdLen)
	err = binary.Read(reader, binary.BigEndian, &a.ProLen)
	err = binary.Read(reader, binary.BigEndian, &a.OpCode)
	if len(frame) < int(8 + 2 * (a.HrdLen + a.ProLen)) {
		return Errer("Arp.Decode: too small frame: %d", len(frame))
	}

	a.SHwAddr = make([]byte, a.HrdLen)
	a.SIpAddr = make([]byte, a.ProLen)
	err = binary.Read(reader, binary.BigEndian, a.SHwAddr)
	err = binary.Read(reader, binary.BigEndian, a.SIpAddr)

	a.THwAddr = make([]byte, a.HrdLen)
	a.TIpAddr = make([]byte, a.ProLen)
	err = binary.Read(reader, binary.BigEndian, a.THwAddr)
	err = binary.Read(reader, binary.BigEndian, a.TIpAddr)

	a.Len = int(reader.Size()) - reader.Len()

	return err
}

func (a *Arp) Encode() []byte {
	writer := new(bytes.Buffer)

	_ = binary.Write(writer, binary.BigEndian, &a.HrdCode)
	_ = binary.Write(writer, binary.BigEndian, &a.ProCode)
	_ = binary.Write(writer, binary.BigEndian, &a.HrdLen)
	_ = binary.Write(writer, binary.BigEndian, &a.ProLen)
	_ = binary.Write(writer, binary.BigEndian, &a.OpCode)
	_ = binary.Write(writer, binary.BigEndian, a.SHwAddr[0:a.HrdLen])
	_ = binary.Write(writer, binary.BigEndian, a.SIpAddr[0:a.ProCode])
	_ = binary.Write(writer, binary.BigEndian, a.THwAddr[0:a.HrdLen])
	_ = binary.Write(writer, binary.BigEndian, a.TIpAddr[0:a.ProCode])

	a.Len = writer.Len()

	return writer.Bytes()
}

func (a *Arp) IsIP4() bool {
	return a.ProCode == ETHPIP4
}
